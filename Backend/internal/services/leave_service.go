package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"

	"github.com/google/uuid"
)

type LeaveService struct {
	db *sql.DB
}

func NewLeaveService(db *sql.DB) *LeaveService {
	return &LeaveService{db: db}
}

// CreateLeaveRequest creates a new leave request
func (s *LeaveService) CreateLeaveRequest(ctx context.Context, tenantID, employeeID uuid.UUID, req models.CreateLeaveRequest) (*models.LeaveRequest, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	// Validate dates
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}

	// Create leave request
	leaveRequest := models.LeaveRequest{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmployeeID:    employeeID,
		LeaveType:     req.LeaveType,
		StartDate:     startDate,
		EndDate:       endDate,
		DaysRequested: req.DaysRequested,
		Reason:        req.Reason,
		Status:        models.LeaveStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO leave_requests (
			id, tenant_id, employee_id, leave_type, start_date, end_date, 
			days_requested, reason, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err = s.db.ExecContext(ctx, query,
		leaveRequest.ID, leaveRequest.TenantID, leaveRequest.EmployeeID,
		leaveRequest.LeaveType, leaveRequest.StartDate, leaveRequest.EndDate,
		leaveRequest.DaysRequested, leaveRequest.Reason, leaveRequest.Status,
		leaveRequest.CreatedAt, leaveRequest.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave request: %w", err)
	}

	return &leaveRequest, nil
}

// GetLeaveRequests retrieves leave requests with optional filters
func (s *LeaveService) GetLeaveRequests(ctx context.Context, tenantID uuid.UUID, employeeID *uuid.UUID, departmentID *uuid.UUID, status *models.LeaveStatus) ([]models.LeaveRequest, error) {
	query := `
		SELECT 
			lr.id, lr.tenant_id, lr.employee_id, lr.leave_type, lr.start_date, lr.end_date,
			lr.days_requested, lr.reason, lr.status, lr.approved_by, lr.approved_at,
			lr.rejection_reason, lr.created_at, lr.updated_at,
			(u.first_name || ' ' || u.last_name) as employee_name, 
			e.employee_code as employee_code, u.role,
			COALESCE((a_user.first_name || ' ' || a_user.last_name), '') as approver_name
		FROM leave_requests lr
		JOIN employees e ON lr.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN employees a_emp ON lr.approved_by = a_emp.id
		LEFT JOIN users a_user ON a_emp.user_id = a_user.id
		WHERE lr.tenant_id = $1
	`

	args := []interface{}{tenantID}
	argCount := 1

	if employeeID != nil {
		argCount++
		query += fmt.Sprintf(" AND lr.employee_id = $%d", argCount)
		args = append(args, *employeeID)
	}

	if departmentID != nil {
		argCount++
		query += fmt.Sprintf(" AND e.department_id = $%d", argCount)
		args = append(args, *departmentID)
	}

	if status != nil {
		argCount++
		query += fmt.Sprintf(" AND lr.status = $%d", argCount)
		args = append(args, *status)
	}

	query += " ORDER BY lr.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave requests: %w", err)
	}
	defer rows.Close()

	leaves := []models.LeaveRequest{}
	for rows.Next() {
		var leave models.LeaveRequest
		err := rows.Scan(
			&leave.ID, &leave.TenantID, &leave.EmployeeID, &leave.LeaveType,
			&leave.StartDate, &leave.EndDate, &leave.DaysRequested, &leave.Reason,
			&leave.Status, &leave.ApprovedBy, &leave.ApprovedAt, &leave.RejectionReason,
			&leave.CreatedAt, &leave.UpdatedAt, &leave.EmployeeName, &leave.EmployeeCode, &leave.Role,
			&leave.ApproverName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave request: %w", err)
		}
		leaves = append(leaves, leave)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return leaves, nil
}

