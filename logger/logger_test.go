package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestLogger_InfoDetail(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")

	actionInfo := ActionInfo{
		Action:            "user_login",
		ActionDescription: "User attempting to login",
		SubAction:         "validate_credentials",
	}

	data := map[string]interface{}{
		"username": "john",
		"password": "secret123",
		"email":    "john@example.com",
	}

	maskingRules := []MaskingRule{
		{Field: "password", Type: MaskingTypeFull},
		{Field: "email", Type: MaskingTypeEmail},
	}

	output := captureOutput(func() {
		logger.InfoDetail(actionInfo, data, maskingRules...)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Service != "test-service" {
		t.Errorf("Expected service 'test-service', got '%s'", log.Service)
	}

	if log.Type != TypeDetail {
		t.Errorf("Expected type 'detail', got '%s'", log.Type)
	}

	if log.Action != "user_login" {
		t.Errorf("Expected action 'user_login', got '%s'", log.Action)
	}

	if log.ActionDescription != "User attempting to login" {
		t.Errorf("Expected actionDescription 'User attempting to login', got '%s'", log.ActionDescription)
	}

	if log.SubAction != "validate_credentials" {
		t.Errorf("Expected subAction 'validate_credentials', got '%s'", log.SubAction)
	}

	maskedData, ok := log.Metadata["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data in metadata")
	}

	if maskedData["password"] != "***" {
		t.Errorf("Expected password to be masked as '***', got '%v'", maskedData["password"])
	}

	if maskedData["email"] != "j***@example.com" {
		t.Errorf("Expected email to be masked as 'j***@example.com', got '%v'", maskedData["email"])
	}

	if maskedData["username"] != "john" {
		t.Errorf("Expected username to remain 'john', got '%v'", maskedData["username"])
	}
}

func TestLogger_ErrorDetail(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")

	actionInfo := ActionInfo{
		Action:            "payment_failed",
		ActionDescription: "Payment processing failed",
	}

	data := map[string]interface{}{
		"cardNumber": "1234-5678-9012-3456",
		"amount":     100.50,
		"error":      "Insufficient funds",
	}

	maskingRules := []MaskingRule{
		{Field: "cardNumber", Type: MaskingTypeCard},
	}

	output := captureOutput(func() {
		logger.ErrorDetail(actionInfo, data, maskingRules...)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Level != LevelError {
		t.Errorf("Expected level 'error', got '%s'", log.Level)
	}

	if log.Type != TypeDetail {
		t.Errorf("Expected type 'detail', got '%s'", log.Type)
	}

	maskedData, ok := log.Metadata["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data in metadata")
	}

	if maskedData["cardNumber"] != "****-****-****-3456" {
		t.Errorf("Expected cardNumber to be masked, got '%v'", maskedData["cardNumber"])
	}
}

// TestLogger_InfoSummary removed - use Flush() instead

func TestLogger_InfoDetailNoMasking(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")

	actionInfo := ActionInfo{
		Action:            "user_view",
		ActionDescription: "User viewing profile",
	}

	data := map[string]interface{}{
		"username": "john",
		"role":     "admin",
	}

	output := captureOutput(func() {
		logger.InfoDetail(actionInfo, data)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	maskedData, ok := log.Metadata["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data in metadata")
	}

	if maskedData["username"] != "john" {
		t.Errorf("Expected username to remain 'john', got '%v'", maskedData["username"])
	}

	if maskedData["role"] != "admin" {
		t.Errorf("Expected role to remain 'admin', got '%v'", maskedData["role"])
	}
}

func TestLogger_SimpleInfo(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")

	output := captureOutput(func() {
		logger.Info("test_action", "Test message",
			WithTransactionID("txn123"),
			WithSessionID("sess456"),
		)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Action != "test_action" {
		t.Errorf("Expected action 'test_action', got '%s'", log.Action)
	}

	if log.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", log.Message)
	}

	if log.TransactionID != "txn123" {
		t.Errorf("Expected transactionId 'txn123', got '%s'", log.TransactionID)
	}

	if log.SessionID != "sess456" {
		t.Errorf("Expected sessionId 'sess456', got '%s'", log.SessionID)
	}
}
