package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
)

type DepartmentService struct {
	db *sql.DB
}

func NewDepartmentService(db *sql.DB) *DepartmentService {
	return &DepartmentService{
		db: db,
	}
}

// CreateDepartmentRequest represents data needed to create a department
type CreateDepartmentRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	HeadID      *string `json:"head_id"`
}

// CreateDepartment creates a new department or revives a deleted one
func (s *DepartmentService) CreateDepartment(tenantID uuid.UUID, req *CreateDepartmentRequest) (*models.Department, error) {
	// Check if department with this name already exists in the tenant
	var existingID string
	var existingDeletedAt *time.Time
	queryCheck := `SELECT id, deleted_at FROM departments WHERE tenant_id = $1 AND name = $2`
	err := s.db.QueryRow(queryCheck, tenantID, req.Name).Scan(&existingID, &existingDeletedAt)

	if err == nil {
		if existingDeletedAt == nil {
			// Department exists and is active
			return nil, fmt.Errorf("department with name '%s' already exists", req.Name)
		}

		// Department exists but is soft-deleted -> Revive it
		reviveQuery := `
			UPDATE departments 
			SET deleted_at = NULL, description = $1, parent_id = $2, head_id = $3, is_active = true, updated_at = $4
			WHERE id = $5`

		var parentID *uuid.UUID
		if req.ParentID != nil && *req.ParentID != "" {
			if pid, err := uuid.Parse(*req.ParentID); err == nil {
				parentID = &pid
			}
		}

		var headID *uuid.UUID
		if req.HeadID != nil && *req.HeadID != "" {
			if hid, err := uuid.Parse(*req.HeadID); err == nil {
				headID = &hid
			}
		}

		_, err = s.db.Exec(reviveQuery, req.Description, parentID, headID, time.Now(), existingID)
		if err != nil {
			return nil, fmt.Errorf("failed to revive department: %w", err)
		}

		// Return the revived department
		parsedID, _ := uuid.Parse(existingID)
		return s.GetDepartmentByID(tenantID, parsedID)

	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing department: %w", err)
	}

	// Insert new department if not exists
	id := uuid.New()

	// Parse optional UUIDs
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		if pid, err := uuid.Parse(*req.ParentID); err == nil {
			parentID = &pid
		}
	}

	var headID *uuid.UUID
	if req.HeadID != nil && *req.HeadID != "" {
		if hid, err := uuid.Parse(*req.HeadID); err == nil {
			headID = &hid
		}
	}

	insertQuery := `
		INSERT INTO departments (id, tenant_id, name, description, parent_id, head_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, name, description, parent_id, head_id, is_active, created_at, updated_at`

	dept := &models.Department{}
	err = s.db.QueryRow(
		insertQuery,
		id,
		tenantID,
		req.Name,
		req.Description,
		parentID,
		headID,
		true, // Active by default
		time.Now(),
		time.Now(),
	).Scan(
		&dept.ID,
		&dept.TenantID,
		&dept.Name,
		&dept.Description,
		&dept.ParentID,
		&dept.HeadID,
		&dept.IsActive,
		&dept.CreatedAt,
		&dept.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return dept, nil
}

// UpdateDepartment updates an existing department
func (s *DepartmentService) UpdateDepartment(tenantID, id uuid.UUID, updates map[string]interface{}) (*models.Department, error) {
	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "name", "description":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "parent_id", "head_id":
			var uuidVal *uuid.UUID
			if strVal, ok := value.(string); ok && strVal != "" {
				if parsed, err := uuid.Parse(strVal); err == nil {
					uuidVal = &parsed
				}
			}
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, uuidVal)
			argIndex++
		case "is_active":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return s.GetDepartmentByID(tenantID, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id, tenantID)

	query := fmt.Sprintf("UPDATE departments SET %s WHERE id = $%d AND tenant_id = $%d",
		fmt.Sprintf("%s", setParts[0]), argIndex, argIndex+1) // Simplified join for brevity, real impl handles comma

	// Fix join logic
	query = fmt.Sprintf("UPDATE departments SET %s WHERE id = $%d AND tenant_id = $%d",
		func() string {
			res := setParts[0]
			for i := 1; i < len(setParts); i++ {
				res += ", " + setParts[i]
			}
			return res
		}(), argIndex, argIndex+1)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return s.GetDepartmentByID(tenantID, id)
}

// GetDepartmentByID retrieves a specific department
func (s *DepartmentService) GetDepartmentByID(tenantID, id uuid.UUID) (*models.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, head_id, parent_id, is_active, created_at, updated_at
		FROM departments 
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	dept := &models.Department{}
	err := s.db.QueryRow(query, id, tenantID).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return dept, nil
}

// GetDepartments retrieves all departments for a tenant
func (s *DepartmentService) GetDepartments(tenantID uuid.UUID) ([]*models.Department, error) {
	query := `
		SELECT id, tenant_id, name, description, head_id, parent_id, is_active, created_at, updated_at
		FROM departments 
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name`

	rows, err := s.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}
	defer rows.Close()

	defer rows.Close()

	departments := make([]*models.Department, 0)
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

// DeleteDepartment hard deletes a department
func (s *DepartmentService) DeleteDepartment(tenantID, id uuid.UUID) error {
	query := `DELETE FROM departments WHERE id = $1 AND tenant_id = $2`
	result, err := s.db.Exec(query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}
