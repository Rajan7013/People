package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type BillingService struct {
	db *sql.DB
}

func NewBillingService(db *sql.DB) *BillingService {
	return &BillingService{db: db}
}

// GenerateInvoice creates a new invoice for a subscription
func (s *BillingService) GenerateInvoice(ctx context.Context, tenantID uuid.UUID, subscriptionID uuid.UUID) (*models.Invoice, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get subscription details
	var amount float64
	var currency, planName string
	query := `
		SELECT s.amount, s.currency, sp.display_name
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.id = $1 AND s.tenant_id = $2`

	err = tx.QueryRowContext(ctx, query, subscriptionID, tenantID).Scan(&amount, &currency, &planName)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Generate invoice number
	var invoiceNumber string
	err = tx.QueryRowContext(ctx, "SELECT generate_invoice_number()").Scan(&invoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice
	invoiceID := uuid.New()
	now := time.Now()
	issueDate := now
	dueDate := now.AddDate(0, 0, 30) // 30 days payment term

	// Calculate tax (example: 10% tax)
	taxRate := 10.0
	taxAmount := amount * (taxRate / 100)
	totalAmount := amount + taxAmount

	// Create line items
	lineItems := []models.InvoiceLineItem{
		{
			Description: fmt.Sprintf("%s Subscription", planName),
			Quantity:    1,
			UnitPrice:   amount,
			Amount:      amount,
		},
	}

	invoiceQuery := `
		INSERT INTO invoices (
			id, invoice_number, tenant_id, subscription_id, subtotal, tax_rate, tax_amount,
			discount_amount, total_amount, currency, status, issue_date, due_date,
			line_items, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`

	err = tx.QueryRowContext(ctx, invoiceQuery,
		invoiceID, invoiceNumber, tenantID, subscriptionID, amount, taxRate, taxAmount,
		0.0, totalAmount, currency, "pending", issueDate, dueDate,
		lineItems, now, now,
	).Scan(&invoiceID)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetInvoiceByID(ctx, invoiceID)
}

// GetInvoiceByID retrieves an invoice by ID
func (s *BillingService) GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error) {
	query := `
		SELECT i.id, i.invoice_number, i.tenant_id, i.subscription_id, i.subtotal, i.tax_rate,
			i.tax_amount, i.discount_amount, i.total_amount, i.currency, i.status,
			i.issue_date, i.due_date, i.paid_at, i.payment_method, i.transaction_id,
			i.payment_gateway, i.billing_details, i.line_items, i.notes,
			i.created_at, i.updated_at, t.name as tenant_name
		FROM invoices i
		JOIN tenants t ON i.tenant_id = t.id
		WHERE i.id = $1`

	invoice := &models.Invoice{}
	err := s.db.QueryRowContext(ctx, query, invoiceID).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.TenantID, &invoice.SubscriptionID,
		&invoice.Subtotal, &invoice.TaxRate, &invoice.TaxAmount, &invoice.DiscountAmount,
		&invoice.TotalAmount, &invoice.Currency, &invoice.Status, &invoice.IssueDate,
		&invoice.DueDate, &invoice.PaidAt, &invoice.PaymentMethod, &invoice.TransactionID,
		&invoice.PaymentGateway, &invoice.BillingDetails, &invoice.LineItems, &invoice.Notes,
		&invoice.CreatedAt, &invoice.UpdatedAt, &invoice.TenantName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoicesByTenant retrieves all invoices for a tenant
func (s *BillingService) GetInvoicesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Invoice, error) {
	query := `
		SELECT i.id, i.invoice_number, i.tenant_id, i.subscription_id, i.subtotal,
			i.total_amount, i.currency, i.status, i.issue_date, i.due_date, i.paid_at,
			i.created_at, t.name as tenant_name
		FROM invoices i
		JOIN tenants t ON i.tenant_id = t.id
		WHERE i.tenant_id = $1
		ORDER BY i.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.TenantID, &invoice.SubscriptionID,
			&invoice.Subtotal, &invoice.TotalAmount, &invoice.Currency, &invoice.Status,
			&invoice.IssueDate, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt,
			&invoice.TenantName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// GetAllInvoices retrieves all invoices with filters
func (s *BillingService) GetAllInvoices(ctx context.Context, filters map[string]interface{}) ([]*models.Invoice, error) {
	query := `
		SELECT i.id, i.invoice_number, i.tenant_id, i.total_amount, i.currency, i.status,
			i.issue_date, i.due_date, i.paid_at, i.created_at, t.name as tenant_name
		FROM invoices i
		JOIN tenants t ON i.tenant_id = t.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND i.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if tenantID, ok := filters["tenant_id"].(string); ok && tenantID != "" {
		query += fmt.Sprintf(" AND i.tenant_id = $%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	query += " ORDER BY i.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		invoice := &models.Invoice{}
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.TenantID, &invoice.TotalAmount,
			&invoice.Currency, &invoice.Status, &invoice.IssueDate, &invoice.DueDate,
			&invoice.PaidAt, &invoice.CreatedAt, &invoice.TenantName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *BillingService) MarkInvoiceAsPaid(ctx context.Context, invoiceID uuid.UUID, paymentMethod, transactionID string) error {
	query := `
		UPDATE invoices 
		SET status = 'paid', paid_at = $1, payment_method = $2, transaction_id = $3, updated_at = $4
		WHERE id = $5`

	_, err := s.db.ExecContext(ctx, query, time.Now(), paymentMethod, transactionID, time.Now(), invoiceID)
	if err != nil {
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	return nil
}

// CancelInvoice cancels an invoice
func (s *BillingService) CancelInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	query := `UPDATE invoices SET status = 'cancelled', updated_at = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, time.Now(), invoiceID)
	if err != nil {
		return fmt.Errorf("failed to cancel invoice: %w", err)
	}
	return nil
}

// GetRevenueStats retrieves revenue statistics
func (s *BillingService) GetRevenueStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_invoices,
			COUNT(*) FILTER (WHERE status = 'paid') as paid_invoices,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'paid'), 0) as total_revenue,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'pending'), 0) as pending_revenue,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'overdue'), 0) as overdue_revenue
		FROM invoices
		WHERE issue_date BETWEEN $1 AND $2`

	var totalInvoices, paidInvoices int
	var totalRevenue, pendingRevenue, overdueRevenue float64

	err := s.db.QueryRowContext(ctx, query, startDate, endDate).Scan(
		&totalInvoices, &paidInvoices, &totalRevenue, &pendingRevenue, &overdueRevenue,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_invoices":  totalInvoices,
		"paid_invoices":   paidInvoices,
		"total_revenue":   totalRevenue,
		"pending_revenue": pendingRevenue,
		"overdue_revenue": overdueRevenue,
		"collection_rate": 0.0,
	}

	if totalInvoices > 0 {
		stats["collection_rate"] = float64(paidInvoices) / float64(totalInvoices) * 100
	}

	return stats, nil
}
