package middleware

import (
	"database/sql"
	"net/http"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rs/zerolog/log"
)

// CheckUserStatus verifies that the user's account is active
// This middleware queries the database on every request to ensure immediate enforcement
// of account suspension/activation changes
func CheckUserStatus(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.GetClaimsFromContext(r.Context())
			if !ok {
				// No claims means user is not authenticated
				// Let the auth middleware handle this
				next.ServeHTTP(w, r)
				return
			}

			// Query database for user's active status
			var isActive bool
			err := db.QueryRow(`
				SELECT is_active 
				FROM users 
				WHERE id = $1 AND deleted_at IS NULL
			`, claims.UserID).Scan(&isActive)

			if err != nil {
				if err == sql.ErrNoRows {
					log.Warn().
						Str("user_id", claims.UserID).
						Msg("User not found in database")
					http.Error(w, "Account not found", http.StatusForbidden)
					return
				}
				log.Error().
					Err(err).
					Str("user_id", claims.UserID).
					Msg("Failed to check user status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !isActive {
				log.Warn().
					Str("user_id", claims.UserID).
					Msg("Inactive user attempted to access system")
				http.Error(w, "Account suspended", http.StatusForbidden)
				return
			}

			// User is active, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates a middleware that requires one of the specified roles
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.GetClaimsFromContext(r.Context())
			if !ok {
				log.Warn().Msg("No claims found in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				log.Warn().
					Str("user_role", claims.Role).
					Strs("allowed_roles", allowedRoles).
					Msg("User role not authorized for this endpoint")
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			log.Debug().
				Str("user_id", claims.UserID).
				Str("role", claims.Role).
				Msg("Role authorization successful")

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSuperAdmin middleware ensures only super_admin can access
func RequireSuperAdmin(next http.Handler) http.Handler {
	return RequireRole("super_admin")(next)
}

// RequireOrgAdmin middleware ensures only org admin can access
func RequireOrgAdmin(next http.Handler) http.Handler {
	return RequireRole("admin")(next)
}

// RequireManager middleware ensures manager or admin can access
func RequireManager(next http.Handler) http.Handler {
	return RequireRole("manager", "admin")(next)
}

// RequireHR middleware ensures HR or admin can access
func RequireHR(next http.Handler) http.Handler {
	return RequireRole("hr", "admin")(next)
}

// RequireTeamLead middleware ensures team lead, manager, or admin can access
func RequireTeamLead(next http.Handler) http.Handler {
	return RequireRole("team_lead", "manager", "admin")(next)
}

// BlockSuperAdminFromCompanyData prevents super admin from accessing company-level endpoints
// Super admin should only access platform-level data (organizations, plans, invoices)
func BlockSuperAdminFromCompanyData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			log.Warn().Msg("No claims found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims.Role == "super_admin" {
			log.Warn().
				Str("user_id", claims.UserID).
				Msg("Super admin attempted to access company data")
			http.Error(w, "Forbidden: super admin cannot access company data", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAnyRole middleware ensures user has at least one valid role (authenticated)
func RequireAnyRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			log.Warn().Msg("No claims found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims.Role == "" {
			log.Warn().
				Str("user_id", claims.UserID).
				Msg("User has no role assigned")
			http.Error(w, "Forbidden: no role assigned", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
