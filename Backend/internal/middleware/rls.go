package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rs/zerolog/log"
)

// RLSMiddleware sets PostgreSQL session variables for Row-Level Security
type RLSMiddleware struct {
	db *sql.DB
}

// NewRLSMiddleware creates a new RLS middleware instance
func NewRLSMiddleware(db *sql.DB) *RLSMiddleware {
	return &RLSMiddleware{db: db}
}

// SetSessionContext is a middleware that sets RLS session variables for each request
func (m *RLSMiddleware) SetSessionContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context (set by auth middleware)
		userClaims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			// No authentication - let the request proceed without RLS context
			// (public endpoints will still work)
			next.ServeHTTP(w, r)
			return
		}

		// Parse tenant ID
		var tenantID uuid.UUID
		var err error
		if userClaims.TenantID != "" {
			tenantID, err = uuid.Parse(userClaims.TenantID)
			if err != nil {
				log.Warn().
					Str("tenant_id", userClaims.TenantID).
					Err(err).
					Msg("Invalid tenant ID in claims")
			}
		}

		// Check tenant status in database
		if tenantID != uuid.Nil {
			var status string
			var deletedAt *string
			err := m.db.QueryRowContext(r.Context(), "SELECT status, deleted_at::text FROM tenants WHERE id = $1", tenantID).Scan(&status, &deletedAt)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Warn().Str("tenant_id", tenantID.String()).Msg("Tenant not found during RLS check")
					http.Error(w, "Organization not found", http.StatusForbidden)
					return
				}
				log.Error().Err(err).Msg("Failed to check tenant status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if deletedAt != nil {
				http.Error(w, "Organization has been deleted", http.StatusForbidden)
				return
			}
			if status == "suspended" || status == "inactive" {
				http.Error(w, fmt.Sprintf("Organization is %s", status), http.StatusForbidden)
				return
			}
		}

		// Set session context in database
		ctx := r.Context()
		ctx, err = m.setDatabaseSessionContext(ctx, tenantID, userClaims.Role)
		if err != nil {
			log.Error().
				Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role", userClaims.Role).
				Msg("Failed to set RLS session context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Store the updated context with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// setDatabaseSessionContext sets PostgreSQL session variables for RLS
func (m *RLSMiddleware) setDatabaseSessionContext(ctx context.Context, tenantID uuid.UUID, role string) (context.Context, error) {
	// Begin a transaction to set session-local variables
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set tenant_id session variable
	if tenantID != uuid.Nil {
		_, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID.String())
		if err != nil {
			tx.Rollback()
			return ctx, fmt.Errorf("failed to set tenant_id: %w", err)
		}
	}

	// Set user_role session variable
	if role != "" {
		_, err = tx.ExecContext(ctx, "SET LOCAL app.current_user_role = $1", role)
		if err != nil {
			tx.Rollback()
			return ctx, fmt.Errorf("failed to set user_role: %w", err)
		}
	}

	// Commit the transaction (this makes the SET LOCAL effective for the session)
	if err = tx.Commit(); err != nil {
		return ctx, fmt.Errorf("failed to commit session context: %w", err)
	}

	log.Debug().
		Str("tenant_id", tenantID.String()).
		Str("role", role).
		Msg("RLS session context set successfully")

	return ctx, nil
}

// Alternative approach using connection pool and session context
// This is more efficient as it doesn't require a transaction for each request

type contextKey string

const (
	rlsContextKey contextKey = "rls_session_set"
)

// SetSessionContextEfficient sets RLS context more efficiently using database connection
func (m *RLSMiddleware) SetSessionContextEfficient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		// Parse IDs
		var tenantID, departmentID, teamID, userID uuid.UUID
		var err error

		if userClaims.UserID != "" {
			userID, err = uuid.Parse(userClaims.UserID)
			if err != nil {
				log.Warn().Err(err).Msg("Invalid user ID")
			}
		}

		if userClaims.TenantID != "" {
			tenantID, err = uuid.Parse(userClaims.TenantID)
			if err != nil {
				log.Warn().Err(err).Msg("Invalid tenant ID")
			}
		}

		if userClaims.DepartmentID != "" {
			departmentID, err = uuid.Parse(userClaims.DepartmentID)
			if err != nil {
				log.Warn().Err(err).Msg("Invalid department ID")
			}
		}

		if userClaims.TeamID != "" {
			teamID, err = uuid.Parse(userClaims.TeamID)
			if err != nil {
				log.Warn().Err(err).Msg("Invalid team ID")
			}
		}

		// Check tenant status in database to ensure it's active
		// This prevents "zombie sessions" where a user is logged in but tenant is suspended/deleted
		if tenantID != uuid.Nil {
			var status string
			var deletedAt *string // Scan as string to handle NULL/TIMESTAMP

			// We use a simple query. Note: This adds a DB round trip per request.
			// Ideally we should cache this, but for strict enforcement we check DB.
			err := m.db.QueryRowContext(r.Context(), "SELECT status, deleted_at::text FROM tenants WHERE id = $1", tenantID).Scan(&status, &deletedAt)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Warn().Str("tenant_id", tenantID.String()).Msg("Tenant not found during RLS check")
					http.Error(w, "Organization not found or access denied", http.StatusForbidden)
					return
				}
				log.Error().Err(err).Msg("Failed to check tenant status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if deletedAt != nil {
				log.Warn().Str("tenant_id", tenantID.String()).Msg("Attempt to access deleted tenant")
				http.Error(w, "Organization has been deleted", http.StatusForbidden)
				return
			}

			if status == "suspended" || status == "inactive" {
				log.Warn().Str("tenant_id", tenantID.String()).Str("status", status).Msg("Attempt to access suspended/inactive tenant")
				http.Error(w, fmt.Sprintf("Organization is %s", status), http.StatusForbidden)
				return
			}
		}

		ctx := r.Context()

		// Use updated set_session_context function with all parameters
		query := `SELECT set_session_context($1, $2, $3, $4, $5)`
		_, err = m.db.ExecContext(ctx, query, userID, tenantID, userClaims.Role, departmentID, teamID)
		if err != nil {
			// Ignore client disconnects/cancellations
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			log.Error().
				Err(err).
				Str("user_id", userID.String()).
				Str("tenant_id", tenantID.String()).
				Str("role", userClaims.Role).
				Str("department_id", departmentID.String()).
				Str("team_id", teamID.String()).
				Msg("Failed to set RLS session context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Mark that RLS context has been set
		ctx = context.WithValue(ctx, rlsContextKey, true)

		log.Debug().
			Str("user_id", userID.String()).
			Str("tenant_id", tenantID.String()).
			Str("role", userClaims.Role).
			Str("department_id", departmentID.String()).
			Str("team_id", teamID.String()).
			Msg("RLS session context set")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// BypassRLSForSuperAdmin allows super admins to bypass RLS when needed
// This should be used sparingly and only for specific admin operations
func (m *RLSMiddleware) BypassRLSForSuperAdmin(ctx context.Context) error {
	userClaims, ok := auth.GetClaimsFromContext(ctx)
	if !ok || userClaims.Role != "super_admin" {
		return fmt.Errorf("only super admins can bypass RLS")
	}

	// Clear the tenant restriction for super admin
	_, err := m.db.ExecContext(ctx, "SET LOCAL app.current_tenant_id = NULL")
	if err != nil {
		return fmt.Errorf("failed to bypass RLS: %w", err)
	}

	return nil
}

// GetSessionContext retrieves the current RLS session context
func (m *RLSMiddleware) GetSessionContext(ctx context.Context) (tenantID uuid.UUID, role string, err error) {
	var tenantIDStr string

	err = m.db.QueryRowContext(ctx, "SELECT current_tenant_id(), current_user_role()").
		Scan(&tenantIDStr, &role)

	if err != nil {
		return uuid.Nil, "", err
	}

	if tenantIDStr != "" {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, "", err
		}
	}

	return tenantID, role, nil
}
