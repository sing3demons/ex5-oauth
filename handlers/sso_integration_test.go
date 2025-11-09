//go:build integration
// +build integration

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"oauth2-server/config"
	"oauth2-server/middleware"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestFirstLoginFlow tests the complete first login flow: no SSO → login → consent → code
func TestFirstLoginFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_first_login")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "first-login-user",
		Email:     "firstlogin@example.com",
		Name:      "First Login User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "first-login-client",
		ClientSecret:  "test-secret",
		Name:          "First Login App",
		RedirectURIs:  []string{"http://localhost:3000/callback"},
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)
	_ = NewAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, ssoSessionRepo, cfg)
	consentHandler := NewConsentHandler(clientRepo, consentRepo, authCodeRepo, sessionRepo, cfg)

	// Step 1: User visits authorization endpoint without SSO session
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=first-login-client&redirect_uri=http://localhost:3000/callback&scope=openid+profile+email&state=test-state", nil)
	w := httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect to login page
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to login, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/auth/login?session_id=") {
		t.Fatalf("Expected redirect to login page, got: %s", location)
	}

	// Extract session ID
	sessionID := strings.TrimPrefix(location, "/auth/login?session_id=")

	// Step 2: User submits login form (simulated)
	// In real flow, user would see login page and submit credentials
	// Here we directly call the login handler with the session
	_, err = sessionRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	// Simulate successful authentication by creating SSO session and auth code
	ssoSessionID, _ := utils.GenerateRandomString(32)
	ssoSession := &models.SSOSession{
		SessionID:     ssoSessionID,
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	// Step 3: After login, user should be redirected to consent screen
	// Check if consent exists (it shouldn't for first login)
	hasConsent, _ := consentRepo.HasConsent(ctx, testUser.ID, testClient.ClientID, []string{"openid", "profile", "email"})
	if hasConsent {
		t.Fatal("Expected no consent for first login")
	}

	// Step 4: User sees consent screen and approves
	form := url.Values{}
	form.Set("action", "allow")
	form.Set("client_id", testClient.ClientID)
	form.Set("scope", "openid profile email")
	form.Set("state", "test-state")
	form.Set("redirect_uri", "http://localhost:3000/callback")

	req = httptest.NewRequest("POST", "/oauth/consent", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), middleware.SSOSessionContextKey, ssoSession))
	w = httptest.NewRecorder()

	consentHandler.HandleConsent(w, req)

	// Should redirect with authorization code
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect with code, got status %d: %s", w.Code, w.Body.String())
	}

	location = w.Header().Get("Location")
	if !strings.Contains(location, "code=") {
		t.Fatalf("Expected authorization code in redirect, got: %s", location)
	}
	if !strings.Contains(location, "state=test-state") {
		t.Fatalf("Expected state parameter in redirect, got: %s", location)
	}

	// Step 5: Verify consent was saved
	consent, err := consentRepo.FindByUserAndClient(ctx, testUser.ID, testClient.ClientID)
	if err != nil {
		t.Fatalf("Expected consent to be saved: %v", err)
	}
	if len(consent.Scopes) != 3 {
		t.Errorf("Expected 3 scopes in consent, got %d", len(consent.Scopes))
	}

	t.Log("First login flow completed successfully")
}

// TestSecondAppWithSSO tests the second app flow: SSO exists → consent exists → auto-approve → code
func TestSecondAppWithSSO(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_second_app")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "second-app-user",
		Email:     "secondapp@example.com",
		Name:      "Second App User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "second-app-client",
		ClientSecret:  "test-secret",
		Name:          "Second App",
		RedirectURIs:  []string{"http://localhost:3001/callback"},
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create existing SSO session (user already logged in)
	ssoSession := &models.SSOSession{
		SessionID:     "existing-sso-session",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	// Create existing consent (user already approved this app)
	consent := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scopes:    []string{"openid", "profile", "email"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent); err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// User visits authorization endpoint with SSO session
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=second-app-client&redirect_uri=http://localhost:3001/callback&scope=openid+profile+email&state=second-state", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
	w := httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect immediately with authorization code (auto-approval)
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect with code, got status %d: %s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "code=") {
		t.Fatalf("Expected authorization code in redirect, got: %s", location)
	}
	if !strings.Contains(location, "state=second-state") {
		t.Fatalf("Expected state parameter in redirect, got: %s", location)
	}
	if strings.Contains(location, "/auth/login") {
		t.Fatalf("Should not redirect to login with valid SSO session")
	}
	if strings.Contains(location, "/oauth/consent") {
		t.Fatalf("Should not redirect to consent with existing consent")
	}

	t.Log("Second app with SSO flow completed successfully (auto-approved)")
}

