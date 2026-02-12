package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
)

type EmployeeHandler struct {
	employeeService *services.EmployeeService
}

func NewEmployeeHandler(employeeService *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
	}
}

// CreateEmployeeRequest represents the request body for creating an employee
type CreateEmployeeRequest struct {
	FirstName    string `json:"first_name" validate:"required"`
	LastName     string `json:"last_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone"`
	Position     string `json:"position"`
	Department   string `json:"department"`
	DepartmentID string `json:"department_id"`
	HireDate     string `json:"hire_date"`
	Role         string `json:"role"`
}

// EmployeeResponse represents the response format for employee data
type EmployeeResponse struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Phone        string `json:"phone,omitempty"`
	Position     string `json:"position,omitempty"`
	Department   string `json:"department,omitempty"`
	DepartmentID string `json:"department_id,omitempty"`
	HireDate     string `json:"hire_date,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreateEmployee handles POST /api/v1/{tenantId}/employees
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
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

	// Only admin, hr, and manager can create employees
	allowedRoles := []string{"admin", "hr", "manager"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		http.Error(w, "First name, last name, and email are required", http.StatusBadRequest)
		return
	}

	// Role hierarchy validation: Only super admin can create admin users
	if req.Role == "admin" && userClaims.Role != "super_admin" {
		http.Error(w, "Only super admins can create admin users", http.StatusForbidden)
		return
	}

	// Create service request
	serviceReq := &services.CreateEmployeeRequest{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		JobTitle:     req.Position,
		DepartmentID: req.DepartmentID,
		HireDate:     req.HireDate,
		Role:         req.Role, // Pass role
	}

	// Create employee
	createdEmployee, err := h.employeeService.CreateEmployee(tenantID, serviceReq)
	if err != nil {
		// Log detailed error for debugging
		fmt.Printf("ERROR creating employee: %v\n", err)

		// Check for duplicate email error
		if strings.Contains(err.Error(), "already in use") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "This email address is already in use",
			})
			return
		}

		http.Error(w, "Failed to create employee: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	response := &EmployeeResponse{
		ID:        createdEmployee.ID.String(),
		FirstName: createdEmployee.FirstName,
		LastName:  createdEmployee.LastName,
		Email:     createdEmployee.Email,
		Status:    createdEmployee.EmploymentStatus,
		Role:      createdEmployee.Role,
		CreatedAt: createdEmployee.CreatedAt.Format(time.RFC3339),
		UpdatedAt: createdEmployee.UpdatedAt.Format(time.RFC3339),
	}

	if createdEmployee.Phone != nil {
		response.Phone = *createdEmployee.Phone
	}
	if createdEmployee.JobTitle != nil {
		response.Position = *createdEmployee.JobTitle
	}
	if createdEmployee.Department != nil {
		response.Department = createdEmployee.Department.Name
		response.DepartmentID = createdEmployee.Department.ID.String()
	}
	if createdEmployee.DateOfJoining != nil {
		response.HireDate = createdEmployee.DateOfJoining.Format("2006-01-02")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetEmployees handles GET /api/v1/{tenantId}/employees
func (h *EmployeeHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
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

	// Extract filters from query params
	filters := make(map[string]interface{})

	if deptID := r.URL.Query().Get("department_id"); deptID != "" {
		filters["department_id"] = deptID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}

	if role := r.URL.Query().Get("role"); role != "" {
		filters["role"] = role
	}

	// Get viewer role from context
	userClaims, ok = auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	employees, err := h.employeeService.GetEmployees(tenantID, filters, userClaims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	response := make([]*EmployeeResponse, 0)
	for _, emp := range employees {
		empResp := &EmployeeResponse{
			ID:        emp.ID.String(),
			FirstName: emp.FirstName,
			LastName:  emp.LastName,
			Email:     emp.Email,
			Status:    emp.EmploymentStatus,
			Role:      emp.Role,
			CreatedAt: emp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: emp.UpdatedAt.Format(time.RFC3339),
		}

		if emp.Phone != nil {
			empResp.Phone = *emp.Phone
		}
		if emp.JobTitle != nil {
			empResp.Position = *emp.JobTitle
		}
		if emp.Department != nil {
			empResp.Department = emp.Department.Name
			empResp.DepartmentID = emp.Department.ID.String()
		}
		if emp.DateOfJoining != nil {
			empResp.HireDate = emp.DateOfJoining.Format("2006-01-02")
		}

		response = append(response, empResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmployee handles GET /api/v1/{tenantId}/employees/{employeeId}
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
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

	employeeIDStr := chi.URLParam(r, "employeeID")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	employee, err := h.employeeService.GetEmployeeByID(tenantID, employeeID)
	if err != nil {
		if err.Error() == "employee not found" {
			http.Error(w, "Employee not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get employee: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Build response
	response := &EmployeeResponse{
		ID:        employee.ID.String(),
		FirstName: employee.FirstName,
		LastName:  employee.LastName,
		Email:     employee.Email,
		Status:    employee.EmploymentStatus,
		Role:      employee.Role,
		CreatedAt: employee.CreatedAt.Format(time.RFC3339),
		UpdatedAt: employee.UpdatedAt.Format(time.RFC3339),
	}

	if employee.Phone != nil {
		response.Phone = *employee.Phone
	}
	if employee.JobTitle != nil {
		response.Position = *employee.JobTitle
	}
	if employee.Department != nil {
		response.Department = employee.Department.Name
		response.DepartmentID = employee.Department.ID.String()
	}
	if employee.DateOfJoining != nil {
		response.HireDate = employee.DateOfJoining.Format("2006-01-02")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateEmployee handles PUT /api/v1/{tenantId}/employees/{employeeId}
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
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

	// Only admin, hr, and manager can update employees
	allowedRoles := []string{"admin", "hr", "manager"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	employeeIDStr := chi.URLParam(r, "employeeID")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedEmployee, err := h.employeeService.UpdateEmployee(tenantID, employeeID, updates)
	if err != nil {
		if err.Error() == "employee not found" {
			http.Error(w, "Employee not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update employee: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Build response
	response := &EmployeeResponse{
		ID:        updatedEmployee.ID.String(),
		FirstName: updatedEmployee.FirstName,
		LastName:  updatedEmployee.LastName,
		Email:     updatedEmployee.Email,
		Status:    updatedEmployee.EmploymentStatus,
		Role:      updatedEmployee.Role,
		CreatedAt: updatedEmployee.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updatedEmployee.UpdatedAt.Format(time.RFC3339),
	}

	if updatedEmployee.Phone != nil {
		response.Phone = *updatedEmployee.Phone
	}
	if updatedEmployee.JobTitle != nil {
		response.Position = *updatedEmployee.JobTitle
	}
	if updatedEmployee.Department != nil {
		response.Department = updatedEmployee.Department.Name
		response.DepartmentID = updatedEmployee.Department.ID.String()
	}
	if updatedEmployee.DateOfJoining != nil {
		response.HireDate = updatedEmployee.DateOfJoining.Format("2006-01-02")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteEmployee handles DELETE /api/v1/{tenantId}/employees/{employeeId}
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
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

	// Only admin and hr can delete employees
	allowedRoles := []string{"admin", "hr"}
	if !slices.Contains(allowedRoles, userClaims.Role) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	employeeIDStr := chi.URLParam(r, "employeeID")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	err = h.employeeService.DeleteEmployee(tenantID, employeeID)
	if err != nil {
		if err.Error() == "employee not found" {
			http.Error(w, "Employee not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete employee: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDepartments handles GET /api/v1/{tenantId}/departments
func (h *EmployeeHandler) GetDepartments(w http.ResponseWriter, r *http.Request) {
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

	departments, err := h.employeeService.GetDepartments(tenantID)
	if err != nil {
		http.Error(w, "Failed to get departments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(departments)
}

// GetMyTeam handles GET /api/v1/company/team-lead/team and /manager/team
func (h *EmployeeHandler) GetMyTeam(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: GetMyTeam called") // Debug log
	userClaims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		fmt.Println("DEBUG: Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Printf("DEBUG: Role=%s, TeamID=%s, DeptID=%s\n", userClaims.Role, userClaims.TeamID, userClaims.DepartmentID)

	tenantID, err := uuid.Parse(userClaims.TenantID)
	if err != nil {
		fmt.Println("DEBUG: Invalid Tenant ID")
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var employees []*models.Employee

	// Build response (reuse existing response structure)
	response := make([]*EmployeeResponse, 0)

	if userClaims.Role == "manager" {
		if userClaims.DepartmentID == "" {
			// Return empty list if manager has no department
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		filters := map[string]interface{}{
			"department_id": userClaims.DepartmentID,
		}

		employees, err = h.employeeService.GetEmployees(tenantID, filters, userClaims)
		if err != nil {
			http.Error(w, "Failed to get department members: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Assume Team Lead (or others using this endpoint in future)
		// efficient check from claims
		if userClaims.TeamID == "" {
			// Return empty list if user has no team
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		teamID, err := uuid.Parse(userClaims.TeamID)
		if err != nil {
			// Should not happen if confirmed non-empty string in claims, but safe check
			http.Error(w, "Invalid team ID in claims", http.StatusBadRequest)
			return
		}

		employees, err = h.employeeService.GetTeamMembers(tenantID, teamID)
		if err != nil {
			http.Error(w, "Failed to get team members: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, emp := range employees {
		empResp := &EmployeeResponse{
			ID:        emp.ID.String(),
			FirstName: emp.FirstName,
			LastName:  emp.LastName,
			Email:     emp.Email,
			Status:    emp.EmploymentStatus,
			Role:      emp.Role,
			CreatedAt: emp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: emp.UpdatedAt.Format(time.RFC3339),
		}

		if emp.Phone != nil {
			empResp.Phone = *emp.Phone
		}
		if emp.JobTitle != nil {
			empResp.Position = *emp.JobTitle
		}
		if emp.Department != nil {
			empResp.Department = emp.Department.Name
		}
		if emp.DateOfJoining != nil {
			empResp.HireDate = emp.DateOfJoining.Format("2006-01-02")
		}

		response = append(response, empResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateEmployeeStatusRequest represents the request body for updating employee status
type UpdateEmployeeStatusRequest struct {
	IsActive         *bool  `json:"is_active"`
	EmploymentStatus string `json:"employment_status"`
}

// UpdateEmployeeStatus handles PUT /api/v1/company/admin/employees/{employeeID}/status
func (h *EmployeeHandler) UpdateEmployeeStatus(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")
	if employeeID == "" {
		http.Error(w, "Employee ID is required", http.StatusBadRequest)
		return
	}

	empUUID, err := uuid.Parse(employeeID)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

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

	var req UpdateEmployeeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate employment status if provided
	if req.EmploymentStatus != "" {
		validStatuses := []string{"active", "inactive", "suspended", "terminated"}
		if !slices.Contains(validStatuses, req.EmploymentStatus) {
			http.Error(w, "Invalid employment status. Must be one of: active, inactive, suspended, terminated", http.StatusBadRequest)
			return
		}
	}

	// Update employee status
	updates := make(map[string]interface{})
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.EmploymentStatus != "" {
		updates["employment_status"] = req.EmploymentStatus
	}

	if len(updates) == 0 {
		http.Error(w, "No updates provided", http.StatusBadRequest)
		return
	}

	employee, err := h.employeeService.UpdateEmployeeStatus(tenantID, empUUID, updates)
	if err != nil {
		fmt.Printf("ERROR updating employee status: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}
