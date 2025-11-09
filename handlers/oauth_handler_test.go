//go:build integration
// +build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"oauth2-server/config"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestUserInfoEndpoint tests the UserInfo endpoint with different scopes
func TestUserInfoEndpoint(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_userinfo")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
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
		ID:    "test-user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	tests := []struct {
		name           string
		scope          string
		expectedClaims []string
		notExpected    []string
	}{
		{
			name:           "Only openid scope returns only sub",
			scope:          "openid",
			expectedClaims: []string{"sub"},
			notExpected:    []string{"email", "name", "email_verified"},
		},
		{
			name:           "openid and email scope returns sub and email claims",
			scope:          "openid email",
			expectedClaims: []string{"sub", "email", "email_verified"},
			notExpected:    []string{"name"},
		},
		{
			name:           "openid and profile scope returns sub and profile claims",
			scope:          "openid profile",
			expectedClaims: []string{"sub", "name"},
			notExpected:    []string{"email", "email_verified"},
		},
		{
			name:           "openid, profile, and email scope returns all claims",
			scope:          "openid profile email",
			expectedClaims: []string{"sub", "name", "email", "email_verified"},
			notExpected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate access token with specific scope
			accessToken, err := utils.GenerateAccessToken(
				testUser.ID,
				testUser.Email,
				testUser.Name,
				tt.scope,
				privateKey,
				cfg.AccessTokenExpiry,
			)
			if err != nil {
				t.Fatalf("Failed to generate access token: %v", err)
			}

			// Create request with Bearer token
			req := httptest.NewRequest("GET", "/oauth/userinfo", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			w := httptest.NewRecorder()

			// Call UserInfo endpoint
			handler.UserInfo(w, req)

			// Check response status
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
				return
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Verify expected claims are present
			for _, claim := range tt.expectedClaims {
				if _, exists := response[claim]; !exists {
					t.Errorf("Expected claim '%s' not found in response: %v", claim, response)
				}
			}

			// Verify unexpected claims are not present
			for _, claim := range tt.notExpected {
				if _, exists := response[claim]; exists {
					t.Errorf("Unexpected claim '%s' found in response: %v", claim, response)
				}
			}

			// Verify sub claim always matches user ID
			if sub, ok := response["sub"].(string); !ok || sub != testUser.ID {
				t.Errorf("Expected sub='%s', got '%v'", testUser.ID, response["sub"])
			}
		})
	}
}

