package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type Claims struct {
	UserID       string `json:"user_id"`
	TenantID     string `json:"tenant_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id,omitempty"`
	TeamID       string `json:"team_id,omitempty"`
	jwt.RegisteredClaims
}

type User struct {
	ID              string     `json:"id"`
	TenantID        *string    `json:"tenant_id,omitempty"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`
	Role            string     `json:"role"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	DepartmentID    *string    `json:"department_id,omitempty"`
	TeamID          *string    `json:"team_id,omitempty"`
	IsActive        bool       `json:"is_active"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

type Service struct {
	db              *sql.DB
	jwtSecret       []byte
	pepperSecret    string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(db *sql.DB, jwtSecret, pepperSecret string, accessTokenTTL, refreshTokenTTL int) *Service {
	return &Service{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		pepperSecret:    pepperSecret,
		accessTokenTTL:  time.Duration(accessTokenTTL) * time.Minute,
		refreshTokenTTL: time.Duration(refreshTokenTTL) * time.Minute,
	}
}

// Login authenticates a user and returns JWT token
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	// Verify password (Argon2id + Pepper)
	valid, err := VerifyPassword(req.Password, user.PasswordHash, s.pepperSecret)
	if err != nil {
		// Log error (could be invalid hash format from legacy bcrypt)
		return nil, ErrInvalidCredentials
	}
	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (simplified - in production, store in Redis)
	refreshToken, _, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// Update last login time
	if err := s.updateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         *user,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user to ensure they're still active
	user, err := s.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new access token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, _, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         *user,
	}, nil
}

// HashPassword hashes a password using Argon2id and a Pepper
func (s *Service) HashPassword(password string) (string, error) {
	return HashPassword(password, s.pepperSecret)
}

