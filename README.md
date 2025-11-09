# OAuth2/OIDC Authorization Server (Stateless)

OAuth2 และ OpenID Connect (OIDC) Authorization Server แบบ stateless ที่พัฒนาด้วย Golang และ MongoDB

## คุณสมบัติ

- ✅ OAuth2 Authorization Code Flow with Session
- ✅ OAuth2 Client Credentials Flow
- ✅ OAuth2 Refresh Token Flow
- ✅ OpenID Connect (OIDC) Support
- ✅ JWT-based Stateless Tokens (RS256)
- ✅ RSA Key Pair Generation
- ✅ HTML Login & Register Pages
- ✅ Session Management
- ✅ User Registration & Authentication
- ✅ Client Registration
- ✅ OIDC Discovery Endpoint
- ✅ JWKS Endpoint
- ✅ UserInfo Endpoint
- ✅ Authorization Code with Session ID Tracking
- ✅ **Single Sign-On (SSO)** - Login once, access multiple applications
- ✅ **User Consent Management** - Remember user permissions across applications
- ✅ **Session Management API** - View and revoke active sessions
- ✅ **Authorization Management API** - View and revoke application permissions
- ✅ **OIDC Prompt Parameter Support** - Control authentication behavior (login, consent, none)

## โครงสร้างโปรเจกต์

```
.
├── config/              # Configuration management
├── database/            # Database connection
├── handlers/            # HTTP handlers
├── models/              # Data models
├── repository/          # Database repositories
├── utils/               # Utility functions (JWT, crypto)
├── main.go              # Application entry point
├── go.mod               # Go modules
└── .env.example         # Environment variables example
```

## การติดตั้ง

### ข้อกำหนด

- Go 1.21+
- MongoDB 4.4+

### ขั้นตอน

1. Clone repository และติดตั้ง dependencies:

```bash
go mod download
```

2. สร้างไฟล์ `.env` จาก `.env.example`:

```bash
cp .env.example .env
```

3. แก้ไขค่าใน `.env` ตามต้องการ:

```env
# Database
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=oauth2_db

# Server
SERVER_PORT=8080

# Token Configuration
ACCESS_TOKEN_EXPIRY=3600
REFRESH_TOKEN_EXPIRY=604800

# SSO Configuration (Optional)
SSO_SESSION_EXPIRY_DAYS=7          # SSO session lifetime (default: 7 days)
SSO_CONSENT_EXPIRY_DAYS=365        # Consent lifetime (default: 1 year)
SSO_COOKIE_SECURE=true             # Require HTTPS (set to false for local dev)
```

**หมายเหตุ:** RSA key pair จะถูกสร้างอัตโนมัติเมื่อรันครั้งแรก และจะถูกเก็บไว้ใน `keys/` directory

4. รัน MongoDB (ถ้ายังไม่ได้รัน):

```bash
# ใช้ Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# หรือรันโดยตรง
mongod
```

5. รันเซิร์ฟเวอร์:

```bash
go run main.go
```

## Logging

The server includes comprehensive structured logging with:
- Detail logs for individual operations
- Summary logs for transaction results
- Data masking for sensitive information
- File and console output support

See [logger/README.md](logger/README.md) for detailed documentation.

## Single Sign-On (SSO)

The OAuth2 Server now supports Single Sign-On functionality, allowing users to authenticate once and seamlessly access multiple client applications without re-entering credentials.

### Key Features

- **Persistent Sessions**: 7-day SSO sessions with secure HTTP-only cookies
- **Automatic Authorization**: Skip login and consent screens for returning users
- **Consent Management**: Remember user permissions for each application
- **Session Security**: IP address and user agent fingerprinting
- **Session Management**: View and revoke active sessions via API
- **Authorization Management**: View and revoke application permissions via API
- **OIDC Compliance**: Full support for `prompt` parameter (login, consent, none, select_account)

### Quick Start

1. **First Login**: User logs in to App A, grants consent → SSO session created
2. **Second App**: User accesses App B → Automatically authorized (no login/consent needed)
3. **Logout**: User logs out → SSO session cleared across all applications

**Documentation**:
- [SSO Usage Guide](docs/SSO_USAGE.md) - Comprehensive examples and flows
- [SSO API Reference](docs/SSO_API_REFERENCE.md) - Complete API documentation
- [SSO Quick Reference](docs/SSO_QUICK_REFERENCE.md) - Quick reference for developers

