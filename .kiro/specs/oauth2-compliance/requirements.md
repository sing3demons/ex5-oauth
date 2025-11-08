# Requirements Document: OAuth2/OIDC Compliance - Enhanced Scope Management

## Introduction

ปรับปรุง OAuth2 Server ให้มีระบบจัดการ scope ที่ดีและเป็นไปตามมาตรฐาน OAuth 2.0 (RFC 6749), OpenID Connect Core 1.0 โดยเน้นการจัดการ scope อย่างละเอียด รองรับ scope hierarchy, client-specific restrictions, และ dynamic scope parameters

## Glossary

- **OAuth2 Server**: ระบบ Authorization Server ที่ออก access tokens
- **OIDC (OpenID Connect)**: Identity layer บน OAuth 2.0
- **Scope**: สิทธิ์การเข้าถึงข้อมูลที่ client ขอ
- **Scope Registry**: ระบบจัดเก็บและจัดการ scope definitions
- **Client**: แอปพลิเคชันที่ใช้ OAuth2 authentication
- **Resource Owner**: ผู้ใช้ที่เป็นเจ้าของข้อมูล
- **Scope Hierarchy**: โครงสร้างแบบ parent-child ของ scopes
- **Parameterized Scope**: Scope ที่มี parameters เช่น "read:user:123"
- **Consent**: การอนุญาตจาก user ให้ client เข้าถึงข้อมูล

## Requirements

### Requirement 1: Scope Definition and Registry

**User Story:** As a system administrator, I want to define and manage available scopes in the system, so that I can control what permissions are available to clients.

#### Acceptance Criteria

1. THE OAuth2 Server SHALL maintain a scope registry with scope definitions including name, description, and required claims
2. THE OAuth2 Server SHALL support standard OIDC scopes: openid, profile, email, phone, address, offline_access
3. THE OAuth2 Server SHALL allow defining custom scopes with associated permissions
4. WHEN defining a scope, THE OAuth2 Server SHALL validate that the scope name contains only allowed characters (alphanumeric, underscore, hyphen, colon, period)
5. THE OAuth2 Server SHALL expose available scopes through the discovery endpoint in scopes_supported field

### Requirement 1.1: Scope Validation and Normalization

**User Story:** As a client developer, I want the OAuth2 server to validate and normalize requested scopes, so that my authorization requests are handled consistently.

#### Acceptance Criteria

1. WHEN a client requests authorization with invalid scopes, THE OAuth2 Server SHALL return an "invalid_scope" error with HTTP 400 status
2. WHEN a client requests authorization with duplicate scopes, THE OAuth2 Server SHALL normalize the scope by removing duplicates while preserving order
3. WHEN a client requests authorization with extra whitespace in scopes, THE OAuth2 Server SHALL trim and normalize the scope string
4. WHERE OIDC is used, THE OAuth2 Server SHALL require "openid" scope in all authorization requests
5. WHEN a client requests authorization without specifying scope, THE OAuth2 Server SHALL use a configurable default scope value

### Requirement 1.2: Client-Specific Scope Restrictions

**User Story:** As a system administrator, I want to restrict which scopes each client can request, so that clients cannot access data beyond their authorization.

#### Acceptance Criteria

1. WHEN registering a client, THE OAuth2 Server SHALL allow specifying allowed_scopes for that client
2. WHEN a client requests authorization, THE OAuth2 Server SHALL validate that requested scopes are a subset of client's allowed_scopes
3. IF a client requests unauthorized scopes, THEN THE OAuth2 Server SHALL return "invalid_scope" error with details of which scopes are not allowed
4. WHERE no allowed_scopes are specified for a client, THE OAuth2 Server SHALL allow all system-defined scopes
5. THE OAuth2 Server SHALL apply scope restrictions to all grant types (authorization_code, refresh_token, client_credentials)

### Requirement 1.3: Scope-Based Claim Filtering

**User Story:** As a resource server, I want tokens to include only claims authorized by the granted scopes, so that user privacy is protected.

#### Acceptance Criteria

1. WHEN generating ID tokens with "profile" scope, THE OAuth2 Server SHALL include name, family_name, given_name, middle_name, nickname, preferred_username, profile, picture, website, gender, birthdate, zoneinfo, locale, updated_at claims
2. WHEN generating ID tokens with "email" scope, THE OAuth2 Server SHALL include email and email_verified claims
3. WHEN generating ID tokens with "phone" scope, THE OAuth2 Server SHALL include phone_number and phone_number_verified claims
4. WHEN generating ID tokens with "address" scope, THE OAuth2 Server SHALL include address claim
5. WHEN generating ID tokens without specific scope, THE OAuth2 Server SHALL include only sub claim and standard JWT claims (iss, aud, exp, iat)

### Requirement 1.4: Scope Downgrade in Token Refresh

**User Story:** As a client developer, I want to request reduced scopes when refreshing tokens, so that I can implement principle of least privilege.

#### Acceptance Criteria

1. WHEN refreshing a token with a scope parameter, THE OAuth2 Server SHALL validate that requested scopes are a subset of original scopes
2. IF requested scopes exceed original scopes, THEN THE OAuth2 Server SHALL return "invalid_scope" error
3. WHEN refreshing a token without scope parameter, THE OAuth2 Server SHALL issue tokens with the original scopes
4. THE OAuth2 Server SHALL allow requesting fewer scopes than originally granted
5. THE OAuth2 Server SHALL store original granted scopes with refresh token for validation

### Requirement 1.5: Scope Consent and User Authorization

**User Story:** As a user, I want to see and approve the specific permissions an application is requesting, so that I can make informed decisions.

#### Acceptance Criteria

