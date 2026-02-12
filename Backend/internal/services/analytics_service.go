package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	db *sql.DB
}

func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetPlatformStats retrieves overall platform statistics
func (s *AnalyticsService) GetPlatformStats(ctx context.Context) (map[string]interface{}, error) {
	// Total tenants
	var totalTenants, activeTenants, suspendedTenants int
	tenantQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'active') as active,
			COUNT(*) FILTER (WHERE status = 'suspended') as suspended
		FROM tenants
		WHERE deleted_at IS NULL`

	err := s.db.QueryRowContext(ctx, tenantQuery).Scan(&totalTenants, &activeTenants, &suspendedTenants)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}

	// Total subscriptions
	var totalSubscriptions, activeSubscriptions, trialSubscriptions int
	subQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'active') as active,
			COUNT(*) FILTER (WHERE status = 'trial') as trial
		FROM subscriptions`

	err = s.db.QueryRowContext(ctx, subQuery).Scan(&totalSubscriptions, &activeSubscriptions, &trialSubscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription stats: %w", err)
	}

	// Total users across all tenants
	var totalUsers, activeUsers int
	userQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM users
		WHERE deleted_at IS NULL`

	err = s.db.QueryRowContext(ctx, userQuery).Scan(&totalUsers, &activeUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Monthly Recurring Revenue (MRR)
	var mrr float64
	mrrQuery := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN billing_cycle = 'monthly' THEN amount
				WHEN billing_cycle = 'yearly' THEN amount / 12
				ELSE 0
			END
		), 0) as mrr
		FROM subscriptions
		WHERE status IN ('active', 'trial')`

	err = s.db.QueryRowContext(ctx, mrrQuery).Scan(&mrr)
	if err != nil {
		return nil, fmt.Errorf("failed to get MRR: %w", err)
	}

	// Total API requests (last 30 days)
	var totalAPIRequests int64
	apiQuery := `
		SELECT COALESCE(SUM(api_requests_count), 0)
		FROM usage_metrics
		WHERE metric_date >= $1`

	err = s.db.QueryRowContext(ctx, apiQuery, time.Now().AddDate(0, 0, -30)).Scan(&totalAPIRequests)
	if err != nil {
		totalAPIRequests = 0
	}

	// Total storage used
	var totalStorageMB int64
	storageQuery := `SELECT COALESCE(SUM(storage_used_mb), 0) FROM tenants WHERE deleted_at IS NULL`

	err = s.db.QueryRowContext(ctx, storageQuery).Scan(&totalStorageMB)
	if err != nil {
		totalStorageMB = 0
	}

	stats := map[string]interface{}{
		"total_tenants":          totalTenants,
		"active_tenants":         activeTenants,
		"suspended_tenants":      suspendedTenants,
		"total_subscriptions":    totalSubscriptions,
		"active_subscriptions":   activeSubscriptions,
		"trial_subscriptions":    trialSubscriptions,
		"total_users":            totalUsers,
		"active_users":           activeUsers,
		"mrr":                    mrr,
		"arr":                    mrr * 12, // Annual Recurring Revenue
		"total_api_requests_30d": totalAPIRequests,
		"total_storage_gb":       float64(totalStorageMB) / 1024,
	}

	return stats, nil
}

