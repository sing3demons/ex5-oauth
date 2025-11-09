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

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestListSessions(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_session_handler")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)

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
		ID:        "test-user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test SSO sessions
	session1 := &models.SSOSession{
		SessionID:     "session-1",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
	}
	if err := ssoSessionRepo.Create(ctx, session1); err != nil {
		t.Fatalf("Failed to create test session 1: %v", err)
	}

	session2 := &models.SSOSession{
		SessionID:     "session-2",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.2",
		UserAgent:     "Chrome/90.0",
	}
	if err := ssoSessionRepo.Create(ctx, session2); err != nil {
		t.Fatalf("Failed to create test session 2: %v", err)
	}

	// Generate access token for authentication
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

	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request
	req := httptest.NewRequest("GET", "/account/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListSessions(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Parse response
	var response ListSessionsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify sessions are returned
	if len(response.Sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(response.Sessions))
	}

	// Verify session data
	foundSession1 := false
	foundSession2 := false
	for _, s := range response.Sessions {
		if s.SessionID == "session-1" {
			foundSession1 = true
			if s.IPAddress != "192.168.1.1" {
				t.Errorf("Expected IP 192.168.1.1, got %s", s.IPAddress)
			}
		}
		if s.SessionID == "session-2" {
			foundSession2 = true
			if s.IPAddress != "192.168.1.2" {
				t.Errorf("Expected IP 192.168.1.2, got %s", s.IPAddress)
			}
		}
	}

	if !foundSession1 || !foundSession2 {
		t.Error("Not all sessions were returned")
	}
}

func TestRevokeSession(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_session_revoke")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)

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
		ID:        "test-user-456",
		Email:     "test2@example.com",
		Name:      "Test User 2",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test SSO session
	session := &models.SSOSession{
		SessionID:     "session-to-revoke",
		UserID:        testUser.ID,
		Authenticated: true,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		LastActivity:  time.Now(),
		IPAddress:     "192.168.1.100",
		UserAgent:     "Test Agent",
	}
	if err := ssoSessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Generate access token for authentication
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

	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request with mux vars
	req := httptest.NewRequest("DELETE", "/account/sessions/session-to-revoke", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req = mux.SetURLVars(req, map[string]string{"session_id": "session-to-revoke"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.RevokeSession(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify session was deleted
	_, err = ssoSessionRepo.FindBySessionID(ctx, "session-to-revoke")
	if err == nil {
		t.Error("Expected session to be deleted, but it still exists")
	}
}

func TestRevokeSessionUnauthorized(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_session_unauth")
	defer db.Drop(ctx)

	// Initialize repositories
	ssoSessionRepo := repository.NewSSOSessionRepository(db)

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

	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request without authorization header
	req := httptest.NewRequest("DELETE", "/account/sessions/some-session", nil)
	req = mux.SetURLVars(req, map[string]string{"session_id": "some-session"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.RevokeSession(rr, req)

	// Check status code
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestListAuthorizations(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_list_authorizations")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

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
		ID:        "test-user-auth-123",
		Email:     "testauth@example.com",
		Name:      "Test Auth User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test clients
	client1 := &models.Client{
		ClientID:     "test-client-1",
		ClientSecret: "secret1",
		Name:         "Test Application 1",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		CreatedAt:    time.Now(),
	}
	if err := clientRepo.Create(ctx, client1); err != nil {
		t.Fatalf("Failed to create test client 1: %v", err)
	}

	client2 := &models.Client{
		ClientID:     "test-client-2",
		ClientSecret: "secret2",
		Name:         "Test Application 2",
		RedirectURIs: []string{"http://localhost:3001/callback"},
		CreatedAt:    time.Now(),
	}
	if err := clientRepo.Create(ctx, client2); err != nil {
		t.Fatalf("Failed to create test client 2: %v", err)
	}

	// Create test consents
	consent1 := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  client1.ClientID,
		Scopes:    []string{"openid", "profile", "email"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent1); err != nil {
		t.Fatalf("Failed to create test consent 1: %v", err)
	}

	consent2 := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  client2.ClientID,
		Scopes:    []string{"openid", "profile"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent2); err != nil {
		t.Fatalf("Failed to create test consent 2: %v", err)
	}

	// Generate access token for authentication
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

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request
	req := httptest.NewRequest("GET", "/account/authorizations", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListAuthorizations(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Parse response
	var response ListAuthorizationsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify authorizations are returned
	if len(response.Authorizations) != 2 {
		t.Errorf("Expected 2 authorizations, got %d", len(response.Authorizations))
	}

	// Verify authorization data
	foundClient1 := false
	foundClient2 := false
	for _, auth := range response.Authorizations {
		if auth.ClientID == "test-client-1" {
			foundClient1 = true
			if auth.ClientName != "Test Application 1" {
				t.Errorf("Expected client name 'Test Application 1', got %s", auth.ClientName)
			}
			if len(auth.Scopes) != 3 {
				t.Errorf("Expected 3 scopes for client 1, got %d", len(auth.Scopes))
			}
		}
		if auth.ClientID == "test-client-2" {
			foundClient2 = true
			if auth.ClientName != "Test Application 2" {
				t.Errorf("Expected client name 'Test Application 2', got %s", auth.ClientName)
			}
			if len(auth.Scopes) != 2 {
				t.Errorf("Expected 2 scopes for client 2, got %d", len(auth.Scopes))
			}
		}
	}

	if !foundClient1 || !foundClient2 {
		t.Error("Not all authorizations were returned")
	}
}

func TestRevokeAuthorization(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_revoke_authorization")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

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
		ID:        "test-user-revoke-123",
		Email:     "testrevoke@example.com",
		Name:      "Test Revoke User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:     "test-client-revoke",
		ClientSecret: "secret",
		Name:         "Test Revoke Application",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		CreatedAt:    time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create test consent
	consent := &models.UserConsent{
		UserID:    testUser.ID,
		ClientID:  testClient.ClientID,
		Scopes:    []string{"openid", "profile"},
		GrantedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	if err := consentRepo.Create(ctx, consent); err != nil {
		t.Fatalf("Failed to create test consent: %v", err)
	}

	// Generate access token for authentication
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

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request with mux vars
	req := httptest.NewRequest("DELETE", "/account/authorizations/test-client-revoke", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req = mux.SetURLVars(req, map[string]string{"client_id": "test-client-revoke"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.RevokeAuthorization(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify consent was deleted
	_, err = consentRepo.FindByUserAndClient(ctx, testUser.ID, testClient.ClientID)
	if err == nil {
		t.Error("Expected consent to be deleted, but it still exists")
	}
}

func TestRevokeAuthorizationNotFound(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_revoke_not_found")
	defer db.Drop(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	ssoSessionRepo := repository.NewSSOSessionRepository(db)
	consentRepo := repository.NewUserConsentRepository(db)
	clientRepo := repository.NewClientRepository(db)

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
		ID:        "test-user-notfound-123",
		Email:     "testnotfound@example.com",
		Name:      "Test Not Found User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate access token for authentication
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

	handler := NewSessionHandler(ssoSessionRepo, consentRepo, clientRepo, cfg)

	// Create request for non-existent authorization
	req := httptest.NewRequest("DELETE", "/account/authorizations/non-existent-client", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req = mux.SetURLVars(req, map[string]string{"client_id": "non-existent-client"})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.RevokeAuthorization(rr, req)

	// Check status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}
