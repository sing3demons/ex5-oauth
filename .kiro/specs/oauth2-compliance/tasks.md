# Implementation Plan: Enhanced Scope Management

- [x] 1. Create Scope Registry and Models
  - Create ScopeDefinition model with name, description, claims, hierarchy support
  - Implement ScopeRegistry with methods for registration, validation, and claim retrieval
  - Register standard OIDC scopes (openid, profile, email, phone, address, offline_access)
  - Add global scope registry initialization
  - _Requirements: 1.0, 1.1_

- [x] 2. Enhance Client and AuthCode Models
  - [x] 2.1 Update Client model
    - Add AllowedScopes []string field to Client struct
    - Add GrantTypes []string field to Client struct
    - Update client repository methods to handle new fields
    - _Requirements: 1.2_

  - [x] 2.2 Update AuthorizationCode model
    - Add Nonce string field for ID token replay protection
    - Add CodeChallenge and ChallengeMethod fields for PKCE support
    - Update auth code repository methods
    - _Requirements: 1.5, 2.0_

- [x] 3. Implement Scope Validation Service
  - [x] 3.1 Create scope validator
    - Implement ValidateScope function using registry
    - Implement NormalizeScope to remove duplicates and invalid scopes
    - Implement ValidateScopeName for scope name format validation
    - _Requirements: 1.1_

  - [x] 3.2 Add client scope restriction validation
    - Implement ValidateScopeAgainstAllowed function
    - Check requested scopes against client's AllowedScopes
    - Return detailed error with unauthorized scopes list
    - _Requirements: 1.2_

  - [x] 3.3 Add scope downgrade validation
    - Implement ValidateScopeDowngrade for refresh token flow
    - Ensure requested scopes are subset of original scopes
    - Store original scopes with refresh token
    - _Requirements: 1.4_

- [-] 4. Implement Claim Filtering Service
  - Create ClaimFilter service to filter user claims based on scopes
  - Implement GetClaimsForScopes using scope registry
  - Update ID token generation to use claim filtering
  - Ensure profile scope includes correct claims (name, picture, etc.)
  - Ensure email scope includes email and email_verified
  - _Requirements: 1.3, 1.6_

- [ ] 5. Update Authorization Handler
  - [ ] 5.1 Add scope validation in authorization endpoint
    - Validate requested scopes using scope validator
    - Check scopes against client's AllowedScopes
    - Return invalid_scope error for invalid/unauthorized scopes
    - Use default scope if none provided
    - Require openid scope for OIDC
    - _Requirements: 1.1, 1.2_

  - [ ] 5.2 Store nonce in authorization code
    - Extract nonce parameter from authorization request
    - Store nonce with authorization code
    - Support nonce up to 512 characters
    - _Requirements: 1.5_

- [ ] 6. Update Token Handler
  - [ ] 6.1 Update authorization code grant
    - Validate scopes from authorization code
    - Generate access token with scope claim only (no user claims)
    - Generate ID token with user claims based on scopes using ClaimFilter
    - Include nonce in ID token if present
    - _Requirements: 1.3, 1.6_

  - [ ] 6.2 Update refresh token grant
    - Support scope parameter for scope downgrade
    - Validate requested scopes against original scopes
    - Use original scopes if no scope parameter provided
    - Return invalid_scope error if trying to escalate scopes
    - _Requirements: 1.4_

  - [ ] 6.3 Update client credentials grant
    - Validate requested scopes against client's AllowedScopes
    - Use minimal default scope if none provided
    - Generate access token with scope claim
    - _Requirements: 1.2_

- [ ] 7. Update UserInfo Endpoint
  - Filter returned claims based on access token scope
  - Return only sub claim if only openid scope
  - Include email claims only if email scope present
  - Include profile claims only if profile scope present
  - Support both JWT and JWE access tokens
  - _Requirements: 1.3, 1.6_

- [ ] 8. Update Client Registration Handler
  - [ ] 8.1 Add allowed_scopes support
    - Accept allowed_scopes in client registration request
    - Validate all allowed_scopes exist in registry
    - Store allowed_scopes in database
    - Default to all scopes if not specified
    - _Requirements: 1.2_

  - [ ] 8.2 Add grant_types support
    - Accept grant_types in client registration request
    - Validate grant types are supported
    - Store grant_types in database
    - _Requirements: 1.2_

- [ ] 9. Update Discovery Endpoint
  - Add scopes_supported field with all registered scopes
  - Add grant_types_supported field
  - Add response_types_supported field
  - Include all OIDC required discovery metadata
  - _Requirements: 1.0_

- [ ] 10. Add Integration Tests
  - [ ] 10.1 Test scope validation flow
    - Test authorization with valid scopes
    - Test authorization with invalid scopes returns error
    - Test authorization with unauthorized scopes (client restriction)
    - Test authorization without scope uses default
    - Test openid scope requirement
    - _Requirements: 1.1, 1.2_

  - [ ] 10.2 Test claim filtering
    - Test ID token with openid only returns sub
    - Test ID token with profile includes name
    - Test ID token with email includes email
    - Test UserInfo endpoint filters by scope
    - _Requirements: 1.3, 1.6_

  - [ ] 10.3 Test scope downgrade
    - Test refresh with same scopes succeeds
    - Test refresh with reduced scopes succeeds
    - Test refresh with increased scopes fails
    - _Requirements: 1.4_

- [ ] 11. Add Documentation
  - Update API documentation with scope examples
  - Document client registration with allowed_scopes
  - Document scope validation errors
  - Add scope management guide
  - _Requirements: All_
