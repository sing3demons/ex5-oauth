package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// VerifyPKCE verifies the PKCE code_verifier against the code_challenge
func VerifyPKCE(codeVerifier, codeChallenge, challengeMethod string) bool {
	if challengeMethod == "" || challengeMethod == "plain" {
		// Plain method: verifier must equal challenge
		return codeVerifier == codeChallenge
	}

	if challengeMethod == "S256" {
		// S256 method: SHA256(verifier) must equal challenge
		hash := sha256.Sum256([]byte(codeVerifier))
		computed := base64.RawURLEncoding.EncodeToString(hash[:])
		return computed == codeChallenge
	}

	return false
}

// GenerateCodeVerifier generates a random code_verifier for PKCE
func GenerateCodeVerifier() (string, error) {
	// Generate 32 random bytes (256 bits)
	randomStr, err := GenerateRandomString(32)
	if err != nil {
		return "", err
	}
	
	// Base64 URL encode (without padding)
	return base64.RawURLEncoding.EncodeToString([]byte(randomStr)), nil
}

// GenerateCodeChallenge generates a code_challenge from a code_verifier
func GenerateCodeChallenge(codeVerifier string, method string) string {
	if method == "" || method == "plain" {
		return codeVerifier
	}

	if method == "S256" {
		hash := sha256.Sum256([]byte(codeVerifier))
		return base64.RawURLEncoding.EncodeToString(hash[:])
	}

	return ""
}

// ValidateCodeVerifier checks if a code_verifier meets PKCE requirements
func ValidateCodeVerifier(codeVerifier string) bool {
	// RFC 7636: code_verifier must be 43-128 characters
	// and contain only [A-Z] [a-z] [0-9] - . _ ~
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return false
	}

	for _, c := range codeVerifier {
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~') {
			return false
		}
	}

	return true
}

// NormalizeCodeChallenge removes any padding from base64 URL encoded challenge
func NormalizeCodeChallenge(challenge string) string {
	return strings.TrimRight(challenge, "=")
}
