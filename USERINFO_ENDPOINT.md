# UserInfo Endpoint Documentation

## Overview

The UserInfo endpoint (`/oauth/userinfo`) returns user claims based on the scopes granted in the access token. This implementation follows the OpenID Connect Core 1.0 specification and OAuth 2.0 standards.

## Endpoint Details

- **URL**: `/oauth/userinfo`
- **Method**: `GET`
- **Authentication**: Bearer token (JWT or JWE)

## Features

### 1. Scope-Based Claim Filtering

The endpoint filters returned claims based on the scopes present in the access token:

#### OpenID Scope Only
```bash
# Request with scope: "openid"
GET /oauth/userinfo
Authorization: Bearer <access_token>

# Response
{
  "sub": "user-123"
}
```

#### OpenID + Email Scope
```bash
# Request with scope: "openid email"
GET /oauth/userinfo
Authorization: Bearer <access_token>

# Response
{
  "sub": "user-123",
  "email": "user@example.com",
  "email_verified": true
}
```

#### OpenID + Profile Scope
```bash
# Request with scope: "openid profile"
GET /oauth/userinfo
Authorization: Bearer <access_token>

# Response
{
  "sub": "user-123",
  "name": "John Doe"
}
```

#### OpenID + Profile + Email Scope
```bash
# Request with scope: "openid profile email"
GET /oauth/userinfo
Authorization: Bearer <access_token>

# Response
{
  "sub": "user-123",
  "name": "John Doe",
  "email": "user@example.com",
  "email_verified": true
}
```

### 2. Token Format Support

The endpoint supports both JWT and JWE access tokens:

#### JWT (JSON Web Token)
- Standard signed tokens
- 3-part format: `header.payload.signature`
- Validated using RSA public key

#### JWE (JSON Web Encryption)
- Encrypted tokens for enhanced security
- 5-part format: `header.encryptedKey.iv.ciphertext.tag`
- Decrypted using RSA private key

### 3. Automatic Token Type Detection

The implementation automatically detects the token format:

```go
if utils.IsJWE(tokenString) {
    // Handle JWE token
    jweClaims, err := utils.ValidateJWE(tokenString, privateKey)
    // ...
} else {
    // Handle JWT token
    jwtClaims, err := utils.ValidateToken(tokenString, publicKey)
    // ...
}
```

## Scope to Claims Mapping

The following table shows which claims are included for each scope:

| Scope | Claims Included |
|-------|----------------|
| `openid` | `sub` |
| `profile` | `name`, `family_name`, `given_name`, `middle_name`, `nickname`, `preferred_username`, `profile`, `picture`, `website`, `gender`, `birthdate`, `zoneinfo`, `locale`, `updated_at` |
| `email` | `email`, `email_verified` |
| `phone` | `phone_number`, `phone_number_verified` |
| `address` | `address` |

**Note**: The `sub` (subject) claim is always included regardless of scopes.

## Error Responses

### Missing Authorization Header
```json
{
  "error": "unauthorized",
  "error_description": "Authorization required"
}
```
**HTTP Status**: 401 Unauthorized

### Invalid or Expired Token
```json
{
  "error": "invalid_token",
  "error_description": "Invalid or expired token"
}
```
**HTTP Status**: 401 Unauthorized

### User Not Found
```json
{
  "error": "server_error",
  "error_description": "Failed to find user"
}
```
**HTTP Status**: 500 Internal Server Error

## Implementation Details

### Claim Filtering Service

The claim filtering is handled by the `ClaimFilter` service in `utils/claims.go`:

```go
type ClaimFilter interface {
    FilterClaims(user *models.User, scopes string) map[string]interface{}
    GetIDTokenClaims(user *models.User, scopes string, nonce string) map[string]interface{}
}
```

### Scope Registry

Scopes and their associated claims are defined in the `ScopeRegistry` (`models/scope.go`):

```go
type ScopeDefinition struct {
    Name        string
    Description string
    Claims      []string
    IsDefault   bool
}
```

## Testing

### Unit Tests

Run the UserInfo endpoint tests:

```bash
go test -v -run TestUserInfo ./handlers/
```

### Test Coverage

The test suite covers:
- ✅ OpenID scope only (returns only `sub`)
- ✅ OpenID + Email scope (returns `sub`, `email`, `email_verified`)
- ✅ OpenID + Profile scope (returns `sub`, `name`)
- ✅ OpenID + Profile + Email scope (returns all claims)
- ✅ JWE token support
- ✅ JWT token support
- ✅ Unauthorized access (missing token)
- ✅ Invalid token handling

## Security Considerations

1. **Token Validation**: All tokens are validated for signature and expiration
2. **Scope Enforcement**: Claims are strictly filtered based on granted scopes
3. **Privacy Protection**: Users' data is only exposed according to authorized scopes
4. **Encryption Support**: JWE tokens provide additional security through encryption

## Compliance

This implementation complies with:
- ✅ OAuth 2.0 (RFC 6749)
- ✅ OpenID Connect Core 1.0
- ✅ Requirements 1.3 (Scope-Based Claim Filtering)
- ✅ Requirements 1.6 (Scope in Access Tokens vs ID Tokens)

## Example Usage

### Using cURL

```bash
# Get access token first (via OAuth flow)
ACCESS_TOKEN="eyJhbGc..."

# Call UserInfo endpoint
curl -X GET http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Using JavaScript

```javascript
const response = await fetch('http://localhost:8080/oauth/userinfo', {
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});

const userInfo = await response.json();
console.log(userInfo);
```

### Using Go

```go
req, _ := http.NewRequest("GET", "http://localhost:8080/oauth/userinfo", nil)
req.Header.Set("Authorization", "Bearer "+accessToken)

resp, _ := http.DefaultClient.Do(req)
defer resp.Body.Close()

var userInfo map[string]interface{}
json.NewDecoder(resp.Body).Decode(&userInfo)
```

## Related Documentation

- [Scope Management](SCOPE_MANAGEMENT.md)
- [Token Exchange](JWE_TOKEN_EXCHANGE.md)
- [OAuth2 Compliance Spec](.kiro/specs/oauth2-compliance/)
