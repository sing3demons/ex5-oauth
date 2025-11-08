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
