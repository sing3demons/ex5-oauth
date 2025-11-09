# Implementation Plan

- [x] 1. Create SSO Session model and repository
  - Add `SSOSession` struct to `models/models.go` with fields: ID, SessionID, UserID, Authenticated, CreatedAt, ExpiresAt, LastActivity, IPAddress, UserAgent
  - Create `repository/sso_session_repository.go` with methods: Create, FindBySessionID, UpdateLastActivity, Delete, DeleteExpired, FindByUserID
  - Create MongoDB indexes for session_id (unique), user_id, and expires_at
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 12.1, 12.2_

- [x] 2. Create User Consent model and repository
  - Add `UserConsent` struct to `models/models.go` with fields: ID, UserID, ClientID, Scopes, GrantedAt, ExpiresAt
  - Create `repository/user_consent_repository.go` with methods: Create, FindByUserAndClient, HasConsent, RevokeConsent, ListUserConsents
  - Create MongoDB unique compound index on user_id + client_id
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 9.1, 9.2_

- [x] 3. Implement SSO middleware
  - Create `middleware/sso_middleware.go` with SSOMiddleware function
  - Extract SSO cookie from request
  - Validate session against database (check authenticated and not expired)
  - Update last activity timestamp on valid session
  - Add session to request context for downstream handlers
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 4. Update login handler to create SSO sessions
  - Add SSO cookie constants to `handlers/auth_handler.go` (SSOCookieName, SSOCookieMaxAge, etc.)
  - Inject SSOSessionRepository into AuthHandler
  - After successful authentication, generate 32-byte random session ID
  - Create SSOSession with user ID, timestamps, IP address, and user agent
  - Set HTTP-only secure cookie with session ID
  - _Requirements: 1.1, 1.2, 1.3, 1.5, 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 5. Update authorization handler for SSO check and auto-approval
  - Inject UserConsentRepository into OAuthHandler
  - At start of Authorize method, extract SSO session from request context
  - If SSO session exists and authenticated, check for existing user consent using HasConsent
  - If consent exists and prompt != "consent", generate authorization code immediately and redirect (auto-approval)
  - If consent missing or prompt == "consent", redirect to consent screen
  - If no SSO session or prompt == "login", redirect to login page
  - _Requirements: 2.5, 4.1, 4.2, 4.3, 4.4, 4.5, 11.2, 11.3_

- [x] 6. Create consent screen template
  - Create `templates/consent.html` with HTML form
  - Display client name and requested scopes with descriptions
  - Include hidden fields for client_id, scope, state, redirect_uri
  - Provide "Allow" and "Deny" buttons
  - Style with CSS for professional appearance
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 7. Implement consent handlers
  - Create `handlers/consent_handler.go` with ConsentHandler struct
  - Implement ShowConsent method to render consent template with client info and scope descriptions
  - Implement HandleConsent method to process form submission
  - On approval: save UserConsent record with 1-year expiration, generate authorization code, redirect with code
  - On denial: redirect with access_denied error
  - _Requirements: 3.1, 3.2, 5.5, 9.3, 9.4_

- [x] 8. Implement logout handler
  - Add Logout method to AuthHandler
  - Extract SSO cookie and delete session from database
  - Clear SSO cookie by setting MaxAge to -1
  - Support post_logout_redirect_uri parameter for OIDC compliance
  - Return JSON response if no redirect URI provided
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 9. Add session management endpoints
  - Create `handlers/session_handler.go` with SessionHandler struct
  - Implement ListSessions endpoint (GET /account/sessions) to return all active sessions for authenticated user
  - Implement RevokeSession endpoint (DELETE /account/sessions/{session_id}) to delete specific session
  - Extract user ID from access token for authentication
  - Return appropriate JSON responses with session data
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 10. Add authorization management endpoints
  - Add ListAuthorizations method to SessionHandler (GET /account/authorizations)
  - Add RevokeAuthorization method to SessionHandler (DELETE /account/authorizations/{client_id})
  - Fetch client information to include client name in response
  - Return list of consents with client details and scopes
  - Delete consent record on revocation
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [x] 11. Add OIDC prompt parameter support
  - Update Authorize handler to parse prompt parameter from query string
  - Implement prompt=none: fail with login_required or consent_required if not authenticated/consented
  - Implement prompt=login: force re-authentication by skipping SSO session check
  - Implement prompt=consent: force consent screen even if consent exists
  - Implement prompt=select_account: display account selection (placeholder for future)
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

- [x] 12. Wire up SSO middleware and routes in main.go
  - Import SSO and consent repositories
  - Initialize SSOSessionRepository and UserConsentRepository
  - Apply SSOMiddleware to authorization routes
  - Register consent handler routes (/oauth/consent GET and POST)
  - Register logout route (/auth/logout)
  - Register session management routes (/account/sessions, /account/authorizations)
  - _Requirements: 7.1, 7.2_

- [x] 13. Add unit tests for repositories
  - Write tests for SSOSessionRepository: Create, FindBySessionID, UpdateLastActivity, Delete, DeleteExpired, FindByUserID
  - Write tests for UserConsentRepository: Create, FindByUserAndClient, HasConsent (scope matching), RevokeConsent, ListUserConsents
  - Test edge cases: expired sessions, missing records, scope validation
  - _Requirements: All requirements (validation)_

- [x] 14. Add integration tests for SSO flows
  - Test first login flow: no SSO → login → consent → code
  - Test second app flow: SSO exists → consent exists → auto-approve → code
  - Test logout flow: logout → SSO cleared → requires login
  - Test expired session flow: expired SSO → requires login
  - Test consent revocation flow: revoke → requires consent again
  - Test prompt parameter flows: prompt=login, prompt=consent, prompt=none
  - _Requirements: All requirements (end-to-end validation)_

- [x] 15. Add documentation
  - Update main README.md with SSO feature description
  - Create SSO_USAGE.md with examples of SSO flows
  - Document session management API endpoints
  - Document consent management API endpoints
  - Add configuration examples for SSO settings
  - _Requirements: All requirements (documentation)_
