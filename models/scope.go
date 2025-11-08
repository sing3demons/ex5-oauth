package models

// ScopeDefinition represents a scope with its metadata
type ScopeDefinition struct {
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Claims      []string `json:"claims,omitempty" bson:"claims,omitempty"`
	IsDefault   bool     `json:"is_default" bson:"is_default"`
	ParentScope string   `json:"parent_scope,omitempty" bson:"parent_scope,omitempty"`
	ChildScopes []string `json:"child_scopes,omitempty" bson:"child_scopes,omitempty"`
}

// ScopeRegistry holds all available scopes
type ScopeRegistry struct {
	Scopes map[string]*ScopeDefinition
}

// NewScopeRegistry creates a new scope registry with standard OIDC scopes
func NewScopeRegistry() *ScopeRegistry {
	registry := &ScopeRegistry{
		Scopes: make(map[string]*ScopeDefinition),
	}

	// Standard OIDC scopes
	registry.RegisterScope(&ScopeDefinition{
		Name:        "openid",
		Description: "OpenID Connect authentication",
		Claims:      []string{"sub"},
		IsDefault:   true,
	})

	registry.RegisterScope(&ScopeDefinition{
		Name:        "profile",
		Description: "Access to user profile information",
		Claims: []string{
			"name", "family_name", "given_name", "middle_name",
			"nickname", "preferred_username", "profile", "picture",
			"website", "gender", "birthdate", "zoneinfo", "locale", "updated_at",
		},
		IsDefault: true,
	})

	registry.RegisterScope(&ScopeDefinition{
		Name:        "email",
		Description: "Access to user email address",
		Claims:      []string{"email", "email_verified"},
		IsDefault:   true,
	})

	registry.RegisterScope(&ScopeDefinition{
		Name:        "phone",
		Description: "Access to user phone number",
		Claims:      []string{"phone_number", "phone_number_verified"},
		IsDefault:   false,
	})

	registry.RegisterScope(&ScopeDefinition{
		Name:        "address",
		Description: "Access to user postal address",
		Claims:      []string{"address"},
		IsDefault:   false,
	})

	registry.RegisterScope(&ScopeDefinition{
		Name:        "offline_access",
		Description: "Request refresh token for offline access",
		Claims:      []string{},
		IsDefault:   false,
	})

	return registry
}

// RegisterScope adds a scope to the registry
func (r *ScopeRegistry) RegisterScope(scope *ScopeDefinition) {
	r.Scopes[scope.Name] = scope
}

// GetScope retrieves a scope definition
func (r *ScopeRegistry) GetScope(name string) (*ScopeDefinition, bool) {
	scope, exists := r.Scopes[name]
	return scope, exists
}

// IsValidScope checks if a scope exists in the registry
func (r *ScopeRegistry) IsValidScope(name string) bool {
	_, exists := r.Scopes[name]
	return exists
}

// GetAllScopes returns all registered scopes
func (r *ScopeRegistry) GetAllScopes() []*ScopeDefinition {
	scopes := make([]*ScopeDefinition, 0, len(r.Scopes))
	for _, scope := range r.Scopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// GetDefaultScopes returns scopes marked as default
func (r *ScopeRegistry) GetDefaultScopes() []string {
	var defaults []string
	for name, scope := range r.Scopes {
		if scope.IsDefault {
			defaults = append(defaults, name)
		}
	}
	return defaults
}

// GetClaimsForScopes returns all claims for given scopes
func (r *ScopeRegistry) GetClaimsForScopes(scopes []string) []string {
	claimsMap := make(map[string]bool)
	
	for _, scopeName := range scopes {
		if scope, exists := r.Scopes[scopeName]; exists {
			for _, claim := range scope.Claims {
				claimsMap[claim] = true
			}
		}
	}

	claims := make([]string, 0, len(claimsMap))
	for claim := range claimsMap {
		claims = append(claims, claim)
	}
	return claims
}

// ExpandScopes expands parent scopes to include child scopes
func (r *ScopeRegistry) ExpandScopes(scopes []string) []string {
	expanded := make(map[string]bool)
	
	var expand func(string)
	expand = func(scopeName string) {
		if expanded[scopeName] {
			return
		}
		
		scope, exists := r.Scopes[scopeName]
		if !exists {
			return
		}
		
		expanded[scopeName] = true
		
		// Expand child scopes
		for _, child := range scope.ChildScopes {
			expand(child)
		}
	}

	for _, scopeName := range scopes {
		expand(scopeName)
	}

	result := make([]string, 0, len(expanded))
	for scopeName := range expanded {
		result = append(result, scopeName)
	}
	return result
}
