package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/config"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/db"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/handlers"
	custommiddleware "github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/middleware"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/services"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config               *config.Config
	router               chi.Router
	db                   *sql.DB
	authService          *auth.Service
	employeeService      *services.EmployeeService
	employeeHandler      *handlers.EmployeeHandler
	attendanceService    *services.AttendanceService
	attendanceHandler    *handlers.AttendanceHandler
	biometricService     *services.BiometricService
	biometricHandler     *handlers.BiometricHandler
	leaveService         *services.LeaveService
	leaveHandler         *handlers.LeaveHandler
	payslipService       *services.PayslipService
	payslipHandler       *handlers.PayslipHandler
	systemService        *services.SystemManagementService
	systemHandler        *handlers.SystemManagementHandler
	userSettingsService  *services.UserSettingsService
	userSettingsHandler  *handlers.UserSettingsHandler
	dashboardService     *services.DashboardService
	dashboardHandler     *handlers.DashboardHandler
	subscriptionService  *services.SubscriptionService
	organizationService  *services.OrganizationService
	invoiceService       *services.InvoiceService
	usageTrackingService *services.UsageTrackingService
	analyticsService     *services.AnalyticsService
	departmentHandler    *handlers.DepartmentHandler
	policyHandler        *handlers.PolicyHandler
	superAdminHandler    *handlers.SuperAdminHandler
	tenantHandler        *handlers.TenantHandler
	organizationHandler  *handlers.OrganizationHandler
	rlsMiddleware        *custommiddleware.RLSMiddleware
}

