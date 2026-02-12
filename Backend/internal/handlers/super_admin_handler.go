package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
)

type SuperAdminHandler struct {
	organizationService  *services.OrganizationService
	subscriptionService  *services.SubscriptionService
	invoiceService       *services.InvoiceService
	usageTrackingService *services.UsageTrackingService
	analyticsService     *services.AnalyticsService
	superAdminService    *services.SuperAdminService
}

func NewSuperAdminHandler(
	orgService *services.OrganizationService,
	subService *services.SubscriptionService,
	invService *services.InvoiceService,
	usageService *services.UsageTrackingService,
	analyticsService *services.AnalyticsService,
	superAdminService *services.SuperAdminService,
) *SuperAdminHandler {
	return &SuperAdminHandler{
		organizationService:  orgService,
		subscriptionService:  subService,
		invoiceService:       invService,
		usageTrackingService: usageService,
		analyticsService:     analyticsService,
		superAdminService:    superAdminService,
	}
}

// Middleware to check super_admin role
func (h *SuperAdminHandler) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if userClaims.Role != "super_admin" {
			http.Error(w, "Access denied: Super Admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// SUBSCRIPTION PLANS
// ============================================================================

// GetAllPlans handles GET /api/v1/super-admin/plans
func (h *SuperAdminHandler) GetAllPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subscriptionService.GetAllPlans(r.Context(), true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plans": plans,
	})
}

