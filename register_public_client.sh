#!/bin/bash

echo "Registering Public OAuth2 Client (for PKCE)..."

curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d @- << 'EOF'
{
  "name": "React SSO Client (Public)",
  "redirect_uris": [
    "http://localhost:3000/callback",
    "http://localhost:3001/callback", 
    "http://localhost:3002/callback"
  ],
  "is_public": true
}
EOF

echo ""
echo "This is a PUBLIC client - no client_secret needed!"
echo "Use PKCE (code_verifier/code_challenge) for security."
