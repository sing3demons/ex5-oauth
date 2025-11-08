# OAuth2 Scope Management

## Overview

Scope ใช้ในการควบคุมสิทธิ์การเข้าถึงข้อมูลและ API ตาม OAuth2/OIDC specification

## Supported Scopes

| Scope | Description | Claims Included |
|-------|-------------|-----------------|
| `openid` | **Required** สำหรับ OIDC authentication | `sub` (user ID) |
| `profile` | ข้อมูลโปรไฟล์พื้นฐาน | `name`, `picture` |
| `email` | ข้อมูลอีเมล | `email`, `email_verified` |
| `phone` | ข้อมูลเบอร์โทรศัพท์ | `phone_number`, `phone_number_verified` |
| `address` | ข้อมูลที่อยู่ | `address` |
| `offline_access` | สำหรับขอ refresh token | - |

## Scope Validation

### 1. Validate Scope
ตรวจสอบว่า scope ที่ขอมาถูกต้องหรือไม่

```go
import "oauth2-server/utils"

scope := "openid profile email"
if !utils.ValidateScope(scope) {
    // Invalid scope
}
```

**Rules:**
- ต้องมี `openid` scope เสมอ (OIDC requirement)
- Scope ที่ไม่รู้จักจะถือว่า invalid
- Empty scope จะถือว่า invalid

### 2. Normalize Scope
ลบ duplicates และ invalid scopes

```go
scope := "openid openid profile invalid_scope email"
normalized := utils.NormalizeScope(scope)
// Result: "openid profile email"
```

**Features:**
- ลบ scope ซ้ำ
- ลบ scope ที่ไม่ valid
- Trim whitespace
- รักษาลำดับของ scope

### 3. Check Specific Scope
ตรวจสอบว่ามี scope เฉพาะหรือไม่

```go
scope := "openid profile email"

if utils.HasScope(scope, "email") {
    // Has email scope
}

// Helper functions
utils.ScopeIncludesProfile(scope)  // true
utils.ScopeIncludesEmail(scope)    // true
utils.ScopeIncludesOfflineAccess(scope) // false
```

### 4. Intersect Scopes
หา scope ที่ตรงกันระหว่าง requested และ allowed

```go
requested := "openid profile email phone"
allowed := "openid profile email"

result := utils.IntersectScopes(requested, allowed)
// Result: "openid profile email"
```

**Use Case:** เมื่อ client ขอ scope มากกว่าที่ได้รับอนุญาต

## API Endpoints with Scope

### 1. Authorization Endpoint
```bash
GET /oauth/authorize?
  response_type=code&
  client_id=<client_id>&
  redirect_uri=<redirect_uri>&
  scope=openid%20profile%20email&
  state=xyz
```

**Scope Handling:**
- ถ้าไม่ระบุ scope → ใช้ default: `openid profile email`
- ถ้าระบุ scope → validate และ normalize
- ต้องมี `openid` scope เสมอ

**Error Response:**
```json
{
  "error": "invalid_scope",
  "error_description": "Invalid scope requested"
}
```

### 2. Token Endpoint

#### Authorization Code Grant
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=<auth_code>&
redirect_uri=<redirect_uri>&
client_id=<client_id>&
client_secret=<client_secret>
```

Scope จะมาจาก authorization code ที่ได้รับอนุญาตไว้แล้ว

#### Refresh Token Grant
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
refresh_token=<refresh_token>&
client_id=<client_id>&
client_secret=<client_secret>&
scope=openid%20email
```

**Scope Handling:**
- ถ้าไม่ระบุ scope → ใช้ default
- ถ้าระบุ scope → validate และ normalize
- สามารถขอ scope น้อยกว่าเดิมได้ แต่ไม่สามารถขอเพิ่มได้

