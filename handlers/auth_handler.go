package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// SSO Cookie constants
const (
	SSOCookieName     = "oauth_sso_session"
	SSOCookieMaxAge   = 86400 * 7 // 7 days
	SSOCookiePath     = "/"
	SSOCookieSecure   = true // HTTPS only in production
	SSOCookieHTTPOnly = true // Prevent XSS
	SSOCookieSameSite = http.SameSiteLaxMode
)

type AuthHandler struct {
	userRepo        *repository.UserRepository
	clientRepo      *repository.ClientRepository
	authCodeRepo    *repository.AuthCodeRepository
	sessionRepo     *repository.SessionRepository
	ssoSessionRepo  *repository.SSOSessionRepository
	config          *config.Config
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	authCodeRepo *repository.AuthCodeRepository,
	sessionRepo *repository.SessionRepository,
	ssoSessionRepo *repository.SSOSessionRepository,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		clientRepo:     clientRepo,
		authCodeRepo:   authCodeRepo,
		sessionRepo:    sessionRepo,
		ssoSessionRepo: ssoSessionRepo,
		config:         cfg,
	}
}

func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID, _ = utils.GenerateRandomString(32)
	}

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"SessionID": sessionID,
	}

	tmpl.Execute(w, data)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Name      string `json:"name"`
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required fields")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to hash password")
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	ctx := context.Background()
	if err := h.userRepo.Create(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(w, http.StatusConflict, "user_exists", "User already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create user")
		return
	}

	// Create SSO Session after successful registration
	ssoSessionID, err := utils.GenerateRandomString(32)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate session ID")
		return
	}

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

	if err := h.ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create SSO session")
		return
	}

	// Set SSO Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SSOCookieName,
		Value:    ssoSessionID,
		Path:     SSOCookiePath,
		MaxAge:   SSOCookieMaxAge,
		HttpOnly: SSOCookieHTTPOnly,
		Secure:   SSOCookieSecure,
		SameSite: SSOCookieSameSite,
	})

	if req.SessionID != "" {
		session, err := h.sessionRepo.FindBySessionID(ctx, req.SessionID)
		if err == nil {
			session.UserID = user.ID
			session.Authenticated = true
			h.sessionRepo.Update(ctx, session)

			code, _ := utils.GenerateRandomString(16)
			code = code + "_" + req.SessionID

			authCode := &models.AuthorizationCode{
				Code:            code,
				ClientID:        session.ClientID,
				UserID:          user.ID,
				RedirectURI:     session.RedirectURI,
				Scope:           session.Scope,
				Nonce:           session.Nonce,
				CodeChallenge:   session.CodeChallenge,
				ChallengeMethod: session.ChallengeMethod,
				ExpiresAt:       time.Now().Add(10 * time.Minute),
			}
			h.authCodeRepo.Create(ctx, authCode)

			// Determine response mode
			responseMode := GetResponseMode(r)
			
			// Prepare response parameters
			params := map[string]string{
				"code": code,
			}
			if session.State != "" {
				params["state"] = session.State
			}

			// Send response based on mode
			SendAuthorizationResponse(w, r, session.RedirectURI, params, responseMode)
			return
		}
	}

	// If no session, return JSON response
	response := map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"email": user.Email,
			"name":  user.Name,
		},
	}
	respondJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	session, err := h.sessionRepo.FindBySessionID(ctx, sessionID)

	data := map[string]interface{}{
		"SessionID": sessionID,
	}

	if err == nil {
		client, err := h.clientRepo.FindByClientID(ctx, session.ClientID)
		if err == nil {
			data["ClientName"] = client.Name
		}
		if session.Scope != "" {
			data["Scope"] = session.Scope
			data["Scopes"] = strings.Split(session.Scope, " ")
		}
	}

		// setHeader to return session ID to client
	w.Header().Set("X-Session-ID", sessionID)


	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	ctx := context.Background()
	user, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Auto-register user if not found
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to hash password")
			return
		}

		newUser := &models.User{
			Email:    req.Email,
			Password: hashedPassword,
			Name:     req.Email, // Use email as name by default
		}

		if err := h.userRepo.Create(ctx, newUser); err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to create user")
			return
		}

		// Fetch the user back to get the generated ID
		user, err = h.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve created user")
			return
		}
	} else {
		// User exists, verify password
		if !utils.CheckPasswordHash(req.Password, user.Password) {
			respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		}
	}

	// Create SSO Session after successful authentication
	ssoSessionID, err := utils.GenerateRandomString(32)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate session ID")
		return
	}

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

	if err := h.ssoSessionRepo.Create(ctx, ssoSession); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create SSO session")
		return
	}

	// Set SSO Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SSOCookieName,
		Value:    ssoSessionID,
		Path:     SSOCookiePath,
		MaxAge:   SSOCookieMaxAge,
		HttpOnly: SSOCookieHTTPOnly,
		Secure:   SSOCookieSecure,
		SameSite: SSOCookieSameSite,
	})

	if req.SessionID != "" {
		session, err := h.sessionRepo.FindBySessionID(ctx, req.SessionID)
		if err == nil && !session.Authenticated {
			session.UserID = user.ID
			session.Authenticated = true
			h.sessionRepo.Update(ctx, session)

			code, _ := utils.GenerateRandomString(16)
			code = code + "_" + req.SessionID

			authCode := &models.AuthorizationCode{
				Code:            code,
				ClientID:        session.ClientID,
				UserID:          user.ID,
				RedirectURI:     session.RedirectURI,
				Scope:           session.Scope,
				Nonce:           session.Nonce,
				CodeChallenge:   session.CodeChallenge,
				ChallengeMethod: session.ChallengeMethod,
				ExpiresAt:       time.Now().Add(10 * time.Minute),
			}
			h.authCodeRepo.Create(ctx, authCode)

			// Determine response mode
			responseMode := GetResponseMode(r)
			
			// Prepare response parameters
			params := map[string]string{
				"code": code,
			}
			if session.State != "" {
				params["state"] = session.State
			}

			// Send response based on mode
			SendAuthorizationResponse(w, r, session.RedirectURI, params, responseMode)
			return
		}
	}

	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Name,
		"openid profile email",
		h.config.PrivateKey,
		h.config.AccessTokenExpiry,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate token")
		return
	}

	scope := "openid profile email"
	refreshToken, err := utils.GenerateRefreshToken(user.ID, scope, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.AccessTokenExpiry,
		RefreshToken: refreshToken,
		Scope:        scope,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract SSO cookie and delete session from database
	cookie, err := r.Cookie(SSOCookieName)
	if err == nil && cookie.Value != "" {
		// Delete session from database
		if err := h.ssoSessionRepo.Delete(ctx, cookie.Value); err != nil {
			// Log error but continue with cookie clearing
			// We don't want to fail logout if session is already gone
		}
	}

	// Clear SSO cookie by setting MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     SSOCookieName,
		Value:    "",
		Path:     SSOCookiePath,
		MaxAge:   -1,
		HttpOnly: SSOCookieHTTPOnly,
		Secure:   SSOCookieSecure,
		SameSite: SSOCookieSameSite,
	})

	// Support post_logout_redirect_uri parameter for OIDC compliance
	redirectURI := r.URL.Query().Get("post_logout_redirect_uri")
	if redirectURI != "" {
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return
	}

	// Return JSON response if no redirect URI provided
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
