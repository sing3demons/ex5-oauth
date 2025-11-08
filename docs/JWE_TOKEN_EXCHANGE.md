# JWE & Token Exchange Implementation

## Overview

ระบบรองรับ 2 รูปแบบของ token:
- **JWT (JSON Web Token)**: Signed token ที่สามารถ verify ได้ แต่ payload อ่านได้
- **JWE (JSON Web Encryption)**: Encrypted token ที่ payload ถูกเข้ารหัส อ่านไม่ได้โดยไม่มี private key

## Token Types และ Claims

### Access Token
**JWT Claims:**
```json
{
  "sub": "user_id",
  "scope": "openid profile email",
  "client_id": "optional",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**JWE Claims:** (เหมือน JWT แต่ encrypted)

**วัตถุประสงค์:** ใช้สำหรับ authorization เข้าถึง API

### ID Token
**JWT Claims:**
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "email_verified": true,
  "name": "User Name",
  "picture": "optional",
  "nonce": "optional",
  "aud": "client_id",
  "iss": "oauth2-server",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**JWE Claims:** (เหมือน JWT แต่ encrypted)

**วัตถุประสงค์:** ใช้สำหรับ authentication ระบุตัวตนผู้ใช้

### Refresh Token
**Claims:**
```json
{
  "sub": "user_id",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## API Endpoints

### 1. Token Exchange (RFC 8693)
**Endpoint:** `POST /token/exchange`

**Purpose:** แลกเปลี่ยน token จาก JWT เป็น JWE หรือ JWE เป็น JWT

**Request:**
```bash
curl -X POST http://localhost:8080/token/exchange \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=<existing_token>" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=<client_id>" \
  -d "client_secret=<client_secret>" \
  -d "is_encrypted_jwe=true"
```

**Parameters:**
- `grant_type`: ต้องเป็น `urn:ietf:params:oauth:grant-type:token-exchange`
- `subject_token`: Token ที่ต้องการแลกเปลี่ยน (JWT หรือ JWE)
- `subject_token_type`: ประเภทของ token
  - `urn:ietf:params:oauth:token-type:access_token`
  - `urn:ietf:params:oauth:token-type:refresh_token`
  - `urn:ietf:params:oauth:token-type:id_token`
- `client_id`: Client ID
- `client_secret`: Client Secret
- `is_encrypted_jwe`: `true` = สร้าง JWE, `false` หรือไม่ระบุ = สร้าง JWT
- `scope`: (optional) scope ที่ต้องการ

**Response:**
```json
{
  "access_token": "eyJ...",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJ...",
  "id_token": "eyJ...",
  "scope": "openid profile email"
}
```

### 2. Token Validation
**Endpoint:** `POST /token/validate` หรือ `GET /token/validate`

**Purpose:** ตรวจสอบความถูกต้องของ JWT หรือ JWE token

**Request (POST):**
```bash
curl -X POST http://localhost:8080/token/validate \
  -H "Content-Type: application/json" \
  -d '{"token":"<jwt_or_jwe_token>"}'
```

**Request (GET):**
```bash
curl "http://localhost:8080/token/validate?token=<jwt_or_jwe_token>"
```

**Request (Authorization Header):**
```bash
curl -X GET http://localhost:8080/token/validate \
  -H "Authorization: Bearer <jwt_or_jwe_token>"
```

**Response (Valid JWT):**
```json
{
  "valid": true,
  "token_type": "JWT",
  "claims": {
    "sub": "user123",
    "scope": "openid profile email",
    "email": "",
    "name": ""
  },
  "expires_at": 1234567890,
  "issued_at": 1234567890
}
```

**Response (Valid JWE):**
```json
{
  "valid": true,
  "token_type": "JWE",
  "claims": {
    "sub": "user123",
    "email": "user@example.com",
    "name": "User Name",
    "scope": "openid profile email",
    "aud": "client_id"
  },
  "expires_at": 1234567890,
  "issued_at": 1234567890
}
```

**Response (Invalid):**
```json
{
  "valid": false,
  "error": "token expired"
}
```

## JWE Implementation Details

### Encryption Algorithm
- **Key Encryption:** RSA-OAEP (RSA with SHA-256)
- **Content Encryption:** AES-256-GCM

### JWE Compact Serialization Format
```
header.encrypted_key.iv.ciphertext.
```

5 ส่วนคั่นด้วย `.` (JWT มี 3 ส่วน)

### Token Detection
```go
// ตรวจสอบว่าเป็น JWE หรือ JWT
if utils.IsJWE(token) {
    // 5 parts = JWE
} else if utils.IsJWT(token) {
    // 3 parts = JWT
}
```

## Use Cases

### 1. แลกเปลี่ยน JWT เป็น JWE (เพิ่มความปลอดภัย)
```bash
# ได้ JWT มาจาก OAuth flow
ACCESS_TOKEN="eyJhbGc..."

# แลกเป็น JWE
curl -X POST http://localhost:8080/token/exchange \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$ACCESS_TOKEN" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "is_encrypted_jwe=true"
```

### 2. แลกเปลี่ยน JWE เป็น JWT (เพื่อ interoperability)
```bash
# ได้ JWE มา
JWE_TOKEN="eyJhbGc..."

# แลกเป็น JWT
curl -X POST http://localhost:8080/token/exchange \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=$JWE_TOKEN" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "is_encrypted_jwe=false"
```

### 3. Validate Token ก่อนใช้งาน
```bash
# ตรวจสอบ token ก่อนเรียก API
curl -X POST http://localhost:8080/token/validate \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}"
```

## Security Considerations

### JWT
- ✅ Signature verification ป้องกันการปลอมแปลง
- ⚠️ Payload อ่านได้ (base64 decode)
- ✅ เหมาะสำหรับ public claims
- ✅ Stateless verification

### JWE
- ✅ Payload encrypted ไม่สามารถอ่านได้
- ✅ เหมาะสำหรับ sensitive data
- ⚠️ ต้องมี private key ในการ decrypt
- ⚠️ Performance overhead จากการ encrypt/decrypt

## Testing

รัน test script:
```bash
./test_jwe_token_exchange.sh
```

Test จะทำการ:
1. Register client
2. Register user
3. Login
4. Get authorization code
5. Exchange code for JWT tokens
6. Validate JWT tokens
7. Exchange JWT to JWE
8. Validate JWE tokens
9. Compare token formats

## Code Examples

### Generate JWE Token
```go
import "oauth2-server/utils"

// Generate JWE access token
token, err := utils.GenerateJWEAccessToken(
    userID,
    email,
    name,
    scope,
    publicKey,
    expiryTimestamp,
)
```

### Validate JWE Token
```go
// Validate and decrypt JWE
claims, err := utils.ValidateJWE(jweToken, privateKey)
if err != nil {
    // Invalid or expired
}

// Use claims
userID := claims.UserID
email := claims.Email
```

### Check Token Type
```go
if utils.IsJWE(token) {
    // Handle JWE
    claims, _ := utils.ValidateJWE(token, privateKey)
} else if utils.IsJWT(token) {
    // Handle JWT
    claims, _ := utils.ValidateToken(token, publicKey)
}
```

## References

- [RFC 8693 - OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693)
- [RFC 7516 - JSON Web Encryption (JWE)](https://datatracker.ietf.org/doc/html/rfc7516)
- [RFC 7519 - JSON Web Token (JWT)](https://datatracker.ietf.org/doc/html/rfc7519)
