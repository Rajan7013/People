package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PayslipService struct {
	DB *sqlx.DB
}

func NewPayslipService(db *sqlx.DB) *PayslipService {
	return &PayslipService{DB: db}
}

// GetPayslipsByTenant gets all payslips for a tenant with optional filtering
func (s *PayslipService) GetPayslipsByTenant(tenantID uuid.UUID, filter *models.PayslipFilter, page, limit int) ([]models.Payslip, int, error) {
	log.Printf("GetPayslipsByTenant called for TenantID: %s", tenantID)

	query := `
		SELECT 
			p.id, p.tenant_id, p.employee_id, p.salary_structure_id, 
			p.pay_period_start, p.pay_period_end, p.payment_date, 
			p.basic_salary, p.gross_salary, p.total_deductions, 
			p.net_salary, p.status, p.notes, p.created_at, p.updated_at,
			u.first_name, u.last_name, e.employee_code as emp_code, u.email, u.role, d.name as department_name
		FROM payslips p
		JOIN employees e ON p.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE p.tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 1

	if filter != nil {
		if filter.EmployeeID != nil {
			argIndex++
			query += fmt.Sprintf(" AND p.employee_id = $%d", argIndex)
			args = append(args, *filter.EmployeeID)
		}
		if filter.DepartmentID != nil {
			argIndex++
			query += fmt.Sprintf(" AND e.department_id = $%d", argIndex)
			args = append(args, *filter.DepartmentID)
		}
		if filter.Status != nil {
			argIndex++
			query += fmt.Sprintf(" AND p.status = $%d", argIndex)
			args = append(args, *filter.Status)
		}
		if filter.PayPeriodStart != nil {
			argIndex++
			query += fmt.Sprintf(" AND p.pay_period_start >= $%d", argIndex)
			args = append(args, *filter.PayPeriodStart)
		}
		if filter.PayPeriodEnd != nil {
			argIndex++
			query += fmt.Sprintf(" AND p.pay_period_end <= $%d", argIndex)
			args = append(args, *filter.PayPeriodEnd)
		}
		if filter.PaymentMonth != nil && filter.PaymentYear != nil {
			argIndex++
			query += fmt.Sprintf(" AND EXTRACT(MONTH FROM p.payment_date) = $%d", argIndex)
			args = append(args, *filter.PaymentMonth)
			argIndex++
			query += fmt.Sprintf(" AND EXTRACT(YEAR FROM p.payment_date) = $%d", argIndex)
			args = append(args, *filter.PaymentYear)
		}
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM (` + query + `) as filtered`
	var totalCount int
	log.Printf("Executing Count Query: %s", countQuery)
	err := s.DB.Get(&totalCount, countQuery, args...)
	if err != nil {
		log.Printf("COUNT QUERY ERROR: %v", err)
		return nil, 0, err
	}
	log.Printf("Count success, total: %d", totalCount)

	// Add pagination
	query += " ORDER BY p.pay_period_start DESC"
	if limit > 0 {
		offset := (page - 1) * limit
		argIndex++
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	log.Printf("Executing Main Query: %s", query)
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Printf("MAIN QUERY ERROR: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	payslips := make([]models.Payslip, 0)
	for rows.Next() {
		var p models.Payslip
		var emp models.Employee
		var deptName sql.NullString

		var salaryStructureID sql.NullString
		var paymentDate sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&p.ID, &p.TenantID, &p.EmployeeID, &salaryStructureID,
			&p.PayPeriodStart, &p.PayPeriodEnd, &paymentDate,
			&p.BasicSalary, &p.GrossSalary, &p.TotalDeductions,
			&p.NetSalary, &p.Status, &notes, &p.CreatedAt, &p.UpdatedAt,
			&emp.FirstName, &emp.LastName, &emp.EmployeeCode, &emp.Email, &p.Role, &deptName,
		)
		if err != nil {
			log.Printf("SCAN ERROR in GetPayslipsByTenant: %v", err)
			return nil, 0, err
		}

		// Handle nullable fields
		if salaryStructureID.Valid {
			if id, err := uuid.Parse(salaryStructureID.String); err == nil {
				p.SalaryStructureID = &id
			}
		}
		if paymentDate.Valid {
			p.PaymentDate = &paymentDate.Time
		}
		if notes.Valid {
			p.Notes = &notes.String
		}

		emp.ID = p.EmployeeID
		if deptName.Valid {
			emp.Department = &models.Department{Name: deptName.String}
		}
		p.Employee = &emp

		// Load components
		p.Components, err = s.getPayslipComponents(p.ID)
		if err != nil {
			log.Printf("Error loading payslip components: %v", err)
		}

		payslips = append(payslips, p)
	}

	return payslips, totalCount, nil
}

// GetPayslipsByEmployee gets payslips for a specific employee
func (s *PayslipService) GetPayslipsByEmployee(tenantID, employeeID uuid.UUID, page, limit int) ([]models.Payslip, int, error) {
	filter := &models.PayslipFilter{
		EmployeeID: &employeeID,
	}
	return s.GetPayslipsByTenant(tenantID, filter, page, limit)
}

// GetPayslipByID gets a single payslip by ID
func (s *PayslipService) GetPayslipByID(tenantID, payslipID uuid.UUID) (*models.Payslip, error) {
	query := `
		SELECT 
			p.id, p.tenant_id, p.employee_id, p.salary_structure_id, 
			p.pay_period_start, p.pay_period_end, p.payment_date, 
			p.basic_salary, p.gross_salary, p.total_deductions, 
			p.net_salary, p.status, p.notes, p.created_at, p.updated_at,
			u.first_name, u.last_name, e.employee_code as emp_code, u.email, u.role, d.name as department_name
		FROM payslips p
		JOIN employees e ON p.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE p.tenant_id = $1 AND p.id = $2`

	var p models.Payslip
	var emp models.Employee
	var deptName sql.NullString

	var salaryStructureID sql.NullString
	var paymentDate sql.NullTime
	var notes sql.NullString

	err := s.DB.QueryRow(query, tenantID, payslipID).Scan(
		&p.ID, &p.TenantID, &p.EmployeeID, &salaryStructureID,
		&p.PayPeriodStart, &p.PayPeriodEnd, &paymentDate,
		&p.BasicSalary, &p.GrossSalary, &p.TotalDeductions,
		&p.NetSalary, &p.Status, &notes, &p.CreatedAt, &p.UpdatedAt,
		&emp.FirstName, &emp.LastName, &emp.EmployeeCode, &emp.Email, &p.Role, &deptName,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if salaryStructureID.Valid {
		if id, err := uuid.Parse(salaryStructureID.String); err == nil {
			p.SalaryStructureID = &id
		}
	}
	if paymentDate.Valid {
		p.PaymentDate = &paymentDate.Time
	}
	if notes.Valid {
		p.Notes = &notes.String
	}

	emp.ID = p.EmployeeID
	if deptName.Valid {
		emp.Department = &models.Department{Name: deptName.String}
	}
	p.Employee = &emp

	// Load components
	p.Components, err = s.getPayslipComponents(p.ID)
	if err != nil {
		log.Printf("Error loading payslip components: %v", err)
	}

	return &p, nil
}

// CreatePayslip creates a new payslip
func (s *PayslipService) CreatePayslip(tenantID uuid.UUID, req *models.PayslipCreateRequest) (*models.Payslip, error) {
	// Get employee's current salary structure
	salaryStructure, err := s.getCurrentSalaryStructure(tenantID, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get salary structure: %v", err)
	}

	if salaryStructure == nil {
		return nil, fmt.Errorf("no active salary structure found for employee")
	}

	// Calculate payslip amounts
	grossSalary, totalDeductions, netSalary, components, err := s.calculatePayslip(salaryStructure)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate payslip: %v", err)
	}

	// Create payslip
	payslipID := uuid.New()
	payslip := &models.Payslip{
		ID:                payslipID,
		TenantID:          tenantID,
		EmployeeID:        req.EmployeeID,
		SalaryStructureID: &salaryStructure.ID,
		PayPeriodStart:    req.PayPeriodStart,
		PayPeriodEnd:      req.PayPeriodEnd,
		PaymentDate:       req.PaymentDate,
		BasicSalary:       salaryStructure.BasicSalary,
		GrossSalary:       grossSalary,
		TotalDeductions:   totalDeductions,
		NetSalary:         netSalary,
		Status:            "draft",
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert payslip
	_, err = tx.NamedExec(`
		INSERT INTO payslips (id, tenant_id, employee_id, salary_structure_id, 
			pay_period_start, pay_period_end, payment_date, basic_salary, 
			gross_salary, total_deductions, net_salary, status, notes)
		VALUES (:id, :tenant_id, :employee_id, :salary_structure_id, 
			:pay_period_start, :pay_period_end, :payment_date, :basic_salary, 
			:gross_salary, :total_deductions, :net_salary, :status, :notes)`,
		payslip)
	if err != nil {
		return nil, err
	}

	// Insert payslip components
	for _, component := range components {
		component.PayslipID = payslipID
		component.ID = uuid.New()
		_, err = tx.NamedExec(`
			INSERT INTO payslip_components (id, payslip_id, component_id, 
				component_name, component_type, amount)
			VALUES (:id, :payslip_id, :component_id, :component_name, 
				:component_type, :amount)`,
			component)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	payslip.Components = components
	return payslip, nil
}

// UpdatePayslip updates a payslip
func (s *PayslipService) UpdatePayslip(tenantID, payslipID uuid.UUID, req *models.PayslipUpdateRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 0

	if req.PaymentDate != nil {
		argIndex++
		setParts = append(setParts, fmt.Sprintf("payment_date = $%d", argIndex))
		args = append(args, *req.PaymentDate)
	}

	if req.Status != nil {
		argIndex++
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
	}

	if req.Notes != nil {
		argIndex++
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *req.Notes)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	argIndex++
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE payslips SET %s WHERE tenant_id = $%d AND id = $%d",
		fmt.Sprintf("%s", setParts), argIndex+1, argIndex+2)
	args = append(args, tenantID, payslipID)

	_, err := s.DB.Exec(query, args...)
	return err
}

// DeletePayslip deletes a payslip
func (s *PayslipService) DeletePayslip(tenantID, payslipID uuid.UUID) error {
	_, err := s.DB.Exec("DELETE FROM payslips WHERE tenant_id = $1 AND id = $2", tenantID, payslipID)
	return err
}

// GetPayslipStats gets payslip statistics for the tenant
func (s *PayslipService) GetPayslipStats(tenantID uuid.UUID) (*models.PayslipStats, error) {
	stats := &models.PayslipStats{}

	// Get basic stats
	err := s.DB.QueryRow(`
		SELECT 
			COUNT(*) as total_payslips,
			COUNT(CASE WHEN status = 'draft' THEN 1 END) as pending_approval,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as processed_payslips,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN net_salary ELSE 0 END), 0) as total_payroll,
			COALESCE(AVG(CASE WHEN status = 'paid' THEN net_salary END), 0) as average_salary
		FROM payslips 
		WHERE tenant_id = $1`, tenantID).Scan(
		&stats.TotalPayslips,
		&stats.PendingApproval,
		&stats.ProcessedPayslips,
		&stats.TotalPayroll,
		&stats.AverageSalary,
	)

	return stats, err
}

// Helper functions

func (s *PayslipService) getPayslipComponents(payslipID uuid.UUID) ([]models.PayslipComponent, error) {
	query := `
		SELECT id, payslip_id, component_id, component_name, component_type, amount, created_at
		FROM payslip_components
		WHERE payslip_id = $1
		ORDER BY component_type, component_name`

	var components []models.PayslipComponent
	err := s.DB.Select(&components, query, payslipID)
	return components, err
}

func (s *PayslipService) getCurrentSalaryStructure(tenantID, employeeID uuid.UUID) (*models.EmployeeSalaryStructure, error) {
	query := `
		SELECT id, tenant_id, employee_id, basic_salary, effective_date, end_date, is_active, created_at, updated_at
		FROM employee_salary_structures
		WHERE tenant_id = $1 AND employee_id = $2 AND is_active = true
		AND effective_date <= CURRENT_DATE
		AND (end_date IS NULL OR end_date > CURRENT_DATE)
		ORDER BY effective_date DESC
		LIMIT 1`

	var structure models.EmployeeSalaryStructure
	err := s.DB.Get(&structure, query, tenantID, employeeID)
	if err != nil {
		return nil, err
	}

	// Load components
	componentQuery := `
		SELECT ssc.id, ssc.salary_structure_id, ssc.component_id, ssc.amount, ssc.percentage, ssc.created_at,
			sc.name, sc.type, sc.is_percentage, sc.is_taxable
		FROM salary_structure_components ssc
		JOIN salary_components sc ON ssc.component_id = sc.id
		WHERE ssc.salary_structure_id = $1
		ORDER BY sc.type, sc.name`

	rows, err := s.DB.Query(componentQuery, structure.ID)
	if err != nil {
		return &structure, nil // Return structure without components if query fails
	}
	defer rows.Close()

	for rows.Next() {
		var comp models.SalaryStructureComponent
		var sc models.SalaryComponent

		err := rows.Scan(
			&comp.ID, &comp.SalaryStructureID, &comp.ComponentID,
			&comp.Amount, &comp.Percentage, &comp.CreatedAt,
			&sc.Name, &sc.Type, &sc.IsPercentage, &sc.IsTaxable,
		)
		if err != nil {
			continue
		}

		sc.ID = comp.ComponentID
		comp.Component = &sc
		structure.Components = append(structure.Components, comp)
	}

	return &structure, nil
}

func (s *PayslipService) calculatePayslip(structure *models.EmployeeSalaryStructure) (float64, float64, float64, []models.PayslipComponent, error) {
	basicSalary := structure.BasicSalary
	totalEarnings := basicSalary
	totalDeductions := 0.0
	var components []models.PayslipComponent

	// Add basic salary component
	components = append(components, models.PayslipComponent{
		ComponentID:   uuid.New(), // This should be a predefined basic salary component ID
		ComponentName: "Basic Salary",
		ComponentType: "earning",
		Amount:        basicSalary,
		CreatedAt:     time.Now(),
	})

	// Calculate other components
	for _, structComp := range structure.Components {
		if structComp.Component == nil {
			continue
		}

		var amount float64
		if structComp.Component.IsPercentage {
			if structComp.Percentage != nil {
				amount = basicSalary * (*structComp.Percentage / 100)
			}
		} else {
			if structComp.Amount != nil {
				amount = *structComp.Amount
			}
		}

		// Round to 2 decimal places
		amount = math.Round(amount*100) / 100

		component := models.PayslipComponent{
			ComponentID:   structComp.ComponentID,
			ComponentName: structComp.Component.Name,
			ComponentType: structComp.Component.Type,
			Amount:        amount,
			CreatedAt:     time.Now(),
		}

		components = append(components, component)

		switch structComp.Component.Type {
		case "earning":
			totalEarnings += amount
		case "deduction":
			totalDeductions += amount
		}
	}

	grossSalary := totalEarnings
	netSalary := grossSalary - totalDeductions

	// Round final amounts
	grossSalary = math.Round(grossSalary*100) / 100
	totalDeductions = math.Round(totalDeductions*100) / 100
	netSalary = math.Round(netSalary*100) / 100

	return grossSalary, totalDeductions, netSalary, components, nil
}
