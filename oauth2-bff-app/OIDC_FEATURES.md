# OpenID Connect (OIDC) Features

## Overview

This BFF implementation is fully OIDC compliant with the following features:

## ✅ Implemented OIDC Features

### 1. **Authorization Code Flow with OIDC**
- ✅ `response_type=code`
- ✅ `scope=openid profile email`
- ✅ `nonce` parameter for replay protection
- ✅ `state` parameter for CSRF protection

### 2. **ID Token Validation**
- ✅ Issuer (`iss`) validation
- ✅ Audience (`aud`) validation
- ✅ Expiration (`exp`) validation
- ✅ Issued at (`iat`) validation
- ✅ Nonce validation
- ✅ Signature validation (via OAuth2 server)

### 3. **Token Management**
- ✅ Access Token (short-lived, memory-only)
- ✅ Refresh Token (long-lived, HttpOnly cookie)
- ✅ ID Token (identity claims)
- ✅ Token rotation on refresh

### 4. **UserInfo Endpoint**
- ✅ Fetch user claims from `/oauth/userinfo`
- ✅ Bearer token authentication
- ✅ Scope-based claim filtering

### 5. **Discovery**
- ✅ OIDC Discovery endpoint support
- ✅ Dynamic configuration

## 🔐 Security Features

### ID Token Validation Flow

```typescript
// 1. Receive ID Token from OAuth2 server
const tokens = await exchangeCodeForTokens(code);

// 2. Validate ID Token
const validation = validateIDToken(
  tokens.id_token,
  CLIENT_ID,
  OAUTH2_SERVER,
  nonce  // From session
);

// 3. Check validation result
if (!validation.valid) {
  throw new Error(validation.error);
}

// 4. Use validated claims
const userClaims = validation.claims;
```

### Nonce Flow

```
1. BFF generates random nonce
2. BFF stores nonce in session
3. BFF sends nonce in authorization request
4. OAuth2 server includes nonce in ID Token
5. BFF validates nonce matches session
6. Prevents replay attacks
```

## 📡 API Endpoints

### Authentication Endpoints

#### `GET /auth/login`
Initiate OIDC login flow
```bash
curl http://localhost:3001/auth/login
```

Response:
```json
{
  "authorization_url": "http://localhost:8080/oauth/authorize?..."
}
```

#### `GET /auth/callback`
Handle OAuth2 callback and validate ID Token
```
http://localhost:3001/auth/callback?code=xxx&state=yyy
```

#### `POST /auth/refresh`
Refresh access token using refresh token cookie
```bash
curl -X POST http://localhost:3001/auth/refresh \
  -H "Cookie: refresh_token=xxx"
```

#### `POST /auth/logout`
Clear refresh token cookie
```bash
curl -X POST http://localhost:3001/auth/logout \
  -H "Cookie: refresh_token=xxx"
```

#### `GET /auth/userinfo`
Get user information
```bash
curl http://localhost:3001/auth/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### OIDC Utility Endpoints

#### `GET /auth/discovery`
Get OIDC discovery document
```bash
curl http://localhost:3001/auth/discovery
```

Response:
```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/oauth/authorize",
  "token_endpoint": "http://localhost:8080/oauth/token",
  "userinfo_endpoint": "http://localhost:8080/oauth/userinfo",
  "jwks_uri": "http://localhost:8080/.well-known/jwks.json",
  ...
}
```

#### `POST /auth/validate-token`
Validate and decode ID Token
```bash
curl -X POST http://localhost:3001/auth/validate-token \
  -H "Content-Type: application/json" \
  -d '{"id_token": "eyJhbGc..."}'
