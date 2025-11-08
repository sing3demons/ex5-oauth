package utils

import (
	"testing"
)

func TestValidateScope(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		want  bool
	}{
		{"valid single scope", "openid", true},
		{"valid multiple scopes", "openid profile email", true},
		{"invalid scope", "invalid_scope", false},
		{"mixed valid and invalid", "openid invalid_scope", false},
		{"empty scope", "", false},
		{"all valid scopes", "openid profile email phone address", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateScope(tt.scope); got != tt.want {
				t.Errorf("ValidateScope(%q) = %v, want %v", tt.scope, got, tt.want)
			}
		})
	}
}

func TestNormalizeScope(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		want  string
	}{
		{"remove duplicates", "openid openid profile", "openid profile"},
		{"remove invalid", "openid invalid profile", "openid profile"},
		{"trim spaces", "  openid   profile  ", "openid profile"},
		{"empty scope", "", ""},
		{"all invalid", "invalid1 invalid2", ""},
		{"preserve order", "profile openid email", "profile openid email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeScope(tt.scope); got != tt.want {
				t.Errorf("NormalizeScope(%q) = %q, want %q", tt.scope, got, tt.want)
			}
		})
	}
}

func TestHasScope(t *testing.T) {
	tests := []struct {
		name        string
		scopeString string
		targetScope string
		want        bool
	}{
		{"has scope", "openid profile email", "profile", true},
		{"doesn't have scope", "openid profile", "email", false},
		{"empty scope string", "", "openid", false},
		{"single scope match", "openid", "openid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasScope(tt.scopeString, tt.targetScope); got != tt.want {
				t.Errorf("HasScope(%q, %q) = %v, want %v", tt.scopeString, tt.targetScope, got, tt.want)
			}
		})
	}
}

func TestIntersectScopes(t *testing.T) {
	tests := []struct {
		name      string
		requested string
		allowed   string
		want      string
	}{
		{"full intersection", "openid profile", "openid profile email", "openid profile"},
		{"partial intersection", "openid profile email", "openid email", "openid email"},
		{"no intersection", "openid profile", "email phone", ""},
		{"empty requested", "", "openid profile", ""},
		{"empty allowed", "openid profile", "", ""},
		{"duplicate handling", "openid openid profile", "openid profile", "openid profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntersectScopes(tt.requested, tt.allowed); got != tt.want {
				t.Errorf("IntersectScopes(%q, %q) = %q, want %q", tt.requested, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestGetDefaultScope(t *testing.T) {
	got := GetDefaultScope()
	want := "openid profile email"
	if got != want {
		t.Errorf("GetDefaultScope() = %q, want %q", got, want)
	}
}

func TestRequiresOpenID(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		want  bool
	}{
		{"has openid", "openid profile", true},
		{"no openid", "profile email", false},
		{"only openid", "openid", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequiresOpenID(tt.scope); got != tt.want {
				t.Errorf("RequiresOpenID(%q) = %v, want %v", tt.scope, got, tt.want)
			}
		})
	}
}

func TestScopeIncludes(t *testing.T) {
	scope := "openid profile email offline_access"

	if !ScopeIncludesProfile(scope) {
		t.Error("Expected profile scope to be included")
	}

	if !ScopeIncludesEmail(scope) {
		t.Error("Expected email scope to be included")
	}

	if !ScopeIncludesOfflineAccess(scope) {
		t.Error("Expected offline_access scope to be included")
	}

	if ScopeIncludesProfile("openid email") {
		t.Error("Expected profile scope to not be included")
	}
}

func TestValidateScopeName(t *testing.T) {
	tests := []struct {
		name      string
		scopeName string
		wantValid bool
	}{
		{"valid simple name", "openid", true},
		{"valid with underscore", "offline_access", true},
		{"valid with hyphen", "read-only", true},
		{"valid with colon", "read:user", true},
		{"valid with period", "api.read", true},
		{"valid complex", "api:read:user_data", true},
		{"invalid with space", "open id", false},
		{"invalid with special char", "open@id", false},
		{"invalid with slash", "open/id", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateScopeName(tt.scopeName)
			if tt.wantValid && got != true {
				t.Errorf("ValidateScopeName(%q) = false, want true", tt.scopeName)
			}
			if !tt.wantValid && got != false {
				t.Errorf("ValidateScopeName(%q) = true, want false", tt.scopeName)
			}
		})
	}
}

func TestValidateScopeAgainstAllowed(t *testing.T) {
	tests := []struct {
		name              string
		requested         string
		allowed           []string
		wantValid         bool
		wantUnauthorized  []string
	}{
		{
			name:             "all scopes allowed",
			requested:        "openid profile",
			allowed:          []string{"openid", "profile", "email"},
			wantValid:        true,
			wantUnauthorized: nil,
		},
		{
			name:             "some scopes not allowed",
			requested:        "openid profile email",
			allowed:          []string{"openid", "profile"},
			wantValid:        false,
			wantUnauthorized: []string{"email"},
		},
		{
			name:             "no restrictions",
			requested:        "openid profile email",
			allowed:          []string{},
			wantValid:        true,
			wantUnauthorized: nil,
		},
		{
			name:             "all scopes unauthorized",
			requested:        "profile email",
			allowed:          []string{"openid"},
			wantValid:        false,
			wantUnauthorized: []string{"profile", "email"},
		},
		{
			name:             "single scope allowed",
			requested:        "openid",
			allowed:          []string{"openid", "profile"},
			wantValid:        true,
			wantUnauthorized: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotUnauthorized := ValidateScopeAgainstAllowed(tt.requested, tt.allowed)
			if gotValid != tt.wantValid {
				t.Errorf("ValidateScopeAgainstAllowed() valid = %v, want %v", gotValid, tt.wantValid)
			}
			if len(gotUnauthorized) != len(tt.wantUnauthorized) {
				t.Errorf("ValidateScopeAgainstAllowed() unauthorized = %v, want %v", gotUnauthorized, tt.wantUnauthorized)
			}
		})
	}
}

