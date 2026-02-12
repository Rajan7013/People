package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type OrganizationService struct {
	db                  *sql.DB
	subscriptionService *SubscriptionService
	pepperSecret        string
}

func NewOrganizationService(db *sql.DB, subscriptionService *SubscriptionService, pepperSecret string) *OrganizationService {
	return &OrganizationService{
		db:                  db,
		subscriptionService: subscriptionService,
		pepperSecret:        pepperSecret,
	}
}

// CreateOrganizationRequest represents the data needed to create an organization
type CreateOrganizationRequest struct {
	// Organization details
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
	Country   string `json:"country"`

	// Admin details
	AdminName     string `json:"admin_name"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`

	// Address
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`

	// Business details
	ContactNumber string `json:"contact_number"`
	Website       string `json:"website"`
	Currency      string `json:"currency"`
	Timezone      string `json:"timezone"`

	// Subscription details
	PlanID        uuid.UUID `json:"plan_id"`
	BillingCycle  string    `json:"billing_cycle"`  // monthly, yearly
	TrialDuration int       `json:"trial_duration"` // days
}

// slugify converts a string to a slug
func slugify(s string) string {
	s = strings.ToLower(s)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// CreateOrganization creates a new organization with admin user and subscription
func (s *OrganizationService) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*models.Tenant, error) {
	// Auto-generate subdomain if empty
	if req.Subdomain == "" {
		req.Subdomain = slugify(req.Name)
		// Append random string to ensure uniqueness if needed, but let's start simple
		// Ideally we should check for uniqueness, but database constraint will catch it
	}

	// Ensure email is lowercase for consistent auth matching
	req.AdminEmail = strings.ToLower(req.AdminEmail)

	// 1. Check for existing USER first (by email) - this is the most reliable anchor
	var existingUserID string
	var existingUserTenantID string
	var existingUserDeletedAt *time.Time
	var existingTenantID string
	var existingTenantDeletedAt *time.Time

	// Check if user exists (active or deleted)
	err := s.db.QueryRowContext(ctx, "SELECT id, tenant_id, deleted_at FROM users WHERE email = $1", req.AdminEmail).Scan(&existingUserID, &existingUserTenantID, &existingUserDeletedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// If user exists, they MUST belong to a tenant. Let's find that tenant's status.
	if existingUserID != "" {
		// Find the tenant linked to this user
		err = s.db.QueryRowContext(ctx, "SELECT id, deleted_at FROM tenants WHERE id = $1", existingUserTenantID).Scan(&existingTenantID, &existingTenantDeletedAt)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check linked tenant: %w", err)
		}
	} else {
		// Fallback: Check tenant by admin_email directly (legacy check)
		err = s.db.QueryRowContext(ctx, "SELECT id, deleted_at FROM tenants WHERE admin_email = $1", req.AdminEmail).Scan(&existingTenantID, &existingTenantDeletedAt)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check existing tenant by email: %w", err)
		}
	}

	// SCENARIO 1: Tenant exists and is ACTIVE -> Conflict (already registered)
	if existingTenantID != "" && existingTenantDeletedAt == nil {
		return nil, fmt.Errorf("organization with this admin email already exists")
	}

	// SCENARIO 2: User exists and is ACTIVE, but Tenant is missing/deleted?
	// If user is active but tenant is deleted, it's an "Orphaned Active User". We should treat this as a revival case.
	if existingUserID != "" && existingUserDeletedAt == nil {
		if existingTenantID != "" && existingTenantDeletedAt != nil {
			// This is the "Revival" case. Proceed.
		} else if existingTenantID == "" {
			// User exists but has no valid tenant? This is a data corruption or "multi-tenant user" case (not supported yet).
			// For now, fail safe.
			return nil, fmt.Errorf("user already exists but organization is missing")
		} else {
			// User active, Tenant active (handled by Scenario 1)
			return nil, fmt.Errorf("user already exists")
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var tenantID uuid.UUID

	if existingTenantID != "" && existingTenantDeletedAt != nil {
		// Revive Constraint: We must reuse the ID to "revive" it, OR delete the old one permanently if we want a fresh start.
		// Reviving is safer for data referential integrity if we kept old data.
		// But user might want a "fresh" account.
		// If we revive, we keep old data. If we hard delete, we lose it.
		// User said "delete and try to make same things it is not happeing". They expect it to work.
		// "Revive" is the most robust way to "make it happen" without unique constraint errors on subdomain.

		tenantID = uuid.MustParse(existingTenantID)
		reviveQuery := `
			UPDATE tenants 
			SET status = 'active', 
				deleted_at = NULL, 
				name = $1, 
				subdomain = $2, 
				domain = $3, 
				country = $4, 
				updated_at = $5 
			WHERE id = $6`

		_, err := tx.ExecContext(ctx, reviveQuery,
			req.Name, req.Subdomain, req.Domain, req.Country, time.Now(), tenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to revive tenant: %w", err)
		}

		// Update org details
		// We can't really "Insert" again if it exists. We should "Upsert" or Update.
		// Let's check if details exist.
		var detailCount int
		tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM organization_details WHERE tenant_id = $1", tenantID).Scan(&detailCount)

		if detailCount > 0 {
			updateDetailQuery := `
				UPDATE organization_details 
				SET admin_name = $1, contact_number = $2, website = $3,
					address_line1 = $4, address_line2 = $5, city = $6, state = $7, country = $8, postal_code = $9,
					currency = $10, timezone = $11, updated_at = $12
				WHERE tenant_id = $13`

			currency := req.Currency
			if currency == "" {
				currency = "USD"
			}
			timezone := req.Timezone
			if timezone == "" {
				timezone = "UTC"
			}

			_, err = tx.ExecContext(ctx, updateDetailQuery,
				req.AdminName, req.ContactNumber, req.Website,
				req.AddressLine1, req.AddressLine2, req.City, req.State, req.Country, req.PostalCode,
				currency, timezone, time.Now(), tenantID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update organization details: %w", err)
			}
		} else {
			// Insert details if missing (rare case)
			orgDetailID := uuid.New()
			orgDetailQuery := `
				INSERT INTO organization_details (
					id, tenant_id, admin_name, admin_email, contact_number, website,
					address_line1, address_line2, city, state, country, postal_code,
					currency, timezone, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

			currency := req.Currency
			if currency == "" {
				currency = "USD"
			}
			timezone := req.Timezone
			if timezone == "" {
				timezone = "UTC"
			}

			_, err = tx.ExecContext(ctx, orgDetailQuery,
				orgDetailID, tenantID, req.AdminName, req.AdminEmail, req.ContactNumber, req.Website,
				req.AddressLine1, req.AddressLine2, req.City, req.State, req.Country, req.PostalCode,
				currency, timezone, time.Now(), time.Now(),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create organization details: %w", err)
			}
		}

	} else {
		// 1. Create NEW tenant
		tenantID = uuid.New()
		tenantQuery := `
			INSERT INTO tenants (id, name, subdomain, domain, status, country, admin_email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`

		err = tx.QueryRowContext(ctx, tenantQuery,
			tenantID, req.Name, req.Subdomain, req.Domain, "active", req.Country, req.AdminEmail,
			time.Now(), time.Now(),
		).Scan(&tenantID)

		if err != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		// 2. Create organization details
		orgDetailID := uuid.UUID{}
		orgDetailQuery := `
			INSERT INTO organization_details (
				id, tenant_id, admin_name, admin_email, contact_number, website,
				address_line1, address_line2, city, state, country, postal_code,
				currency, timezone, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

		currency := req.Currency
		if currency == "" {
			currency = "USD"
		}
		timezone := req.Timezone
		if timezone == "" {
			timezone = "UTC"
		}

		orgDetailID = uuid.New()
		_, err = tx.ExecContext(ctx, orgDetailQuery,
			orgDetailID, tenantID, req.AdminName, req.AdminEmail, req.ContactNumber, req.Website,
			req.AddressLine1, req.AddressLine2, req.City, req.State, req.Country, req.PostalCode,
			currency, timezone, time.Now(), time.Now(),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create organization details: %w", err)
		}
	}

	// 3. Create admin user
	// 3. Create OR Update admin user
	userID := uuid.New()
	if existingUserID != "" {
		userID = uuid.MustParse(existingUserID)
	}

	var hashedPassword string
	if req.AdminPassword != "" {
		hashedPassword, err = auth.HashPassword(req.AdminPassword, s.pepperSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	if existingUserID != "" {
		// Update existing user (whether active orphaned, or soft-deleted)
		// We revive functionality here implicitly by setting deleted_at = NULL
		userUpdateQuery := `
			UPDATE users 
			SET tenant_id = $1, 
				password_hash = COALESCE(NULLIF($2, ''), password_hash), 
				role = 'admin', 
				first_name = $3, 
				last_name = '', 
				is_active = true, 
				deleted_at = NULL, 
				updated_at = $4 
			WHERE id = $5`

		_, err = tx.ExecContext(ctx, userUpdateQuery,
			tenantID, string(hashedPassword), req.AdminName, time.Now(), userID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update admin user: %w", err)
		}
	} else {
		// Create new user
		userQuery := `
			INSERT INTO users (id, tenant_id, email, password_hash, role, first_name, last_name, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err = tx.ExecContext(ctx, userQuery,
			userID, tenantID, req.AdminEmail, string(hashedPassword), "admin",
			req.AdminName, "", true, time.Now(), time.Now(),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	}

	// 4. Get plan details for subscription
	var planAmount float64
	var planYearlyPrice float64
	var planCurrency string
	planQuery := `SELECT price_monthly, price_yearly, currency FROM subscription_plans WHERE id = $1`

	err = tx.QueryRowContext(ctx, planQuery, req.PlanID).Scan(&planAmount, &planYearlyPrice, &planCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan details: %w", err)
	}

	// Determine amount based on billing cycle
	amount := planAmount
	if req.BillingCycle == "yearly" {
		amount = planYearlyPrice
	}

	// 5. Create subscription
	subscriptionID := uuid.New()
	currentPeriodStart := time.Now()
	var currentPeriodEnd time.Time
	var trialEndsAt *time.Time

	if req.TrialDuration > 0 {
		trialEnd := currentPeriodStart.AddDate(0, 0, req.TrialDuration)
		trialEndsAt = &trialEnd
		currentPeriodEnd = trialEnd
	} else {
		if req.BillingCycle == "monthly" {
			currentPeriodEnd = currentPeriodStart.AddDate(0, 1, 0)
		} else {
			currentPeriodEnd = currentPeriodStart.AddDate(1, 0, 0)
		}
	}

	subscriptionQuery := `
		INSERT INTO subscriptions (
			id, tenant_id, plan_id, status, billing_cycle, amount, currency,
			trial_ends_at, current_period_start, current_period_end, auto_renew,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tenant_id) DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			status = EXCLUDED.status,
			billing_cycle = EXCLUDED.billing_cycle,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			trial_ends_at = EXCLUDED.trial_ends_at,
			current_period_start = EXCLUDED.current_period_start,
			current_period_end = EXCLUDED.current_period_end,
			auto_renew = EXCLUDED.auto_renew,
			updated_at = EXCLUDED.updated_at`

	status := "active"
	if req.TrialDuration > 0 {
		status = "trial"
	}

	_, err = tx.ExecContext(ctx, subscriptionQuery,
		subscriptionID, tenantID, req.PlanID, status, req.BillingCycle, amount, planCurrency,
		trialEndsAt, currentPeriodStart, currentPeriodEnd, true, time.Now(), time.Now(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created organization
	return s.GetOrganizationByID(ctx, tenantID)
}

// GetOrganizationByID retrieves an organization with all details
func (s *OrganizationService) GetOrganizationByID(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT 
			t.id, t.name, t.subdomain, t.domain, t.status, t.country, t.admin_email,
			t.storage_used_mb, t.api_requests_count, t.last_activity_at, t.settings,
			t.created_at, t.updated_at,
			od.admin_name, od.contact_number, od.website, od.address_line1, od.address_line2,
			od.city, od.state, od.postal_code, od.currency, od.timezone,
			s.id as sub_id, s.plan_id, s.status as sub_status, s.billing_cycle, s.amount,
			s.current_period_end, sp.display_name as plan_name,
			(SELECT COUNT(*) FROM users WHERE tenant_id = t.id AND deleted_at IS NULL) as total_users,
			(SELECT COUNT(*) FROM employees WHERE tenant_id = t.id AND employment_status = 'active') as active_employees
		FROM tenants t
		LEFT JOIN organization_details od ON t.id = od.tenant_id
		LEFT JOIN subscriptions s ON t.id = s.tenant_id
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE t.id = $1 AND t.deleted_at IS NULL`

	tenant := &models.Tenant{
		OrganizationDetail: &models.OrganizationDetail{},
		Subscription:       &models.Subscription{Plan: &models.SubscriptionPlan{}},
	}

	var subID, planID *uuid.UUID
	var subStatus, billingCycle, planName *string
	var amount *float64
	var periodEnd *time.Time
	var settingsBytes []byte

	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Subdomain, &tenant.Domain, &tenant.Status,
		&tenant.Country, &tenant.AdminEmail, &tenant.StorageUsedMB, &tenant.APIRequestsCount,
		&tenant.LastActivityAt, &settingsBytes, &tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.OrganizationDetail.AdminName, &tenant.OrganizationDetail.ContactNumber,
		&tenant.OrganizationDetail.Website, &tenant.OrganizationDetail.AddressLine1,
		&tenant.OrganizationDetail.AddressLine2, &tenant.OrganizationDetail.City,
		&tenant.OrganizationDetail.State, &tenant.OrganizationDetail.PostalCode,
		&tenant.OrganizationDetail.Currency, &tenant.OrganizationDetail.Timezone,
		&subID, &planID, &subStatus, &billingCycle, &amount, &periodEnd, &planName,
		&tenant.TotalUsers, &tenant.ActiveEmployees,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if settingsBytes != nil {
		if err := json.Unmarshal(settingsBytes, &tenant.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	// Populate subscription if exists
	if subID != nil {
		tenant.Subscription.ID = *subID
		tenant.Subscription.PlanID = *planID
		tenant.Subscription.Status = *subStatus
		tenant.Subscription.BillingCycle = *billingCycle
		tenant.Subscription.Amount = *amount
		tenant.Subscription.CurrentPeriodEnd = *periodEnd
		tenant.Subscription.Plan.DisplayName = *planName
	}

	return tenant, nil
}

// GetAllOrganizations retrieves all organizations with filters
func (s *OrganizationService) GetAllOrganizations(ctx context.Context, filters map[string]interface{}) ([]*models.Tenant, error) {
	query := `
		SELECT 
			t.id, t.name, t.subdomain, t.domain, t.status, t.country, t.admin_email,
			t.storage_used_mb, t.api_requests_count, t.created_at,
			s.status as sub_status, sp.display_name as plan_name,
			(SELECT COUNT(*) FROM users WHERE tenant_id = t.id AND deleted_at IS NULL) as total_users
		FROM tenants t
		LEFT JOIN subscriptions s ON t.id = s.tenant_id
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE t.deleted_at IS NULL AND t.status != 'deleted'`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if country, ok := filters["country"].(string); ok && country != "" {
		query += fmt.Sprintf(" AND t.country = $%d", argIndex)
		args = append(args, country)
		argIndex++
	}

	if planID, ok := filters["plan_id"].(string); ok && planID != "" {
		query += fmt.Sprintf(" AND s.plan_id = $%d", argIndex)
		args = append(args, planID)
		argIndex++
	}

	query += " ORDER BY t.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{
			Subscription: &models.Subscription{Plan: &models.SubscriptionPlan{}},
		}

		var subStatus, planName *string

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Subdomain, &tenant.Domain, &tenant.Status,
			&tenant.Country, &tenant.AdminEmail, &tenant.StorageUsedMB, &tenant.APIRequestsCount,
			&tenant.CreatedAt, &subStatus, &planName, &tenant.TotalUsers,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		if subStatus != nil {
			tenant.Subscription.Status = *subStatus
		}
		if planName != nil {
			tenant.Subscription.Plan.DisplayName = *planName
		}

		tenants = append(tenants, tenant)
		fmt.Printf("[DEBUG] GetAllOrganizations: Found Tenant ID=%s Name='%s' Status=%s CreatedAt=%v\n", tenant.ID, tenant.Name, tenant.Status, tenant.CreatedAt)
	}

	return tenants, nil
}

// BlockOrganization blocks/suspends an organization
func (s *OrganizationService) BlockOrganization(ctx context.Context, tenantID uuid.UUID) error {
	query := `UPDATE tenants SET status = 'suspended', updated_at = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to block organization: %w", err)
	}
	return nil
}

// UnblockOrganization unblocks/activates an organization
func (s *OrganizationService) UnblockOrganization(ctx context.Context, tenantID uuid.UUID) error {
	query := `UPDATE tenants SET status = 'active', updated_at = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to unblock organization: %w", err)
	}
	return nil
}

// DeleteOrganization hard deletes an organization and everything cascading from it
func (s *OrganizationService) DeleteOrganization(ctx context.Context, tenantID uuid.UUID) error {
	// Robust Hard Delete:
	// We manually delete dependent tables in reverse order of dependency to ensure
	// the deletion succeeds even if "ON DELETE CASCADE" is missing from some tables in the actual DB schema.

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Delete High-Level Tenant Logs & Settings
	// These are generally independent or link only to Tenant/User
	if _, err := tx.ExecContext(ctx, "DELETE FROM audit_logs WHERE tenant_id = $1", tenantID); err != nil {
		fmt.Printf("Error deleting audit_logs: %v\n", err) // Non-fatal, but good to know
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM api_request_logs WHERE tenant_id = $1", tenantID); err != nil {
		// Ignore if table missing
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM usage_metrics WHERE tenant_id = $1", tenantID); err != nil {
		// Ignore
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM system_settings WHERE tenant_id = $1", tenantID); err != nil {
		// Ignore
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM organization_details WHERE tenant_id = $1", tenantID); err != nil {
		// Ignore
	}

	// 2. Delete User Links (Must be before Users)
	// Tables: user_themes, user_preferences, security_settings, user_profiles
	userTables := []string{"user_themes", "user_preferences", "security_settings", "user_profiles"}
	for _, table := range userTables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1", table), tenantID); err != nil {
			fmt.Printf("Error deleting %s: %v\n", table, err)
		}
	}

	// 3. Delete Operational Data (Attendance, Biometrics)
	// Order: attendance_records -> biometric_attendance_logs -> biometric_devices
	if _, err := tx.ExecContext(ctx, "DELETE FROM attendance_records WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete attendance_records: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM biometric_attendance_logs WHERE tenant_id = $1", tenantID); err != nil {
		fmt.Printf("Error deleting biometric_attendance_logs: %v\n", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM biometric_devices WHERE tenant_id = $1", tenantID); err != nil {
		fmt.Printf("Error deleting biometric_devices: %v\n", err)
	}

	// 4. Delete Leave Data (Must be before Employees)
	// We delete from BOTH tables to ensure cleanup across schema versions (migrated vs fresh).
	// We use ExecContext and ignore errors because one of these tables might not exist.
	tx.ExecContext(ctx, "DELETE FROM leave_requests WHERE tenant_id = $1", tenantID)
	tx.ExecContext(ctx, "DELETE FROM leave_applications WHERE tenant_id = $1", tenantID)

	if _, err := tx.ExecContext(ctx, "DELETE FROM leave_balances WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete leave_balances: %w", err)
	}

	// Delete Leave Types & Policies (Referenced by Balances/Requests, Reference Tenant)
	// These must be deleted AFTER balances, but BEFORE tenant.
	tx.ExecContext(ctx, "DELETE FROM leave_types WHERE tenant_id = $1", tenantID)
	tx.ExecContext(ctx, "DELETE FROM attendance_policies WHERE tenant_id = $1", tenantID)

	// 5. Delete Employees (Refer Users and Departments)
	if _, err := tx.ExecContext(ctx, "DELETE FROM employees WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete employees: %w", err)
	}

	// 6. Delete Departments
	// Break self-reference first
	if _, err := tx.ExecContext(ctx, "UPDATE departments SET parent_id = NULL WHERE tenant_id = $1", tenantID); err != nil {
		// Ignore error
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM departments WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete departments: %w", err)
	}

	// 7. Delete Users
	if _, err := tx.ExecContext(ctx, "DELETE FROM users WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}

	// 8. Delete Billing Data
	if _, err := tx.ExecContext(ctx, "DELETE FROM invoices WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete invoices: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM subscriptions WHERE tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("failed to delete subscriptions: %w", err)
	}

	// 9. Delete Tenant
	fmt.Printf("[DEBUG] DeleteOrganization: Attempting to delete tenant ID=%s\n", tenantID)

	// Check if it exists before trying to delete (inside transaction)
	var count int
	tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tenants WHERE id = $1", tenantID).Scan(&count)
	fmt.Printf("[DEBUG] DeleteOrganization: Tenant exists count=%d\n", count)

	query := `DELETE FROM tenants WHERE id = $1`
	result, err := tx.ExecContext(ctx, query, tenantID)
	if err != nil {
		fmt.Printf("Error deleting tenant: %v\n", err)
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("[DEBUG] DeleteOrganization: Rows affected=%d\n", rowsAffected)

	if rowsAffected == 0 {
		// If count was > 0 but rowsAffected is 0, something is blocking it silently?
		return fmt.Errorf("organization not found")
	}

	return tx.Commit()
}

// UpdateOrganization updates organization details
func (s *OrganizationService) UpdateOrganization(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*models.Tenant, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update tenants table
	tenantFields := map[string]bool{"name": true, "subdomain": true, "domain": true, "country": true, "admin_email": true, "status": true}
	tenantUpdates := make(map[string]interface{})
	orgDetailUpdates := make(map[string]interface{})

	for field, value := range updates {
		if tenantFields[field] {
			tenantUpdates[field] = value
		} else {
			orgDetailUpdates[field] = value
		}
	}

	// Update tenant if needed
	if len(tenantUpdates) > 0 {
		setParts := []string{}
		args := []interface{}{}
		argIndex := 1

		for field, value := range tenantUpdates {
			if field == "settings" {
				settingsJSON, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal settings: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, settingsJSON)
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, value)
			}
			argIndex++
		}

		setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
		args = append(args, time.Now())
		argIndex++

		args = append(args, tenantID)

		query := fmt.Sprintf("UPDATE tenants SET %s WHERE id = $%d", joinStringsOrg(setParts, ", "), argIndex)
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update tenant: %w", err)
		}
	}

	// Update organization_details if needed
	if len(orgDetailUpdates) > 0 {
		setParts := []string{}
		args := []interface{}{}
		argIndex := 1

		for field, value := range orgDetailUpdates {
			if field == "metadata" {
				metadataJSON, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, metadataJSON)
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, value)
			}
			argIndex++
		}

		setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
		args = append(args, time.Now())
		argIndex++

		args = append(args, tenantID)

		query := fmt.Sprintf("UPDATE organization_details SET %s WHERE tenant_id = $%d", joinStringsOrg(setParts, ", "), argIndex)
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update organization details: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetOrganizationByID(ctx, tenantID)
}

// Helper function to join strings
func joinStringsOrg(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// OrganizationProfile represents the organization profile data for admin view
type OrganizationProfile struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"` // Read-only (admin email)
	Phone          string     `json:"phone"`
	Address        string     `json:"address"`
	Industry       string     `json:"industry"`
	Size           string     `json:"size"`
	LogoURL        string     `json:"logo_url"`
	CreatedAt      time.Time  `json:"created_at"` // Read-only
	SubscriptionID *uuid.UUID `json:"subscription_id"`
	PlanName       string     `json:"plan_name"`
	PlanStatus     string     `json:"plan_status"`
	BillingCycle   string     `json:"billing_cycle"`
}

// UpdateOrganizationProfileRequest represents the request to update organization profile
type UpdateOrganizationProfileRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	LogoURL string `json:"logo_url"`
}

// GetOrganizationProfile retrieves the organization profile for admin view
func (s *OrganizationService) GetOrganizationProfile(tenantID uuid.UUID) (*OrganizationProfile, error) {
	var profile OrganizationProfile

	query := `
		SELECT 
			t.id, t.name, t.admin_email, 
			COALESCE(od.contact_number, '') as phone,
			COALESCE(od.address_line1 || ' ' || od.address_line2 || ' ' || od.city || ' ' || od.state || ' ' || od.postal_code, '') as address,
			COALESCE(t.country, '') as industry,
			'' as size,
			'' as logo_url,
			t.created_at,
			s.id as subscription_id,
			COALESCE(sp.display_name, '') as plan_name,
			COALESCE(s.status, '') as plan_status,
			COALESCE(s.billing_cycle, '') as billing_cycle
		FROM tenants t
		LEFT JOIN organization_details od ON t.id = od.tenant_id
		LEFT JOIN subscriptions s ON t.id = s.tenant_id
		LEFT JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE t.id = $1 AND t.deleted_at IS NULL
	`

	err := s.db.QueryRow(query, tenantID).Scan(
		&profile.ID,
		&profile.Name,
		&profile.Email,
		&profile.Phone,
		&profile.Address,
		&profile.Industry,
		&profile.Size,
		&profile.LogoURL,
		&profile.CreatedAt,
		&profile.SubscriptionID,
		&profile.PlanName,
		&profile.PlanStatus,
		&profile.BillingCycle,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get organization profile: %w", err)
	}

	return &profile, nil
}

// UpdateOrganizationProfile updates the organization profile (admin-editable fields only)
func (s *OrganizationService) UpdateOrganizationProfile(tenantID uuid.UUID, req UpdateOrganizationProfileRequest) (*OrganizationProfile, error) {
	// Update only allowed fields in tenants table
	query := `
		UPDATE tenants
		SET 
			name = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := s.db.Exec(query, req.Name, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization name: %w", err)
	}

	// Update organization_details for phone and address
	detailQuery := `
		UPDATE organization_details
		SET 
			contact_number = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = $2
	`

	_, err = s.db.Exec(detailQuery, req.Phone, tenantID)
	if err != nil {
		// If organization_details doesn't exist, that's okay
		fmt.Printf("Note: Could not update organization_details: %v\n", err)
	}

	// Return updated profile
	return s.GetOrganizationProfile(tenantID)
}
