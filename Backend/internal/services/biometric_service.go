package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

// BiometricService handles biometric device and attendance operations
type BiometricService struct {
	db *sql.DB
}

// NewBiometricService creates a new biometric service
func NewBiometricService(db *sql.DB) *BiometricService {
	return &BiometricService{db: db}
}

// RegisterDevice registers a new biometric device
func (s *BiometricService) RegisterDevice(ctx context.Context, tenantID uuid.UUID, req models.DeviceRegistrationRequest) (*models.BiometricDevice, error) {
	deviceID := uuid.New()
	now := time.Now()

	device := &models.BiometricDevice{
		ID:           deviceID,
		TenantID:     tenantID,
		DeviceName:   req.DeviceName,
		DeviceType:   req.DeviceType,
		SerialNumber: req.SerialNumber,
		MacAddress:   req.MacAddress,
		IPAddress:    req.IPAddress,
		Port:         req.Port,
		Location:     req.Location,
		Department:   req.Department,
		Status:       models.DeviceStatusActive,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Config:       req.Config,
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO biometric_devices (
			id, tenant_id, device_name, device_type, serial_number, mac_address, 
			ip_address, port, location, department, status, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		device.ID, device.TenantID, device.DeviceName, device.DeviceType, device.SerialNumber,
		device.MacAddress, device.IPAddress, device.Port, device.Location, device.Department,
		device.Status, device.IsActive, device.CreatedAt, device.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	return device, nil
}

// GetDevices retrieves all biometric devices for a tenant
func (s *BiometricService) GetDevices(ctx context.Context, tenantID uuid.UUID) ([]*models.BiometricDevice, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, device_name, device_type, serial_number, mac_address, 
		        ip_address, port, location, department, status, last_sync_at, is_active, 
		        created_at, updated_at
		 FROM biometric_devices 
		 WHERE tenant_id = $1 AND is_active = true
		 ORDER BY device_name`,
		tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.BiometricDevice
	for rows.Next() {
		device := &models.BiometricDevice{}
		var macAddress, ipAddress, location, department sql.NullString
		var port sql.NullInt64
		var lastSyncAt sql.NullTime

		err := rows.Scan(
			&device.ID, &device.TenantID, &device.DeviceName, &device.DeviceType,
			&device.SerialNumber, &macAddress, &ipAddress, &port, &location,
			&department, &device.Status, &lastSyncAt, &device.IsActive,
			&device.CreatedAt, &device.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		// Handle nullable fields
		if macAddress.Valid {
			device.MacAddress = &macAddress.String
		}
		if ipAddress.Valid {
			device.IPAddress = &ipAddress.String
		}
		if port.Valid {
			portInt := int(port.Int64)
			device.Port = &portInt
		}
		if location.Valid {
			device.Location = &location.String
		}
		if department.Valid {
			device.Department = &department.String
		}
		if lastSyncAt.Valid {
			device.LastSyncAt = &lastSyncAt.Time
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// UpdateDeviceStatus updates the status of a biometric device
func (s *BiometricService) UpdateDeviceStatus(ctx context.Context, deviceID uuid.UUID, status models.BiometricDeviceStatus) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE biometric_devices 
		 SET status = $1, last_sync_at = CASE WHEN $1 = 'active' THEN NOW() ELSE last_sync_at END, updated_at = NOW()
		 WHERE id = $2`,
		status, deviceID)

	if err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	return nil
}

// ProcessBiometricLog processes a biometric attendance log entry
func (s *BiometricService) ProcessBiometricLog(ctx context.Context, tenantID uuid.UUID, log models.BiometricAttendanceLog) error {
	// First, find the employee by employee code
	var employeeID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT e.id 
		 FROM employees e 
		 WHERE e.tenant_id = $1 AND e.employee_code = $2 AND e.deleted_at IS NULL`,
		tenantID, log.EmployeeCode).Scan(&employeeID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("employee not found with code: %s", log.EmployeeCode)
		}
		return fmt.Errorf("failed to find employee: %w", err)
	}

	// Insert the biometric log
	logID := uuid.New()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO biometric_attendance_logs (
			id, tenant_id, device_id, employee_code, employee_id, timestamp, 
			event_type, biometric_data, confidence, is_processed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, NOW())`,
		logID, tenantID, log.DeviceID, log.EmployeeCode, employeeID,
		log.Timestamp, log.EventType, log.BiometricData, log.Confidence)

	if err != nil {
		return fmt.Errorf("failed to insert biometric log: %w", err)
	}

	// Process the log into attendance record
	return s.processLogToAttendance(ctx, logID, tenantID, employeeID, log)
}

