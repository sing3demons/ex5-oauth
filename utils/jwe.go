package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// JWEClaims represents generic claims in a JWE token
type JWEClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Scope  string `json:"scope,omitempty"`
	Aud    string `json:"aud,omitempty"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// JWEAccessTokenClaims for access tokens
type JWEAccessTokenClaims struct {
	UserID   string `json:"sub"`
	Scope    string `json:"scope"`
	ClientID string `json:"client_id,omitempty"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// JWEIDTokenClaims for ID tokens
type JWEIDTokenClaims struct {
	UserID        string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	Aud           string `json:"aud"`
	Iss           string `json:"iss"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
}

// JWERefreshTokenClaims for refresh tokens
type JWERefreshTokenClaims struct {
	UserID string `json:"sub"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// EncryptJWE encrypts data into JWE format using RSA-OAEP + AES-256-GCM
func EncryptJWE(data interface{}, publicKey *rsa.PublicKey) (string, error) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Generate random AES key (256-bit)
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt AES key with RSA-OAEP
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, nonce, jsonData, nil)

	// JWE Compact Serialization: header.encryptedKey.iv.ciphertext.tag
	// For simplicity, we combine ciphertext and tag (GCM already includes tag)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RSA-OAEP","enc":"A256GCM"}`))
	encKey := base64.RawURLEncoding.EncodeToString(encryptedKey)
	iv := base64.RawURLEncoding.EncodeToString(nonce)
	ct := base64.RawURLEncoding.EncodeToString(ciphertext)

	return fmt.Sprintf("%s.%s.%s.%s.", header, encKey, iv, ct), nil
}

// DecryptJWE decrypts JWE token using RSA private key
func DecryptJWE(jweToken string, privateKey *rsa.PrivateKey, target interface{}) error {
	parts := strings.Split(jweToken, ".")
	if len(parts) != 5 {
		return errors.New("invalid JWE format")
	}

	// Decode encrypted key
	encryptedKey, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	// Decrypt AES key with RSA-OAEP
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Decode IV (nonce)
	nonce, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Decode ciphertext
	ciphertext, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Unmarshal into target
	if err := json.Unmarshal(plaintext, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// GenerateJWEAccessToken creates an encrypted access token
func GenerateJWEAccessToken(userID, email, name, scope string, publicKey *rsa.PublicKey, expiry int64) (string, error) {
	claims := JWEAccessTokenClaims{
		UserID: userID,
		Scope:  scope,
		Exp:    expiry,
		Iat:    time.Now().Unix(),
	}
	return EncryptJWE(claims, publicKey)
}

// GenerateJWERefreshToken creates an encrypted refresh token
func GenerateJWERefreshToken(userID string, publicKey *rsa.PublicKey, expiry int64) (string, error) {
	claims := JWERefreshTokenClaims{
		UserID: userID,
		Exp:    expiry,
		Iat:    time.Now().Unix(),
	}
	return EncryptJWE(claims, publicKey)
}

// GenerateJWEIDToken creates an encrypted ID token with filtered claims
func GenerateJWEIDToken(userID, clientID string, userClaims map[string]interface{}, publicKey *rsa.PublicKey, expiry int64) (string, error) {
	// Start with user claims (already filtered by scope)
	claims := make(map[string]interface{})
	
	// Add all user claims
	for key, value := range userClaims {
		claims[key] = value
	}
	
	// Add standard JWT claims
	claims["sub"] = userID
	claims["aud"] = clientID
	claims["iss"] = "oauth2-server"
	claims["exp"] = expiry
	claims["iat"] = time.Now().Unix()
	
	return EncryptJWE(claims, publicKey)
}

// GenerateJWEIDTokenLegacy creates an encrypted ID token with explicit claims (deprecated)
func GenerateJWEIDTokenLegacy(userID, email, name, clientID string, publicKey *rsa.PublicKey, expiry int64) (string, error) {
	claims := JWEIDTokenClaims{
		UserID:        userID,
		Email:         email,
		EmailVerified: true,
		Name:          name,
		Aud:           clientID,
		Iss:           "oauth2-server",
		Exp:           expiry,
		Iat:           time.Now().Unix(),
	}
	return EncryptJWE(claims, publicKey)
}

// ValidateJWE validates and decrypts a JWE token
func ValidateJWE(jweToken string, privateKey *rsa.PrivateKey) (*JWEClaims, error) {
	var claims JWEClaims
	if err := DecryptJWE(jweToken, privateKey, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

// IsJWE checks if a token is in JWE format (5 parts)
func IsJWE(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 5
}

// IsJWT checks if a token is in JWT format (3 parts)
func IsJWT(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3
}
