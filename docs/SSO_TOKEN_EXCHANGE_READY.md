# ✅ SSO with Token Exchange - Ready to Use!

## Good News! 🎉

ระบบของคุณ **มี Token Exchange handler อยู่แล้ว** และพร้อมใช้งาน SSO แบบไม่ใช้ Cookie!

## What You Already Have

### ✅ Token Exchange Handler
- File: `handlers/token_exchange_handler.go`
- รองรับ RFC 8693 Token Exchange
- รองรับทั้ง JWT และ JWE tokens
- มี scope validation
- มี client authentication

### ✅ Grant Type Support
```go
const TokenExchangeGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
```

### ✅ Token Types Support
- Access Token: `urn:ietf:params:oauth:token-type:access_token`
- Refresh Token: `urn:ietf:params:oauth:token-type:refresh_token`
- ID Token: `urn:ietf:params:oauth:token-type:id_token`

## How to Use SSO Right Now

### 1. User Logs into App A

```bash
# Normal OAuth flow
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE" \
  -d "client_id=app-a" \
  -d "client_secret=secret-a" \
  -d "redirect_uri=http://localhost:3000/callback"

# Response:
{
  "access_token": "eyJhbGc...",  # Store this!
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "...",
  "scope": "openid profile email"
}
```

### 2. User Accesses App B - Token Exchange (SSO!)

```bash
# Exchange App A's token for App B's token
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=eyJhbGc..." \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=app-b" \
  -d "client_secret=secret-b" \
  -d "scope=openid profile"

# Response:
{
  "access_token": "eyJhbGc...",  # New token for App B!
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "...",
  "scope": "openid profile"
}
```

**🎉 No login required! User is automatically authenticated!**

## Complete Example Flow

```javascript
// sso-client.js
class SSOClient {
  constructor() {
    this.tokens = new Map();
  }

  // Step 1: Login to App A
  async loginToAppA() {
    // Normal OAuth flow
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code: 'AUTH_CODE_FROM_CALLBACK',
        client_id: 'app-a',
        client_secret: 'secret-a',
        redirect_uri: 'http://localhost:3000/callback'
      })
    });

    const tokens = await response.json();
    this.tokens.set('app-a', tokens);
    console.log('✅ Logged into App A');
    return tokens;
  }

  // Step 2: Access App B using token exchange (SSO!)
  async accessAppB() {
    // Get App A's token
    const appATokens = this.tokens.get('app-a');
    if (!appATokens) {
      throw new Error('Please login to App A first');
    }

    // Exchange for App B token
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: appATokens.access_token,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: 'app-b',
        client_secret: 'secret-b',
        scope: 'openid profile'
      })
    });

    const tokens = await response.json();
    this.tokens.set('app-b', tokens);
    console.log('✅ Accessed App B without login!');
    return tokens;
  }

  // Step 3: Access App C (same pattern)
  async accessAppC() {
    const appATokens = this.tokens.get('app-a');
    
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: appATokens.access_token,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: 'app-c',
        client_secret: 'secret-c',
        scope: 'openid'
      })
    });

    const tokens = await response.json();
    this.tokens.set('app-c', tokens);
    console.log('✅ Accessed App C without login!');
    return tokens;
  }
}

// Usage
const sso = new SSOClient();

// User logs into App A
await sso.loginToAppA();

// User accesses App B - automatic SSO!
await sso.accessAppB();

// User accesses App C - automatic SSO!
await sso.accessAppC();

console.log('🎉 SSO working! User logged into 3 apps with 1 login!');
```

## React Example

```typescript
// useSSO.ts
import { useState, useCallback } from 'react';

interface TokenSet {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope: string;
}

export function useSSO() {
  const [tokens, setTokens] = useState<Map<string, TokenSet>>(new Map());

  const exchangeToken = useCallback(async (
    subjectToken: string,
    targetClientId: string,
    targetClientSecret: string,
    scope?: string
  ): Promise<TokenSet> => {
    const response = await fetch('http://localhost:8080/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: subjectToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: targetClientId,
        client_secret: targetClientSecret,
        ...(scope && { scope })
      })
    });

    if (!response.ok) {
      throw new Error('Token exchange failed');
    }

    const newTokens = await response.json();
    
    // Update state
    setTokens(prev => new Map(prev).set(targetClientId, newTokens));
    
    return newTokens;
  }, []);

  const getTokenForApp = useCallback(async (
    appId: string,
    appSecret: string,
    sourceAppId?: string
  ): Promise<string> => {
    // Check if we already have a valid token
    const existing = tokens.get(appId);
    if (existing && !isExpired(existing.access_token)) {
      return existing.access_token;
    }

    // Find a source token
    const sourceToken = sourceAppId 
      ? tokens.get(sourceAppId)?.access_token
      : Array.from(tokens.values())[0]?.access_token;

    if (!sourceToken) {
      throw new Error('No source token available');
    }

    // Exchange
    const newTokens = await exchangeToken(sourceToken, appId, appSecret);
    return newTokens.access_token;
  }, [tokens, exchangeToken]);

  return { tokens, exchangeToken, getTokenForApp };
}

function isExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
}

// Component usage
function AppB() {
  const { getTokenForApp } = useSSO();
  const [data, setData] = useState(null);

  useEffect(() => {
    async function fetchData() {
      try {
        // Get token for App B (will exchange if needed)
        const token = await getTokenForApp('app-b', 'secret-b', 'app-a');
        
        // Use token to fetch data
        const response = await fetch('http://localhost:8081/api/data', {
          headers: { Authorization: `Bearer ${token}` }
        });
        
        setData(await response.json());
      } catch (error) {
        console.error('Failed to fetch data:', error);
      }
    }

    fetchData();
  }, [getTokenForApp]);

  return <div>{data ? JSON.stringify(data) : 'Loading...'}</div>;
}
```

