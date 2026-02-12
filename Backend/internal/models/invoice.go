package models

import (
	"time"

	"github.com/google/uuid"
)

// Invoice represents a billing record
type Invoice struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	InvoiceNumber  string     `json:"invoice_number" db:"invoice_number"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty" db:"subscription_id"`

	// Financials
	Subtotal       float64 `json:"subtotal" db:"subtotal"`
	TaxRate        float64 `json:"tax_rate" db:"tax_rate"`
	TaxAmount      float64 `json:"tax_amount" db:"tax_amount"`
	DiscountAmount float64 `json:"discount_amount" db:"discount_amount"`
	TotalAmount    float64 `json:"total_amount" db:"total_amount"`
	Currency       string  `json:"currency" db:"currency"`

	// Status
	Status string `json:"status" db:"status"` // draft, pending, paid, overdue, cancelled, refunded

	// Dates
	IssueDate time.Time  `json:"issue_date" db:"issue_date"`
	DueDate   time.Time  `json:"due_date" db:"due_date"`
	PaidAt    *time.Time `json:"paid_at,omitempty" db:"paid_at"`

	// Payment details
	PaymentMethod  *string `json:"payment_method,omitempty" db:"payment_method"`
	TransactionID  *string `json:"transaction_id,omitempty" db:"transaction_id"`
	PaymentGateway *string `json:"payment_gateway,omitempty" db:"payment_gateway"`

	// JSON fields
	BillingDetails map[string]interface{} `json:"billing_details" db:"billing_details"`
	LineItems      []InvoiceLineItem      `json:"line_items" db:"line_items"`

	// Notes
	Notes string `json:"notes" db:"notes"`

	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	TenantName string `json:"tenant_name,omitempty" db:"tenant_name"`
}

// InvoiceLineItem represents an item in an invoice
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// Implementing Scanner/Valuer for JSONB is handled manually in service currently,
// but can be added here if needed. For now sticking to service-level handling as per previous pattern.
