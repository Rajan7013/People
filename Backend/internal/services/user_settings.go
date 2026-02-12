package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserProfile represents a user's profile information
type UserProfile struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                uuid.UUID  `json:"user_id" db:"user_id"`
	TenantID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Phone                 *string    `json:"phone" db:"phone"`
	Address               *string    `json:"address" db:"address"`
	DateOfBirth           *time.Time `json:"date_of_birth" db:"date_of_birth"`
	EmergencyContactName  *string    `json:"emergency_contact_name" db:"emergency_contact_name"`
	EmergencyContactPhone *string    `json:"emergency_contact_phone" db:"emergency_contact_phone"`
	ProfilePictureURL     *string    `json:"profile_picture_url" db:"profile_picture_url"`
	Bio                   *string    `json:"bio" db:"bio"`
	JobTitle              *string    `json:"job_title" db:"job_title"`
	DepartmentName        *string    `json:"department_name" db:"department_name"`
	ManagerName           *string    `json:"manager_name" db:"manager_name"`
	HireDate              *time.Time `json:"hire_date" db:"hire_date"`
	WorkLocation          *string    `json:"work_location" db:"work_location"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// UserPreference represents a user preference setting
type UserPreference struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	TenantID        uuid.UUID `json:"tenant_id" db:"tenant_id"`
	PreferenceKey   string    `json:"preference_key" db:"preference_key"`
	PreferenceValue *string   `json:"preference_value" db:"preference_value"`
	PreferenceType  string    `json:"preference_type" db:"preference_type"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// SecuritySettings represents user security preferences
