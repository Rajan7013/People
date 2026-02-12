package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type AttendanceService struct {
	db *sql.DB
}

func NewAttendanceService(db *sql.DB) *AttendanceService {
	return &AttendanceService{db: db}
}

// CheckIn records employee check-in
func (s *AttendanceService) CheckIn(ctx context.Context, tenantID, employeeID uuid.UUID, notes string) (*models.AttendanceRecord, error) {
	return s.CheckInWithSource(ctx, tenantID, employeeID, notes, models.SourceManual, nil, nil)
}

// CheckInWithSource allows specifying the attendance source and optional device/log references
func (s *AttendanceService) CheckInWithSource(ctx context.Context, tenantID, employeeID uuid.UUID, notes string, source models.AttendanceSource, deviceID, biometricLogID *uuid.UUID) (*models.AttendanceRecord, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// Check if already checked in today
	var existingID sql.NullString
	var existingCheckIn sql.NullTime
	err := s.db.QueryRowContext(ctx,
		"SELECT id, check_in_time FROM attendance_records WHERE tenant_id = $1 AND employee_id = $2 AND date = $3",
		tenantID, employeeID, today).Scan(&existingID, &existingCheckIn)

	if err == nil && existingCheckIn.Valid {
		return nil, fmt.Errorf("employee already checked in today")
	}

	// Determine status (simple logic for now)
	status := "present"
	hour := now.Hour()
	if hour > 9 { // If after 9 AM, consider late
		status = "late"
	}

	record := models.AttendanceRecord{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		Date:           now,
		CheckInTime:    &now,
		Status:         status,
		Notes:          &notes,
		Source:         source,
		DeviceID:       deviceID,
		BiometricLogID: biometricLogID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err == sql.ErrNoRows {
		// Insert new record
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO attendance_records (id, tenant_id, employee_id, date, check_in_time, status, notes, source, device_id, biometric_log_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			record.ID, record.TenantID, record.EmployeeID, today, record.CheckInTime,
			record.Status, record.Notes, record.Source, record.DeviceID, record.BiometricLogID, record.CreatedAt, record.UpdatedAt)
	} else if existingID.Valid {
		// Update existing record
		record.ID = uuid.MustParse(existingID.String)
		_, err = s.db.ExecContext(ctx, `
			UPDATE attendance_records 
			SET check_in_time = $1, status = $2, notes = $3, source = $4, device_id = $5, biometric_log_id = $6, updated_at = $7
			WHERE id = $8`,
			record.CheckInTime, record.Status, record.Notes, record.Source, record.DeviceID, record.BiometricLogID, record.UpdatedAt, record.ID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save check-in record: %w", err)
	}

	return &record, nil
}

// CheckOut records employee check-out
func (s *AttendanceService) CheckOut(ctx context.Context, tenantID, employeeID uuid.UUID, notes string) (*models.AttendanceRecord, error) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// Get existing record
	var record models.AttendanceRecord
	var checkInTime, checkOutTime sql.NullTime
	var totalHours sql.NullFloat64
	var existingNotes sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, employee_id, date, check_in_time, check_out_time, 
		        total_hours, overtime_hours, status, notes, created_at, updated_at
		 FROM attendance_records WHERE tenant_id = $1 AND employee_id = $2 AND date = $3`,
		tenantID, employeeID, today).Scan(
		&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
		&checkInTime, &checkOutTime, &totalHours, &record.OvertimeHours,
		&record.Status, &existingNotes, &record.CreatedAt, &record.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no check-in record found for today")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance record: %w", err)
	}

	if checkOutTime.Valid {
		return nil, fmt.Errorf("employee already checked out today")
	}

	if !checkInTime.Valid {
		return nil, fmt.Errorf("employee has not checked in today")
	}

	record.CheckInTime = &checkInTime.Time
	record.CheckOutTime = &now

	// Calculate total hours
	totalHoursFloat := now.Sub(checkInTime.Time).Hours()
	record.TotalHours = &totalHoursFloat

	// Calculate overtime (assuming 8 hours is standard)
	if totalHoursFloat > 8.0 {
		record.OvertimeHours = totalHoursFloat - 8.0
	}

	// Combine notes if provided
	if notes != "" {
		if existingNotes.Valid && existingNotes.String != "" {
			combinedNotes := existingNotes.String + "; " + notes
			record.Notes = &combinedNotes
		} else {
			record.Notes = &notes
		}
	} else if existingNotes.Valid {
		record.Notes = &existingNotes.String
	}

	record.UpdatedAt = now

	// Update record
	_, err = s.db.ExecContext(ctx, `
		UPDATE attendance_records 
		SET check_out_time = $1, total_hours = $2, overtime_hours = $3, notes = $4, updated_at = $5
		WHERE id = $6`,
		record.CheckOutTime, record.TotalHours, record.OvertimeHours, record.Notes, record.UpdatedAt, record.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to save check-out record: %w", err)
	}

	return &record, nil
}

// GetTodayAttendance gets today's attendance records
func (s *AttendanceService) GetTodayAttendance(ctx context.Context, tenantID uuid.UUID) ([]models.AttendanceRecord, error) {
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.source, ar.device_id, ar.biometric_log_id,
		       ar.created_at, ar.updated_at,
		       u.first_name, u.last_name,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND ar.date = $2
		ORDER BY ar.check_in_time DESC NULLS LAST`

	rows, err := s.db.QueryContext(ctx, query, tenantID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query today's attendance: %w", err)
	}
	defer rows.Close()

	records := make([]models.AttendanceRecord, 0)
	for rows.Next() {
		var record models.AttendanceRecord
		var checkInTime, checkOutTime, approvedAt sql.NullTime
		var totalHours sql.NullFloat64
		var approvedBy sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
			&checkInTime, &checkOutTime, &record.BreakDurationMinutes,
			&totalHours, &record.OvertimeHours, &record.Status,
			&record.IsApproved, &approvedBy, &approvedAt,
			&notes, &record.CreatedAt, &record.UpdatedAt,
			&record.FirstName, &record.LastName, &record.EmployeeName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}

		// Handle nullable fields
		if checkInTime.Valid {
			record.CheckInTime = &checkInTime.Time
		}
		if checkOutTime.Valid {
			record.CheckOutTime = &checkOutTime.Time
		}
		if totalHours.Valid {
			record.TotalHours = &totalHours.Float64
		}
		if approvedBy.Valid {
			approvedByUUID := uuid.MustParse(approvedBy.String)
			record.ApprovedBy = &approvedByUUID
		}
		if approvedAt.Valid {
			record.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			record.Notes = &notes.String
		}

		records = append(records, record)
	}

	return records, nil
}

