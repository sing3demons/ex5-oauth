# SSO Quick Reference Guide

Quick reference for developers implementing or using the SSO functionality.

## Configuration

### Environment Variables

```bash
SSO_SESSION_EXPIRY_DAYS=7          # Default: 7 days
SSO_CONSENT_EXPIRY_DAYS=365        # Default: 1 year
SSO_COOKIE_SECURE=true             # Set to false for local dev
```

### Cookie Settings

```
Name:        oauth_sso_session
Max Age:     604800 seconds (7 days)
Path:        /
HttpOnly:    true
Secure:      true (production)
SameSite:    Lax
```

## SSO Flow Patterns

### Pattern 1: First Login (Full Flow)

```
User → /oauth/authorize → Login Page → Consent Page → Authorization Code
```

**Time**: ~2-5 seconds (user interaction)

### Pattern 2: Returning User (Auto-Approval)

```
User → /oauth/authorize → Authorization Code (immediate)
```

**Time**: < 100ms (no user interaction)

### Pattern 3: New App (Consent Only)

```
User → /oauth/authorize → Consent Page → Authorization Code
```

**Time**: ~1-3 seconds (user interaction)

## API Endpoints

### Session Management

```bash
# List sessions
GET /account/sessions
Authorization: Bearer {token}

# Revoke session
DELETE /account/sessions/{session_id}
Authorization: Bearer {token}
```

### Authorization Management

```bash
# List authorizations
GET /account/authorizations
Authorization: Bearer {token}

# Revoke authorization
DELETE /account/authorizations/{client_id}
Authorization: Bearer {token}
```

### Logout

```bash
# Simple logout
POST /auth/logout

# Logout with redirect
POST /auth/logout?post_logout_redirect_uri=https://example.com/goodbye
```

## OIDC Prompt Parameter

### prompt=none

```bash
# Silent authentication check
GET /oauth/authorize?...&prompt=none

# Returns:
# - Authorization code (if authenticated + consent exists)
# - login_required error (if not authenticated)
# - consent_required error (if consent missing)
```

### prompt=login

```bash
# Force re-authentication
GET /oauth/authorize?...&prompt=login

# Always shows login page, even with valid SSO session
```

### prompt=consent

```bash
# Force consent screen
GET /oauth/authorize?...&prompt=consent

# Always shows consent page, even with existing consent
```

### prompt=select_account

```bash
# Account selection (future)
GET /oauth/authorize?...&prompt=select_account

# Currently treated as normal flow
```

## Common Use Cases

### Use Case 1: Check Authentication Status

```bash
# Use prompt=none to check without user interaction
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=my-app&\
redirect_uri=http://localhost:3000/callback&\
scope=openid&\
state=check123&\
prompt=none" \
-b cookies.txt

# Success: Redirects with code
# Failure: Redirects with login_required or consent_required
```

### Use Case 2: Force Fresh Login

```bash
# Use prompt=login for high-security operations
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=my-app&\
redirect_uri=http://localhost:3000/callback&\
scope=openid&\
state=secure123&\
prompt=login"

# Always shows login page
```

### Use Case 3: Request Additional Permissions

```bash
# Use prompt=consent to show consent for new scopes
curl "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=my-app&\
redirect_uri=http://localhost:3000/callback&\
scope=openid profile email&\
state=newscope123&\
prompt=consent"

# Shows consent page with all requested scopes
```

### Use Case 4: Security Audit

```bash
# List all active sessions
ACCESS_TOKEN="your_token_here"

curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq '.sessions[] | {session_id, ip_address, last_activity}'

# Revoke suspicious session
curl -X DELETE http://localhost:8080/account/sessions/suspicious_session_id \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Use Case 5: Privacy Management

```bash
# List authorized apps
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  | jq '.authorizations[] | {client_name, scopes, granted_at}'

# Revoke unused app
curl -X DELETE http://localhost:8080/account/authorizations/old-app \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Database Queries

### Check SSO Session

```javascript
// MongoDB
db.sso_sessions.findOne({
  session_id: "SESSION_ID",
  expires_at: { $gt: new Date() }
})
```

### Check User Consent

```javascript
// MongoDB
db.user_consents.findOne({
  user_id: "USER_ID",
  client_id: "CLIENT_ID"
})
```

### List User Sessions

```javascript
// MongoDB
db.sso_sessions.find({
  user_id: "USER_ID",
  expires_at: { $gt: new Date() }
}).sort({ created_at: -1 })
```

### Cleanup Expired Sessions

```javascript
// MongoDB
db.sso_sessions.deleteMany({
  expires_at: { $lt: new Date() }
})
```

## Response Examples

### Successful Authorization

```http
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?code=AUTH_CODE&state=xyz123
Set-Cookie: oauth_sso_session=SESSION_ID; HttpOnly; Secure; SameSite=Lax; Max-Age=604800
```

### Login Required (prompt=none)

```http
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?error=login_required&error_description=User+authentication+required&state=xyz123
```

### Consent Required (prompt=none)

```http
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?error=consent_required&error_description=User+consent+required&state=xyz123
```

### Session List Response

```json
{
  "sessions": [
    {
      "session_id": "abc123...",
      "created_at": "2025-11-09T10:00:00Z",
      "last_activity": "2025-11-09T15:30:00Z",
      "expires_at": "2025-11-16T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0..."
    }
  ]
}
```

### Authorization List Response

```json
{
  "authorizations": [
    {
      "client_id": "my-app",
      "client_name": "My Application",
      "scopes": ["openid", "profile", "email"],
      "granted_at": "2025-11-01T10:00:00Z",
      "expires_at": "2026-11-01T10:00:00Z"
    }
  ]
}
```

## Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid_request` | 400 | Malformed request |
| `invalid_token` | 401 | Invalid access token |
| `forbidden` | 403 | Access denied |
| `not_found` | 404 | Resource not found |
| `login_required` | 302 | Authentication required (prompt=none) |
| `consent_required` | 302 | Consent required (prompt=none) |
| `server_error` | 500 | Internal error |

## Security Checklist

- [ ] HTTPS enabled in production
- [ ] `SSO_COOKIE_SECURE=true` in production
- [ ] Regular session cleanup (expired sessions)
- [ ] Monitor for suspicious activity
- [ ] User education on session management
- [ ] Rate limiting on login endpoint
- [ ] Audit logging enabled

## Performance Metrics

| Operation | Target | Typical |
|-----------|--------|---------|
| SSO session lookup | < 10ms | ~5ms |
| Consent check | < 10ms | ~5ms |
| Auto-approval flow | < 100ms | ~50ms |
| Session cleanup (10k) | < 5s | ~2s |

## Troubleshooting

### SSO Not Working

1. Check cookie exists: Browser DevTools → Application → Cookies
2. Verify session in DB: `db.sso_sessions.findOne({session_id: "..."})`
3. Check expiration: `expires_at` > current time
4. Verify cookie domain matches

### Consent Not Remembered

1. Check consent in DB: `db.user_consents.findOne({user_id: "...", client_id: "..."})`
2. Verify scopes match exactly
3. Check consent not expired
4. Ensure no `prompt=consent` parameter

### Auto-Approval Not Working

1. Verify SSO session valid
2. Check consent exists
3. Ensure no `prompt=login` or `prompt=consent`
4. Verify scopes match

## Additional Resources

- [SSO Usage Guide](./SSO_USAGE.md) - Comprehensive examples
- [SSO API Reference](./SSO_API_REFERENCE.md) - Complete API docs
- [Main README](../README.md) - Project overview
