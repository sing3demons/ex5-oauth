package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecretPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}

	// Hash same password again - should be different (bcrypt uses salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password second time: %v", err)
	}

	if hash == hash2 {
		t.Error("Two hashes of same password should be different (due to salt)")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mySecretPassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Check correct password
	if !CheckPasswordHash(password, hash) {
		t.Error("Expected password to match hash")
	}

	// Check incorrect password
	if CheckPasswordHash("wrongPassword", hash) {
		t.Error("Expected wrong password to not match hash")
	}

	// Check empty password
	if CheckPasswordHash("", hash) {
		t.Error("Expected empty password to not match hash")
	}
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	// Test with invalid hash
	if CheckPasswordHash("password", "invalid_hash") {
		t.Error("Expected invalid hash to fail")
	}

	// Test with empty hash
	if CheckPasswordHash("password", "") {
		t.Error("Expected empty hash to fail")
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 32

	str, err := GenerateRandomString(length)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}

	if len(str) != length {
		t.Errorf("Expected string length %d, got %d", length, len(str))
	}

	// Generate another string - should be different
	str2, err := GenerateRandomString(length)
	if err != nil {
		t.Fatalf("Failed to generate second random string: %v", err)
	}

	if str == str2 {
		t.Error("Two random strings should be different")
	}
}

func TestGenerateRandomStringDifferentLengths(t *testing.T) {
	lengths := []int{8, 16, 32, 64, 128}

	for _, length := range lengths {
		str, err := GenerateRandomString(length)
		if err != nil {
			t.Fatalf("Failed to generate random string of length %d: %v", length, err)
		}

		if len(str) != length {
			t.Errorf("Expected string length %d, got %d", length, len(str))
		}
	}
}

func TestGenerateRandomStringZeroLength(t *testing.T) {
	str, err := GenerateRandomString(0)
	if err != nil {
		t.Fatalf("Failed to generate zero-length string: %v", err)
	}

	if len(str) != 0 {
		t.Errorf("Expected empty string, got length %d", len(str))
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("Failed to hash empty password: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash even for empty password")
	}

	// Should be able to verify empty password
	if !CheckPasswordHash("", hash) {
		t.Error("Expected empty password to match its hash")
	}
}

func TestHashPasswordLong(t *testing.T) {
	// Test with long password (bcrypt max is 72 bytes)
	longPassword := string(make([]byte, 72))
	for i := range longPassword {
		longPassword = string(append([]byte(longPassword[:i]), 'a'))
	}

	hash, err := HashPassword(longPassword)
	if err != nil {
		t.Fatalf("Failed to hash long password: %v", err)
	}

	if !CheckPasswordHash(longPassword, hash) {
		t.Error("Expected long password to match hash")
	}
}

func TestHashPasswordTooLong(t *testing.T) {
	// Test with password exceeding bcrypt limit (>72 bytes)
	tooLongPassword := string(make([]byte, 100))
	for i := 0; i < 100; i++ {
		tooLongPassword = string(append([]byte(tooLongPassword[:i]), 'a'))
	}

	_, err := HashPassword(tooLongPassword)
	if err == nil {
		t.Error("Expected error for password exceeding 72 bytes")
	}
}

func TestGenerateRandomStringUniqueness(t *testing.T) {
	// Generate multiple strings and check uniqueness
	generated := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		str, err := GenerateRandomString(32)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}

		if generated[str] {
			t.Error("Generated duplicate random string")
		}

		generated[str] = true
	}

	if len(generated) != count {
		t.Errorf("Expected %d unique strings, got %d", count, len(generated))
	}
}
