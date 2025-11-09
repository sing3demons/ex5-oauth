//go:build integration
// +build integration

package handlers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"oauth2-server/config"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper function to parse and validate ID tokens
func parseIDToken(tokenString string, publicKey *rsa.PublicKey) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result := make(map[string]interface{})
		for k, v := range claims {
			result[k] = v
		}
		return result, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// Helper function to parse and validate access tokens
func parseAccessToken(tokenString string, publicKey *rsa.PublicKey) (*utils.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &utils.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*utils.AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// TestScopeValidationFlow tests the complete scope validation flow
func TestScopeValidationFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_scope_validation")
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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create test client with allowed scopes
	testClient := &models.Client{
		ClientID:      "test-client-123",
		ClientSecret:  "test-secret",
		RedirectURIs:  []string{"https://example.com/callback"},
		Name:          "Test Client",
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create test client with no scope restrictions
	unrestrictedClient := &models.Client{
		ClientID:      "unrestricted-client",
		ClientSecret:  "unrestricted-secret",
		RedirectURIs:  []string{"https://example.com/callback"},
		Name:          "Unrestricted Client",
		AllowedScopes: []string{}, // Empty means all scopes allowed
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, unrestrictedClient); err != nil {
		t.Fatalf("Failed to create unrestricted client: %v", err)
	}

	tests := []struct {
		name           string
		clientID       string
		scope          string
		expectedStatus int
		expectedError  string
		checkRedirect  bool
	}{
		{
			name:           "Valid scopes - should succeed",
			clientID:       "test-client-123",
			scope:          "openid profile email",
			expectedStatus: http.StatusFound,
			checkRedirect:  true,
		},
		{
			name:           "Valid subset of scopes - should succeed",
			clientID:       "test-client-123",
			scope:          "openid profile",
			expectedStatus: http.StatusFound,
			checkRedirect:  true,
		},
		{
			name:           "Invalid scope - should fail",
			clientID:       "test-client-123",
			scope:          "openid invalid_scope",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_scope",
		},
		{
			name:           "Unauthorized scope (client restriction) - should fail",
			clientID:       "test-client-123",
			scope:          "openid profile email phone",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_scope",
		},
		{
			name:           "No scope parameter - should use default",
			clientID:       "test-client-123",
			scope:          "",
			expectedStatus: http.StatusFound,
			checkRedirect:  true,
		},
		{
			name:           "Missing openid scope - should fail",
			clientID:       "test-client-123",
			scope:          "profile email",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_scope",
		},
		{
			name:           "Unrestricted client with all scopes - should succeed",
			clientID:       "unrestricted-client",
			scope:          "openid profile email phone address",
			expectedStatus: http.StatusFound,
			checkRedirect:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build authorization request
			params := url.Values{}
			params.Set("response_type", "code")
			params.Set("client_id", tt.clientID)
			params.Set("redirect_uri", "https://example.com/callback")
			params.Set("state", "test-state")
			if tt.scope != "" {
				params.Set("scope", tt.scope)
			}

			req := httptest.NewRequest("GET", "/oauth/authorize?"+params.Encode(), nil)
			w := httptest.NewRecorder()

			// Call authorize endpoint
			handler.Authorize(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.checkRedirect {
				// Should redirect to login page
				location := w.Header().Get("Location")
				if !strings.HasPrefix(location, "/auth/login?session_id=") {
					t.Errorf("Expected redirect to login page, got: %s", location)
				}

				// Extract session ID and verify session was created
				sessionID := strings.TrimPrefix(location, "/auth/login?session_id=")
				session, err := sessionRepo.FindBySessionID(ctx, sessionID)
				if err != nil {
					t.Errorf("Failed to find session: %v", err)
					return
				}

				// Verify scope was stored correctly
				expectedScope := tt.scope
				if expectedScope == "" {
					expectedScope = utils.GetDefaultScope()
				}
				if session.Scope != expectedScope {
					t.Errorf("Expected scope '%s', got '%s'", expectedScope, session.Scope)
				}
			} else {
				// Should return error
				var errorResp models.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Fatalf("Failed to parse error response: %v", err)
				}

				if errorResp.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorResp.Error)
				}
			}
		})
	}
}

