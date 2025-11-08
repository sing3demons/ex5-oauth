package utils

import (
	"oauth2-server/models"
	"strings"
)

// ClaimFilter interface for filtering user claims based on scopes
type ClaimFilter interface {
	// FilterClaims filters user claims based on granted scopes
	FilterClaims(user *models.User, scopes string) map[string]interface{}
	
	// GetIDTokenClaims gets claims for ID token based on scopes
	GetIDTokenClaims(user *models.User, scopes string, nonce string) map[string]interface{}
}

// claimFilter implements ClaimFilter interface
type claimFilter struct {
	registry *models.ScopeRegistry
}

// NewClaimFilter creates a new claim filter
func NewClaimFilter(registry *models.ScopeRegistry) ClaimFilter {
	return &claimFilter{
		registry: registry,
	}
}

// GlobalClaimFilter is the global claim filter instance
var GlobalClaimFilter ClaimFilter

func init() {
	// Ensure GlobalScopeRegistry is initialized
	if GlobalScopeRegistry == nil {
		GlobalScopeRegistry = models.NewScopeRegistry()
	}
	GlobalClaimFilter = NewClaimFilter(GlobalScopeRegistry)
}

// FilterClaims filters user claims based on granted scopes
func (f *claimFilter) FilterClaims(user *models.User, scopes string) map[string]interface{} {
	claims := make(map[string]interface{})
	
	// Always include sub (subject) claim
	claims["sub"] = user.ID
	
	// Get allowed claims from scope registry
	scopeList := strings.Split(scopes, " ")
	allowedClaims := f.registry.GetClaimsForScopes(scopeList)
	
	// Build map for O(1) lookup
	claimMap := make(map[string]bool)
	for _, c := range allowedClaims {
		claimMap[c] = true
	}
	
	// Add email claims if email scope is present
	if claimMap["email"] {
		claims["email"] = user.Email
	}
	if claimMap["email_verified"] {
		claims["email_verified"] = true
	}
	
	// Add profile claims if profile scope is present
	if claimMap["name"] {
		claims["name"] = user.Name
	}
	
	// Additional profile claims (if available in User model)
	// These would be added when User model is extended
	if claimMap["picture"] {
		// claims["picture"] = user.Picture
	}
	if claimMap["preferred_username"] {
		// claims["preferred_username"] = user.PreferredUsername
	}
	if claimMap["profile"] {
		// claims["profile"] = user.ProfileURL
	}
	if claimMap["website"] {
		// claims["website"] = user.Website
	}
	if claimMap["gender"] {
		// claims["gender"] = user.Gender
	}
	if claimMap["birthdate"] {
		// claims["birthdate"] = user.Birthdate
	}
	if claimMap["zoneinfo"] {
		// claims["zoneinfo"] = user.Zoneinfo
	}
	if claimMap["locale"] {
		// claims["locale"] = user.Locale
	}
	if claimMap["updated_at"] {
		// claims["updated_at"] = user.UpdatedAt
	}
	
	// Phone claims (if phone scope is present)
	if claimMap["phone_number"] {
		// claims["phone_number"] = user.PhoneNumber
	}
	if claimMap["phone_number_verified"] {
		// claims["phone_number_verified"] = user.PhoneNumberVerified
	}
	
	// Address claim (if address scope is present)
	if claimMap["address"] {
		// claims["address"] = user.Address
	}
	
	return claims
}

// GetIDTokenClaims gets claims for ID token based on scopes
// This is specifically for ID tokens and includes nonce if provided
func (f *claimFilter) GetIDTokenClaims(user *models.User, scopes string, nonce string) map[string]interface{} {
	// Start with filtered claims based on scopes
	claims := f.FilterClaims(user, scopes)
	
	// Add nonce if provided (for replay protection)
	if nonce != "" {
		claims["nonce"] = nonce
	}
	
	return claims
}

// Helper functions for backward compatibility and convenience

// FilterClaimsForUser filters user claims based on scopes (helper function)
func FilterClaimsForUser(user *models.User, scopes string) map[string]interface{} {
	return GlobalClaimFilter.FilterClaims(user, scopes)
}

// GetIDTokenClaimsForUser gets ID token claims for user (helper function)
func GetIDTokenClaimsForUser(user *models.User, scopes string, nonce string) map[string]interface{} {
	return GlobalClaimFilter.GetIDTokenClaims(user, scopes, nonce)
}
