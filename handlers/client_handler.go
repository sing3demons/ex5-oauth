package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"oauth2-server/models"
	"oauth2-server/repository"
	"oauth2-server/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

type ClientHandler struct {
	clientRepo *repository.ClientRepository
}

func NewClientHandler(clientRepo *repository.ClientRepository) *ClientHandler {
	return &ClientHandler{
		clientRepo: clientRepo,
	}
}

func (h *ClientHandler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string   `json:"name"`
		RedirectURIs  []string `json:"redirect_uris"`
		AllowedScopes []string `json:"allowed_scopes,omitempty"`
		GrantTypes    []string `json:"grant_types,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" || len(req.RedirectURIs) == 0 {
		respondError(w, http.StatusBadRequest, "invalid_request", "Missing required fields")
		return
	}

	clientID, err := utils.GenerateRandomString(32)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate client ID")
		return
	}

	clientSecret, err := utils.GenerateRandomString(64)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to generate client secret")
		return
	}

	client := &models.Client{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURIs:  req.RedirectURIs,
		Name:          req.Name,
		AllowedScopes: req.AllowedScopes,
		GrantTypes:    req.GrantTypes,
	}

	ctx := context.Background()
	if err := h.clientRepo.Create(ctx, client); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(w, http.StatusConflict, "client_exists", "Client already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "server_error", "Failed to create client")
		return
	}

	response := map[string]interface{}{
		"client_id":     client.ClientID,
		"client_secret": client.ClientSecret,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
	}
	
	if len(client.AllowedScopes) > 0 {
		response["allowed_scopes"] = client.AllowedScopes
	}
	
	if len(client.GrantTypes) > 0 {
		response["grant_types"] = client.GrantTypes
	}

	respondJSON(w, http.StatusCreated, response)
}