#### Client Credentials Grant
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&
client_id=<client_id>&
client_secret=<client_secret>&
scope=openid
```

**Scope Handling:**
- ถ้าไม่ระบุ scope → ใช้ `openid` (minimal)
- Client credentials ไม่ควรขอ `profile` หรือ `email` เพราะไม่มี user context

### 3. UserInfo Endpoint
```bash
GET /oauth/userinfo
Authorization: Bearer <access_token>
```

**Response based on Scope:**

**Scope: `openid`**
```json
{
  "sub": "user123"
}
```

**Scope: `openid profile`**
```json
{
  "sub": "user123",
  "name": "John Doe"
}
```

**Scope: `openid profile email`**
```json
{
  "sub": "user123",
  "name": "John Doe",
  "email": "john@example.com"
}
```

### 4. Token Exchange
```bash
POST /token/exchange
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:token-exchange&
subject_token=<existing_token>&
subject_token_type=urn:ietf:params:oauth:token-type:access_token&
client_id=<client_id>&
client_secret=<client_secret>&
scope=openid%20email&
is_encrypted_jwe=true
```

**Scope Handling:**
- ถ้าไม่ระบุ scope → ใช้ default
- ถ้าระบุ scope → validate และ normalize
- สามารถเปลี่ยน scope ได้ตามที่ client มีสิทธิ์

## Token Claims by Scope

### Access Token Claims
```json
{
  "sub": "user123",
  "scope": "openid profile email",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Note:** Access token ไม่มี user info ละเอียด เก็บแค่ scope

### ID Token Claims

**Minimal (openid only):**
```json
{
  "sub": "user123",
  "aud": "client_id",
  "iss": "oauth2-server",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**With profile scope:**
```json
{
  "sub": "user123",
  "name": "John Doe",
  "aud": "client_id",
  "iss": "oauth2-server",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**With email scope:**
```json
{
  "sub": "user123",
  "email": "john@example.com",
  "email_verified": true,
  "aud": "client_id",
  "iss": "oauth2-server",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Best Practices

### 1. Request Minimal Scopes
ขอเฉพาะ scope ที่จำเป็นเท่านั้น

```bash
# Good - ขอเฉพาะที่ต้องการ
scope=openid email

# Bad - ขอทุกอย่าง
scope=openid profile email phone address
```

### 2. Validate Scope in API
ตรวจสอบ scope ก่อนให้บริการ

```go
func ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
    token := extractToken(r)
    claims, _ := utils.ValidateToken(token, publicKey)
    
    // ต้องมี email scope
    if !utils.ScopeIncludesEmail(claims.Scope) {
        respondError(w, http.StatusForbidden, "insufficient_scope", 
            "Email scope required")
        return
    }
    
    // Continue...
}
```

### 3. Use Scope Downgrade
ลด scope เมื่อ refresh token

```bash
# Original scope
scope=openid profile email phone

# Refresh with reduced scope
scope=openid email
```

### 4. Separate Access Token and ID Token
- **Access Token**: ใช้สำหรับ API authorization (มีแค่ scope)
- **ID Token**: ใช้สำหรับ user authentication (มี user info)

## Error Handling

### Invalid Scope Error
```json
{
  "error": "invalid_scope",
  "error_description": "Invalid scope requested"
}
```

**Causes:**
- Scope ที่ไม่รู้จัก
- ไม่มี `openid` scope (OIDC requirement)
- Format ไม่ถูกต้อง

### Insufficient Scope Error
```json
{
  "error": "insufficient_scope",
  "error_description": "The request requires higher privileges"
}
```

**Causes:**
- Token ไม่มี scope ที่จำเป็นสำหรับ API
- ใช้ในการป้องกัน API endpoints

## Testing Scope

### Test Invalid Scope
```bash
curl -X GET "http://localhost:8080/oauth/authorize?\
response_type=code&\
client_id=test&\
redirect_uri=http://localhost:3000/callback&\
scope=invalid_scope&\
state=xyz"

# Expected: invalid_scope error
```

### Test Scope Filtering in UserInfo
```bash
# Get token with only openid scope
TOKEN=$(curl -s -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code&..." | jq -r .access_token)

# UserInfo should only return sub
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/oauth/userinfo

# Response: {"sub":"user123"}
```

### Test Scope Normalization
```bash
# Request with duplicates and invalid scopes
scope="openid openid profile invalid email"

# Server will normalize to: "openid profile email"
```

## Code Examples

### Add Custom Scope
```go
// In utils/scope.go
var ValidScopes = map[string]bool{
    "openid":  true,
    "profile": true,
    "email":   true,
    "phone":   true,
    "address": true,
    "offline_access": true,
    "custom_scope": true,  // Add your custom scope
}
```

### Protect API with Scope
```go
func RequireScope(requiredScope string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            claims, err := utils.ValidateToken(token, publicKey)
            if err != nil {
                respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid token")
                return
            }
            
            if !utils.HasScope(claims.Scope, requiredScope) {
                respondError(w, http.StatusForbidden, "insufficient_scope", 
                    fmt.Sprintf("Requires %s scope", requiredScope))
                return
            }
            
            next(w, r)
        }
    }
}

// Usage
r.HandleFunc("/api/profile", RequireScope("profile")(ProfileHandler))
r.HandleFunc("/api/email", RequireScope("email")(EmailHandler))
```

## References

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749#section-3.3)
- [OpenID Connect Core 1.0 - Scope Values](https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims)
