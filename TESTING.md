# Testing Guide

This document explains how to run tests in the OAuth2 server project.

## Test Types

The project has two types of tests:

### 1. Unit Tests
- Fast, isolated tests that don't require external dependencies
- Run by default with `go test`
- Don't require MongoDB

### 2. Integration Tests
- Tests that require MongoDB connection
- Tagged with `//go:build integration`
- Must be explicitly enabled with `-tags=integration`

## Running Tests

### Run Unit Tests Only (Default)

```bash
# Run all unit tests
go test ./...

# Run unit tests in a specific package
go test ./handlers

# Run with verbose output
go test -v ./handlers
```

### Run Integration Tests

Integration tests require MongoDB to be running on `localhost:27017`.

```bash
# Start MongoDB (if using Docker)
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Run integration tests
go test -tags=integration ./handlers -v

# Run specific integration test
go test -tags=integration ./handlers -run TestScopeValidationFlow -v
```

### Run All Tests (Unit + Integration)

```bash
# Run all tests including integration tests
go test -tags=integration ./... -v
```

## Test Coverage

```bash
# Generate coverage report for unit tests
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Generate coverage report including integration tests
go test -tags=integration ./... -coverprofile=coverage-full.out
go tool cover -html=coverage-full.out
```

## Integration Test Files

Files with integration tests:
- `handlers/oauth_handler_test.go` - UserInfo endpoint tests
- `handlers/scope_integration_test.go` - Scope validation, claim filtering, and scope downgrade tests

## CI/CD Considerations

In CI/CD pipelines:

1. **Fast feedback**: Run unit tests first
   ```bash
   go test ./...
   ```

2. **Full validation**: Run integration tests with MongoDB service
   ```bash
   # Start MongoDB service
   docker run -d -p 27017:27017 mongo:latest
   
   # Wait for MongoDB to be ready
   sleep 5
   
   # Run integration tests
   go test -tags=integration ./... -v
   ```

## Test Database

Integration tests use separate test databases:
- `oauth2_test_userinfo`
- `oauth2_test_scope_validation`
- `oauth2_test_claim_filtering`
- `oauth2_test_scope_downgrade`
- etc.

Each test creates and drops its own database to ensure isolation.

## Skipping Tests

Integration tests automatically skip if MongoDB is not available:

```go
client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    t.Skip("MongoDB not available, skipping integration test")
    return
}
```

## Best Practices

1. **Keep unit tests fast** - No external dependencies
2. **Use integration tests for end-to-end flows** - Test with real database
3. **Clean up test data** - Each test should clean up after itself
4. **Use descriptive test names** - Make it clear what is being tested
5. **Test both success and failure cases** - Cover error scenarios

## Example Test Commands

```bash
# Quick check (unit tests only)
go test ./...

# Full validation before commit
go test -tags=integration ./... -v

# Test specific feature
go test -tags=integration ./handlers -run TestScopeValidation -v

# Benchmark tests
go test -bench=. ./...

# Race condition detection
go test -race ./...
```
