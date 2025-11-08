package handlers

import (
	"context"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"time"
)

const (
	TokenExchangeGrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
	AccessTokenType        = "urn:ietf:params:oauth:token-type:access_token"
	RefreshTokenType       = "urn:ietf:params:oauth:token-type:refresh_token"
	IDTokenType            = "urn:ietf:params:oauth:token-type:id_token"
)

type TokenExchangeHandler struct {
	userRepo   *repository.UserRepository
	clientRepo *repository.ClientRepository
	config     *config.Config
}

func NewTokenExchangeHandler(
	userRepo *repository.UserRepository,
	clientRepo *repository.ClientRepository,
	cfg *config.Config,
) *TokenExchangeHandler {
	return &TokenExchangeHandler{
		userRepo:   userRepo,
		clientRepo: clientRepo,
		config:     cfg,
	}
}

type TokenExchangeRequest struct {
	GrantType          string `json:"grant_type"`
	SubjectToken       string `json:"subject_token"`
	SubjectTokenType   string `json:"subject_token_type"`
	RequestedTokenType string `json:"requested_token_type,omitempty"`
	Scope              string `json:"scope,omitempty"`
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	IsEncryptedJWE     bool   `json:"is_encrypted_jwe,omitempty"`
}

type TokenExchangeResponse struct {
	AccessToken     string `json:"access_token"`
	IssuedTokenType string `json:"issued_token_type"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int64  `json:"expires_in"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	IDToken         string `json:"id_token,omitempty"`
	Scope           string `json:"scope,omitempty"`
}

func (h *TokenExchangeHandler) HandleTokenExchange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	req := TokenExchangeRequest{
		GrantType:          r.FormValue("grant_type"),
		SubjectToken:       r.FormValue("subject_token"),
		SubjectTokenType:   r.FormValue("subject_token_type"),
		RequestedTokenType: r.FormValue("requested_token_type"),
		Scope:              r.FormValue("scope"),
		ClientID:           r.FormValue("client_id"),
		ClientSecret:       r.FormValue("client_secret"),
		IsEncryptedJWE:     r.FormValue("is_encrypted_jwe") == "true",
	}

	if req.GrantType != TokenExchangeGrantType {
		respondError(w, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
		return
	}

	if req.SubjectToken == "" || req.SubjectTokenType == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()
	client, err := h.clientRepo.FindByClientID(ctx, req.ClientID)
	if err != nil || client.ClientSecret != req.ClientSecret {
		respondError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
	}

	var userID, email, name, scope string
	if utils.IsJWE(req.SubjectToken) {
		claims, err := utils.ValidateJWE(req.SubjectToken, h.config.PrivateKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid subject token")
			return
		}
		userID = claims.UserID
		email = claims.Email
		name = claims.Name
		scope = claims.Scope
	} else if utils.IsJWT(req.SubjectToken) {
		claims, err := utils.ValidateToken(req.SubjectToken, h.config.PublicKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid subject token")
			return
		}
		userID = claims.UserID
		email = claims.Email
		name = claims.Name
		scope = claims.Scope
	} else {
		respondError(w, http.StatusBadRequest, "invalid_grant", "Invalid token format")
		return
	}

	// Get user from database to ensure user exists and get latest info
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to find user")
		return
	}

	if email == "" {
		email = user.Email
	}
	if name == "" {
		name = user.Name
	}

	// Validate and normalize scope
	if req.Scope != "" {
		if !utils.ValidateScope(req.Scope) {
			respondError(w, http.StatusBadRequest, "invalid_scope", "Invalid scope requested")
			return
		}
		scope = utils.NormalizeScope(req.Scope)
	} else if scope == "" {
		scope = utils.GetDefaultScope()
	}

	var accessToken, refreshToken, idToken string
	expiresIn := h.config.AccessTokenExpiry

	if req.IsEncryptedJWE {
		accessToken, err = utils.GenerateJWEAccessToken(
			userID,
			email,
			name,
			scope,
			h.config.PublicKey,
			time.Now().Add(time.Duration(expiresIn)*time.Second).Unix(),
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
			return
		}

		refreshToken, err = utils.GenerateJWERefreshToken(
			userID,
			h.config.PublicKey,
			time.Now().Add(time.Duration(h.config.RefreshTokenExpiry)*time.Second).Unix(),
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
			return
		}

		// Generate ID token with filtered claims based on scopes
		userClaims := utils.GetIDTokenClaimsForUser(user, scope, "")
		idToken, err = utils.GenerateJWEIDToken(
			userID,
			req.ClientID,
			userClaims,
			h.config.PublicKey,
			time.Now().Add(time.Duration(expiresIn)*time.Second).Unix(),
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
			return
		}
	} else {
		accessToken, err = utils.GenerateAccessToken(
			userID,
			email,
			name,
			scope,
			h.config.PrivateKey,
			expiresIn,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
			return
		}

		refreshToken, err = utils.GenerateRefreshToken(
			userID,
			scope,
			h.config.PrivateKey,
			h.config.RefreshTokenExpiry,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate refresh token")
			return
		}

		// Generate ID token with filtered claims based on scopes
		userClaims := utils.GetIDTokenClaimsForUser(user, scope, "")
		idToken, err = utils.GenerateIDToken(
			userID,
			req.ClientID,
			userClaims,
			h.config.PrivateKey,
			expiresIn,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
			return
		}
	}

	response := TokenExchangeResponse{
		AccessToken:     accessToken,
		IssuedTokenType: AccessTokenType,
		TokenType:       "Bearer",
		ExpiresIn:       expiresIn,
		RefreshToken:    refreshToken,
		IDToken:         idToken,
		Scope:           scope,
	}

	respondJSON(w, http.StatusOK, response)
}
