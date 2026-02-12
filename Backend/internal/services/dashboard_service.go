package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
)

type DashboardService struct {
	db *sql.DB
}

func NewDashboardService(db *sql.DB) *DashboardService {
	return &DashboardService{db: db}
}

type DashboardStats struct {
	TotalEmployees  int `json:"total_employees"`
	PresentToday    int `json:"present_today"`
	OnLeave         int `json:"on_leave"`
	PendingRequests int `json:"pending_requests"`
}

func (s *DashboardService) GetStats(claims *auth.Claims) (*DashboardStats, error) {
	stats := &DashboardStats{}
	tenantID := claims.TenantID
	userID := claims.UserID

	// Helper to build role-based filters
	// Returns: extraJoin, whereClause, args
	buildFilter := func(baseArgs []interface{}) (string, string, []interface{}) {
		query := ""
		args := baseArgs
		argIdx := len(args) + 1

		// Exclude self from counts
		query += fmt.Sprintf(" AND u.id != $%d", argIdx)
		args = append(args, userID)
		argIdx++

		switch claims.Role {
		case "admin":
			// Admin sees all active employees (excluding themselves)
			// No extra filters needed beyond tenant
		case "hr":
			// HR sees Manager, Team Lead, Employee
			query += " AND u.role IN ('manager', 'team_lead', 'employee')"
		case "manager":
			// Manager sees their Department (Team Lead, Employee)
			if claims.DepartmentID != "" {
				query += fmt.Sprintf(" AND e.department_id = $%d", argIdx)
				args = append(args, claims.DepartmentID)
				argIdx++
			} else {
				// If Manager has no department, they see no one
				query += " AND 1=0"
			}
			query += " AND u.role IN ('team_lead', 'employee')"
		case "team_lead":
			// Team Lead sees their Team (Employee)
			if claims.TeamID != "" {
				query += fmt.Sprintf(" AND u.team_id = $%d", argIdx)
				args = append(args, claims.TeamID)
				argIdx++
			} else {
				// If Team Lead has no team, they see no one
				query += " AND 1=0"
			}
			query += " AND u.role = 'employee'"
		case "employee":
			// Employee sees nothing (or just valid 0s)
			query += " AND 1=0"
		}
		return "", query, args
	}

	// 1. Total Employees
	// Base Query
	baseQuery := `
		SELECT COUNT(*) 
		FROM employees e
		JOIN users u ON e.user_id = u.id
		WHERE e.tenant_id = $1 
		AND e.employment_status = 'active'
	`
	_, filter, args := buildFilter([]interface{}{tenantID})
	err := s.db.QueryRow(baseQuery+filter, args...).Scan(&stats.TotalEmployees)
	if err != nil {
		return nil, err
	}

	// 2. Present Today (Attendance)
	// Need to join with users/employees to apply RBAC
	attendanceQuery := `
		SELECT COUNT(ar.id) 
		FROM attendance_records ar
		JOIN employees e ON ar.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE ar.tenant_id = $1 
		AND ar.date = $2 
		AND ar.status = 'present'
	`
	startOfDay := time.Now().Truncate(24 * time.Hour)
	_, attFilter, attArgs := buildFilter([]interface{}{tenantID, startOfDay})
	err = s.db.QueryRow(attendanceQuery+attFilter, attArgs...).Scan(&stats.PresentToday)
	if err != nil {
		return nil, err
	}

	// 3. On Leave (Approved leaves containing today)
	leaveQuery := `
		SELECT COUNT(la.id) 
		FROM leave_applications la
		JOIN employees e ON la.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE la.tenant_id = $1 
		AND la.status = 'approved'
		AND la.start_date <= $2 AND la.end_date >= $2
	`
	today := time.Now().Format("2006-01-02")
	_, leaveFilter, leaveArgs := buildFilter([]interface{}{tenantID, today})
	err = s.db.QueryRow(leaveQuery+leaveFilter, leaveArgs...).Scan(&stats.OnLeave)
	if err != nil {
		return nil, err
	}

	// 4. Pending Requests
	requestQuery := `
		SELECT COUNT(la.id) 
		FROM leave_applications la
		JOIN employees e ON la.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE la.tenant_id = $1 
		AND la.status = 'pending'
	`
	_, reqFilter, reqArgs := buildFilter([]interface{}{tenantID})
	err = s.db.QueryRow(requestQuery+reqFilter, reqArgs...).Scan(&stats.PendingRequests)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
