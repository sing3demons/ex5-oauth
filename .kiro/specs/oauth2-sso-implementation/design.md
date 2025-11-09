# Design Document

## Overview

This design document describes the implementation of Single Sign-On (SSO) functionality for the OAuth2 Server. The SSO system enables users to authenticate once and seamlessly access multiple client applications without re-entering credentials. The design introduces two new data models (SSO Session and User Consent), two new repositories, middleware for session validation, and updates to existing authorization and authentication handlers.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    OAuth2 Server                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         SSO Session (Browser Cookie)             │  │
│  │  - Session ID (32-byte random string)            │  │
│  │  - User ID                                       │  │
│  │  - Authenticated: true                           │  │
│  │  - Created At / Expires At (7 days)              │  │
│  │  - Last Activity                                 │  │
│  │  - IP Address / User Agent (fingerprinting)     │  │
│  └──────────────────────────────────────────────────┘  │
│                          ↓                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │      User Consent Records                        │  │
│  │  (per user-client pair)                          │  │
│  │  - User ID + Client ID                           │  │
│  │  - Approved Scopes                               │  │
│  │  - Granted At / Expires At (1 year)              │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

```
User Request → SSO Middleware → Authorization Handler
                    ↓                      ↓
              SSO Session            Check Consent
              Repository             Repository
                                          ↓
                                   Auto-approve OR
                                   Show Consent Screen
```


## Components and Interfaces

### 1. Data Models

#### SSO Session Model

Located in `models/models.go`:

```go
type SSOSession struct {
    ID            string    `bson:"_id,omitempty" json:"id"`
    SessionID     string    `bson:"session_id" json:"session_id"`
    UserID        string    `bson:"user_id" json:"user_id"`
    Authenticated bool      `bson:"authenticated" json:"authenticated"`
    CreatedAt     time.Time `bson:"created_at" json:"created_at"`
    ExpiresAt     time.Time `bson:"expires_at" json:"expires_at"`
    LastActivity  time.Time `bson:"last_activity" json:"last_activity"`
    IPAddress     string    `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
    UserAgent     string    `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
}
```

**Design Decisions:**
- `SessionID` is a 32-byte random string used as the cookie value
- `ExpiresAt` is set to 7 days from creation for long-lived sessions
- `IPAddress` and `UserAgent` provide session fingerprinting for security
- `LastActivity` enables activity-based session timeout (future enhancement)

#### User Consent Model

Located in `models/models.go`:

```go
type UserConsent struct {
    ID        string    `bson:"_id,omitempty" json:"id"`
    UserID    string    `bson:"user_id" json:"user_id"`
    ClientID  string    `bson:"client_id" json:"client_id"`
    Scopes    []string  `bson:"scopes" json:"scopes"`
    GrantedAt time.Time `bson:"granted_at" json:"granted_at"`
    ExpiresAt time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}
```

**Design Decisions:**
- Composite key of `UserID` + `ClientID` ensures one consent record per user-client pair
- `Scopes` array stores all approved scopes for the client
- `ExpiresAt` set to 1 year allows long-term consent with periodic re-validation
- Consent is checked before auto-approving authorization requests


### 2. Repository Interfaces

#### SSO Session Repository

Located in `repository/sso_session_repository.go`:

```go
type SSOSessionRepository struct {
    collection *mongo.Collection
}

func NewSSOSessionRepository(db *mongo.Database) *SSOSessionRepository

// Core CRUD operations
func (r *SSOSessionRepository) Create(ctx context.Context, session *SSOSession) error
func (r *SSOSessionRepository) FindBySessionID(ctx context.Context, sessionID string) (*SSOSession, error)
func (r *SSOSessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error
func (r *SSOSessionRepository) Delete(ctx context.Context, sessionID string) error

// Maintenance operations
func (r *SSOSessionRepository) DeleteExpired(ctx context.Context) (int64, error)
func (r *SSOSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*SSOSession, error)
```

**Design Decisions:**
- Follows existing repository pattern used in `SessionRepository`
- `UpdateLastActivity` is optimized for frequent updates without full document replacement
- `DeleteExpired` returns count for monitoring and logging
- `FindByUserID` supports session management UI

#### User Consent Repository

Located in `repository/user_consent_repository.go`:

