# Design Document: Enhanced Scope Management System

## Overview

ออกแบบระบบจัดการ scope ที่ครบถ้วนและเป็นมาตรฐานสำหรับ OAuth2/OIDC Server โดยเน้น:
- Scope Registry สำหรับจัดเก็บและจัดการ scope definitions
- Client-specific scope restrictions
- Scope validation และ normalization
- Claim filtering based on scopes
- Support for scope hierarchy และ dynamic parameters

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     OAuth2 Server                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │   Handlers   │─────▶│ Scope Service│                     │
│  └──────────────┘      └──────┬───────┘                     │
│         │                      │                              │
│         │                      ▼                              │
│         │              ┌──────────────┐                      │
│         │              │Scope Registry│                      │
│         │              └──────────────┘                      │
│         │                      │                              │
│         ▼                      ▼                              │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │  Repository  │      │  Validators  │                     │
│  └──────────────┘      └──────────────┘                     │
│         │                                                     │
│         ▼                                                     │
│  ┌──────────────┐                                            │
│  │   MongoDB    │                                            │
│  └──────────────┘                                            │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Scope Registry

**Purpose:** จัดเก็บและจัดการ scope definitions

**Data Structure:**
```go
type ScopeDefinition struct {
    Name        string   // scope name (e.g., "openid", "profile")
    Description string   // human-readable description
    Claims      []string // claims included in this scope
    IsDefault   bool     // included in default scope set
    ParentScope string   // parent scope (for hierarchy)
    ChildScopes []string // child scopes (for hierarchy)
}

type ScopeRegistry struct {
    Scopes map[string]*ScopeDefinition
}
```

**Methods:**
```go
// Initialize with standard OIDC scopes
func NewScopeRegistry() *ScopeRegistry

// Register a new scope
func (r *ScopeRegistry) RegisterScope(scope *ScopeDefinition)

// Get scope definition
func (r *ScopeRegistry) GetScope(name string) (*ScopeDefinition, bool)

// Validate scope exists
func (r *ScopeRegistry) IsValidScope(name string) bool

// Get all scopes
func (r *ScopeRegistry) GetAllScopes() []*ScopeDefinition

// Get default scopes
func (r *ScopeRegistry) GetDefaultScopes() []string

// Get claims for given scopes
func (r *ScopeRegistry) GetClaimsForScopes(scopes []string) []string

// Expand parent scopes to include children
func (r *ScopeRegistry) ExpandScopes(scopes []string) []string
```

**Standard OIDC Scopes:**
- `openid`: Required for OIDC, includes `sub` claim
- `profile`: User profile info (name, picture, etc.)
- `email`: Email address and verification status
- `phone`: Phone number and verification status
- `address`: Postal address
- `offline_access`: Request refresh token

### 2. Client Model Enhancement

**Current:**
```go
type Client struct {
    ClientID     string
    ClientSecret string
    RedirectURIs []string
    Name         string
}
```

**Enhanced:**
```go
type Client struct {
    ClientID      string
    ClientSecret  string
    RedirectURIs  []string
    Name          string
    AllowedScopes []string  // NEW: scopes this client can request
    GrantTypes    []string  // NEW: allowed grant types
    CreatedAt     time.Time
}
```

### 3. Authorization Code Enhancement

**Current:**
```go
type AuthorizationCode struct {
    Code        string
    ClientID    string
    UserID      string
    RedirectURI string
    Scope       string
    ExpiresAt   time.Time
}
```

**Enhanced:**
```go
type AuthorizationCode struct {
    Code            string
    ClientID        string
    UserID          string
    RedirectURI     string
    Scope           string
    Nonce           string  // NEW: for ID token replay protection
    CodeChallenge   string  // NEW: for PKCE
    ChallengeMethod string  // NEW: S256 or plain
    ExpiresAt       time.Time
}
```

### 4. Scope Validation Service