// TestLogoutFlow tests the logout flow: logout → SSO cleared → requires login
func TestLogoutFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_logout")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "logout-user",
		Email:     "logout@example.com",
		Name:      "Logout User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "logout-client",
		ClientSecret:  "test-secret",
		Name:          "Logout App",
		RedirectURIs:  []string{"http://localhost:3002/callback"},
		AllowedScopes: []string{"openid", "profile"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create SSO session
	ssoSessionID := "logout-sso-session"
	ssoSession := &models.SSOSession{
		SessionID:     ssoSessionID,
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	authHandler := NewAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, ssoSessionRepo, cfg)
	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Step 1: Verify SSO session exists
	foundSession, err := ssoSessionRepo.FindBySessionID(ctx, ssoSessionID)
	if err != nil {
		t.Fatalf("SSO session should exist before logout: %v", err)
	}
	if foundSession.SessionID != ssoSessionID {
		t.Fatalf("Expected session ID %s, got %s", ssoSessionID, foundSession.SessionID)
	}

	// Step 2: User logs out
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "oauth_sso_session",
		Value: ssoSessionID,
	})
	w := httptest.NewRecorder()

	authHandler.Logout(w, req)

	// Should return success
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Step 3: Verify SSO session was deleted
	_, err = ssoSessionRepo.FindBySessionID(ctx, ssoSessionID)
	if err == nil {
		t.Fatal("SSO session should be deleted after logout")
	}

	// Step 4: Verify cookie was cleared
	cookies := w.Result().Cookies()
	foundClearedCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "oauth_sso_session" && cookie.MaxAge == -1 {
			foundClearedCookie = true
			break
		}
	}
	if !foundClearedCookie {
		t.Fatal("SSO cookie should be cleared after logout")
	}

	// Step 5: Try to access authorization endpoint without SSO session
	req = httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=logout-client&redirect_uri=http://localhost:3002/callback&scope=openid+profile&state=logout-state", nil)
	w = httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect to login page
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to login, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/auth/login") {
		t.Fatalf("Expected redirect to login page after logout, got: %s", location)
	}

	t.Log("Logout flow completed successfully")
}

// TestExpiredSessionFlow tests the expired session flow: expired SSO → requires login
func TestExpiredSessionFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_expired_session")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "expired-user",
		Email:     "expired@example.com",
		Name:      "Expired User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "expired-client",
		ClientSecret:  "test-secret",
		Name:          "Expired App",
		RedirectURIs:  []string{"http://localhost:3003/callback"},
		AllowedScopes: []string{"openid", "profile"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create expired SSO session
	expiredSession := &models.SSOSession{
		SessionID:     "expired-sso-session",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now().Add(-8 * 24 * time.Hour), // 8 days ago
		ExpiresAt:     time.Now().Add(-1 * time.Hour),      // Expired 1 hour ago
		LastActivity:  time.Now().Add(-8 * 24 * time.Hour),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, expiredSession); err != nil {
		t.Fatalf("Failed to create expired SSO session: %v", err)
	}

	// Setup SSO middleware
	ssoMiddleware := middleware.SSOMiddleware(ssoSessionRepo)
	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create request with expired SSO cookie
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=expired-client&redirect_uri=http://localhost:3003/callback&scope=openid+profile&state=expired-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  "oauth_sso_session",
		Value: "expired-sso-session",
	})

	// Apply SSO middleware
	w := httptest.NewRecorder()
	handler := ssoMiddleware(http.HandlerFunc(oauthHandler.Authorize))
	handler.ServeHTTP(w, req)

	// Should redirect to login page (expired session ignored by middleware)
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to login, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/auth/login") {
		t.Fatalf("Expected redirect to login page with expired session, got: %s", location)
	}

	t.Log("Expired session flow completed successfully")
}

