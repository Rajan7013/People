package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type SubscriptionService struct {
	db *sql.DB
}

func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// GetAllPlans retrieves all subscription plans
func (s *SubscriptionService) GetAllPlans(ctx context.Context, includeInactive bool) ([]*models.SubscriptionPlan, error) {
	query := `
		SELECT id, name, display_name, description, price_monthly, price_yearly, currency,
			max_users, max_storage_gb, max_api_requests_monthly, max_departments,
			features, is_active, is_visible, sort_order, created_at, updated_at
		FROM subscription_plans`

	if !includeInactive {
		query += ` WHERE is_active = true`
	}

	query += ` ORDER BY sort_order, name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []*models.SubscriptionPlan
	for rows.Next() {
		plan := &models.SubscriptionPlan{}
		var featuresBytes []byte

		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.DisplayName, &plan.Description,
			&plan.PriceMonthly, &plan.PriceYearly, &plan.Currency,
			&plan.MaxUsers, &plan.MaxStorageGB, &plan.MaxAPIRequestsMonthly, &plan.MaxDepartments,
			&featuresBytes, &plan.IsActive, &plan.IsVisible, &plan.SortOrder,
			&plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		if featuresBytes != nil {
			if err := json.Unmarshal(featuresBytes, &plan.Features); err != nil {
				return nil, fmt.Errorf("failed to unmarshal features: %w", err)
			}
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

// GetPlanByID retrieves a specific plan by ID
func (s *SubscriptionService) GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, name, display_name, description, price_monthly, price_yearly, currency,
			max_users, max_storage_gb, max_api_requests_monthly, max_departments,
			features, is_active, is_visible, sort_order, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1`

	plan := &models.SubscriptionPlan{}
	var featuresBytes []byte

	err := s.db.QueryRowContext(ctx, query, planID).Scan(
		&plan.ID, &plan.Name, &plan.DisplayName, &plan.Description,
		&plan.PriceMonthly, &plan.PriceYearly, &plan.Currency,
		&plan.MaxUsers, &plan.MaxStorageGB, &plan.MaxAPIRequestsMonthly, &plan.MaxDepartments,
		&featuresBytes, &plan.IsActive, &plan.IsVisible, &plan.SortOrder,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	if featuresBytes != nil {
		if err := json.Unmarshal(featuresBytes, &plan.Features); err != nil {
			return nil, fmt.Errorf("failed to unmarshal features: %w", err)
		}
	}

	return plan, nil
}

// CreatePlan creates a new subscription plan or revives a deleted one
func (s *SubscriptionService) CreatePlan(ctx context.Context, plan *models.SubscriptionPlan) (*models.SubscriptionPlan, error) {
	// Check if plan with this name already exists
	var existingID string
	var existingDeletedAt *time.Time
	// Assuming 'name' is the unique logical identifier.
	queryCheck := `SELECT id, deleted_at FROM subscription_plans WHERE name = $1`
	// Note: subscription_plans has no tenant_id, it's global.

	err := s.db.QueryRowContext(ctx, queryCheck, plan.Name).Scan(&existingID, &existingDeletedAt)
	if err == nil {
		if existingDeletedAt == nil {
			return nil, fmt.Errorf("plan with name '%s' already exists", plan.Name)
		}

		// Plan exists but is soft-deleted -> Revive it
		reviveQuery := `
			UPDATE subscription_plans 
			SET is_active = $1, is_visible = $2, display_name = $3, description = $4,
				price_monthly = $5, price_yearly = $6, currency = $7,
				max_users = $8, max_storage_gb = $9, max_api_requests_monthly = $10, max_departments = $11,
				features = $12, sort_order = $13,
				updated_at = $14
			WHERE id = $15`

		featuresJSON, _ := json.Marshal(plan.Features)

		_, err = s.db.ExecContext(ctx, reviveQuery,
			plan.IsActive, plan.IsVisible, plan.DisplayName, plan.Description,
			plan.PriceMonthly, plan.PriceYearly, plan.Currency,
			plan.MaxUsers, plan.MaxStorageGB, plan.MaxAPIRequestsMonthly, plan.MaxDepartments,
			featuresJSON, plan.SortOrder,
			time.Now(), existingID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to revive plan: %w", err)
		}

		planID, _ := uuid.Parse(existingID)
		return s.GetPlanByID(ctx, planID)

	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing plan: %w", err)
	}

	// Insert new plan
	plan.ID = uuid.New()
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	query := `
		INSERT INTO subscription_plans (
			id, name, display_name, description, price_monthly, price_yearly, currency,
			max_users, max_storage_gb, max_api_requests_monthly, max_departments,
			features, is_active, is_visible, sort_order, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id`

	err = s.db.QueryRowContext(ctx, query,
		plan.ID, plan.Name, plan.DisplayName, plan.Description,
		plan.PriceMonthly, plan.PriceYearly, plan.Currency,
		plan.MaxUsers, plan.MaxStorageGB, plan.MaxAPIRequestsMonthly, plan.MaxDepartments,
		featuresJSON, plan.IsActive, plan.IsVisible, plan.SortOrder,
		plan.CreatedAt, plan.UpdatedAt,
	).Scan(&plan.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan, nil
}

// UpdatePlan updates an existing subscription plan
func (s *SubscriptionService) UpdatePlan(ctx context.Context, planID uuid.UUID, updates map[string]interface{}) (*models.SubscriptionPlan, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	allowedFields := map[string]bool{
		"display_name": true, "description": true, "price_monthly": true, "price_yearly": true,
		"max_users": true, "max_storage_gb": true, "max_api_requests_monthly": true,
		"max_departments": true, "features": true, "is_active": true, "is_visible": true, "sort_order": true,
	}

	for field, value := range updates {
		if allowedFields[field] {
			if field == "features" {
				featuresJSON, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal features: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, featuresJSON)
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, value)
			}
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return s.GetPlanByID(ctx, planID)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, planID)

	query := fmt.Sprintf("UPDATE subscription_plans SET %s WHERE id = $%d",
		joinStrings(setParts, ", "), argIndex)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return s.GetPlanByID(ctx, planID)
}

// DeletePlan hard deletes a plan (if not in use)
func (s *SubscriptionService) DeletePlan(ctx context.Context, planID uuid.UUID) error {
	// Hard delete the plan.
	// Note: If the plan is in use by any subscription, this will fail with a Foreign Key Violation constraint error.
	// This is desired behavior: we should not delete plans that are currently assigned to tenants.
	query := `DELETE FROM subscription_plans WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, planID)
	if err != nil {
		// Check for foreign key violation (Postgres error code 23503)
		// We can string match for simplicity or check pq driver error
		if errStr := err.Error(); len(errStr) > 0 {
			// Basic check for FK violation indication
			if result == nil { // Likely an error before execution or during constraint check
				// We can just return the error wrapped
			}
		}
		return fmt.Errorf("failed to delete plan (it might be in use): %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plan not found")
	}

	return nil
}

// GetSubscriptionByTenantID retrieves a tenant's subscription
func (s *SubscriptionService) GetSubscriptionByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT s.id, s.tenant_id, s.plan_id, s.status, s.billing_cycle, s.amount, s.currency,
			s.trial_ends_at, s.current_period_start, s.current_period_end, s.cancelled_at,
			s.auto_renew, s.notes, s.metadata, s.created_at, s.updated_at,
			p.name, p.display_name, p.price_monthly, p.price_yearly,
			t.name as tenant_name
		FROM subscriptions s
		JOIN subscription_plans p ON s.plan_id = p.id
		JOIN tenants t ON s.tenant_id = t.id
		WHERE s.tenant_id = $1`

	sub := &models.Subscription{Plan: &models.SubscriptionPlan{}}
	var metadataBytes []byte

	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.PlanID, &sub.Status, &sub.BillingCycle, &sub.Amount, &sub.Currency,
		&sub.TrialEndsAt, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelledAt,
		&sub.AutoRenew, &sub.Notes, &metadataBytes, &sub.CreatedAt, &sub.UpdatedAt,
		&sub.Plan.Name, &sub.Plan.DisplayName, &sub.Plan.PriceMonthly, &sub.Plan.PriceYearly,
		&sub.TenantName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if metadataBytes != nil {
		if err := json.Unmarshal(metadataBytes, &sub.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return sub, nil
}

// CreateSubscription creates a new subscription for a tenant
func (s *SubscriptionService) CreateSubscription(ctx context.Context, sub *models.Subscription) (*models.Subscription, error) {
	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(sub.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO subscriptions (
			id, tenant_id, plan_id, status, billing_cycle, amount, currency,
			trial_ends_at, current_period_start, current_period_end, auto_renew,
			notes, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	err = s.db.QueryRowContext(ctx, query,
		sub.ID, sub.TenantID, sub.PlanID, sub.Status, sub.BillingCycle, sub.Amount, sub.Currency,
		sub.TrialEndsAt, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.AutoRenew,
		sub.Notes, metadataJSON, sub.CreatedAt, sub.UpdatedAt,
	).Scan(&sub.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return s.GetSubscriptionByTenantID(ctx, sub.TenantID)
}

// RenewSubscription renews a subscription for another period or updates the plan
func (s *SubscriptionService) RenewSubscription(ctx context.Context, tenantID uuid.UUID, newPlanID *uuid.UUID, newBillingCycle *string) (*models.Subscription, error) {
	sub, err := s.GetSubscriptionByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Update Plan ID if provided
	if newPlanID != nil {
		sub.PlanID = *newPlanID
	}

	// Update Billing Cycle if provided
	if newBillingCycle != nil {
		sub.BillingCycle = *newBillingCycle
	}

	// Fetch plan details to calculate amount
	plan, err := s.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan details: %w", err)
	}

	// Calculate new amount and period
	if sub.BillingCycle == "monthly" {
		sub.Amount = plan.PriceMonthly
		sub.CurrentPeriodStart = time.Now()
		sub.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0)
	} else {
		sub.Amount = plan.PriceYearly
		sub.CurrentPeriodStart = time.Now()
		sub.CurrentPeriodEnd = time.Now().AddDate(1, 0, 0)
	}

	query := `
		UPDATE subscriptions 
		SET plan_id = $1, billing_cycle = $2, amount = $3, 
			current_period_start = $4, current_period_end = $5, 
			status = 'active', updated_at = $6
		WHERE tenant_id = $7`

	_, err = s.db.ExecContext(ctx, query,
		sub.PlanID, sub.BillingCycle, sub.Amount,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		time.Now(), tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to renew/update subscription: %w", err)
	}

	return s.GetSubscriptionByTenantID(ctx, tenantID)
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(ctx context.Context, tenantID uuid.UUID) error {
	query := `UPDATE subscriptions SET status = 'cancelled', cancelled_at = $1, updated_at = $2 WHERE tenant_id = $3`
	_, err := s.db.ExecContext(ctx, query, time.Now(), time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	return nil
}

// Helper function to join strings
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