// TestClaimFiltering tests ID token and UserInfo claim filtering based on scopes
func TestClaimFiltering(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_claim_filtering")
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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create test user
	testUser := &models.User{
		ID:        "test-user-claims",
		Email:     "claims@example.com",
		Name:      "Claims Test User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "test-client-claims",
		ClientSecret:  "test-secret-claims",
		RedirectURIs:  []string{"https://example.com/callback"},
		Name:          "Test Client Claims",
		AllowedScopes: []string{"openid", "profile", "email"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	tests := []struct {
		name           string
		scope          string
		expectedClaims []string
		notExpected    []string
	}{
		{
			name:           "ID token with openid only returns sub",
			scope:          "openid",
			expectedClaims: []string{"sub", "iss", "aud", "exp", "iat"},
			notExpected:    []string{"email", "name", "email_verified"},
		},
		{
			name:           "ID token with profile includes name",
			scope:          "openid profile",
			expectedClaims: []string{"sub", "name", "iss", "aud", "exp", "iat"},
			notExpected:    []string{"email", "email_verified"},
		},
		{
			name:           "ID token with email includes email",
			scope:          "openid email",
			expectedClaims: []string{"sub", "email", "email_verified", "iss", "aud", "exp", "iat"},
			notExpected:    []string{"name"},
		},
		{
			name:           "ID token with all scopes includes all claims",
			scope:          "openid profile email",
			expectedClaims: []string{"sub", "name", "email", "email_verified", "iss", "aud", "exp", "iat"},
			notExpected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create authorization code
			code, _ := utils.GenerateRandomString(16)
			authCode := &models.AuthorizationCode{
				Code:        code,
				ClientID:    testClient.ClientID,
				UserID:      testUser.ID,
				RedirectURI: "https://example.com/callback",
				Scope:       tt.scope,
				Nonce:       "test-nonce",
				ExpiresAt:   time.Now().Add(10 * time.Minute),
				CreatedAt:   time.Now(),
			}
			if err := authCodeRepo.Create(ctx, authCode); err != nil {
				t.Fatalf("Failed to create auth code: %v", err)
			}

			// Exchange code for tokens
			form := url.Values{}
			form.Set("grant_type", "authorization_code")
			form.Set("code", code)
			form.Set("client_id", testClient.ClientID)
			form.Set("client_secret", testClient.ClientSecret)
			form.Set("redirect_uri", "https://example.com/callback")

			req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.Token(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Token exchange failed with status %d: %s", w.Code, w.Body.String())
			}

			var tokenResp models.TokenResponse
			if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
				t.Fatalf("Failed to parse token response: %v", err)
			}

			// Verify ID token claims
			if tokenResp.IDToken == "" {
				t.Fatal("ID token not returned")
			}

			// Parse ID token
			idTokenClaims, err := parseIDToken(tokenResp.IDToken, publicKey)
			if err != nil {
				t.Fatalf("Failed to validate ID token: %v", err)
			}

			// Check expected claims are present
			for _, claim := range tt.expectedClaims {
				if _, exists := idTokenClaims[claim]; !exists {
					t.Errorf("Expected claim '%s' not found in ID token: %v", claim, idTokenClaims)
				}
			}

			// Check unexpected claims are not present
			for _, claim := range tt.notExpected {
				if _, exists := idTokenClaims[claim]; exists {
					t.Errorf("Unexpected claim '%s' found in ID token: %v", claim, idTokenClaims)
				}
			}

			// Verify nonce is included
			if nonce, ok := idTokenClaims["nonce"].(string); !ok || nonce != "test-nonce" {
				t.Errorf("Expected nonce='test-nonce', got '%v'", idTokenClaims["nonce"])
			}
		})
	}
}