// TestUserInfoEndpointWithJWE tests the UserInfo endpoint with JWE tokens
func TestUserInfoEndpointWithJWE(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_userinfo_jwe")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
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
		ID:        "test-user-jwe-123",
		Email:     "jwe@example.com",
		Name:      "JWE Test User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Test with JWE token containing only openid scope
	scope := "openid"
	jweToken, err := utils.GenerateJWEAccessToken(
		testUser.ID,
		testUser.Email,
		testUser.Name,
		scope,
		publicKey,
		time.Now().Add(time.Hour).Unix(),
	)
	if err != nil {
		t.Fatalf("Failed to generate JWE token: %v", err)
	}

	// Create request with Bearer token
	req := httptest.NewRequest("GET", "/oauth/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+jweToken)
	w := httptest.NewRecorder()

	// Call UserInfo endpoint
	handler.UserInfo(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify only sub claim is present (openid scope only)
	if _, exists := response["sub"]; !exists {
		t.Errorf("Expected 'sub' claim not found in response: %v", response)
	}

	// Verify email and name are not present
	if _, exists := response["email"]; exists {
		t.Errorf("Unexpected 'email' claim found in response with only openid scope: %v", response)
	}
	if _, exists := response["name"]; exists {
		t.Errorf("Unexpected 'name' claim found in response with only openid scope: %v", response)
	}
}

// TestUserInfoEndpointUnauthorized tests unauthorized access
func TestUserInfoEndpointUnauthorized(t *testing.T) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_userinfo_unauth")
	defer db.Drop(ctx)

	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)

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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Test without Authorization header
	req := httptest.NewRequest("GET", "/oauth/userinfo", nil)
	w := httptest.NewRecorder()
	handler.UserInfo(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test with invalid token
	req = httptest.NewRequest("GET", "/oauth/userinfo", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	handler.UserInfo(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
	}
}

// TestAuthorizeWithPromptParameter tests the OIDC prompt parameter support
func TestAuthorizeWithPromptParameter(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_prompt")
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
		ID:        "test-user-prompt",
		Email:     "prompt@example.com",
		Name:      "Prompt Test User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "test-client-prompt",
		ClientSecret:  "test-secret",
		Name:          "Test Client",
		RedirectURIs:  []string{"http://localhost:3000/callback"},
		AllowedScopes: []string{"openid", "profile", "email"},
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	t.Run("prompt=none without SSO session returns login_required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid&state=xyz&prompt=none", nil)
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect with login_required error
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "error=login_required") {
			t.Errorf("Expected login_required error in redirect, got: %s", location)
		}
		if !contains(location, "state=xyz") {
			t.Errorf("Expected state parameter in redirect, got: %s", location)
		}
	})

	t.Run("prompt=none with SSO but no consent returns consent_required", func(t *testing.T) {
		// Create SSO session
		ssoSession := &models.SSOSession{
			SessionID:     "test-sso-session-1",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
		}
		if err := ssoSessionRepo.Create(ctx, ssoSession); err != nil {
			t.Fatalf("Failed to create SSO session: %v", err)
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid&state=abc&prompt=none", nil)
		// Add SSO session to context
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect with consent_required error
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "error=consent_required") {
			t.Errorf("Expected consent_required error in redirect, got: %s", location)
		}
		if !contains(location, "state=abc") {
			t.Errorf("Expected state parameter in redirect, got: %s", location)
		}
	})

	t.Run("prompt=none with SSO and consent returns authorization code", func(t *testing.T) {
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

		ssoSession := &models.SSOSession{
			SessionID:     "test-sso-session-2",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid+profile&state=def&prompt=none", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect with authorization code
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "code=") {
			t.Errorf("Expected authorization code in redirect, got: %s", location)
		}
		if !contains(location, "state=def") {
			t.Errorf("Expected state parameter in redirect, got: %s", location)
		}
		if contains(location, "error=") {
			t.Errorf("Unexpected error in redirect, got: %s", location)
		}
	})

	t.Run("prompt=login forces re-authentication even with SSO", func(t *testing.T) {
		ssoSession := &models.SSOSession{
			SessionID:     "test-sso-session-3",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid&state=ghi&prompt=login", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect to login page
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "/auth/login") {
			t.Errorf("Expected redirect to login page, got: %s", location)
		}
		if !contains(location, "session_id=") {
			t.Errorf("Expected session_id in redirect, got: %s", location)
		}
	})

	t.Run("prompt=consent forces consent screen even with existing consent", func(t *testing.T) {
		ssoSession := &models.SSOSession{
			SessionID:     "test-sso-session-4",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid+profile&state=jkl&prompt=consent", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect to consent screen
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "/oauth/consent") {
			t.Errorf("Expected redirect to consent screen, got: %s", location)
		}
		if !contains(location, "client_id=test-client-prompt") {
			t.Errorf("Expected client_id in redirect, got: %s", location)
		}
		if !contains(location, "state=jkl") {
			t.Errorf("Expected state parameter in redirect, got: %s", location)
		}
	})

	t.Run("prompt=select_account forces re-authentication", func(t *testing.T) {
		ssoSession := &models.SSOSession{
			SessionID:     "test-sso-session-5",
			UserID:        testUser.ID,
			Authenticated: true,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
			LastActivity:  time.Now(),
		}

		req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&client_id=test-client-prompt&redirect_uri=http://localhost:3000/callback&scope=openid&state=mno&prompt=select_account", nil)
		req = req.WithContext(context.WithValue(req.Context(), "sso_session", ssoSession))
		w := httptest.NewRecorder()

		handler.Authorize(w, req)

		// Should redirect to login page (placeholder behavior)
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if !contains(location, "/auth/login") {
			t.Errorf("Expected redirect to login page, got: %s", location)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