**Interface:**
```go
type ScopeValidator interface {
    // Validate scope format and existence
    ValidateScope(scope string) error
    
    // Normalize scope (remove duplicates, trim)
    NormalizeScope(scope string) string
    
    // Validate against client's allowed scopes
    ValidateClientScopes(requested string, client *Client) error
    
    // Validate scope downgrade (for refresh)
    ValidateScopeDowngrade(requested, original string) error
    
    // Check if openid scope is present
    RequiresOpenID(scope string) bool
}
```

**Implementation:**
```go
type scopeValidator struct {
    registry *ScopeRegistry
}

func (v *scopeValidator) ValidateScope(scope string) error {
    if scope == "" {
        return errors.New("scope cannot be empty")
    }
    
    scopes := strings.Split(scope, " ")
    for _, s := range scopes {
        if !v.registry.IsValidScope(s) {
            return fmt.Errorf("invalid scope: %s", s)
        }
    }
    return nil
}

func (v *scopeValidator) ValidateClientScopes(requested string, client *Client) error {
    if len(client.AllowedScopes) == 0 {
        return nil // no restrictions
    }
    
    allowedMap := make(map[string]bool)
    for _, s := range client.AllowedScopes {
        allowedMap[s] = true
    }
    
    requestedScopes := strings.Split(requested, " ")
    var unauthorized []string
    
    for _, s := range requestedScopes {
        if !allowedMap[s] {
            unauthorized = append(unauthorized, s)
        }
    }
    
    if len(unauthorized) > 0 {
        return fmt.Errorf("unauthorized scopes: %v", unauthorized)
    }
    return nil
}
```

### 5. Claim Filtering Service

**Purpose:** Filter user claims based on granted scopes

**Interface:**
```go
type ClaimFilter interface {
    // Get claims for user based on scopes
    FilterClaims(user *User, scopes string) map[string]interface{}
    
    // Get claims for ID token
    GetIDTokenClaims(user *User, scopes string) map[string]interface{}
}
```

**Implementation:**
```go
func (f *claimFilter) FilterClaims(user *User, scopes string) map[string]interface{} {
    claims := make(map[string]interface{})
    
    // Always include sub
    claims["sub"] = user.ID
    
    scopeList := strings.Split(scopes, " ")
    allowedClaims := f.registry.GetClaimsForScopes(scopeList)
    
    claimMap := make(map[string]bool)
    for _, c := range allowedClaims {
        claimMap[c] = true
    }
    
    // Add claims based on scope
    if claimMap["email"] {
        claims["email"] = user.Email
        claims["email_verified"] = true
    }
    
    if claimMap["name"] {
        claims["name"] = user.Name
    }
    
    // ... more claims
    
    return claims
}
```

## Data Models

### Database Schema

**clients collection:**
```json
{
  "_id": ObjectId,
  "client_id": "string",
  "client_secret": "hashed_string",
  "redirect_uris": ["string"],
  "name": "string",
  "allowed_scopes": ["openid", "profile", "email"],
  "grant_types": ["authorization_code", "refresh_token"],
  "created_at": ISODate
}
```

**authorization_codes collection:**
```json
{
  "code": "string",
  "client_id": "string",
  "user_id": "string",
  "redirect_uri": "string",
  "scope": "openid profile email",
  "nonce": "string",
  "code_challenge": "string",
  "challenge_method": "S256",
  "expires_at": ISODate,
  "created_at": ISODate
}
```

**user_consents collection (NEW):**
```json
{
  "_id": ObjectId,
  "user_id": "string",
  "client_id": "string",
  "scopes": ["openid", "profile", "email"],
  "granted_at": ISODate,
  "expires_at": ISODate
}
```

## Error Handling

### Error Types

```go
type ScopeError struct {
    Code        string // "invalid_scope", "unauthorized_scope"
    Description string
    Details     interface{}
}
```

### Error Responses

**Invalid Scope:**
```json
{
  "error": "invalid_scope",
  "error_description": "Requested scope 'invalid_scope_name' is not supported"
}
```

**Unauthorized Scope:**
```json
{
  "error": "invalid_scope",
  "error_description": "Client is not authorized for scopes: profile, email"
}
```

**Missing OpenID Scope:**
```json
{
  "error": "invalid_scope",
  "error_description": "OpenID scope is required for OIDC authentication"
}
```

