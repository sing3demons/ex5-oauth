package logger

import (
	"encoding/json"
	"testing"
)

func TestLogger_AddSuccess(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// Add first value
	logger.AddSuccess("userId", "user1")

	// Add second value with same key - should create array
	logger.AddSuccess("userId", "user2")

	// Add third value - should append to array
	logger.AddSuccess("userId", "user3")

	// Add different key
	logger.AddSuccess("action", "login")

	output := captureOutput(func() {
		logger.Flush(200)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	// Check userId is an array
	userIds, ok := log.Metadata["userId"].([]interface{})
	if !ok {
		t.Fatalf("Expected userId to be array, got %T", log.Metadata["userId"])
	}

	if len(userIds) != 3 {
		t.Errorf("Expected userId array length 3, got %d", len(userIds))
	}

	if userIds[0] != "user1" {
		t.Errorf("Expected first userId to be 'user1', got '%v'", userIds[0])
	}

	if userIds[1] != "user2" {
		t.Errorf("Expected second userId to be 'user2', got '%v'", userIds[1])
	}

	if userIds[2] != "user3" {
		t.Errorf("Expected third userId to be 'user3', got '%v'", userIds[2])
	}

	// Check action is still a single value
	if log.Metadata["action"] != "login" {
		t.Errorf("Expected action to be 'login', got '%v'", log.Metadata["action"])
	}
}

func TestLogger_AddSuccessMixedTypes(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// Add different types with same key
	logger.AddSuccess("data", "string value")
	logger.AddSuccess("data", 123)
	logger.AddSuccess("data", true)
	logger.AddSuccess("data", map[string]interface{}{"key": "value"})

	output := captureOutput(func() {
		logger.Flush(200)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	dataArray, ok := log.Metadata["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data to be array, got %T", log.Metadata["data"])
	}

	if len(dataArray) != 4 {
		t.Errorf("Expected data array length 4, got %d", len(dataArray))
	}

	if dataArray[0] != "string value" {
		t.Errorf("Expected first data to be 'string value', got '%v'", dataArray[0])
	}

	if dataArray[1] != float64(123) {
		t.Errorf("Expected second data to be 123, got '%v'", dataArray[1])
	}

	if dataArray[2] != true {
		t.Errorf("Expected third data to be true, got '%v'", dataArray[2])
	}
}

func TestLogger_AddSuccessSingleValue(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// Add only one value - should not be array
	logger.AddSuccess("status", "completed")

	output := captureOutput(func() {
		logger.Flush(200)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	// Should be a single value, not array
	if log.Metadata["status"] != "completed" {
		t.Errorf("Expected status to be 'completed', got '%v'", log.Metadata["status"])
	}

	// Ensure it's not an array
	if _, isArray := log.Metadata["status"].([]interface{}); isArray {
		t.Error("Expected status to be single value, not array")
	}
}

func TestLogger_AddSuccessWithAddMetadata(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// Mix AddSuccess and AddMetadata
	logger.AddSuccess("userId", "user1")
	logger.AddMetadata("requestId", "req123")
	logger.AddSuccess("userId", "user2")
	logger.AddMetadata("method", "POST")

	output := captureOutput(func() {
		logger.Flush(200)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	// Check userId is array
	userIds, ok := log.Metadata["userId"].([]interface{})
	if !ok {
		t.Fatalf("Expected userId to be array, got %T", log.Metadata["userId"])
	}

	if len(userIds) != 2 {
		t.Errorf("Expected userId array length 2, got %d", len(userIds))
	}

	// Check other metadata
	if log.Metadata["requestId"] != "req123" {
		t.Errorf("Expected requestId to be 'req123', got '%v'", log.Metadata["requestId"])
	}

	if log.Metadata["method"] != "POST" {
		t.Errorf("Expected method to be 'POST', got '%v'", log.Metadata["method"])
	}
}

func TestLogger_AddSuccessOverwriteWithAddMetadata(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// AddSuccess creates array
	logger.AddSuccess("key", "value1")
	logger.AddSuccess("key", "value2")

	// AddMetadata overwrites the array
	logger.AddMetadata("key", "single value")

	output := captureOutput(func() {
		logger.Flush(200)
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	// Should be overwritten to single value
	if log.Metadata["key"] != "single value" {
		t.Errorf("Expected key to be 'single value', got '%v'", log.Metadata["key"])
	}

	// Ensure it's not an array
	if _, isArray := log.Metadata["key"].([]interface{}); isArray {
		t.Error("Expected key to be single value after AddMetadata, not array")
	}
}
