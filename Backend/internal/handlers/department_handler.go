package handlers

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
)

type DepartmentHandler struct {
	departmentService *services.DepartmentService
}

func NewDepartmentHandler(departmentService *services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
	}
}

// CreateDepartment handles POST /api/v1/{tenantId}/departments
func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
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

	allowedRoles := []string{"admin", "hr"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req services.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Department name is required", http.StatusBadRequest)
		return
	}

	dept, err := h.departmentService.CreateDepartment(tenantID, &req)
	if err != nil {
		http.Error(w, "Failed to create department: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dept)
}

// UpdateDepartment handles PUT /api/v1/{tenantId}/departments/{departmentId}
func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
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

	allowedRoles := []string{"admin", "hr"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	deptIDStr := chi.URLParam(r, "departmentID")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedDept, err := h.departmentService.UpdateDepartment(tenantID, deptID, updates)
	if err != nil {
		if err.Error() == "department not found" {
			http.Error(w, "Department not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update department: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDept)
}

// GetDepartments handles GET /api/v1/{tenantId}/departments
func (h *DepartmentHandler) GetDepartments(w http.ResponseWriter, r *http.Request) {
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

	departments, err := h.departmentService.GetDepartments(tenantID)
	if err != nil {
		http.Error(w, "Failed to get departments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(departments)
}

// DeleteDepartment handles DELETE /api/v1/{tenantId}/departments/{departmentId}
func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
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

	allowedRoles := []string{"admin", "hr"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	deptIDStr := chi.URLParam(r, "departmentID")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	err = h.departmentService.DeleteDepartment(tenantID, deptID)
	if err != nil {
		http.Error(w, "Failed to delete department: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