// TestConsentRevocationFlow tests the consent revocation flow: revoke → requires consent again
func TestConsentRevocationFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_consent_revocation")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "revoke-user",
		Email:     "revoke@example.com",
		Name:      "Revoke User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "revoke-client",
		ClientSecret:  "test-secret",
		Name:          "Revoke App",
		RedirectURIs:  []string{"http://localhost:3004/callback"},
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create SSO session
	ssoSession := &models.SSOSession{
		SessionID:     "revoke-sso-session",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	// Create consent
	consent := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scopes:    []string{"openid", "profile", "email"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent); err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)
	sessionHandler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Step 1: Verify auto-approval works with consent
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=revoke-client&redirect_uri=http://localhost:3004/callback&scope=openid+profile+email&state=before-revoke", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
	w := httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should auto-approve
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect with code, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "code=") {
		t.Fatalf("Expected authorization code before revocation, got: %s", location)
	}

	// Step 2: Revoke consent
	accessToken, err := utils.GenerateAccessToken(
		testUser.ID,
		testUser.Email,
		testUser.Name,
		"openid profile email",
		privateKey,
		3600,
	)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	req = httptest.NewRequest("DELETE", "/account/authorizations/revoke-client", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req = mux.SetURLVars(req, map[string]string{"client_id": "revoke-client"})
	w = httptest.NewRecorder()

	sessionHandler.RevokeAuthorization(w, req)

	// Should succeed
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for revocation, got %d: %s", w.Code, w.Body.String())
	}

	// Step 3: Verify consent was deleted
	_, err = consentRepo.FindByUserAndClient(ctx, testUser.ID, testClient.ClientID)
	if err == nil {
		t.Fatal("Consent should be deleted after revocation")
	}

	// Step 4: Try to authorize again - should require consent
	req = httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=revoke-client&redirect_uri=http://localhost:3004/callback&scope=openid+profile+email&state=after-revoke", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
	w = httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect to consent screen
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to consent, got status %d", w.Code)
	}

	location = w.Header().Get("Location")
	if !strings.Contains(location, "/oauth/consent") {
		t.Fatalf("Expected redirect to consent screen after revocation, got: %s", location)
	}

	t.Log("Consent revocation flow completed successfully")
}