// TestUserInfoClaimFiltering tests UserInfo endpoint claim filtering
func TestUserInfoClaimFiltering(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_userinfo_filtering")
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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create test user
	testUser := &models.User{
		ID:        "test-user-userinfo",
		Email:     "userinfo@example.com",
		Name:      "UserInfo Test User",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		scope          string
		expectedClaims []string
		notExpected    []string
	}{
		{
			name:           "UserInfo with openid only returns sub",
			scope:          "openid",
			expectedClaims: []string{"sub"},
			notExpected:    []string{"email", "name", "email_verified"},
		},
		{
			name:           "UserInfo with profile includes name",
			scope:          "openid profile",
			expectedClaims: []string{"sub", "name"},
			notExpected:    []string{"email", "email_verified"},
		},
		{
			name:           "UserInfo with email includes email",
			scope:          "openid email",
			expectedClaims: []string{"sub", "email", "email_verified"},
			notExpected:    []string{"name"},
		},
		{
			name:           "UserInfo with all scopes includes all claims",
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

			// Call UserInfo endpoint
			req := httptest.NewRequest("GET", "/oauth/userinfo", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			w := httptest.NewRecorder()

			handler.UserInfo(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("UserInfo request failed with status %d: %s", w.Code, w.Body.String())
			}

			var userInfo map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &userInfo); err != nil {
				t.Fatalf("Failed to parse UserInfo response: %v", err)
			}

			// Check expected claims are present
			for _, claim := range tt.expectedClaims {
				if _, exists := userInfo[claim]; !exists {
					t.Errorf("Expected claim '%s' not found in UserInfo: %v", claim, userInfo)
				}
			}

			// Check unexpected claims are not present
			for _, claim := range tt.notExpected {
				if _, exists := userInfo[claim]; exists {
					t.Errorf("Unexpected claim '%s' found in UserInfo: %v", claim, userInfo)
				}
			}
		})
	}
}

// TestScopeDowngrade tests scope downgrade in refresh token flow
func TestScopeDowngrade(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_scope_downgrade")
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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create test user with explicit string ID
	userID, _ := utils.GenerateRandomString(32)
	testUser := &models.User{
		ID:        userID,
		Email:     "downgrade@example.com",
		Name:      "Downgrade Test User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Verify the user was created with our ID
	createdUser, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find created user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "test-client-downgrade",
		ClientSecret:  "test-secret-downgrade",
		RedirectURIs:  []string{"https://example.com/callback"},
		Name:          "Test Client Downgrade",
		AllowedScopes: []string{"openid", "profile", "email", "phone"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Original scope for the initial token
	originalScope := "openid profile email phone"

	// Create an authorization code and exchange it for tokens to get a valid refresh token
	code, _ := utils.GenerateRandomString(16)
	authCode := &models.AuthorizationCode{
		Code:        code,
		ClientID:    testClient.ClientID,
		UserID:      createdUser.ID, // Use the ID from the created user
		RedirectURI: "https://example.com/callback",
		Scope:       originalScope,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}
	if err := authCodeRepo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// Exchange code for tokens
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", testClient.ClientID)
	form.Set("client_secret", testClient.ClientSecret)
	form.Set("redirect_uri", "https://example.com/callback")

	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.Token(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Initial token exchange failed: %s", w.Body.String())
	}

	var initialTokenResp models.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &initialTokenResp); err != nil {
		t.Fatalf("Failed to parse initial token response: %v", err)
	}

	refreshToken := initialTokenResp.RefreshToken

	tests := []struct {
		name           string
		requestedScope string
		expectedStatus int
		expectedError  string
		shouldSucceed  bool
	}{
		{
			name:           "Refresh with same scopes succeeds",
			requestedScope: "openid profile email phone",
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name:           "Refresh with reduced scopes succeeds",
			requestedScope: "openid profile",
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name:           "Refresh with single scope succeeds",
			requestedScope: "openid",
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
		{
			name:           "Refresh with increased scopes fails",
			requestedScope: "openid profile email phone address",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_scope",
			shouldSucceed:  false,
		},
		{
			name:           "Refresh with new scope fails",
			requestedScope: "openid address",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_scope",
			shouldSucceed:  false,
		},
		{
			name:           "Refresh without scope parameter uses original scopes",
			requestedScope: "",
			expectedStatus: http.StatusOK,
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build refresh token request
			form := url.Values{}
			form.Set("grant_type", "refresh_token")
			form.Set("refresh_token", refreshToken)
			form.Set("client_id", testClient.ClientID)
			form.Set("client_secret", testClient.ClientSecret)
			if tt.requestedScope != "" {
				form.Set("scope", tt.requestedScope)
			}

			req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.Token(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.shouldSucceed {
				// Parse successful response
				var tokenResp models.TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
					t.Fatalf("Failed to parse token response: %v", err)
				}

				// Verify access token was issued
				if tokenResp.AccessToken == "" {
					t.Error("Access token not returned")
				}

				// Verify new refresh token was issued
				if tokenResp.RefreshToken == "" {
					t.Error("Refresh token not returned")
				}

				// Verify scope in response
				expectedScope := tt.requestedScope
				if expectedScope == "" {
					expectedScope = originalScope
				}
				if tokenResp.Scope != expectedScope {
					t.Errorf("Expected scope '%s', got '%s'", expectedScope, tokenResp.Scope)
				}

				// Validate the new access token contains correct scope
				claims, err := parseAccessToken(tokenResp.AccessToken, publicKey)
				if err != nil {
					t.Fatalf("Failed to validate access token: %v", err)
				}

				if claims.Scope != expectedScope {
					t.Errorf("Access token scope mismatch. Expected '%s', got '%s'", expectedScope, claims.Scope)
				}
			} else {
				// Parse error response
				var errorResp models.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Fatalf("Failed to parse error response: %v", err)
				}

				if errorResp.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorResp.Error)
				}
			}
		})
	}
}

