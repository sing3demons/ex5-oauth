package handlers

import (
	"context"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/middleware"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type OAuthHandler struct {
	userRepo     *repository.UserRepository
	clientRepo   *repository.ClientRepository
	authCodeRepo *repository.AuthCodeRepository
	sessionRepo  *repository.SessionRepository
	consentRepo  *repository.UserConsentRepository
	config       *config.Config
}

func NewOAuthHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	authCodeRepo *repository.AuthCodeRepository,
	sessionRepo *repository.SessionRepository,
	consentRepo *repository.UserConsentRepository,
	cfg *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		sessionRepo:  sessionRepo,
		consentRepo:  consentRepo,
		config:       cfg,
	}
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	nonce := r.URL.Query().Get("nonce")
	codeChallenge := r.URL.Query().Get("code_challenge")
	challengeMethod := r.URL.Query().Get("code_challenge_method")
	prompt := r.URL.Query().Get("prompt")

	// Validate nonce length (max 512 characters as per OIDC spec)
	if len(nonce) > 512 {
		respondError(w, http.StatusBadRequest, "invalid_request", "Nonce exceeds maximum length of 512 characters")
		return
	}

	if responseType != "code" {
		respondError(w, http.StatusBadRequest, "unsupported_response_type", "Only 'code' response type is supported")
		return
	}

	if clientID == "" || redirectURI == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusBadRequest, "invalid_client", "Client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to validate client")
		return
	}

	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirect = true
			break
		}
	}

	if !validRedirect {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid redirect URI")
		return
	}

	// Validate and normalize scope
	if scope == "" {
		scope = utils.GetDefaultScope()
	} else {
		// Validate scope format and existence
		if err := utils.GlobalScopeValidator.ValidateScope(scope); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
		scope = utils.NormalizeScope(scope)
	}

	// Validate scopes against client's AllowedScopes
	if err := utils.GlobalScopeValidator.ValidateScopeAgainstAllowed(scope, client.AllowedScopes); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
		return
	}

	// OIDC requires openid scope
	if !utils.RequiresOpenID(scope) {
		respondError(w, http.StatusBadRequest, "invalid_scope", "OpenID scope is required")
		return
	}

	// Validate PKCE parameters if present
	if codeChallenge != "" {
		if challengeMethod == "" {
			challengeMethod = "plain" // Default to plain if not specified
		}
		if challengeMethod != "S256" && challengeMethod != "plain" {
			respondError(w, http.StatusBadRequest, "invalid_request", "Invalid code_challenge_method")
			return
		}
	}

	// Extract SSO session from request context
	ssoSession, _ := r.Context().Value(middleware.SSOSessionContextKey).(*models.SSOSession)

	// Handle prompt=login: force re-authentication
	if prompt == "login" {
		ssoSession = nil // Ignore SSO session to force login
	}

	// Handle prompt=select_account: display account selection (placeholder for future)
	// For now, treat it like prompt=login and force re-authentication
	if prompt == "select_account" {
		ssoSession = nil // Force account selection by requiring login
	}

	// Handle prompt=none: fail immediately if not authenticated or no consent
	if prompt == "none" {
		// Check if user is authenticated
		if ssoSession == nil || !ssoSession.Authenticated {
			// Build error redirect URL
			errorURL := redirectURI + "?error=login_required&error_description=User+authentication+required"
			if state != "" {
				errorURL += "&state=" + state
			}
			http.Redirect(w, r, errorURL, http.StatusFound)
			return
		}

		// User is authenticated, check for consent
		requestedScopes := strings.Split(scope, " ")
		hasConsent, err := h.consentRepo.HasConsent(ctx, ssoSession.UserID, clientID, requestedScopes)
		if err != nil || !hasConsent {
			// Build error redirect URL
			errorURL := redirectURI + "?error=consent_required&error_description=User+consent+required"
			if state != "" {
				errorURL += "&state=" + state
			}
			http.Redirect(w, r, errorURL, http.StatusFound)
			return
		}

		// Both authentication and consent exist, proceed with auto-approval
		// Fall through to the auto-approval logic below
	}

	// Check if SSO session exists and is authenticated
	if ssoSession != nil && ssoSession.Authenticated {
		// Parse scopes for consent check
		requestedScopes := strings.Split(scope, " ")
		
		// Check for existing user consent
		hasConsent, err := h.consentRepo.HasConsent(ctx, ssoSession.UserID, clientID, requestedScopes)
		if err != nil {
			// Log error but continue to consent screen
			hasConsent = false
		}

		// Handle prompt=consent: force consent screen
		if prompt == "consent" {
			hasConsent = false
		}

		// Auto-approve if consent exists
		if hasConsent {
			// Generate authorization code immediately
			code, err := utils.GenerateRandomString(16)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate authorization code")
				return
			}

			authCode := &models.AuthorizationCode{
				Code:            code,
				ClientID:        clientID,
				UserID:          ssoSession.UserID,
				RedirectURI:     redirectURI,
				Scope:           scope,
				Nonce:           nonce,
				CodeChallenge:   codeChallenge,
				ChallengeMethod: challengeMethod,
				ExpiresAt:       time.Now().Add(10 * time.Minute),
			}

			if err := h.authCodeRepo.Create(ctx, authCode); err != nil {
				respondError(w, http.StatusInternalServerError, "server_error", "Failed to create authorization code")
				return
			}

			// Redirect with authorization code
			redirectURL := redirectURI + "?code=" + code
			if state != "" {
				redirectURL += "&state=" + state
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		// No consent - redirect to consent screen
		consentURL := "/oauth/consent?client_id=" + clientID +
			"&scope=" + scope +
			"&redirect_uri=" + redirectURI +
			"&response_type=" + responseType
		if state != "" {
			consentURL += "&state=" + state
		}
		if nonce != "" {
			consentURL += "&nonce=" + nonce
		}
		if codeChallenge != "" {
			consentURL += "&code_challenge=" + codeChallenge
			if challengeMethod != "" {
				consentURL += "&code_challenge_method=" + challengeMethod
			}
		}
		http.Redirect(w, r, consentURL, http.StatusFound)
		return
	}

	// No SSO session or no consent - create OAuth session and redirect to login
	sessionID, err := utils.GenerateRandomString(32)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate session")
		return
	}

	session := &models.Session{
		SessionID:       sessionID,
		ClientID:        clientID,
		RedirectURI:     redirectURI,
		Scope:           scope,
		State:           state,
		ResponseType:    responseType,
		Nonce:           nonce,
		CodeChallenge:   codeChallenge,
		ChallengeMethod: challengeMethod,
		Authenticated:   false,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}

	if err := h.sessionRepo.Create(ctx, session); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create session")
		return
	}

	loginURL := "/auth/login?session_id=" + sessionID
	http.Redirect(w, r, loginURL, http.StatusFound)
}

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
	case "urn:ietf:params:oauth:grant-type:token-exchange":
		// Delegate to token exchange handler
		h.handleTokenExchange(w, r)
	default:
		respondError(w, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
	}
}

