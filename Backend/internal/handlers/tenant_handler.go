package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
	"github.com/rs/zerolog/log"
)

type TenantHandler struct {
	organizationService *services.OrganizationService
}

func NewTenantHandler(organizationService *services.OrganizationService) *TenantHandler {
	return &TenantHandler{
		organizationService: organizationService,
	}
}

// GetConfig retrieves the configuration/details of the current authenticated tenant
func (h *TenantHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	tenant, err := h.organizationService.GetOrganizationByID(r.Context(), tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", claims.TenantID).Msg("Failed to get tenant config")
		http.Error(w, "Failed to retrieve configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// UpdateConfig updates the configuration/details of the current authenticated tenant
func (h *TenantHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize updates: Prevent restricted fields from being updated by Org Admin
	// Super Admin might be able to update plan/subscription, but Org Admin shouldn't via this endpoint
	delete(updates, "status")
	delete(updates, "subscription")
	delete(updates, "plan_id")
	delete(updates, "storage_used_mb")
	delete(updates, "api_requests_count")

	updatedTenant, err := h.organizationService.UpdateOrganization(r.Context(), tenantID, updates)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", claims.TenantID).Msg("Failed to update tenant config")
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTenant)
}
