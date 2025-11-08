# SSO with Token Exchange (Cookie-less)

## Overview

SSO แบบ Token Exchange ใช้ **OAuth 2.0 Token Exchange (RFC 8693)** แทน cookies ทำให้:
- ✅ ทำงานได้กับ mobile apps และ SPA
- ✅ ไม่มีปัญหา cross-domain
- ✅ ไม่ต้องพึ่ง browser cookies
- ✅ รองรับ microservices architecture
- ✅ Stateless และ scalable

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Token Exchange Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  App A Login → Get Token A                                  │
│       ↓                                                      │
│  Store Token A in App A                                     │
│       ↓                                                      │
│  App B needs auth → Exchange Token A for Token B            │
│       ↓                                                      │
│  OAuth Server validates Token A                             │
│       ↓                                                      │
│  Issue Token B for App B                                    │
│       ↓                                                      │
│  App B can now access resources                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Key Concepts

### 1. Subject Token (Token A)
- Token ที่ user มีอยู่แล้วจาก App A
- ใช้เป็นหลักฐานว่า user login แล้ว
- มี scope และ permissions ของ App A

### 2. Token Exchange
- แลก Subject Token เป็น Token ใหม่สำหรับ App B
- Token ใหม่มี scope และ audience ที่เหมาะกับ App B
- ไม่ต้อง login ซ้ำ

### 3. Actor Token (Optional)
- Token ของ service ที่ทำการ exchange
- ใช้สำหรับ delegation scenarios

## Token Exchange Request (RFC 8693)

```http
POST /oauth/token HTTP/1.1
Host: auth.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:token-exchange
&subject_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
&subject_token_type=urn:ietf:params:oauth:token-type:access_token
&requested_token_type=urn:ietf:params:oauth:token-type:access_token
&audience=app-b-client-id
&scope=openid profile email
&client_id=app-b-client-id
&client_secret=app-b-secret
```

### Parameters:

- **grant_type**: `urn:ietf:params:oauth:grant-type:token-exchange`
- **subject_token**: Token ที่มีอยู่แล้ว (from App A)
- **subject_token_type**: ประเภทของ subject token
  - `urn:ietf:params:oauth:token-type:access_token`
  - `urn:ietf:params:oauth:token-type:refresh_token`
  - `urn:ietf:params:oauth:token-type:id_token`
- **requested_token_type**: ประเภทของ token ที่ต้องการ
- **audience**: Client ID ของ app ที่ต้องการใช้ token
- **scope**: Scopes ที่ต้องการ (ต้องเป็น subset ของ subject token)
- **client_id**: Client ID ของ app ที่ request
- **client_secret**: Client secret

## Token Exchange Response

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "openid profile email",
  "refresh_token": "optional_refresh_token"
}
```

## Complete SSO Flow

### Scenario: User wants to access App B after logging into App A

```
┌──────┐    ┌───────┐    ┌───────┐    ┌──────────────┐
│ User │    │ App A │    │ App B │    │ OAuth Server │
└──┬───┘    └───┬───┘    └───┬───┘    └──────┬───────┘
   │            │            │               │
   │ 1. Login   │            │               │
   ├───────────>│            │               │
   │            │            │               │
   │            │ 2. OAuth   │               │
   │            │    Flow    │               │
   │            ├───────────────────────────>│
   │            │<───────────────────────────┤
   │            │ Token A                    │
   │            │                            │
   │ 3. Store   │            │               │
   │    Token A │            │               │
   │<───────────┤            │               │
   │            │            │               │
   │ 4. Access  │            │               │
   │    App B   │            │               │
   ├────────────────────────>│               │
   │            │            │               │
   │            │            │ 5. Need auth  │
   │            │            │    Check local│
   │            │            │    storage    │
   │            │            │    (no token) │
   │            │            │               │
   │            │ 6. Request │               │
   │            │    Token A │               │
   │            │<───────────┤               │
   │            │            │               │
   │ 7. Provide │            │               │
   │    Token A │            │               │
   │<───────────┤            │               │
   │────────────────────────>│               │
   │            │            │               │
   │            │            │ 8. Exchange   │
   │            │            │    Token A    │
   │            │            │    for Token B│
   │            │            ├──────────────>│
   │            │            │               │
   │            │            │ 9. Validate   │
   │            │            │    Token A    │
   │            │            │    Check user │
   │            │            │    Check scope│
   │            │            │               │
   │            │            │ 10. Issue     │
   │            │            │     Token B   │
   │            │            │<──────────────┤
   │            │            │               │
   │ 11. Store  │            │               │
   │     Token B│            │               │
   │<────────────────────────┤               │
   │            │            │               │
   │ 12. Access │            │               │
   │     App B  │            │               │
   │     (with  │            │               │
   │     Token B)│           │               │
   │<────────────────────────┤               │
