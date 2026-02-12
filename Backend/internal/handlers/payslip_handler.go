package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PayslipHandler struct {
	PayslipService *services.PayslipService
}

func NewPayslipHandler(payslipService *services.PayslipService) *PayslipHandler {
	return &PayslipHandler{
		PayslipService: payslipService,
	}
}

// GetPayslips handles GET /api/payslips
func (h *PayslipHandler) GetPayslips(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert string UUIDs to uuid.UUID
	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	filter := &models.PayslipFilter{}

	// Employee ID filter
	if empIDStr := r.URL.Query().Get("employee_id"); empIDStr != "" {
		if empID, err := uuid.Parse(empIDStr); err == nil {
			filter.EmployeeID = &empID
		}
	}

	// Department ID filter
	if deptIDStr := r.URL.Query().Get("department_id"); deptIDStr != "" {
		if deptID, err := uuid.Parse(deptIDStr); err == nil {
			filter.DepartmentID = &deptID
		}
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Date filters
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.PayPeriodStart = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.PayPeriodEnd = &endDate
		}
	}

	// Month/Year filters
	if monthStr := r.URL.Query().Get("month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil && month >= 1 && month <= 12 {
			filter.PaymentMonth = &month
		}
	}

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.PaymentYear = &year
		}
	}

	// Role-based filtering
	if userClaims.Role == "employee" {
		// Employees can only see their own payslips
		if employeeID, err := uuid.Parse(userClaims.UserID); err == nil {
			filter.EmployeeID = &employeeID
		}
	}

	payslips, totalCount, err := h.PayslipService.GetPayslipsByTenant(tenantID, filter, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"payslips": payslips,
		"total":    totalCount,
		"page":     page,
		"limit":    limit,
		"has_more": totalCount > page*limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPayslip handles GET /api/payslips/{id}
func (h *PayslipHandler) GetPayslip(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	payslipIDStr := chi.URLParam(r, "id")
	payslipID, err := uuid.Parse(payslipIDStr)
	if err != nil {
		http.Error(w, "Invalid payslip ID", http.StatusBadRequest)
		return
	}

	payslip, err := h.PayslipService.GetPayslipByID(tenantID, payslipID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "Payslip not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Role-based access control
	if userClaims.Role == "employee" {
		if userEmployeeID, err := uuid.Parse(userClaims.UserID); err == nil && payslip.EmployeeID != userEmployeeID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payslip)
}

// CreatePayslip handles POST /api/payslips
func (h *PayslipHandler) CreatePayslip(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Only admin and hr can create payslips
	if userClaims.Role != "admin" && userClaims.Role != "hr" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req models.PayslipCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmployeeID == uuid.Nil {
		http.Error(w, "Employee ID is required", http.StatusBadRequest)
		return
	}

	if req.PayPeriodStart.IsZero() || req.PayPeriodEnd.IsZero() {
		http.Error(w, "Pay period start and end dates are required", http.StatusBadRequest)
		return
	}

	if req.PayPeriodEnd.Before(req.PayPeriodStart) {
		http.Error(w, "Pay period end must be after start date", http.StatusBadRequest)
		return
	}

	payslip, err := h.PayslipService.CreatePayslip(tenantID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payslip)
}

// UpdatePayslip handles PUT /api/payslips/{id}
func (h *PayslipHandler) UpdatePayslip(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Only admin and hr can update payslips
	if userClaims.Role != "admin" && userClaims.Role != "hr" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	payslipIDStr := chi.URLParam(r, "id")
	payslipID, err := uuid.Parse(payslipIDStr)
	if err != nil {
		http.Error(w, "Invalid payslip ID", http.StatusBadRequest)
		return
	}

	var req models.PayslipUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"draft":     true,
			"approved":  true,
			"paid":      true,
			"cancelled": true,
		}
		if !validStatuses[*req.Status] {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
	}

	err = h.PayslipService.UpdatePayslip(tenantID, payslipID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payslip updated successfully"})
}

// DeletePayslip handles DELETE /api/payslips/{id}
func (h *PayslipHandler) DeletePayslip(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Only admin can delete payslips
	if userClaims.Role != "admin" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	payslipIDStr := chi.URLParam(r, "id")
	payslipID, err := uuid.Parse(payslipIDStr)
	if err != nil {
		http.Error(w, "Invalid payslip ID", http.StatusBadRequest)
		return
	}

	err = h.PayslipService.DeletePayslip(tenantID, payslipID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payslip deleted successfully"})
}

// GetPayslipStats handles GET /api/payslips/stats
func (h *PayslipHandler) GetPayslipStats(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Only admin, hr, and manager can view stats
	if userClaims.Role == "employee" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	stats, err := h.PayslipService.GetPayslipStats(tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetEmployeePayslips handles GET /api/employees/{id}/payslips
func (h *PayslipHandler) GetEmployeePayslips(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	employeeIDStr := chi.URLParam(r, "id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	// Role-based access control
	if userClaims.Role == "employee" {
		if userEmployeeID, err := uuid.Parse(userClaims.UserID); err != nil || employeeID != userEmployeeID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	payslips, totalCount, err := h.PayslipService.GetPayslipsByEmployee(tenantID, employeeID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"payslips": payslips,
		"total":    totalCount,
		"page":     page,
		"limit":    limit,
		"has_more": totalCount > page*limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
