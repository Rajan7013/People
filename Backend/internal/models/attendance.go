package models

import (
	"time"

	"github.com/google/uuid"
)

// AttendancePolicy represents an attendance policy for a tenant
type AttendancePolicy struct {
	ID                       uuid.UUID  `json:"id" db:"id"`
	TenantID                 uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name                     string     `json:"name" db:"name"`
	WorkingHoursPerDay       float64    `json:"working_hours_per_day" db:"working_hours_per_day"`
	WorkingDays              []string   `json:"working_days" db:"working_days"`
	GracePeriodMinutes       int        `json:"grace_period_minutes" db:"grace_period_minutes"`
	BreakDurationMinutes     int        `json:"break_duration_minutes" db:"break_duration_minutes"`
	OvertimeThresholdMinutes int        `json:"overtime_threshold_minutes" db:"overtime_threshold_minutes"`
	IsDefault                bool       `json:"is_default" db:"is_default"`
	IsActive                 bool       `json:"is_active" db:"is_active"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt                *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// AttendanceRecord represents an employee's attendance for a specific date
type AttendanceRecord struct {
	ID                   uuid.UUID        `json:"id" db:"id"`
	TenantID             uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	EmployeeID           uuid.UUID        `json:"employee_id" db:"employee_id"`
	Date                 time.Time        `json:"date" db:"date"`
	CheckInTime          *time.Time       `json:"check_in_time" db:"check_in_time"`
	CheckOutTime         *time.Time       `json:"check_out_time" db:"check_out_time"`
	BreakDurationMinutes int              `json:"break_duration_minutes" db:"break_duration_minutes"`
	TotalHours           *float64         `json:"total_hours" db:"total_hours"`
	OvertimeHours        float64          `json:"overtime_hours" db:"overtime_hours"`
	Status               string           `json:"status" db:"status"`
	IsApproved           bool             `json:"is_approved" db:"is_approved"`
	ApprovedBy           *uuid.UUID       `json:"approved_by" db:"approved_by"`
	ApprovedAt           *time.Time       `json:"approved_at" db:"approved_at"`
	Notes                *string          `json:"notes" db:"notes"`
	Source               AttendanceSource `json:"source" db:"source"`
	DeviceID             *uuid.UUID       `json:"device_id,omitempty" db:"device_id"`
	BiometricLogID       *uuid.UUID       `json:"biometric_log_id,omitempty" db:"biometric_log_id"`
	CreatedAt            time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" db:"updated_at"`

	// Joined fields for API responses
	EmployeeName string `json:"employee_name,omitempty" db:"employee_name"`
	FirstName    string `json:"first_name,omitempty" db:"first_name"`
	LastName     string `json:"last_name,omitempty" db:"last_name"`
	Role         string `json:"role,omitempty" db:"role"`
}

// CheckInRequest represents a check-in request
type CheckInRequest struct {
	EmployeeID string `json:"employee_id,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// CheckOutRequest represents a check-out request
type CheckOutRequest struct {
	EmployeeID string `json:"employee_id,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// AttendanceStats represents attendance statistics
type AttendanceStats struct {
	TotalEmployees   int     `json:"total_employees"`
	PresentToday     int     `json:"present_today"`
	AbsentToday      int     `json:"absent_today"`
	LateToday        int     `json:"late_today"`
	AverageHoursWeek float64 `json:"average_hours_week"`
	AttendanceRate   float64 `json:"attendance_rate"`
}

// AttendanceSummary represents a summary for an employee
type AttendanceSummary struct {
	EmployeeID         uuid.UUID `json:"employee_id"`
	EmployeeName       string    `json:"employee_name"`
	TotalDaysPresent   int       `json:"total_days_present"`
	TotalDaysAbsent    int       `json:"total_days_absent"`
	TotalLateDays      int       `json:"total_late_days"`
	TotalHoursWorked   float64   `json:"total_hours_worked"`
	TotalOvertimeHours float64   `json:"total_overtime_hours"`
	AttendanceRate     float64   `json:"attendance_rate"`
}
