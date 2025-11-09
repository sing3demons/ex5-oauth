package middleware

import (
	"context"
	"net/http"
	"oauth2-server/repository"
	"time"
)

const (
	// SSOCookieName is the name of the SSO session cookie
	SSOCookieName = "oauth_sso_session"
	
	// SSOSessionContextKey is the context key for storing SSO session
	SSOSessionContextKey = "sso_session"
)

// SSOMiddleware creates middleware that validates SSO sessions from cookies
// and adds them to the request context for downstream handlers
func SSOMiddleware(ssoRepo *repository.SSOSessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract SSO cookie from request
			cookie, err := r.Cookie(SSOCookieName)
			if err == nil && cookie.Value != "" {
				// Validate session against database
				session, err := ssoRepo.FindBySessionID(r.Context(), cookie.Value)
				if err == nil && session != nil {
					// Check if session is authenticated and not expired
					if session.Authenticated && session.ExpiresAt.After(time.Now()) {
						// Update last activity timestamp on valid session
						_ = ssoRepo.UpdateLastActivity(r.Context(), session.SessionID)
						
						// Add session to request context for downstream handlers
						ctx := context.WithValue(r.Context(), SSOSessionContextKey, session)
						r = r.WithContext(ctx)
					}
				}
			}
			
			// Continue to next handler (non-blocking even if session validation fails)
			next.ServeHTTP(w, r)
		})
	}
}
