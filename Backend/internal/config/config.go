package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port           int      `json:"port"`
	Environment    string   `json:"environment"`
	LogLevel       string   `json:"log_level"`
	ReadTimeout    int      `json:"read_timeout"`
	WriteTimeout   int      `json:"write_timeout"`
	IdleTimeout    int      `json:"idle_timeout"`
	AllowedOrigins []string `json:"allowed_origins"`

	// Database configuration
	DatabaseURL string `json:"database_url"`
	DBHost      string `json:"db_host"`
	DBPort      int    `json:"db_port"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
	DBSSLMode   string `json:"db_ssl_mode"`

	// Redis configuration
	RedisURL      string `json:"redis_url"`
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// Password Security
	PepperSecret  string `json:"pepper_secret"`
	EncryptionKey string `json:"encryption_key"`

	// JWT configuration
	JWTSecret       string `json:"jwt_secret"`
	AccessTokenTTL  int    `json:"access_token_ttl"`  // in minutes
	RefreshTokenTTL int    `json:"refresh_token_ttl"` // in minutes (absolute timeout)

	// File storage
	S3Endpoint  string `json:"s3_endpoint"`
	S3Region    string `json:"s3_region"`
	S3Bucket    string `json:"s3_bucket"`
	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`

	// Email configuration
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
}

func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		// Server defaults
		Port:         getEnvAsInt("PORT", 8080),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
		WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
		IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),

		// Database configuration
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnvAsInt("DB_PORT", 5432),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "peopleos"),
		DBSSLMode:   getEnv("DB_SSL_MODE", "disable"),

		// Redis configuration
		RedisURL:      getEnv("REDIS_URL", ""),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Password Security
		PepperSecret:  getEnv("PEPPER_SECRET", "change-me-to-a-long-random-string-in-prod"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "change-me-to-a-32-byte-secret-in-production"),

		// JWT configuration
		JWTSecret:       getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		AccessTokenTTL:  getEnvAsInt("ACCESS_TOKEN_TTL", 15),   // 15 minutes idle timeout
		RefreshTokenTTL: getEnvAsInt("REFRESH_TOKEN_TTL", 720), // 12 hours absolute timeout

		// File storage
		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3Bucket:    getEnv("S3_BUCKET", "peopleos-documents"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),

		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@peopleos.com"),

		// CORS configuration
		AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000", "https://*.peopleos.com"}),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Assume comma-separated values
	return strings.Split(value, ",")
}
