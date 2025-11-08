package logger

import (
	"encoding/json"
	"testing"
)

func TestLogger_Flush(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn123", "sess456")

	// Add some detail logs
	logger.InfoDetail(ActionInfo{
		Action:            "step1",
		ActionDescription: "First step",
	}, map[string]interface{}{"data": "value1"})

	logger.InfoDetail(ActionInfo{
		Action:            "step2",
		ActionDescription: "Second step",
	}, map[string]interface{}{"data": "value2"})

	logger.AddMetadata("totalSteps", 2)
	logger.AddMetadata("result", "success")

	output := captureOutput(func() {
		logger.Flush(200, "Request completed successfully")
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Type != TypeSummary {
		t.Errorf("Expected type 'summary', got '%s'", log.Type)
	}

	if log.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", log.StatusCode)
	}

	if log.Message != "Request completed successfully" {
		t.Errorf("Expected message 'Request completed successfully', got '%s'", log.Message)
	}

	if log.Action != "" {
		t.Errorf("Expected action to be empty for summary, got '%s'", log.Action)
	}

	if log.TransactionID != "txn123" {
		t.Errorf("Expected transactionId 'txn123', got '%s'", log.TransactionID)
	}

	if log.SessionID != "sess456" {
		t.Errorf("Expected sessionId 'sess456', got '%s'", log.SessionID)
	}

	if log.Metadata["totalSteps"] != float64(2) {
		t.Errorf("Expected totalSteps to be 2, got '%v'", log.Metadata["totalSteps"])
	}

	// Check if logs were cleaned up
	if len(logger.detailLogs) != 0 {
		t.Errorf("Expected detail logs to be cleaned up, got %d logs", len(logger.detailLogs))
	}
}

func TestLogger_FlushError(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn789", "sess999")

	// Add some detail logs
	logger.InfoDetail(ActionInfo{
		Action: "step1",
	}, map[string]interface{}{"data": "value1"})

	logger.ErrorDetail(ActionInfo{
		Action: "step2_failed",
	}, map[string]interface{}{"error": "something went wrong"})

	logger.AddMetadata("failedStep", "step2")

	output := captureOutput(func() {
		logger.FlushError(500, "Request failed with error")
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Type != TypeSummary {
		t.Errorf("Expected type 'summary', got '%s'", log.Type)
	}

	if log.Level != LevelError {
		t.Errorf("Expected level 'error', got '%s'", log.Level)
	}

	if log.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", log.StatusCode)
	}

	if log.Message != "Request failed with error" {
		t.Errorf("Expected message 'Request failed with error', got '%s'", log.Message)
	}

	if log.Action != "" {
		t.Errorf("Expected action to be empty for summary, got '%s'", log.Action)
	}

	if log.Metadata["failedStep"] != "step2" {
		t.Errorf("Expected failedStep to be 'step2', got '%v'", log.Metadata["failedStep"])
	}

	// Check if logs were cleaned up
	if len(logger.detailLogs) != 0 {
		t.Errorf("Expected detail logs to be cleaned up, got %d logs", len(logger.detailLogs))
	}
}

func TestLogger_FlushWithDuration(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")
	logger.StartTransaction("txn111", "sess222")

	// Simulate some work
	logger.InfoDetail(ActionInfo{Action: "work"}, map[string]interface{}{"status": "done"})

	output := captureOutput(func() {
		logger.Flush(201, "Resource created successfully")
	})

	var log DetailLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("Failed to unmarshal log: %v", err)
	}

	if log.Duration < 0 {
		t.Errorf("Expected duration to be >= 0, got %d", log.Duration)
	}

	if log.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", log.StatusCode)
	}
}

func TestLogger_MultipleFlush(t *testing.T) {
	logger := NewLogger("test-service", "1.0.0")

	// First transaction
	logger.StartTransaction("txn1", "sess1")
	logger.InfoDetail(ActionInfo{Action: "action1"}, map[string]interface{}{"data": "1"})
	
	output1 := captureOutput(func() {
		logger.Flush(200, "Transaction 1 complete")
	})

	var log1 DetailLog
	json.Unmarshal([]byte(output1), &log1)

	if log1.TransactionID != "txn1" {
		t.Errorf("Expected transactionId 'txn1', got '%s'", log1.TransactionID)
	}

	// Second transaction
	logger.StartTransaction("txn2", "sess2")
	logger.InfoDetail(ActionInfo{Action: "action2"}, map[string]interface{}{"data": "2"})
	
	output2 := captureOutput(func() {
		logger.Flush(200, "Transaction 2 complete")
	})

	var log2 DetailLog
	json.Unmarshal([]byte(output2), &log2)

	if log2.TransactionID != "txn2" {
		t.Errorf("Expected transactionId 'txn2', got '%s'", log2.TransactionID)
	}

	// Ensure logs are independent
	if log1.TransactionID == log2.TransactionID {
		t.Error("Transaction IDs should be different")
	}
}