```go
type UserConsentRepository struct {
    collection *mongo.Collection
}

func NewUserConsentRepository(db *mongo.Database) *UserConsentRepository

// Core operations
func (r *UserConsentRepository) Create(ctx context.Context, consent *UserConsent) error
func (r *UserConsentRepository) FindByUserAndClient(ctx context.Context, userID, clientID string) (*UserConsent, error)
func (r *UserConsentRepository) HasConsent(ctx context.Context, userID, clientID string, scopes []string) (bool, error)

// Management operations
func (r *UserConsentRepository) RevokeConsent(ctx context.Context, userID, clientID string) error
func (r *UserConsentRepository) ListUserConsents(ctx context.Context, userID string) ([]*UserConsent, error)
```

**Design Decisions:**
- `HasConsent` performs scope validation to ensure all requested scopes are approved
- `RevokeConsent` deletes the consent record, requiring re-authorization
- `ListUserConsents` supports user consent management UI
- Uses MongoDB unique index on `user_id + client_id` to prevent duplicates


### 3. SSO Middleware

Located in `middleware/sso_middleware.go`:

```go
func SSOMiddleware(ssoRepo *repository.SSOSessionRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract SSO cookie
            cookie, err := r.Cookie("oauth_sso_session")
            if err == nil {
                // Validate session
                session, err := ssoRepo.FindBySessionID(r.Context(), cookie.Value)
                if err == nil && session.Authenticated && session.ExpiresAt.After(time.Now()) {
                    // Update activity
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

**Design Decisions:**
- Middleware runs before authorization handlers to populate SSO session
- Non-blocking: errors don't prevent request processing
- Updates `LastActivity` on every valid request for session tracking
- Uses context to pass session data to handlers
- Cookie name: `oauth_sso_session` (configurable via constants)

### 4. Cookie Configuration

Located in `handlers/auth_handler.go` (constants):

```go
const (
    SSOCookieName     = "oauth_sso_session"
    SSOCookieMaxAge   = 86400 * 7  // 7 days
    SSOCookiePath     = "/"
    SSOCookieSecure   = true        // HTTPS only in production
    SSOCookieHTTPOnly = true        // Prevent XSS
    SSOCookieSameSite = http.SameSiteLaxMode
)
```

**Design Decisions:**
- `HttpOnly` prevents JavaScript access, mitigating XSS attacks
- `Secure` ensures cookies only sent over HTTPS in production
- `SameSite=Lax` balances security and usability for OAuth flows
- 7-day expiration matches SSO session lifetime
- Path `/` makes cookie available to all OAuth endpoints


### 5. Handler Updates

#### Updated Authorization Handler

Located in `handlers/oauth_handler.go`:

**Key Changes:**
1. Check for SSO session from context at the start of `Authorize()`
2. If SSO session exists and is authenticated:
   - Check for existing user consent using `HasConsent()`
   - If consent exists: auto-generate authorization code and redirect
   - If no consent: show consent screen
3. If no SSO session: redirect to login (existing behavior)
4. Support `prompt` parameter for OIDC compliance

**Pseudo-code:**
```go
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
    // ... existing parameter validation ...
    
    // Check for SSO session
    ssoSession, _ := r.Context().Value("sso_session").(*models.SSOSession)
    
    if ssoSession != nil && ssoSession.Authenticated {
        // Handle prompt parameter
        prompt := r.URL.Query().Get("prompt")
        if prompt == "login" {
            // Force re-authentication
            goto ShowLogin
        }
        
        // Check consent
        hasConsent, _ := h.consentRepo.HasConsent(ctx, ssoSession.UserID, clientID, scopes)
        
        if hasConsent && prompt != "consent" {
            // Auto-approve: generate code and redirect
            generateAuthCodeAndRedirect()
            return
        }
        
        // Show consent screen
        showConsentScreen()
        return
    }
    
ShowLogin:
    // No SSO session - redirect to login
    // ... existing login redirect logic ...
}
```

#### Updated Login Handler

Located in `handlers/auth_handler.go`:

**Key Changes:**
1. After successful authentication, create SSO session
2. Set SSO cookie with secure configuration
3. Continue with existing OAuth session logic

**Pseudo-code:**
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
        ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
        LastActivity:  time.Now(),
        IPAddress:     r.RemoteAddr,
        UserAgent:     r.UserAgent(),
    }
    h.ssoSessionRepo.Create(ctx, ssoSession)
    
    // 2. Set SSO Cookie
    http.SetCookie(w, &http.Cookie{
        Name:     SSOCookieName,
        Value:    ssoSessionID,
        Path:     SSOCookiePath,
        MaxAge:   SSOCookieMaxAge,
        HttpOnly: SSOCookieHTTPOnly,
        Secure:   SSOCookieSecure,
        SameSite: SSOCookieSameSite,
    })
    
    // 3. Continue with existing OAuth flow
    // ... generate authorization code and redirect ...
}
```