// TestScopeDowngradeWithMultipleRefreshes tests multiple refresh operations with scope changes
func TestScopeDowngradeWithMultipleRefreshes(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("oauth2_test_multiple_refresh")
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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, consentRepo, cfg)

	// Create test user with explicit string ID
	userID, _ := utils.GenerateRandomString(32)
	testUser := &models.User{
		ID:        userID,
		Email:     "multirefresh@example.com",
		Name:      "Multi Refresh Test User",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Verify the user was created with our ID
	createdUser, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find created user: %v", err)
	}

	// Create test client
	testClient := &models.Client{
		ClientID:      "test-client-multi-refresh",
		ClientSecret:  "test-secret-multi-refresh",
		RedirectURIs:  []string{"https://example.com/callback"},
		Name:          "Test Client Multi Refresh",
		AllowedScopes: []string{"openid", "profile", "email", "phone"},
		CreatedAt:     time.Now(),
	}
	if err := clientRepo.Create(ctx, testClient); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Start with full scope - create initial tokens via authorization code
	currentScope := "openid profile email phone"
	
	code, _ := utils.GenerateRandomString(16)
	authCode := &models.AuthorizationCode{
		Code:        code,
		ClientID:    testClient.ClientID,
		UserID:      createdUser.ID, // Use the ID from the created user
		RedirectURI: "https://example.com/callback",
		Scope:       currentScope,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}
	if err := authCodeRepo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// Exchange code for tokens
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", testClient.ClientID)
	form.Set("client_secret", testClient.ClientSecret)
	form.Set("redirect_uri", "https://example.com/callback")

	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.Token(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Initial token exchange failed: %s", w.Body.String())
	}

	var initialTokenResp models.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &initialTokenResp); err != nil {
		t.Fatalf("Failed to parse initial token response: %v", err)
	}

	refreshToken := initialTokenResp.RefreshToken

	// First refresh: downgrade to openid profile email
	form = url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", testClient.ClientID)
	form.Set("client_secret", testClient.ClientSecret)
	form.Set("scope", "openid profile email")

	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	handler.Token(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("First refresh failed: %s", w.Body.String())
	}

	var tokenResp1 models.TokenResponse
	json.Unmarshal(w.Body.Bytes(), &tokenResp1)

	// Verify scope was downgraded
	if tokenResp1.Scope != "openid profile email" {
		t.Errorf("Expected scope 'openid profile email', got '%s'", tokenResp1.Scope)
	}

	// Second refresh: try to escalate back to phone (should fail)
	form = url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokenResp1.RefreshToken)
	form.Set("client_id", testClient.ClientID)
	form.Set("client_secret", testClient.ClientSecret)
	form.Set("scope", "openid profile email phone")

	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	handler.Token(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected scope escalation to fail, got status %d", w.Code)
	}

	var errorResp models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	if errorResp.Error != "invalid_scope" {
		t.Errorf("Expected error 'invalid_scope', got '%s'", errorResp.Error)
	}

	// Third refresh: further downgrade to openid only (should succeed)
	form = url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokenResp1.RefreshToken)
	form.Set("client_id", testClient.ClientID)
	form.Set("client_secret", testClient.ClientSecret)
	form.Set("scope", "openid")

	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	handler.Token(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Third refresh failed: %s", w.Body.String())
	}

	var tokenResp2 models.TokenResponse
	json.Unmarshal(w.Body.Bytes(), &tokenResp2)

	if tokenResp2.Scope != "openid" {
		t.Errorf("Expected scope 'openid', got '%s'", tokenResp2.Scope)
	}
}
