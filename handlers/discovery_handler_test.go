package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"oauth2-server/models"
	"testing"
)

func TestDiscoveryHandler_WellKnown(t *testing.T) {
	// Create a scope registry with standard OIDC scopes
	registry := models.NewScopeRegistry()
	
	// Create discovery handler
	handler := NewDiscoveryHandler("https://example.com", registry)
	
	// Create test request
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.WellKnown(w, req)
	
	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Parse response
	var discovery map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &discovery); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Verify required OIDC fields
	requiredFields := []string{
		"issuer",
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"response_types_supported",
		"subject_types_supported",
		"id_token_signing_alg_values_supported",
	}
	
	for _, field := range requiredFields {
		if _, exists := discovery[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}
	
	// Verify scopes_supported field
	if _, exists := discovery["scopes_supported"]; !exists {
		t.Error("Missing scopes_supported field")
	}
	
	scopes, ok := discovery["scopes_supported"].([]interface{})
	if !ok {
		t.Fatal("scopes_supported is not an array")
	}
	
	// Verify standard OIDC scopes are present
	expectedScopes := map[string]bool{
		"openid":         false,
		"profile":        false,
		"email":          false,
		"phone":          false,
		"address":        false,
		"offline_access": false,
	}
	
	for _, scope := range scopes {
		scopeName, ok := scope.(string)
		if !ok {
			continue
		}
		if _, exists := expectedScopes[scopeName]; exists {
			expectedScopes[scopeName] = true
		}
	}
	
	for scope, found := range expectedScopes {
		if !found {
			t.Errorf("Expected scope %s not found in scopes_supported", scope)
		}
	}
	
	// Verify grant_types_supported field
	if _, exists := discovery["grant_types_supported"]; !exists {
		t.Error("Missing grant_types_supported field")
	}
	
	grantTypes, ok := discovery["grant_types_supported"].([]interface{})
	if !ok {
		t.Fatal("grant_types_supported is not an array")
	}
	
	expectedGrantTypes := map[string]bool{
		"authorization_code": false,
		"refresh_token":      false,
		"client_credentials": false,
	}
	
	for _, gt := range grantTypes {
		gtName, ok := gt.(string)
		if !ok {
			continue
		}
		if _, exists := expectedGrantTypes[gtName]; exists {
			expectedGrantTypes[gtName] = true
		}
	}
	
	for gt, found := range expectedGrantTypes {
		if !found {
			t.Errorf("Expected grant type %s not found in grant_types_supported", gt)
		}
	}
	
	// Verify response_types_supported field
	if _, exists := discovery["response_types_supported"]; !exists {
		t.Error("Missing response_types_supported field")
	}
	
	responseTypes, ok := discovery["response_types_supported"].([]interface{})
	if !ok {
		t.Fatal("response_types_supported is not an array")
	}
	
	if len(responseTypes) == 0 {
		t.Error("response_types_supported should not be empty")
	}
	
	// Verify claims_supported field
	if _, exists := discovery["claims_supported"]; !exists {
		t.Error("Missing claims_supported field")
	}
	
	claims, ok := discovery["claims_supported"].([]interface{})
	if !ok {
		t.Fatal("claims_supported is not an array")
	}
	
	// Verify standard claims are present
	expectedClaims := map[string]bool{
		"sub":   false,
		"iss":   false,
		"aud":   false,
		"exp":   false,
		"iat":   false,
		"email": false,
		"name":  false,
	}
	
	for _, claim := range claims {
		claimName, ok := claim.(string)
		if !ok {
			continue
		}
		if _, exists := expectedClaims[claimName]; exists {
			expectedClaims[claimName] = true
		}
	}
	
	for claim, found := range expectedClaims {
		if !found {
			t.Errorf("Expected claim %s not found in claims_supported", claim)
		}
	}
}

func TestDiscoveryHandler_GetScopesSupported(t *testing.T) {
	registry := models.NewScopeRegistry()
	handler := NewDiscoveryHandler("https://example.com", registry)
	
	scopes := handler.getScopesSupported()
	
	if len(scopes) == 0 {
		t.Error("Expected at least one scope")
	}
	
	// Verify standard OIDC scopes
	expectedScopes := []string{"openid", "profile", "email", "phone", "address", "offline_access"}
	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s] = true
	}
	
	for _, expected := range expectedScopes {
		if !scopeMap[expected] {
			t.Errorf("Expected scope %s not found", expected)
		}
	}
}

func TestDiscoveryHandler_GetClaimsSupported(t *testing.T) {
	registry := models.NewScopeRegistry()
	handler := NewDiscoveryHandler("https://example.com", registry)
	
	claims := handler.getClaimsSupported()
	
	if len(claims) == 0 {
		t.Error("Expected at least one claim")
	}
	
	// Verify standard JWT claims
	expectedClaims := []string{"sub", "iss", "aud", "exp", "iat"}
	claimMap := make(map[string]bool)
	for _, c := range claims {
		claimMap[c] = true
	}
	
	for _, expected := range expectedClaims {
		if !claimMap[expected] {
			t.Errorf("Expected claim %s not found", expected)
		}
	}
	
	// Verify profile claims
	profileClaims := []string{"name", "email"}
	for _, claim := range profileClaims {
		if !claimMap[claim] {
			t.Errorf("Expected profile claim %s not found", claim)
		}
	}
}
