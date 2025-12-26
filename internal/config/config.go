package config

import (
	"os"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    string
	BaseURL string

	// Database
	DatabasePath string

	// S3 Storage
	S3BucketName string
	S3Endpoint   string
	S3Region     string

	// Dispatcher
	DispatcherMode string
	WorkerURL      string
	CallbackURL    string

	// Livestream
	RTMPPort    int
	StreamsPath string

	// CORS
	AllowedOrigins []string

	// JWT
	JWTSecret string
	JWTExpiry time.Duration
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		// Server
		Port:    getEnv("PORT", "8080"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),

		// Database
		DatabasePath: getEnv("DATABASE_PATH", "./data/katie.db"),

		// S3
		S3BucketName: os.Getenv("S3_BUCKET_NAME"),
		S3Endpoint:   os.Getenv("S3_ENDPOINT"),
		S3Region:     getEnv("S3_REGION", "auto"),

		// Dispatcher
		DispatcherMode: getEnv("DISPATCHER_MODE", "local"),
		WorkerURL:      os.Getenv("WORKER_URL"),
		CallbackURL:    os.Getenv("CALLBACK_URL"),

		// Livestream
		RTMPPort:    1935,
		StreamsPath: getEnv("STREAMS_PATH", "./data/streams"),

		// CORS
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "loll-nice-haha"),
		JWTExpiry: parseDuration(getEnv("JWT_EXPIRY", "24h")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
