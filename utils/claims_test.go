package utils

import (
	"oauth2-server/models"
	"testing"
)

func TestClaimFilter_FilterClaims(t *testing.T) {
	registry := models.NewScopeRegistry()
	filter := NewClaimFilter(registry)

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name          string
		scopes        string
		expectedClaims []string
		notExpectedClaims []string
	}{
		{
			name:          "openid only returns sub",
			scopes:        "openid",
			expectedClaims: []string{"sub"},
			notExpectedClaims: []string{"email", "name"},
		},
		{
			name:          "openid profile includes name",
			scopes:        "openid profile",
			expectedClaims: []string{"sub", "name"},
			notExpectedClaims: []string{"email"},
		},
		{
			name:          "openid email includes email and email_verified",
			scopes:        "openid email",
			expectedClaims: []string{"sub", "email", "email_verified"},
			notExpectedClaims: []string{"name"},
		},
		{
			name:          "openid profile email includes all",
			scopes:        "openid profile email",
			expectedClaims: []string{"sub", "name", "email", "email_verified"},
			notExpectedClaims: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := filter.FilterClaims(user, tt.scopes)

			// Check expected claims are present
			for _, claim := range tt.expectedClaims {
				if _, exists := claims[claim]; !exists {
					t.Errorf("Expected claim %s to be present", claim)
				}
			}

			// Check not expected claims are absent
			for _, claim := range tt.notExpectedClaims {
				if _, exists := claims[claim]; exists {
					t.Errorf("Expected claim %s to be absent", claim)
				}
			}

			// Verify sub is always present
			if claims["sub"] != user.ID {
				t.Errorf("Expected sub claim to be %s, got %v", user.ID, claims["sub"])
			}
		})
	}
}

func TestClaimFilter_GetIDTokenClaims(t *testing.T) {
	registry := models.NewScopeRegistry()
	filter := NewClaimFilter(registry)

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	t.Run("includes nonce when provided", func(t *testing.T) {
		nonce := "test-nonce-123"
		claims := filter.GetIDTokenClaims(user, "openid profile email", nonce)

		if claims["nonce"] != nonce {
			t.Errorf("Expected nonce to be %s, got %v", nonce, claims["nonce"])
		}
	})

	t.Run("excludes nonce when empty", func(t *testing.T) {
		claims := filter.GetIDTokenClaims(user, "openid profile email", "")

		if _, exists := claims["nonce"]; exists {
			t.Error("Expected nonce to be absent when not provided")
		}
	})

	t.Run("includes filtered claims", func(t *testing.T) {
		claims := filter.GetIDTokenClaims(user, "openid email", "")

		if claims["sub"] != user.ID {
			t.Errorf("Expected sub to be %s, got %v", user.ID, claims["sub"])
		}

		if claims["email"] != user.Email {
			t.Errorf("Expected email to be %s, got %v", user.Email, claims["email"])
		}

		if _, exists := claims["name"]; exists {
			t.Error("Expected name to be absent without profile scope")
		}
	})
}

func TestGlobalClaimFilter(t *testing.T) {
	// Ensure global instances are initialized
	if GlobalScopeRegistry == nil {
		t.Fatal("GlobalScopeRegistry is not initialized")
	}
	if GlobalClaimFilter == nil {
		t.Fatal("GlobalClaimFilter is not initialized")
	}

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	t.Run("FilterClaimsForUser helper", func(t *testing.T) {
		claims := FilterClaimsForUser(user, "openid profile email")

		if claims["sub"] != user.ID {
			t.Errorf("Expected sub to be %s, got %v", user.ID, claims["sub"])
		}

		if claims["email"] != user.Email {
			t.Errorf("Expected email to be %s, got %v", user.Email, claims["email"])
		}

		if claims["name"] != user.Name {
			t.Errorf("Expected name to be %s, got %v", user.Name, claims["name"])
		}
	})

	t.Run("GetIDTokenClaimsForUser helper", func(t *testing.T) {
		nonce := "test-nonce"
		claims := GetIDTokenClaimsForUser(user, "openid email", nonce)

		if claims["nonce"] != nonce {
			t.Errorf("Expected nonce to be %s, got %v", nonce, claims["nonce"])
		}

		if claims["email"] != user.Email {
			t.Errorf("Expected email to be %s, got %v", user.Email, claims["email"])
		}
	})
}
