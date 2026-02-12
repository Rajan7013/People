package models

import (
	"time"

	"github.com/google/uuid"
)

// BiometricDeviceType represents the type of biometric device
type BiometricDeviceType string

const (
	DeviceTypeFingerprint BiometricDeviceType = "fingerprint"
	DeviceTypeFaceID      BiometricDeviceType = "face_id"
	DeviceTypeIris        BiometricDeviceType = "iris"
	DeviceTypeVoice       BiometricDeviceType = "voice"
	DeviceTypeCard        BiometricDeviceType = "card"
	DeviceTypeHybrid      BiometricDeviceType = "hybrid"
)

// BiometricDeviceStatus represents the status of a biometric device
type BiometricDeviceStatus string

const (
	DeviceStatusActive   BiometricDeviceStatus = "active"
	DeviceStatusInactive BiometricDeviceStatus = "inactive"
	DeviceStatusOffline  BiometricDeviceStatus = "offline"
	DeviceStatusError    BiometricDeviceStatus = "error"
)

// BiometricDevice represents a biometric attendance device
type BiometricDevice struct {
	ID           uuid.UUID             `json:"id" db:"id"`
	TenantID     uuid.UUID             `json:"tenant_id" db:"tenant_id"`
	DeviceName   string                `json:"device_name" db:"device_name"`
	DeviceType   BiometricDeviceType   `json:"device_type" db:"device_type"`
	SerialNumber string                `json:"serial_number" db:"serial_number"`
	MacAddress   *string               `json:"mac_address,omitempty" db:"mac_address"`
	IPAddress    *string               `json:"ip_address,omitempty" db:"ip_address"`
	Port         *int                  `json:"port,omitempty" db:"port"`
	Location     *string               `json:"location,omitempty" db:"location"`
	Department   *string               `json:"department,omitempty" db:"department"`
	Status       BiometricDeviceStatus `json:"status" db:"status"`
	LastSyncAt   *time.Time            `json:"last_sync_at,omitempty" db:"last_sync_at"`
	IsActive     bool                  `json:"is_active" db:"is_active"`
	CreatedAt    time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at" db:"updated_at"`

	// Configuration settings
	Config BiometricDeviceConfig `json:"config,omitempty"`
}

// BiometricDeviceConfig represents device-specific configuration
type BiometricDeviceConfig struct {
	APIEndpoint    *string `json:"api_endpoint,omitempty"`
	APIKey         *string `json:"api_key,omitempty"`
	SyncInterval   int     `json:"sync_interval"` // in minutes
	TimeoutSeconds int     `json:"timeout_seconds"`
	Protocol       string  `json:"protocol"` // TCP, HTTP, MQTT, etc.
	Vendor         string  `json:"vendor"`   // ZKTeco, eSSL, etc.
	Model          string  `json:"model"`
}

// BiometricAttendanceLog represents raw attendance data from biometric devices
type BiometricAttendanceLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	DeviceID      uuid.UUID  `json:"device_id" db:"device_id"`
	EmployeeCode  string     `json:"employee_code" db:"employee_code"`
	EmployeeID    *uuid.UUID `json:"employee_id,omitempty" db:"employee_id"`
	Timestamp     time.Time  `json:"timestamp" db:"timestamp"`
	EventType     string     `json:"event_type" db:"event_type"`                   // "check_in", "check_out", "break_start", "break_end"
	BiometricData *string    `json:"biometric_data,omitempty" db:"biometric_data"` // encrypted biometric template
	Confidence    *float64   `json:"confidence,omitempty" db:"confidence"`         // matching confidence percentage
	IsProcessed   bool       `json:"is_processed" db:"is_processed"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	ErrorMessage  *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	DeviceName   string `json:"device_name,omitempty" db:"device_name"`
	EmployeeName string `json:"employee_name,omitempty" db:"employee_name"`
}

// AttendanceSource represents the source of attendance data
type AttendanceSource string

const (
	SourceManual    AttendanceSource = "manual"
	SourceBiometric AttendanceSource = "biometric"
	SourceMobile    AttendanceSource = "mobile"
	SourceWeb       AttendanceSource = "web"
	SourceAPI       AttendanceSource = "api"
)

// BiometricSyncRequest represents a request to sync data from a biometric device
type BiometricSyncRequest struct {
	DeviceID  uuid.UUID `json:"device_id" binding:"required"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Force     bool      `json:"force"` // Force sync even if recently synced
}

// BiometricSyncResponse represents the response from a sync operation
type BiometricSyncResponse struct {
	DeviceID       uuid.UUID `json:"device_id"`
	SyncStarted    time.Time `json:"sync_started"`
	SyncCompleted  time.Time `json:"sync_completed"`
	RecordsSynced  int       `json:"records_synced"`
	RecordsSkipped int       `json:"records_skipped"`
	Errors         []string  `json:"errors,omitempty"`
	Status         string    `json:"status"` // "success", "partial", "failed"
}

// DeviceRegistrationRequest represents a request to register a new biometric device
type DeviceRegistrationRequest struct {
	DeviceName   string                `json:"device_name" binding:"required"`
	DeviceType   BiometricDeviceType   `json:"device_type" binding:"required"`
	SerialNumber string                `json:"serial_number" binding:"required"`
	MacAddress   *string               `json:"mac_address,omitempty"`
	IPAddress    *string               `json:"ip_address,omitempty"`
	Port         *int                  `json:"port,omitempty"`
	Location     *string               `json:"location,omitempty"`
	Department   *string               `json:"department,omitempty"`
	Config       BiometricDeviceConfig `json:"config"`
}
