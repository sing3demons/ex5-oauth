# Single Sign-On (SSO) Design

## Overview

การทำให้ OAuth2 server รองรับ SSO จะช่วยให้ผู้ใช้ login ครั้งเดียว แล้วสามารถเข้าใช้งานหลาย applications ได้โดยไม่ต้อง login ซ้ำ

## Current State vs SSO State

### ปัจจุบัน (Without SSO)
```
User → App A → OAuth Server → Login → App A
User → App B → OAuth Server → Login Again! → App B
User → App C → OAuth Server → Login Again! → App C
```

### หลังมี SSO
```
User → App A → OAuth Server → Login (first time) → App A
User → App B → OAuth Server → Auto approve (already logged in) → App B
User → App C → OAuth Server → Auto approve (already logged in) → App C
```

## Architecture Design

### 1. SSO Session Management

```
┌─────────────────────────────────────────────────────────┐
│                    OAuth2 Server                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         SSO Session (Browser Cookie)             │  │
│  │  - Session ID                                    │  │
│  │  - User ID                                       │  │
│  │  - Authenticated: true                           │  │
│  │  - Created At / Expires At                       │  │
│  │  - Last Activity                                 │  │
│  └──────────────────────────────────────────────────┘  │
│                          ↓                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │      OAuth Authorization Sessions                │  │
│  │  (per application authorization request)         │  │
│  │  - Links to SSO Session                          │  │
│  │  - Client ID                                     │  │
│  │  - Requested Scopes                              │  │
│  │  - Consent Given                                 │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 2. Data Models

#### SSO Session Model
```go
type SSOSession struct {
    ID            string    `bson:"_id,omitempty" json:"id"`
    SessionID     string    `bson:"session_id" json:"session_id"`           // Cookie value
    UserID        string    `bson:"user_id" json:"user_id"`
    Authenticated bool      `bson:"authenticated" json:"authenticated"`
    CreatedAt     time.Time `bson:"created_at" json:"created_at"`
    ExpiresAt     time.Time `bson:"expires_at" json:"expires_at"`
    LastActivity  time.Time `bson:"last_activity" json:"last_activity"`
    IPAddress     string    `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
    UserAgent     string    `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
}
```

#### User Consent Model (for remembering app permissions)
```go
type UserConsent struct {
    ID        string    `bson:"_id,omitempty" json:"id"`
    UserID    string    `bson:"user_id" json:"user_id"`
    ClientID  string    `bson:"client_id" json:"client_id"`
    Scopes    []string  `bson:"scopes" json:"scopes"`              // Approved scopes
    GrantedAt time.Time `bson:"granted_at" json:"granted_at"`
    ExpiresAt time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}
```

### 3. SSO Flow

#### First Login (No SSO Session)
```
1. User → App A → /oauth/authorize
2. OAuth Server checks SSO cookie → Not found
3. Redirect to /auth/login with return_url
4. User enters credentials
5. Create SSO Session + Set HTTP-only cookie
6. Create OAuth Session (linked to SSO Session)
7. Show consent screen (if needed)
8. Generate authorization code
9. Redirect back to App A with code
```

#### Subsequent Login (SSO Session Exists)
```
1. User → App B → /oauth/authorize
2. OAuth Server checks SSO cookie → Found & Valid
3. Check if user already consented to App B
   - If YES: Skip consent, generate code immediately
   - If NO: Show consent screen