type SecuritySettings struct {
	ID                 uuid.UUID      `json:"id" db:"id"`
	UserID             uuid.UUID      `json:"user_id" db:"user_id"`
	TenantID           uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	TwoFactorEnabled   bool           `json:"two_factor_enabled" db:"two_factor_enabled"`
	SessionTimeout     int            `json:"session_timeout" db:"session_timeout"`
	LoginNotifications bool           `json:"login_notifications" db:"login_notifications"`
	DeviceTracking     bool           `json:"device_tracking" db:"device_tracking"`
	IPRestrictions     pq.StringArray `json:"ip_restrictions" db:"ip_restrictions"`
	LastPasswordChange *time.Time     `json:"last_password_change" db:"last_password_change"`
	PasswordExpiryDays int            `json:"password_expiry_days" db:"password_expiry_days"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// UserTheme represents user UI customization preferences
type UserTheme struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	TenantID         uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ThemeName        string    `json:"theme_name" db:"theme_name"`
	PrimaryColor     string    `json:"primary_color" db:"primary_color"`
	SecondaryColor   string    `json:"secondary_color" db:"secondary_color"`
	FontSize         string    `json:"font_size" db:"font_size"`
	CompactMode      bool      `json:"compact_mode" db:"compact_mode"`
	SidebarCollapsed bool      `json:"sidebar_collapsed" db:"sidebar_collapsed"`
	Language         string    `json:"language" db:"language"`
	Timezone         string    `json:"timezone" db:"timezone"`
	DateFormat       string    `json:"date_format" db:"date_format"`
	TimeFormat       string    `json:"time_format" db:"time_format"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UserSettingsService handles user profile and settings operations
type UserSettingsService struct {
	db *sql.DB
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(db *sql.DB) *UserSettingsService {
	return &UserSettingsService{db: db}
}

// ===== USER PROFILE METHODS =====

// GetUserProfile retrieves user profile information
func (s *UserSettingsService) GetUserProfile(userID uuid.UUID) (*UserProfile, error) {
	query := `
		SELECT id, user_id, tenant_id, phone, address, date_of_birth, 
		       emergency_contact_name, emergency_contact_phone, profile_picture_url,
		       bio, job_title, department_name, manager_name, hire_date, 
		       work_location, created_at, updated_at
		FROM user_profiles 
		WHERE user_id = $1`

	var profile UserProfile
	err := s.db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.TenantID, &profile.Phone, &profile.Address,
		&profile.DateOfBirth, &profile.EmergencyContactName, &profile.EmergencyContactPhone,
		&profile.ProfilePictureURL, &profile.Bio, &profile.JobTitle, &profile.DepartmentName,
		&profile.ManagerName, &profile.HireDate, &profile.WorkLocation,
		&profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Profile doesn't exist yet
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &profile, nil
}

// UpdateUserProfile creates or updates user profile information
func (s *UserSettingsService) UpdateUserProfile(userID, tenantID uuid.UUID, profile UserProfile) error {
	query := `
		INSERT INTO user_profiles (
			user_id, tenant_id, phone, address, date_of_birth, emergency_contact_name,
			emergency_contact_phone, profile_picture_url, bio, job_title, department_name,
			manager_name, hire_date, work_location, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			date_of_birth = EXCLUDED.date_of_birth,
			emergency_contact_name = EXCLUDED.emergency_contact_name,
			emergency_contact_phone = EXCLUDED.emergency_contact_phone,
			profile_picture_url = EXCLUDED.profile_picture_url,
			bio = EXCLUDED.bio,
			job_title = EXCLUDED.job_title,
			department_name = EXCLUDED.department_name,
			manager_name = EXCLUDED.manager_name,
			hire_date = EXCLUDED.hire_date,
			work_location = EXCLUDED.work_location,
			updated_at = CURRENT_TIMESTAMP`

	_, err := s.db.Exec(query, userID, tenantID, profile.Phone, profile.Address, profile.DateOfBirth,
		profile.EmergencyContactName, profile.EmergencyContactPhone, profile.ProfilePictureURL,
		profile.Bio, profile.JobTitle, profile.DepartmentName, profile.ManagerName,
		profile.HireDate, profile.WorkLocation)

	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

// ===== USER PREFERENCES METHODS =====

// GetUserPreferences retrieves all user preferences
func (s *UserSettingsService) GetUserPreferences(userID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT preference_key, preference_value, preference_type 
		FROM user_preferences 
		WHERE user_id = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	defer rows.Close()

	preferences := make(map[string]interface{})
	for rows.Next() {
		var key, valueStr, valueType string
		err := rows.Scan(&key, &valueStr, &valueType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan preference: %w", err)
		}

		// Convert value based on type
		switch valueType {
		case "boolean":
			preferences[key] = valueStr == "true"
		case "integer":
			var intVal int
			if err := json.Unmarshal([]byte(valueStr), &intVal); err == nil {
				preferences[key] = intVal
			} else {
				preferences[key] = valueStr
			}
		case "json":
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(valueStr), &jsonVal); err == nil {
				preferences[key] = jsonVal
			} else {
				preferences[key] = valueStr
			}
		default:
			preferences[key] = valueStr
		}
	}

	return preferences, nil
}

// UpdateUserPreference creates or updates a user preference
func (s *UserSettingsService) UpdateUserPreference(userID, tenantID uuid.UUID, key string, value interface{}, valueType string) error {
	var valueStr string
	var err error

	// Convert value to string based on type
	switch valueType {
	case "boolean":
		if boolVal, ok := value.(bool); ok {
			valueStr = fmt.Sprintf("%t", boolVal)
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}
	case "integer":
		valueStr = fmt.Sprintf("%v", value)
	case "json":
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON value: %w", err)
		}
		valueStr = string(bytes)
	default:
		valueStr = fmt.Sprintf("%v", value)
	}

	query := `
		INSERT INTO user_preferences (user_id, tenant_id, preference_key, preference_value, preference_type)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, preference_key) 
		DO UPDATE SET 
			preference_value = EXCLUDED.preference_value,
			preference_type = EXCLUDED.preference_type,
			updated_at = CURRENT_TIMESTAMP`

	_, err = s.db.Exec(query, userID, tenantID, key, valueStr, valueType)
	if err != nil {
		return fmt.Errorf("failed to update user preference: %w", err)
	}

	return nil
}

// ===== SECURITY SETTINGS METHODS =====

// GetSecuritySettings retrieves user security settings
func (s *UserSettingsService) GetSecuritySettings(userID uuid.UUID) (*SecuritySettings, error) {
	query := `
		SELECT id, user_id, tenant_id, two_factor_enabled, session_timeout, 
		       login_notifications, device_tracking, ip_restrictions, 
		       last_password_change, password_expiry_days, created_at, updated_at
		FROM security_settings 
		WHERE user_id = $1`

	var settings SecuritySettings
	err := s.db.QueryRow(query, userID).Scan(
		&settings.ID, &settings.UserID, &settings.TenantID, &settings.TwoFactorEnabled,
		&settings.SessionTimeout, &settings.LoginNotifications, &settings.DeviceTracking,
		&settings.IPRestrictions, &settings.LastPasswordChange, &settings.PasswordExpiryDays,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get security settings: %w", err)
	}

	return &settings, nil
}

// UpdateSecuritySettings updates user security settings
func (s *UserSettingsService) UpdateSecuritySettings(userID uuid.UUID, settings SecuritySettings) error {
	query := `
		UPDATE security_settings 
		SET two_factor_enabled = $1, session_timeout = $2, login_notifications = $3,
		    device_tracking = $4, ip_restrictions = $5, password_expiry_days = $6,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $7`

	_, err := s.db.Exec(query, settings.TwoFactorEnabled, settings.SessionTimeout,
		settings.LoginNotifications, settings.DeviceTracking, settings.IPRestrictions,
		settings.PasswordExpiryDays, userID)

	if err != nil {
		return fmt.Errorf("failed to update security settings: %w", err)
	}

	return nil
}

// ===== THEME SETTINGS METHODS =====

// GetUserTheme retrieves user theme settings
func (s *UserSettingsService) GetUserTheme(userID uuid.UUID) (*UserTheme, error) {
	query := `
		SELECT id, user_id, tenant_id, theme_name, primary_color, secondary_color,
		       font_size, compact_mode, sidebar_collapsed, language, timezone,
		       date_format, time_format, created_at, updated_at
		FROM user_themes 
		WHERE user_id = $1`

	var theme UserTheme
	err := s.db.QueryRow(query, userID).Scan(
		&theme.ID, &theme.UserID, &theme.TenantID, &theme.ThemeName,
		&theme.PrimaryColor, &theme.SecondaryColor, &theme.FontSize,
		&theme.CompactMode, &theme.SidebarCollapsed, &theme.Language,
		&theme.Timezone, &theme.DateFormat, &theme.TimeFormat,
		&theme.CreatedAt, &theme.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user theme: %w", err)
	}

	return &theme, nil
}

// UpdateUserTheme updates user theme settings
func (s *UserSettingsService) UpdateUserTheme(userID uuid.UUID, theme UserTheme) error {
	query := `
		UPDATE user_themes 
		SET theme_name = $1, primary_color = $2, secondary_color = $3, font_size = $4,
		    compact_mode = $5, sidebar_collapsed = $6, language = $7, timezone = $8,
		    date_format = $9, time_format = $10, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $11`

	_, err := s.db.Exec(query, theme.ThemeName, theme.PrimaryColor, theme.SecondaryColor,
		theme.FontSize, theme.CompactMode, theme.SidebarCollapsed, theme.Language,
		theme.Timezone, theme.DateFormat, theme.TimeFormat, userID)

	if err != nil {
		return fmt.Errorf("failed to update user theme: %w", err)
	}

	return nil
}
