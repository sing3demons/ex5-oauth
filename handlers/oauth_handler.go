package handlers

import (
	"context"
	"net/http"
	"net/url"
	"oauth2-server/config"
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
	config       *config.Config
}

func NewOAuthHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	authCodeRepo *repository.AuthCodeRepository,
	sessionRepo *repository.SessionRepository,
	cfg *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		sessionRepo:  sessionRepo,
		config:       cfg,
	}
}

func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")

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

	sessionID, err := utils.GenerateRandomString(32)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate session")
		return
	}

	session := &models.Session{
		SessionID:     sessionID,
		ClientID:      clientID,
		RedirectURI:   redirectURI,
		Scope:         scope,
		State:         state,
		ResponseType:  responseType,
		Authenticated: false,
		ExpiresAt:     time.Now().Add(10 * time.Minute),
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
	default:
		respondError(w, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
	}
}

func (h *OAuthHandler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	if code == "" || clientID == "" || clientSecret == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil || client.ClientSecret != clientSecret {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
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

	h.authCodeRepo.Delete(ctx, code)

	user, err := h.userRepo.FindByID(ctx, authCode.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
	}

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

	refreshToken, err := utils.GenerateRefreshToken(user.ID, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	idToken, err := utils.GenerateIDToken(
		user.ID,
		user.Email,
		user.Name,
		clientID,
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

	claims, err := utils.ValidateToken(refreshToken, h.config.PublicKey)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid refresh token")
		return
	}

	user, err := h.userRepo.FindByID(ctx, claims.Subject)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
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
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.AccessTokenExpiry,
		RefreshToken: newRefreshToken,
		Scope:        "openid profile email",
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *OAuthHandler) handleClientCredentialsGrant(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	scope := r.FormValue("scope")

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
	claims, err := utils.ValidateToken(tokenString, h.config.PublicKey)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
		return
	}

	userInfo := models.UserInfo{
		Sub:   claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
	}

	respondJSON(w, http.StatusOK, userInfo)
}