// GetAttendanceStats gets attendance statistics
func (s *AttendanceService) GetAttendanceStats(ctx context.Context, tenantID uuid.UUID) (*models.AttendanceStats, error) {
	today := time.Now().Format("2006-01-02")

	stats := &models.AttendanceStats{}

	// Total employees
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM employees WHERE tenant_id = $1 AND employment_status = 'active'", tenantID).Scan(&stats.TotalEmployees)
	if err != nil {
		return nil, fmt.Errorf("failed to get total employees: %w", err)
	}

	// Today's stats
	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM attendance_records WHERE tenant_id = $1 AND date = $2 AND status IN ('present', 'late')",
		tenantID, today).Scan(&stats.PresentToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get present count: %w", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM attendance_records WHERE tenant_id = $1 AND date = $2 AND status = 'late'",
		tenantID, today).Scan(&stats.LateToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get late count: %w", err)
	}

	stats.AbsentToday = stats.TotalEmployees - stats.PresentToday

	// Average hours this week
	weekStart := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(total_hours), 0) FROM attendance_records 
		 WHERE tenant_id = $1 AND date >= $2 AND total_hours IS NOT NULL`,
		tenantID, weekStart).Scan(&stats.AverageHoursWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get average hours: %w", err)
	}

	// Simple attendance rate calculation
	if stats.TotalEmployees > 0 {
		stats.AttendanceRate = float64(stats.PresentToday) / float64(stats.TotalEmployees) * 100
	}

	return stats, nil
}

// GetEmployeeCurrentStatus gets current check-in status for an employee
func (s *AttendanceService) GetEmployeeCurrentStatus(ctx context.Context, tenantID, employeeID uuid.UUID) (*models.AttendanceRecord, error) {
	today := time.Now().Format("2006-01-02")

	var record models.AttendanceRecord
	var checkInTime, checkOutTime sql.NullTime
	var totalHours sql.NullFloat64
	var notes sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		        ar.check_out_time, ar.total_hours, ar.overtime_hours, ar.status, 
		        ar.notes, ar.created_at, ar.updated_at,
		        u.first_name, u.last_name,
		        CONCAT(u.first_name, ' ', u.last_name) as employee_name
		 FROM attendance_records ar
		 JOIN employees e ON ar.employee_id = e.id
		 JOIN users u ON e.user_id = u.id
		 WHERE ar.tenant_id = $1 AND ar.employee_id = $2 AND ar.date = $3`,
		tenantID, employeeID, today).Scan(
		&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
		&checkInTime, &checkOutTime, &totalHours, &record.OvertimeHours,
		&record.Status, &notes, &record.CreatedAt, &record.UpdatedAt,
		&record.FirstName, &record.LastName, &record.EmployeeName)

	if err == sql.ErrNoRows {
		// Return empty record if no attendance today
		return &models.AttendanceRecord{
			TenantID:   tenantID,
			EmployeeID: employeeID,
			Date:       time.Now(),
			Status:     "absent",
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get employee status: %w", err)
	}

	// Handle nullable fields
	if checkInTime.Valid {
		record.CheckInTime = &checkInTime.Time
	}
	if checkOutTime.Valid {
		record.CheckOutTime = &checkOutTime.Time
	}
	if totalHours.Valid {
		record.TotalHours = &totalHours.Float64
	}
	if notes.Valid {
		record.Notes = &notes.String
	}

	return &record, nil
}

