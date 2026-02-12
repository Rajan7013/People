package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type UsageTrackingService struct {
	db *sql.DB
}

func NewUsageTrackingService(db *sql.DB) *UsageTrackingService {
	return &UsageTrackingService{db: db}
}

// RecordDailyMetrics records or updates daily usage metrics for a tenant
func (s *UsageTrackingService) RecordDailyMetrics(ctx context.Context, tenantID uuid.UUID, date time.Time) error {
	metricDate := date.Truncate(24 * time.Hour)

	// Aggregate metrics for the day
	var totalUsers, activeUsers, newUsers int
	var storageUsedMB int64
	var apiRequestsCount, loginsCount, uniqueLoginsCount int
	var attendanceRecordsCount, leaveRequestsCount, employeesAddedCount int

	// Get user metrics
	userQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE last_login_at::date = $2) as active_users,
			COUNT(*) FILTER (WHERE created_at::date = $2) as new_users
		FROM users
		WHERE tenant_id = $1 AND deleted_at IS NULL`

	err := s.db.QueryRowContext(ctx, userQuery, tenantID, metricDate).Scan(&totalUsers, &activeUsers, &newUsers)
	if err != nil {
		return fmt.Errorf("failed to get user metrics: %w", err)
	}

	// Get API request metrics
	apiQuery := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(DISTINCT user_id) as unique_users
		FROM api_request_logs
		WHERE tenant_id = $1 AND created_at::date = $2`

	err = s.db.QueryRowContext(ctx, apiQuery, tenantID, metricDate).Scan(&apiRequestsCount, &uniqueLoginsCount)
	if err != nil {
		// If no API logs table yet, set to 0
		apiRequestsCount = 0
		uniqueLoginsCount = 0
	}

	// Get attendance records count
	attendanceQuery := `
		SELECT COUNT(*) FROM attendance_records
		WHERE tenant_id = $1 AND date = $2`

	err = s.db.QueryRowContext(ctx, attendanceQuery, tenantID, metricDate).Scan(&attendanceRecordsCount)
	if err != nil {
		attendanceRecordsCount = 0
	}

	// Get leave requests count
	leaveQuery := `
		SELECT COUNT(*) FROM leave_applications
		WHERE tenant_id = $1 AND created_at::date = $2`

	err = s.db.QueryRowContext(ctx, leaveQuery, tenantID, metricDate).Scan(&leaveRequestsCount)
	if err != nil {
		leaveRequestsCount = 0
	}

	// Get employees added count
	employeeQuery := `
		SELECT COUNT(*) FROM employees
		WHERE tenant_id = $1 AND created_at::date = $2`

	err = s.db.QueryRowContext(ctx, employeeQuery, tenantID, metricDate).Scan(&employeesAddedCount)
	if err != nil {
		employeesAddedCount = 0
	}

	// Insert or update usage metrics
	upsertQuery := `
		INSERT INTO usage_metrics (
			id, tenant_id, metric_date, total_users, active_users, new_users,
			storage_used_mb, api_requests_count, logins_count, unique_logins_count,
			attendance_records_count, leave_requests_count, employees_added_count,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (tenant_id, metric_date) DO UPDATE SET
			total_users = EXCLUDED.total_users,
			active_users = EXCLUDED.active_users,
			new_users = EXCLUDED.new_users,
			storage_used_mb = EXCLUDED.storage_used_mb,
			api_requests_count = EXCLUDED.api_requests_count,
			logins_count = EXCLUDED.logins_count,
			unique_logins_count = EXCLUDED.unique_logins_count,
			attendance_records_count = EXCLUDED.attendance_records_count,
			leave_requests_count = EXCLUDED.leave_requests_count,
			employees_added_count = EXCLUDED.employees_added_count,
			updated_at = EXCLUDED.updated_at`

	_, err = s.db.ExecContext(ctx, upsertQuery,
		uuid.New(), tenantID, metricDate, totalUsers, activeUsers, newUsers,
		storageUsedMB, apiRequestsCount, loginsCount, uniqueLoginsCount,
		attendanceRecordsCount, leaveRequestsCount, employeesAddedCount,
		time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to record usage metrics: %w", err)
	}

	return nil
}