// TestPromptParameterLogin tests prompt=login forces re-authentication
func TestPromptParameterLogin(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_prompt_login")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "prompt-login-user",
		Email:     "promptlogin@example.com",
		Name:      "Prompt Login User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "prompt-login-client",
		ClientSecret:  "test-secret",
		Name:          "Prompt Login App",
		RedirectURIs:  []string{"http://localhost:3005/callback"},
		AllowedScopes: []string{"openid", "profile"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create valid SSO session
	ssoSession := &models.SSOSession{
		SessionID:     "prompt-login-sso",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	// Create consent
	consent := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scopes:    []string{"openid", "profile"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent); err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Request with prompt=login should force re-authentication even with valid SSO
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=prompt-login-client&redirect_uri=http://localhost:3005/callback&scope=openid+profile&state=login-state&prompt=login", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
	w := httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect to login page
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to login, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/auth/login") {
		t.Fatalf("Expected redirect to login page with prompt=login, got: %s", location)
	}

	t.Log("prompt=login flow completed successfully")
}

// TestPromptParameterConsent tests prompt=consent forces consent screen
func TestPromptParameterConsent(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_prompt_consent")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "prompt-consent-user",
		Email:     "promptconsent@example.com",
		Name:      "Prompt Consent User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "prompt-consent-client",
		ClientSecret:  "test-secret",
		Name:          "Prompt Consent App",
		RedirectURIs:  []string{"http://localhost:3006/callback"},
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create valid SSO session
	ssoSession := &models.SSOSession{
		SessionID:     "prompt-consent-sso",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		t.Fatalf("Failed to create SSO session: %v", err)
	}

	// Create existing consent
	consent := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scopes:    []string{"openid", "profile", "email"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent); err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Request with prompt=consent should force consent screen even with existing consent
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=prompt-consent-client&redirect_uri=http://localhost:3006/callback&scope=openid+profile+email&state=consent-state&prompt=consent", nil)
	req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
	w := httptest.NewRecorder()

	oauthHandler.Authorize(w, req)

	// Should redirect to consent screen
	if w.Code != http.StatusFound {
		t.Fatalf("Expected redirect to consent, got status %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "/oauth/consent") {
		t.Fatalf("Expected redirect to consent screen with prompt=consent, got: %s", location)
	}

	t.Log("prompt=consent flow completed successfully")
}

// TestPromptParameterNone tests prompt=none behavior
func TestPromptParameterNone(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_prompt_none")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

	// Load test keys
	privateKey, publicKey, err := utils.LoadTestKeys()
	if err != nil {
		t.Fatalf("Failed to load test keys: %v", err)
	}

	cfg := &config.Config{
		PrivateKey:         privateKey,
		PublicKey:          publicKey,
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}

	// Create test user
	testUser := &models.User{
		ID:        "prompt-none-user",
		Email:     "promptnone@example.com",
		Name:      "Prompt None User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "prompt-none-client",
		ClientSecret:  "test-secret",
		Name:          "Prompt None App",
		RedirectURIs:  []string{"http://localhost:3007/callback"},
		AllowedScopes: []string{"openid", "profile"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	oauthHandler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Test 1: prompt=none without SSO session returns login_required
	t.Run("without SSO returns login_required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=prompt-none-client&redirect_uri=http://localhost:3007/callback&scope=openid+profile&state=none-state-1&prompt=none", nil)
		w := httptest.NewRecorder()

		oauthHandler.Authorize(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("Expected redirect, got status %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !strings.Contains(location, "error=login_required") {
			t.Fatalf("Expected login_required error with prompt=none and no SSO, got: %s", location)
		}
		if !strings.Contains(location, "state=none-state-1") {
			t.Fatalf("Expected state parameter in error redirect, got: %s", location)
		}
	})

	// Test 2: prompt=none with SSO but no consent returns consent_required
	t.Run("with SSO but no consent returns consent_required", func(t *testing.T) {
		ssoSession := &models.SSOSession{
			SessionID:     "prompt-none-sso-1",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
			IPAddress:     "192.168.1.1",
			UserAgent:     "Test Agent",
		}
		if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
			t.Fatalf("Failed to create SSO session: %v", err)
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=prompt-none-client&redirect_uri=http://localhost:3007/callback&scope=openid+profile&state=none-state-2&prompt=none", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		oauthHandler.Authorize(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("Expected redirect, got status %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !strings.Contains(location, "error=consent_required") {
			t.Fatalf("Expected consent_required error with prompt=none and no consent, got: %s", location)
		}
		if !strings.Contains(location, "state=none-state-2") {
			t.Fatalf("Expected state parameter in error redirect, got: %s", location)
		}
	})

	// Test 3: prompt=none with SSO and consent returns authorization code
	t.Run("with SSO and consent returns code", func(t *testing.T) {
		ssoSession := &models.SSOSession{
			SessionID:     "prompt-none-sso-2",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
			IPAddress:     "192.168.1.1",
			UserAgent:     "Test Agent",
		}
		if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
			t.Fatalf("Failed to create SSO session: %v", err)
		}

		consent := &models.UserConsent{
			UserID:    testUser.ID,
			ClientID:  testClient.ClientID,
			Scopes:    []string{"openid", "profile"},
			GrantedAt: time.Now(),
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		}
		if err := consentRepo.Create(ctx, consent); err != nil {
			t.Fatalf("Failed to create consent: %v", err)
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=prompt-none-client&redirect_uri=http://localhost:3007/callback&scope=openid+profile&state=none-state-3&prompt=none", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		oauthHandler.Authorize(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("Expected redirect, got status %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !strings.Contains(location, "code=") {
			t.Fatalf("Expected authorization code with prompt=none, SSO, and consent, got: %s", location)
		}
		if !strings.Contains(location, "state=none-state-3") {
			t.Fatalf("Expected state parameter in redirect, got: %s", location)
		}
		if strings.Contains(location, "error=") {
			t.Fatalf("Unexpected error with prompt=none, SSO, and consent, got: %s", location)
		}
	})

	t.Log("prompt=none flow completed successfully")
}
