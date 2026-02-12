package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"

	"github.com/google/uuid"
)

// UserSettingsHandler handles user settings HTTP requests
type UserSettingsHandler struct {
	userSettingsService *services.UserSettingsService
}

// NewUserSettingsHandler creates a new user settings handler
func NewUserSettingsHandler(userSettingsService *services.UserSettingsService) *UserSettingsHandler {
	return &UserSettingsHandler{
		userSettingsService: userSettingsService,
	}
}

// Helper function to get user info from context
func (h *UserSettingsHandler) getUserInfo(r *http.Request) (*uuid.UUID, *uuid.UUID, string, error) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		return nil, nil, "", fmt.Errorf("unauthorized")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid user ID")
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid tenant ID")
	}

	return &userID, &tenantID, claims.Role, nil
}

// ===== USER PROFILE ENDPOINTS =====

// GetUserProfile retrieves user profile information
func (h *UserSettingsHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.userSettingsService.GetUserProfile(*userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profile": profile,
	})
}

// UpdateUserProfile updates user profile information
func (h *UserSettingsHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Phone                 *string `json:"phone"`
		Address               *string `json:"address"`
		DateOfBirth           *string `json:"date_of_birth"`
		EmergencyContactName  *string `json:"emergency_contact_name"`
		EmergencyContactPhone *string `json:"emergency_contact_phone"`
		Bio                   *string `json:"bio"`
		JobTitle              *string `json:"job_title"`
		DepartmentName        *string `json:"department_name"`
		ManagerName           *string `json:"manager_name"`
		HireDate              *string `json:"hire_date"`
		WorkLocation          *string `json:"work_location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	profile := services.UserProfile{
		UserID:                *userID,
		TenantID:              *tenantID,
		Phone:                 req.Phone,
		Address:               req.Address,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Bio:                   req.Bio,
		JobTitle:              req.JobTitle,
		DepartmentName:        req.DepartmentName,
		ManagerName:           req.ManagerName,
		WorkLocation:          req.WorkLocation,
	}

	// Parse dates if provided
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		if dob, err := time.Parse("2006-01-02", *req.DateOfBirth); err == nil {
			profile.DateOfBirth = &dob
		}
	}

	if req.HireDate != nil && *req.HireDate != "" {
		if hireDate, err := time.Parse("2006-01-02", *req.HireDate); err == nil {
			profile.HireDate = &hireDate
		}
	}

	err = h.userSettingsService.UpdateUserProfile(*userID, *tenantID, profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}

// ===== USER PREFERENCES ENDPOINTS =====

// GetUserPreferences retrieves user preferences
func (h *UserSettingsHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	preferences, err := h.userSettingsService.GetUserPreferences(*userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preferences": preferences,
	})
}

// UpdateUserPreferences updates multiple user preferences
func (h *UserSettingsHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update each preference
	for key, value := range req {
		var valueType string
		switch value.(type) {
		case bool:
			valueType = "boolean"
		case int, int64, float64:
			valueType = "integer"
		case map[string]interface{}, []interface{}:
			valueType = "json"
		default:
			valueType = "string"
		}

		err := h.userSettingsService.UpdateUserPreference(*userID, *tenantID, key, value, valueType)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update preference %s: %v", key, err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Preferences updated successfully",
	})
}

// ===== SECURITY SETTINGS ENDPOINTS =====

// GetSecuritySettings retrieves user security settings
func (h *UserSettingsHandler) GetSecuritySettings(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	settings, err := h.userSettingsService.GetSecuritySettings(*userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"security_settings": settings,
	})
}

// UpdateSecuritySettings updates user security settings
func (h *UserSettingsHandler) UpdateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var settings services.SecuritySettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	settings.UserID = *userID

	err = h.userSettingsService.UpdateSecuritySettings(*userID, settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Security settings updated successfully",
	})
}

// ===== THEME SETTINGS ENDPOINTS =====

// GetUserTheme retrieves user theme settings
func (h *UserSettingsHandler) GetUserTheme(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	theme, err := h.userSettingsService.GetUserTheme(*userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"theme": theme,
	})
}

// UpdateUserTheme updates user theme settings
func (h *UserSettingsHandler) UpdateUserTheme(w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := h.getUserInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var theme services.UserTheme
	if err := json.NewDecoder(r.Body).Decode(&theme); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	theme.UserID = *userID

	err = h.userSettingsService.UpdateUserTheme(*userID, theme)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Theme settings updated successfully",
	})
}
