#!/bin/bash

echo "🔐 Registering OAuth2 Client for Todo App..."
echo ""
echo "This script registers a confidential OAuth2 client with the OAuth2 server"
echo "for the Todo App backend (BFF pattern)."
echo ""

# Check if OAuth2 server is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Error: OAuth2 server is not running on http://localhost:8080"
    echo "Please start the OAuth2 server first:"
    echo "  ./oauth2-server"
    echo ""
    exit 1
fi

echo "✓ OAuth2 server is running"
echo ""

# Register the client
RESPONSE=$(curl -s -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Todo App with SSO",
    "redirect_uris": [
      "http://localhost:4000/auth/callback"
    ],
    "is_public": false
  }')

# Check if registration was successful
if [ $? -ne 0 ]; then
    echo "❌ Error: Failed to register client"
    exit 1
fi

# Parse response
CLIENT_ID=$(echo "$RESPONSE" | jq -r '.client_id')
CLIENT_SECRET=$(echo "$RESPONSE" | jq -r '.client_secret')

if [ "$CLIENT_ID" = "null" ] || [ -z "$CLIENT_ID" ]; then
    echo "❌ Error: Client registration failed"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "✅ Client registered successfully!"
echo ""
echo "Client Details:"
echo "$RESPONSE" | jq '.'
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📝 IMPORTANT: Update your backend/.env file with these values:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "OAUTH2_CLIENT_ID=$CLIENT_ID"
echo "OAUTH2_CLIENT_SECRET=$CLIENT_SECRET"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "You can automatically update the .env file by running:"
echo ""
echo "  cd oauth2-bff-app/backend"
echo "  sed -i.bak \"s/OAUTH2_CLIENT_ID=.*/OAUTH2_CLIENT_ID=$CLIENT_ID/\" .env"
echo "  sed -i.bak \"s/OAUTH2_CLIENT_SECRET=.*/OAUTH2_CLIENT_SECRET=$CLIENT_SECRET/\" .env"
echo ""
echo "Or manually copy the values above into oauth2-bff-app/backend/.env"
echo ""
