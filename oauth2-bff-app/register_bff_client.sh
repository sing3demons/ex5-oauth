#!/bin/bash

echo "🔐 Registering Confidential Client for BFF..."
echo ""

RESPONSE=$(curl -s -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Todo App BFF Client",
    "redirect_uris": [
      "http://localhost:3001/auth/callback"
    ],
    "is_public": false
  }')

echo "$RESPONSE" | jq '.'

CLIENT_ID=$(echo "$RESPONSE" | jq -r '.client_id')
CLIENT_SECRET=$(echo "$RESPONSE" | jq -r '.client_secret')

echo ""
echo "✅ Client registered successfully!"
echo ""
echo "📝 Update your backend/.env file with:"
echo "CLIENT_ID=$CLIENT_ID"
echo "CLIENT_SECRET=$CLIENT_SECRET"
echo ""
echo "Or run this command:"
echo "cd backend"
echo "sed -i '' 's/CLIENT_ID=.*/CLIENT_ID=$CLIENT_ID/' .env"
echo "sed -i '' 's/CLIENT_SECRET=.*/CLIENT_SECRET=$CLIENT_SECRET/' .env"
