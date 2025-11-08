package handlers

import (
	"context"
	"encoding/json"
	"html/template"
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

type AuthHandler struct {
	userRepo     *repository.UserRepository
	clientRepo   *repository.ClientRepository
	authCodeRepo *repository.AuthCodeRepository
	sessionRepo  *repository.SessionRepository
	config       *config.Config
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	authCodeRepo *repository.AuthCodeRepository,
	sessionRepo *repository.SessionRepository,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		sessionRepo:  sessionRepo,
		config:       cfg,
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

	response := map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"email": user.Email,
			"name":  user.Name,
		},
	}

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

			redirectURL, _ := url.Parse(session.RedirectURI)
			q := redirectURL.Query()
			q.Set("code", code)
			if session.State != "" {
				q.Set("state", session.State)
			}
			redirectURL.RawQuery = q.Encode()

			response["redirect_uri"] = redirectURL.String()
		}
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
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}

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

			redirectURL, _ := url.Parse(session.RedirectURI)
			q := redirectURL.Query()
			q.Set("code", code)
			if session.State != "" {
				q.Set("state", session.State)
			}
			redirectURL.RawQuery = q.Encode()

			respondJSON(w, http.StatusOK, map[string]string{
				"redirect_uri": redirectURL.String(),
				"code":         code,
			})
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
