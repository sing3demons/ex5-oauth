package utils

import (
	"os"
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	if privateKey == nil {
		t.Error("Expected non-nil private key")
	}

	if privateKey.PublicKey.N == nil {
		t.Error("Expected valid public key")
	}
}

func TestSaveAndLoadPrivateKey(t *testing.T) {
	// Generate key
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Save to temp file
	tempFile := "test_private.pem"
	defer os.Remove(tempFile)

	err = SavePrivateKeyToFile(privateKey, tempFile)
	if err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	// Load back
	loadedKey, err := LoadPrivateKeyFromFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	if loadedKey == nil {
		t.Error("Expected non-nil loaded key")
	}

	// Verify keys match
	if privateKey.N.Cmp(loadedKey.N) != 0 {
		t.Error("Loaded key doesn't match original")
	}
}

func TestSaveAndLoadPublicKey(t *testing.T) {
	// Generate key
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Save to temp file
	tempFile := "test_public.pem"
	defer os.Remove(tempFile)

	err = SavePublicKeyToFile(publicKey, tempFile)
	if err != nil {
		t.Fatalf("Failed to save public key: %v", err)
	}

	// Load back
	loadedKey, err := LoadPublicKeyFromFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to load public key: %v", err)
	}

	if loadedKey == nil {
		t.Error("Expected non-nil loaded key")
	}

	// Verify keys match
	if publicKey.N.Cmp(loadedKey.N) != 0 {
		t.Error("Loaded public key doesn't match original")
	}
}

func TestLoadPrivateKeyFromNonExistentFile(t *testing.T) {
	_, err := LoadPrivateKeyFromFile("nonexistent.pem")
	if err == nil {
		t.Error("Expected error when loading from nonexistent file")
	}
}

func TestLoadPublicKeyFromNonExistentFile(t *testing.T) {
	_, err := LoadPublicKeyFromFile("nonexistent.pem")
	if err == nil {
		t.Error("Expected error when loading from nonexistent file")
	}
}

func TestSavePrivateKeyToInvalidPath(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Try to save to invalid path
	err = SavePrivateKeyToFile(privateKey, "/invalid/path/key.pem")
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestSavePublicKeyToInvalidPath(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Try to save to invalid path
	err = SavePublicKeyToFile(publicKey, "/invalid/path/key.pem")
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestGenerateRSAKeyPairSmallSize(t *testing.T) {
	// Test with minimum size
	privateKey, err := GenerateRSAKeyPair(1024)
	if err != nil {
		t.Fatalf("Failed to generate 1024-bit key: %v", err)
	}

	if privateKey == nil {
		t.Error("Expected non-nil private key")
	}
}
