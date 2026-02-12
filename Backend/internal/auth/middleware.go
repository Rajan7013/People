package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type contextKey string

const (
	UserContextKey   contextKey = "user"
	TenantContextKey contextKey = "tenant"
	ClaimsContextKey contextKey = "claims"
)

// Middleware creates an authentication middleware
func (s *Service) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from cookie first, then fall back to Authorization header
			var token string

			// Log all cookies for debugging
			cookies := r.Cookies()
			log.Debug().Int("cookie_count", len(cookies)).Msg("Received cookies")
			for _, c := range cookies {
				preview := c.Value
				if len(preview) > 20 {
					preview = preview[:20] + "..."
				}
				log.Debug().Str("name", c.Name).Str("value_preview", preview).Msg("Cookie")
			}

			// Try to get token from HttpOnly cookie
			cookie, err := r.Cookie("access_token")
			if err == nil && cookie.Value != "" {
				token = cookie.Value
				log.Info().Msg("✅ Token found in HttpOnly cookie")
			} else {
				log.Debug().Err(err).Msg("No access_token cookie found")
				// Fall back to Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					log.Warn().Msg("Missing both cookie and Authorization header")
					s.writeErrorResponse(w, http.StatusUnauthorized, "Missing authorization")
					return
				}

				// Check Bearer token format
				tokenParts := strings.Split(authHeader, " ")
				if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
					s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format")
					return
				}

				token = tokenParts[1]
				log.Debug().Msg("Token found in Authorization header")
			}

			// Validate token
			claims, err := s.ValidateToken(token)
			if err != nil {
				log.Warn().Err(err).Msg("Invalid token")
				s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Check if user and tenant are still active
			if err := s.CheckUserStatus(r.Context(), claims.UserID); err != nil {
				log.Warn().Err(err).Str("user_id", claims.UserID).Msg("User or tenant inactive/deleted")
				// Using 403 Forbidden to distinguish from invalid token (401)
				// The frontend will intercept this and clear the session
				s.writeErrorResponse(w, http.StatusForbidden, err.Error())
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			ctx = context.WithValue(ctx, UserContextKey, claims.UserID)
			ctx = context.WithValue(ctx, TenantContextKey, claims.TenantID)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that requires a specific role level
// Users with a higher role level also have access (e.g., Admin can access Manager routes)
func (s *Service) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
			if !ok {
				s.writeErrorResponse(w, http.StatusUnauthorized, "Missing authentication")
				return
			}

			// Check if user's role satisfies the required role level
			if !HasRole(claims.Role, requiredRole) {
				s.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireTenantAccess middleware ensures user can only access their tenant's data
func (s *Service) RequireTenantAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
			if !ok {
				s.writeErrorResponse(w, http.StatusUnauthorized, "Missing authentication")
				return
			}

			// Get tenant ID from URL path (e.g., /api/v1/{tenantID}/employees)
			// This is a simplified implementation - in production, you might want more robust parsing
			pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(pathSegments) >= 3 {
				urlTenantID := pathSegments[2] // /api/v1/{tenantID}/...

				// Super admins can access any tenant
				if claims.Role != "super_admin" && claims.TenantID != urlTenantID {
					s.writeErrorResponse(w, http.StatusForbidden, "Access denied to tenant data")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user ID from request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserContextKey).(string)
	return userID, ok
}

// GetTenantFromContext extracts tenant ID from request context
func GetTenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantContextKey).(string)
	return tenantID, ok
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// writeErrorResponse writes a JSON error response
func (s *Service) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