// generateToken creates a JWT token for the user
func (s *Service) generateToken(user *User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenTTL)

	// Helper to safely dereference pointers
	tenantID := ""
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	departmentID := ""
	if user.DepartmentID != nil {
		departmentID = *user.DepartmentID
	}
	teamID := ""
	if user.TeamID != nil {
		teamID = *user.TeamID
	}

	claims := &Claims{
		UserID:       user.ID,
		TenantID:     tenantID,
		Email:        user.Email,
		Role:         user.Role,
		DepartmentID: departmentID,
		TeamID:       teamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "peopleos-api",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a refresh token (longer TTL)
func (s *Service) GenerateRefreshToken(user *User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.refreshTokenTTL)

	// Helper to safely dereference pointers
	tenantID := ""
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	departmentID := ""
	if user.DepartmentID != nil {
		departmentID = *user.DepartmentID
	}
	teamID := ""
	if user.TeamID != nil {
		teamID = *user.TeamID
	}

	claims := &Claims{
		UserID:       user.ID,
		TenantID:     tenantID,
		Email:        user.Email,
		Role:         user.Role,
		DepartmentID: departmentID,
		TeamID:       teamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "peopleos-api-refresh",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// getUserByEmail retrieves a user by email
func (s *Service) getUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.role, u.first_name, u.last_name, 
			   u.is_active, u.last_login_at, u.email_verified_at, u.created_at, u.updated_at,
			   e.department_id, u.team_id
		FROM users u
		LEFT JOIN employees e ON u.id = e.user_id
		WHERE u.email = $1 AND u.deleted_at IS NULL
	`

	log.Info().
		Str("email", email).
		Str("query", query).
		Msg("Executing getUserByEmail query")

	user := &User{}
	var tenantID sql.NullString
	var passwordHash sql.NullString
	var departmentID sql.NullString
	var teamID sql.NullString

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &tenantID, &user.Email, &passwordHash,
		&user.Role, &user.FirstName, &user.LastName, &user.IsActive,
		&user.LastLoginAt, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
		&departmentID, &teamID,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("email", email).
			Msg("getUserByEmail failed")
		return nil, err
	}

	// Set tenant_id if not NULL (super_admin has NULL tenant_id)
	if tenantID.Valid {
		user.TenantID = &tenantID.String
	}

	// Set password hash only if not NULL
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}

	// Set department_id if not NULL
	if departmentID.Valid {
		user.DepartmentID = &departmentID.String
	}

	// Set team_id if not NULL
	if teamID.Valid {
		user.TeamID = &teamID.String
	}

	log.Info().
		Str("email", email).
		Str("user_id", user.ID).
		Msg("getUserByEmail succeeded")

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.role, u.first_name, u.last_name, 
			   u.is_active, u.last_login_at, u.email_verified_at, u.created_at, u.updated_at,
			   e.department_id, u.team_id
		FROM users u
		LEFT JOIN employees e ON u.id = e.user_id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	user := &User{}
	var tenantID sql.NullString
	var passwordHash sql.NullString
	var departmentID sql.NullString
	var teamID sql.NullString

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &tenantID, &user.Email, &passwordHash,
		&user.Role, &user.FirstName, &user.LastName, &user.IsActive,
		&user.LastLoginAt, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
		&departmentID, &teamID,
	)

	if err != nil {
		return nil, err
	}

	// Set tenant_id if not NULL
	if tenantID.Valid {
		user.TenantID = &tenantID.String
	}

	// Set password hash only if not NULL
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}

	// Set department_id if not NULL
	if departmentID.Valid {
		user.DepartmentID = &departmentID.String
	}

	// Set team_id if not NULL
	if teamID.Valid {
		user.TeamID = &teamID.String
	}

	return user, nil
}

// updateLastLogin updates the user's last login timestamp
func (s *Service) updateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

// FindOrCreateGoogleUser finds an existing user by google_id or creates a new one
func (s *Service) FindOrCreateGoogleUser(ctx context.Context, googleUser *GoogleUserInfo) (*User, error) {
	// Ensure email is lowercase for consistent matching
	googleUser.Email = strings.ToLower(googleUser.Email)

	log.Info().
		Str("email", googleUser.Email).
		Str("google_id", googleUser.ID).
		Msg("Starting FindOrCreateGoogleUser")

	// Try to find existing user by google_id
	user, err := s.getUserByGoogleID(ctx, googleUser.ID)
	if err == nil {
		// User exists with google_id, update last login and return
		log.Info().Msg("Found user by google_id")
		_ = s.updateLastLogin(ctx, user.ID)
		return user, nil
	}
	log.Info().Err(err).Msg("User not found by google_id, checking by email")

	// Check if user exists by email (from old email/password login)
	existingUser, err := s.getUserByEmail(ctx, googleUser.Email)
	if err == nil {
		// User exists with this email, link their Google account
		log.Info().
			Str("email", googleUser.Email).
			Str("google_id", googleUser.ID).
			Str("user_id", existingUser.ID).
			Msg("Linking existing user account with Google")

		updateQuery := `
			UPDATE users 
			SET google_id = $1, 
			    password_hash = NULL,
			    email_verified_at = COALESCE(email_verified_at, NOW()),
			    last_login_at = NOW()
			WHERE id = $2
			RETURNING id, tenant_id, email, role, first_name, last_name, is_active, last_login_at, email_verified_at, created_at, updated_at
		`

		updatedUser := &User{}
		err = s.db.QueryRowContext(ctx, updateQuery, googleUser.ID, existingUser.ID).Scan(
			&updatedUser.ID, &updatedUser.TenantID, &updatedUser.Email, &updatedUser.Role,
			&updatedUser.FirstName, &updatedUser.LastName, &updatedUser.IsActive,
			&updatedUser.LastLoginAt, &updatedUser.EmailVerifiedAt, &updatedUser.CreatedAt, &updatedUser.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to update user with google_id")
			return nil, fmt.Errorf("failed to link Google account: %w", err)
		}

		log.Info().Msg("Successfully linked Google account")
		return updatedUser, nil
	}
	log.Info().Err(err).Msg("User not found by email either, creating new user")

	// User doesn't exist at all, create new user
	// Get default tenant (PeopleOS HQ)
	var defaultTenantID string
	err = s.db.QueryRowContext(ctx, `
		SELECT id FROM tenants WHERE name = 'PeopleOS HQ' LIMIT 1
	`).Scan(&defaultTenantID)

	if err != nil {
		// If default tenant not found, we cannot auto-create user.
		// Return UserNotFound so frontend can show "Email not registered"
		return nil, ErrUserNotFound
	}

	// Create new user
	now := time.Now()
	query := `
		INSERT INTO users (tenant_id, email, google_id, first_name, last_name, role, is_active, email_verified_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, 'employee', true, $6, $7)
		RETURNING id, tenant_id, email, role, first_name, last_name, is_active, last_login_at, email_verified_at, created_at, updated_at
	`

	newUser := &User{}
	err = s.db.QueryRowContext(ctx, query,
		defaultTenantID, googleUser.Email, googleUser.ID,
		googleUser.GivenName, googleUser.FamilyName, now, now,
	).Scan(
		&newUser.ID, &newUser.TenantID, &newUser.Email, &newUser.Role,
		&newUser.FirstName, &newUser.LastName, &newUser.IsActive,
		&newUser.LastLoginAt, &newUser.EmailVerifiedAt, &newUser.CreatedAt, &newUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// getUserByGoogleID retrieves a user by Google ID
func (s *Service) getUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, first_name, last_name, 
			   is_active, last_login_at, email_verified_at, created_at, updated_at
		FROM users 
		WHERE google_id = $1 AND deleted_at IS NULL
	`

	user := &User{}
	var passwordHash sql.NullString

	err := s.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID, &user.TenantID, &user.Email, &passwordHash,
		&user.Role, &user.FirstName, &user.LastName, &user.IsActive,
		&user.LastLoginAt, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set password hash only if not NULL
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}

	return user, nil
}

// CheckUserStatus verifies if the user and their tenant are active and not deleted
func (s *Service) CheckUserStatus(ctx context.Context, userID string) error {
	// 1. Check User Status
	query := `
		SELECT u.is_active, u.deleted_at, u.tenant_id
		FROM users u
		WHERE u.id = $1
	`
	var isActive bool
	var deletedAt sql.NullTime
	var tenantID sql.NullString

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&isActive, &deletedAt, &tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	if deletedAt.Valid {
		return errors.New("user account has been deleted")
	}

	if !isActive {
		return ErrUserInactive
	}

	// 2. Check Tenant Status (if user belongs to a tenant)
	if tenantID.Valid {
		tenantQuery := `
			SELECT status, deleted_at
			FROM tenants
			WHERE id = $1
		`
		var status string
		var tenantDeletedAt sql.NullTime

		err := s.db.QueryRowContext(ctx, tenantQuery, tenantID.String).Scan(&status, &tenantDeletedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("organization not found")
			}
			return err
		}

		if tenantDeletedAt.Valid {
			return errors.New("organization has been deleted")
		}

		if status != "active" {
			return errors.New("organization is suspended")
		}
	}

	return nil
}

// GenerateAccessToken generates a JWT access token for a user
func (s *Service) GenerateAccessToken(userID, tenantID, email, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
