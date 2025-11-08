package utils

import (
	"errors"
	"fmt"
	"oauth2-server/models"
	"strings"
)

// GlobalScopeRegistry is the global scope registry instance
var GlobalScopeRegistry *models.ScopeRegistry

// GlobalScopeValidator is the global scope validator instance
var GlobalScopeValidator ScopeValidator

func init() {
	GlobalScopeRegistry = models.NewScopeRegistry()
	GlobalScopeValidator = NewScopeValidator(GlobalScopeRegistry)
}

// ScopeValidator interface for scope validation operations
type ScopeValidator interface {
	// ValidateScope validates scope format and existence, returns error if invalid
	ValidateScope(scope string) error
	
	// NormalizeScope removes duplicates and invalid scopes
	NormalizeScope(scope string) string
	
	// ValidateScopeName validates scope name format
	ValidateScopeName(name string) error
	
	// ValidateScopeAgainstAllowed validates against client's allowed scopes
	ValidateScopeAgainstAllowed(requested string, allowed []string) error
	
	// ValidateScopeDowngrade validates scope downgrade for refresh token flow
	ValidateScopeDowngrade(requested, original string) error
	
	// RequiresOpenID checks if openid scope is present
	RequiresOpenID(scope string) bool
}

// scopeValidator implements ScopeValidator interface
type scopeValidator struct {
	registry *models.ScopeRegistry
}

// NewScopeValidator creates a new scope validator
func NewScopeValidator(registry *models.ScopeRegistry) ScopeValidator {
	return &scopeValidator{
		registry: registry,
	}
}

// ValidateScope validates scope format and existence
func (v *scopeValidator) ValidateScope(scope string) error {
	if scope == "" {
		return errors.New("scope cannot be empty")
	}

	scopes := strings.Split(scope, " ")
	var invalidScopes []string
	
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s != "" && !v.registry.IsValidScope(s) {
			invalidScopes = append(invalidScopes, s)
		}
	}
	
	if len(invalidScopes) > 0 {
		return fmt.Errorf("invalid scopes: %v", invalidScopes)
	}
	
	return nil
}

// NormalizeScope removes duplicates, trims whitespace, and removes invalid scopes
func (v *scopeValidator) NormalizeScope(scope string) string {
	if scope == "" {
		return ""
	}

	scopes := strings.Split(scope, " ")
	seen := make(map[string]bool)
	var normalized []string

	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s != "" && v.registry.IsValidScope(s) && !seen[s] {
			seen[s] = true
			normalized = append(normalized, s)
		}
	}

	return strings.Join(normalized, " ")
}

// ValidateScopeName validates scope name format
// Allows alphanumeric, underscore, hyphen, colon, period
func (v *scopeValidator) ValidateScopeName(name string) error {
	if name == "" {
		return errors.New("scope name cannot be empty")
	}
	
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == ':' || ch == '.') {
			return fmt.Errorf("invalid scope name format: %s (only alphanumeric, underscore, hyphen, colon, and period allowed)", name)
		}
	}
	
	return nil
}

// ValidateScopeAgainstAllowed validates requested scopes against client's allowed scopes
func (v *scopeValidator) ValidateScopeAgainstAllowed(requested string, allowed []string) error {
	if len(allowed) == 0 {
		// No restrictions - all scopes allowed
		return nil
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

	if len(unauthorized) > 0 {
		return fmt.Errorf("client is not authorized for scopes: %v", unauthorized)
	}
	
	return nil
}

// ValidateScopeDowngrade validates that requested scopes are subset of original scopes
func (v *scopeValidator) ValidateScopeDowngrade(requested, original string) error {
	if requested == "" {
		// Empty requested scope means use original scopes
		return nil
	}
	
	// Build map of original scopes
	originalMap := make(map[string]bool)
	for _, s := range strings.Split(original, " ") {
		s = strings.TrimSpace(s)
		if s != "" {
			originalMap[s] = true
		}
	}
	
	// Check if all requested scopes are in original
	requestedScopes := strings.Split(requested, " ")
	var escalated []string
	
	for _, s := range requestedScopes {
		s = strings.TrimSpace(s)
		if s != "" && !originalMap[s] {
			escalated = append(escalated, s)
		}
	}
	
	if len(escalated) > 0 {
		return fmt.Errorf("cannot escalate scopes during refresh: %v", escalated)
	}
	
	return nil
}

// RequiresOpenID checks if openid scope is present (required for OIDC)
func (v *scopeValidator) RequiresOpenID(scope string) bool {
	return HasScope(scope, "openid")
}

// ValidateScope checks if requested scopes are valid (backward compatible helper)
func ValidateScope(scope string) bool {
	return GlobalScopeValidator.ValidateScope(scope) == nil
}

// ValidateScopeName checks if a scope name is valid format (backward compatible helper)
func ValidateScopeName(name string) bool {
	return GlobalScopeValidator.ValidateScopeName(name) == nil
}

// NormalizeScope removes duplicates and invalid scopes (backward compatible helper)
func NormalizeScope(scope string) string {
	return GlobalScopeValidator.NormalizeScope(scope)
}

// ValidateScopeAgainstAllowed checks if requested scopes are within allowed scopes (backward compatible helper)
// Returns (isValid, unauthorizedScopes)
func ValidateScopeAgainstAllowed(requested string, allowed []string) (bool, []string) {
	err := GlobalScopeValidator.ValidateScopeAgainstAllowed(requested, allowed)
	if err == nil {
		return true, nil
	}
	
	// Extract unauthorized scopes from error message
	// This maintains backward compatibility with existing code
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

	return false, unauthorized
}

// ValidateScopeDowngrade validates scope downgrade for refresh token flow (helper)
func ValidateScopeDowngrade(requested, original string) error {
	return GlobalScopeValidator.ValidateScopeDowngrade(requested, original)
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
