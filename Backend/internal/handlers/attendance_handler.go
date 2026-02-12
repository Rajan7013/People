package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
)

type AttendanceHandler struct {
	attendanceService *services.AttendanceService
}

func NewAttendanceHandler(attendanceService *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
	}
}

// CheckIn handles employee check-in
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID: "+userClaims.TenantID, http.StatusBadRequest)
		return
	}

	var req models.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID: "+userClaims.UserID, http.StatusBadRequest)
		return
	}

	// Get employee ID for the user
	employeeID, err := h.attendanceService.GetEmployeeIDByUserID(r.Context(), tenantID, userID)
	if err != nil {
		http.Error(w, "Employee record not found: "+err.Error(), http.StatusNotFound)
		return
	}

	record, err := h.attendanceService.CheckIn(r.Context(), tenantID, employeeID, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Check-in successful",
		"record":  record,
	})
}

// CheckOut handles employee check-out
func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
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

	var req models.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get employee ID for the user
	employeeID, err := h.attendanceService.GetEmployeeIDByUserID(r.Context(), tenantID, userID)
	if err != nil {
		http.Error(w, "Employee record not found", http.StatusNotFound)
		return
	}

	record, err := h.attendanceService.CheckOut(r.Context(), tenantID, employeeID, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Check-out successful",
		"record":  record,
	})
}

// GetEmployeeAttendance gets attendance records for a specific employee
func (h *AttendanceHandler) GetEmployeeAttendance(w http.ResponseWriter, r *http.Request) {
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

	employeeIDStr := chi.URLParam(r, "employeeId")

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default to last 30 days
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.attendanceService.GetEmployeeAttendance(r.Context(), tenantID, employeeID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
	})
}

// GetAttendanceRecords gets all attendance records
func (h *AttendanceHandler) GetAttendanceRecords(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default to last 7 days
		startDate = time.Now().AddDate(0, 0, -7)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	records, total, err := h.attendanceService.GetAttendanceRecords(r.Context(), tenantID, startDate, endDate, page, limit, search, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records":     records,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	})
}

// GetTodayAttendance gets today's attendance records
func (h *AttendanceHandler) GetTodayAttendance(w http.ResponseWriter, r *http.Request) {
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

	records, err := h.attendanceService.GetTodayAttendance(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "GetTodayAttendance error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
		"date":    time.Now().Format("2006-01-02"),
	})
}

// GetAttendanceStats gets attendance statistics
func (h *AttendanceHandler) GetAttendanceStats(w http.ResponseWriter, r *http.Request) {
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

	stats, err := h.attendanceService.GetAttendanceStats(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": stats,
	})
}

// GetEmployeeStatus gets current check-in status for an employee
func (h *AttendanceHandler) GetEmployeeStatus(w http.ResponseWriter, r *http.Request) {
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

	employeeIDStr := chi.URLParam(r, "employeeId")

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	record, err := h.attendanceService.GetEmployeeCurrentStatus(r.Context(), tenantID, employeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"employee_id":    employeeID,
		"is_checked_in":  record.CheckInTime != nil && record.CheckOutTime == nil,
		"is_checked_out": record.CheckOutTime != nil,
		"status":         record.Status,
		"check_in_time":  record.CheckInTime,
		"check_out_time": record.CheckOutTime,
		"total_hours":    record.TotalHours,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetCurrentUserStatus gets current check-in status for the authenticated user
func (h *AttendanceHandler) GetCurrentUserStatus(w http.ResponseWriter, r *http.Request) {
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

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get employee ID for the user
	employeeID, err := h.attendanceService.GetEmployeeIDByUserID(r.Context(), tenantID, userID)
	if err != nil {
		http.Error(w, "Employee record not found: "+err.Error(), http.StatusNotFound)
		return
	}

	record, err := h.attendanceService.GetEmployeeCurrentStatus(r.Context(), tenantID, employeeID)
	if err != nil {
		http.Error(w, "GetEmployeeCurrentStatus error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"is_checked_in": record != nil && record.CheckInTime != nil && record.CheckOutTime == nil,
		"record":        record,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// UpdateAttendanceRecord updates an attendance record (admin only)
func (h *AttendanceHandler) UpdateAttendanceRecord(w http.ResponseWriter, r *http.Request) {
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

	recordIDStr := chi.URLParam(r, "recordId")

	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Remove fields that shouldn't be updated via this endpoint
	delete(updates, "id")
	delete(updates, "tenant_id")
	delete(updates, "employee_id")
	delete(updates, "created_at")

	err = h.attendanceService.UpdateAttendanceRecord(r.Context(), tenantID, recordID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Attendance record updated successfully",
	})
}

// GetDepartmentAttendance gets attendance records for the user's department (Manager only)
func (h *AttendanceHandler) GetDepartmentAttendance(w http.ResponseWriter, r *http.Request) {
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

	// Double check role
	if !slices.Contains([]string{"manager", "admin", "hr"}, userClaims.Role) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if userClaims.DepartmentID == "" {
		http.Error(w, "User does not belong to a department", http.StatusBadRequest)
		return
	}

	departmentID, err := uuid.Parse(userClaims.DepartmentID)
	if err != nil {
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default to last 7 days
		startDate = time.Now().AddDate(0, 0, -7)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.attendanceService.GetDepartmentAttendance(r.Context(), tenantID, departmentID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get department attendance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
	})
}

// GetTeamAttendance gets attendance records for the user's team (Team Lead, Manager)
func (h *AttendanceHandler) GetTeamAttendance(w http.ResponseWriter, r *http.Request) {
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

	// Helper function check
	// Note: using inline check since contains helper might be removed or tricky with package scope
	allowed := false
	for _, role := range []string{"team_lead", "manager", "admin", "hr"} {
		if userClaims.Role == role {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if userClaims.TeamID == "" {
		http.Error(w, "User is not assigned to a team", http.StatusBadRequest)
		return
	}

	teamID, err := uuid.Parse(userClaims.TeamID)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		// Default to last 7 days
		startDate = time.Now().AddDate(0, 0, -7)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.attendanceService.GetTeamAttendance(r.Context(), tenantID, teamID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get team attendance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
	})
}

// CreateAttendancePolicy handles POST /api/v1/company/attendance/policies
func (h *AttendanceHandler) CreateAttendancePolicy(w http.ResponseWriter, r *http.Request) {
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

	// Only admin and HR can create attendance policies
	if userClaims.Role != "admin" && userClaims.Role != "hr" && userClaims.Role != "super_admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var req models.AttendancePolicy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	policy, err := h.attendanceService.CreateAttendancePolicy(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}
