package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type InvoiceService struct {
	db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(ctx context.Context, invoice *models.Invoice) (*models.Invoice, error) {
	invoice.ID = uuid.New()

	// Generate Invoice Number if not provided
	if invoice.InvoiceNumber == "" {
		var invoiceNum string
		err := s.db.QueryRowContext(ctx, "SELECT generate_invoice_number()").Scan(&invoiceNum)
		if err != nil {
			return nil, fmt.Errorf("failed to generate invoice number: %w", err)
		}
		invoice.InvoiceNumber = invoiceNum
	}

	billingDetailsJSON, err := json.Marshal(invoice.BillingDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal billing details: %w", err)
	}

	lineItemsJSON, err := json.Marshal(invoice.LineItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal line items: %w", err)
	}

	query := `
		INSERT INTO invoices (
			id, invoice_number, tenant_id, subscription_id,
			subtotal, tax_rate, tax_amount, discount_amount, total_amount, currency,
			status, issue_date, due_date,
			billing_details, line_items, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`

	err = s.db.QueryRowContext(ctx, query,
		invoice.ID, invoice.InvoiceNumber, invoice.TenantID, invoice.SubscriptionID,
		invoice.Subtotal, invoice.TaxRate, invoice.TaxAmount, invoice.DiscountAmount, invoice.TotalAmount, invoice.Currency,
		invoice.Status, invoice.IssueDate, invoice.DueDate,
		billingDetailsJSON, lineItemsJSON, invoice.Notes,
		time.Now(), time.Now(),
	).Scan(&invoice.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return s.GetInvoiceByID(ctx, invoice.ID)
}

// GetInvoiceByID retrieves an invoice by ID
func (s *InvoiceService) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	query := `
		SELECT 
			i.id, i.invoice_number, i.tenant_id, i.subscription_id,
			i.subtotal, i.tax_rate, i.tax_amount, i.discount_amount, i.total_amount, i.currency,
			i.status, i.issue_date, i.due_date, i.paid_at,
			i.payment_method, i.transaction_id, i.payment_gateway,
			i.billing_details, i.line_items, i.notes,
			i.created_at, i.updated_at,
			t.name as tenant_name
		FROM invoices i
		JOIN tenants t ON i.tenant_id = t.id
		WHERE i.id = $1`

	invoice := &models.Invoice{}
	var billingDetailsBytes, lineItemsBytes []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.TenantID, &invoice.SubscriptionID,
		&invoice.Subtotal, &invoice.TaxRate, &invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount, &invoice.Currency,
		&invoice.Status, &invoice.IssueDate, &invoice.DueDate, &invoice.PaidAt,
		&invoice.PaymentMethod, &invoice.TransactionID, &invoice.PaymentGateway,
		&billingDetailsBytes, &lineItemsBytes, &invoice.Notes,
		&invoice.CreatedAt, &invoice.UpdatedAt,
		&invoice.TenantName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if billingDetailsBytes != nil {
		if err := json.Unmarshal(billingDetailsBytes, &invoice.BillingDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal billing details: %w", err)
		}
	}

	if lineItemsBytes != nil {
		if err := json.Unmarshal(lineItemsBytes, &invoice.LineItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal line items: %w", err)
		}
	}

	return invoice, nil
}

// GetInvoices retrieves invoices with filtering
func (s *InvoiceService) GetInvoices(ctx context.Context, filters map[string]interface{}) ([]*models.Invoice, error) {
	query := `
		SELECT 
			i.id, i.invoice_number, i.tenant_id, i.subscription_id,
			i.subtotal, i.tax_rate, i.tax_amount, i.discount_amount, i.total_amount, i.currency,
			i.status, i.issue_date, i.due_date, i.paid_at,
			i.created_at,
			t.name as tenant_name
		FROM invoices i
		JOIN tenants t ON i.tenant_id = t.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if tenantID, ok := filters["tenant_id"].(string); ok && tenantID != "" {
		query += fmt.Sprintf(" AND i.tenant_id = $%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND i.status = $%d", argIndex)
		args = append(args, status)
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
			&invoice.ID, &invoice.InvoiceNumber, &invoice.TenantID, &invoice.SubscriptionID,
			&invoice.Subtotal, &invoice.TaxRate, &invoice.TaxAmount, &invoice.DiscountAmount, &invoice.TotalAmount, &invoice.Currency,
			&invoice.Status, &invoice.IssueDate, &invoice.DueDate, &invoice.PaidAt,
			&invoice.CreatedAt,
			&invoice.TenantName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// PayInvoice marks an invoice as paid (Fake Payment)
func (s *InvoiceService) PayInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	query := `
		UPDATE invoices 
		SET status = 'paid', paid_at = $1, payment_method = 'credit_card', 
			transaction_id = $2, updated_at = $3
		WHERE id = $4`

	// Generate fake transaction ID
	txID := "txn_" + uuid.New().String()[:8]

	_, err := s.db.ExecContext(ctx, query, time.Now(), txID, time.Now(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to pay invoice: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// UpdateInvoice updates an existing invoice
func (s *InvoiceService) UpdateInvoice(ctx context.Context, id uuid.UUID, updates *models.Invoice) (*models.Invoice, error) {
	// First, check if invoice exists and get current status
	current, err := s.GetInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Prevent updates to paid invoices
	if current.Status == "paid" {
		return nil, fmt.Errorf("cannot update paid invoice")
	}

	// Marshal JSON fields
	billingDetailsJSON, err := json.Marshal(updates.BillingDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal billing details: %w", err)
	}

	lineItemsJSON, err := json.Marshal(updates.LineItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal line items: %w", err)
	}

	query := `
		UPDATE invoices 
		SET subtotal = $1, tax_rate = $2, tax_amount = $3, discount_amount = $4, 
			total_amount = $5, currency = $6, status = $7, issue_date = $8, due_date = $9,
			billing_details = $10, line_items = $11, notes = $12, updated_at = $13
		WHERE id = $14`

	_, err = s.db.ExecContext(ctx, query,
		updates.Subtotal, updates.TaxRate, updates.TaxAmount, updates.DiscountAmount,
		updates.TotalAmount, updates.Currency, updates.Status, updates.IssueDate, updates.DueDate,
		billingDetailsJSON, lineItemsJSON, updates.Notes, time.Now(), id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// DeleteInvoice soft deletes an invoice
func (s *InvoiceService) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	// Check if invoice exists and get current status
	invoice, err := s.GetInvoiceByID(ctx, id)
	if err != nil {
		return err
	}

	// Prevent deletion of paid invoices
	if invoice.Status == "paid" {
		return fmt.Errorf("cannot delete paid invoice")
	}

	// Hard delete since we don't have deleted_at column
	query := `DELETE FROM invoices WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// GeneratePDF generates a PDF for the invoice using gofpdf
func (s *InvoiceService) GeneratePDF(ctx context.Context, id uuid.UUID) ([]byte, error) {
	invoice, err := s.GetInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Colors
	primaryColor := []int{37, 99, 235} // Blue-600
	grayColor := []int{107, 114, 128}  // Gray-500
	lightGray := []int{243, 244, 246}  // Gray-100

	// Logo Placeholder (Text for now)
	// Logo Placeholder (Text for now)
	// Logo (Left)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Cell(100, 10, "PeopleOS")

	// INVOICE Label (Right)
	pdf.SetXY(120, 10)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, "INVOICE")

	pdf.Ln(15) // Move down

	// Store Y position for two columns
	yPos := pdf.GetY()

	// Column 1: Bill To (Left)
	pdf.SetXY(10, yPos)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, "Bill To:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, invoice.TenantName)
	pdf.Ln(5)
	if invoice.BillingDetails["address"] != nil {
		pdf.Cell(0, 5, fmt.Sprintf("%v", invoice.BillingDetails["address"]))
		pdf.Ln(5)
	}
	if invoice.BillingDetails["email"] != nil {
		pdf.Cell(0, 5, fmt.Sprintf("%v", invoice.BillingDetails["email"]))
		pdf.Ln(5)
	}

	// Column 2: Invoice Details (Right)
	pdf.SetXY(120, yPos)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Invoice #:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, invoice.InvoiceNumber)

	pdf.SetXY(120, yPos+6)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, invoice.IssueDate.Format("Jan 02, 2006"))

	pdf.SetXY(120, yPos+12)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Due Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, invoice.DueDate.Format("Jan 02, 2006"))

	// Move cursor below the lowest column
	pdf.SetXY(10, yPos+30)

	// Items Table Header
	pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 7, "Description", "1", 0, "", true, 0, "") // Reduced height to 7
	pdf.CellFormat(30, 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 7, "Total", "1", 1, "R", true, 0, "")

	// Items
	pdf.SetFont("Arial", "", 10)
	total := 0.0
	for _, item := range invoice.LineItems {
		amount := float64(item.Quantity) * item.UnitPrice
		total += amount

		pdf.CellFormat(90, 7, item.Description, "1", 0, "", false, 0, "") // Reduced height to 7
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("%s%.2f", invoice.Currency, item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("%s%.2f", invoice.Currency, amount), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 6, "")
	pdf.Cell(35, 6, "Subtotal:")
	pdf.CellFormat(35, 6, fmt.Sprintf("%s%.2f", invoice.Currency, total), "", 1, "R", false, 0, "")

	pdf.Cell(120, 6, "")
	pdf.Cell(35, 6, fmt.Sprintf("Tax (%.0f%%):", invoice.TaxRate))
	pdf.CellFormat(35, 6, fmt.Sprintf("%s%.2f", invoice.Currency, invoice.TaxAmount), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(120, 8, "")
	pdf.Cell(35, 8, "Total:")
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.CellFormat(35, 8, fmt.Sprintf("%s%.2f", invoice.Currency, invoice.TotalAmount), "", 1, "R", false, 0, "")

	// Footer - inline instead of absolute positioning
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(grayColor[0], grayColor[1], grayColor[2])
	pdf.CellFormat(0, 5, "Thank you for your business. Please make checks payable to PeopleOS Inc.", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
