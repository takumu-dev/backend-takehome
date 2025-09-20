package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	CORS        CORSConfig
	Logging     LoggingConfig
	RateLimit   RateLimitConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	DSN      string
	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	Environment    string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	DefaultRequestsPerSecond float64
	DefaultBurstSize         int
	AuthRequestsPerSecond    float64
	AuthBurstSize            int
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "3306"))
	
	// Parse CORS allowed origins
	allowedOrigins := []string{}
	if originsStr := getEnv("ALLOWED_ORIGINS", ""); originsStr != "" {
		allowedOrigins = strings.Split(originsStr, ",")
		// Trim whitespace from each origin
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "blog_platform"),
			DSN:      getEnv("DB_DSN", "root:@tcp(localhost:3306)/blog_platform?parseTime=true"),
			// Connection pool settings
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25"), 25),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "5"), 5),
			ConnMaxLifetime: parseInt(getEnv("DB_CONN_MAX_LIFETIME", "5"), 5), // minutes
			ConnMaxIdleTime: parseInt(getEnv("DB_CONN_MAX_IDLE_TIME", "1"), 1), // minutes
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
			Environment:    getEnv("APP_ENV", "development"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		RateLimit: RateLimitConfig{
			DefaultRequestsPerSecond: parseFloat(getEnv("RATE_LIMIT_DEFAULT_RPS", "10"), 10),
			DefaultBurstSize:         parseInt(getEnv("RATE_LIMIT_DEFAULT_BURST", "20"), 20),
			AuthRequestsPerSecond:    parseFloat(getEnv("RATE_LIMIT_AUTH_RPS", "2"), 2),
			AuthBurstSize:            parseInt(getEnv("RATE_LIMIT_AUTH_BURST", "5"), 5),
		},
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseFloat parses a string to float64 with fallback
func parseFloat(str string, fallback float64) float64 {
	if value, err := strconv.ParseFloat(str, 64); err == nil {
		return value
	}
	return fallback
}

// parseInt parses a string to int with fallback
func parseInt(str string, fallback int) int {
	if value, err := strconv.Atoi(str); err == nil {
		return value
	}
	return fallback
}