4. Generate authorization code
5. Redirect back to App B with code
```

### 4. Cookie Configuration

```go
// SSO Cookie settings
const (
    SSOCookieName     = "oauth_sso_session"
    SSOCookieMaxAge   = 86400 * 7  // 7 days
    SSOCookiePath     = "/"
    SSOCookieSecure   = true        // HTTPS only in production
    SSOCookieHTTPOnly = true        // Prevent XSS
    SSOCookieSameSite = http.SameSiteLaxMode
)
```

### 5. Implementation Components

#### A. SSO Session Repository
```go
type SSOSessionRepository interface {
    Create(ctx context.Context, session *SSOSession) error
    FindBySessionID(ctx context.Context, sessionID string) (*SSOSession, error)
    UpdateLastActivity(ctx context.Context, sessionID string) error
    Delete(ctx context.Context, sessionID string) error
    DeleteExpired(ctx context.Context) error
}
```

#### B. User Consent Repository
```go
type UserConsentRepository interface {
    Create(ctx context.Context, consent *UserConsent) error
    FindByUserAndClient(ctx context.Context, userID, clientID string) (*UserConsent, error)
    HasConsent(ctx context.Context, userID, clientID string, scopes []string) (bool, error)
    RevokeConsent(ctx context.Context, userID, clientID string) error
    ListUserConsents(ctx context.Context, userID string) ([]*UserConsent, error)
}
```

#### C. SSO Middleware
```go
func SSOMiddleware(ssoRepo *SSOSessionRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check SSO cookie
            cookie, err := r.Cookie(SSOCookieName)
            if err == nil {
                // Validate SSO session
                session, err := ssoRepo.FindBySessionID(r.Context(), cookie.Value)
                if err == nil && session.Authenticated && session.ExpiresAt.After(time.Now()) {
                    // Update last activity
                    ssoRepo.UpdateLastActivity(r.Context(), session.SessionID)
                    
                    // Add to context
                    ctx := context.WithValue(r.Context(), "sso_session", session)
                    r = r.WithContext(ctx)
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### D. Updated Authorization Handler
```go
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
    // ... existing parameter validation ...
    
    // Check for existing SSO session
    ssoSession := r.Context().Value("sso_session").(*SSOSession)
    
    if ssoSession != nil && ssoSession.Authenticated {
        // User is already logged in via SSO
        
        // Check if user has already consented to this client
        hasConsent, err := h.consentRepo.HasConsent(
            ctx, 
            ssoSession.UserID, 
            clientID, 
            strings.Split(scope, " "),
        )
        
        if err == nil && hasConsent {
            // Auto-approve: Skip login and consent
            code, _ := utils.GenerateRandomString(16)
            authCode := &models.AuthorizationCode{
                Code:        code,
                ClientID:    clientID,
                UserID:      ssoSession.UserID,
                RedirectURI: redirectURI,
                Scope:       scope,
                Nonce:       nonce,
                ExpiresAt:   time.Now().Add(10 * time.Minute),
            }
            h.authCodeRepo.Create(ctx, authCode)
            
            // Redirect immediately
            redirectURL, _ := url.Parse(redirectURI)
            q := redirectURL.Query()
            q.Set("code", code)
            if state != "" {
                q.Set("state", state)
            }
            redirectURL.RawQuery = q.Encode()
            http.Redirect(w, r, redirectURL.String(), http.StatusFound)
            return
        }
        
        // User is logged in but hasn't consented to this app
        // Show consent screen
        h.showConsentScreen(w, r, ssoSession.UserID, clientID, scope, state)
        return
    }
    
    // No SSO session - redirect to login
    // ... existing login redirect logic ...
}
```

#### E. Updated Login Handler
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... existing authentication logic ...
    
    // After successful authentication:
    
    // 1. Create SSO Session
    ssoSessionID, _ := utils.GenerateRandomString(32)
    ssoSession := &models.SSOSession{
        SessionID:     ssoSessionID,
        UserID:        user.ID,
        Authenticated: true,
        CreatedAt:     time.Now(),
        ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // 7 days
        LastActivity:  time.Now(),
        IPAddress:     r.RemoteAddr,
        UserAgent:     r.UserAgent(),
    }
    h.ssoSessionRepo.Create(ctx, ssoSession)
    
    // 2. Set SSO Cookie
    http.SetCookie(w, &http.Cookie{
        Name:     SSOCookieName,
        Value:    ssoSessionID,
        Path:     "/",
        MaxAge:   86400 * 7, // 7 days
        HttpOnly: true,
        Secure:   true, // HTTPS only
        SameSite: http.SameSiteLaxMode,
    })
    
    // 3. Continue with OAuth flow
    // ... existing OAuth session logic ...
}
```

#### F. Consent Screen Handler
```go
func (h *OAuthHandler) ShowConsent(w http.ResponseWriter, r *http.Request) {
    // Get parameters from query or session
    clientID := r.URL.Query().Get("client_id")
    scope := r.URL.Query().Get("scope")
    
    // Get client info
    client, _ := h.clientRepo.FindByClientID(r.Context(), clientID)
    
    // Get scope descriptions
    scopeList := strings.Split(scope, " ")
    scopeDescriptions := make([]string, 0)
    for _, s := range scopeList {
        if scopeDef := models.GlobalScopeRegistry.GetScope(s); scopeDef != nil {
            scopeDescriptions = append(scopeDescriptions, scopeDef.Description)
        }
    }
    
    // Render consent page
    data := map[string]interface{}{
        "ClientName":         client.Name,
        "Scopes":            scopeList,
        "ScopeDescriptions": scopeDescriptions,
    }
    
    tmpl, _ := template.ParseFiles("templates/consent.html")
    tmpl.Execute(w, data)
}

func (h *OAuthHandler) HandleConsent(w http.ResponseWriter, r *http.Request) {
    // Get SSO session
    ssoSession := r.Context().Value("sso_session").(*SSOSession)
    
    // Get form data
    approved := r.FormValue("approved")
    clientID := r.FormValue("client_id")
    scope := r.FormValue("scope")
    state := r.FormValue("state")
    redirectURI := r.FormValue("redirect_uri")
    
    if approved != "true" {
        // User denied consent
        redirectURL, _ := url.Parse(redirectURI)
        q := redirectURL.Query()
        q.Set("error", "access_denied")
        q.Set("error_description", "User denied consent")
        if state != "" {
            q.Set("state", state)
        }
        redirectURL.RawQuery = q.Encode()
        http.Redirect(w, r, redirectURL.String(), http.StatusFound)
        return
    }
    
    // Save consent
    consent := &models.UserConsent{
        UserID:    ssoSession.UserID,
        ClientID:  clientID,
        Scopes:    strings.Split(scope, " "),
        GrantedAt: time.Now(),
        ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year
    }
    h.consentRepo.Create(r.Context(), consent)
    
    // Generate authorization code
    code, _ := utils.GenerateRandomString(16)
    authCode := &models.AuthorizationCode{
        Code:        code,
        ClientID:    clientID,
        UserID:      ssoSession.UserID,
        RedirectURI: redirectURI,
        Scope:       scope,
        ExpiresAt:   time.Now().Add(10 * time.Minute),
    }
    h.authCodeRepo.Create(r.Context(), authCode)
    
    // Redirect with code
    redirectURL, _ := url.Parse(redirectURI)
    q := redirectURL.Query()
    q.Set("code", code)
    if state != "" {
        q.Set("state", state)
    }
    redirectURL.RawQuery = q.Encode()
    http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
```

#### G. Logout Handler
```go
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Get SSO session from cookie
    cookie, err := r.Cookie(SSOCookieName)
    if err == nil {
        // Delete SSO session from database
        h.ssoSessionRepo.Delete(r.Context(), cookie.Value)
    }
    
    // Clear SSO cookie
    http.SetCookie(w, &http.Cookie{
        Name:     SSOCookieName,
        Value:    "",
        Path:     "/",
        MaxAge:   -1, // Delete cookie
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    })
    
    // Support post_logout_redirect_uri (OIDC)
    redirectURI := r.URL.Query().Get("post_logout_redirect_uri")
    if redirectURI != "" {
        http.Redirect(w, r, redirectURI, http.StatusFound)
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Logged out successfully",
    })
}
```

### 6. Consent Screen Template

```html
<!-- templates/consent.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Required</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
        .app-info { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .scopes { list-style: none; padding: 0; }
        .scopes li { padding: 10px; margin: 5px 0; background: #fff; border: 1px solid #ddd; border-radius: 3px; }
        .buttons { margin-top: 20px; }
        button { padding: 10px 20px; margin-right: 10px; cursor: pointer; }
        .approve { background: #4CAF50; color: white; border: none; }
        .deny { background: #f44336; color: white; border: none; }
    </style>
</head>
<body>
    <h2>Authorization Required</h2>
    
    <div class="app-info">
        <strong>{{.ClientName}}</strong> is requesting access to your account.
    </div>
    
    <h3>This application will be able to:</h3>
    <ul class="scopes">
        {{range .ScopeDescriptions}}
        <li>✓ {{.}}</li>
        {{end}}
    </ul>
    
    <form method="POST" action="/oauth/consent">
        <input type="hidden" name="client_id" value="{{.ClientID}}">
        <input type="hidden" name="scope" value="{{.Scope}}">
        <input type="hidden" name="state" value="{{.State}}">
        <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
        
        <div class="buttons">
            <button type="submit" name="approved" value="true" class="approve">Allow</button>
            <button type="submit" name="approved" value="false" class="deny">Deny</button>
        </div>
    </form>
</body>
</html>
```

### 7. Additional Features

#### A. Session Management Endpoints

```go
// List active SSO sessions for a user
GET /account/sessions

// Revoke specific SSO session
DELETE /account/sessions/{session_id}

// List authorized applications
GET /account/authorizations

// Revoke application authorization
DELETE /account/authorizations/{client_id}
```

#### B. Security Enhancements

1. **Session Fingerprinting**: Store IP + User-Agent to detect session hijacking
2. **Activity Timeout**: Auto-logout after inactivity (e.g., 30 minutes)
3. **Concurrent Session Limit**: Limit number of active sessions per user
4. **Device Tracking**: Show user which devices are logged in

#### C. Prompt Parameter Support (OIDC)

```go
// Support prompt parameter in authorization request
prompt := r.URL.Query().Get("prompt")

switch prompt {
case "none":
    // Don't show login or consent - fail if not authenticated
case "login":
    // Force re-authentication even if SSO session exists
case "consent":
    // Force consent screen even if already consented
case "select_account":
    // Show account selection (if multiple accounts)
}
```

## Implementation Checklist

- [ ] 1. Create SSO Session model and repository
- [ ] 2. Create User Consent model and repository
- [ ] 3. Implement SSO middleware
- [ ] 4. Update authorization handler for SSO check
- [ ] 5. Update login handler to create SSO session
- [ ] 6. Create consent screen template
- [ ] 7. Implement consent handlers
- [ ] 8. Implement logout handler
- [ ] 9. Add session management endpoints
- [ ] 10. Add prompt parameter support
- [ ] 11. Add security features (fingerprinting, timeouts)
- [ ] 12. Add tests for SSO flows
- [ ] 13. Update documentation

## Benefits

✅ **Better UX**: Login once, access all apps  
✅ **Security**: Centralized authentication  
✅ **Control**: Users can see and revoke app access  
✅ **Compliance**: Follows OIDC best practices  
✅ **Scalability**: Easy to add new applications  

## Migration Path

1. **Phase 1**: Implement SSO sessions (backward compatible)
2. **Phase 2**: Add consent management
3. **Phase 3**: Add session management UI
4. **Phase 4**: Add advanced security features

## Testing Scenarios

1. ✅ First login creates SSO session
2. ✅ Second app uses existing SSO session
3. ✅ Consent is remembered across sessions
4. ✅ Logout clears SSO session
5. ✅ Expired SSO session requires re-login
6. ✅ Prompt=login forces re-authentication
7. ✅ Prompt=consent forces consent screen
8. ✅ Session hijacking detection works