```

## Implementation

### 1. Token Exchange Handler

```go
// handlers/token_exchange_handler.go
package handlers

import (
	"context"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"
	"time"
)

const (
	GrantTypeTokenExchange = "urn:ietf:params:oauth:grant-type:token-exchange"
	TokenTypeAccessToken   = "urn:ietf:params:oauth:token-type:access_token"
	TokenTypeRefreshToken  = "urn:ietf:params:oauth:token-type:refresh_token"
	TokenTypeIDToken       = "urn:ietf:params:oauth:token-type:id_token"
)

type TokenExchangeHandler struct {
	userRepo   *repository.UserRepository
	clientRepo *repository.ClientRepository
	config     *config.Config
}

func NewTokenExchangeHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	cfg *config.Config,
) *TokenExchangeHandler {
	return &TokenExchangeHandler{
		userRepo:   userRepo,
		clientRepo: clientRepo,
		config:     cfg,
	}
}

func (h *TokenExchangeHandler) HandleTokenExchange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	// Extract parameters
	subjectToken := r.FormValue("subject_token")
	subjectTokenType := r.FormValue("subject_token_type")
	requestedTokenType := r.FormValue("requested_token_type")
	audience := r.FormValue("audience")
	requestedScope := r.FormValue("scope")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Validate required parameters
	if subjectToken == "" || subjectTokenType == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing subject_token or subject_token_type")
		return
	}

	if clientID == "" || clientSecret == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing client credentials")
		return
	}

	ctx := context.Background()

	// Validate client
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil || client.ClientSecret != clientSecret {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
	}

	// Validate subject token based on type
	var userID string
	var originalScope string

	switch subjectTokenType {
	case TokenTypeAccessToken:
		// Validate access token
		claims, err := utils.ValidateToken(subjectToken, h.config.PublicKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_token", "Invalid subject token")
			return
		}
		userID = claims.UserID
		originalScope = claims.Scope

	case TokenTypeRefreshToken:
		// Validate refresh token
		claims, err := utils.ValidateRefreshToken(subjectToken, h.config.PublicKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_token", "Invalid subject token")
			return
		}
		userID = claims.UserID
		originalScope = claims.Scope

	case TokenTypeIDToken:
		// For ID tokens, we need custom validation
		respondError(w, http.StatusBadRequest, "unsupported_token_type", "ID token exchange not yet supported")
		return

	default:
		respondError(w, http.StatusBadRequest, "unsupported_token_type", "Unsupported subject_token_type")
		return
	}

	// Get user
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_token", "User not found")
		return
	}

	// Determine scope for new token
	newScope := requestedScope
	if newScope == "" {
		// Use original scope if not specified
		newScope = originalScope
	} else {
		// Validate scope downgrade (requested scope must be subset of original)
		if err := utils.ValidateScopeDowngrade(newScope, originalScope); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
	}

	// Validate scopes against target client's allowed scopes
	if len(client.AllowedScopes) > 0 {
		if err := utils.GlobalScopeValidator.ValidateScopeAgainstAllowed(newScope, client.AllowedScopes); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
	}

	// Determine requested token type (default to access token)
	if requestedTokenType == "" {
		requestedTokenType = TokenTypeAccessToken
	}

	// Generate new token based on requested type
	var response map[string]interface{}

	switch requestedTokenType {
	case TokenTypeAccessToken:
		// Generate new access token with audience
		accessToken, err := utils.GenerateAccessTokenWithAudience(
			user.ID,
			user.Email,
			user.Name,
			newScope,
			audience, // Set audience to target client
			h.config.PrivateKey,
			h.config.AccessTokenExpiry,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate token")
			return
		}

		// Optionally generate refresh token
		refreshToken, _ := utils.GenerateRefreshToken(user.ID, newScope, h.config.PrivateKey, h.config.RefreshTokenExpiry)

		response = map[string]interface{}{
			"access_token":       accessToken,
			"issued_token_type":  TokenTypeAccessToken,
			"token_type":         "Bearer",
			"expires_in":         h.config.AccessTokenExpiry,
			"scope":              newScope,
			"refresh_token":      refreshToken,
		}

	case TokenTypeRefreshToken:
		// Generate refresh token
		refreshToken, err := utils.GenerateRefreshToken(user.ID, newScope, h.config.PrivateKey, h.config.RefreshTokenExpiry)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate token")
			return
		}

		response = map[string]interface{}{
			"access_token":      refreshToken,
			"issued_token_type": TokenTypeRefreshToken,
			"token_type":        "Bearer",
			"expires_in":        h.config.RefreshTokenExpiry,
			"scope":             newScope,
		}

	default:
		respondError(w, http.StatusBadRequest, "unsupported_token_type", "Unsupported requested_token_type")
		return
	}

	respondJSON(w, http.StatusOK, response)
}
```

### 2. Enhanced JWT Utils with Audience Support

```go
// utils/jwt.go - Add audience support
func GenerateAccessTokenWithAudience(userID, email, name, scope, audience string, privateKey *rsa.PrivateKey, expiry int64) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		Scope:  scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Audience:  jwt.ClaimStrings{audience}, // Set audience
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "oauth2-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// Validate token with audience check
func ValidateTokenWithAudience(tokenString string, publicKey *rsa.PublicKey, expectedAudience string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		// Validate audience
		if expectedAudience != "" {
			found := false
			for _, aud := range claims.Audience {
				if aud == expectedAudience {
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New("invalid audience")
			}
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
```

### 3. Update OAuth Handler to Support Token Exchange

```go
// handlers/oauth_handler.go - Update Token method
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	case "client_credentials":
		h.handleClientCredentialsGrant(w, r)
	case GrantTypeTokenExchange:
		// Delegate to token exchange handler
		h.tokenExchangeHandler.HandleTokenExchange(w, r)
	default:
		respondError(w, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
	}
}
```

### 4. Update Discovery Endpoint

```go
// handlers/discovery_handler.go - Add token exchange support
func (h *DiscoveryHandler) WellKnown(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		// ... existing fields ...
		"grant_types_supported": []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
			"urn:ietf:params:oauth:grant-type:token-exchange", // Add this
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_post",
			"client_secret_basic",
		},
	}

	respondJSON(w, http.StatusOK, discovery)
}
```

## Client-Side Implementation

### JavaScript/TypeScript Example

```typescript
// sso-manager.ts
class SSOManager {
  private tokens: Map<string, TokenSet> = new Map();

