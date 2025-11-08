#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== OAuth2 Server - JWE & Token Exchange Test ==="
echo ""

# 1. Register Client
echo "1. Registering OAuth2 Client..."
CLIENT_RESPONSE=$(curl -s -X POST "$BASE_URL/clients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Test JWE Client",
    "redirect_uris": ["http://localhost:3000/callback"]
  }')

CLIENT_ID=$(echo $CLIENT_RESPONSE | grep -o '"client_id":"[^"]*"' | cut -d'"' -f4)
CLIENT_SECRET=$(echo $CLIENT_RESPONSE | grep -o '"client_secret":"[^"]*"' | cut -d'"' -f4)

echo "Client ID: $CLIENT_ID"
echo "Client Secret: $CLIENT_SECRET"
echo ""

# 2. Register User
echo "2. Registering User..."
USER_EMAIL="jwe_test@example.com"
USER_PASSWORD="password123"

curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "email=$USER_EMAIL&password=$USER_PASSWORD&name=JWE Test User" > /dev/null

echo "User registered: $USER_EMAIL"
echo ""

# 3. Login and get session
echo "3. Login to get session..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "email=$USER_EMAIL&password=$USER_PASSWORD")

SESSION_ID=$(echo $LOGIN_RESPONSE | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo "Session ID: $SESSION_ID"
echo ""

# 4. Get Authorization Code
echo "4. Getting Authorization Code..."
AUTH_RESPONSE=$(curl -s -L "$BASE_URL/oauth/authorize?response_type=code&client_id=$CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=xyz&session_id=$SESSION_ID")

AUTH_CODE=$(echo $AUTH_RESPONSE | grep -o 'code=[^&"]*' | cut -d'=' -f2)
echo "Authorization Code: $AUTH_CODE"
echo ""

# 5. Exchange code for JWT tokens (normal)
echo "5. Exchange code for JWT tokens..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=$AUTH_CODE&redirect_uri=http://localhost:3000/callback&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET")

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
ID_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"id_token":"[^"]*"' | cut -d'"' -f4)

echo "JWT Access Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
echo "JWT ID Token (first 50 chars): ${ID_TOKEN:0:50}..."
echo ""

# 6. Validate JWT tokens
echo "6. Validating JWT Access Token..."
VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/token/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$ACCESS_TOKEN\"}")

echo "Validation Response:"
echo $VALIDATE_RESPONSE | jq '.'
echo ""

# 7. Token Exchange to JWE
echo "7. Token Exchange: JWT -> JWE (with is_encrypted_jwe=true)..."
EXCHANGE_RESPONSE=$(curl -s -X POST "$BASE_URL/token/exchange" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange&subject_token=$ACCESS_TOKEN&subject_token_type=urn:ietf:params:oauth:token-type:access_token&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&is_encrypted_jwe=true")

JWE_ACCESS_TOKEN=$(echo $EXCHANGE_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
JWE_ID_TOKEN=$(echo $EXCHANGE_RESPONSE | grep -o '"id_token":"[^"]*"' | cut -d'"' -f4)

echo "JWE Access Token (first 80 chars): ${JWE_ACCESS_TOKEN:0:80}..."
echo "JWE ID Token (first 80 chars): ${JWE_ID_TOKEN:0:80}..."
echo ""

# 8. Validate JWE tokens
echo "8. Validating JWE Access Token..."
JWE_VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/token/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$JWE_ACCESS_TOKEN\"}")

echo "JWE Validation Response:"
echo $JWE_VALIDATE_RESPONSE | jq '.'
echo ""

# 9. Compare JWT vs JWE structure
echo "9. Token Format Comparison:"
echo "JWT parts (should be 3): $(echo $ACCESS_TOKEN | tr '.' '\n' | wc -l)"
echo "JWE parts (should be 5): $(echo $JWE_ACCESS_TOKEN | tr '.' '\n' | wc -l)"
echo ""

# 10. Validate ID Token
echo "10. Validating JWE ID Token..."
ID_VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/token/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$JWE_ID_TOKEN\"}")

echo "JWE ID Token Validation:"
echo $ID_VALIDATE_RESPONSE | jq '.'
echo ""

echo "=== Test Complete ==="
