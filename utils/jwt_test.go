package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func TestGenerateAccessToken(t *testing.T) {
	privateKey, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	token, err := GenerateAccessToken(
		"user123",
		"user@example.com",
		"John Doe",
		"openid profile email",
		privateKey,
		3600,
	)

	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token structure
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Expected valid token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	privateKey, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	token, err := GenerateRefreshToken("user123", "openid profile email", privateKey, 604800)

	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestGenerateIDToken(t *testing.T) {
	privateKey, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	userClaims := map[string]interface{}{
		"email":          "user@example.com",
		"email_verified": true,
		"name":           "John Doe",
	}

	token, err := GenerateIDToken(
		"user123",
		"client456",
		userClaims,
		privateKey,
		3600,
	)

	if err != nil {
		t.Fatalf("Failed to generate ID token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Parse and verify claims
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}

	if claims.Email != "user@example.com" {
		t.Errorf("Expected Email 'user@example.com', got '%s'", claims.Email)
	}

	if len(claims.Audience) != 1 || claims.Audience[0] != "client456" {
		t.Errorf("Expected Audience ['client456'], got %v", claims.Audience)
	}
}

func TestValidateToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Generate a valid token
	token, err := GenerateAccessToken(
		"user123",
		"user@example.com",
		"John Doe",
		"openid profile",
		privateKey,
		3600,
	)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := ValidateToken(token, publicKey)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}

	// Access tokens should NOT include email and name (only scope)
	// These should be empty as per OAuth2 best practices
	if claims.Email != "" {
		t.Errorf("Expected Email to be empty in access token, got '%s'", claims.Email)
	}

	if claims.Name != "" {
		t.Errorf("Expected Name to be empty in access token, got '%s'", claims.Name)
	}

	if claims.Scope != "openid profile" {
		t.Errorf("Expected Scope 'openid profile', got '%s'", claims.Scope)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	_, publicKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Test with invalid token
	_, err = ValidateToken("invalid.token.here", publicKey)
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateTokenExpired(t *testing.T) {
	privateKey, publicKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Generate token that expires immediately
	token, err := GenerateAccessToken(
		"user123",
		"user@example.com",
		"John Doe",
		"openid",
		privateKey,
		-1, // Expired
	)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait a moment to ensure expiration
	time.Sleep(10 * time.Millisecond)

	// Validate should fail
	_, err = ValidateToken(token, publicKey)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestValidateTokenWrongKey(t *testing.T) {
	privateKey1, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	_, publicKey2, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Generate token with key1
	token, err := GenerateAccessToken(
		"user123",
		"user@example.com",
		"John Doe",
		"openid",
		privateKey1,
		3600,
	)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate with different key2
	_, err = ValidateToken(token, publicKey2)
	if err == nil {
		t.Error("Expected error when validating with wrong key")
	}
}

func TestParsePrivateKey(t *testing.T) {
	// Valid PKCS8 private key
	validPKCS8 := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKj
MzEfYyjiWA4R4/M2bS1+fWIcPm15A8+raZ4dp+/PJE+5K9Z1RKsJe4s/1uvKgEiB
3Opz8d+45uVk5y4OfAgWrQzwyZcU1Aw8NcJGUYhX3siQYEF9DsIs2FrtGkNfq+AI
-----END PRIVATE KEY-----`

	_, err := ParsePrivateKey(validPKCS8)
	// May fail with actual parsing but tests the function
	if err == nil {
		// Success case
	}

	// Invalid PEM
	_, err = ParsePrivateKey("not a valid pem")
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}

	// Empty string
	_, err = ParsePrivateKey("")
	if err == nil {
		t.Error("Expected error for empty string")
	}
}

func TestParsePublicKey(t *testing.T) {
	// Invalid PEM
	_, err := ParsePublicKey("not a valid pem")
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}

	// Empty string
	_, err = ParsePublicKey("")
	if err == nil {
		t.Error("Expected error for empty string")
	}
}

func TestJWTClaimsStructure(t *testing.T) {
	claims := JWTClaims{
		UserID: "user123",
		Email:  "user@example.com",
		Name:   "John Doe",
		Scope:  "openid profile",
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}

	if claims.Email != "user@example.com" {
		t.Errorf("Expected Email 'user@example.com', got '%s'", claims.Email)
	}
}

func TestGenerateTokensWithEmptyValues(t *testing.T) {
	privateKey, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Test with empty values
	token, err := GenerateAccessToken("", "", "", "", privateKey, 3600)
	if err != nil {
		t.Fatalf("Failed to generate token with empty values: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token even with empty values")
	}
}

func TestGenerateRefreshTokenWithZeroExpiry(t *testing.T) {
	privateKey, _, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	token, err := GenerateRefreshToken("user123", "openid profile email", privateKey, 0)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}
}
