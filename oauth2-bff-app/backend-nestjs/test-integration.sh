#!/bin/bash

# Integration test script for NestJS backend
# This script tests all the endpoints and flows

set -e

BASE_URL="http://localhost:3001"
OAUTH_URL="http://localhost:8080"

echo "🧪 Starting NestJS Backend Integration Tests"
echo "=============================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local expected_status=$4
    local extra_args=${5:-""}
    
    echo -n "Testing: $name... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" $extra_args "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method $extra_args "$url")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (Status: $status_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected: $expected_status, Got: $status_code)"
        echo "Response: $body"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo "📋 Test 1: Health Check"
echo "----------------------"
test_endpoint "Health endpoint" "GET" "$BASE_URL/health" "200"
echo ""

echo "📋 Test 2: OAuth2/OIDC Endpoints"
echo "--------------------------------"
test_endpoint "Login initiation" "GET" "$BASE_URL/auth/login" "200"
test_endpoint "OIDC Discovery" "GET" "$BASE_URL/auth/discovery" "200"
test_endpoint "JWKS endpoint" "GET" "$BASE_URL/auth/jwks" "200"
echo ""

echo "📋 Test 3: Protected Endpoints (Should Fail Without Auth)"
echo "---------------------------------------------------------"
test_endpoint "Get todos without auth" "GET" "$BASE_URL/api/todos" "401"
test_endpoint "Get userinfo without auth" "GET" "$BASE_URL/auth/userinfo" "401"
test_endpoint "Refresh without cookie" "POST" "$BASE_URL/auth/refresh" "401"
echo ""

echo "📋 Test 4: Token Utilities"
echo "--------------------------"
# Test decode token with a sample JWT (this will fail validation but should decode)
SAMPLE_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OWDEjXwaYtuAXA5R8SoulwM0dG2HE5Y5QLr_BpXhR6zXjPDt3wU69p"

test_endpoint "Decode JWT token" "POST" "$BASE_URL/auth/decode-token" "200" \
    "-H 'Content-Type: application/json' -d '{\"token\":\"$SAMPLE_TOKEN\"}'"
echo ""

echo "📋 Test 5: CORS Configuration"
echo "-----------------------------"
echo -n "Testing CORS headers... "
cors_response=$(curl -s -I -H "Origin: http://localhost:5173" "$BASE_URL/health")
if echo "$cors_response" | grep -q "Access-Control-Allow-Origin"; then
    echo -e "${GREEN}✓ PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}✗ FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

echo "📋 Test 6: Error Handling"
echo "-------------------------"
test_endpoint "Invalid endpoint" "GET" "$BASE_URL/invalid/endpoint" "404"
test_endpoint "Invalid todo ID" "GET" "$BASE_URL/api/todos/invalid-id" "401"
echo ""

echo "=============================================="
echo "📊 Test Summary"
echo "=============================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    exit 1
fi
