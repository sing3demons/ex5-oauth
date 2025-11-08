#!/bin/bash

echo "Registering OAuth2 Client..."

curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d @- << 'EOF'
{
  "name": "Test SSO Client",
  "redirect_uris": [
    "http://localhost:3000/callback",
    "http://localhost:3001/callback", 
    "http://localhost:3002/callback"
  ]
}
EOF

echo ""
echo "Save the client_id and client_secret for use in your React app!"
