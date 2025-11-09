# Single Sign-On (SSO) Usage Guide

This guide provides comprehensive examples and explanations for using the Single Sign-On (SSO) functionality in the OAuth2 Server.

## Table of Contents

- [Overview](#overview)
- [SSO Flow Examples](#sso-flow-examples)
- [Session Management](#session-management)
- [Authorization Management](#authorization-management)
- [OIDC Prompt Parameter](#oidc-prompt-parameter)
- [Configuration](#configuration)
- [Security Considerations](#security-considerations)

## Overview

Single Sign-On (SSO) allows users to authenticate once and access multiple client applications without re-entering credentials. The OAuth2 Server maintains a secure server-side session and remembers user consent for each application.

### How It Works

1. **First Login**: User authenticates and grants consent to App A
   - SSO session created (7-day lifetime)
   - User consent saved (1-year lifetime)
   - Secure HTTP-only cookie set

2. **Subsequent Apps**: User accesses App B
   - SSO session validated from cookie
   - Existing consent checked
   - Authorization code generated automatically (no login/consent screens)

3. **Logout**: User logs out
   - SSO session deleted from database
   - Cookie cleared
   - Next access requires re-authentication

## SSO Flow Examples

### Example 1: First Application Login

**Scenario**: User accesses App A for the first time.

```bash
# Step 1: Client initiates authorization
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile email&
  state=random123

# Step 2: No SSO session exists → Redirect to login
# User sees login page at /auth/login

# Step 3: User submits credentials
POST http://localhost:8080/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "session_id": "oauth_session_abc123"
}

# Step 4: SSO session created, consent screen shown
# User sees consent page at /oauth/consent

# Step 5: User approves consent
POST http://localhost:8080/oauth/consent
Content-Type: application/x-www-form-urlencoded

approved=true&
client_id=app-a&
scope=openid profile email&
state=random123&
redirect_uri=http://localhost:3000/callback

# Step 6: Authorization code generated and redirected
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?code=AUTH_CODE&state=random123
Set-Cookie: oauth_sso_session=SESSION_ID; HttpOnly; Secure; SameSite=Lax; Max-Age=604800
```

**Result**: 
- SSO session created with 7-day expiration
- User consent saved for App A
- Authorization code issued

### Example 2: Second Application (Auto-Approval)

**Scenario**: Same user accesses App B immediately after.

```bash
# Step 1: Client initiates authorization
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-b&
  redirect_uri=http://localhost:4000/callback&
  scope=openid profile&
  state=xyz789
Cookie: oauth_sso_session=SESSION_ID

# Step 2: SSO session validated ✓
# Step 3: Check consent for App B → Not found
# Step 4: Show consent screen

# Step 5: User approves consent for App B
POST http://localhost:8080/oauth/consent
Content-Type: application/x-www-form-urlencoded

approved=true&
client_id=app-b&
scope=openid profile&
state=xyz789&
redirect_uri=http://localhost:4000/callback

# Step 6: Authorization code generated immediately
HTTP/1.1 302 Found
Location: http://localhost:4000/callback?code=AUTH_CODE_2&state=xyz789
```

**Result**:
- No login required (SSO session valid)
- Consent screen shown once for App B
- Authorization code issued

### Example 3: Returning to App A (Full Auto-Approval)

**Scenario**: User returns to App A after some time.

```bash
# Step 1: Client initiates authorization
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile email&
  state=new_state_456
Cookie: oauth_sso_session=SESSION_ID

# Step 2: SSO session validated ✓
# Step 3: Consent exists for App A ✓
# Step 4: Authorization code generated immediately (< 100ms)

HTTP/1.1 302 Found
Location: http://localhost:3000/callback?code=AUTH_CODE_3&state=new_state_456
```

**Result**:
- No login screen (SSO session valid)
- No consent screen (consent already granted)
- Instant authorization code generation

### Example 4: Logout

**Scenario**: User logs out from any application.

```bash
# Step 1: User clicks logout
POST http://localhost:8080/auth/logout
Cookie: oauth_sso_session=SESSION_ID

# Response
HTTP/1.1 200 OK
Set-Cookie: oauth_sso_session=; Max-Age=-1

{
  "message": "Logged out successfully"
}

# Step 2: User tries to access App A again
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile email&
  state=after_logout

# Step 3: No SSO session → Redirect to login
HTTP/1.1 302 Found
Location: /auth/login?session_id=NEW_SESSION_ID
```

**Result**:
- SSO session deleted from database
- Cookie cleared
- Next access requires re-authentication

### Example 5: Logout with Redirect

**Scenario**: OIDC-compliant logout with redirect.

```bash
POST http://localhost:8080/auth/logout?
  post_logout_redirect_uri=https://example.com/goodbye
Cookie: oauth_sso_session=SESSION_ID

# Response
HTTP/1.1 302 Found
Location: https://example.com/goodbye
Set-Cookie: oauth_sso_session=; Max-Age=-1
```

## Session Management

### List Active Sessions

View all active SSO sessions for the authenticated user.

```bash
GET http://localhost:8080/account/sessions
Authorization: Bearer ACCESS_TOKEN

# Response
{
  "sessions": [
    {
      "session_id": "abc123def456...",
      "created_at": "2025-11-09T10:00:00Z",
      "last_activity": "2025-11-09T15:30:00Z",
      "expires_at": "2025-11-16T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)..."
    },
    {
      "session_id": "xyz789ghi012...",
      "created_at": "2025-11-08T14:00:00Z",
      "last_activity": "2025-11-09T09:00:00Z",
      "expires_at": "2025-11-15T14:00:00Z",
      "ip_address": "192.168.1.50",
      "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)..."
    }
  ]
}
```

**Use Cases**:
- Security audit: See where you're logged in
- Detect suspicious sessions (unknown IP/device)
- Monitor session activity

### Revoke Specific Session

Revoke a specific SSO session (e.g., logout from a specific device).

```bash
DELETE http://localhost:8080/account/sessions/abc123def456...
Authorization: Bearer ACCESS_TOKEN

# Response
{
  "message": "Session revoked successfully"
}
```

**Use Cases**:
- Logout from a specific device
- Revoke suspicious session
- Remote logout (e.g., lost phone)

### Example: Security Audit Workflow

```bash
# 1. List all sessions
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 2. Identify suspicious session (unknown IP)
# Session ID: xyz789ghi012... from IP 203.0.113.42

# 3. Revoke suspicious session
curl -X DELETE http://localhost:8080/account/sessions/xyz789ghi012... \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 4. Verify session removed
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Authorization Management

### List Authorized Applications

View all applications that have been granted access to your account.

```bash
GET http://localhost:8080/account/authorizations
Authorization: Bearer ACCESS_TOKEN

# Response
{
  "authorizations": [
    {
      "client_id": "app-a",
      "client_name": "My Application A",
      "scopes": ["openid", "profile", "email"],
      "granted_at": "2025-11-01T10:00:00Z",
      "expires_at": "2026-11-01T10:00:00Z"
    },
    {
      "client_id": "app-b",
      "client_name": "My Application B",
      "scopes": ["openid", "profile"],
      "granted_at": "2025-11-05T14:30:00Z",
      "expires_at": "2026-11-05T14:30:00Z"
    }
  ]
}
```

**Use Cases**:
- Review which apps have access to your data
- Audit application permissions
- Identify unused applications

### Revoke Application Authorization

Revoke consent for a specific application.

```bash
DELETE http://localhost:8080/account/authorizations/app-b
Authorization: Bearer ACCESS_TOKEN

# Response
{
  "message": "Authorization revoked successfully"
}
```

**Result**:
- Consent record deleted
- Next authorization request will show consent screen
- Existing access tokens remain valid until expiration

### Example: Privacy Management Workflow

```bash
# 1. List all authorized applications
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 2. Identify unused application
# Client ID: old-app (last used 6 months ago)

# 3. Revoke authorization
curl -X DELETE http://localhost:8080/account/authorizations/old-app \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 4. User tries to access old-app
# → Consent screen shown (authorization revoked)

# 5. Verify authorization removed
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## OIDC Prompt Parameter

The `prompt` parameter controls authentication and consent behavior, following the OIDC specification.

### prompt=none

**Behavior**: Fail immediately if user is not authenticated or consent is missing.

```bash
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile&
  state=abc123&
  prompt=none

# Case 1: User not authenticated
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?
  error=login_required&
  error_description=User+authentication+required&
  state=abc123

# Case 2: User authenticated but no consent
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?
  error=consent_required&
  error_description=User+consent+required&
  state=abc123

# Case 3: User authenticated and consent exists
HTTP/1.1 302 Found
Location: http://localhost:3000/callback?code=AUTH_CODE&state=abc123
```

**Use Cases**:
- Silent token refresh in iframe
- Check authentication status without user interaction
- Background authorization checks

### prompt=login

**Behavior**: Force re-authentication even if SSO session exists.

```bash
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile&
  state=xyz789&
  prompt=login
Cookie: oauth_sso_session=VALID_SESSION

# Result: SSO session ignored, login page shown
HTTP/1.1 302 Found
Location: /auth/login?session_id=NEW_SESSION_ID
```

**Use Cases**:
- High-security operations (e.g., financial transactions)
- Account switching
- Verify user identity before sensitive action

### prompt=consent

**Behavior**: Force consent screen even if consent was previously granted.

```bash
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile email&
  state=consent123&
  prompt=consent
Cookie: oauth_sso_session=VALID_SESSION

# Result: Consent screen shown even if consent exists
HTTP/1.1 302 Found
Location: /oauth/consent?client_id=app-a&scope=openid+profile+email&...
```

**Use Cases**:
- Request additional permissions
- Re-confirm user consent
- Compliance requirements (explicit consent)

### prompt=select_account

**Behavior**: Display account selection screen (placeholder for future implementation).

```bash
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile&
  state=select123&
  prompt=select_account

# Current behavior: Treated as normal flow
# Future: Show account selection UI
```

**Use Cases**:
- Multi-account support
- Allow user to choose which account to use
- Switch between personal and work accounts

### Combining Prompt Values

Multiple prompt values can be combined with spaces:

```bash
# Force login and consent
prompt=login consent

# Example
GET http://localhost:8080/oauth/authorize?
  response_type=code&
  client_id=app-a&
  redirect_uri=http://localhost:3000/callback&
  scope=openid profile&
  state=combined123&
  prompt=login%20consent
```

## Configuration

### Environment Variables

```bash
# SSO Session Configuration
SSO_SESSION_EXPIRY_DAYS=7          # SSO session lifetime (default: 7 days)
SSO_CONSENT_EXPIRY_DAYS=365        # Consent lifetime (default: 1 year)

# Cookie Security
SSO_COOKIE_SECURE=true             # Require HTTPS (set to false for local dev)
SSO_COOKIE_DOMAIN=                 # Cookie domain (empty = current domain)
SSO_COOKIE_PATH=/                  # Cookie path (default: /)

# Future Enhancements
SSO_MAX_CONCURRENT_SESSIONS=5      # Max sessions per user
SSO_ACTIVITY_TIMEOUT_MINUTES=30    # Inactivity timeout
```

### Cookie Configuration

The SSO cookie is configured with the following security settings:

```go
Cookie Name:     oauth_sso_session
Max Age:         604800 seconds (7 days)
Path:            /
HttpOnly:        true              // Prevent JavaScript access
Secure:          true              // HTTPS only (production)
SameSite:        Lax               // CSRF protection
```

### Database Indexes

The following MongoDB indexes are created automatically:

```javascript
// sso_sessions collection
db.sso_sessions.createIndex({ "session_id": 1 }, { unique: true })
db.sso_sessions.createIndex({ "user_id": 1 })
db.sso_sessions.createIndex({ "expires_at": 1 })

// user_consents collection
db.user_consents.createIndex({ "user_id": 1, "client_id": 1 }, { unique: true })
db.user_consents.createIndex({ "user_id": 1 })
```

## Security Considerations

### Session Security

1. **Secure Storage**: Sessions stored server-side, not in cookie
2. **Random Session IDs**: 32-byte cryptographically random strings
3. **HTTP-Only Cookies**: Prevent XSS attacks
4. **Secure Flag**: HTTPS-only in production
5. **SameSite Protection**: Mitigate CSRF attacks

### Session Fingerprinting

Each SSO session includes:
- **IP Address**: Client IP at session creation
- **User Agent**: Browser/device information

**Future Enhancement**: Detect session hijacking by comparing fingerprints on each request.

### Consent Security

1. **Scope Validation**: Requested scopes validated against stored consent
2. **Expiration**: 1-year maximum lifetime
3. **Revocation**: Users can revoke consent at any time
4. **Audit Trail**: All consent grants and revocations logged

### Best Practices

1. **Use HTTPS**: Always use HTTPS in production
2. **Monitor Sessions**: Regularly review active sessions
3. **Revoke Unused Apps**: Remove authorization for unused applications
4. **Logout on Shared Devices**: Always logout on public/shared computers
5. **Review Permissions**: Periodically audit application permissions

### Security Checklist

- [ ] HTTPS enabled in production
- [ ] `SSO_COOKIE_SECURE=true` in production
- [ ] Regular session cleanup (expired sessions)
- [ ] Monitor for suspicious activity (IP changes, etc.)
- [ ] User education on session management
- [ ] Implement rate limiting on login endpoint
- [ ] Enable audit logging for security events

## Troubleshooting

### SSO Session Not Working

**Symptoms**: User redirected to login despite recent authentication.

**Possible Causes**:
1. Cookie not set (check browser dev tools)
2. Session expired (> 7 days old)
3. Session deleted from database
4. Cookie domain mismatch
5. Secure flag issue (HTTPS required)

**Solutions**:
```bash
# Check if cookie exists
# Browser Dev Tools → Application → Cookies → oauth_sso_session

# Check session in database
db.sso_sessions.findOne({ session_id: "SESSION_ID" })

# Verify session not expired
db.sso_sessions.findOne({ 
  session_id: "SESSION_ID",
  expires_at: { $gt: new Date() }
})
```

### Consent Not Remembered

**Symptoms**: Consent screen shown every time.

**Possible Causes**:
1. Consent not saved to database
2. Scope mismatch (requesting different scopes)
3. Consent expired (> 1 year old)
4. `prompt=consent` parameter used

**Solutions**:
```bash
# Check consent in database
db.user_consents.findOne({ 
  user_id: "USER_ID",
  client_id: "CLIENT_ID"
})

# Verify scopes match
# Requested: ["openid", "profile", "email"]
# Stored: ["openid", "profile"]  ← Missing "email"
```

### Auto-Approval Not Working

**Symptoms**: Login or consent screen shown despite valid SSO and consent.

**Possible Causes**:
1. `prompt=login` or `prompt=consent` parameter
2. SSO session invalid
3. Consent missing or expired
4. Scope mismatch

**Debug Steps**:
```bash
# 1. Check SSO session
curl -X GET http://localhost:8080/account/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 2. Check consent
curl -X GET http://localhost:8080/account/authorizations \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Verify no prompt parameter in authorization URL
# Bad:  prompt=login
# Good: (no prompt parameter)
```

## Examples

### Complete SSO Flow (cURL)

```bash
#!/bin/bash

# Configuration
SERVER="http://localhost:8080"
CLIENT_ID="test-app"
CLIENT_SECRET="test-secret"
REDIRECT_URI="http://localhost:3000/callback"

# Step 1: Register client (one-time)
curl -X POST $SERVER/clients/register \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Application\",
    \"redirect_uris\": [\"$REDIRECT_URI\"]
  }"

# Step 2: Start authorization (browser would follow redirects)
# This will redirect to login page
curl -v "$SERVER/oauth/authorize?response_type=code&client_id=$CLIENT_ID&redirect_uri=$REDIRECT_URI&scope=openid%20profile%20email&state=random123"

# Step 3: Login (in real flow, user fills form)
# Note: Extract session_id from login page redirect
SESSION_ID="extracted_from_redirect"

curl -X POST $SERVER/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d "{
    \"email\": \"user@example.com\",
    \"password\": \"password123\",
    \"session_id\": \"$SESSION_ID\"
  }"

# Step 4: Approve consent (in real flow, user clicks button)
curl -X POST $SERVER/oauth/consent \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -b cookies.txt \
  -d "approved=true&client_id=$CLIENT_ID&scope=openid%20profile%20email&state=random123&redirect_uri=$REDIRECT_URI"

# Step 5: Extract authorization code from redirect
# Location: http://localhost:3000/callback?code=AUTH_CODE&state=random123

# Step 6: Exchange code for tokens
curl -X POST $SERVER/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&redirect_uri=$REDIRECT_URI"

# Step 7: Access second app (auto-approval with SSO)
curl -v "$SERVER/oauth/authorize?response_type=code&client_id=app-b&redirect_uri=http://localhost:4000/callback&scope=openid%20profile&state=xyz789" \
  -b cookies.txt

# Result: Immediate redirect with authorization code (no login/consent)
```

## Additional Resources

- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [OIDC Prompt Parameter](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest)
- [OAuth2 Security Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)
