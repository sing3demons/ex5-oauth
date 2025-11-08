package utils

import (
	"testing"
	"time"
)

func TestEncryptDecryptJWE(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	testData := map[string]interface{}{
		"user_id": "user123",
		"email":   "test@example.com",
		"name":    "Test User",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	jweToken, err := EncryptJWE(testData, publicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt JWE: %v", err)
	}
	if jweToken == "" {
		t.Error("Expected non-empty JWE token")
	}

	var decrypted map[string]interface{}
	err = DecryptJWE(jweToken, privateKey, &decrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt JWE: %v", err)
	}

	if decrypted["user_id"] != "user123" {
		t.Errorf("Expected user_id 'user123', got '%v'", decrypted["user_id"])
	}
}

func TestGenerateJWEAccessToken(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	expiry := time.Now().Add(time.Hour).Unix()
	token, err := GenerateJWEAccessToken(
		"user123",
		"test@example.com",
		"Test User",
		"openid profile email",
		publicKey,
		expiry,
	)
	if err != nil {
		t.Fatalf("Failed to generate JWE access token: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}

	claims, err := ValidateJWE(token, privateKey)
	if err != nil {
		t.Fatalf("Failed to validate JWE token: %v", err)
	}
	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}
}

func TestGenerateJWERefreshToken(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	expiry := time.Now().Add(24 * time.Hour).Unix()
	token, err := GenerateJWERefreshToken("user123", publicKey, expiry)
	if err != nil {
		t.Fatalf("Failed to generate JWE refresh token: %v", err)
	}

	claims, err := ValidateJWE(token, privateKey)
	if err != nil {
		t.Fatalf("Failed to validate JWE token: %v", err)
	}
	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", claims.UserID)
	}
}

func TestGenerateJWEIDToken(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	userClaims := map[string]interface{}{
		"email":          "test@example.com",
		"email_verified": true,
		"name":           "Test User",
	}

	expiry := time.Now().Add(time.Hour).Unix()
	token, err := GenerateJWEIDToken(
		"user123",
		"client456",
		userClaims,
		publicKey,
		expiry,
	)
	if err != nil {
		t.Fatalf("Failed to generate JWE ID token: %v", err)
	}

	claims, err := ValidateJWE(token, privateKey)
	if err != nil {
		t.Fatalf("Failed to validate JWE token: %v", err)
	}
	if claims.Aud != "client456" {
		t.Errorf("Expected Aud 'client456', got '%s'", claims.Aud)
	}
}

func TestValidateJWEExpired(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	publicKey := &privateKey.PublicKey

	expiry := time.Now().Add(-time.Hour).Unix()
	token, err := GenerateJWEAccessToken(
		"user123",
		"test@example.com",
		"Test User",
		"openid",
		publicKey,
		expiry,
	)
	if err != nil {
		t.Fatalf("Failed to generate JWE token: %v", err)
	}

	_, err = ValidateJWE(token, privateKey)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestDecryptJWEWithWrongKey(t *testing.T) {
	privateKey1, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	privateKey2, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	publicKey1 := &privateKey1.PublicKey
	expiry := time.Now().Add(time.Hour).Unix()

	token, err := GenerateJWEAccessToken(
		"user123",
		"test@example.com",
		"Test User",
		"openid",
		publicKey1,
		expiry,
	)
	if err != nil {
		t.Fatalf("Failed to generate JWE token: %v", err)
	}

	_, err = ValidateJWE(token, privateKey2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key")
	}
}

func TestIsJWEAndIsJWT(t *testing.T) {
	jweToken := "header.encrypted_key.iv.ciphertext."
	if !IsJWE(jweToken) {
		t.Error("Expected IsJWE to return true for JWE format")
	}
	if IsJWT(jweToken) {
		t.Error("Expected IsJWT to return false for JWE format")
	}

	jwtToken := "header.payload.signature"
	if !IsJWT(jwtToken) {
		t.Error("Expected IsJWT to return true for JWT format")
	}
	if IsJWE(jwtToken) {
		t.Error("Expected IsJWE to return false for JWT format")
	}
}

func TestDecryptJWEInvalidToken(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	var result map[string]interface{}
	err = DecryptJWE("invalid.jwe.token", privateKey, &result)
	if err == nil {
		t.Error("Expected error when decrypting invalid token")
	}
}
