# OAuth2/OIDC Authorization Server (Stateless)

OAuth2 และ OpenID Connect (OIDC) Authorization Server แบบ stateless ที่พัฒนาด้วย Golang และ MongoDB

## คุณสมบัติ

- ✅ OAuth2 Authorization Code Flow
- ✅ OAuth2 Client Credentials Flow
- ✅ OAuth2 Refresh Token Flow
- ✅ OpenID Connect (OIDC) Support
- ✅ JWT-based Stateless Tokens (RS256)
- ✅ RSA Key Pair Generation
- ✅ User Registration & Authentication
- ✅ Client Registration
- ✅ OIDC Discovery Endpoint
- ✅ JWKS Endpoint
- ✅ UserInfo Endpoint

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
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=oauth2_db
SERVER_PORT=8080
ACCESS_TOKEN_EXPIRY=3600
REFRESH_TOKEN_EXPIRY=604800
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

## API Endpoints

### Authentication

#### Register User
```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Login
```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
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
Authorization: Bearer ACCESS_TOKEN
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

## ตัวอย่างการใช้งาน

### 1. ลงทะเบียนผู้ใช้

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### 2. ลงทะเบียน OAuth Client

```bash
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test App",
    "redirect_uris": ["http://localhost:3000/callback"]
  }'
```

### 3. Login เพื่อรับ Access Token

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 4. ขอ Authorization Code

```bash
curl -X GET "http://localhost:8080/oauth/authorize?response_type=code&client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid profile email&state=random_state" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. แลก Authorization Code เป็น Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&client_id=CLIENT_ID&client_secret=CLIENT_SECRET&redirect_uri=http://localhost:3000/callback"
```

### 6. ดึงข้อมูล UserInfo

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

## License

MIT
