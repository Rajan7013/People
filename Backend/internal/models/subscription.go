package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionPlan represents a pricing tier/plan
type SubscriptionPlan struct {
	ID                    uuid.UUID              `json:"id" db:"id"`
	Name                  string                 `json:"name" db:"name"`
	DisplayName           string                 `json:"display_name" db:"display_name"`
	Description           *string                `json:"description,omitempty" db:"description"`
	PriceMonthly          float64                `json:"price_monthly" db:"price_monthly"`
	PriceYearly           float64                `json:"price_yearly" db:"price_yearly"`
	Currency              string                 `json:"currency" db:"currency"`
	MaxUsers              *int                   `json:"max_users,omitempty" db:"max_users"`
	MaxStorageGB          *int                   `json:"max_storage_gb,omitempty" db:"max_storage_gb"`
	MaxAPIRequestsMonthly *int                   `json:"max_api_requests_monthly,omitempty" db:"max_api_requests_monthly"`
	MaxDepartments        *int                   `json:"max_departments,omitempty" db:"max_departments"`
	Features              map[string]interface{} `json:"features" db:"features"`
	IsActive              bool                   `json:"is_active" db:"is_active"`
	IsVisible             bool                   `json:"is_visible" db:"is_visible"`
	SortOrder             int                    `json:"sort_order" db:"sort_order"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// Subscription represents a tenant's subscription to a plan
type Subscription struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	TenantID           uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	PlanID             uuid.UUID              `json:"plan_id" db:"plan_id"`
	Status             string                 `json:"status" db:"status"`               // trial, active, past_due, cancelled, expired
	BillingCycle       string                 `json:"billing_cycle" db:"billing_cycle"` // monthly, yearly
	Amount             float64                `json:"amount" db:"amount"`
	Currency           string                 `json:"currency" db:"currency"`
	TrialEndsAt        *time.Time             `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	CurrentPeriodStart time.Time              `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time              `json:"current_period_end" db:"current_period_end"`
	CancelledAt        *time.Time             `json:"cancelled_at,omitempty" db:"cancelled_at"`
	AutoRenew          bool                   `json:"auto_renew" db:"auto_renew"`
	Notes              *string                `json:"notes,omitempty" db:"notes"`
	Metadata           map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`

	// Joined fields
	Plan       *SubscriptionPlan `json:"plan,omitempty"`
	TenantName string            `json:"tenant_name,omitempty"`
}

// OrganizationDetail represents extended tenant information
type OrganizationDetail struct {
	ID                        uuid.UUID              `json:"id" db:"id"`
	TenantID                  uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	AdminName                 *string                `json:"admin_name,omitempty" db:"admin_name"`
	AdminEmail                *string                `json:"admin_email,omitempty" db:"admin_email"`
	ContactNumber             *string                `json:"contact_number,omitempty" db:"contact_number"`
	Website                   *string                `json:"website,omitempty" db:"website"`
	AddressLine1              *string                `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2              *string                `json:"address_line2,omitempty" db:"address_line2"`
	City                      *string                `json:"city,omitempty" db:"city"`
	State                     *string                `json:"state,omitempty" db:"state"`
	Country                   *string                `json:"country,omitempty" db:"country"`
	PostalCode                *string                `json:"postal_code,omitempty" db:"postal_code"`
	CompanyRegistrationNumber *string                `json:"company_registration_number,omitempty" db:"company_registration_number"`
	TaxID                     *string                `json:"tax_id,omitempty" db:"tax_id"`
	Currency                  string                 `json:"currency" db:"currency"`
	Timezone                  string                 `json:"timezone" db:"timezone"`
	LogoURL                   *string                `json:"logo_url,omitempty" db:"logo_url"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt                 time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time              `json:"updated_at" db:"updated_at"`
}

// UsageMetric represents daily usage metrics for a tenant
type UsageMetric struct {
	ID                     uuid.UUID              `json:"id" db:"id"`
	TenantID               uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	MetricDate             time.Time              `json:"metric_date" db:"metric_date"`
	TotalUsers             int                    `json:"total_users" db:"total_users"`
	ActiveUsers            int                    `json:"active_users" db:"active_users"`
	NewUsers               int                    `json:"new_users" db:"new_users"`
	StorageUsedMB          int64                  `json:"storage_used_mb" db:"storage_used_mb"`
	StorageDocumentsMB     int64                  `json:"storage_documents_mb" db:"storage_documents_mb"`
	StorageAttachmentsMB   int64                  `json:"storage_attachments_mb" db:"storage_attachments_mb"`
	APIRequestsCount       int                    `json:"api_requests_count" db:"api_requests_count"`
	APIRequestsSuccess     int                    `json:"api_requests_success" db:"api_requests_success"`
	APIRequestsFailed      int                    `json:"api_requests_failed" db:"api_requests_failed"`
	LoginsCount            int                    `json:"logins_count" db:"logins_count"`
	UniqueLoginsCount      int                    `json:"unique_logins_count" db:"unique_logins_count"`
	AttendanceRecordsCount int                    `json:"attendance_records_count" db:"attendance_records_count"`
	LeaveRequestsCount     int                    `json:"leave_requests_count" db:"leave_requests_count"`
	EmployeesAddedCount    int                    `json:"employees_added_count" db:"employees_added_count"`
	AdditionalMetrics      map[string]interface{} `json:"additional_metrics,omitempty" db:"additional_metrics"`
	CreatedAt              time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
}

// APIRequestLog represents a single API request log entry
type APIRequestLog struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID            *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Method            string     `json:"method" db:"method"`
	Endpoint          string     `json:"endpoint" db:"endpoint"`
	StatusCode        *int       `json:"status_code,omitempty" db:"status_code"`
	ResponseTimeMS    *int       `json:"response_time_ms,omitempty" db:"response_time_ms"`
	RequestSizeBytes  *int       `json:"request_size_bytes,omitempty" db:"request_size_bytes"`
	ResponseSizeBytes *int       `json:"response_size_bytes,omitempty" db:"response_size_bytes"`
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	ErrorMessage      *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// Tenant represents an organization/tenant (updated model)
type Tenant struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Subdomain        string                 `json:"subdomain" db:"subdomain"`
	Domain           *string                `json:"domain,omitempty" db:"domain"`
	Status           string                 `json:"status" db:"status"` // active, suspended, inactive
	Country          *string                `json:"country,omitempty" db:"country"`
	AdminEmail       *string                `json:"admin_email,omitempty" db:"admin_email"`
	StorageUsedMB    int64                  `json:"storage_used_mb" db:"storage_used_mb"`
	APIRequestsCount int                    `json:"api_requests_count" db:"api_requests_count"`
	LastActivityAt   *time.Time             `json:"last_activity_at,omitempty" db:"last_activity_at"`
	Settings         map[string]interface{} `json:"settings,omitempty" db:"settings"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`

	// Joined fields
	Subscription       *Subscription       `json:"subscription,omitempty"`
	OrganizationDetail *OrganizationDetail `json:"organization_detail,omitempty"`
	TotalUsers         int                 `json:"total_users,omitempty"`
	ActiveEmployees    int                 `json:"active_employees,omitempty"`
}