  // Store token after login
  async login(appId: string): Promise<TokenSet> {
    // Normal OAuth flow
    const tokens = await this.performOAuthFlow(appId);
    this.tokens.set(appId, tokens);
    return tokens;
  }

  // Get token for another app using token exchange
  async getTokenForApp(targetAppId: string, sourceAppId?: string): Promise<TokenSet> {
    // Check if we already have a token for target app
    if (this.tokens.has(targetAppId)) {
      const tokens = this.tokens.get(targetAppId)!;
      if (!this.isTokenExpired(tokens.access_token)) {
        return tokens;
      }
    }

    // Find a valid token from any app
    const sourceToken = this.findValidToken(sourceAppId);
    if (!sourceToken) {
      throw new Error('No valid token available for exchange');
    }

    // Exchange token
    const newTokens = await this.exchangeToken(sourceToken, targetAppId);
    this.tokens.set(targetAppId, newTokens);
    return newTokens;
  }

  private async exchangeToken(subjectToken: string, targetClientId: string): Promise<TokenSet> {
    const response = await fetch('https://auth.example.com/oauth/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        subject_token: subjectToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        audience: targetClientId,
        client_id: targetClientId,
        client_secret: 'target-client-secret',
      }),
    });

    if (!response.ok) {
      throw new Error('Token exchange failed');
    }

    return await response.json();
  }

  private findValidToken(preferredAppId?: string): string | null {
    // Try preferred app first
    if (preferredAppId && this.tokens.has(preferredAppId)) {
      const tokens = this.tokens.get(preferredAppId)!;
      if (!this.isTokenExpired(tokens.access_token)) {
        return tokens.access_token;
      }
    }

    // Find any valid token
    for (const [_, tokens] of this.tokens) {
      if (!this.isTokenExpired(tokens.access_token)) {
        return tokens.access_token;
      }
    }

    return null;
  }

  private isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }
}