// processLogToAttendance converts a biometric log to attendance record
func (s *BiometricService) processLogToAttendance(ctx context.Context, logID, tenantID, employeeID uuid.UUID, biometricLog models.BiometricAttendanceLog) error {
	date := biometricLog.Timestamp.Format("2006-01-02")

	// Check if attendance record exists for this date
	var recordID uuid.UUID
	var existingCheckIn, existingCheckOut sql.NullTime

	err := s.db.QueryRowContext(ctx,
		`SELECT id, check_in_time, check_out_time 
		 FROM attendance_records 
		 WHERE tenant_id = $1 AND employee_id = $2 AND date = $3`,
		tenantID, employeeID, date).Scan(&recordID, &existingCheckIn, &existingCheckOut)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing attendance: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create new attendance record
		recordID = uuid.New()
		status := "present"
		if biometricLog.EventType == "check_in" {
			// Check if late (assuming 9:00 AM standard time)
			standardStart := time.Date(biometricLog.Timestamp.Year(), biometricLog.Timestamp.Month(), biometricLog.Timestamp.Day(), 9, 0, 0, 0, biometricLog.Timestamp.Location())
			if biometricLog.Timestamp.After(standardStart) {
				status = "late"
			}
		}

		var checkInTime, checkOutTime *time.Time
		switch biometricLog.EventType {
		case "check_in":
			checkInTime = &biometricLog.Timestamp
		case "check_out":
			checkOutTime = &biometricLog.Timestamp
		}

		_, err = s.db.ExecContext(ctx,
			`INSERT INTO attendance_records (
				id, tenant_id, employee_id, date, check_in_time, check_out_time, 
				status, source, device_id, biometric_log_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
			recordID, tenantID, employeeID, date, checkInTime, checkOutTime,
			status, models.SourceBiometric, biometricLog.DeviceID, logID)
	} else {
		// Update existing record
		if biometricLog.EventType == "check_in" && !existingCheckIn.Valid {
			_, err = s.db.ExecContext(ctx,
				`UPDATE attendance_records 
				 SET check_in_time = $1, source = $2, device_id = $3, biometric_log_id = $4, updated_at = NOW()
				 WHERE id = $5`,
				biometricLog.Timestamp, models.SourceBiometric, biometricLog.DeviceID, logID, recordID)
		} else if biometricLog.EventType == "check_out" && !existingCheckOut.Valid {
			// Calculate total hours
			var totalHours *float64
			if existingCheckIn.Valid {
				duration := biometricLog.Timestamp.Sub(existingCheckIn.Time)
				hours := duration.Hours()
				totalHours = &hours
			}

			_, err = s.db.ExecContext(ctx,
				`UPDATE attendance_records 
				 SET check_out_time = $1, total_hours = $2, source = $3, device_id = $4, updated_at = NOW()
				 WHERE id = $5`,
				biometricLog.Timestamp, totalHours, models.SourceBiometric, biometricLog.DeviceID, recordID)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to update attendance record: %w", err)
	}

	// Mark log as processed
	_, err = s.db.ExecContext(ctx,
		`UPDATE biometric_attendance_logs 
		 SET is_processed = true, processed_at = NOW() 
		 WHERE id = $1`,
		logID)

	if err != nil {
		log.Printf("Failed to mark biometric log as processed: %v", err)
		// Don't return error as the main operation succeeded
	}

	return nil
} // SyncDeviceData simulates syncing data from a biometric device
func (s *BiometricService) SyncDeviceData(ctx context.Context, tenantID uuid.UUID, req models.BiometricSyncRequest) (*models.BiometricSyncResponse, error) {
	startTime := time.Now()

	// Update device sync status
	err := s.UpdateDeviceStatus(ctx, req.DeviceID, models.DeviceStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to update device status: %w", err)
	}

	// In a real implementation, this would connect to the actual biometric device
	// For now, we'll simulate successful sync
	response := &models.BiometricSyncResponse{
		DeviceID:       req.DeviceID,
		SyncStarted:    startTime,
		SyncCompleted:  time.Now(),
		RecordsSynced:  0,
		RecordsSkipped: 0,
		Status:         "success",
	}

	// Update last sync time
	_, err = s.db.ExecContext(ctx,
		`UPDATE biometric_devices 
		 SET last_sync_at = NOW(), updated_at = NOW()
		 WHERE id = $1`,
		req.DeviceID)

	if err != nil {
		log.Printf("Failed to update device sync time: %v", err)
	}

	return response, nil
}

// GetBiometricLogs retrieves biometric attendance logs with pagination
func (s *BiometricService) GetBiometricLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.BiometricAttendanceLog, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT bl.id, bl.tenant_id, bl.device_id, bl.employee_code, bl.employee_id, 
		        bl.timestamp, bl.event_type, bl.confidence, bl.is_processed, bl.processed_at, 
		        bl.error_message, bl.created_at,
		        bd.device_name,
		        COALESCE(CONCAT(u.first_name, ' ', u.last_name), bl.employee_code) as employee_name
		 FROM biometric_attendance_logs bl
		 LEFT JOIN biometric_devices bd ON bl.device_id = bd.id
		 LEFT JOIN employees e ON bl.employee_id = e.id
		 LEFT JOIN users u ON e.user_id = u.id
		 WHERE bl.tenant_id = $1
		 ORDER BY bl.timestamp DESC
		 LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get biometric logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.BiometricAttendanceLog
	for rows.Next() {
		log := &models.BiometricAttendanceLog{}
		var employeeID sql.NullString
		var confidence sql.NullFloat64
		var processedAt sql.NullTime
		var errorMessage sql.NullString
		var deviceName sql.NullString
		var employeeName sql.NullString

		err := rows.Scan(
			&log.ID, &log.TenantID, &log.DeviceID, &log.EmployeeCode, &employeeID,
			&log.Timestamp, &log.EventType, &confidence, &log.IsProcessed, &processedAt,
			&errorMessage, &log.CreatedAt, &deviceName, &employeeName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan biometric log: %w", err)
		}

		// Handle nullable fields
		if employeeID.Valid {
			id, _ := uuid.Parse(employeeID.String)
			log.EmployeeID = &id
		}
		if confidence.Valid {
			log.Confidence = &confidence.Float64
		}
		if processedAt.Valid {
			log.ProcessedAt = &processedAt.Time
		}
		if errorMessage.Valid {
			log.ErrorMessage = &errorMessage.String
		}
		if deviceName.Valid {
			log.DeviceName = deviceName.String
		}
		if employeeName.Valid {
			log.EmployeeName = employeeName.String
		}

		logs = append(logs, log)
	}

	return logs, nil
}
