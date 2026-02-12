package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load .env from multiple possible locations
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Warn().Msg("No .env file found")
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal().Msg("DATABASE_URL not set in environment")
	}

	// Use pgx.Connect directly (not database/sql)
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to database")
	}
	defer conn.Close(context.Background())

	var version string
	if err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatal().Err(err).Msg("Query failed")
	}
	log.Info().Str("version", version).Msg("Connected to Postgres")

	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name   = 'tenants'
	);`

	if err := conn.QueryRow(context.Background(), query).Scan(&exists); err != nil {
		log.Fatal().Err(err).Msg("Failed to check if table exists")
	}

	if exists {
		fmt.Println("VERIFICATION_SUCCESS: Tenants table exists!")
	} else {
		fmt.Println("VERIFICATION_FAILURE: Tenants table does NOT exist.")
	}
}
