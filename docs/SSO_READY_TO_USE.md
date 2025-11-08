# ✅ SSO Token Exchange - พร้อมใช้งาน 100%!

## 🎉 ยืนยัน: ใช้ได้เลยโดยไม่ต้อง code เพิ่ม!

ระบบ OAuth2 ของคุณ **รองรับ SSO แบบ Token Exchange แล้ว** และพร้อมใช้งานทันที!

## ✅ สิ่งที่เพิ่งทำเสร็จ

เพิ่ม Token Exchange support ใน OAuth Token endpoint:

```go
// handlers/oauth_handler.go
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
    grantType := r.FormValue("grant_type")
    
    switch grantType {
    case "authorization_code":
        h.handleAuthorizationCodeGrant(w, r)
    case "refresh_token":
        h.handleRefreshTokenGrant(w, r)
    case "client_credentials":
        h.handleClientCredentialsGrant(w, r)
    case "urn:ietf:params:oauth:grant-type:token-exchange":  // ✨ เพิ่มใหม่!
        h.handleTokenExchange(w, r)
    default:
        respondError(w, http.StatusBadRequest, "unsupported_grant_type", ...)
    }
}
```

## 🚀 วิธีใช้งาน SSO ทันที

### Endpoint เดียว: `/oauth/token`

ใช้ endpoint เดียวกันกับ grant types อื่นๆ ตามมาตรฐาน OAuth 2.0

### ตัวอย่างการใช้งาน

#### 1. User Login App A (ครั้งแรก)

```bash
# ขั้นตอนปกติ - OAuth Authorization Code Flow
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "client_id=app-a-client-id" \
  -d "client_secret=app-a-secret" \
  -d "redirect_uri=http://localhost:3000/callback"

# Response:
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "...",
  "scope": "openid profile email"
}
```

**💾 เก็บ `access_token` ไว้!**

#### 2. User เข้า App B (SSO - ไม่ต้อง Login!)

```bash
# ใช้ Token Exchange - แลก Token A เป็น Token B
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=app-b-client-id" \
  -d "client_secret=app-b-secret" \
  -d "scope=openid profile"

# Response:
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",  # Token ใหม่สำหรับ App B!
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "...",
  "scope": "openid profile"
}
```

**🎉 ได้ Token สำหรับ App B โดยไม่ต้อง Login!**

#### 3. User เข้า App C (SSO ต่อเนื่อง!)

```bash
# แลก Token A (หรือ Token B) เป็น Token C
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=TOKEN_FROM_APP_A_OR_B" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=app-c-client-id" \
  -d "client_secret=app-c-secret"

# Response: Token สำหรับ App C!
```

## 📱 Client-Side Implementation

### JavaScript/TypeScript

```typescript
class SSOClient {
  private tokens = new Map<string, string>();

  // Login to first app
  async login(appId: string, authCode: string): Promise<string> {
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code: authCode,
        client_id: appId,
        client_secret: 'secret',
        redirect_uri: 'http://localhost:3000/callback'
      })
    });

    const data = await response.json();
    this.tokens.set(appId, data.access_token);
    return data.access_token;
  }

  // Get token for another app (SSO!)
  async getTokenForApp(targetAppId: string, targetSecret: string): Promise<string> {
    // Get any existing token
    const sourceToken = Array.from(this.tokens.values())[0];
    if (!sourceToken) {
      throw new Error('No token available. Please login first.');
    }

    // Exchange token
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: sourceToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: targetAppId,
        client_secret: targetSecret
      })
    });

    const data = await response.json();
    this.tokens.set(targetAppId, data.access_token);
    return data.access_token;
  }
}

// Usage
const sso = new SSOClient();

// Step 1: Login to App A
await sso.login('app-a', 'auth_code_from_callback');

// Step 2: Access App B (automatic SSO!)
const tokenB = await sso.getTokenForApp('app-b', 'secret-b');

// Step 3: Access App C (automatic SSO!)
const tokenC = await sso.getTokenForApp('app-c', 'secret-c');

console.log('✅ Logged into 3 apps with 1 login!');
```

### React Hook

```typescript
import { useState, useCallback } from 'react';

export function useSSO() {
  const [tokens, setTokens] = useState<Map<string, string>>(new Map());

  const exchangeToken = useCallback(async (
    subjectToken: string,
    targetClientId: string,
    targetClientSecret: string
  ) => {
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: subjectToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: targetClientId,
        client_secret: targetClientSecret
      })
    });

    const data = await response.json();
    setTokens(prev => new Map(prev).set(targetClientId, data.access_token));
    return data.access_token;
  }, []);

  return { tokens, exchangeToken };
}

// Component
function AppB() {
  const { exchangeToken } = useSSO();
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    async function getToken() {
      const sourceToken = localStorage.getItem('app_a_token');
      if (sourceToken) {
        const newToken = await exchangeToken(
          sourceToken,
          'app-b-client-id',
          'app-b-secret'
        );
        setToken(newToken);
      }
    }
    getToken();
  }, [exchangeToken]);

  return <div>Token: {token}</div>;
}
```

## 🧪 ทดสอบ SSO

```bash
# 1. Start server
go run main.go

# 2. Register two clients
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{"name":"App A","redirect_uris":["http://localhost:3000/callback"]}'

curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{"name":"App B","redirect_uris":["http://localhost:3001/callback"]}'

# 3. Get auth code for App A (via browser)
# Open: http://localhost:8080/oauth/authorize?response_type=code&client_id=CLIENT_A&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random

# 4. Exchange code for token (App A)
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=CODE_FROM_STEP_3" \
  -d "client_id=CLIENT_A" \
  -d "client_secret=SECRET_A" \
  -d "redirect_uri=http://localhost:3000/callback"

# Save access_token from response

# 5. Exchange token for App B (SSO!)
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=ACCESS_TOKEN_FROM_STEP_4" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=CLIENT_B" \
  -d "client_secret=SECRET_B"

# ✅ You got App B token without login!
```

## ✨ Features

### ✅ ฝั่ง Server (พร้อมใช้งาน 100%)
- Token Exchange handler
- Scope validation
- Token validation (JWT & JWE)
- Client authentication
- Scope downgrade support
- Standard OAuth 2.0 endpoint

### 📱 ฝั่ง Client (ต้อง implement)
- Token storage
- Token exchange logic
- Token caching
- Automatic refresh

## 🎯 สรุป

### คำตอบคำถาม: "ใช้ได้เลย โดยไม่ต้อง code เพิ่มใช่ไหม"

**✅ ใช่! ฝั่ง Server พร้อมใช้งานแล้ว 100%**

ไม่ต้อง code เพิ่มฝั่ง server เลย! แค่:

1. **Start server**: `go run main.go`
2. **Register clients**: ใช้ `/clients/register`
3. **Use Token Exchange**: ส่ง request ไปที่ `/oauth/token` ด้วย `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`

**📱 ฝั่ง Client ต้อง implement:**
- Token management (เก็บ, cache, exchange)
- ใช้ code examples ที่ให้ไว้

## 🚀 เริ่มใช้งานได้เลย!

```bash
# Start server
go run main.go

# Server พร้อมรับ Token Exchange requests ที่:
# POST http://localhost:8080/oauth/token
# grant_type=urn:ietf:params:oauth:grant-type:token-exchange
```

**SSO แบบ Token Exchange พร้อมใช้งาน 100%!** 🎉
