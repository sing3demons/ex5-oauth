package handlers

import (
	"context"
	"html/template"
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

type ConsentHandler struct {
	clientRepo   *repository.ClientRepository
	consentRepo  *repository.UserConsentRepository
	authCodeRepo *repository.AuthCodeRepository
	sessionRepo  *repository.SessionRepository
	config       *config.Config
}

func NewConsentHandler(
	clientRepo *repository.ClientRepository,
	consentRepo *repository.UserConsentRepository,
	authCodeRepo *repository.AuthCodeRepository,
	sessionRepo *repository.SessionRepository,
	cfg *config.Config,
) *ConsentHandler {
	return &ConsentHandler{
		clientRepo:   clientRepo,
		consentRepo:  consentRepo,
		authCodeRepo: authCodeRepo,
		sessionRepo:  sessionRepo,
		config:       cfg,
	}
}

// ShowConsent renders the consent screen with client info and scope descriptions
func (h *ConsentHandler) ShowConsent(w http.ResponseWriter, r *http.Request) {
	// Extract parameters from query string
	clientID := r.URL.Query().Get("client_id")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")
	nonce := r.URL.Query().Get("nonce")

	if clientID == "" || redirectURI == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	ctx := context.Background()

	// Fetch client information
	client, err := h.clientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusBadRequest, "invalid_client", "Client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to fetch client")
		return
	}

	// Parse scopes
	scopes := strings.Split(scope, " ")
	scopeDescriptions := make([]string, len(scopes))

	// Get scope descriptions from registry
	for i, scopeName := range scopes {
		if scopeDef, exists := utils.GlobalScopeRegistry.GetScope(scopeName); exists {
			scopeDescriptions[i] = scopeDef.Description
		} else {
			scopeDescriptions[i] = "Access to " + scopeName
		}
	}

	// Prepare template data
	data := map[string]interface{}{
		"ClientName":            client.Name,
		"ClientID":              clientID,
		"Scopes":                scopes,
		"ScopeDescriptions":     scopeDescriptions,
		"ScopeString":           scope,
		"State":                 state,
		"RedirectURI":           redirectURI,
		"ResponseType":          responseType,
		"CodeChallenge":         codeChallenge,
		"CodeChallengeMethod":   codeChallengeMethod,
		"Nonce":                 nonce,
	}

	// Render consent template
	tmpl, err := template.ParseFiles("templates/consent.html")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to load template")
		return
	}

	tmpl.Execute(w, data)
}

// HandleConsent processes the consent form submission
func (h *ConsentHandler) HandleConsent(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	action := r.FormValue("action")
	clientID := r.FormValue("client_id")
	scope := r.FormValue("scope")
	state := r.FormValue("state")
	redirectURI := r.FormValue("redirect_uri")
	codeChallenge := r.FormValue("code_challenge")
	codeChallengeMethod := r.FormValue("code_challenge_method")
	nonce := r.FormValue("nonce")

	if clientID == "" || redirectURI == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required parameters")
		return
	}

	// Get SSO session from context
	ssoSession, ok := r.Context().Value(middleware.SSOSessionContextKey).(*models.SSOSession)
	if !ok || ssoSession == nil || !ssoSession.Authenticated {
		// No valid SSO session - redirect to login
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ctx := context.Background()

	// Handle denial
	if action == "deny" {
		// Build error redirect URL
		errorURL := redirectURI + "?error=access_denied&error_description=User+denied+consent"
		if state != "" {
			errorURL += "&state=" + state
		}
		http.Redirect(w, r, errorURL, http.StatusFound)
		return
	}

	// Handle approval
	if action == "allow" {
		// Parse scopes
		scopes := strings.Split(scope, " ")

		// Save UserConsent record with 1-year expiration
		consent := &models.UserConsent{
			UserID:    ssoSession.UserID,
			ClientID:  clientID,
			Scopes:    scopes,
			GrantedAt: time.Now(),
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year
		}

		// Create or update consent
		if err := h.consentRepo.Create(ctx, consent); err != nil {
			// If consent already exists, it's fine - the unique index will prevent duplicates
			// We can continue with authorization code generation
			if !mongo.IsDuplicateKeyError(err) {
				respondError(w, http.StatusInternalServerError, "server_error", "Failed to save consent")
				return
			}
		}

		// Generate authorization code
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
			ChallengeMethod: codeChallengeMethod,
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

	// Invalid action
	respondError(w, http.StatusBadRequest, "invalid_request", "Invalid action")
}