// GetAttendanceRecords gets attendance records for all employees with pagination and filters
func (s *AttendanceService) GetAttendanceRecords(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, page, limit int, search, status string) ([]models.AttendanceRecord, int, error) {
	offset := (page - 1) * limit

	// Base Query
	query := `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.created_at, ar.updated_at,
		       u.first_name, u.last_name, u.role,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND ar.date >= $2 AND ar.date <= $3`

	args := []interface{}{tenantID, startDate, endDate}
	argIdx := 4

	// Add Filters
	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND ar.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query += fmt.Sprintf(" AND (LOWER(u.first_name) LIKE $%d OR LOWER(u.last_name) LIKE $%d OR LOWER(u.email) LIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, searchTerm) // same arg for all 3 placeholders? No, postgres driver might need separate?
		// Actually, standard sql injection protection re-uses index if supported, but sql/database usually expects arg per placeholder.
		// Simpler: use named args or just append same arg 3 times?
		// Or better: use ILIKE and one placeholder?
		// Postgres supports $N param reuse.
		// "AND (u.first_name ILIKE $4 OR u.last_name ILIKE $4 OR u.email ILIKE $4)"
		// Let's retry with reusable index
	}

	// Re-construct query with cleaner logic for Search
	query = `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.created_at, ar.updated_at,
		       u.first_name, u.last_name, u.role,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND ar.date >= $2 AND ar.date <= $3`

	args = []interface{}{tenantID, startDate, endDate}
	argIdx = 4

	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND ar.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		// Using ILIKE for case-insensitive search in Postgres
		query += fmt.Sprintf(" AND (u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, searchTerm)
		argIdx++
	}

	// Add Order and Pagination
	query += fmt.Sprintf(" ORDER BY ar.date DESC, u.first_name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query attendance records: %w", err)
	}
	defer rows.Close()

	var records []models.AttendanceRecord
	for rows.Next() {
		var record models.AttendanceRecord
		var checkInTime, checkOutTime, approvedAt sql.NullTime
		var totalHours sql.NullFloat64
		var approvedBy sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
			&checkInTime, &checkOutTime, &record.BreakDurationMinutes,
			&totalHours, &record.OvertimeHours, &record.Status,
			&record.IsApproved, &approvedBy, &approvedAt,
			&notes, &record.CreatedAt, &record.UpdatedAt,
			&record.FirstName, &record.LastName, &record.Role, &record.EmployeeName)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan attendance record: %w", err)
		}

		// Handle nullable fields
		if checkInTime.Valid {
			record.CheckInTime = &checkInTime.Time
		}
		if checkOutTime.Valid {
			record.CheckOutTime = &checkOutTime.Time
		}
		if totalHours.Valid {
			record.TotalHours = &totalHours.Float64
		}
		if approvedBy.Valid {
			approvedByUUID := uuid.MustParse(approvedBy.String)
			record.ApprovedBy = &approvedByUUID
		}
		if approvedAt.Valid {
			record.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			record.Notes = &notes.String
		}

		records = append(records, record)
	}

	// Get total count (using same filters)
	countQuery := `
		SELECT COUNT(*) 
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND ar.date >= $2 AND ar.date <= $3`

	countArgs := []interface{}{tenantID, startDate, endDate}
	countArgIdx := 4

	if status != "" && status != "all" {
		countQuery += fmt.Sprintf(" AND ar.status = $%d", countArgIdx)
		countArgs = append(countArgs, status)
		countArgIdx++
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		countQuery += fmt.Sprintf(" AND (u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d)", countArgIdx, countArgIdx, countArgIdx)
		countArgs = append(countArgs, searchTerm)
		countArgIdx++
	}

	var total int
	err = s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return records, total, nil
}