func (h *OAuthHandler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || clientID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client")
		return
	}

	// For confidential clients, verify client_secret
	// For public clients (PKCE), verify code_verifier instead
	if client.ClientSecret != "" {
		// Confidential client - require client_secret
		if clientSecret == "" || client.ClientSecret != clientSecret {
			respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
			return
		}
	}

	authCode, err := h.authCodeRepo.FindByCode(ctx, code)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid authorization code")
		return
	}

	if authCode.ExpiresAt.Before(time.Now()) {
		h.authCodeRepo.Delete(ctx, code)
		respondError(w, http.StatusBadRequest, "invalid_grant", "Authorization code expired")
		return
	}

	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		respondError(w, http.StatusBadRequest, "invalid_grant", "Code mismatch")
		return
	}

	// Verify PKCE if code_challenge was used
	if authCode.CodeChallenge != "" {
		if codeVerifier == "" {
			respondError(w, http.StatusBadRequest, "invalid_request", "code_verifier required for PKCE")
			return
		}
		if !utils.VerifyPKCE(codeVerifier, authCode.CodeChallenge, authCode.ChallengeMethod) {
			respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid code_verifier")
			return
		}
	}

	h.authCodeRepo.Delete(ctx, code)

	user, err := h.userRepo.FindByID(ctx, authCode.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
	}

	// Validate scopes from authorization code (already validated during authorization)
	// Scopes are stored in authCode.Scope
	
	// Generate access token with scope claim only (no user claims)
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Name,
		authCode.Scope,
		h.config.PrivateKey,
		h.config.AccessTokenExpiry,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, authCode.Scope, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	// Generate ID token with user claims based on scopes using ClaimFilter
	// Include nonce in ID token if present (for replay protection)
	userClaims := utils.GetIDTokenClaimsForUser(user, authCode.Scope, authCode.Nonce)
	idToken, err := utils.GenerateIDToken(
		user.ID,
		clientID,
		userClaims,
		h.config.PrivateKey,
		h.config.AccessTokenExpiry,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.AccessTokenExpiry,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        authCode.Scope,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *OAuthHandler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	requestedScope := r.FormValue("scope")

	if refreshToken == "" || clientID == "" || clientSecret == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil || client.ClientSecret != clientSecret {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
	}

	claims, err := utils.ValidateRefreshToken(refreshToken, h.config.PublicKey)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid refresh token")
		return
	}

	// Use UserID field which contains the actual user ID (Subject may be empty due to JSON tag conflict)
	userID := claims.UserID
	if userID == "" {
		userID = claims.Subject
	}
	
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
	}

	// Support scope parameter for scope downgrade
	scope := requestedScope
	if scope == "" {
		// Use original scopes if no scope parameter provided
		scope = claims.Scope
	} else {
		// Validate requested scope format and existence
		if err := utils.GlobalScopeValidator.ValidateScope(scope); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
		scope = utils.NormalizeScope(scope)
		
		// Validate requested scopes against original scopes (scope downgrade validation)
		// Return invalid_scope error if trying to escalate scopes
		if err := utils.ValidateScopeDowngrade(scope, claims.Scope); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Name,
		scope,
		h.config.PrivateKey,
		h.config.AccessTokenExpiry,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, scope, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.AccessTokenExpiry,
		RefreshToken: newRefreshToken,
		Scope:        scope,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *OAuthHandler) handleClientCredentialsGrant(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	requestedScope := r.FormValue("scope")

	if clientID == "" || clientSecret == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil || client.ClientSecret != clientSecret {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
	}

	// Use minimal default scope if none provided
	scope := requestedScope
	if scope == "" {
		scope = "openid"
	} else {
		// Validate requested scope format and existence
		if err := utils.GlobalScopeValidator.ValidateScope(scope); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
			return
		}
		scope = utils.NormalizeScope(scope)
	}
	
	// Validate requested scopes against client's AllowedScopes
	if err := utils.GlobalScopeValidator.ValidateScopeAgainstAllowed(scope, client.AllowedScopes); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_scope", err.Error())
		return
	}

	// Generate access token with scope claim
	accessToken, err := utils.GenerateAccessToken(
		clientID,
		"",
		client.Name,
		scope,
		h.config.PrivateKey,
		h.config.AccessTokenExpiry,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
		return
	}

	response := models.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   h.config.AccessTokenExpiry,
		Scope:       scope,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *OAuthHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authorization required")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	
	var scope string
	var userID string

	// Support both JWT and JWE tokens
	if utils.IsJWE(tokenString) {
		jweClaims, err := utils.ValidateJWE(tokenString, h.config.PrivateKey)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}
		userID = jweClaims.UserID
		scope = jweClaims.Scope
	} else {
		jwtClaims, err := utils.ValidateToken(tokenString, h.config.PublicKey)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}
		userID = jwtClaims.UserID
		scope = jwtClaims.Scope
	}

	// Get user from database
	ctx := context.Background()
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
	}

	// Filter claims based on scope using claim filtering service
	filteredClaims := utils.FilterClaimsForUser(user, scope)
	
	respondJSON(w, http.StatusOK, filteredClaims)
}

func (h *OAuthHandler) handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	// Create a temporary TokenExchangeHandler to handle the request
	tokenExchangeHandler := NewTokenExchangeHandler(h.userRepo, h.clientRepo, h.config)
	tokenExchangeHandler.HandleTokenExchange(w, r)
}
