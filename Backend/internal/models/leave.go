package models

import (
	"time"

	"github.com/google/uuid"
)

// LeaveType represents different types of leave
type LeaveType string

const (
	LeaveTypeSick      LeaveType = "sick"
	LeaveTypeCasual    LeaveType = "casual"
	LeaveTypeAnnual    LeaveType = "annual"
	LeaveTypeMaternity LeaveType = "maternity"
	LeaveTypePaternity LeaveType = "paternity"
	LeaveTypeUnpaid    LeaveType = "unpaid"
)

// LeaveTypeConfig represents a leave type configuration
type LeaveTypeConfig struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name              string     `json:"name" db:"name"`
	ShortCode         string     `json:"short_code" db:"short_code"`
	AnnualLimit       *int       `json:"annual_limit" db:"annual_limit"`
	AccrualType       *string    `json:"accrual_type" db:"accrual_type"`
	AccrualAmount     *float64   `json:"accrual_amount" db:"accrual_amount"`
	CarryforwardLimit *int       `json:"carryforward_limit" db:"carryforward_limit"`
	IsPaid            bool       `json:"is_paid" db:"is_paid"`
	RequiresApproval  bool       `json:"requires_approval" db:"requires_approval"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// LeaveStatus represents the status of a leave request
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

// LeaveRequest represents an employee's leave request
type LeaveRequest struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	TenantID        uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	EmployeeID      uuid.UUID   `json:"employee_id" db:"employee_id"`
	LeaveType       LeaveType   `json:"leave_type" db:"leave_type"`
	StartDate       time.Time   `json:"start_date" db:"start_date"`
	EndDate         time.Time   `json:"end_date" db:"end_date"`
	DaysRequested   float64     `json:"days_requested" db:"days_requested"`
	Reason          string      `json:"reason" db:"reason"`
	Status          LeaveStatus `json:"status" db:"status"`
	ApprovedBy      *uuid.UUID  `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt      *time.Time  `json:"approved_at,omitempty" db:"approved_at"`
	RejectionReason *string     `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`

	// Joined fields for API responses
	EmployeeName string `json:"employee_name,omitempty" db:"employee_name"`
	EmployeeCode string `json:"employee_code,omitempty" db:"employee_code"`
	Role         string `json:"role,omitempty" db:"role"`
	ApproverName string `json:"approver_name,omitempty" db:"approver_name"`
}

// CreateLeaveRequest represents a request to create a leave
type CreateLeaveRequest struct {
	LeaveType     LeaveType `json:"leave_type" binding:"required"`
	StartDate     string    `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate       string    `json:"end_date" binding:"required"`   // YYYY-MM-DD
	DaysRequested float64   `json:"days_requested" binding:"required"`
	Reason        string    `json:"reason" binding:"required"`
}

// ApproveLeaveRequest represents a request to approve a leave
type ApproveLeaveRequest struct {
	ApproverNotes string `json:"approver_notes,omitempty"`
}

// RejectLeaveRequest represents a request to reject a leave
type RejectLeaveRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

// LeaveBalance represents an employee's leave balance
type LeaveBalance struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TenantID      uuid.UUID `json:"tenant_id" db:"tenant_id"`
	EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
	LeaveType     LeaveType `json:"leave_type" db:"leave_type"`
	TotalDays     float64   `json:"total_days" db:"total_days"`
	UsedDays      float64   `json:"used_days" db:"used_days"`
	AvailableDays float64   `json:"available_days" db:"available_days"`
	Year          int       `json:"year" db:"year"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
