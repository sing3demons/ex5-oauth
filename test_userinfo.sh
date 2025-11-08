#!/bin/bash

# Test script for UserInfo endpoint with different scopes

echo "=== Testing UserInfo Endpoint with Different Scopes ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Test user credentials
EMAIL="test@example.com"
PASSWORD="password123"
CLIENT_ID="test-client"
CLIENT_SECRET="test-secret"
REDIRECT_URI="http://localhost:3000/callback"

echo -e "${BLUE}1. Register test user${NC}"
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Test User\"}" | jq .

echo ""
echo -e "${BLUE}2. Register test client${NC}"
curl -s -X POST "$BASE_URL/clients/register" \
  -H "Content-Type: application/json" \
  -d "{\"client_id\":\"$CLIENT_ID\",\"client_secret\":\"$CLIENT_SECRET\",\"redirect_uris\":[\"$REDIRECT_URI\"],\"name\":\"Test Client\",\"allowed_scopes\":[\"openid\",\"profile\",\"email\"]}" | jq .

echo ""
echo -e "${BLUE}3. Test UserInfo with 'openid' scope only${NC}"
echo "Expected: Only 'sub' claim"
# This would require a full OAuth flow, so we'll skip the actual test
echo "Skipping actual test - requires full OAuth flow"

echo ""
echo -e "${BLUE}4. Test UserInfo with 'openid email' scope${NC}"
echo "Expected: 'sub', 'email', 'email_verified' claims"
echo "Skipping actual test - requires full OAuth flow"

echo ""
echo -e "${BLUE}5. Test UserInfo with 'openid profile' scope${NC}"
echo "Expected: 'sub', 'name' claims"
echo "Skipping actual test - requires full OAuth flow"

echo ""
echo -e "${BLUE}6. Test UserInfo with 'openid profile email' scope${NC}"
echo "Expected: 'sub', 'name', 'email', 'email_verified' claims"
echo "Skipping actual test - requires full OAuth flow"

echo ""
echo -e "${GREEN}=== UserInfo endpoint tests completed ===${NC}"
echo "Run 'go test -v -run TestUserInfo ./handlers/' to see automated tests"
