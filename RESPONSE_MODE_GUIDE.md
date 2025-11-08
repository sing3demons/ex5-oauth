# Response Mode Guide

## Overview

ระบบรองรับการเลือก response format ได้ 2 แบบ:
1. **JSON Response** - สำหรับ API, SPA, Mobile apps
2. **Redirect** - สำหรับ traditional web apps (OAuth standard)

## Response Modes

### 1. `query` (Default)
ส่ง parameters ใน query string และ redirect (OAuth 2.0 standard)

```
https://client.example.com/callback?code=AUTH_CODE&state=xyz
```

### 2. `json`
Return JSON response โดยไม่ redirect

```json
{
  "redirect_uri": "https://client.example.com/callback",
  "code": "AUTH_CODE",
  "state": "xyz"
}
```

### 3. `fragment`
ส่ง parameters ใน URL fragment และ redirect (สำหรับ implicit flow)

```
https://client.example.com/callback#code=AUTH_CODE&state=xyz
```

### 4. `form_post`
ส่ง HTML form ที่ auto-submit ไปยัง redirect_uri

```html
<form method="post" action="https://client.example.com/callback">
  <input type="hidden" name="code" value="AUTH_CODE"/>
  <input type="hidden" name="state" value="xyz"/>
</form>
```

## How to Use

### Method 1: Explicit `response_mode` Parameter

เพิ่ม `response_mode` parameter ใน authorization request:

```bash
# JSON Response
GET /oauth/authorize?
  response_type=code&
  client_id=CLIENT_ID&
  redirect_uri=https://example.com/callback&
  scope=openid&
  response_mode=json

# Query Response (default)
GET /oauth/authorize?
  response_type=code&
  client_id=CLIENT_ID&
  redirect_uri=https://example.com/callback&
  scope=openid&
  response_mode=query

# Fragment Response
GET /oauth/authorize?
  response_type=code&
  client_id=CLIENT_ID&
  redirect_uri=https://example.com/callback&
  scope=openid&
  response_mode=fragment

# Form Post Response
GET /oauth/authorize?
  response_type=code&
  client_id=CLIENT_ID&
  redirect_uri=https://example.com/callback&
  scope=openid&
  response_mode=form_post
```

### Method 2: Accept Header

ระบบจะตรวจสอบ `Accept` header และ `Content-Type` header:

```bash
# JSON Response (auto-detected)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "session_id": "SESSION_ID"
  }'

# Response:
{
  "redirect_uri": "https://example.com/callback?code=AUTH_CODE&state=xyz",
  "code": "AUTH_CODE",
  "state": "xyz"
}
```

### Method 3: Default Behavior

ถ้าไม่ระบุ response_mode:
- **Browser request** → redirect (query mode)
- **API request** (with JSON headers) → JSON response

## Use Cases

### 1. Single Page Application (SPA)

```javascript
// Login with JSON response
async function login(email, password, sessionId) {
  const response = await fetch('http://localhost:8080/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    },
    body: JSON.stringify({
      email,
      password,
      session_id: sessionId
    })
  });

  const data = await response.json();
  
  // data = {
  //   redirect_uri: "https://app.example.com/callback?code=...",
  //   code: "AUTH_CODE",
  //   state: "xyz"
  // }

  // Extract code and exchange for tokens
  const code = data.code;
  const tokens = await exchangeCodeForTokens(code);
  
  return tokens;
}
```

### 2. Mobile App

```swift
// iOS Example
func login(email: String, password: String, sessionId: String) async throws -> AuthCode {
    let url = URL(string: "http://localhost:8080/auth/login")!
    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    request.setValue("application/json", forHTTPHeaderField: "Content-Type")
    request.setValue("application/json", forHTTPHeaderField: "Accept")
    
    let body = [
        "email": email,
        "password": password,
        "session_id": sessionId
    ]
    request.httpBody = try JSONEncoder().encode(body)
    
    let (data, _) = try await URLSession.shared.data(for: request)
    let response = try JSONDecoder().decode(AuthResponse.self, from: data)
    
    // response.code contains the authorization code
    return response.code
}

struct AuthResponse: Codable {
    let redirect_uri: String
    let code: String
    let state: String?
}
```

### 3. Traditional Web App