func TestValidateScopeDowngrade(t *testing.T) {
	tests := []struct {
		name      string
		requested string
		original  string
		wantErr   bool
	}{
		{
			name:      "same scopes",
			requested: "openid profile email",
			original:  "openid profile email",
			wantErr:   false,
		},
		{
			name:      "reduced scopes",
			requested: "openid profile",
			original:  "openid profile email",
			wantErr:   false,
		},
		{
			name:      "escalated scopes",
			requested: "openid profile email phone",
			original:  "openid profile email",
			wantErr:   true,
		},
		{
			name:      "empty requested uses original",
			requested: "",
			original:  "openid profile email",
			wantErr:   false,
		},
		{
			name:      "single scope downgrade",
			requested: "openid",
			original:  "openid profile",
			wantErr:   false,
		},
		{
			name:      "different scope escalation",
			requested: "openid phone",
			original:  "openid profile",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScopeDowngrade(tt.requested, tt.original)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScopeDowngrade() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeValidator_ValidateScope(t *testing.T) {
	validator := GlobalScopeValidator

	tests := []struct {
		name    string
		scope   string
		wantErr bool
	}{
		{"valid single scope", "openid", false},
		{"valid multiple scopes", "openid profile email", false},
		{"invalid scope", "invalid_scope", true},
		{"mixed valid and invalid", "openid invalid_scope", true},
		{"empty scope", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateScope(tt.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScope() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeValidator_ValidateScopeName(t *testing.T) {
	validator := GlobalScopeValidator

	tests := []struct {
		name      string
		scopeName string
		wantErr   bool
	}{
		{"valid simple name", "openid", false},
		{"valid with underscore", "offline_access", false},
		{"valid with colon", "read:user", false},
		{"invalid with space", "open id", true},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateScopeName(tt.scopeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScopeName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeValidator_ValidateScopeAgainstAllowed(t *testing.T) {
	validator := GlobalScopeValidator

	tests := []struct {
		name      string
		requested string
		allowed   []string
		wantErr   bool
	}{
		{
			name:      "all scopes allowed",
			requested: "openid profile",
			allowed:   []string{"openid", "profile", "email"},
			wantErr:   false,
		},
		{
			name:      "some scopes not allowed",
			requested: "openid profile email",
			allowed:   []string{"openid", "profile"},
			wantErr:   true,
		},
		{
			name:      "no restrictions",
			requested: "openid profile email",
			allowed:   []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateScopeAgainstAllowed(tt.requested, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScopeAgainstAllowed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