func New(cfg *config.Config) (*Server, error) {
	// Initialize database connection
	database, err := db.Connect(cfg)
	if err != nil {
		return nil, err
	}

	// MANUAL MIGRATION FIX: Update employment_status check constraint
	log.Info().Msg("Applying migration for employment_status constraint...")
	// We swallow error on drop in case it doesn't exist (though it does), or use IF EXISTS syntax if postgres supports it for constraint
	_, _ = database.Exec(`ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_employment_status_check`)
	_, err = database.Exec(`ALTER TABLE employees ADD CONSTRAINT employees_employment_status_check 
        CHECK (employment_status IN ('active', 'inactive', 'terminated', 'suspended'))`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to apply employment_status constraint migration")
		// We warn but don't fail, in case it's already compliant or some other issue
	} else {
		log.Info().Msg("Successfully applied employment_status constraint migration")
	}

	// Initialize auth service
	authService := auth.NewService(
		database,
		cfg.JWTSecret,
		cfg.PepperSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	// Initialize employee service and handler
	employeeService := services.NewEmployeeService(database, cfg.PepperSecret, cfg.EncryptionKey)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)

	// Initialize attendance service and handler
	attendanceService := services.NewAttendanceService(database)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceService)

	// Initialize biometric service and handler
	biometricService := services.NewBiometricService(database)
	biometricHandler := handlers.NewBiometricHandler(biometricService)

	// Initialize leave service and handler
	leaveService := services.NewLeaveService(database)
	leaveHandler := handlers.NewLeaveHandler(leaveService)

	// Initialize payslip service and handler
	dbx := sqlx.NewDb(database, "pgx")
	payslipService := services.NewPayslipService(dbx)
	payslipHandler := handlers.NewPayslipHandler(payslipService)

	// Initialize system management service and handler
	systemService := services.NewSystemManagementService(database)
	systemHandler := handlers.NewSystemManagementHandler(systemService)

	// Initialize user settings service and handler
	userSettingsService := services.NewUserSettingsService(database)
	userSettingsHandler := handlers.NewUserSettingsHandler(userSettingsService)

	// Initialize dashboard service and handler
	dashboardService := services.NewDashboardService(database)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// Initialize Super Admin services
	subscriptionService := services.NewSubscriptionService(database)
	organizationService := services.NewOrganizationService(database, subscriptionService, cfg.PepperSecret)
	invoiceService := services.NewInvoiceService(database)
	usageTrackingService := services.NewUsageTrackingService(database)
	analyticsService := services.NewAnalyticsService(database)
	departmentService := services.NewDepartmentService(database)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	policyService := services.NewPolicyService(database)
	policyHandler := handlers.NewPolicyHandler(policyService)
	superAdminService := services.NewSuperAdminService(database, cfg.PepperSecret)

	superAdminHandler := handlers.NewSuperAdminHandler(
		organizationService,
		subscriptionService,
		invoiceService,
		usageTrackingService,
		analyticsService,
		superAdminService,
	)

	tenantHandler := handlers.NewTenantHandler(organizationService)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)

	// Initialize RLS middleware
	rlsMiddleware := custommiddleware.NewRLSMiddleware(database)

	s := &Server{
		config:               cfg,
		router:               chi.NewRouter(),
		db:                   database,
		authService:          authService,
		employeeService:      employeeService,
		employeeHandler:      employeeHandler,
		attendanceService:    attendanceService,
		attendanceHandler:    attendanceHandler,
		biometricService:     biometricService,
		biometricHandler:     biometricHandler,
		leaveService:         leaveService,
		leaveHandler:         leaveHandler,
		payslipService:       payslipService,
		payslipHandler:       payslipHandler,
		systemService:        systemService,
		systemHandler:        systemHandler,
		userSettingsService:  userSettingsService,
		userSettingsHandler:  userSettingsHandler,
		dashboardService:     dashboardService,
		dashboardHandler:     dashboardHandler,
		subscriptionService:  subscriptionService,
		organizationService:  organizationService,
		invoiceService:       invoiceService,
		usageTrackingService: usageTrackingService,
		analyticsService:     analyticsService,
		departmentHandler:    departmentHandler,
		policyHandler:        policyHandler,
		superAdminHandler:    superAdminHandler,
		tenantHandler:        tenantHandler,
		organizationHandler:  organizationHandler,
		rlsMiddleware:        rlsMiddleware,
	}

	// Initialize Google OAuth
	auth.InitGoogleOAuth()

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupMiddleware() {
	// Basic middleware stack
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Timeout middleware
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Security Headers
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Content-Security-Policy", "default-src 'self' http: https: data: blob: 'unsafe-inline' 'unsafe-eval'")
			next.ServeHTTP(w, r)
		})
	})

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	// Health check endpoints
	s.router.Get("/health", s.healthHandler)
	s.router.Get("/ready", s.readinessHandler)

	// API v1 routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// ========================================
		// PUBLIC ROUTES
		// ========================================
		r.Route("/auth", func(r chi.Router) {
			authHandler := auth.NewHandler(s.authService)
			authHandler.RegisterRoutes(r)
		})

		// ========================================
		// PLATFORM ROUTES (Super Admin ONLY)
		// ========================================
		r.Route("/platform", func(r chi.Router) {
			r.Use(s.authService.Middleware())                 // JWT validation
			r.Use(custommiddleware.CheckUserStatus(s.db))     // Check user is_active status
			r.Use(s.rlsMiddleware.SetSessionContextEfficient) // RLS context
			r.Use(custommiddleware.RequireSuperAdmin)         // RBAC check

			// Organizations
			r.Route("/organizations", func(r chi.Router) {
				r.Get("/", s.superAdminHandler.GetAllOrganizations)
				r.Post("/", s.superAdminHandler.CreateOrganization)
				r.Get("/{id}", s.superAdminHandler.GetOrganization)
				r.Put("/{id}", s.superAdminHandler.UpdateOrganization)
				r.Delete("/{id}", s.superAdminHandler.DeleteOrganization)
				r.Post("/{id}/block", s.superAdminHandler.BlockOrganization)
				r.Post("/{id}/unblock", s.superAdminHandler.UnblockOrganization)
				r.Post("/{id}/renew", s.superAdminHandler.RenewOrganizationSubscription)
			})

			// Subscription Plans
			r.Route("/plans", func(r chi.Router) {
				r.Get("/", s.superAdminHandler.GetAllPlans)
				r.Post("/", s.superAdminHandler.CreatePlan)
				r.Get("/{id}", s.superAdminHandler.GetPlan)
				r.Put("/{id}", s.superAdminHandler.UpdatePlan)
				r.Delete("/{id}", s.superAdminHandler.DeletePlan)
			})

			// Invoices
			r.Route("/invoices", func(r chi.Router) {
				r.Get("/", s.superAdminHandler.GetAllInvoices)
				r.Post("/generate", s.superAdminHandler.GenerateBill)
				r.Get("/{id}", s.superAdminHandler.GetInvoice)
				r.Put("/{id}", s.superAdminHandler.UpdateInvoice)
				r.Delete("/{id}", s.superAdminHandler.DeleteInvoice)
				r.Post("/{id}/mark-paid", s.superAdminHandler.MarkInvoiceAsPaid)
				r.Get("/{id}/download", s.superAdminHandler.DownloadInvoice)
			})

			// Analytics & Usage
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/platform-stats", s.superAdminHandler.GetPlatformStats)
				r.Get("/tenant-growth", s.superAdminHandler.GetTenantGrowth)
				r.Get("/revenue", s.superAdminHandler.GetRevenueMetrics)
			})

			r.Route("/usage", func(r chi.Router) {
				r.Get("/organizations/{id}", s.superAdminHandler.GetOrganizationUsage)
			})

			// Super Admin Management
			r.Route("/admins", func(r chi.Router) {
				r.Post("/", s.superAdminHandler.CreateSuperAdmin)
				r.Get("/", s.superAdminHandler.GetAllSuperAdmins)
			})

			// System Management
			r.Route("/system", func(r chi.Router) {
				// System Settings
				r.Get("/settings", s.systemHandler.GetSettings)
				r.Post("/settings", s.systemHandler.CreateSetting)
				r.Put("/settings/{key}", s.systemHandler.UpdateSetting)

				// Audit Logs
				r.Get("/audit-logs", s.systemHandler.GetAuditLogs)

				// System Metrics
				r.Get("/metrics", s.systemHandler.GetMetrics)
				r.Post("/metrics", s.systemHandler.RecordMetric)

				// System Backups
				r.Get("/backups", s.systemHandler.GetBackups)
				r.Post("/backups", s.systemHandler.CreateBackup)
			})
		})

		// ========================================
		// COMPANY ROUTES (Tenant-scoped)
		// ========================================
		r.Route("/company", func(r chi.Router) {
			r.Use(s.authService.Middleware())                      // JWT validation
			r.Use(custommiddleware.CheckUserStatus(s.db))          // Check user is_active status
			r.Use(s.rlsMiddleware.SetSessionContextEfficient)      // RLS context
			r.Use(custommiddleware.BlockSuperAdminFromCompanyData) // Prevent Super Admin access

			// ------------------------------------
			// Admin Routes (Org Admin)
			// ------------------------------------
			r.Route("/admin", func(r chi.Router) {
				r.Use(custommiddleware.RequireOrgAdmin)

				// Employee Management
				r.Route("/employees", func(r chi.Router) {
					r.Get("/", s.getEmployeesHandler)
					r.Post("/", s.createEmployeeHandler)
					r.Get("/{employeeID}", s.getEmployeeHandler)
					r.Put("/{employeeID}", s.updateEmployeeHandler)
					r.Put("/{employeeID}/status", s.employeeHandler.UpdateEmployeeStatus)
					r.Delete("/{employeeID}", s.deleteEmployeeHandler)
				})

				// Department Management
				r.Route("/departments", func(r chi.Router) {
					r.Get("/", s.departmentHandler.GetDepartments)
					r.Post("/", s.departmentHandler.CreateDepartment)
					r.Put("/{departmentID}", s.departmentHandler.UpdateDepartment)
					r.Delete("/{departmentID}", s.departmentHandler.DeleteDepartment)
				})

				// Policy Configuration
				r.Route("/policies", func(r chi.Router) {
					// Attendance Policy
					r.Get("/attendance", s.policyHandler.GetAttendancePolicy)
					r.Put("/attendance", s.policyHandler.UpdateAttendancePolicy)

					// Salary Components
					r.Get("/salary-components", s.policyHandler.GetSalaryComponents)
					r.Post("/salary-components", s.policyHandler.CreateSalaryComponent)

					// Leave Types
					r.Get("/leave-types", s.policyHandler.GetLeaveTypes)
					r.Post("/leave-types", s.policyHandler.CreateLeaveType)
				})

				// Tenant Configuration
				r.Route("/config", func(r chi.Router) {
					r.Get("/", s.tenantHandler.GetConfig)
					r.Put("/", s.tenantHandler.UpdateConfig)
				})

				// Biometric Devices
				r.Route("/biometric", func(r chi.Router) {
					r.Get("/devices", s.biometricHandler.GetDevices)
					r.Post("/devices", s.biometricHandler.RegisterDevice)
				})

				// Organization Profile
				r.Get("/organization", s.organizationHandler.GetOrganizationProfile)
				r.Put("/organization", s.organizationHandler.UpdateOrganizationProfile)
			})

			// ------------------------------------
			// Manager Routes (Manager + Admin)
			// ------------------------------------
			r.Route("/manager", func(r chi.Router) {
				r.Use(custommiddleware.RequireManager)

				// Team Management
				r.Get("/team", s.employeeHandler.GetMyTeam)

				// Department Leaves
				r.Get("/leaves", s.leaveHandler.GetDepartmentLeaves)
				r.Put("/leaves/{id}/approve", s.leaveHandler.ApproveLeave)
				r.Put("/leaves/{id}/reject", s.leaveHandler.RejectLeave)

				// Department Attendance
				r.Get("/attendance", s.attendanceHandler.GetDepartmentAttendance)
			})

			// ------------------------------------
			// HR Routes (HR + Admin)
			// ------------------------------------
			r.Route("/hr", func(r chi.Router) {
				r.Use(custommiddleware.RequireHR)

				// Employee Directory (Filtered view)
				r.Get("/employees", s.getEmployeesHandler)
				r.Get("/employees/{employeeID}", s.getEmployeeHandler)
				r.Put("/employees/{employeeID}", s.updateEmployeeHandler)

				// Department & Position Metadata (For filtering)
				r.Get("/departments", s.departmentHandler.GetDepartments)
				// r.Get("/positions", s.departmentHandler.GetPositions) // If it exists, otherwise just departments

				// Attendance Management
				r.Route("/attendance", func(r chi.Router) {
					r.Get("/", s.attendanceHandler.GetAttendanceRecords)
					r.Get("/today", s.attendanceHandler.GetTodayAttendance)
					r.Get("/stats", s.attendanceHandler.GetAttendanceStats)
					r.Get("/employees/{employeeId}", s.attendanceHandler.GetEmployeeAttendance)
					r.Put("/records/{recordId}", s.attendanceHandler.UpdateAttendanceRecord)
					r.Post("/policies", s.attendanceHandler.CreateAttendancePolicy)
				})

				// Leave Management
				r.Route("/leaves", func(r chi.Router) {
					r.Get("/", s.getLeavesHandler)
					r.Put("/{id}/approve", s.approveLeaveHandler)
					r.Put("/{id}/reject", s.rejectLeaveHandler)
					r.Post("/types", s.leaveHandler.CreateLeaveType)
				})

				// Payslip Management
				r.Route("/payslips", func(r chi.Router) {
					r.Get("/", s.payslipHandler.GetPayslips)
					r.Post("/", s.payslipHandler.CreatePayslip)
					r.Get("/stats", s.payslipHandler.GetPayslipStats)
					r.Put("/{id}", s.payslipHandler.UpdatePayslip)
					r.Delete("/{id}", s.payslipHandler.DeletePayslip)
				})

				// Biometric Logs
				r.Get("/biometric/logs", s.biometricHandler.GetBiometricLogs)
			})

			// ------------------------------------
			// Team Lead Routes (Team Lead + Manager + Admin)
			// ------------------------------------
			r.Route("/team-lead", func(r chi.Router) {
				r.Use(custommiddleware.RequireTeamLead)

				// Team view
				r.Get("/team", s.employeeHandler.GetMyTeam)
				r.Get("/attendance", s.attendanceHandler.GetTeamAttendance)
			})

			// ------------------------------------
			// Employee Routes (All Authenticated Roles)
			// ------------------------------------
			r.Route("/employee", func(r chi.Router) {
				// No specific role check needed as all users are employees in this context
				// But we ensure they are authenticated via middleware stack

				// Profile
				r.Get("/profile", s.profileHandler)

				// Attendance
				r.Route("/attendance", func(r chi.Router) {
					r.Get("/", s.attendanceHandler.GetAttendanceRecords) // Will filter by self via RLS
					r.Get("/my-status", s.attendanceHandler.GetCurrentUserStatus)
					r.Post("/checkin", s.attendanceHandler.CheckIn)
					r.Post("/checkout", s.attendanceHandler.CheckOut)
				})

				// Leaves
				r.Route("/leaves", func(r chi.Router) {
					r.Get("/", s.getLeavesHandler) // Will filter by self via RLS
					r.Post("/", s.createLeaveHandler)
					r.Get("/types", s.policyHandler.GetLeaveTypes)
				})

				// Payslips
				r.Get("/payslips", s.payslipHandler.GetPayslips) // Will filter by self via RLS (need to verify handler)
				r.Get("/payslips/{id}", s.payslipHandler.GetPayslip)

				// Dashboard
				r.Get("/dashboard/stats", s.getDashboardStatsHandler)

				// Settings
				r.Route("/settings", func(r chi.Router) {
					r.Get("/profile", s.userSettingsHandler.GetUserProfile)
					r.Put("/profile", s.userSettingsHandler.UpdateUserProfile)
					r.Get("/preferences", s.userSettingsHandler.GetUserPreferences)
					r.Put("/preferences", s.userSettingsHandler.UpdateUserPreferences)
					r.Get("/security", s.userSettingsHandler.GetSecuritySettings)
					r.Put("/security", s.userSettingsHandler.UpdateSecuritySettings)
					r.Get("/theme", s.userSettingsHandler.GetUserTheme)
					r.Put("/theme", s.userSettingsHandler.UpdateUserTheme)
				})
			})
		})
	})
}

