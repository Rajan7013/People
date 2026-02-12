package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Connection string from your .env
	connStr := "postgresql://postgres:Peoples9331%24%40%40%21-%2B%3Dcv.@db.boxukijspmnlbbjndugy.supabase.co:5432/postgres"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	fmt.Println("Connected to database. Running migration 035...")

	migrationSQL, err := os.ReadFile("migrations/036_add_deleted_at_to_plans.sql")
	if err != nil {
		log.Fatal("Could not read migration file:", err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Migration 035 applied successfully! Soft delete rules dropped.")
}
