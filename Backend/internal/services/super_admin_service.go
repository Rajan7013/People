package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/models"
	"github.com/rs/zerolog/log"
)

type SuperAdminService struct {
	db           *sql.DB
	pepperSecret string
}

func NewSuperAdminService(db *sql.DB, pepperSecret string) *SuperAdminService {
	return &SuperAdminService{
		db:           db,
		pepperSecret: pepperSecret,
	}
}

// GenerateSecurePassword generates a cryptographically secure random password
// Format: 12 characters with uppercase, lowercase, numbers, and symbols
func (s *SuperAdminService) GenerateSecurePassword() (string, error) {
	const (
		length    = 12
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		numbers   = "0123456789"
		symbols   = "!@#$%^&*"
		allChars  = uppercase + lowercase + numbers + symbols
	)

	password := make([]byte, length)

	// Ensure at least one character from each category
	categories := []string{uppercase, lowercase, numbers, symbols}
	for i, category := range categories {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(category))))
		if err != nil {
			return "", err
		}
		password[i] = category[n.Int64()]
	}

	// Fill remaining characters randomly
	for i := len(categories); i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}
		password[i] = allChars[n.Int64()]
	}

	// Shuffle the password to avoid predictable pattern
	for i := length - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}

// CreateSuperAdmin creates a new super admin user
func (s *SuperAdminService) CreateSuperAdmin(ctx context.Context, req models.CreateSuperAdminRequest) (*models.SuperAdminResponse, error) {
	// Validate email format
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	// Check if email already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)
	`, req.Email).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to check email existence")
		return nil, errors.New("failed to validate email")
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Generate secure password
	tempPassword, err := s.GenerateSecurePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate secure password")
		return nil, errors.New("failed to generate password")
	}

	// Hash password with Argon2id + pepper
	passwordHash, err := auth.HashPassword(tempPassword, s.pepperSecret)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, errors.New("failed to hash password")
	}

	// Create super admin user
	userID := uuid.New().String()
	var createdAt, updatedAt string

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (
			id, 
			email, 
			password_hash, 
			role, 
			first_name, 
			last_name, 
			tenant_id,
			is_active
		) VALUES ($1, $2, $3, $4, $5, $6, NULL, true)
		RETURNING created_at, updated_at
	`, userID, req.Email, passwordHash, "super_admin", req.FirstName, req.LastName).Scan(&createdAt, &updatedAt)

	if err != nil {
		log.Error().Err(err).
			Str("email", req.Email).
			Msg("Failed to create super admin")
		return nil, errors.New("failed to create super admin")
	}

	log.Info().
		Str("user_id", userID).
		Str("email", req.Email).
		Msg("Super admin created successfully")

	// Parse timestamps
	createdAtTime, _ := parseTimestamp(createdAt)
	updatedAtTime, _ := parseTimestamp(updatedAt)

	return &models.SuperAdminResponse{
		ID:                userID,
		Email:             req.Email,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Role:              "super_admin",
		TemporaryPassword: tempPassword, // Return temp password only on creation
		IsActive:          true,
		CreatedAt:         createdAtTime,
		UpdatedAt:         updatedAtTime,
	}, nil
}

// GetAllSuperAdmins retrieves all super admin users
func (s *SuperAdminService) GetAllSuperAdmins(ctx context.Context) ([]models.SuperAdminListItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			id, 
			email, 
			first_name, 
			last_name, 
			is_active, 
			created_at,
			last_login_at
		FROM users
		WHERE role = 'super_admin' AND deleted_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch super admins")
		return nil, errors.New("failed to fetch super admins")
	}
	defer rows.Close()

	var admins []models.SuperAdminListItem
	for rows.Next() {
		var admin models.SuperAdminListItem
		var createdAt string
		var lastLoginAt sql.NullString

		err := rows.Scan(
			&admin.ID,
			&admin.Email,
			&admin.FirstName,
			&admin.LastName,
			&admin.IsActive,
			&createdAt,
			&lastLoginAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan super admin row")
			continue
		}

		admin.CreatedAt, _ = parseTimestamp(createdAt)
		if lastLoginAt.Valid {
			t, _ := parseTimestamp(lastLoginAt.String)
			admin.LastLoginAt = &t
		}

		admins = append(admins, admin)
	}

	return admins, nil
}

// Helper function to parse timestamp strings
func parseTimestamp(ts string) (time.Time, error) {
	// Try multiple formats
	formats := []string{
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05-07",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", ts)
}