// GetLeaveRequestByID retrieves a single leave request by ID
func (s *LeaveService) GetLeaveRequestByID(ctx context.Context, tenantID, leaveID uuid.UUID) (*models.LeaveRequest, error) {
	query := `
		SELECT 
			lr.id, lr.tenant_id, lr.employee_id, lr.leave_type, lr.start_date, lr.end_date,
			lr.days_requested, lr.reason, lr.status, lr.approved_by, lr.approved_at,
			lr.rejection_reason, lr.created_at, lr.updated_at,
			(u.first_name || ' ' || u.last_name) as employee_name, 
			e.employee_code as employee_code, u.role,
			COALESCE((a_user.first_name || ' ' || a_user.last_name), '') as approver_name
		FROM leave_requests lr
		JOIN employees e ON lr.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN employees a_emp ON lr.approved_by = a_emp.id
		LEFT JOIN users a_user ON a_emp.user_id = a_user.id
		WHERE lr.id = $1 AND lr.tenant_id = $2
	`

	var leave models.LeaveRequest
	err := s.db.QueryRowContext(ctx, query, leaveID, tenantID).Scan(
		&leave.ID, &leave.TenantID, &leave.EmployeeID, &leave.LeaveType,
		&leave.StartDate, &leave.EndDate, &leave.DaysRequested, &leave.Reason,
		&leave.Status, &leave.ApprovedBy, &leave.ApprovedAt, &leave.RejectionReason,
		&leave.CreatedAt, &leave.UpdatedAt, &leave.EmployeeName, &leave.EmployeeCode, &leave.Role,
		&leave.ApproverName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("leave request not found")
		}
		return nil, fmt.Errorf("failed to get leave request: %w", err)
	}

	return &leave, nil
}