func (s *Server) Routes() http.Handler {
	return s.router
}

func (s *Server) Close() {
	// Close database connections, Redis, etc.
	if s.db != nil {
		db.Close(s.db)
	}
	log.Info().Msg("Server connections closed")
}

// Handler functions (basic implementations)
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"peopleos-api"}`))
}

func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","service":"peopleos-api","error":"database connection failed"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"peopleos-api"}`))
}

func (s *Server) profileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Missing authentication"}`))
		return
	}

	// Return user profile (without sensitive data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := `{
		"user_id": "` + claims.UserID + `",
		"tenant_id": "` + claims.TenantID + `",
		"email": "` + claims.Email + `",
		"role": "` + claims.Role + `"
	}`
	w.Write([]byte(response))
}

// Employee handlers - delegate to EmployeeHandler
func (s *Server) getEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	s.employeeHandler.GetEmployees(w, r)
}

func (s *Server) createEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	s.employeeHandler.CreateEmployee(w, r)
}

func (s *Server) getEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	s.employeeHandler.GetEmployee(w, r)
}

func (s *Server) updateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	s.employeeHandler.UpdateEmployee(w, r)
}

func (s *Server) deleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	s.employeeHandler.DeleteEmployee(w, r)
}

// Departments handler

func (s *Server) getLeavesHandler(w http.ResponseWriter, r *http.Request) {
	s.leaveHandler.GetLeaveRequests(w, r)
}

func (s *Server) createLeaveHandler(w http.ResponseWriter, r *http.Request) {
	s.leaveHandler.CreateLeaveRequest(w, r)
}

func (s *Server) approveLeaveHandler(w http.ResponseWriter, r *http.Request) {
	s.leaveHandler.ApproveLeave(w, r)
}

func (s *Server) rejectLeaveHandler(w http.ResponseWriter, r *http.Request) {
	s.leaveHandler.RejectLeave(w, r)
}

func (s *Server) getDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	s.dashboardHandler.GetStats(w, r)
}
