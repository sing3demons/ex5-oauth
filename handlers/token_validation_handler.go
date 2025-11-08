package handlers

import (
	"encoding/json"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/utils"
	"strings"
)

type TokenValidationHandler struct {
	config *config.Config
}

func NewTokenValidationHandler(cfg *config.Config) *TokenValidationHandler {
	return &TokenValidationHandler{
		config: cfg,
	}
}

type TokenValidationRequest struct {
	Token string `json:"token"`
}

type TokenValidationResponse struct {
	Valid     bool                   `json:"valid"`
	TokenType string                 `json:"token_type,omitempty"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ExpiresAt int64                  `json:"expires_at,omitempty"`
	IssuedAt  int64                  `json:"issued_at,omitempty"`
}

func (h *TokenValidationHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req TokenValidationRequest

	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
			return
		}
	} else {
		req.Token = r.URL.Query().Get("token")
		if req.Token == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				req.Token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
	}

	if req.Token == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing token parameter")
		return
	}

	response := TokenValidationResponse{}

	if utils.IsJWE(req.Token) {
		claims, err := utils.ValidateJWE(req.Token, h.config.PrivateKey)
		if err != nil {
			response.Valid = false
			response.Error = err.Error()
		} else {
			response.Valid = true
			response.TokenType = "JWE"
			response.Claims = map[string]interface{}{
				"sub":   claims.UserID,
				"email": claims.Email,
				"name":  claims.Name,
				"scope": claims.Scope,
				"aud":   claims.Aud,
			}
			response.ExpiresAt = claims.Exp
			response.IssuedAt = claims.Iat
		}
	} else if utils.IsJWT(req.Token) {
		claims, err := utils.ValidateToken(req.Token, h.config.PublicKey)
		if err != nil {
			response.Valid = false
			response.Error = err.Error()
		} else {
			response.Valid = true
			response.TokenType = "JWT"
			response.Claims = map[string]interface{}{
				"sub":   claims.UserID,
				"email": claims.Email,
				"name":  claims.Name,
				"scope": claims.Scope,
			}
			if claims.ExpiresAt != nil {
				response.ExpiresAt = claims.ExpiresAt.Unix()
			}
			if claims.IssuedAt != nil {
				response.IssuedAt = claims.IssuedAt.Unix()
			}
		}
	} else {
		response.Valid = false
		response.Error = "Invalid token format"
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *TokenValidationHandler) ValidateTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authorization header required")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Bearer token required")
			return
		}

		var valid bool

		if utils.IsJWE(token) {
			_, err := utils.ValidateJWE(token, h.config.PrivateKey)
			valid = err == nil
		} else if utils.IsJWT(token) {
			_, err := utils.ValidateToken(token, h.config.PublicKey)
			valid = err == nil
		}

		if !valid {
			respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}

		next(w, r)
	}
}
