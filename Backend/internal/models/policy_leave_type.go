package models

import (
	"time"

	"github.com/google/uuid"
)

// PolicyLeaveType represents a leave type configuration for policy management
type PolicyLeaveType struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Name              string     `json:"name"`
	AnnualLimit       int        `json:"annual_limit"`
	CarryForwardLimit int        `json:"carry_forward_limit"`
	IsPaid            bool       `json:"is_paid"`
	RequiresApproval  bool       `json:"requires_approval"`
	AccrualRate       float64    `json:"accrual_rate"`
	Description       string     `json:"description"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}
