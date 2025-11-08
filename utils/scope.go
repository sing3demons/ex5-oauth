package utils

import (
	"oauth2-server/models"
	"strings"
)

// GlobalScopeRegistry is the global scope registry instance
var GlobalScopeRegistry *models.ScopeRegistry

func init() {
	GlobalScopeRegistry = models.NewScopeRegistry()
}

// ValidateScope checks if requested scopes are valid
func ValidateScope(scope string) bool {
	if scope == "" {
		return false
	}

	scopes := strings.Split(scope, " ")
	for _, s := range scopes {
		if s != "" && !GlobalScopeRegistry.IsValidScope(s) {
			return false
		}
	}
	return true
}

// ValidateScopeName checks if a scope name is valid format
// Allows alphanumeric, underscore, hyphen, colon, period
func ValidateScopeName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == ':' || ch == '.') {
			return false
		}
	}
	return true
}

// NormalizeScope removes duplicates and invalid scopes
func NormalizeScope(scope string) string {
	if scope == "" {
		return ""
	}

	scopes := strings.Split(scope, " ")
	seen := make(map[string]bool)
	var normalized []string

	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s != "" && GlobalScopeRegistry.IsValidScope(s) && !seen[s] {
			seen[s] = true
			normalized = append(normalized, s)
		}
	}

	return strings.Join(normalized, " ")
}

// ValidateScopeAgainstAllowed checks if requested scopes are within allowed scopes
// Returns (isValid, unauthorizedScopes)
func ValidateScopeAgainstAllowed(requested string, allowed []string) (bool, []string) {
	if len(allowed) == 0 {
		// No restrictions - all scopes allowed
		return true, nil
	}

	allowedMap := make(map[string]bool)
	for _, s := range allowed {
		allowedMap[s] = true
	}

	requestedScopes := strings.Split(requested, " ")
	var unauthorized []string

	for _, s := range requestedScopes {
		s = strings.TrimSpace(s)
		if s != "" && !allowedMap[s] {
			unauthorized = append(unauthorized, s)
		}
	}

	return len(unauthorized) == 0, unauthorized
}

// HasScope checks if a scope string contains a specific scope
func HasScope(scopeString, targetScope string) bool {
	scopes := strings.Split(scopeString, " ")
	for _, s := range scopes {
		if s == targetScope {
			return true
		}
	}
	return false
}

// IntersectScopes returns the intersection of requested and allowed scopes
func IntersectScopes(requested, allowed string) string {
	if requested == "" || allowed == "" {
		return ""
	}

	requestedScopes := strings.Split(requested, " ")
	allowedMap := make(map[string]bool)
	
	for _, s := range strings.Split(allowed, " ") {
		if s != "" {
			allowedMap[s] = true
		}
	}

	var intersection []string
	seen := make(map[string]bool)

	for _, s := range requestedScopes {
		s = strings.TrimSpace(s)
		if s != "" && allowedMap[s] && !seen[s] {
			seen[s] = true
			intersection = append(intersection, s)
		}
	}

	return strings.Join(intersection, " ")
}

// GetDefaultScope returns default scope if none provided
func GetDefaultScope() string {
	return "openid profile email"
}

// RequiresOpenID checks if openid scope is present (required for OIDC)
func RequiresOpenID(scope string) bool {
	return HasScope(scope, "openid")
}

// ScopeIncludesProfile checks if profile scope is requested
func ScopeIncludesProfile(scope string) bool {
	return HasScope(scope, "profile")
}

// ScopeIncludesEmail checks if email scope is requested
func ScopeIncludesEmail(scope string) bool {
	return HasScope(scope, "email")
}

// ScopeIncludesOfflineAccess checks if offline_access scope is requested
func ScopeIncludesOfflineAccess(scope string) bool {
	return HasScope(scope, "offline_access")
}