## API Endpoints

### Authentication

#### Show Register Page
```bash
GET /auth/register?session_id=SESSION_ID
```

#### Register User
```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "session_id": "optional_session_id"
}
```

#### Show Login Page
```bash
GET /auth/login?session_id=SESSION_ID
```

#### Login
```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "session_id": "optional_session_id"
}
```

#### Logout (SSO)
```bash
POST /auth/logout
# Optional: post_logout_redirect_uri parameter
POST /auth/logout?post_logout_redirect_uri=https://example.com/logged-out
```

### Client Management

#### Register OAuth Client
```bash
POST /clients/register
Content-Type: application/json

{
  "name": "My Application",
  "redirect_uris": ["http://localhost:3000/callback"]
}
```

### OAuth2/OIDC Flow

#### Authorization Endpoint
```bash
GET /oauth/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&scope=openid profile email&state=STATE

# จะ redirect ไปหน้า login พร้อม session_id
# หลัง login สำเร็จจะ redirect กลับพร้อม authorization code
```

#### Token Endpoint (Authorization Code)
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=AUTH_CODE&client_id=CLIENT_ID&client_secret=CLIENT_SECRET&redirect_uri=REDIRECT_URI
```

#### Token Endpoint (Refresh Token)
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&refresh_token=REFRESH_TOKEN&client_id=CLIENT_ID&client_secret=CLIENT_SECRET
```

#### Token Endpoint (Client Credentials)
```bash
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=CLIENT_ID&client_secret=CLIENT_SECRET&scope=SCOPE
```

#### UserInfo Endpoint
```bash
GET /oauth/userinfo
Authorization: Bearer ACCESS_TOKEN
```

#### OIDC Discovery
```bash
GET /.well-known/openid-configuration
```

#### JWKS Endpoint
```bash
GET /.well-known/jwks.json
```

### SSO Session Management

#### List Active Sessions
```bash
GET /account/sessions
Authorization: Bearer ACCESS_TOKEN
```

#### Revoke Specific Session
```bash
DELETE /account/sessions/{session_id}
Authorization: Bearer ACCESS_TOKEN
```

### SSO Authorization Management

#### List Authorized Applications
```bash
GET /account/authorizations
Authorization: Bearer ACCESS_TOKEN
```

#### Revoke Application Authorization
```bash
DELETE /account/authorizations/{client_id}
Authorization: Bearer ACCESS_TOKEN
```

## ตัวอย่างการใช้งาน

### วิธีที่ 1: ผ่าน Browser (แนะนำ)

1. **ลงทะเบียน OAuth Client:**
```bash
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test App",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'
```

2. **เปิด Browser และไปที่:**
```
http://localhost:8080/oauth/authorize?response_type=code&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random123
```

3. **Login หรือ Register** ผ่านหน้า web

4. **รับ Authorization Code** จาก callback URL

5. **แลก Code เป็น Token:**
```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=YOUR_CODE&client_id=CLIENT_ID&client_secret=CLIENT_SECRET&redirect_uri=http://localhost:3000/callback"
```

ดูรายละเอียดเพิ่มเติมใน [test_browser_flow.md](test_browser_flow.md)

### วิธีที่ 2: ผ่าน API โดยตรง

#### 1. ลงทะเบียนผู้ใช้ (ไม่ผ่าน OAuth flow)

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

#### 2. Login เพื่อรับ Access Token (ไม่ผ่าน OAuth flow)

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### 3. ดึงข้อมูล UserInfo

```bash
curl -X GET http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Security Features

- Password hashing ด้วย bcrypt
- JWT-based stateless authentication with RS256 (RSA asymmetric signing)
- RSA 2048-bit key pair
- Authorization code expiration (10 นาที)
- Access token expiration (configurable)
- Refresh token rotation
- Client secret validation
- Redirect URI validation
- Public key distribution via JWKS endpoint

## Testing

The project includes comprehensive unit and integration tests.

### Quick Test

```bash
# Run unit tests only (fast, no MongoDB required)
go test ./...
```

### Full Test Suite

```bash
# Run all tests including integration tests (requires MongoDB)
go test -tags=integration ./... -v
```

For detailed testing instructions, see [TESTING.md](TESTING.md).

## License

MIT
