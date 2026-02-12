package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/config"
	"github.com/rs/zerolog/log"
)

// Connect creates a database connection
func Connect(cfg *config.Config) (*sql.DB, error) {
	var dsn string

	// Use DATABASE_URL if available, otherwise build from individual components
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
			cfg.DBSSLMode,
		)
	}

	log.Info().
		Str("host", cfg.DBHost).
		Int("port", cfg.DBPort).
		Str("database", cfg.DBName).
		Msg("Connecting to database...")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Database connection established successfully")

	return db, nil
}

// Close closes the database connection
func Close(db *sql.DB) error {
	if db != nil {
		log.Info().Msg("Closing database connection...")
		return db.Close()
	}
	return nil
}