```

Response:
```json
{
  "valid": true,
  "claims": {
    "iss": "http://localhost:8080",
    "sub": "user123",
    "aud": "client_id",
    "exp": 1234567890,
    "iat": 1234567800,
    "nonce": "abc123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

#### `POST /auth/decode-token`
Decode JWT without validation (debugging)
```bash
curl -X POST http://localhost:3001/auth/decode-token \
  -H "Content-Type: application/json" \
  -d '{"token": "eyJhbGc..."}'
```

#### `GET /auth/session`
Get current session info
```bash
curl http://localhost:3001/auth/session \
  -H "Cookie: refresh_token=xxx"
```

Response:
```json
{
  "has_refresh_token": true,
  "expires_at": "2024-12-31T23:59:59.000Z",
  "user_id": "user123",
  "scope": "openid profile email"
}
```

## 🔍 ID Token Claims

### Standard Claims
- `iss` - Issuer (OAuth2 server URL)
- `sub` - Subject (user ID)
- `aud` - Audience (client ID)
- `exp` - Expiration time
- `iat` - Issued at time
- `nonce` - Nonce for replay protection

### Profile Claims (scope: profile)
- `name` - Full name
- `given_name` - First name
- `family_name` - Last name
- `middle_name` - Middle name
- `nickname` - Nickname
- `preferred_username` - Username
- `profile` - Profile page URL
- `picture` - Profile picture URL
- `website` - Website URL
- `gender` - Gender
- `birthdate` - Birthday
- `zoneinfo` - Timezone
- `locale` - Locale
- `updated_at` - Last update time

### Email Claims (scope: email)
- `email` - Email address
- `email_verified` - Email verification status

### Phone Claims (scope: phone)
- `phone_number` - Phone number
- `phone_number_verified` - Phone verification status

### Address Claims (scope: address)
- `address` - Postal address object

## 🧪 Testing OIDC Features

### Test ID Token Validation

```bash
# 1. Login and get ID token
curl http://localhost:3001/auth/login

# 2. Complete OAuth flow (browser)
# Get id_token from callback

# 3. Validate ID token
curl -X POST http://localhost:3001/auth/validate-token \
  -H "Content-Type: application/json" \
  -d '{"id_token": "YOUR_ID_TOKEN"}'
```

### Test Nonce Protection

```bash
# Try to replay an old ID token with different nonce
# Should fail validation
```

### Test Token Refresh

```bash
# 1. Get initial tokens
# 2. Wait for access token to expire
# 3. Refresh should work automatically
curl -X POST http://localhost:3001/auth/refresh \
  -H "Cookie: refresh_token=xxx"
```

## 📚 OIDC Specifications

This implementation follows:
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OAuth 2.0 Token Introspection RFC 7662](https://tools.ietf.org/html/rfc7662)

## 🔄 Token Lifecycle

```
┌─────────────────────────────────────────────────────────┐
│                    Token Lifecycle                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  1. Login                                                │
│     ├─ Generate nonce                                    │
│     ├─ Redirect to OAuth2 server                         │
│     └─ Store nonce in session                            │
│                                                          │
│  2. Callback                                             │
│     ├─ Exchange code for tokens                          │
│     ├─ Validate ID Token (iss, aud, exp, nonce)         │
│     ├─ Store refresh_token in HttpOnly cookie           │
│     └─ Return access_token to frontend                   │
│                                                          │
│  3. Use Access Token                                     │
│     ├─ Frontend stores in memory                         │
│     ├─ Send with API requests                            │
│     └─ Auto-refresh before expiry                        │
│                                                          │
│  4. Refresh                                              │
│     ├─ Use refresh_token from cookie                     │
│     ├─ Get new access_token                              │
│     ├─ Rotate refresh_token (optional)                   │
│     └─ Update cookie                                     │
│                                                          │
│  5. Logout                                               │
│     ├─ Clear refresh_token cookie                        │
│     ├─ Clear frontend memory                             │
│     └─ Optionally revoke tokens                          │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 🎯 Best Practices

1. ✅ **Always validate ID tokens** - Don't trust without verification
2. ✅ **Use nonce** - Prevents replay attacks
3. ✅ **Validate all claims** - iss, aud, exp, iat, nonce
4. ✅ **Store refresh tokens securely** - HttpOnly cookies
5. ✅ **Never expose client_secret** - Keep on server only
6. ✅ **Use HTTPS in production** - Protect tokens in transit
7. ✅ **Implement token rotation** - Limit damage from compromise
8. ✅ **Set appropriate token lifetimes** - Balance security and UX
9. ✅ **Log security events** - Monitor for suspicious activity
10. ✅ **Handle errors gracefully** - Don't leak sensitive info

## 🚀 Production Considerations

### Token Storage
- ✅ Use Redis for session storage (not in-memory)
- ✅ Encrypt refresh tokens at rest
- ✅ Set appropriate cookie attributes

### Monitoring
- ✅ Log all authentication events
- ✅ Monitor failed validation attempts
- ✅ Alert on suspicious patterns
- ✅ Track token usage metrics

### Performance
- ✅ Cache JWKS keys
- ✅ Cache discovery document
- ✅ Use connection pooling
- ✅ Implement rate limiting

### Security
- ✅ Rotate client secrets regularly
- ✅ Implement token revocation
- ✅ Use short-lived access tokens
- ✅ Implement account lockout
- ✅ Add MFA support
