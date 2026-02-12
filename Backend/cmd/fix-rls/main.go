package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	fmt.Println("✅ Connected to Database")

	// 1. Grant Permissions (Just in case)
	_, err = db.Exec(`
		GRANT CONNECT ON DATABASE peopleos TO peopleos_app;
		GRANT USAGE ON SCHEMA public TO peopleos_app;
		GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO peopleos_app;
		GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO peopleos_app;
	`)
	if err != nil {
		fmt.Printf("⚠️  Failed to grant permissions: %v\n", err)
	} else {
		fmt.Println("✅ Granted permissions to 'peopleos_app'")
	}

	// 1.5 Create Helper Function (if missing)
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION current_app_tenant() RETURNS uuid AS $$
			SELECT NULLIF(current_setting('app.current_tenant', TRUE), '')::uuid;
		$$ LANGUAGE sql STABLE;
	`)
	if err != nil {
		fmt.Printf("⚠️  Failed to create helper function: %v\n", err)
	} else {
		fmt.Println("✅ Created/Verified 'current_app_tenant()' function")
	}

	// 2. Force Enable RLS
	tables := []string{
		"system_settings", "tenants", "users", "employees", "departments",
		"attendance_policies", "attendance_records", "leave_types", "leave_balances",
		"leave_applications", "leave_requests", "audit_logs", "system_backups",
		"system_metrics", "api_keys", "webhooks", "user_profiles", "user_preferences",
		"security_settings", "user_themes", "salary_components", "employee_salary_structures",
		"salary_structure_components", "payslips", "payslip_components", "biometric_devices",
		"biometric_attendance_logs",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY;", table))
		if err != nil {
			fmt.Printf("⚠️  Failed to enable RLS on %s: %v\n", table, err)
		} else {
			fmt.Printf("✅ Enabled RLS on %s\n", table)
		}

		// Drop existing policy to avoid conflict
		_, _ = db.Exec(fmt.Sprintf("DROP POLICY IF EXISTS tenant_isolation_%s ON %s;", table, table))
		_, _ = db.Exec(fmt.Sprintf("DROP POLICY IF EXISTS tenant_isolation_sys_settings ON %s;", table)) // cleanup old name
	}

	// 3. Re-create Policy for system_settings specifically (as Proof of Concept)
	_, err = db.Exec(`
		CREATE POLICY tenant_isolation_sys_settings ON system_settings
			USING (tenant_id = current_app_tenant());
	`)
	if err != nil {
		fmt.Printf("⚠️  Failed to recreate policy on system_settings: %v\n", err)
	} else {
		fmt.Println("✅ Re-created RLS policy for 'system_settings'")
	}
}