## Testing Strategy

### Unit Tests

1. **Scope Registry Tests:**
   - Test scope registration
   - Test scope retrieval
   - Test scope expansion (hierarchy)
   - Test claim retrieval

2. **Scope Validator Tests:**
   - Test valid scope validation
   - Test invalid scope rejection
   - Test scope normalization
   - Test client scope restriction
   - Test scope downgrade validation

3. **Claim Filter Tests:**
   - Test claim filtering by scope
   - Test profile scope claims
   - Test email scope claims
   - Test minimal scope (openid only)

### Integration Tests

1. **Authorization Flow:**
   - Request with valid scopes
   - Request with invalid scopes
   - Request with unauthorized scopes (client restriction)
   - Request without scope (use default)

2. **Token Generation:**
   - Access token with correct scope claim
   - ID token with correct user claims
   - Verify claim filtering works

3. **Token Refresh:**
   - Refresh with same scopes
   - Refresh with reduced scopes
   - Refresh with increased scopes (should fail)

### End-to-End Tests

1. Complete OAuth flow with scope validation
2. UserInfo endpoint returns correct claims
3. Client registration with allowed_scopes
4. Scope consent flow

## Implementation Plan

### Phase 1: Core Scope Registry (High Priority)
- Create ScopeDefinition model
- Implement ScopeRegistry
- Register standard OIDC scopes
- Update scope validation to use registry

### Phase 2: Client Scope Restrictions (High Priority)
- Add AllowedScopes to Client model
- Update client registration handler
- Implement client scope validation
- Update authorization handler

### Phase 3: Claim Filtering (High Priority)
- Implement ClaimFilter service
- Update ID token generation
- Update UserInfo endpoint
- Add tests for claim filtering

### Phase 4: Enhanced Features (Medium Priority)
- Add nonce support
- Add PKCE support
- Implement scope downgrade validation
- Add user consent storage

### Phase 5: Advanced Features (Low Priority)
- Implement scope hierarchy
- Add dynamic scope parameters
- Create scope management API
- Add scope analytics

## Security Considerations

1. **Scope Validation:**
   - Always validate scopes against registry
   - Enforce client scope restrictions
   - Require openid scope for OIDC

2. **Claim Protection:**
   - Never include claims not authorized by scope
   - Filter sensitive data based on scope
   - Log scope usage for audit

3. **Client Restrictions:**
   - Store allowed_scopes in database
   - Validate on every authorization request
   - Prevent scope escalation

4. **Token Security:**
   - Include scope in access token
   - Validate scope when using token
   - Support scope downgrade, not upgrade

## Performance Considerations

1. **Scope Registry:**
   - Load once at startup
   - Cache in memory
   - No database queries for validation

2. **Claim Filtering:**
   - Use map lookups for O(1) performance
   - Pre-compute allowed claims per scope
   - Minimize database queries

3. **Validation:**
   - Validate early in request pipeline
   - Cache validation results per request
   - Use efficient string operations

## Migration Strategy

1. **Database Migration:**
   - Add allowed_scopes field to clients (nullable)
   - Add nonce, code_challenge to auth_codes (nullable)
   - Create user_consents collection

2. **Backward Compatibility:**
   - Existing clients without allowed_scopes: allow all scopes
   - Existing auth codes without nonce: skip nonce validation
   - Gradual rollout of new features

3. **Deployment:**
   - Deploy code changes
   - Run database migration
   - Update client configurations
   - Monitor for errors

## Monitoring and Observability

1. **Metrics:**
   - Scope validation failures
   - Unauthorized scope attempts
   - Most requested scopes
   - Claim filtering performance

2. **Logging:**
   - Log scope validation errors
   - Log unauthorized scope attempts
   - Log scope changes in refresh

3. **Alerts:**
   - High rate of scope validation failures
   - Unusual scope request patterns
   - Client attempting unauthorized scopes

## References

- [RFC 6749 - OAuth 2.0](https://datatracker.ietf.org/doc/html/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [RFC 7662 - Token Introspection](https://datatracker.ietf.org/doc/html/rfc7662)
