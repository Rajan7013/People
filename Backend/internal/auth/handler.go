package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/logout", h.Logout)

	// Google OAuth routes
	r.Get("/google", h.GoogleLogin)
	r.Get("/google/callback", h.GoogleCallback)
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Email is required")
		return
	}
	if req.Password == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Authenticate user
	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("Login failed")

		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		} else if errors.Is(err, ErrUserInactive) {
			h.writeErrorResponse(w, http.StatusForbidden, "User account is inactive")
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Set refresh token as HttpOnly cookie
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   true, // Always true for SameSite=None
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}
	http.SetCookie(w, cookie)

	tenantIDStr := ""
	if response.User.TenantID != nil {
		tenantIDStr = *response.User.TenantID
	}

	log.Info().
		Str("user_id", response.User.ID).
		Str("tenant_id", tenantIDStr).
		Str("email", response.User.Email).
		Str("role", response.User.Role).
		Msg("User logged in successfully")

	// Return response without refresh token in body
	// Create a minimal response structure
	loginResp := map[string]interface{}{
		"token":      response.Token,
		"expires_at": response.ExpiresAt,
		"user":       response.User,
	}

	h.writeJSONResponse(w, http.StatusOK, loginResp)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Try to get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			h.writeErrorResponse(w, http.StatusUnauthorized, "Refresh token required")
			return
		}
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid cookie")
		return
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Refresh token required")
		return
	}

	// Refresh token
	response, err := h.service.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Token refresh failed")

		// Clear cookie if invalid
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/api/v1/auth/refresh",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})

		switch err {
		case ErrInvalidToken, ErrTokenExpired:
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		case ErrUserInactive:
			h.writeErrorResponse(w, http.StatusForbidden, "User account is inactive")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Set new refresh token as HttpOnly cookie (rotation)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	tenantIDStr := ""
	if response.User.TenantID != nil {
		tenantIDStr = *response.User.TenantID
	}

	log.Info().
		Str("user_id", response.User.ID).
		Str("tenant_id", tenantIDStr).
		Msg("Token refreshed successfully")

	// Return response without refresh token in body
	refreshResp := map[string]interface{}{
		"token":      response.Token,
		"expires_at": response.ExpiresAt,
		"user":       response.User,
	}

	h.writeJSONResponse(w, http.StatusOK, refreshResp)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// Profile returns the current user's profile
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetClaimsFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Missing authentication")
		return
	}

	// Fetch full user profile
	user, err := h.service.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user profile")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, user)
}

// writeJSONResponse writes a JSON response
func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// writeErrorResponse writes a JSON error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]string{
		"error": message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}

// GoogleLogin initiates the Google OAuth flow
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate random state for CSRF protection
	state := generateRandomState()

	// Store state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Get the OAuth URL and redirect
	url := GoogleOAuthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state to prevent CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		log.Error().Err(err).Msg("Invalid OAuth state")
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid OAuth state")
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := ExchangeCodeForToken(code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange code for token")
		h.writeErrorResponse(w, http.StatusUnauthorized, "Failed to exchange authorization code")
		return
	}

	// Get user info from Google
	userInfo, err := GetGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info from Google")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info from Google")
		return
	}

	// Check if user exists or create new user
	user, err := h.service.FindOrCreateGoogleUser(r.Context(), userInfo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find or create user")
		if errors.Is(err, ErrUserInactive) {
			h.writeErrorResponse(w, http.StatusForbidden, "User account is inactive")
		} else if errors.Is(err, ErrUserNotFound) {
			h.writeErrorResponse(w, http.StatusNotFound, "Email is not registered")
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to process user login")
		}
		return
	}

	// Generate JWT token
	tenantIDStr := ""
	if user.TenantID != nil {
		tenantIDStr = *user.TenantID
	}
	jwtToken, err := h.service.GenerateAccessToken(user.ID, tenantIDStr, user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	// Generate refresh token
	refreshToken, _, err := h.service.GenerateRefreshToken(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	log.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("JWT and Refresh tokens generated successfully")

	// Return JSON response with token and user data
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"token": jwtToken,
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"role":      user.Role,
			"tenant_id": user.TenantID,
		},
	})
}

// generateRandomState generates a random state string for CSRF protection
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
