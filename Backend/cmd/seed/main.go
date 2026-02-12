package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration (including .env)
	cfg, err := config.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load config, using defaults")
	}

	if cfg.DatabaseURL == "" {
		log.Fatal().Msg("DATABASE_URL is not set")
	}

	log.Info().Msg("Connecting to database...")
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer conn.Close(context.Background())

	ctx := context.Background()

	// 1. Create Default Tenant if not exists
	var tenantID string
	var tenantName = "PeopleOS HQ"

	err = conn.QueryRow(ctx, "SELECT id FROM tenants WHERE name = $1", tenantName).Scan(&tenantID)
	if err == pgx.ErrNoRows {
		log.Info().Msg("Creating default tenant: PeopleOS HQ")
		err = conn.QueryRow(ctx, `
			INSERT INTO tenants (name, subdomain, status, plan)
			VALUES ($1::text, $2::text, 'active', 'enterprise')
			RETURNING id::text
		`, tenantName, "hq").Scan(&tenantID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create tenant")
		}
	} else if err != nil {
		log.Fatal().Err(err).Msg("Failed to query tenant")
	} else {
		log.Info().Str("id", tenantID).Msg("PeopleOS HQ tenant already exists")
	}

	// 2. Create Super Admin User
	adminEmail := "rajanprasaila@gmail.com"
	initialPassword := "PeopleOS@2026" // Temporary initial password

	var userID string
	err = conn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", adminEmail).Scan(&userID)

	if err == pgx.ErrNoRows {
		log.Info().Str("email", adminEmail).Msg("Creating Super Admin user")

		// Hash password using the SAME logic as the backend
		hashedPassword, err := auth.HashPassword(initialPassword, cfg.PepperSecret)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to hash password")
		}

		// Insert User
		err = conn.QueryRow(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, role, first_name, last_name, is_active, email_verified_at)
			VALUES ($1, $2, $3, 'super_admin', 'Rajan', 'Prasaila', true, NOW())
			RETURNING id
		`, tenantID, adminEmail, hashedPassword).Scan(&userID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create user")
		}

		// Create Employee Record (Required for many features)
		// Assuming we have an 'employees' table linked to 'users'
		employeeCode := "EMP-001"
		_, err = conn.Exec(ctx, `
			INSERT INTO employees (tenant_id, user_id, employee_code, job_title, employment_type, date_of_joining)
			VALUES ($1, $2, $3, 'Chief Executive Officer', 'full_time', $4)
		`, tenantID, userID, employeeCode, time.Now())

		if err != nil {
			// It might fail if employee record already exists (unlikely if user didn't exist)
			log.Warn().Err(err).Msg("Failed to create employee record (ignoring if duplicate)")
		}

		log.Info().Msg("---------------------------------------------------------")
		log.Info().Msg(" SUPER ADMIN CREATED SUCCESSFULLY")
		log.Info().Str("Email", adminEmail).Msg("")
		log.Info().Str("Password", initialPassword).Msg("")
		log.Info().Msg(" PLEASE CHANGE THIS PASSWORD AFTER FIRST LOGIN")
		log.Info().Msg("---------------------------------------------------------")

	} else if err != nil {
		log.Fatal().Err(err).Msg("Failed to query user")
	} else {
		log.Info().Str("id", userID).Msg("Super Admin user already exists")
	}

	// 3. Populate System Settings for this Tenant
	_, err = conn.Exec(ctx, `
		INSERT INTO system_settings (tenant_id, setting_key, setting_value, setting_type, description)
		SELECT $1, d.key, d.val, d.type, d.description
		FROM (VALUES
			('company_name', 'PeopleOS HQ', 'string', 'Company name'),
			('timezone', 'Asia/Kathmandu', 'string', 'Default timezone'),
			('currency', 'NPR', 'string', 'Default currency')
		) AS d(key, val, type, description)
		WHERE NOT EXISTS (SELECT 1 FROM system_settings WHERE tenant_id = $1 AND setting_key = d.key)
	`, tenantID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to seed system settings")
	}

	log.Info().Msg("Seeding completed successfully!")
}