1. WHEN displaying consent screen, THE OAuth2 Server SHALL show human-readable descriptions for each requested scope
2. THE OAuth2 Server SHALL allow users to selectively approve or deny individual scopes
3. WHEN a user denies required scopes, THE OAuth2 Server SHALL return "access_denied" error
4. THE OAuth2 Server SHALL store user consent decisions per client and scope combination
5. WHEN a client requests previously consented scopes, THE OAuth2 Server SHALL skip consent screen unless prompt=consent is specified

### Requirement 1.6: Scope in Access Tokens vs ID Tokens

**User Story:** As a security engineer, I want clear separation between access token scopes and ID token claims, so that tokens serve their intended purposes.

#### Acceptance Criteria

1. WHEN generating access tokens, THE OAuth2 Server SHALL include scope claim with space-separated scope values
2. WHEN generating access tokens, THE OAuth2 Server SHALL NOT include user profile claims (name, email, etc.)
3. WHEN generating ID tokens, THE OAuth2 Server SHALL include user claims based on granted scopes
4. WHEN generating ID tokens, THE OAuth2 Server SHALL include scope claim for reference
5. THE OAuth2 Server SHALL use access tokens for API authorization and ID tokens for user authentication

### Requirement 1.7: Scope Hierarchy and Dependencies

**User Story:** As a system administrator, I want to define scope hierarchies and dependencies, so that granting a parent scope automatically grants child scopes.

#### Acceptance Criteria

1. THE OAuth2 Server SHALL support defining parent-child relationships between scopes
2. WHEN a client requests a parent scope, THE OAuth2 Server SHALL automatically grant all child scopes
3. THE OAuth2 Server SHALL validate that scope hierarchies do not contain circular dependencies
4. WHEN displaying consent screen, THE OAuth2 Server SHALL show parent scopes with their included child scopes
5. THE OAuth2 Server SHALL expand parent scopes to include all child scopes in token scope claim

### Requirement 1.8: Dynamic Scope Parameters

**User Story:** As a client developer, I want to request scopes with parameters, so that I can request fine-grained permissions.

#### Acceptance Criteria

1. THE OAuth2 Server SHALL support parameterized scopes in format "scope:parameter" (e.g., "read:user:123")
2. WHEN validating parameterized scopes, THE OAuth2 Server SHALL validate both the scope name and parameter format
3. THE OAuth2 Server SHALL allow defining parameter validation rules for each scope
4. WHEN storing granted scopes, THE OAuth2 Server SHALL preserve scope parameters
5. THE OAuth2 Server SHALL include parameterized scopes in token scope claim

### Requirement 2: PKCE Support (RFC 7636)

**User Story:** As a mobile app developer, I want to use PKCE for authorization code flow, so that my app is protected against authorization code interception attacks.

#### Acceptance Criteria

1. WHEN a client sends an authorization request with code_challenge parameter, THE OAuth2 Server SHALL store the code_challenge with the authorization code
2. WHEN a client exchanges an authorization code with PKCE, THE OAuth2 Server SHALL require the code_verifier parameter
3. WHEN validating PKCE, THE OAuth2 Server SHALL verify that SHA256(code_verifier) equals the stored code_challenge
4. WHERE PKCE is used with code_challenge_method "plain", THE OAuth2 Server SHALL verify that code_verifier equals code_challenge
5. IF PKCE verification fails, THEN THE OAuth2 Server SHALL return "invalid_grant" error and invalidate the authorization code

### Requirement 3: Token Introspection Endpoint (RFC 7662)

**User Story:** As a resource server, I want to introspect access tokens to validate them and get token metadata, so that I can make authorization decisions.

#### Acceptance Criteria

1. THE OAuth2 Server SHALL provide a token introspection endpoint at /oauth/introspect
2. WHEN a valid token is introspected, THE OAuth2 Server SHALL return active=true with token metadata
3. WHEN an invalid or expired token is introspected, THE OAuth2 Server SHALL return active=false
4. THE OAuth2 Server SHALL require client authentication for introspection requests
5. THE OAuth2 Server SHALL return scope, client_id, username, token_type, exp, iat, sub in introspection response

### Requirement 4: Token Revocation Endpoint (RFC 7009)

**User Story:** As a user, I want to revoke my access tokens and refresh tokens, so that I can immediately terminate access for applications I no longer trust.

#### Acceptance Criteria

1. THE OAuth2 Server SHALL provide a token revocation endpoint at /oauth/revoke
2. WHEN a valid token is revoked, THE OAuth2 Server SHALL invalidate the token immediately
3. WHEN revoking a refresh token, THE OAuth2 Server SHALL also invalidate all associated access tokens
4. THE OAuth2 Server SHALL require client authentication for revocation requests
5. THE OAuth2 Server SHALL return HTTP 200 for both successful revocations and already-revoked tokens

## Priority

**High Priority (Focus on Scope Management):**
- Requirement 1: Scope Definition and Registry
- Requirement 1.1: Scope Validation and Normalization
- Requirement 1.2: Client-Specific Scope Restrictions
- Requirement 1.3: Scope-Based Claim Filtering
- Requirement 1.6: Scope in Access Tokens vs ID Tokens

**Medium Priority:**
- Requirement 1.4: Scope Downgrade in Token Refresh
- Requirement 1.5: Scope Consent and User Authorization
- Requirement 2: PKCE Support
- Requirement 3: Token Introspection
- Requirement 4: Token Revocation

**Low Priority (Advanced Features):**
- Requirement 1.7: Scope Hierarchy and Dependencies
- Requirement 1.8: Dynamic Scope Parameters

## Out of Scope

- Multi-factor authentication (MFA)
- Social login integration
- User management UI
- Admin dashboard
- Metrics and monitoring
- Rate limiting (will be separate spec)
- Consent screen UI (will be separate spec)