// GetEmployeeAttendance gets attendance records for a specific employee
func (s *AttendanceService) GetEmployeeAttendance(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) ([]models.AttendanceRecord, error) {
	query := `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.created_at, ar.updated_at,
		       u.first_name, u.last_name,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND ar.employee_id = $2 
		  AND ar.date >= $3 AND ar.date <= $4
		ORDER BY ar.date DESC`

	rows, err := s.db.QueryContext(ctx, query, tenantID, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query employee attendance: %w", err)
	}
	defer rows.Close()

	var records []models.AttendanceRecord
	for rows.Next() {
		var record models.AttendanceRecord
		var checkInTime, checkOutTime, approvedAt sql.NullTime
		var totalHours sql.NullFloat64
		var approvedBy sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
			&checkInTime, &checkOutTime, &record.BreakDurationMinutes,
			&totalHours, &record.OvertimeHours, &record.Status,
			&record.IsApproved, &approvedBy, &approvedAt,
			&notes, &record.CreatedAt, &record.UpdatedAt,
			&record.FirstName, &record.LastName, &record.EmployeeName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}

		// Handle nullable fields
		if checkInTime.Valid {
			record.CheckInTime = &checkInTime.Time
		}
		if checkOutTime.Valid {
			record.CheckOutTime = &checkOutTime.Time
		}
		if totalHours.Valid {
			record.TotalHours = &totalHours.Float64
		}
		if approvedBy.Valid {
			approvedByUUID := uuid.MustParse(approvedBy.String)
			record.ApprovedBy = &approvedByUUID
		}
		if approvedAt.Valid {
			record.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			record.Notes = &notes.String
		}

		records = append(records, record)
	}

	return records, nil
}

