package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type LeaveHandler struct {
	leaveService *services.LeaveService
}

func NewLeaveHandler(leaveService *services.LeaveService) *LeaveHandler {
	return &LeaveHandler{
		leaveService: leaveService,
	}
}

// Helper to get authorized context info
func (h *LeaveHandler) getContextInfo(r *http.Request) (*uuid.UUID, *uuid.UUID, string, error) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		return nil, nil, "", fmt.Errorf("unauthorized")
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid tenant ID")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid user ID")
	}

	return &tenantID, &userID, claims.Role, nil
}

// CreateLeaveRequest handles POST /api/leaves
func (h *LeaveHandler) CreateLeaveRequest(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, _, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Employee ID
	employeeID, err := h.leaveService.GetEmployeeIDByUserID(r.Context(), *tenantID, *userID)
	if err != nil {
		http.Error(w, "Employee record not found", http.StatusForbidden)
		return
	}

	var req models.CreateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode leave request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	leave, err := h.leaveService.CreateLeaveRequest(r.Context(), *tenantID, employeeID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create leave request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(leave)
}

// GetLeaveRequests handles GET /api/leaves
func (h *LeaveHandler) GetLeaveRequests(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	statusParam := r.URL.Query().Get("status")
	var status *models.LeaveStatus
	if statusParam != "" {
		s := models.LeaveStatus(statusParam)
		status = &s
	}

	var leaves []models.LeaveRequest

	// If user is manager/hr/admin, show all leaves; otherwise show only their own
	if role == "manager" || role == "hr" || role == "admin" || role == "super_admin" {
		leaves, err = h.leaveService.GetLeaveRequests(r.Context(), *tenantID, nil, nil, status)
	} else {
		// Get Employee ID for filtering
		employeeID, err := h.leaveService.GetEmployeeIDByUserID(r.Context(), *tenantID, *userID)
		if err != nil {
			http.Error(w, "Employee record not found", http.StatusForbidden)
			return
		}
		leaves, err = h.leaveService.GetLeaveRequests(r.Context(), *tenantID, &employeeID, nil, status)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get leave requests")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaves)
}

// GetDepartmentLeaves handles GET /api/v1/company/manager/leaves (Manager only)
func (h *LeaveHandler) GetDepartmentLeaves(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Double check role (middleware should handle this, but good for safety)
	if role != "manager" && role != "admin" && role != "hr" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok || claims.DepartmentID == "" {
		http.Error(w, "User does not belong to a department", http.StatusBadRequest)
		return
	}

	departmentID, err := uuid.Parse(claims.DepartmentID)
	if err != nil {
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	// Parse status query param
	statusParam := r.URL.Query().Get("status")
	var status *models.LeaveStatus
	if statusParam != "" {
		s := models.LeaveStatus(statusParam)
		status = &s
	}

	leaves, err := h.leaveService.GetLeaveRequests(r.Context(), *tenantID, nil, &departmentID, status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get department leave requests")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaves)
}

// GetPendingLeaves handles GET /api/leaves/pending (for managers)
func (h *LeaveHandler) GetPendingLeaves(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only managers, HR, and admins can see all pending leaves
	if role != "manager" && role != "hr" && role != "admin" && role != "super_admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	leaves, err := h.leaveService.GetPendingLeaveRequests(r.Context(), *tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending leave requests")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaves)
}

// ApproveLeave handles PUT /api/leaves/{id}/approve
func (h *LeaveHandler) ApproveLeave(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only managers, HR, and admins can approve leaves
	if role != "manager" && role != "hr" && role != "admin" && role != "super_admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get Approver Employee ID
	approverID, err := h.leaveService.GetEmployeeIDByUserID(r.Context(), *tenantID, *userID)
	if err != nil {
		http.Error(w, "Approver employee record not found", http.StatusForbidden)
		return
	}

	leaveIDStr := chi.URLParam(r, "leaveID") // Updated param name to match server.go
	if leaveIDStr == "" {
		leaveIDStr = chi.URLParam(r, "id") // Fallback
	}

	leaveID, err := uuid.Parse(leaveIDStr)
	if err != nil {
		http.Error(w, "Invalid leave ID", http.StatusBadRequest)
		return
	}

	err = h.leaveService.ApproveLeaveRequest(r.Context(), *tenantID, leaveID, approverID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to approve leave request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Leave request approved successfully",
	})
}

// RejectLeave handles PUT /api/leaves/{id}/reject
func (h *LeaveHandler) RejectLeave(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only managers, HR, and admins can reject leaves
	if role != "manager" && role != "hr" && role != "admin" && role != "super_admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get Approver Employee ID
	approverID, err := h.leaveService.GetEmployeeIDByUserID(r.Context(), *tenantID, *userID)
	if err != nil {
		http.Error(w, "Approver employee record not found", http.StatusForbidden)
		return
	}

	leaveIDStr := chi.URLParam(r, "leaveID") // Updated param name to match server.go
	if leaveIDStr == "" {
		leaveIDStr = chi.URLParam(r, "id") // Fallback
	}

	leaveID, err := uuid.Parse(leaveIDStr)
	if err != nil {
		http.Error(w, "Invalid leave ID", http.StatusBadRequest)
		return
	}

	var req models.RejectLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.leaveService.RejectLeaveRequest(r.Context(), *tenantID, leaveID, approverID, req.RejectionReason)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reject leave request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Leave request rejected successfully",
	})
}

// CreateLeaveType handles POST /api/v1/company/leave-types
func (h *LeaveHandler) CreateLeaveType(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin and HR can create leave types
	if role != "admin" && role != "hr" && role != "super_admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var req models.LeaveTypeConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	leaveType, err := h.leaveService.CreateLeaveType(r.Context(), *tenantID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create leave type")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(leaveType)
}
