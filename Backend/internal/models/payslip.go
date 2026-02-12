package models

import (
	"time"

	"github.com/google/uuid"
)

// SalaryComponent represents a salary component (earning or deduction)
type SalaryComponent struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name         string     `json:"name" db:"name"`
	Type         string     `json:"type" db:"type"` // "earning" or "deduction"
	IsFixed      bool       `json:"is_fixed" db:"is_fixed"`
	Amount       float64    `json:"amount" db:"amount"`
	Percentage   float64    `json:"percentage" db:"percentage"`
	Description  string     `json:"description" db:"description"`
	IsPercentage bool       `json:"is_percentage" db:"is_percentage"`
	IsTaxable    bool       `json:"is_taxable" db:"is_taxable"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// EmployeeSalaryStructure represents an employee's salary configuration
type EmployeeSalaryStructure struct {
	ID            uuid.UUID                  `json:"id" db:"id"`
	TenantID      uuid.UUID                  `json:"tenant_id" db:"tenant_id"`
	EmployeeID    uuid.UUID                  `json:"employee_id" db:"employee_id"`
	BasicSalary   float64                    `json:"basic_salary" db:"basic_salary"`
	EffectiveDate time.Time                  `json:"effective_date" db:"effective_date"`
	EndDate       *time.Time                 `json:"end_date,omitempty" db:"end_date"`
	IsActive      bool                       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at" db:"updated_at"`
	Components    []SalaryStructureComponent `json:"components,omitempty"`
	Employee      *Employee                  `json:"employee,omitempty"`
}

// SalaryStructureComponent represents a component in an employee's salary structure
type SalaryStructureComponent struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	SalaryStructureID uuid.UUID        `json:"salary_structure_id" db:"salary_structure_id"`
	ComponentID       uuid.UUID        `json:"component_id" db:"component_id"`
	Amount            *float64         `json:"amount,omitempty" db:"amount"`
	Percentage        *float64         `json:"percentage,omitempty" db:"percentage"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	Component         *SalaryComponent `json:"component,omitempty"`
}

// Payslip represents a payslip record
type Payslip struct {
	ID                uuid.UUID          `json:"id" db:"id"`
	TenantID          uuid.UUID          `json:"tenant_id" db:"tenant_id"`
	EmployeeID        uuid.UUID          `json:"employee_id" db:"employee_id"`
	SalaryStructureID *uuid.UUID         `json:"salary_structure_id,omitempty" db:"salary_structure_id"`
	PayPeriodStart    time.Time          `json:"pay_period_start" db:"pay_period_start"`
	PayPeriodEnd      time.Time          `json:"pay_period_end" db:"pay_period_end"`
	PaymentDate       *time.Time         `json:"payment_date,omitempty" db:"payment_date"`
	BasicSalary       float64            `json:"basic_salary" db:"basic_salary"`
	GrossSalary       float64            `json:"gross_salary" db:"gross_salary"`
	TotalDeductions   float64            `json:"total_deductions" db:"total_deductions"`
	NetSalary         float64            `json:"net_salary" db:"net_salary"`
	Status            string             `json:"status" db:"status"`
	Notes             *string            `json:"notes,omitempty" db:"notes"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
	Employee          *Employee          `json:"employee,omitempty"`
	Components        []PayslipComponent `json:"components,omitempty"`
	Role              string             `json:"role,omitempty" db:"role"`
}

// PayslipComponent represents a detailed breakdown of payslip components
type PayslipComponent struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PayslipID     uuid.UUID `json:"payslip_id" db:"payslip_id"`
	ComponentID   uuid.UUID `json:"component_id" db:"component_id"`
	ComponentName string    `json:"component_name" db:"component_name"`
	ComponentType string    `json:"component_type" db:"component_type"`
	Amount        float64   `json:"amount" db:"amount"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// PayslipCreateRequest represents the request to create a new payslip
type PayslipCreateRequest struct {
	EmployeeID     uuid.UUID  `json:"employee_id" validate:"required"`
	PayPeriodStart time.Time  `json:"pay_period_start" validate:"required"`
	PayPeriodEnd   time.Time  `json:"pay_period_end" validate:"required"`
	PaymentDate    *time.Time `json:"payment_date,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
}

// PayslipUpdateRequest represents the request to update a payslip
type PayslipUpdateRequest struct {
	PaymentDate *time.Time `json:"payment_date,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
}

// PayslipFilter represents filters for payslip queries
type PayslipFilter struct {
	EmployeeID     *uuid.UUID `json:"employee_id,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	Status         *string    `json:"status,omitempty"`
	PayPeriodStart *time.Time `json:"pay_period_start,omitempty"`
	PayPeriodEnd   *time.Time `json:"pay_period_end,omitempty"`
	PaymentMonth   *int       `json:"payment_month,omitempty"`
	PaymentYear    *int       `json:"payment_year,omitempty"`
}

// PayslipStats represents payslip statistics
type PayslipStats struct {
	TotalPayslips     int64   `json:"total_payslips"`
	PendingApproval   int64   `json:"pending_approval"`
	ProcessedPayslips int64   `json:"processed_payslips"`
	TotalPayroll      float64 `json:"total_payroll"`
	AverageSalary     float64 `json:"average_salary"`
}