### 6. Consent Screen Handler

Located in `handlers/consent_handler.go` (new file):

```go
type ConsentHandler struct {
    clientRepo  *repository.ClientRepository
    consentRepo *repository.UserConsentRepository
    authCodeRepo *repository.AuthCodeRepository
}

func (h *ConsentHandler) ShowConsent(w http.ResponseWriter, r *http.Request)
func (h *ConsentHandler) HandleConsent(w http.ResponseWriter, r *http.Request)
```

**ShowConsent Flow:**
1. Extract parameters from query string (client_id, scope, state, redirect_uri)
2. Fetch client information for display
3. Get scope descriptions from scope registry
4. Render consent template with client name and scope descriptions

**HandleConsent Flow:**
1. Get SSO session from context
2. Parse form data (approved, client_id, scope, state, redirect_uri)
3. If denied: redirect with `access_denied` error
4. If approved:
   - Save consent record with 1-year expiration
   - Generate authorization code
   - Redirect with code and state

**Template Data Structure:**
```go
type ConsentData struct {
    ClientName         string
    ClientID           string
    Scopes             []string
    ScopeDescriptions  []string
    State              string
    RedirectURI        string
}
```

### 7. Logout Handler

Located in `handlers/auth_handler.go`:

```go
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Get SSO session from cookie
    cookie, err := r.Cookie(SSOCookieName)
    if err == nil {
        // Delete from database
        h.ssoSessionRepo.Delete(r.Context(), cookie.Value)
    }
    
    // Clear cookie
    http.SetCookie(w, &http.Cookie{
        Name:     SSOCookieName,
        Value:    "",
        Path:     SSOCookiePath,
        MaxAge:   -1,
        HttpOnly: SSOCookieHTTPOnly,
        Secure:   SSOCookieSecure,
        SameSite: SSOCookieSameSite,
    })
    
    // Support OIDC post_logout_redirect_uri
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

**Design Decisions:**
- Deletes SSO session from database to invalidate across all devices
- Clears cookie by setting MaxAge to -1
- Supports OIDC `post_logout_redirect_uri` parameter
- Returns JSON response if no redirect URI provided


### 8. Session Management Endpoints

Located in `handlers/session_handler.go` (new file):

```go
type SessionHandler struct {
    ssoSessionRepo *repository.SSOSessionRepository
    consentRepo    *repository.UserConsentRepository
}

// GET /account/sessions - List active SSO sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request)

// DELETE /account/sessions/{session_id} - Revoke specific session
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request)

// GET /account/authorizations - List authorized applications
func (h *SessionHandler) ListAuthorizations(w http.ResponseWriter, r *http.Request)