interface TokenSet {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  scope: string;
}

// Usage
const ssoManager = new SSOManager();

// User logs into App A
await ssoManager.login('app-a-client-id');

// Later, user accesses App B - automatic token exchange!
const appBTokens = await ssoManager.getTokenForApp('app-b-client-id', 'app-a-client-id');
```

### React Hook Example

```typescript
// useSSOToken.ts
import { useState, useEffect } from 'react';

export function useSSOToken(clientId: string) {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    async function getToken() {
      try {
        setLoading(true);
        
        // Check local storage first
        const stored = localStorage.getItem(`token_${clientId}`);
        if (stored && !isExpired(stored)) {
          setToken(stored);
          setLoading(false);
          return;
        }

        // Try token exchange from another app
        const sourceToken = findAnyValidToken();
        if (sourceToken) {
          const newToken = await exchangeToken(sourceToken, clientId);
          localStorage.setItem(`token_${clientId}`, newToken);
          setToken(newToken);
        } else {
          // Need to login
          setError(new Error('No valid token available'));
        }
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    }

    getToken();
  }, [clientId]);

  return { token, loading, error };
}

// Component usage
function AppB() {
  const { token, loading, error } = useSSOToken('app-b-client-id');

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Please login</div>;

  return <div>Logged in with token: {token}</div>;
}
```

## Advantages of Token Exchange SSO

### ✅ Pros:

1. **No Cookie Dependency**
   - Works with mobile apps
   - Works with SPA
   - No CORS issues

2. **Explicit Control**
   - Apps explicitly request tokens
   - Clear token lifecycle
   - Easy to debug

3. **Microservices Friendly**
   - Services can exchange tokens
   - Delegation support
   - Service-to-service auth

4. **Scope Control**
   - Can request different scopes per app
   - Scope downgrade built-in
   - Fine-grained permissions

5. **Stateless**
   - No server-side session
   - Scales horizontally
   - JWT-based

### ⚠️ Considerations:

1. **Token Storage**
   - Client must securely store tokens
   - Use secure storage (Keychain, Keystore)
   - Consider token encryption

2. **Token Lifetime**
   - Balance security vs UX
   - Use refresh tokens
   - Implement token rotation

3. **Network Calls**
   - Each exchange requires API call
   - Cache tokens locally
   - Handle offline scenarios

## Security Best Practices

1. **Validate Subject Token**
   - Check signature
   - Check expiration
   - Check issuer

2. **Scope Validation**
   - Only allow scope downgrade
   - Validate against client's allowed scopes
   - Log scope changes

3. **Audience Validation**
   - Set audience in new token
   - Validate audience when using token
   - Prevent token reuse across apps

4. **Rate Limiting**
   - Limit token exchange requests
   - Detect abuse patterns
   - Implement backoff

5. **Audit Logging**
   - Log all token exchanges
   - Track token lineage
   - Monitor suspicious activity

## Comparison: Cookie vs Token Exchange

| Feature | Cookie-based SSO | Token Exchange SSO |
|---------|------------------|-------------------|
| Mobile Support | ❌ Limited | ✅ Full |
| SPA Support | ⚠️ Complex | ✅ Native |
| Cross-domain | ⚠️ Tricky | ✅ Easy |
| Microservices | ❌ No | ✅ Yes |
| Stateless | ❌ No | ✅ Yes |
| Implementation | Simple | Moderate |
| Browser Support | ✅ Native | ✅ API-based |
| Security | ✅ HttpOnly | ⚠️ Client storage |

## Conclusion

Token Exchange SSO เหมาะกับ:
- ✅ Modern applications (SPA, Mobile)
- ✅ Microservices architecture
- ✅ Cross-platform scenarios
- ✅ API-first design

คุณมี Token Exchange handler อยู่แล้วใน `handlers/token_exchange_handler.go` แค่เพิ่ม SSO logic เข้าไปก็ใช้งานได้เลย! 🚀
