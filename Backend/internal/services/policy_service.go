package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type PolicyService struct {
	db *sql.DB
}

func NewPolicyService(db *sql.DB) *PolicyService {
	return &PolicyService{db: db}
}

// Request structs
type AttendancePolicyRequest struct {
	WorkingHoursPerDay   *float64 `json:"working_hours_per_day"`
	GracePeriodMinutes   *int     `json:"grace_period_minutes"`
	OvertimeThreshold    *int     `json:"overtime_threshold_minutes"`
	RequiredHoursPerWeek *float64 `json:"required_hours_per_week"`
	RequiredDaysPerMonth *int     `json:"required_days_per_month"`
	LateFinePerMinute    *float64 `json:"late_fine_per_minute"`
}

type SalaryComponentRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	IsFixed     bool    `json:"is_fixed"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
	Description string  `json:"description"`
}

type LeaveTypeRequest struct {
	Name              string  `json:"name"`
	AnnualLimit       int     `json:"annual_limit"`
	CarryForwardLimit int     `json:"carry_forward_limit"`
	IsPaid            bool    `json:"is_paid"`
	RequiresApproval  bool    `json:"requires_approval"`
	AccrualRate       float64 `json:"accrual_rate"`
	Description       string  `json:"description"`
}

// AttendancePolicy represents the attendance policy configuration
type AttendancePolicy struct {
	ID                   uuid.UUID `json:"id"`
	TenantID             uuid.UUID `json:"tenant_id"`
	WorkingHoursPerDay   float64   `json:"working_hours_per_day"`
	GracePeriodMinutes   int       `json:"grace_period_minutes"`
	OvertimeThreshold    int       `json:"overtime_threshold_minutes"`
	RequiredHoursPerWeek float64   `json:"required_hours_per_week"`
	RequiredDaysPerMonth int       `json:"required_days_per_month"`
	LateFinePerMinute    float64   `json:"late_fine_per_minute"`
}

// GetAttendancePolicy retrieves the attendance policy for a tenant
// If no policy exists, it returns default values
func (s *PolicyService) GetAttendancePolicy(tenantID uuid.UUID) (*AttendancePolicy, error) {
	var policy AttendancePolicy

	err := s.db.QueryRow(`
		SELECT id, tenant_id, working_hours_per_day, grace_period_minutes, 
		       overtime_threshold_minutes, required_hours_per_week, required_days_per_month
		FROM attendance_policies
		WHERE tenant_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`, tenantID).Scan(
		&policy.ID,
		&policy.TenantID,
		&policy.WorkingHoursPerDay,
		&policy.GracePeriodMinutes,
		&policy.OvertimeThreshold,
		&policy.RequiredHoursPerWeek,
		&policy.RequiredDaysPerMonth,
	)

	if err == sql.ErrNoRows {
		// Return default policy
		return &AttendancePolicy{
			ID:                   uuid.New(),
			TenantID:             tenantID,
			WorkingHoursPerDay:   8.0,
			GracePeriodMinutes:   15,
			OvertimeThreshold:    480, // 8 hours in minutes
			RequiredHoursPerWeek: 40.0,
			RequiredDaysPerMonth: 22,
			LateFinePerMinute:    0.0,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get attendance policy: %w", err)
	}

	// Get late fine from tenant settings
	var settingsJSON sql.NullString
	err = s.db.QueryRow(`
		SELECT settings FROM tenants WHERE id = $1
	`, tenantID).Scan(&settingsJSON)

	if err == nil && settingsJSON.Valid {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON.String), &settings); err == nil {
			if payroll, ok := settings["payroll"].(map[string]interface{}); ok {
				if lateFine, ok := payroll["late_fine_per_minute"].(float64); ok {
					policy.LateFinePerMinute = lateFine
				}
			}
		}
	}

	return &policy, nil
}

// UpdateAttendancePolicy updates the attendance policy for a tenant
func (s *PolicyService) UpdateAttendancePolicy(tenantID uuid.UUID, req AttendancePolicyRequest) (*AttendancePolicy, error) {
	// First, check if policy exists
	var policyID uuid.UUID
	err := s.db.QueryRow(`
		SELECT id FROM attendance_policies WHERE tenant_id = $1 AND deleted_at IS NULL
	`, tenantID).Scan(&policyID)

	if err == sql.ErrNoRows {
		// Create new policy
		policyID = uuid.New()
		_, err = s.db.Exec(`
			INSERT INTO attendance_policies (
				id, tenant_id, working_hours_per_day, grace_period_minutes,
				overtime_threshold_minutes, required_hours_per_week, required_days_per_month
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, policyID, tenantID, 8.0, 15, 480, 40.0, 22)
		if err != nil {
			return nil, fmt.Errorf("failed to create attendance policy: %w", err)
		}
	}

	// Update policy fields if provided
	if req.WorkingHoursPerDay != nil {
		_, err = s.db.Exec(`
			UPDATE attendance_policies SET working_hours_per_day = $1 WHERE id = $2
		`, *req.WorkingHoursPerDay, policyID)
		if err != nil {
			return nil, err
		}
	}

	if req.GracePeriodMinutes != nil {
		_, err = s.db.Exec(`
			UPDATE attendance_policies SET grace_period_minutes = $1 WHERE id = $2
		`, *req.GracePeriodMinutes, policyID)
		if err != nil {
			return nil, err
		}
	}

	if req.OvertimeThreshold != nil {
		_, err = s.db.Exec(`
			UPDATE attendance_policies SET overtime_threshold_minutes = $1 WHERE id = $2
		`, *req.OvertimeThreshold, policyID)
		if err != nil {
			return nil, err
		}
	}

	if req.RequiredHoursPerWeek != nil {
		_, err = s.db.Exec(`
			UPDATE attendance_policies SET required_hours_per_week = $1 WHERE id = $2
		`, *req.RequiredHoursPerWeek, policyID)
		if err != nil {
			return nil, err
		}
	}

	if req.RequiredDaysPerMonth != nil {
		_, err = s.db.Exec(`
			UPDATE attendance_policies SET required_days_per_month = $1 WHERE id = $2
		`, *req.RequiredDaysPerMonth, policyID)
		if err != nil {
			return nil, err
		}
	}

	// Update late fine in tenant settings
	if req.LateFinePerMinute != nil {
		var settingsJSON sql.NullString
		err = s.db.QueryRow(`SELECT settings FROM tenants WHERE id = $1`, tenantID).Scan(&settingsJSON)

		settings := make(map[string]interface{})
		if settingsJSON.Valid {
			json.Unmarshal([]byte(settingsJSON.String), &settings)
		}

		if settings["payroll"] == nil {
			settings["payroll"] = make(map[string]interface{})
		}
		payroll := settings["payroll"].(map[string]interface{})
		payroll["late_fine_per_minute"] = *req.LateFinePerMinute

		settingsBytes, _ := json.Marshal(settings)
		_, err = s.db.Exec(`UPDATE tenants SET settings = $1 WHERE id = $2`, settingsBytes, tenantID)
		if err != nil {
			return nil, err
		}
	}

	return s.GetAttendancePolicy(tenantID)
}

