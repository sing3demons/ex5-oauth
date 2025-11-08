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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, cfg)

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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, cfg)

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

	handler := NewOAuthHandler(userRepo, clientRepo, authCodeRepo, sessionRepo, cfg)

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
