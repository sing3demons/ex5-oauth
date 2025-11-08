#!/bin/bash

# OAuth2/OIDC Server API Test Script

BASE_URL="http://localhost:8080"

echo "=== OAuth2/OIDC Server API Test (RS256) ==="
echo ""

# 0. Check JWKS endpoint
echo "0. Checking JWKS endpoint..."
JWKS_RESPONSE=$(curl -s -X GET "$BASE_URL/.well-known/jwks.json")
echo "JWKS Response: $JWKS_RESPONSE"
echo ""

# 1. Register User
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }')
echo "Response: $REGISTER_RESPONSE"
echo ""

# 2. Login
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')
echo "Response: $LOGIN_RESPONSE"
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo "Access Token: $ACCESS_TOKEN"
echo ""

# 3. Register OAuth Client
echo "3. Registering OAuth client..."
CLIENT_RESPONSE=$(curl -s -X POST "$BASE_URL/clients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Application",
    "redirect_uris": ["http://localhost:3000/callback"]
  }')
echo "Response: $CLIENT_RESPONSE"
CLIENT_ID=$(echo $CLIENT_RESPONSE | grep -o '"client_id":"[^"]*' | cut -d'"' -f4)
CLIENT_SECRET=$(echo $CLIENT_RESPONSE | grep -o '"client_secret":"[^"]*' | cut -d'"' -f4)
echo "Client ID: $CLIENT_ID"
echo "Client Secret: $CLIENT_SECRET"
echo ""

# 4. Get Authorization Code
echo "4. Getting authorization code..."
AUTH_RESPONSE=$(curl -s -X GET "$BASE_URL/oauth/authorize?response_type=code&client_id=$CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random123" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $AUTH_RESPONSE"
AUTH_CODE=$(echo $AUTH_RESPONSE | grep -o '"code":"[^"]*' | cut -d'"' -f4)
echo "Authorization Code: $AUTH_CODE"
echo ""

# 5. Exchange code for tokens
echo "5. Exchanging authorization code for tokens..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=$AUTH_CODE&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&redirect_uri=http://localhost:3000/callback")
echo "Response: $TOKEN_RESPONSE"
NEW_ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)
ID_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"id_token":"[^"]*' | cut -d'"' -f4)
echo "New Access Token: $NEW_ACCESS_TOKEN"
echo "Refresh Token: $REFRESH_TOKEN"
echo "ID Token: $ID_TOKEN"
echo ""

# 6. Get UserInfo
echo "6. Getting user info..."
USERINFO_RESPONSE=$(curl -s -X GET "$BASE_URL/oauth/userinfo" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN")
echo "Response: $USERINFO_RESPONSE"
echo ""

# 7. Refresh Token
echo "7. Refreshing token..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET")
echo "Response: $REFRESH_RESPONSE"
echo ""

# 8. Client Credentials Grant
echo "8. Testing client credentials grant..."
CLIENT_CREDS_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&scope=api:read")
echo "Response: $CLIENT_CREDS_RESPONSE"
echo ""

# 9. OIDC Discovery
echo "9. Getting OIDC discovery configuration..."
DISCOVERY_RESPONSE=$(curl -s -X GET "$BASE_URL/.well-known/openid-configuration")
echo "Response: $DISCOVERY_RESPONSE"
echo ""

echo "=== Test Complete ==="
echo ""
echo "Note: All tokens are now signed with RS256 (RSA)"
