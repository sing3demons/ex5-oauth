package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	userRepo     *repository.UserRepository
	clientRepo   *repository.ClientRepository
	authCodeRepo *repository.AuthCodeRepository
	config       *config.Config
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	authCodeRepo *repository.AuthCodeRepository,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		config:       cfg,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
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

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	refreshToken, err := utils.GenerateRefreshToken(user.ID, h.config.PrivateKey, h.config.RefreshTokenExpiry)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
		return
	}

	response := models.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.AccessTokenExpiry,
		RefreshToken: refreshToken,
		Scope:        "openid profile email",
	}

	respondJSON(w, http.StatusOK, response)
}