// ApproveLeaveRequest approves a leave request
func (s *LeaveService) ApproveLeaveRequest(ctx context.Context, tenantID, leaveID, approverID uuid.UUID) error {
	// Get leave request first
	var status models.LeaveStatus
	query := `
		SELECT status
		FROM leave_requests
		WHERE id = $1 AND tenant_id = $2
	`
	err := s.db.QueryRowContext(ctx, query, leaveID, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("leave request not found")
		}
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// Check if already approved/rejected
	if status != models.LeaveStatusPending {
		return fmt.Errorf("leave request is already %s", status)
	}

	// Update leave request status
	now := time.Now()
	updateQuery := `
		UPDATE leave_requests
		SET status = $1, approved_by = $2, approved_at = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = s.db.ExecContext(ctx, updateQuery, models.LeaveStatusApproved, approverID, now, now, leaveID)
	if err != nil {
		return fmt.Errorf("failed to approve leave request: %w", err)
	}

	return nil
}

// RejectLeaveRequest rejects a leave request
func (s *LeaveService) RejectLeaveRequest(ctx context.Context, tenantID, leaveID, approverID uuid.UUID, reason string) error {
	// Get leave request first
	var status models.LeaveStatus
	query := `
		SELECT status
		FROM leave_requests
		WHERE id = $1 AND tenant_id = $2
	`
	err := s.db.QueryRowContext(ctx, query, leaveID, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("leave request not found")
		}
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// Check if already approved/rejected
	if status != models.LeaveStatusPending {
		return fmt.Errorf("leave request is already %s", status)
	}

	// Update leave request status
	now := time.Now()
	updateQuery := `
		UPDATE leave_requests
		SET status = $1, approved_by = $2, approved_at = $3, rejection_reason = $4, updated_at = $5
		WHERE id = $6
	`
	_, err = s.db.ExecContext(ctx, updateQuery,
		models.LeaveStatusRejected, approverID, now, reason, now, leaveID,
	)
	if err != nil {
		return fmt.Errorf("failed to reject leave request: %w", err)
	}

	return nil
}

// GetPendingLeaveRequests retrieves all pending leave requests (for managers)
func (s *LeaveService) GetPendingLeaveRequests(ctx context.Context, tenantID uuid.UUID) ([]models.LeaveRequest, error) {
	status := models.LeaveStatusPending
	return s.GetLeaveRequests(ctx, tenantID, nil, nil, &status)
}

// GetEmployeeIDByUserID retrieves the employee ID associated with a user ID
func (s *LeaveService) GetEmployeeIDByUserID(ctx context.Context, tenantID, userID uuid.UUID) (uuid.UUID, error) {
	var employeeID uuid.UUID
	query := `
		SELECT id
		FROM employees
		WHERE user_id = $1 AND tenant_id = $2
	`
	err := s.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&employeeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("employee record not found for user")
		}
		return uuid.Nil, fmt.Errorf("failed to get employee ID: %w", err)
	}

	return employeeID, nil
}

// CreateLeaveType creates a new leave type or revives a soft-deleted one
func (s *LeaveService) CreateLeaveType(ctx context.Context, tenantID uuid.UUID, req models.LeaveTypeConfig) (*models.LeaveTypeConfig, error) {
	// Check for existing leave type by name or short code
	var existingID string
	var existingDeletedAt *time.Time

	checkQuery := `
		SELECT id, deleted_at 
		FROM leave_types 
		WHERE tenant_id = $1 AND (name = $2 OR short_code = $3)
	`
	err := s.db.QueryRowContext(ctx, checkQuery, tenantID, req.Name, req.ShortCode).Scan(&existingID, &existingDeletedAt)

	if err == nil {
		if existingDeletedAt == nil {
			return nil, fmt.Errorf("leave type with name '%s' or short code '%s' already exists", req.Name, req.ShortCode)
		}

		// Revive soft-deleted leave type
		reviveQuery := `
			UPDATE leave_types
			SET is_active = $1, 
			    annual_limit = $2, 
			    accrual_type = $3, 
			    accrual_amount = $4, 
			    carryforward_limit = $5, 
			    is_paid = $6, 
			    requires_approval = $7,
			    deleted_at = NULL,
			    updated_at = $8
			WHERE id = $9
		`
		_, err = s.db.ExecContext(ctx, reviveQuery,
			req.IsActive, req.AnnualLimit, req.AccrualType, req.AccrualAmount,
			req.CarryforwardLimit, req.IsPaid, req.RequiresApproval,
			time.Now(), existingID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to revive leave type: %w", err)
		}

		// Return refreshed object
		return s.GetLeaveTypeByID(ctx, tenantID, uuid.MustParse(existingID))

	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing leave type: %w", err)
	}

	// Create new leave type
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO leave_types (
			id, tenant_id, name, short_code, annual_limit, accrual_type, 
			accrual_amount, carryforward_limit, is_paid, requires_approval, 
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = s.db.ExecContext(ctx, query,
		id, tenantID, req.Name, req.ShortCode, req.AnnualLimit, req.AccrualType,
		req.AccrualAmount, req.CarryforwardLimit, req.IsPaid, req.RequiresApproval,
		req.IsActive, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create leave type: %w", err)
	}

	req.ID = id
	req.TenantID = tenantID
	req.CreatedAt = now
	req.UpdatedAt = now

	return &req, nil
}

// GetLeaveTypeByID gets a specific leave type
func (s *LeaveService) GetLeaveTypeByID(ctx context.Context, tenantID, id uuid.UUID) (*models.LeaveTypeConfig, error) {
	query := `
		SELECT id, tenant_id, name, short_code, annual_limit, accrual_type,
		       accrual_amount, carryforward_limit, is_paid, requires_approval,
		       is_active, created_at, updated_at
		FROM leave_types
		WHERE id = $1 AND tenant_id = $2
	`
	var lt models.LeaveTypeConfig
	err := s.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&lt.ID, &lt.TenantID, &lt.Name, &lt.ShortCode, &lt.AnnualLimit, &lt.AccrualType,
		&lt.AccrualAmount, &lt.CarryforwardLimit, &lt.IsPaid, &lt.RequiresApproval,
		&lt.IsActive, &lt.CreatedAt, &lt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("leave type not found")
		}
		return nil, fmt.Errorf("failed to get leave type: %w", err)
	}

	return &lt, nil
}