// GetTeamAttendance gets attendance records for a specific team
func (s *AttendanceService) GetTeamAttendance(ctx context.Context, tenantID, teamID uuid.UUID, startDate, endDate time.Time) ([]models.AttendanceRecord, error) {
	query := `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.created_at, ar.updated_at,
		       u.first_name, u.last_name,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND u.team_id = $2
		  AND ar.date >= $3 AND ar.date <= $4
		ORDER BY ar.date DESC, u.first_name ASC`

	rows, err := s.db.QueryContext(ctx, query, tenantID, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query team attendance: %w", err)
	}
	defer rows.Close()

	var records []models.AttendanceRecord
	for rows.Next() {
		var record models.AttendanceRecord
		var checkInTime, checkOutTime, approvedAt sql.NullTime
		var totalHours sql.NullFloat64
		var approvedBy sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
			&checkInTime, &checkOutTime, &record.BreakDurationMinutes,
			&totalHours, &record.OvertimeHours, &record.Status,
			&record.IsApproved, &approvedBy, &approvedAt,
			&notes, &record.CreatedAt, &record.UpdatedAt,
			&record.FirstName, &record.LastName, &record.EmployeeName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}

		// Handle nullable fields
		if checkInTime.Valid {
			record.CheckInTime = &checkInTime.Time
		}
		if checkOutTime.Valid {
			record.CheckOutTime = &checkOutTime.Time
		}
		if totalHours.Valid {
			record.TotalHours = &totalHours.Float64
		}
		if approvedBy.Valid {
			approvedByUUID := uuid.MustParse(approvedBy.String)
			record.ApprovedBy = &approvedByUUID
		}
		if approvedAt.Valid {
			record.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			record.Notes = &notes.String
		}

		records = append(records, record)
	}

	return records, nil
}

// GetDepartmentAttendance gets attendance records for a specific department
func (s *AttendanceService) GetDepartmentAttendance(ctx context.Context, tenantID, departmentID uuid.UUID, startDate, endDate time.Time) ([]models.AttendanceRecord, error) {
	query := `
		SELECT ar.id, ar.tenant_id, ar.employee_id, ar.date, ar.check_in_time, 
		       ar.check_out_time, ar.break_duration_minutes, ar.total_hours, 
		       ar.overtime_hours, ar.status, ar.is_approved, ar.approved_by, 
		       ar.approved_at, ar.notes, ar.created_at, ar.updated_at,
		       u.first_name, u.last_name,
		       CONCAT(u.first_name, ' ', u.last_name) as employee_name
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 AND e.department_id = $2
		  AND ar.date >= $3 AND ar.date <= $4
		ORDER BY ar.date DESC, u.first_name ASC`

	rows, err := s.db.QueryContext(ctx, query, tenantID, departmentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query department attendance: %w", err)
	}
	defer rows.Close()

	var records []models.AttendanceRecord
	for rows.Next() {
		var record models.AttendanceRecord
		var checkInTime, checkOutTime, approvedAt sql.NullTime
		var totalHours sql.NullFloat64
		var approvedBy sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.Date,
			&checkInTime, &checkOutTime, &record.BreakDurationMinutes,
			&totalHours, &record.OvertimeHours, &record.Status,
			&record.IsApproved, &approvedBy, &approvedAt,
			&notes, &record.CreatedAt, &record.UpdatedAt,
			&record.FirstName, &record.LastName, &record.EmployeeName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}

		// Handle nullable fields
		if checkInTime.Valid {
			record.CheckInTime = &checkInTime.Time
		}
		if checkOutTime.Valid {
			record.CheckOutTime = &checkOutTime.Time
		}
		if totalHours.Valid {
			record.TotalHours = &totalHours.Float64
		}
		if approvedBy.Valid {
			approvedByUUID := uuid.MustParse(approvedBy.String)
			record.ApprovedBy = &approvedByUUID
		}
		if approvedAt.Valid {
			record.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			record.Notes = &notes.String
		}

		records = append(records, record)
	}

	return records, nil
}

