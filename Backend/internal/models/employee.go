package models

import (
	"time"

	"github.com/google/uuid"
)

// Employee represents an employee in the system
type Employee struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	TenantID          uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	UserID            *uuid.UUID  `json:"user_id,omitempty" db:"user_id"`
	EmployeeCode      string      `json:"employee_code" db:"employee_code"`
	DepartmentID      *uuid.UUID  `json:"department_id,omitempty" db:"department_id"`
	ManagerID         *uuid.UUID  `json:"manager_id,omitempty" db:"manager_id"`
	JobTitle          *string     `json:"job_title,omitempty" db:"job_title"`
	EmploymentType    *string     `json:"employment_type,omitempty" db:"employment_type"`
	EmploymentStatus  string      `json:"employment_status" db:"employment_status"`
	DateOfJoining     *time.Time  `json:"date_of_joining,omitempty" db:"date_of_joining"`
	DateOfLeaving     *time.Time  `json:"date_of_leaving,omitempty" db:"date_of_leaving"`
	Salary            *float64    `json:"salary,omitempty" db:"salary"`
	Phone             *string     `json:"phone,omitempty" db:"phone"`
	NationalID        *string     `json:"national_id,omitempty" db:"national_id"`
	BankAccountNumber *string     `json:"bank_account_number,omitempty" db:"bank_account_number"`
	NationalIDHash    *string     `json:"-" db:"national_id_hash"` // Blind index, never returned to API
	EmergencyContact  interface{} `json:"emergency_contact,omitempty" db:"emergency_contact"`
	Address           interface{} `json:"address,omitempty" db:"address"`
	PersonalDetails   interface{} `json:"personal_details,omitempty" db:"personal_details"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`

	// Derived fields (from joins or computed)
	FirstName  string      `json:"first_name,omitempty"`
	LastName   string      `json:"last_name,omitempty"`
	Email      string      `json:"email,omitempty"`
	Role       string      `json:"role,omitempty"`
	Department *Department `json:"department,omitempty"`
}

// Department represents a department in the system
type Department struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	HeadID      *uuid.UUID `json:"head_id,omitempty" db:"head_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