// GetPlan handles GET /api/v1/super-admin/plans/{id}
func (h *SuperAdminHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	plan, err := h.subscriptionService.GetPlanByID(r.Context(), planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// UpdatePlan handles PUT /api/v1/super-admin/plans/{id}
func (h *SuperAdminHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	plan, err := h.subscriptionService.UpdatePlan(r.Context(), planID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// CreatePlan handles POST /api/v1/super-admin/plans
func (h *SuperAdminHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan models.SubscriptionPlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if plan.Name == "" || plan.DisplayName == "" || plan.PriceMonthly < 0 {
		http.Error(w, "Name, display name, and valid monthly price are required", http.StatusBadRequest)
		return
	}

	createdPlan, err := h.subscriptionService.CreatePlan(r.Context(), &plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPlan)
}

// DeletePlan handles DELETE /api/v1/super-admin/plans/{id}
func (h *SuperAdminHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.DeletePlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// ORGANIZATIONS
// ============================================================================

// GetAllOrganizations handles GET /api/v1/super-admin/organizations
func (h *SuperAdminHandler) GetAllOrganizations(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]interface{})

	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if country := r.URL.Query().Get("country"); country != "" {
		filters["country"] = country
	}
	if planID := r.URL.Query().Get("plan_id"); planID != "" {
		filters["plan_id"] = planID
	}

	organizations, err := h.organizationService.GetAllOrganizations(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organizations": organizations,
		"total":         len(organizations),
	})
}

// GetOrganization handles GET /api/v1/super-admin/organizations/{id}
func (h *SuperAdminHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	org, err := h.organizationService.GetOrganizationByID(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// CreateOrganization handles POST /api/v1/super-admin/organizations
func (h *SuperAdminHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req services.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.AdminEmail == "" || req.AdminPassword == "" {
		http.Error(w, "Name, admin email, and admin password are required", http.StatusBadRequest)
		return
	}

	org, err := h.organizationService.CreateOrganization(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

// UpdateOrganization handles PUT /api/v1/super-admin/organizations/{id}
func (h *SuperAdminHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	org, err := h.organizationService.UpdateOrganization(r.Context(), tenantID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// BlockOrganization handles POST /api/v1/super-admin/organizations/{id}/block
func (h *SuperAdminHandler) BlockOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	if err := h.organizationService.BlockOrganization(r.Context(), tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Organization blocked successfully"})
}

// UnblockOrganization handles POST /api/v1/super-admin/organizations/{id}/unblock
func (h *SuperAdminHandler) UnblockOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	if err := h.organizationService.UnblockOrganization(r.Context(), tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Organization unblocked successfully"})
}

// RenewOrganizationSubscription handles POST /api/v1/super-admin/organizations/{id}/renew
func (h *SuperAdminHandler) RenewOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PlanID       string `json:"plan_id"`
		BillingCycle string `json:"billing_cycle"`
	}

	// Try to decode body, but ignore error if body is empty (just renew)
	json.NewDecoder(r.Body).Decode(&req)

	var newPlanID *uuid.UUID
	var newBillingCycle *string

	if req.PlanID != "" {
		if id, err := uuid.Parse(req.PlanID); err == nil {
			newPlanID = &id
		}
	}

	if req.BillingCycle != "" {
		newBillingCycle = &req.BillingCycle
	}

	subscription, err := h.subscriptionService.RenewSubscription(r.Context(), tenantID, newPlanID, newBillingCycle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// DeleteOrganization handles DELETE /api/v1/super-admin/organizations/{id}
func (h *SuperAdminHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	if err := h.organizationService.DeleteOrganization(r.Context(), tenantID); err != nil {
		if strings.Contains(err.Error(), "organization not found") {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// INVOICES & BILLING
// ============================================================================

// GetAllInvoices handles GET /api/v1/super-admin/invoices
func (h *SuperAdminHandler) GetAllInvoices(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]interface{})

	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		filters["tenant_id"] = tenantID
	}

	invoices, err := h.invoiceService.GetInvoices(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoices": invoices,
		"total":    len(invoices),
	})
}

// GetInvoice handles GET /api/v1/super-admin/invoices/{id}
func (h *SuperAdminHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceService.GetInvoiceByID(r.Context(), invoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

// MarkInvoiceAsPaid handles POST /api/v1/super-admin/invoices/{id}/pay
func (h *SuperAdminHandler) MarkInvoiceAsPaid(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if _, err := h.invoiceService.PayInvoice(r.Context(), invoiceID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Invoice marked as paid"})
}

// GenerateBill handles POST /api/v1/super-admin/invoices/generate
func (h *SuperAdminHandler) GenerateBill(w http.ResponseWriter, r *http.Request) {
	// Request struct to handle string dates from frontend
	var req struct {
		TenantID       string                   `json:"tenant_id"`
		SubscriptionID string                   `json:"subscription_id,omitempty"`
		Subtotal       float64                  `json:"subtotal"`
		TaxAmount      float64                  `json:"tax_amount"`
		TotalAmount    float64                  `json:"total_amount"`
		Currency       string                   `json:"currency"`
		Status         string                   `json:"status"`
		IssueDate      string                   `json:"issue_date"`
		DueDate        string                   `json:"due_date"`
		BillingDetails map[string]interface{}   `json:"billing_details,omitempty"`
		LineItems      []models.InvoiceLineItem `json:"line_items"`
		Notes          string                   `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode invoice request: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Parse dates
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid issue date format: %v", err), http.StatusBadRequest)
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid due date format: %v", err), http.StatusBadRequest)
		return
	}

	// Parse subscription ID if provided
	var subscriptionID *uuid.UUID
	if req.SubscriptionID != "" {
		id, err := uuid.Parse(req.SubscriptionID)
		if err != nil {
			http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
			return
		}
		subscriptionID = &id
	}

	// Create invoice model
	invoice := &models.Invoice{
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		Subtotal:       req.Subtotal,
		TaxRate:        0, // Can be calculated or passed
		TaxAmount:      req.TaxAmount,
		DiscountAmount: 0,
		TotalAmount:    req.TotalAmount,
		Currency:       req.Currency,
		Status:         req.Status,
		IssueDate:      issueDate,
		DueDate:        dueDate,
		BillingDetails: req.BillingDetails,
		LineItems:      req.LineItems,
		Notes:          req.Notes,
	}

	createdInvoice, err := h.invoiceService.CreateInvoice(r.Context(), invoice)
	if err != nil {
		log.Printf("Failed to create invoice: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdInvoice)
}

// DownloadInvoice handles GET /api/v1/super-admin/invoices/{id}/download
func (h *SuperAdminHandler) DownloadInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	pdfBytes, err := h.invoiceService.GeneratePDF(r.Context(), invoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"invoice-%s.pdf\"", invoiceID))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// UpdateInvoice handles PUT /api/v1/super-admin/invoices/:id
func (h *SuperAdminHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		TenantID       string                   `json:"tenant_id"`
		Subtotal       float64                  `json:"subtotal"`
		TaxRate        float64                  `json:"tax_rate"`
		TaxAmount      float64                  `json:"tax_amount"`
		DiscountAmount float64                  `json:"discount_amount"`
		TotalAmount    float64                  `json:"total_amount"`
		Currency       string                   `json:"currency"`
		Status         string                   `json:"status"`
		IssueDate      time.Time                `json:"issue_date"`
		DueDate        time.Time                `json:"due_date"`
		BillingDetails map[string]interface{}   `json:"billing_details"`
		LineItems      []models.InvoiceLineItem `json:"line_items"`
		Notes          string                   `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	updates := &models.Invoice{
		TenantID:       tenantID,
		Subtotal:       req.Subtotal,
		TaxRate:        req.TaxRate,
		TaxAmount:      req.TaxAmount,
		DiscountAmount: req.DiscountAmount,
		TotalAmount:    req.TotalAmount,
		Currency:       req.Currency,
		Status:         req.Status,
		IssueDate:      req.IssueDate,
		DueDate:        req.DueDate,
		BillingDetails: req.BillingDetails,
		LineItems:      req.LineItems,
		Notes:          req.Notes,
	}

	invoice, err := h.invoiceService.UpdateInvoice(r.Context(), invoiceID, updates)
	if err != nil {
		log.Printf("Failed to update invoice: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoice": invoice,
	})
}

// DeleteInvoice handles DELETE /api/v1/super-admin/invoices/:id
func (h *SuperAdminHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.invoiceService.DeleteInvoice(r.Context(), invoiceID); err != nil {
		log.Printf("Failed to delete invoice: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// ANALYTICS & USAGE
// ============================================================================

// GetPlatformStats handles GET /api/v1/super-admin/analytics/platform-stats
func (h *SuperAdminHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analyticsService.GetPlatformStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetTenantGrowth handles GET /api/v1/super-admin/analytics/tenant-growth
func (h *SuperAdminHandler) GetTenantGrowth(w http.ResponseWriter, r *http.Request) {
	days := 30 // default
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		// Parse days parameter
		var d int
		if _, err := fmt.Sscanf(daysParam, "%d", &d); err == nil && d > 0 {
			days = d
		}
	}

	growth, err := h.analyticsService.GetTenantGrowth(r.Context(), days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"growth": growth,
		"days":   days,
	})
}

// GetRevenueMetrics handles GET /api/v1/super-admin/analytics/revenue
func (h *SuperAdminHandler) GetRevenueMetrics(w http.ResponseWriter, r *http.Request) {
	// Default to current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Parse date parameters if provided
	if start := r.URL.Query().Get("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	metrics, err := h.analyticsService.GetRevenueMetrics(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetOrganizationUsage handles GET /api/v1/super-admin/usage/organizations/{id}
func (h *SuperAdminHandler) GetOrganizationUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	summary, err := h.usageTrackingService.GetUsageSummary(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// ============================================================================
// SUPER ADMIN MANAGEMENT
// ============================================================================

// CreateSuperAdmin handles POST /api/v1/super-admin/admins
func (h *SuperAdminHandler) CreateSuperAdmin(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSuperAdminRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create super admin
	response, err := h.superAdminService.CreateSuperAdmin(r.Context(), req)
	if err != nil {
		log.Printf("Failed to create super admin: %v", err)

		// Check for specific errors
		if err.Error() == "email already exists" {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		http.Error(w, "Failed to create super admin", http.StatusInternalServerError)
		return
	}

	// Log success
	log.Printf("Super admin created successfully: %s (%s)", response.Email, response.ID)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAllSuperAdmins handles GET /api/v1/super-admin/admins
func (h *SuperAdminHandler) GetAllSuperAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.superAdminService.GetAllSuperAdmins(r.Context())
	if err != nil {
		log.Printf("Failed to fetch super admins: %v", err)
		http.Error(w, "Failed to fetch super admins", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"super_admins": admins,
		"total":        len(admins),
	})
}