// GetSalaryComponents retrieves all salary components for a tenant
func (s *PolicyService) GetSalaryComponents(tenantID uuid.UUID) ([]*models.SalaryComponent, error) {
	rows, err := s.db.Query(`
		SELECT id, tenant_id, name, type, is_fixed, amount, percentage, description, created_at, updated_at
		FROM salary_components
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY type, name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get salary components: %w", err)
	}
	defer rows.Close()

	var components []*models.SalaryComponent
	for rows.Next() {
		var comp models.SalaryComponent
		err := rows.Scan(
			&comp.ID,
			&comp.TenantID,
			&comp.Name,
			&comp.Type,
			&comp.IsFixed,
			&comp.Amount,
			&comp.Percentage,
			&comp.Description,
			&comp.CreatedAt,
			&comp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		components = append(components, &comp)
	}

	return components, nil
}

// CreateSalaryComponent creates a new salary component
func (s *PolicyService) CreateSalaryComponent(tenantID uuid.UUID, req SalaryComponentRequest) (*models.SalaryComponent, error) {
	componentID := uuid.New()

	_, err := s.db.Exec(`
		INSERT INTO salary_components (
			id, tenant_id, name, type, is_fixed, amount, percentage, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, componentID, tenantID, req.Name, req.Type, req.IsFixed, req.Amount, req.Percentage, req.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to create salary component: %w", err)
	}

	// Fetch and return the created component
	var comp models.SalaryComponent
	err = s.db.QueryRow(`
		SELECT id, tenant_id, name, type, is_fixed, amount, percentage, description, created_at, updated_at
		FROM salary_components
		WHERE id = $1
	`, componentID).Scan(
		&comp.ID,
		&comp.TenantID,
		&comp.Name,
		&comp.Type,
		&comp.IsFixed,
		&comp.Amount,
		&comp.Percentage,
		&comp.Description,
		&comp.CreatedAt,
		&comp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &comp, nil
}

// GetLeaveTypes retrieves all leave types for a tenant
func (s *PolicyService) GetLeaveTypes(tenantID uuid.UUID) ([]*models.PolicyLeaveType, error) {
	rows, err := s.db.Query(`
		SELECT id, tenant_id, name, annual_limit, carry_forward_limit, is_paid, 
		       requires_approval, accrual_rate, description, created_at, updated_at
		FROM leave_types
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave types: %w", err)
	}
	defer rows.Close()

	var leaveTypes []*models.PolicyLeaveType
	for rows.Next() {
		var lt models.PolicyLeaveType
		err := rows.Scan(
			&lt.ID,
			&lt.TenantID,
			&lt.Name,
			&lt.AnnualLimit,
			&lt.CarryForwardLimit,
			&lt.IsPaid,
			&lt.RequiresApproval,
			&lt.AccrualRate,
			&lt.Description,
			&lt.CreatedAt,
			&lt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		leaveTypes = append(leaveTypes, &lt)
	}

	return leaveTypes, nil
}

// CreateLeaveType creates a new leave type
func (s *PolicyService) CreateLeaveType(tenantID uuid.UUID, req LeaveTypeRequest) (*models.PolicyLeaveType, error) {
	leaveTypeID := uuid.New()

	_, err := s.db.Exec(`
		INSERT INTO leave_types (
			id, tenant_id, name, annual_limit, carry_forward_limit, is_paid,
			requires_approval, accrual_rate, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, leaveTypeID, tenantID, req.Name, req.AnnualLimit, req.CarryForwardLimit,
		req.IsPaid, req.RequiresApproval, req.AccrualRate, req.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to create leave type: %w", err)
	}

	// Fetch and return the created leave type
	var lt models.PolicyLeaveType
	err = s.db.QueryRow(`
		SELECT id, tenant_id, name, annual_limit, carry_forward_limit, is_paid,
		       requires_approval, accrual_rate, description, created_at, updated_at
		FROM leave_types
		WHERE id = $1
	`, leaveTypeID).Scan(
		&lt.ID,
		&lt.TenantID,
		&lt.Name,
		&lt.AnnualLimit,
		&lt.CarryForwardLimit,
		&lt.IsPaid,
		&lt.RequiresApproval,
		&lt.AccrualRate,
		&lt.Description,
		&lt.CreatedAt,
		&lt.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &lt, nil
}
