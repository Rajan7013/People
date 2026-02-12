package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/config"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Force connection string if not set (fallback for dev)
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	}

	ctx := context.Background()

	// ---------------------------------------------------------
	// 1. Setup Admin Connection (to create app user)
	// ---------------------------------------------------------
	dbAdmin, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open DB as admin: %v", err)
	}
	defer dbAdmin.Close()

	if err := dbAdmin.Ping(); err != nil {
		log.Fatalf("Failed to ping DB as admin: %v", err)
	}
	fmt.Println("✅ Connected as Admin (postgres)")

	// Create App User (if not exists)
	_, err = dbAdmin.ExecContext(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'peopleos_app') THEN
				CREATE ROLE peopleos_app WITH LOGIN PASSWORD 'peopleos_secure_pass';
			END IF;
			ALTER ROLE peopleos_app WITH NOBYPASSRLS;
		END
		$$;
	`)
	if err != nil {
		log.Fatalf("Failed to create app user: %v", err)
	}

	// Grant Permissions
	grantQuery := fmt.Sprintf(`
		GRANT CONNECT ON DATABASE %s TO peopleos_app;
		GRANT USAGE ON SCHEMA public TO peopleos_app;
		GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO peopleos_app;
		GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO peopleos_app;
	`, cfg.DBName)

	_, err = dbAdmin.ExecContext(ctx, grantQuery)
	if err != nil {
		log.Fatalf("Failed to grant permissions: %v", err)
	}
	fmt.Println("✅ Configured 'peopleos_app' user with NOBYPASSRLS")

	// ---------------------------------------------------------
	// 2. Setup Test Tenants (as Admin)
	// ---------------------------------------------------------
	tenantA := "11111111-1111-1111-1111-111111111111"
	tenantB := "99999999-9999-9999-9999-999999999999"

	// Cleanup old test data
	dbAdmin.ExecContext(ctx, "DELETE FROM tenants WHERE id IN ($1, $2)", tenantA, tenantB)

	// Create Tenants
	_, err = dbAdmin.ExecContext(ctx, `
		INSERT INTO tenants (id, name, subdomain, status) VALUES 
		($1, 'Victim Corp', 'victim', 'active'),
		($2, 'Hacker Inc', 'hacker', 'active')
	`, tenantA, tenantB)
	if err != nil {
		log.Fatalf("Failed to create tenants: %v", err)
	}
	fmt.Println("✅ Created Test Tenants")

	// ---------------------------------------------------------
	// 3. Connect as App User (RLS Enforced)
	// ---------------------------------------------------------
	appConnStr := fmt.Sprintf("postgres://peopleos_app:peopleos_secure_pass@%s:%d/%s?sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	dbApp, err := sql.Open("pgx", appConnStr)
	if err != nil {
		log.Fatalf("Failed to connect as app user: %v", err)
	}
	defer dbApp.Close()
	fmt.Println("✅ Connected as App User (peopleos_app)")

	// ---------------------------------------------------------
	// 4. Create Victim Data (as App User, assuming Tenant A)
	// ---------------------------------------------------------
	txA, err := dbApp.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// ---------------------------------------------------------
	// DEBUG: Investigate RLS Status
	// ---------------------------------------------------------
	fmt.Println("\n🔎 DEBUG: Inspecting RLS Configuration...")

	// Check if RLS is enabled on table
	var rlsEnabled bool
	err = dbApp.QueryRowContext(ctx, "SELECT relrowsecurity FROM pg_class WHERE relname = 'system_settings'").Scan(&rlsEnabled)
	if err != nil {
		fmt.Printf("⚠️  Could not query pg_class: %v\n", err)
	} else {
		fmt.Printf("   > RLS Enabled on 'system_settings': %v\n", rlsEnabled)
	}

	// Check active policies
	rows, err := dbApp.QueryContext(ctx, "SELECT polname, polcmd FROM pg_policy WHERE polrelid = 'system_settings'::regclass")
	if err != nil {
		fmt.Printf("⚠️  Could not query pg_policy: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("   > Active Policies on 'system_settings':")
		for rows.Next() {
			var polName, polCmd string
			rows.Scan(&polName, &polCmd)
			fmt.Printf("     - %s (%s)\n", polName, polCmd)
		}
	}

	// Check current user and role attributes
	var currentUser, currentRole string
	var isSuperuser, bypassRLS bool
	err = dbApp.QueryRowContext(ctx, `
		SELECT current_user, current_role, rolsuper, rolbypassrls 
		FROM pg_roles WHERE rolname = current_user
	`).Scan(&currentUser, &currentRole, &isSuperuser, &bypassRLS)

	if err != nil {
		fmt.Printf("⚠️  Could not query user info: %v\n", err)
	} else {
		fmt.Printf("   > Current User: %s (Role: %s)\n", currentUser, currentRole)
		fmt.Printf("   > Is Superuser: %v\n", isSuperuser)
		fmt.Printf("   > Bypass RLS: %v\n", bypassRLS)
	}

	_, err = txA.ExecContext(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", tenantA))
	if err != nil {
		txA.Rollback()
		log.Fatalf("Failed to set Victim context: %v", err)
	}

	_, err = txA.ExecContext(ctx, `
		INSERT INTO system_settings (tenant_id, setting_key, setting_value, is_sensitive)
		VALUES ($1, 'SECRET_KEY', 'VICTIM_SUPER_SECRET_DATA', true)
	`, tenantA)
	if err != nil {
		txA.Rollback()
		log.Fatalf("Failed to insert victim data: %v", err)
	}
	txA.Commit()
	fmt.Println("✅ Inserted Sensitive Data as Victim (via App User)")

	// ---------------------------------------------------------
	// 5. ATTEMPT HACK: Try to read data as Hacker
	// ---------------------------------------------------------
	fmt.Println("\n🕵️  STARTING RLS ATTACK SIMULATION...")

	txB, err := dbApp.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer txB.Rollback()

	// Switch to Hacker Context
	_, err = txB.ExecContext(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", tenantB))
	if err != nil {
		log.Fatalf("Failed to set Hacker context: %v", err)
	}

	// Try to Select Victim's Data (Should return 0 rows)
	var count int
	err = txB.QueryRowContext(ctx, "SELECT count(*) FROM system_settings").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query as hacker: %v", err)
	}

	if count == 0 {
		fmt.Println("✅ SUCCESS: Hacker sees 0 rows! RLS is working.")
	} else {
		fmt.Printf("❌ FAILURE: Hacker saw %d rows! RLS IS BROKEN.\n", count)
		os.Exit(1)
	}

	// ---------------------------------------------------------
	// 6. ATTEMPT HACK: Try to Insert into Victim's Tenant
	// ---------------------------------------------------------
	fmt.Println("\n🕵️  STARTING INJECTION ATTACK...")

	_, err = txB.ExecContext(ctx, `
		INSERT INTO system_settings (tenant_id, setting_key, setting_value)
		VALUES ($1, 'HACKED_KEY', 'I_WAS_HERE')
	`, tenantA)

	if err != nil {
		fmt.Println("✅ SUCCESS: Injection Blocked! DB Error:", err)
	} else {
		fmt.Println("❌ FAILURE: Hacker successfully injected data into Victim tenant!")
		os.Exit(1)
	}

	fmt.Println("\n🎉 VERIFICATION COMPLETE: MILITARY GRADE SECURITY CONFIRMED.")
}