// GetUsageByTenant retrieves usage metrics for a tenant within a date range
func (s *UsageTrackingService) GetUsageByTenant(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*models.UsageMetric, error) {
	query := `
		SELECT id, tenant_id, metric_date, total_users, active_users, new_users,
			storage_used_mb, storage_documents_mb, storage_attachments_mb,
			api_requests_count, api_requests_success, api_requests_failed,
			logins_count, unique_logins_count, attendance_records_count,
			leave_requests_count, employees_added_count, additional_metrics,
			created_at, updated_at
		FROM usage_metrics
		WHERE tenant_id = $1 AND metric_date BETWEEN $2 AND $3
		ORDER BY metric_date DESC`

	rows, err := s.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*models.UsageMetric
	for rows.Next() {
		metric := &models.UsageMetric{}
		err := rows.Scan(
			&metric.ID, &metric.TenantID, &metric.MetricDate, &metric.TotalUsers,
			&metric.ActiveUsers, &metric.NewUsers, &metric.StorageUsedMB,
			&metric.StorageDocumentsMB, &metric.StorageAttachmentsMB,
			&metric.APIRequestsCount, &metric.APIRequestsSuccess, &metric.APIRequestsFailed,
			&metric.LoginsCount, &metric.UniqueLoginsCount, &metric.AttendanceRecordsCount,
			&metric.LeaveRequestsCount, &metric.EmployeesAddedCount, &metric.AdditionalMetrics,
			&metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetUsageSummary retrieves aggregated usage summary for a tenant
func (s *UsageTrackingService) GetUsageSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			COALESCE(SUM(api_requests_count), 0) as total_api_requests,
			COALESCE(AVG(api_requests_count), 0) as avg_daily_requests,
			COALESCE(MAX(storage_used_mb), 0) as max_storage_mb,
			COALESCE(AVG(active_users), 0) as avg_active_users,
			COALESCE(SUM(logins_count), 0) as total_logins,
			COALESCE(SUM(new_users), 0) as total_new_users,
			COALESCE(SUM(attendance_records_count), 0) as total_attendance_records,
			COALESCE(SUM(leave_requests_count), 0) as total_leave_requests
		FROM usage_metrics
		WHERE tenant_id = $1 AND metric_date BETWEEN $2 AND $3`

	var totalAPIRequests, maxStorageMB, totalLogins, totalNewUsers int64
	var totalAttendanceRecords, totalLeaveRequests int64
	var avgDailyRequests, avgActiveUsers float64

	err := s.db.QueryRowContext(ctx, query, tenantID, startDate, endDate).Scan(
		&totalAPIRequests, &avgDailyRequests, &maxStorageMB, &avgActiveUsers,
		&totalLogins, &totalNewUsers, &totalAttendanceRecords, &totalLeaveRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	summary := map[string]interface{}{
		"total_api_requests":       totalAPIRequests,
		"avg_daily_requests":       avgDailyRequests,
		"max_storage_mb":           maxStorageMB,
		"avg_active_users":         avgActiveUsers,
		"total_logins":             totalLogins,
		"total_new_users":          totalNewUsers,
		"total_attendance_records": totalAttendanceRecords,
		"total_leave_requests":     totalLeaveRequests,
	}

	return summary, nil
}

// LogAPIRequest logs an API request for usage tracking
func (s *UsageTrackingService) LogAPIRequest(ctx context.Context, log *models.APIRequestLog) error {
	query := `
		INSERT INTO api_request_logs (
			id, tenant_id, user_id, method, endpoint, status_code, response_time_ms,
			request_size_bytes, response_size_bytes, ip_address, user_agent, error_message,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	_, err := s.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.UserID, log.Method, log.Endpoint, log.StatusCode,
		log.ResponseTimeMS, log.RequestSizeBytes, log.ResponseSizeBytes, log.IPAddress,
		log.UserAgent, log.ErrorMessage, log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to log API request: %w", err)
	}

	return nil
}

// GetAPIUsageStats retrieves API usage statistics for a tenant
func (s *UsageTrackingService) GetAPIUsageStats(ctx context.Context, tenantID uuid.UUID, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 300) as successful_requests,
			COUNT(*) FILTER (WHERE status_code >= 400) as failed_requests,
			AVG(response_time_ms) as avg_response_time,
			MAX(response_time_ms) as max_response_time,
			COUNT(DISTINCT user_id) as unique_users
		FROM api_request_logs
		WHERE tenant_id = $1 AND created_at >= $2`

	var totalRequests, successfulRequests, failedRequests, uniqueUsers int
	var avgResponseTime, maxResponseTime float64

	err := s.db.QueryRowContext(ctx, query, tenantID, startDate).Scan(
		&totalRequests, &successfulRequests, &failedRequests,
		&avgResponseTime, &maxResponseTime, &uniqueUsers,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get API usage stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_requests":       totalRequests,
		"successful_requests":  successfulRequests,
		"failed_requests":      failedRequests,
		"avg_response_time_ms": avgResponseTime,
		"max_response_time_ms": maxResponseTime,
		"unique_users":         uniqueUsers,
		"success_rate":         0.0,
	}

	if totalRequests > 0 {
		stats["success_rate"] = float64(successfulRequests) / float64(totalRequests) * 100
	}

	return stats, nil
}
