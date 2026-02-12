package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	SettingKey   string     `json:"setting_key" db:"setting_key"`
	SettingValue *string    `json:"setting_value" db:"setting_value"`
	SettingType  string     `json:"setting_type" db:"setting_type"`
	Description  *string    `json:"description" db:"description"`
	IsSensitive  bool       `json:"is_sensitive" db:"is_sensitive"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy    *uuid.UUID `json:"updated_by" db:"updated_by"`
}

// AuditLog represents a system audit log entry
type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	UserID       *uuid.UUID      `json:"user_id" db:"user_id"`
	Action       string          `json:"action" db:"action"`
	ResourceType string          `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id" db:"resource_id"`
	OldValues    json.RawMessage `json:"old_values" db:"old_values"`
	NewValues    json.RawMessage `json:"new_values" db:"new_values"`
	IPAddress    *string         `json:"ip_address" db:"ip_address"`
	UserAgent    *string         `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// SystemBackup represents a database backup record
type SystemBackup struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	BackupType   string     `json:"backup_type" db:"backup_type"`
	FileName     string     `json:"file_name" db:"file_name"`
	FileSize     *int64     `json:"file_size" db:"file_size"`
	FilePath     *string    `json:"file_path" db:"file_path"`
	Status       string     `json:"status" db:"status"`
	ErrorMessage *string    `json:"error_message" db:"error_message"`
	CreatedBy    *uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
}

// SystemMetric represents a performance metric
type SystemMetric struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	MetricName  string          `json:"metric_name" db:"metric_name"`
	MetricValue *float64        `json:"metric_value" db:"metric_value"`
	MetricUnit  *string         `json:"metric_unit" db:"metric_unit"`
	Tags        json.RawMessage `json:"tags" db:"tags"`
	RecordedAt  time.Time       `json:"recorded_at" db:"recorded_at"`
}

// APIKey represents an API key for external access
type APIKey struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	KeyName     string          `json:"key_name" db:"key_name"`
	APIKeyHash  string          `json:"-" db:"api_key_hash"` // Never expose the hash
	KeyPrefix   *string         `json:"key_prefix" db:"key_prefix"`
	Permissions json.RawMessage `json:"permissions" db:"permissions"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	ExpiresAt   *time.Time      `json:"expires_at" db:"expires_at"`
	LastUsedAt  *time.Time      `json:"last_used_at" db:"last_used_at"`
	UsageCount  int             `json:"usage_count" db:"usage_count"`
	CreatedBy   *uuid.UUID      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name            string          `json:"name" db:"name"`
	EndpointURL     string          `json:"endpoint_url" db:"endpoint_url"`
	Events          pq.StringArray  `json:"events" db:"events"`
	SecretKey       *string         `json:"secret_key,omitempty" db:"secret_key"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	RetryCount      int             `json:"retry_count" db:"retry_count"`
	TimeoutSeconds  int             `json:"timeout_seconds" db:"timeout_seconds"`
	Headers         json.RawMessage `json:"headers" db:"headers"`
	LastTriggeredAt *time.Time      `json:"last_triggered_at" db:"last_triggered_at"`
	LastStatus      *int            `json:"last_status" db:"last_status"`
	FailureCount    int             `json:"failure_count" db:"failure_count"`
	CreatedBy       *uuid.UUID      `json:"created_by" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// SystemManagementService handles all system management operations
type SystemManagementService struct {
	db *sql.DB
}

// NewSystemManagementService creates a new system management service
func NewSystemManagementService(db *sql.DB) *SystemManagementService {
	return &SystemManagementService{db: db}
}

// ===== SYSTEM SETTINGS =====

