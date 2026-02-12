package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SystemManagementHandler handles system management HTTP requests
type SystemManagementHandler struct {
	systemService *services.SystemManagementService
}

// NewSystemManagementHandler creates a new system management handler
func NewSystemManagementHandler(systemService *services.SystemManagementService) *SystemManagementHandler {
	return &SystemManagementHandler{
		systemService: systemService,
	}
}

// Helper function to get tenant and user IDs from context
func (h *SystemManagementHandler) getContextInfo(r *http.Request) (*uuid.UUID, *uuid.UUID, string, error) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		return nil, nil, "", fmt.Errorf("unauthorized")
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid tenant ID")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid user ID")
	}

	return &tenantID, &userID, claims.Role, nil
}

// ===== SYSTEM SETTINGS ENDPOINTS =====

// GetSettings retrieves all system settings for a tenant
func (h *SystemManagementHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to view system settings
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	settings, err := h.systemService.GetSettings(*tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": settings,
	})
}

// UpdateSetting updates a specific system setting
func (h *SystemManagementHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to modify system settings
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	settingKey := chi.URLParam(r, "key")
	if settingKey == "" {
		http.Error(w, "Setting key is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.systemService.UpdateSetting(*tenantID, settingKey, req.Value, userID)
	if err != nil {
		if err.Error() == "setting not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Setting updated successfully",
	})
}

// CreateSetting creates a new system setting
func (h *SystemManagementHandler) CreateSetting(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to create system settings
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	var req struct {
		Key         string  `json:"key"`
		Value       string  `json:"value"`
		Type        string  `json:"type"`
		Description *string `json:"description"`
		IsSensitive bool    `json:"is_sensitive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value are required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "string"
	}

	err = h.systemService.CreateSetting(*tenantID, req.Key, req.Value, req.Type, req.Description, req.IsSensitive, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Setting created successfully",
	})
}

// ===== AUDIT LOGS ENDPOINTS =====

// GetAuditLogs retrieves audit logs for a tenant
func (h *SystemManagementHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin and hr roles to view audit logs
	if role != "admin" && role != "hr" {
		http.Error(w, "Access denied. Admin or HR role required.", http.StatusForbidden)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	logs, err := h.systemService.GetAuditLogs(*tenantID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// ===== SYSTEM METRICS ENDPOINTS =====

// GetMetrics retrieves system metrics
func (h *SystemManagementHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to view system metrics
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	metricName := r.URL.Query().Get("metric_name")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	var metricNamePtr *string
	if metricName != "" {
		metricNamePtr = &metricName
	}

	metrics, err := h.systemService.GetMetrics(*tenantID, metricNamePtr, nil, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metrics,
		"limit":   limit,
	})
}

// RecordMetric records a new system metric
func (h *SystemManagementHandler) RecordMetric(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to record metrics
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	var req struct {
		MetricName  string                 `json:"metric_name"`
		MetricValue *float64               `json:"metric_value"`
		MetricUnit  *string                `json:"metric_unit"`
		Tags        map[string]interface{} `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.MetricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	var tagsJSON json.RawMessage
	if req.Tags != nil {
		var err error
		tagsJSON, err = json.Marshal(req.Tags)
		if err != nil {
			http.Error(w, "Invalid tags format", http.StatusBadRequest)
			return
		}
	}

	err = h.systemService.RecordMetric(*tenantID, req.MetricName, req.MetricValue, req.MetricUnit, tagsJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Metric recorded successfully",
	})
}

// ===== SYSTEM BACKUPS ENDPOINTS =====

// CreateBackup initiates a new system backup
func (h *SystemManagementHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to create backups
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	var req struct {
		BackupType string `json:"backup_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BackupType == "" {
		req.BackupType = "full"
	}

	// Validate backup type
	validTypes := map[string]bool{"full": true, "incremental": true, "data_only": true}
	if !validTypes[req.BackupType] {
		http.Error(w, "Invalid backup type", http.StatusBadRequest)
		return
	}

	// Generate filename based on timestamp and type
	fileName := fmt.Sprintf("backup_%s_%s_%d.sql", req.BackupType, tenantID.String()[:8], time.Now().Unix())

	backup, err := h.systemService.CreateBackup(*tenantID, req.BackupType, fileName, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Implement actual backup logic here (async)
	// For now, just mark as completed
	go func() {
		// Simulate backup process
		time.Sleep(2 * time.Second)
		h.systemService.UpdateBackupStatus(backup.ID, "completed", nil, nil, nil)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Backup initiated successfully",
		"backup":  backup,
	})
}

// GetBackups retrieves backup records
func (h *SystemManagementHandler) GetBackups(w http.ResponseWriter, r *http.Request) {
	tenantID, _, role, err := h.getContextInfo(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only allow admin role to view backups
	if role != "admin" {
		http.Error(w, "Access denied. Admin role required.", http.StatusForbidden)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	backups, err := h.systemService.GetBackups(*tenantID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
		"limit":   limit,
		"offset":  offset,
	})
}