// GetTenantGrowth retrieves tenant growth statistics
func (s *AnalyticsService) GetTenantGrowth(ctx context.Context, days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as new_tenants
		FROM tenants
		WHERE created_at >= $1 AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date ASC`

	rows, err := s.db.QueryContext(ctx, query, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant growth: %w", err)
	}
	defer rows.Close()

	var growth []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var newTenants int

		err := rows.Scan(&date, &newTenants)
		if err != nil {
			return nil, fmt.Errorf("failed to scan growth data: %w", err)
		}

		growth = append(growth, map[string]interface{}{
			"date":        date.Format("2006-01-02"),
			"new_tenants": newTenants,
		})
	}

	return growth, nil
}

// GetRevenueMetrics retrieves revenue metrics over time
func (s *AnalyticsService) GetRevenueMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Total revenue
	var totalRevenue, paidRevenue, pendingRevenue float64
	revenueQuery := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'paid'), 0) as paid,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'pending'), 0) as pending
		FROM invoices
		WHERE issue_date BETWEEN $1 AND $2`

	err := s.db.QueryRowContext(ctx, revenueQuery, startDate, endDate).Scan(&totalRevenue, &paidRevenue, &pendingRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue metrics: %w", err)
	}

	// Revenue by plan
	planRevenueQuery := `
		SELECT 
			sp.display_name,
			COUNT(DISTINCT s.tenant_id) as subscribers,
			COALESCE(SUM(s.amount), 0) as revenue
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.status IN ('active', 'trial')
		GROUP BY sp.display_name
		ORDER BY revenue DESC`

	rows, err := s.db.QueryContext(ctx, planRevenueQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan revenue: %w", err)
	}
	defer rows.Close()

	var planRevenue []map[string]interface{}
	for rows.Next() {
		var planName string
		var subscribers int
		var revenue float64

		err := rows.Scan(&planName, &subscribers, &revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan revenue: %w", err)
		}

		planRevenue = append(planRevenue, map[string]interface{}{
			"plan":        planName,
			"subscribers": subscribers,
			"revenue":     revenue,
		})
	}

	metrics := map[string]interface{}{
		"total_revenue":   totalRevenue,
		"paid_revenue":    paidRevenue,
		"pending_revenue": pendingRevenue,
		"plan_revenue":    planRevenue,
	}

	return metrics, nil
}

// GetTopTenantsByUsage retrieves top tenants by various usage metrics
func (s *AnalyticsService) GetTopTenantsByUsage(ctx context.Context, metric string, limit int) ([]map[string]interface{}, error) {
	var query string

	switch metric {
	case "api_requests":
		query = `
			SELECT t.id, t.name, SUM(um.api_requests_count) as total
			FROM tenants t
			JOIN usage_metrics um ON t.id = um.tenant_id
			WHERE um.metric_date >= $1
			GROUP BY t.id, t.name
			ORDER BY total DESC
			LIMIT $2`
	case "storage":
		query = `
			SELECT id, name, storage_used_mb as total
			FROM tenants
			WHERE deleted_at IS NULL
			ORDER BY storage_used_mb DESC
			LIMIT $1`
	case "users":
		query = `
			SELECT t.id, t.name, COUNT(u.id) as total
			FROM tenants t
			JOIN users u ON t.id = u.tenant_id
			WHERE t.deleted_at IS NULL AND u.deleted_at IS NULL
			GROUP BY t.id, t.name
			ORDER BY total DESC
			LIMIT $1`
	default:
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	var rows *sql.Rows
	var err error

	if metric == "api_requests" {
		rows, err = s.db.QueryContext(ctx, query, time.Now().AddDate(0, 0, -30), limit)
	} else if metric == "storage" {
		rows, err = s.db.QueryContext(ctx, query, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query top tenants: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var tenantID uuid.UUID
		var tenantName string
		var total interface{}

		err := rows.Scan(&tenantID, &tenantName, &total)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		results = append(results, map[string]interface{}{
			"tenant_id":   tenantID,
			"tenant_name": tenantName,
			"total":       total,
		})
	}

	return results, nil
}

// GetChurnRate calculates the churn rate for a period
func (s *AnalyticsService) GetChurnRate(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	// Count tenants at start of period
	var tenantsAtStart int
	startQuery := `
		SELECT COUNT(*) FROM tenants
		WHERE created_at < $1 AND deleted_at IS NULL`

	err := s.db.QueryRowContext(ctx, startQuery, startDate).Scan(&tenantsAtStart)
	if err != nil {
		return 0, fmt.Errorf("failed to get tenants at start: %w", err)
	}

	// Count churned tenants (cancelled subscriptions)
	var churnedTenants int
	churnQuery := `
		SELECT COUNT(*) FROM subscriptions
		WHERE status = 'cancelled' AND cancelled_at BETWEEN $1 AND $2`

	err = s.db.QueryRowContext(ctx, churnQuery, startDate, endDate).Scan(&churnedTenants)
	if err != nil {
		return 0, fmt.Errorf("failed to get churned tenants: %w", err)
	}

	if tenantsAtStart == 0 {
		return 0, nil
	}

	churnRate := float64(churnedTenants) / float64(tenantsAtStart) * 100
	return churnRate, nil
}