// GetSettings retrieves all system settings for a tenant
func (s *SystemManagementService) GetSettings(tenantID uuid.UUID) ([]SystemSetting, error) {
	query := `
		SELECT id, tenant_id, setting_key, setting_value, setting_type, description, 
		       is_sensitive, created_at, updated_at, updated_by 
		FROM system_settings 
		WHERE tenant_id = $1 
		ORDER BY setting_key`

	rows, err := s.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get system settings: %w", err)
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var setting SystemSetting
		err := rows.Scan(&setting.ID, &setting.TenantID, &setting.SettingKey, &setting.SettingValue,
			&setting.SettingType, &setting.Description, &setting.IsSensitive,
			&setting.CreatedAt, &setting.UpdatedAt, &setting.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		// Mask sensitive values
		if setting.IsSensitive && setting.SettingValue != nil {
			masked := "****"
			setting.SettingValue = &masked
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

// UpdateSetting updates a system setting
func (s *SystemManagementService) UpdateSetting(tenantID uuid.UUID, settingKey, settingValue string, userID *uuid.UUID) error {
	query := `
		UPDATE system_settings 
		SET setting_value = $1, updated_by = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE tenant_id = $3 AND setting_key = $4`

	result, err := s.db.Exec(query, settingValue, userID, tenantID, settingKey)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("setting not found")
	}

	return nil
}

// CreateSetting creates a new system setting
func (s *SystemManagementService) CreateSetting(tenantID uuid.UUID, settingKey, settingValue, settingType string, description *string, isSensitive bool, userID *uuid.UUID) error {
	query := `
		INSERT INTO system_settings (tenant_id, setting_key, setting_value, setting_type, description, is_sensitive, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Exec(query, tenantID, settingKey, settingValue, settingType, description, isSensitive, userID)
	if err != nil {
		return fmt.Errorf("failed to create setting: %w", err)
	}

	return nil
}

// ===== AUDIT LOGS =====

// CreateAuditLog creates a new audit log entry
func (s *SystemManagementService) CreateAuditLog(tenantID uuid.UUID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, oldValues, newValues json.RawMessage, ipAddress, userAgent *string) error {
	query := `
		INSERT INTO audit_logs (tenant_id, user_id, action, resource_type, resource_id, old_values, new_values, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.db.Exec(query, tenantID, userID, action, resourceType, resourceID, oldValues, newValues, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a tenant with pagination
func (s *SystemManagementService) GetAuditLogs(tenantID uuid.UUID, limit, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource_type, resource_id, 
		       old_values, new_values, ip_address, user_agent, created_at 
		FROM audit_logs 
		WHERE tenant_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(&log.ID, &log.TenantID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.OldValues, &log.NewValues, &log.IPAddress, &log.UserAgent, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// ===== SYSTEM METRICS =====

// RecordMetric records a new system metric
func (s *SystemManagementService) RecordMetric(tenantID uuid.UUID, metricName string, metricValue *float64, metricUnit *string, tags json.RawMessage) error {
	query := `
		INSERT INTO system_metrics (tenant_id, metric_name, metric_value, metric_unit, tags)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, tenantID, metricName, metricValue, metricUnit, tags)
	if err != nil {
		return fmt.Errorf("failed to record metric: %w", err)
	}

	return nil
}

// GetMetrics retrieves system metrics for a tenant
func (s *SystemManagementService) GetMetrics(tenantID uuid.UUID, metricName *string, since *time.Time, limit int) ([]SystemMetric, error) {
	baseQuery := `
		SELECT id, tenant_id, metric_name, metric_value, metric_unit, tags, recorded_at 
		FROM system_metrics 
		WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2

	if metricName != nil {
		baseQuery += fmt.Sprintf(" AND metric_name = $%d", argIndex)
		args = append(args, *metricName)
		argIndex++
	}

	if since != nil {
		baseQuery += fmt.Sprintf(" AND recorded_at >= $%d", argIndex)
		args = append(args, *since)
		argIndex++
	}

	baseQuery += fmt.Sprintf(" ORDER BY recorded_at DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	var metrics []SystemMetric
	for rows.Next() {
		var metric SystemMetric
		err := rows.Scan(&metric.ID, &metric.TenantID, &metric.MetricName, &metric.MetricValue,
			&metric.MetricUnit, &metric.Tags, &metric.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// ===== SYSTEM BACKUPS =====

// CreateBackup creates a new backup record
func (s *SystemManagementService) CreateBackup(tenantID uuid.UUID, backupType, fileName string, createdBy *uuid.UUID) (*SystemBackup, error) {
	backup := SystemBackup{
		ID:         uuid.New(),
		TenantID:   tenantID,
		BackupType: backupType,
		FileName:   fileName,
		Status:     "pending",
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
	}

	query := `
		INSERT INTO system_backups (id, tenant_id, backup_type, file_name, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Exec(query, backup.ID, backup.TenantID, backup.BackupType, backup.FileName, backup.Status, backup.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	return &backup, nil
}

// UpdateBackupStatus updates the status of a backup
func (s *SystemManagementService) UpdateBackupStatus(backupID uuid.UUID, status string, fileSize *int64, filePath, errorMessage *string) error {
	query := `
		UPDATE system_backups 
		SET status = $1, file_size = $2, file_path = $3, error_message = $4, completed_at = $5
		WHERE id = $6`

	var completedAt *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now()
		completedAt = &now
	}

	_, err := s.db.Exec(query, status, fileSize, filePath, errorMessage, completedAt, backupID)
	if err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}

	return nil
}

// GetBackups retrieves backup records for a tenant
func (s *SystemManagementService) GetBackups(tenantID uuid.UUID, limit, offset int) ([]SystemBackup, error) {
	query := `
		SELECT id, tenant_id, backup_type, file_name, file_size, file_path, 
		       status, error_message, created_by, created_at, completed_at 
		FROM system_backups 
		WHERE tenant_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get backups: %w", err)
	}
	defer rows.Close()

	var backups []SystemBackup
	for rows.Next() {
		var backup SystemBackup
		err := rows.Scan(&backup.ID, &backup.TenantID, &backup.BackupType, &backup.FileName,
			&backup.FileSize, &backup.FilePath, &backup.Status, &backup.ErrorMessage,
			&backup.CreatedBy, &backup.CreatedAt, &backup.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backup: %w", err)
		}
		backups = append(backups, backup)
	}

	return backups, nil
}
