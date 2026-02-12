package services

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/security"
)

type EmployeeService struct {
	db            *sql.DB
	pepperSecret  string
	encryptionKey string
}

func NewEmployeeService(db *sql.DB, pepperSecret, encryptionKey string) *EmployeeService {
	return &EmployeeService{
		db:            db,
		pepperSecret:  pepperSecret,
		encryptionKey: encryptionKey,
	}
}

// CreateEmployeeRequest represents the data needed to create an employee
type CreateEmployeeRequest struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	JobTitle          string `json:"job_title"`
	DepartmentID      string `json:"department_id"`
	HireDate          string `json:"hire_date"`
	NationalID        string `json:"national_id"`
	BankAccountNumber string `json:"bank_account_number"`
	Role              string `json:"role"`
}

// CreateEmployee creates a new employee record along with user account
func (s *EmployeeService) CreateEmployee(tenantID uuid.UUID, req *CreateEmployeeRequest) (*models.Employee, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Parse hire date
	var hireDate time.Time
	if req.HireDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.HireDate); err == nil {
			hireDate = parsed
		} else {
			return nil, fmt.Errorf("invalid hire date format: %w", err)
		}
	} else {
		hireDate = time.Now()
	}

	// Generate secure temporary password (consistent for admin notification)
	tempPassword := s.generateTempPassword()
	hashedPassword, err := auth.HashPassword(tempPassword, s.pepperSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check for existing user by email (active or deleted)
	var existingUserID string
	var existingUserDeletedAt *time.Time
	err = s.db.QueryRowContext(context.Background(), "SELECT id, deleted_at FROM users WHERE email = $1", req.Email).Scan(&existingUserID, &existingUserDeletedAt)

	var userID uuid.UUID

	switch {
	case err == nil:
		if existingUserDeletedAt == nil {
			// User exists and is active
			return nil, fmt.Errorf("email address %s is already in use", req.Email)
		}

		// User exists but is soft-deleted -> Revive/Update
		userID = uuid.MustParse(existingUserID)

		role := req.Role
		if role == "" {
			role = "employee"
		}

		// Update user with new details and reactivate
		updateUserQuery := `
			UPDATE users
			SET tenant_id = $1, password_hash = $2, role = $3, first_name = $4, last_name = $5, is_active = true, deleted_at = NULL, updated_at = $6
			WHERE id = $7`

		_, err = tx.Exec(updateUserQuery, tenantID, string(hashedPassword), role, req.FirstName, req.LastName, time.Now(), userID)
		if err != nil {
			return nil, fmt.Errorf("failed to revive user account: %w", err)
		}

	case err == sql.ErrNoRows:
		// Create new user
		userID = uuid.New()
		userQuery := `
			INSERT INTO users (id, tenant_id, email, password_hash, role, first_name, last_name, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		role := req.Role
		if role == "" {
			role = "employee"
		}

		_, err = tx.Exec(userQuery,
			userID,
			tenantID,
			req.Email,
			string(hashedPassword),
			role,
			req.FirstName,
			req.LastName,
			true,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create user account: %w", err)
		}

	default:
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Parse department ID if provided
	var departmentID *uuid.UUID
	if req.DepartmentID != "" {
		if deptUUID, err := uuid.Parse(req.DepartmentID); err == nil {
			departmentID = &deptUUID
		}
	}

	// Encrypt sensitive data
	encryptedSalary, err := security.Encrypt(fmt.Sprintf("%f", 0.0), s.encryptionKey) // Default salary 0
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt salary: %w", err)
	}

	phoneVal := ""
	if req.Phone != "" {
		phoneVal = req.Phone
	}
	encryptedPhone, err := security.Encrypt(phoneVal, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt phone: %w", err)
	}

	encryptedNationalID, err := security.Encrypt(req.NationalID, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt national_id: %w", err)
	}

	nationalIDHash := security.HashBlindIndex(req.NationalID, s.encryptionKey)

	encryptedBankAccount, err := security.Encrypt(req.BankAccountNumber, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bank_account: %w", err)
	}

	// Generate employee code
	employeeCode := fmt.Sprintf("EMP%06d", time.Now().Unix()%1000000)

	// Create employee record
	employeeID := uuid.New()
	employeeQuery := `
		INSERT INTO employees (
			id, tenant_id, user_id, employee_code, department_id, job_title,
			employment_type, employment_status, date_of_joining, phone, salary,
			national_id, national_id_hash, bank_account_number,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err = tx.Exec(employeeQuery,
		employeeID,
		tenantID,
		userID,
		employeeCode,
		departmentID,
		req.JobTitle,
		"full_time",
		"active",
		hireDate,
		encryptedPhone,       // Encrypted
		encryptedSalary,      // Encrypted
		encryptedNationalID,  // Encrypted
		nationalIDHash,       // Blind Index
		encryptedBankAccount, // Encrypted
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee record: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created employee
	return s.GetEmployeeByID(tenantID, employeeID)
}

// GetEmployees retrieves all employees for a tenant with optional filtering
func (s *EmployeeService) GetEmployees(tenantID uuid.UUID, filters map[string]interface{}, claims *auth.Claims) ([]*models.Employee, error) {
	// Base query joins with users table to filter by role
	baseQuery := `
		SELECT 
			e.id, e.tenant_id, e.user_id, e.employee_code, e.department_id, e.manager_id,
			e.job_title, e.employment_type, e.employment_status, e.date_of_joining, e.date_of_leaving,
			e.salary, e.phone, e.created_at, e.updated_at,
			u.first_name, u.last_name, u.email, u.role,
			d.name as department_name
		FROM employees e
		JOIN users u ON e.user_id = u.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.tenant_id = $1 AND e.deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argIdx := 2

	// Role-based filtering (Hierarchical & Structural)
	switch claims.Role {
	case "admin":
		// Admin sees everyone (HR, Manager, TeamLead, Employee)
	case "hr":
		// HR sees Manager, TeamLead, Employee
		baseQuery += " AND u.role IN ('manager', 'team_lead', 'employee')"
	case "manager":
		// Manager sees Team Lead, Employee within their Department
		baseQuery += " AND u.role IN ('team_lead', 'employee')"
		if claims.DepartmentID != "" {
			baseQuery += fmt.Sprintf(" AND e.department_id = $%d", argIdx)
			args = append(args, claims.DepartmentID)
			argIdx++
		} else {
			// Manager without department sees no one
			return []*models.Employee{}, nil
		}
	case "team_lead":
		// Team Lead sees Employee within their Team
		baseQuery += " AND u.role = 'employee'"
		if claims.TeamID != "" {
			baseQuery += fmt.Sprintf(" AND u.team_id = $%d", argIdx)
			args = append(args, claims.TeamID)
			argIdx++
		} else {
			// TL without team sees no one
			return []*models.Employee{}, nil
		}
	case "employee":
		// Employee sees nothing via this endpoint
		return []*models.Employee{}, nil
	}

	// Apply optional filters (only if they don't conflict with enforced filters)
	if val, ok := filters["department_id"]; ok && val != "" {
		// Only apply if user is NOT manager (manager is already locked to their dept)
		if claims.Role != "manager" {
			baseQuery += fmt.Sprintf(" AND e.department_id = $%d", argIdx)
			args = append(args, val)
			argIdx++
		}
	}

	if val, ok := filters["status"]; ok && val != "" {
		baseQuery += fmt.Sprintf(" AND e.employment_status = $%d", argIdx)
		args = append(args, val)
		argIdx++
	}

	if val, ok := filters["role"]; ok && val != "" {
		// Only apply if it doesn't violate hierarchy? Service trusts caller to not pass invalid role?
		// But usually `filters["role"]` is "show me all managers".
		// If Admin asks for managers, fine.
		// If Manager asks for "team_lead", fine.
		// If Manager asks for "manager", they shouldn't see any because of line 248.
		// So strict hierarchy logic above takes precedence.
		baseQuery += fmt.Sprintf(" AND u.role = $%d", argIdx)
		args = append(args, val)
		argIdx++
	}

	baseQuery += " ORDER BY e.created_at DESC"

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}
	defer rows.Close()

	employees := []*models.Employee{}
	for rows.Next() {
		employee := &models.Employee{}
		var departmentName *string
		var encryptedSalary sql.NullString
		var encryptedPhone sql.NullString

		err := rows.Scan(
			&employee.ID,
			&employee.TenantID,
			&employee.UserID,
			&employee.EmployeeCode,
			&employee.DepartmentID,
			&employee.ManagerID,
			&employee.JobTitle,
			&employee.EmploymentType,
			&employee.EmploymentStatus,
			&employee.DateOfJoining,
			&employee.DateOfLeaving,
			&encryptedSalary,
			&encryptedPhone,
			&employee.CreatedAt,
			&employee.UpdatedAt,
			&employee.FirstName,
			&employee.LastName,
			&employee.Email,
			&employee.Role,
			&departmentName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}

		// Decrypt phone if present
		if encryptedPhone.Valid && encryptedPhone.String != "" {
			decryptedPhone, err := security.Decrypt(encryptedPhone.String, s.encryptionKey)
			if err == nil && decryptedPhone != "" {
				employee.Phone = &decryptedPhone
			}
		}

		// Decrypt salary if present
		if encryptedSalary.Valid && encryptedSalary.String != "" {
			decryptedSalary, err := security.Decrypt(encryptedSalary.String, s.encryptionKey)
			if err == nil && decryptedSalary != "" {
				// Parse float
				var salary float64
				if _, err := fmt.Sscanf(decryptedSalary, "%f", &salary); err == nil {
					employee.Salary = &salary
				}
			}
		}

		// Set department name if available
		if departmentName != nil {
			employee.Department = &models.Department{
				Name: *departmentName,
			}
		}

		employees = append(employees, employee)
	}

	return employees, nil
}

// GetEmployeeByID retrieves a specific employee by ID
func (s *EmployeeService) GetEmployeeByID(tenantID, employeeID uuid.UUID) (*models.Employee, error) {
	query := `
		SELECT 
			e.id, e.tenant_id, e.user_id, e.employee_code, e.department_id, e.manager_id,
			e.job_title, e.employment_type, e.employment_status, e.date_of_joining,
			e.date_of_leaving, e.salary, e.phone, 
			e.national_id, e.bank_account_number,
			e.created_at, e.updated_at,
			u.first_name, u.last_name, u.email, u.role,
			d.name as department_name, d.id as department_id
		FROM employees e
		INNER JOIN users u ON e.user_id = u.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.tenant_id = $1 AND e.id = $2 AND e.deleted_at IS NULL`

	employee := &models.Employee{}
	var departmentName *string
	var deptID *uuid.UUID
	var encryptedSalary, encryptedPhone, encryptedNationalID, encryptedBankAccount sql.NullString

	err := s.db.QueryRow(query, tenantID, employeeID).Scan(
		&employee.ID,
		&employee.TenantID,
		&employee.UserID,
		&employee.EmployeeCode,
		&employee.DepartmentID,
		&employee.ManagerID,
		&employee.JobTitle,
		&employee.EmploymentType,
		&employee.EmploymentStatus,
		&employee.DateOfJoining,
		&employee.DateOfLeaving,
		&encryptedSalary,
		&encryptedPhone,
		&encryptedNationalID,
		&encryptedBankAccount,
		&employee.CreatedAt,
		&employee.UpdatedAt,
		&employee.FirstName,
		&employee.LastName,
		&employee.Email,
		&employee.Role,
		&departmentName,
		&deptID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// Decrypt sensitive fields
	if encryptedPhone.Valid && encryptedPhone.String != "" {
		decrypted, err := security.Decrypt(encryptedPhone.String, s.encryptionKey)
		if err == nil && decrypted != "" {
			employee.Phone = &decrypted
		}
	}

	if encryptedSalary.Valid && encryptedSalary.String != "" {
		decrypted, err := security.Decrypt(encryptedSalary.String, s.encryptionKey)
		if err == nil && decrypted != "" {
			var salary float64
			if _, err := fmt.Sscanf(decrypted, "%f", &salary); err == nil {
				employee.Salary = &salary
			}
		}
	}

	if encryptedNationalID.Valid && encryptedNationalID.String != "" {
		decrypted, err := security.Decrypt(encryptedNationalID.String, s.encryptionKey)
		if err == nil && decrypted != "" {
			employee.NationalID = &decrypted
		}
	}

	if encryptedBankAccount.Valid && encryptedBankAccount.String != "" {
		decrypted, err := security.Decrypt(encryptedBankAccount.String, s.encryptionKey)
		if err == nil && decrypted != "" {
			employee.BankAccountNumber = &decrypted
		}
	}

	if departmentName != nil {
		employee.Department = &models.Department{
			Name: *departmentName,
		}
		if deptID != nil {
			employee.Department.ID = *deptID
		}
	}

	return employee, nil
}

// UpdateEmployee updates an existing employee
func (s *EmployeeService) UpdateEmployee(tenantID, employeeID uuid.UUID, updates map[string]interface{}) (*models.Employee, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// First get the employee to find the user_id
	var userID uuid.UUID
	err = tx.QueryRow("SELECT user_id FROM employees WHERE id = $1 AND tenant_id = $2", employeeID, tenantID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee user_id: %w", err)
	}

	// Separate user updates from employee updates
	userUpdates := make(map[string]interface{})
	employeeUpdates := make(map[string]interface{})

	for field, value := range updates {
		switch field {
		case "first_name", "last_name", "email", "role":
			userUpdates[field] = value
		case "position":
			employeeUpdates["job_title"] = value
		case "job_title", "employment_status", "phone", "department_id", "national_id", "bank_account_number":
			employeeUpdates[field] = value
		}
	}

	// Update user table if needed
	if len(userUpdates) > 0 {
		userSetParts := []string{}
		userArgs := []interface{}{}
		argIndex := 1

		for field, value := range userUpdates {
			userSetParts = append(userSetParts, fmt.Sprintf("%s = $%d", field, argIndex))
			userArgs = append(userArgs, value)
			argIndex++
		}

		userSetParts = append(userSetParts, fmt.Sprintf("updated_at = $%d", argIndex))
		userArgs = append(userArgs, time.Now())
		argIndex++

		userArgs = append(userArgs, userID)

		userQuery := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
			fmt.Sprintf("%s", userSetParts[0]), argIndex)

		// Build full query with all SET parts
		if len(userSetParts) > 1 {
			setParts := userSetParts[0]
			for i := 1; i < len(userSetParts); i++ {
				setParts += ", " + userSetParts[i]
			}
			userQuery = fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", setParts, argIndex)
		}

		_, err = tx.Exec(userQuery, userArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update employee table if needed
	if len(employeeUpdates) > 0 {
		employeeSetParts := []string{}
		employeeArgs := []interface{}{}
		argIndex := 1

		for field, value := range employeeUpdates {
			// Handle Department ID parsing
			if field == "department_id" && value != nil {
				if deptUUID, err := uuid.Parse(value.(string)); err == nil {
					employeeSetParts = append(employeeSetParts, fmt.Sprintf("%s = $%d", field, argIndex))
					employeeArgs = append(employeeArgs, deptUUID)
					argIndex++
				}
				continue
			}

			// Handle Encryption for sensitive fields
			if (field == "phone" || field == "national_id" || field == "bank_account_number") && value != nil {
				strVal, ok := value.(string)
				if !ok {
					// Skip if not string (should not happen with json decoding usually, but good to be safe)
					continue
				}

				encryptedVal, err := security.Encrypt(strVal, s.encryptionKey)
				if err != nil {
					return nil, fmt.Errorf("failed to encrypt %s: %w", field, err)
				}

				employeeSetParts = append(employeeSetParts, fmt.Sprintf("%s = $%d", field, argIndex))
				employeeArgs = append(employeeArgs, encryptedVal)
				argIndex++

				// If National ID changed, update hash too
				if field == "national_id" {
					hash := security.HashBlindIndex(strVal, s.encryptionKey)
					employeeSetParts = append(employeeSetParts, fmt.Sprintf("national_id_hash = $%d", argIndex))
					employeeArgs = append(employeeArgs, hash)
					argIndex++
				}
				continue
			}

			// Default handling for other fields
			employeeSetParts = append(employeeSetParts, fmt.Sprintf("%s = $%d", field, argIndex))
			employeeArgs = append(employeeArgs, value)
			argIndex++
		}

		employeeSetParts = append(employeeSetParts, fmt.Sprintf("updated_at = $%d", argIndex))
		employeeArgs = append(employeeArgs, time.Now())
		argIndex++

		employeeArgs = append(employeeArgs, employeeID)

		if len(employeeSetParts) > 0 {
			setParts := employeeSetParts[0]
			for i := 1; i < len(employeeSetParts); i++ {
				setParts += ", " + employeeSetParts[i]
			}
			employeeQuery := fmt.Sprintf("UPDATE employees SET %s WHERE id = $%d AND deleted_at IS NULL", setParts, argIndex)

			_, err = tx.Exec(employeeQuery, employeeArgs...)
			if err != nil {
				return nil, fmt.Errorf("failed to update employee: %w", err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated employee
	return s.GetEmployeeByID(tenantID, employeeID)
}

// DeleteEmployee hard deletes an employee (and their user account)
func (s *EmployeeService) DeleteEmployee(tenantID, employeeID uuid.UUID) error {
	// 1. Get User ID associated with the employee
	var userID uuid.UUID
	queryGet := `SELECT user_id FROM employees WHERE id = $1 AND tenant_id = $2`
	err := s.db.QueryRow(queryGet, employeeID, tenantID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("employee not found")
		}
		return fmt.Errorf("failed to get user id for employee: %w", err)
	}

	// 2. Hard Delete the User (Cascades to Employee)
	queryDelete := `DELETE FROM users WHERE id = $1 AND tenant_id = $2`
	result, err := s.db.Exec(queryDelete, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete user/employee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// GetDepartments retrieves all departments for a tenant
func (s *EmployeeService) GetDepartments(tenantID uuid.UUID) ([]*models.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, head_id, parent_id, is_active, created_at, updated_at
		FROM departments 
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY name`

	rows, err := s.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		dept := &models.Department{}
		err := rows.Scan(
			&dept.ID,
			&dept.TenantID,
			&dept.Name,
			&dept.Description,
			&dept.HeadID,
			&dept.ParentID,
			&dept.IsActive,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}
		departments = append(departments, dept)
	}

	return departments, nil
}

// GetTeamMembers retrieves all employees belonging to a specific team
func (s *EmployeeService) GetTeamMembers(tenantID, teamID uuid.UUID) ([]*models.Employee, error) {
	query := `
		SELECT 
			e.id, e.tenant_id, e.user_id, e.employee_code, e.department_id, e.manager_id,
			e.job_title, e.employment_type, e.employment_status, e.date_of_joining,
			e.date_of_leaving, e.salary, e.phone, e.created_at, e.updated_at,
			u.first_name, u.last_name, u.email, u.role,
			d.name as department_name
		FROM employees e
		INNER JOIN users u ON e.user_id = u.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.tenant_id = $1 AND u.team_id = $2 AND e.deleted_at IS NULL
		ORDER BY u.first_name, u.last_name`

	rows, err := s.db.Query(query, tenantID, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	var employees []*models.Employee
	for rows.Next() {
		employee := &models.Employee{}
		var departmentName *string

		err := rows.Scan(
			&employee.ID,
			&employee.TenantID,
			&employee.UserID,
			&employee.EmployeeCode,
			&employee.DepartmentID,
			&employee.ManagerID,
			&employee.JobTitle,
			&employee.EmploymentType,
			&employee.EmploymentStatus,
			&employee.DateOfJoining,
			&employee.DateOfLeaving,
			&employee.Salary,
			&employee.Phone,
			&employee.CreatedAt,
			&employee.UpdatedAt,
			&employee.FirstName,
			&employee.LastName,
			&employee.Email,
			&employee.Role,
			&departmentName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		if departmentName != nil {
			employee.Department = &models.Department{
				Name: *departmentName,
			}
		}

		employees = append(employees, employee)
	}

	return employees, nil
}

// UpdateEmployeeStatus updates the active status and/or employment status of an employee
// This method updates both the users table (is_active) and employees table (employment_status)
// in a single transaction to ensure consistency
func (s *EmployeeService) UpdateEmployeeStatus(tenantID, employeeID uuid.UUID, updates map[string]interface{}) (*models.Employee, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get employee to find associated user_id
	var userID uuid.UUID
	err = tx.QueryRow(`
		SELECT user_id 
		FROM employees 
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, employeeID, tenantID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// Update users table if is_active is provided
	if isActive, ok := updates["is_active"]; ok {
		_, err = tx.Exec(`
			UPDATE users 
			SET is_active = $1, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $2 AND tenant_id = $3
		`, isActive, userID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to update user status: %w", err)
		}
	}

	// Update employees table if employment_status is provided
	if employmentStatus, ok := updates["employment_status"]; ok {
		_, err = tx.Exec(`
			UPDATE employees 
			SET employment_status = $1, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $2 AND tenant_id = $3
		`, employmentStatus, employeeID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to update employment status: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch and return updated employee
	return s.GetEmployeeByID(tenantID, employeeID)
}

// generateTempPassword generates a secure temporary password
func (s *EmployeeService) generateTempPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