## Mobile App Example (React Native)

```typescript
// SSOManager.ts
import AsyncStorage from '@react-native-async-storage/async-storage';

class SSOManager {
  private static instance: SSOManager;
  private tokens: Map<string, TokenSet> = new Map();

  static getInstance(): SSOManager {
    if (!SSOManager.instance) {
      SSOManager.instance = new SSOManager();
    }
    return SSOManager.instance;
  }

  async initialize() {
    // Load tokens from secure storage
    const stored = await AsyncStorage.getItem('sso_tokens');
    if (stored) {
      this.tokens = new Map(JSON.parse(stored));
    }
  }

  async exchangeToken(
    subjectToken: string,
    targetClientId: string,
    targetClientSecret: string
  ): Promise<TokenSet> {
    const response = await fetch('https://auth.example.com/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: subjectToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        client_id: targetClientId,
        client_secret: targetClientSecret,
      }).toString()
    });

    const tokens = await response.json();
    this.tokens.set(targetClientId, tokens);
    
    // Persist to secure storage
    await AsyncStorage.setItem('sso_tokens', JSON.stringify(Array.from(this.tokens)));
    
    return tokens;
  }

  async getTokenForApp(appId: string, appSecret: string): Promise<string> {
    // Check cache
    const cached = this.tokens.get(appId);
    if (cached && !this.isExpired(cached.access_token)) {
      return cached.access_token;
    }

    // Find any valid token to exchange
    for (const [_, tokens] of this.tokens) {
      if (!this.isExpired(tokens.access_token)) {
        const newTokens = await this.exchangeToken(
          tokens.access_token,
          appId,
          appSecret
        );
        return newTokens.access_token;
      }
    }

    throw new Error('No valid tokens available');
  }

  private isExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }
}

export default SSOManager.getInstance();
```

## Testing the SSO Flow

```bash
# 1. Register two clients
curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{"name":"App A","redirect_uris":["http://localhost:3000/callback"]}'

curl -X POST http://localhost:8080/clients/register \
  -H "Content-Type: application/json" \
  -d '{"name":"App B","redirect_uris":["http://localhost:3001/callback"]}'

# 2. Get authorization code for App A (via browser)
# Open: http://localhost:8080/oauth/authorize?client_id=CLIENT_A&...

# 3. Exchange code for tokens (App A)
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=CODE_FROM_STEP_2" \
  -d "client_id=CLIENT_A" \
  -d "client_secret=SECRET_A" \
  -d "redirect_uri=http://localhost:3000/callback"

# Save the access_token from response

# 4. Exchange App A token for App B token (SSO!)
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=ACCESS_TOKEN_FROM_STEP_3" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "requested_token_type=urn:ietf:params:oauth:token-type:access_token" \
  -d "client_id=CLIENT_B" \
  -d "client_secret=SECRET_B"

# 🎉 You now have a token for App B without logging in again!
```

## What Makes This SSO?

### Traditional (No SSO):
```
User → App A → Login → Token A
User → App B → Login AGAIN → Token B  😫
User → App C → Login AGAIN → Token C  😫
```

### With Token Exchange SSO:
```
User → App A → Login → Token A
User → App B → Exchange Token A → Token B  ✨ (No login!)
User → App C → Exchange Token A → Token C  ✨ (No login!)
```

## Advantages

✅ **No Cookies** - Works everywhere  
✅ **Mobile Friendly** - Native support  
✅ **SPA Ready** - No CORS issues  
✅ **Microservices** - Service-to-service auth  
✅ **Stateless** - Scales horizontally  
✅ **Secure** - JWT-based with validation  
✅ **Flexible** - Scope control per app  
✅ **Already Implemented** - Ready to use!  

## Next Steps

1. ✅ **You're ready!** - Handler already exists
2. 📱 Implement client-side token management
3. 🔒 Add token caching and refresh logic
4. 📊 Add monitoring and logging
5. 🧪 Test with multiple apps

## Summary

คุณไม่ต้องทำอะไรเพิ่มเติมฝั่ง server! ระบบรองรับ SSO แบบ Token Exchange อยู่แล้ว

แค่ implement client-side logic เพื่อ:
1. เก็บ token จาก App A
2. ใช้ token exchange เมื่อเข้า App B
3. Cache tokens ไว้ใช้ซ้ำ

**SSO แบบไม่ใช้ Cookie พร้อมใช้งานแล้ว!** 🚀
