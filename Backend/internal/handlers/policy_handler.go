package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
)

type PolicyHandler struct {
	policyService *services.PolicyService
}

func NewPolicyHandler(policyService *services.PolicyService) *PolicyHandler {
	return &PolicyHandler{
		policyService: policyService,
	}
}

// GetAttendancePolicy handles GET /api/v1/company/admin/policies/attendance
func (h *PolicyHandler) GetAttendancePolicy(w http.ResponseWriter, r *http.Request) {
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

	policy, err := h.policyService.GetAttendancePolicy(tenantID)
	if err != nil {
		http.Error(w, "Failed to get attendance policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// UpdateAttendancePolicy handles PUT /api/v1/company/admin/policies/attendance
func (h *PolicyHandler) UpdateAttendancePolicy(w http.ResponseWriter, r *http.Request) {
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

	var req services.AttendancePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	policy, err := h.policyService.UpdateAttendancePolicy(tenantID, req)
	if err != nil {
		http.Error(w, "Failed to update attendance policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// GetSalaryComponents handles GET /api/v1/company/admin/policies/salary-components
func (h *PolicyHandler) GetSalaryComponents(w http.ResponseWriter, r *http.Request) {
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

	components, err := h.policyService.GetSalaryComponents(tenantID)
	if err != nil {
		http.Error(w, "Failed to get salary components: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(components)
}

// CreateSalaryComponent handles POST /api/v1/company/admin/policies/salary-components
func (h *PolicyHandler) CreateSalaryComponent(w http.ResponseWriter, r *http.Request) {
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

	var req services.SalaryComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate type
	if req.Type != "earning" && req.Type != "deduction" {
		http.Error(w, "Type must be 'earning' or 'deduction'", http.StatusBadRequest)
		return
	}

	component, err := h.policyService.CreateSalaryComponent(tenantID, req)
	if err != nil {
		http.Error(w, "Failed to create salary component: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(component)
}

// GetLeaveTypes handles GET /api/v1/company/admin/policies/leave-types
func (h *PolicyHandler) GetLeaveTypes(w http.ResponseWriter, r *http.Request) {
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

	leaveTypes, err := h.policyService.GetLeaveTypes(tenantID)
	if err != nil {
		http.Error(w, "Failed to get leave types: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaveTypes)
}

// CreateLeaveType handles POST /api/v1/company/admin/policies/leave-types
func (h *PolicyHandler) CreateLeaveType(w http.ResponseWriter, r *http.Request) {
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

	var req services.LeaveTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	leaveType, err := h.policyService.CreateLeaveType(tenantID, req)
	if err != nil {
		http.Error(w, "Failed to create leave type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(leaveType)
}
