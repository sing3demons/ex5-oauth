# Logger Package

Structured logging system with detail and summary logs, data masking, and transaction tracking.

## Features

- **Detail Logs**: Individual operation logs with optional data masking
- **Summary Logs**: Transaction summary with accumulated metadata
- **Data Masking**: Protect sensitive data (passwords, emails, cards)
- **File & Console Output**: Configurable output destinations
- **Transaction Tracking**: Track operations across a request lifecycle

## Usage

### Basic Logging

```go
logger := logger.NewLogger("my-service", "1.0.0")

// Simple log
logger.Info("user_action", "User performed action",
    logger.WithUserID("user123"),
    logger.WithMetadata("key", "value"),
)
```

### Detail Logging with Masking

```go
actionInfo := logger.ActionInfo{
    Action:            "user_login",
    ActionDescription: "User attempting to login",
    SubAction:         "validate_credentials",
}

data := map[string]interface{}{
    "email":    "user@example.com",
    "password": "secret123",
}

maskingRules := []logger.MaskingRule{
    {Field: "password", Type: logger.MaskingTypeFull},
    {Field: "email", Type: logger.MaskingTypeEmail},
}

// Log detail with masking
logger.InfoDetail(actionInfo, data, maskingRules...)
```

**Note:** Only use `InfoDetail`, `DebugDetail`, `WarnDetail`, `ErrorDetail` for detail logs. For summary logs, use `Flush()` or `FlushError()` only.

### Transaction with Summary

```go
logger := logger.NewWithConfig("my-service", "1.0.0", logConfig)

// Start transaction
logger.StartTransaction("txn-123", "session-456")

// Log details
logger.InfoDetail(logger.ActionInfo{
    Action: "step1",
    ActionDescription: "Processing step 1",
}, map[string]interface{}{
    "input": "data",
})

logger.InfoDetail(logger.ActionInfo{
    Action: "step2",
    ActionDescription: "Processing step 2",
}, map[string]interface{}{
    "result": "success",
})

// Add metadata for summary
logger.AddMetadata("totalSteps", 2)
logger.AddMetadata("processedItems", 100)

// Flush summary (success)
logger.Flush(200, "Request completed successfully")

// Or flush with error
// logger.FlushError(500, "Request failed")
```

## Masking Types

### Full Masking
```go
{Field: "password", Type: logger.MaskingTypeFull}
// "secret123" → "***"
```

### Partial Masking
```go
{Field: "token", Type: logger.MaskingTypePartial}
// "abc123xyz" → "a*******z"
```

### Email Masking
```go
{Field: "email", Type: logger.MaskingTypeEmail}
// "user@example.com" → "u***@example.com"
```

### Card Masking
```go
{Field: "cardNumber", Type: logger.MaskingTypeCard}
// "1234-5678-9012-3456" → "****-****-****-3456"
```

## Configuration

### From Config File (YAML)

```yaml
service: "my-service"
version: "1.0.0"

logging:
  summary:
    path: "./logs/summary/"
    console: true
    file: true
  detail:
    path: "./logs/detail/"
    console: true
    file: true
```

### From Environment Variables

```bash
SERVICE_NAME=my-service
SERVICE_VERSION=1.0.0
LOG_SUMMARY_PATH=./logs/summary/
LOG_SUMMARY_CONSOLE=true
LOG_SUMMARY_FILE=true
LOG_DETAIL_PATH=./logs/detail/
LOG_DETAIL_CONSOLE=true
LOG_DETAIL_FILE=true
```

### Programmatically

```go
logConfig := logger.LoggingConfig{
    Summary: logger.LogOutputConfig{
        Path:    "./logs/summary/",
        Console: true,
        File:    true,
    },
    Detail: logger.LogOutputConfig{
        Path:    "./logs/detail/",
        Console: true,
        File:    false,
    },
}

logger := logger.NewWithConfig("my-service", "1.0.0", logConfig)
```

## Log Structure

### Detail Log
```json
{
  "timestamp": "2024-11-09T10:30:45.123Z",
  "level": "info",
  "type": "detail",
  "service": "oauth2-server",
  "version": "1.0.0",
  "transactionId": "txn-123",
  "sessionId": "session-456",
  "action": "user_login",
  "actionDescription": "User attempting to login",
  "subAction": "validate_credentials",
  "message": "{\"email\":\"user@example.com\"}",
  "userId": "user-789",
  "metadata": {
    "data": {
      "email": "u***@example.com",
      "password": "***"
    }
  }
}
```

### Summary Log
```json
{
  "timestamp": "2024-11-09T10:30:46.456Z",
  "level": "info",
  "type": "summary",
  "service": "oauth2-server",
  "version": "1.0.0",
  "transactionId": "txn-123",
  "sessionId": "session-456",
  "message": "Request completed successfully",
  "statusCode": 200,
  "duration": 1333,
  "metadata": {
    "detailLogCount": 5,
    "duration": 1333,
    "totalSteps": 2,
    "processedItems": 100
  }
}
```

**Note:** Summary logs don't have `action` field. Use `statusCode` and `message` to indicate success or failure.

## Best Practices

1. **Start Transaction Early**: Call `StartTransaction()` at the beginning of request handling
2. **Use Detail Logs**: Log important steps with `InfoDetail()` or `ErrorDetail()`
3. **Add Metadata**: Use `AddMetadata()` to accumulate summary information
4. **Use AddSuccess for Multiple Values**: When same key appears multiple times, use `AddSuccess()` to automatically create arrays
5. **Always Flush**: Call `Flush()` or `FlushError()` at the end to write summary
6. **Mask Sensitive Data**: Always mask passwords, tokens, PII data
7. **Use Appropriate Levels**: Info for normal flow, Error for failures, Debug for troubleshooting

## AddSuccess vs AddMetadata

- **AddMetadata(key, value)**: Sets or overwrites a key with a single value
- **AddSuccess(key, value)**: Adds value to key, creating array if key already exists

```go
// Using AddMetadata
logger.AddMetadata("userId", "user1")
logger.AddMetadata("userId", "user2") // Overwrites: "user2"

// Using AddSuccess
logger.AddSuccess("userId", "user1")
logger.AddSuccess("userId", "user2") // Creates array: ["user1", "user2"]
logger.AddSuccess("userId", "user3") // Appends: ["user1", "user2", "user3"]
```

## Example: HTTP Handler

```go
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Start transaction
    txnID := generateTransactionID()
    sessionID := r.URL.Query().Get("session_id")
    
    h.logger.StartTransaction(txnID, sessionID)
    
    // Log request details
    h.logger.InfoDetail(logger.ActionInfo{
        Action: "http_request_start",
        ActionDescription: "Processing HTTP request",
    }, map[string]interface{}{
        "method": r.Method,
        "path": r.URL.Path,
    })
    
    // Process request
    result, err := h.processRequest(r)
    if err != nil {
        h.logger.ErrorDetail(logger.ActionInfo{
            Action: "processing_failed",
            ActionDescription: "Request processing failed",
        }, map[string]interface{}{
            "error": err.Error(),
        })
        
        h.logger.FlushError(500, err.Error())
        http.Error(w, err.Error(), 500)
        return
    }
    
    // Add summary metadata
    h.logger.AddMetadata("resultCount", len(result))
    
    // Success
    h.logger.Flush(200, "Request completed successfully")
    json.NewEncoder(w).Encode(result)
}
```
