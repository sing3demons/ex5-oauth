# OAuth2 Client Registration Guide

This document explains how to register the Todo App as an OAuth2 client with the OAuth2 server.

## Overview

The Todo App uses the **Backend-for-Frontend (BFF)** pattern with a **confidential OAuth2 client**. This means:

- The backend acts as an OAuth2 client
- Client credentials (ID and secret) are stored securely on the server
- The frontend never sees the client secret
- The backend handles all OAuth2 token exchanges

## Prerequisites

Before registering the client, ensure:

1. **OAuth2 Server is running** on `http://localhost:8080`
   ```bash
   # Start the OAuth2 server
   ./oauth2-server
   ```

2. **jq is installed** (for JSON parsing)
   ```bash
   # macOS
   brew install jq
   
   # Ubuntu/Debian
   sudo apt-get install jq
   ```

## Registration Methods

### Method 1: Automated Script (Recommended)

Use the provided registration script:

```bash
# From the project root
./oauth2-bff-app/register_todo_client.sh
```

This script will:
1. Check if the OAuth2 server is running
2. Register a new confidential client
3. Display the client credentials
4. Provide instructions for updating the `.env` file

**Example Output:**
```
🔐 Registering OAuth2 Client for Todo App...

✓ OAuth2 server is running

✅ Client registered successfully!

Client Details:
{
  "client_id": "X4jNSxivzBWKG3L0Tm2pYc0zKYprN0p9",
  "client_secret": "9VzaZ-0uLo8mx04hPja3Xp_xCQHncbYgn4MOP7cST2J92ojDiDKQ7j6XnyrZdOs-",
  "name": "Todo App with SSO",
  "redirect_uris": [
    "http://localhost:4000/auth/callback"
  ],
  "is_public": false
}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 IMPORTANT: Update your backend/.env file with these values:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

OAUTH2_CLIENT_ID=X4jNSxivzBWKG3L0Tm2pYc0zKYprN0p9
OAUTH2_CLIENT_SECRET=9VzaZ-0uLo8mx04hPja3Xp_xCQHncbYgn4MOP7cST2J92ojDiDKQ7j6XnyrZdOs-
```

### Method 2: Manual cURL Request

If you prefer to register manually:

```bash
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Todo App with SSO",
    "redirect_uris": [
      "http://localhost:4000/auth/callback"
    ],
    "is_public": false
  }'
```

## Configuration

After registration, update the backend environment variables:

### 1. Copy the Example Environment File

```bash
cd oauth2-bff-app/backend
cp .env.example .env
```

### 2. Update the .env File

Edit `oauth2-bff-app/backend/.env` and add your client credentials:

```bash
# Server Configuration
PORT=4000
NODE_ENV=development

# OAuth2 Configuration
OAUTH2_SERVER_URL=http://localhost:8080
OAUTH2_CLIENT_ID=<your-client-id-from-registration>
OAUTH2_CLIENT_SECRET=<your-client-secret-from-registration>
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback

# Database Configuration
MONGODB_URI=mongodb://localhost:27017/todo_app

# Security
SESSION_SECRET=<generate-a-random-secret-key>

# CORS
FRONTEND_URL=http://localhost:3000
CORS_ORIGIN=http://localhost:3000
```

### 3. Generate a Session Secret

For production, generate a strong session secret:

```bash
# Generate a random 32-byte hex string
node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
```

## Client Configuration Details

### Client Type: Confidential

The Todo App uses a **confidential client** because:

- The backend can securely store the client secret
- The client secret is never exposed to the browser
- This provides better security than public clients
- Suitable for server-side applications

### Redirect URI

The redirect URI must match exactly:

```
http://localhost:4000/auth/callback
```

**Important:** 
- The OAuth2 server validates redirect URIs strictly
- Any mismatch will result in an authorization error
- For production, register additional URIs (e.g., `https://api.todo.example.com/auth/callback`)

### Registered Scopes

The client will request the following scopes during authorization:

- `openid` - Required for OIDC authentication
- `profile` - Access to user profile information (name, picture)
- `email` - Access to user email address

## Security Best Practices

### 1. Protect Client Credentials

- **Never commit** `.env` files to version control
- Store credentials in environment variables or secret management systems
- Rotate credentials periodically
- Use different credentials for development and production

### 2. Environment-Specific Clients

Register separate clients for each environment:

- **Development:** `http://localhost:4000/auth/callback`
- **Staging:** `https://api-staging.todo.example.com/auth/callback`
- **Production:** `https://api.todo.example.com/auth/callback`

### 3. Validate Redirect URIs

The OAuth2 server will only redirect to registered URIs. This prevents:

- Authorization code interception
- Redirect-based attacks
- Unauthorized client access

## Troubleshooting

### Error: OAuth2 server is not running

**Solution:** Start the OAuth2 server first:
```bash
./oauth2-server
```

### Error: Client registration failed

**Possible causes:**
1. OAuth2 server is not accessible
2. Invalid JSON in the request
3. Server configuration issues

**Solution:** Check the OAuth2 server logs and verify the endpoint is accessible:
```bash
curl http://localhost:8080/health
```

### Error: redirect_uri_mismatch

**Cause:** The redirect URI in the authorization request doesn't match the registered URI.

**Solution:** Ensure the `OAUTH2_REDIRECT_URI` in `.env` matches exactly:
```
http://localhost:4000/auth/callback
```

### Error: invalid_client

**Cause:** Client ID or secret is incorrect.

**Solution:** 
1. Verify the credentials in `.env` match the registration response
2. Re-register the client if credentials are lost
3. Check for extra whitespace or newlines in the `.env` file

## Verification

After registration and configuration, verify the setup:

### 1. Check Environment Variables

```bash
cd oauth2-bff-app/backend
cat .env | grep OAUTH2
```

Expected output:
```
OAUTH2_SERVER_URL=http://localhost:8080
OAUTH2_CLIENT_ID=<your-client-id>
OAUTH2_CLIENT_SECRET=<your-client-secret>
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback
```

### 2. Test OAuth2 Server Connectivity

```bash
curl http://localhost:8080/.well-known/openid-configuration
```

This should return the OAuth2 server's OIDC discovery document.

### 3. Verify Client Registration

You can verify the client is registered by checking the OAuth2 server's client list (if available) or by attempting an authorization flow.

## Next Steps

After successful client registration:

1. ✅ Client registered with OAuth2 server
2. ✅ Credentials stored in `.env` file
3. ⏭️ Implement backend OAuth2 authentication routes (Task 3)
4. ⏭️ Implement token exchange and validation (Task 3.2)
5. ⏭️ Create session management (Task 3.3)

## Additional Resources

- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 for Browser-Based Apps](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps)
- [Backend-for-Frontend Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/backends-for-frontends)

## Support

If you encounter issues:

1. Check the OAuth2 server logs
2. Verify all prerequisites are met
3. Review the troubleshooting section
4. Ensure environment variables are correctly set