// UpdateAttendanceRecord updates an attendance record
func (s *AttendanceService) UpdateAttendanceRecord(ctx context.Context, tenantID, recordID uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	// Add updates
	for field, value := range updates {
		switch field {
		case "status", "notes", "check_in_time", "check_out_time", "total_hours", "overtime_hours":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE conditions
	args = append(args, recordID, tenantID)

	// Build final query
	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE attendance_records 
		SET %s 
		WHERE id = $%d AND tenant_id = $%d`,
		setClause, argIndex, argIndex+1)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update attendance record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attendance record not found")
	}

	return nil
}

// GetEmployeeIDByUserID gets the employee ID for a given user ID
func (s *AttendanceService) GetEmployeeIDByUserID(ctx context.Context, tenantID, userID uuid.UUID) (uuid.UUID, error) {
	var employeeID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM employees WHERE tenant_id = $1 AND user_id = $2",
		tenantID, userID).Scan(&employeeID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get employee ID for user: %w", err)
	}

	return employeeID, nil
}

// CreateAttendancePolicy creates a new attendance policy or revives a soft-deleted one
func (s *AttendanceService) CreateAttendancePolicy(ctx context.Context, tenantID uuid.UUID, req models.AttendancePolicy) (*models.AttendancePolicy, error) {
	// Check for existing policy by name
	var existingID string
	var existingDeletedAt *time.Time

	// Assuming 'attendance_policies' table has columns: id, tenant_id, name, ... deleted_at
	checkQuery := `
		SELECT id, deleted_at 
		FROM attendance_policies 
		WHERE tenant_id = $1 AND name = $2
	`
	err := s.db.QueryRowContext(ctx, checkQuery, tenantID, req.Name).Scan(&existingID, &existingDeletedAt)

	if err == nil {
		if existingDeletedAt == nil {
			return nil, fmt.Errorf("attendance policy with name '%s' already exists", req.Name)
		}

		// Revive soft-deleted policy
		reviveQuery := `
			UPDATE attendance_policies
			SET working_hours_per_day = $1,
			    working_days = $2,
			    grace_period_minutes = $3,
			    break_duration_minutes = $4,
			    overtime_threshold_minutes = $5,
			    is_default = $6,
			    is_active = $7,
			    deleted_at = NULL,
			    updated_at = $8
			WHERE id = $9
		`

		// Marshal working days to JSON
		// existing code struct likely has WorkingDays as []string or similar that needs DB driver support or manual marshaling
		// Ideally we use a helper or the model already handles it. Assuming direct pass for now but will verify model.

		_, err = s.db.ExecContext(ctx, reviveQuery,
			req.WorkingHoursPerDay, req.WorkingDays, req.GracePeriodMinutes,
			req.BreakDurationMinutes, req.OvertimeThresholdMinutes, req.IsDefault,
			req.IsActive, time.Now(), existingID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to revive attendance policy: %w", err)
		}

		// Return refreshed object (stub, would normally fetch)
		req.ID = uuid.MustParse(existingID)
		req.TenantID = tenantID
		req.UpdatedAt = time.Now()
		return &req, nil

	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing attendance policy: %w", err)
	}

	// Create new policy
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO attendance_policies (
			id, tenant_id, name, working_hours_per_day, working_days,
			grace_period_minutes, break_duration_minutes, overtime_threshold_minutes,
			is_default, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err = s.db.ExecContext(ctx, query,
		id, tenantID, req.Name, req.WorkingHoursPerDay, req.WorkingDays,
		req.GracePeriodMinutes, req.BreakDurationMinutes, req.OvertimeThresholdMinutes,
		req.IsDefault, req.IsActive, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create attendance policy: %w", err)
	}

	req.ID = id
	req.TenantID = tenantID
	req.CreatedAt = now
	req.UpdatedAt = now

	return &req, nil
}