```html
<!-- Traditional OAuth flow with redirect -->
<form action="/oauth/authorize" method="GET">
  <input type="hidden" name="response_type" value="code">
  <input type="hidden" name="client_id" value="CLIENT_ID">
  <input type="hidden" name="redirect_uri" value="https://example.com/callback">
  <input type="hidden" name="scope" value="openid profile email">
  <input type="hidden" name="state" value="random_state">
  <!-- No response_mode = default redirect behavior -->
  <button type="submit">Login with OAuth</button>
</form>
```

### 4. React Application

```typescript
// useAuth.ts
import { useState } from 'react';

interface AuthResponse {
  redirect_uri: string;
  code: string;
  state?: string;
}

export function useAuth() {
  const [loading, setLoading] = useState(false);

  const login = async (email: string, password: string, sessionId: string) => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify({
          email,
          password,
          session_id: sessionId,
        }),
      });

      if (!response.ok) {
        throw new Error('Login failed');
      }

      const data: AuthResponse = await response.json();
      
      // Got authorization code without redirect!
      return data.code;
    } finally {
      setLoading(false);
    }
  };

  return { login, loading };
}

// LoginComponent.tsx
function LoginComponent() {
  const { login } = useAuth();
  const [sessionId] = useState(() => new URLSearchParams(window.location.search).get('session_id'));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    
    const code = await login(
      formData.get('email') as string,
      formData.get('password') as string,
      sessionId!
    );

    // Exchange code for tokens
    const tokens = await exchangeCodeForTokens(code);
    
    // Store tokens and redirect to app
    localStorage.setItem('access_token', tokens.access_token);
    window.location.href = '/dashboard';
  };

  return (
    <form onSubmit={handleSubmit}>
      <input type="email" name="email" required />
      <input type="password" name="password" required />
      <button type="submit">Login</button>
    </form>
  );
}
```

## Complete Flow Examples

### Flow 1: SPA with JSON Response

```
1. User clicks "Login" in SPA
   ↓
2. SPA redirects to: /oauth/authorize?response_mode=json&...
   ↓
3. User sees login page
   ↓
4. User submits credentials (POST /auth/login with JSON headers)
   ↓
5. Server returns JSON:
   {
     "redirect_uri": "https://app.com/callback?code=...",
     "code": "AUTH_CODE",
     "state": "xyz"
   }
   ↓
6. SPA extracts code from JSON (no page redirect!)
   ↓
7. SPA exchanges code for tokens (POST /oauth/token)
   ↓
8. SPA stores tokens and shows dashboard
```

### Flow 2: Traditional Web with Redirect

```
1. User clicks "Login" button
   ↓
2. Browser redirects to: /oauth/authorize?...
   ↓
3. User sees login page
   ↓
4. User submits credentials (POST /auth/login from HTML form)
   ↓
5. Server redirects to: https://app.com/callback?code=AUTH_CODE&state=xyz
   ↓
6. App backend receives code
   ↓
7. App backend exchanges code for tokens
   ↓
8. App backend creates session and shows dashboard
```

## Testing

### Test JSON Response

```bash
# Start authorization flow with JSON mode
curl -X GET "http://localhost:8080/oauth/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid&state=test&response_mode=json"

# This will redirect to login page with session_id

# Login with JSON response
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "session_id": "SESSION_ID_FROM_STEP_1"
  }'

# Response:
{
  "redirect_uri": "http://localhost:3000/callback?code=AUTH_CODE&state=test",
  "code": "AUTH_CODE",
  "state": "test"
}
```

### Test Redirect Response

```bash
# Login without JSON headers (will redirect)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "session_id": "SESSION_ID"
  }' \
  -L  # Follow redirects

# Will redirect to callback URL
```

## Benefits

### ✅ For SPA/Mobile Apps:
- No page redirects
- Better UX
- Easier state management
- Direct code extraction

### ✅ For Traditional Web Apps:
- Standard OAuth flow
- Browser handles redirects
- Simpler implementation

### ✅ For Developers:
- Flexible integration
- One server, multiple client types
- Standards-compliant

## Summary

| Client Type | Recommended Mode | Why |
|-------------|-----------------|-----|
| SPA | `json` | No page reload, better UX |
| Mobile App | `json` | Native handling, no webview redirect |
| Traditional Web | `query` (default) | Standard OAuth flow |
| Hybrid App | `json` or `fragment` | Depends on architecture |

**ระบบรองรับทั้ง JSON และ Redirect แล้ว!** 🎉
