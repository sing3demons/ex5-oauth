package handlers

import (
	"net/http"
	"oauth2-server/models"
)

type DiscoveryHandler struct {
	issuer   string
	registry *models.ScopeRegistry
}

func NewDiscoveryHandler(issuer string, registry *models.ScopeRegistry) *DiscoveryHandler {
	return &DiscoveryHandler{
		issuer:   issuer,
		registry: registry,
	}
}

func (h *DiscoveryHandler) WellKnown(w http.ResponseWriter, r *http.Request) {
	// Get all registered scopes
	scopes := h.getScopesSupported()
	
	// Get all claims from all scopes
	claims := h.getClaimsSupported()
	
	discovery := map[string]interface{}{
		// Required OIDC Discovery fields
		"issuer":                                h.issuer,
		"authorization_endpoint":                h.issuer + "/oauth/authorize",
		"token_endpoint":                        h.issuer + "/oauth/token",
		"jwks_uri":                              h.issuer + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code", "token", "id_token", "code id_token", "code token", "id_token token", "code id_token token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		
		// Recommended OIDC Discovery fields
		"userinfo_endpoint":                     h.issuer + "/oauth/userinfo",
		"scopes_supported":                      scopes,
		"claims_supported":                      claims,
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials", "implicit"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		
		// Additional useful fields
		"response_modes_supported":              []string{"query", "fragment"},
		"code_challenge_methods_supported":      []string{"S256", "plain"},
		"token_endpoint_auth_signing_alg_values_supported": []string{"RS256"},
		"userinfo_signing_alg_values_supported": []string{"RS256"},
		"request_parameter_supported":           false,
		"request_uri_parameter_supported":       false,
		"require_request_uri_registration":      false,
		"claims_parameter_supported":            false,
	}

	respondJSON(w, http.StatusOK, discovery)
}

// getScopesSupported returns all registered scope names
func (h *DiscoveryHandler) getScopesSupported() []string {
	allScopes := h.registry.GetAllScopes()
	scopes := make([]string, 0, len(allScopes))
	
	for _, scope := range allScopes {
		scopes = append(scopes, scope.Name)
	}
	
	return scopes
}

// getClaimsSupported returns all claims from all registered scopes
func (h *DiscoveryHandler) getClaimsSupported() []string {
	allScopes := h.registry.GetAllScopes()
	claimsMap := make(map[string]bool)
	
	// Always include standard JWT claims
	claimsMap["sub"] = true
	claimsMap["iss"] = true
	claimsMap["aud"] = true
	claimsMap["exp"] = true
	claimsMap["iat"] = true
	claimsMap["auth_time"] = true
	claimsMap["nonce"] = true
	claimsMap["acr"] = true
	claimsMap["amr"] = true
	claimsMap["azp"] = true
	
	// Add claims from all scopes
	for _, scope := range allScopes {
		for _, claim := range scope.Claims {
			claimsMap[claim] = true
		}
	}
	
	claims := make([]string, 0, len(claimsMap))
	for claim := range claimsMap {
		claims = append(claims, claim)
	}
	
	return claims
}
