package handlers

import (
	"net/http"
)

type DiscoveryHandler struct {
	issuer string
}

func NewDiscoveryHandler(issuer string) *DiscoveryHandler {
	return &DiscoveryHandler{
		issuer: issuer,
	}
}

func (h *DiscoveryHandler) WellKnown(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		"issuer":                                h.issuer,
		"authorization_endpoint":                h.issuer + "/oauth/authorize",
		"token_endpoint":                        h.issuer + "/oauth/token",
		"userinfo_endpoint":                     h.issuer + "/oauth/userinfo",
		"jwks_uri":                              h.issuer + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code", "token", "id_token"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"sub", "email", "name"},
	}

	respondJSON(w, http.StatusOK, discovery)
}
