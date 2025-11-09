package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"oauth2-server/config"
	"oauth2-server/repository"
	"oauth2-server/utils"
	"strings"

	"github.com/gorilla/mux"
)

// SessionHandler handles session management endpoints for SSO sessions
// It provides endpoints to list and revoke user sessions
type SessionHandler struct {
	ssoSessionRepo *repository.SSOSessionRepository
	consentRepo    *repository.UserConsentRepository
	clientRepo     *repository.ClientRepository
	config         *config.Config
}

func NewSessionHandler(
	ssoSessionRepo *repository.SSOSessionRepository,
	consentRepo *repository.UserConsentRepository,
	clientRepo *repository.ClientRepository,
	cfg *config.Config,
) *SessionHandler {
	return &SessionHandler{
		ssoSessionRepo: ssoSessionRepo,
		consentRepo:    consentRepo,
		clientRepo:     clientRepo,
		config:         cfg,
	}
}

// SessionResponse represents a session in the API response
type SessionResponse struct {
	SessionID    string `json:"session_id"`
	CreatedAt    string `json:"created_at"`
	LastActivity string `json:"last_activity"`
	ExpiresAt    string `json:"expires_at"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// ListSessionsResponse represents the response for listing sessions
type ListSessionsResponse struct {
	Sessions []SessionResponse `json:"sessions"`
}

// AuthorizationResponse represents an authorization in the API response
type AuthorizationResponse struct {
	ClientID   string   `json:"client_id"`
	ClientName string   `json:"client_name"`
	Scopes     []string `json:"scopes"`
	GrantedAt  string   `json:"granted_at"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
}

// ListAuthorizationsResponse represents the response for listing authorizations
type ListAuthorizationsResponse struct {
	Authorizations []AuthorizationResponse `json:"authorizations"`
}

// extractUserIDFromToken extracts and validates the user ID from the Authorization header
func (h *SessionHandler) extractUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", &AuthError{Code: "unauthorized", Message: "Authorization required"}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	
	// Support both JWT and JWE tokens
	if utils.IsJWE(tokenString) {
		jweClaims, err := utils.ValidateJWE(tokenString, h.config.PrivateKey)
		if err != nil {
			return "", &AuthError{Code: "invalid_token", Message: "Invalid or expired token"}
		}
		return jweClaims.UserID, nil
	}
	
	jwtClaims, err := utils.ValidateToken(tokenString, h.config.PublicKey)
	if err != nil {
		return "", &AuthError{Code: "invalid_token", Message: "Invalid or expired token"}
	}
	
	return jwtClaims.UserID, nil
}

// AuthError represents an authentication error
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// ListSessions returns all active SSO sessions for the authenticated user
// GET /account/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from access token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			respondError(w, http.StatusUnauthorized, authErr.Code, authErr.Message)
			return
		}
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication failed")
		return
	}

	ctx := context.Background()
	
	// Fetch all sessions for the user
	sessions, err := h.ssoSessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve sessions")
		return
	}

	// Convert to response format
	sessionResponses := make([]SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		sessionResponses = append(sessionResponses, SessionResponse{
			SessionID:    session.SessionID,
			CreatedAt:    session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActivity: session.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:    session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			IPAddress:    session.IPAddress,
			UserAgent:    session.UserAgent,
		})
	}

	response := ListSessionsResponse{
		Sessions: sessionResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RevokeSession deletes a specific SSO session by session ID
// DELETE /account/sessions/{session_id}
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from access token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			respondError(w, http.StatusUnauthorized, authErr.Code, authErr.Message)
			return
		}
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication failed")
		return
	}

	// Extract session_id from URL path
	vars := mux.Vars(r)
	sessionID := vars["session_id"]
	
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Session ID is required")
		return
	}

	ctx := context.Background()
	
	// Verify the session belongs to the authenticated user
	session, err := h.ssoSessionRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Session not found")
		return
	}

	if session.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden", "You can only revoke your own sessions")
		return
	}

	// Delete the session
	if err := h.ssoSessionRepo.Delete(ctx, sessionID); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to revoke session")
		return
	}

	// Return success response
	response := map[string]string{
		"message": "Session revoked successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListAuthorizations returns all active authorizations (consents) for the authenticated user
// GET /account/authorizations
func (h *SessionHandler) ListAuthorizations(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from access token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			respondError(w, http.StatusUnauthorized, authErr.Code, authErr.Message)
			return
		}
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication failed")
		return
	}

	ctx := context.Background()
	
	// Fetch all consents for the user
	consents, err := h.consentRepo.ListUserConsents(ctx, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to retrieve authorizations")
		return
	}

	// Convert to response format with client information
	authResponses := make([]AuthorizationResponse, 0, len(consents))
	for _, consent := range consents {
		// Fetch client information to include client name
		client, err := h.clientRepo.FindByClientID(ctx, consent.ClientID)
		clientName := consent.ClientID // Default to client ID if fetch fails
		if err == nil && client != nil {
			clientName = client.Name
		}

		expiresAt := ""
		if !consent.ExpiresAt.IsZero() {
			expiresAt = consent.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}

		authResponses = append(authResponses, AuthorizationResponse{
			ClientID:   consent.ClientID,
			ClientName: clientName,
			Scopes:     consent.Scopes,
			GrantedAt:  consent.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:  expiresAt,
		})
	}

	response := ListAuthorizationsResponse{
		Authorizations: authResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RevokeAuthorization revokes authorization (consent) for a specific client application
// DELETE /account/authorizations/{client_id}
func (h *SessionHandler) RevokeAuthorization(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from access token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			respondError(w, http.StatusUnauthorized, authErr.Code, authErr.Message)
			return
		}
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication failed")
		return
	}

	// Extract client_id from URL path
	vars := mux.Vars(r)
	clientID := vars["client_id"]
	
	if clientID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Client ID is required")
		return
	}

	ctx := context.Background()
	
	// Verify the consent exists for this user and client
	consent, err := h.consentRepo.FindByUserAndClient(ctx, userID, clientID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Authorization not found")
		return
	}

	// Additional check to ensure the consent belongs to the authenticated user
	if consent.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden", "You can only revoke your own authorizations")
		return
	}

	// Delete the consent record
	if err := h.consentRepo.RevokeConsent(ctx, userID, clientID); err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to revoke authorization")
		return
	}

	// Return success response
	response := map[string]string{
		"message": "Authorization revoked successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