// DELETE /account/authorizations/{client_id} - Revoke app authorization
func (h *SessionHandler) RevokeAuthorization(w http.ResponseWriter, r *http.Request)
```

**Authentication:**
- All endpoints require valid access token
- Extract user ID from JWT token
- Only allow users to manage their own sessions/consents

**Response Formats:**

ListSessions:
```json
{
  "sessions": [
    {
      "session_id": "abc123...",
      "created_at": "2025-11-09T10:00:00Z",
      "last_activity": "2025-11-09T15:30:00Z",
      "expires_at": "2025-11-16T10:00:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0..."
    }
  ]
}
```

ListAuthorizations:
```json
{
  "authorizations": [
    {
      "client_id": "my-app",
      "client_name": "My Application",
      "scopes": ["openid", "profile", "email"],
      "granted_at": "2025-11-01T10:00:00Z",
      "expires_at": "2026-11-01T10:00:00Z"
    }
  ]
}
```


## Data Models

### Database Collections

#### sso_sessions Collection

```javascript
{
  "_id": ObjectId("..."),
  "session_id": "abc123def456...",  // 32-byte random string
  "user_id": "user123",
  "authenticated": true,
  "created_at": ISODate("2025-11-09T10:00:00Z"),
  "expires_at": ISODate("2025-11-16T10:00:00Z"),
  "last_activity": ISODate("2025-11-09T15:30:00Z"),
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

**Indexes:**
- `session_id`: unique index for fast lookup
- `user_id`: index for listing user sessions
- `expires_at`: index for efficient cleanup of expired sessions

#### user_consents Collection

```javascript
{
  "_id": ObjectId("..."),
  "user_id": "user123",
  "client_id": "my-app",
  "scopes": ["openid", "profile", "email"],
  "granted_at": ISODate("2025-11-09T10:00:00Z"),
  "expires_at": ISODate("2026-11-09T10:00:00Z")
}
```

**Indexes:**
- `user_id + client_id`: unique compound index
- `user_id`: index for listing user consents
- `expires_at`: index for cleanup (future enhancement)

### State Transitions

#### SSO Session Lifecycle

```
[Created] → [Active] → [Expired]
              ↓
          [Revoked]
```

- **Created**: Session created after successful authentication
- **Active**: Session is valid and can be used for authorization
- **Expired**: Session exceeded 7-day lifetime or inactivity timeout
- **Revoked**: User explicitly logged out or session was revoked

#### User Consent Lifecycle

```
[Granted] → [Active] → [Expired]
              ↓
          [Revoked]
```

- **Granted**: User approved authorization request
- **Active**: Consent is valid and enables auto-approval
- **Expired**: Consent exceeded 1-year lifetime (requires re-consent)
- **Revoked**: User explicitly revoked application access


## Error Handling

### SSO-Specific Error Scenarios

#### 1. Invalid SSO Session

**Scenario:** SSO cookie exists but session not found in database or expired

**Handling:**
- Middleware silently ignores invalid session
- Authorization handler treats request as unauthenticated
- User redirected to login page
- No error response to client

**Rationale:** Expired sessions are normal; don't expose internal state

#### 2. Missing Consent

**Scenario:** Valid SSO session but no consent for requested client/scopes

**Handling:**
- Display consent screen
- User can approve or deny
- If denied: redirect with `error=access_denied`

#### 3. Prompt Parameter Conflicts

**Scenario:** `prompt=none` but user not authenticated or consent missing

**Handling:**
- Return error immediately without redirecting to login
- Error: `login_required` or `consent_required`
- Follow OIDC specification for prompt parameter

**Example Response:**
```
HTTP/1.1 302 Found
Location: https://client.example.com/callback?
  error=login_required&
  error_description=User+authentication+required&
  state=xyz
```

#### 4. Concurrent Session Limit (Future Enhancement)

**Scenario:** User exceeds maximum allowed concurrent sessions

**Handling:**
- Revoke oldest session automatically
- Log security event
- Continue with new session creation

#### 5. Session Hijacking Detection (Future Enhancement)

**Scenario:** IP address or user agent changes during session

**Handling:**
- Log security warning
- Optionally require re-authentication
- Configurable strictness level

### Error Response Format

All SSO-related errors follow OAuth2 error response format:

```json
{
  "error": "error_code",
  "error_description": "Human-readable description"
}
```

**Common Error Codes:**
- `invalid_request`: Malformed request parameters
- `access_denied`: User denied consent
- `login_required`: Authentication required (prompt=none)
- `consent_required`: Consent required (prompt=none)
- `server_error`: Internal server error


## Testing Strategy

### Unit Tests

#### Repository Tests

**SSOSessionRepository:**
- `TestCreate`: Verify session creation with all fields
- `TestFindBySessionID`: Test successful lookup and not found cases
- `TestUpdateLastActivity`: Verify timestamp update
- `TestDelete`: Verify session deletion
- `TestDeleteExpired`: Verify expired sessions are removed, active sessions remain
- `TestFindByUserID`: Verify listing all sessions for a user

**UserConsentRepository:**
- `TestCreate`: Verify consent creation
- `TestFindByUserAndClient`: Test lookup by composite key
- `TestHasConsent`: Verify scope matching logic (exact match, subset, superset)
- `TestRevokeConsent`: Verify deletion
- `TestListUserConsents`: Verify listing all consents for a user

#### Middleware Tests

**SSOMiddleware:**
- `TestValidSession`: Verify session loaded into context
- `TestExpiredSession`: Verify expired session ignored
- `TestMissingCookie`: Verify request continues without session
- `TestInvalidSessionID`: Verify invalid session ignored
- `TestLastActivityUpdate`: Verify activity timestamp updated

#### Handler Tests

**Authorization Handler:**
- `TestAuthorizeWithSSOAndConsent`: Verify auto-approval flow
- `TestAuthorizeWithSSONoConsent`: Verify consent screen shown
- `TestAuthorizeWithoutSSO`: Verify login redirect
- `TestAuthorizePromptLogin`: Verify forced re-authentication
- `TestAuthorizePromptConsent`: Verify forced consent screen
- `TestAuthorizePromptNone`: Verify error when not authenticated

**Login Handler:**
- `TestLoginCreatesSSOSession`: Verify SSO session created
- `TestLoginSetsCookie`: Verify cookie set with correct attributes
- `TestLoginWithExistingSSO`: Verify existing session handling

**Consent Handler:**
- `TestShowConsent`: Verify template rendering with correct data
- `TestHandleConsentApproved`: Verify consent saved and code generated
- `TestHandleConsentDenied`: Verify error redirect

**Logout Handler:**
- `TestLogoutDeletesSession`: Verify session deleted from database
- `TestLogoutClearsCookie`: Verify cookie cleared
- `TestLogoutWithRedirect`: Verify post_logout_redirect_uri handling

### Integration Tests

#### End-to-End SSO Flow

**Test 1: First Login**
1. User visits App A authorization endpoint
2. No SSO session exists
3. Redirected to login page
4. User enters credentials
5. SSO session created, cookie set
6. Consent screen shown
7. User approves
8. Consent saved
9. Authorization code generated
10. Redirected to App A with code

**Test 2: Second App with SSO**
1. User visits App B authorization endpoint
2. SSO session exists and valid
3. Consent exists for App B
4. Authorization code generated immediately
5. Redirected to App B with code (no login/consent screens)

**Test 3: Logout**
1. User calls logout endpoint
2. SSO session deleted
3. Cookie cleared
4. User visits App C authorization endpoint
5. Redirected to login (SSO session gone)

**Test 4: Expired Session**
1. Create SSO session with past expiration
2. User visits authorization endpoint
3. Session ignored (expired)
4. Redirected to login

**Test 5: Consent Revocation**
1. User has active SSO session and consent for App A
2. User revokes consent via API
3. User visits App A authorization endpoint
4. Consent screen shown (consent revoked)

### Performance Tests

**Metrics to Measure:**
- SSO session lookup time: < 10ms
- Consent check time: < 10ms
- Auto-approval flow (SSO + consent): < 100ms total
- Session cleanup for 10,000 expired sessions: < 5 seconds

### Security Tests

**Test Scenarios:**
- Cookie attributes (HttpOnly, Secure, SameSite)
- Session expiration enforcement
- Consent scope validation
- CSRF protection via state parameter
- Session hijacking detection (IP/User-Agent changes)


## Implementation Phases

### Phase 1: Core SSO Session Management (MVP)

**Deliverables:**
- SSO Session model and repository
- SSO middleware
- Updated login handler to create SSO sessions
- Updated authorization handler to check SSO sessions
- Basic logout functionality

**Success Criteria:**
- User can log in once and access multiple apps
- SSO sessions persist for 7 days
- Logout clears SSO session

### Phase 2: Consent Management

**Deliverables:**
- User Consent model and repository
- Consent screen template
- Consent handler (show and handle)
- Auto-approval logic in authorization handler

**Success Criteria:**
- Users see consent screen on first app authorization
- Subsequent authorizations auto-approve with existing consent
- Consent persists for 1 year

### Phase 3: Session Management UI

**Deliverables:**
- Session management endpoints
- Authorization management endpoints
- API documentation

**Success Criteria:**
- Users can list active sessions
- Users can revoke specific sessions
- Users can list and revoke app authorizations

### Phase 4: Advanced Features

**Deliverables:**
- OIDC prompt parameter support
- Session fingerprinting and hijacking detection
- Activity-based session timeout
- Concurrent session limits
- Expired session cleanup job

**Success Criteria:**
- Full OIDC compliance for prompt parameter
- Security events logged for suspicious activity
- Automatic cleanup of expired sessions

## Migration Strategy

### Backward Compatibility

**Existing Behavior Preserved:**
- Users without SSO sessions continue to work normally
- Existing OAuth flows unchanged for first-time users
- No breaking changes to client applications

**Gradual Rollout:**
1. Deploy SSO code with feature flag disabled
2. Enable SSO for internal testing
3. Monitor metrics (session creation, auto-approvals)
4. Enable SSO for all users
5. Monitor for issues and rollback if needed

### Database Migration

**No migration required:**
- New collections created automatically
- Existing collections unchanged
- No data transformation needed

**Indexes to Create:**
```javascript
// sso_sessions collection
db.sso_sessions.createIndex({ "session_id": 1 }, { unique: true })
db.sso_sessions.createIndex({ "user_id": 1 })
db.sso_sessions.createIndex({ "expires_at": 1 })

// user_consents collection
db.user_consents.createIndex({ "user_id": 1, "client_id": 1 }, { unique: true })
db.user_consents.createIndex({ "user_id": 1 })
```

## Configuration

### Environment Variables

```bash
# SSO Configuration
SSO_ENABLED=true
SSO_SESSION_EXPIRY_DAYS=7
SSO_COOKIE_SECURE=true  # Set to false for local development
SSO_CONSENT_EXPIRY_DAYS=365

# Security
SSO_MAX_CONCURRENT_SESSIONS=5  # Future enhancement
SSO_ACTIVITY_TIMEOUT_MINUTES=30  # Future enhancement
```

### Feature Flags

```go
type SSOConfig struct {
    Enabled                bool
    SessionExpiryDays      int
    ConsentExpiryDays      int
    CookieSecure           bool
    MaxConcurrentSessions  int
    ActivityTimeoutMinutes int
}
```

## Monitoring and Observability

### Metrics to Track

**Session Metrics:**
- Active SSO sessions count
- SSO session creation rate
- SSO session expiration rate
- Average session lifetime

**Authorization Metrics:**
- Auto-approval rate (SSO + consent)
- Consent screen display rate
- Consent approval rate
- Consent denial rate

**Performance Metrics:**
- SSO session lookup latency (p50, p95, p99)
- Consent check latency (p50, p95, p99)
- Auto-approval flow latency (p50, p95, p99)

### Logging

**Log Events:**
- SSO session created (user_id, session_id, ip_address)
- SSO session expired (user_id, session_id)
- SSO session revoked (user_id, session_id, reason)
- Consent granted (user_id, client_id, scopes)
- Consent revoked (user_id, client_id)
- Auto-approval (user_id, client_id)
- Security events (session hijacking attempts, etc.)

**Log Format:**
```json
{
  "timestamp": "2025-11-09T15:30:00Z",
  "level": "info",
  "event": "sso_session_created",
  "user_id": "user123",
  "session_id": "abc123...",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

## Security Considerations

### Cookie Security

- **HttpOnly**: Prevents JavaScript access to mitigate XSS
- **Secure**: Ensures cookies only sent over HTTPS
- **SameSite=Lax**: Balances security and OAuth redirect compatibility
- **Path=/**: Scoped to entire OAuth server

### Session Security

- **Random Session IDs**: 32-byte cryptographically random strings
- **Expiration**: 7-day maximum lifetime
- **Fingerprinting**: IP address and user agent stored for detection
- **Database Storage**: Sessions stored server-side, not in cookie

### Consent Security

- **Scope Validation**: Requested scopes validated against stored consent
- **Expiration**: 1-year maximum lifetime requires periodic re-consent
- **Revocation**: Users can revoke consent at any time
- **Audit Trail**: Consent grants and revocations logged

### CSRF Protection

- **State Parameter**: Required for all authorization requests
- **Validated on Callback**: State parameter validated before issuing code
- **Nonce Support**: OIDC nonce included in ID tokens

## Future Enhancements

### Activity-Based Timeout

- Track last activity timestamp
- Auto-expire sessions after 30 minutes of inactivity
- Configurable timeout period

### Concurrent Session Limits

- Limit number of active sessions per user
- Revoke oldest session when limit exceeded
- Configurable limit per user or globally

### Session Hijacking Detection

- Compare IP address and user agent on each request
- Flag suspicious changes
- Optionally require re-authentication

### Remember Me

- Extend session lifetime for "remember me" option
- Separate cookie for long-lived sessions
- Configurable maximum lifetime (e.g., 30 days)

### Multi-Factor Authentication

- Require MFA for sensitive operations
- Store MFA status in SSO session
- Step-up authentication for high-risk actions
