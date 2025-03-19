package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	MongoURI       string
	DatabaseName   string
	TimeoutSeconds int
	BatchSize      int
}

// LoadEnv loads environment variables from .env file if it exists
func LoadEnv() error {
	// Load .env file if it exists
	// If file doesn't exist, godotenv returns an error but we can ignore it
	// as we'll fall back to OS environment variables
	if err := godotenv.Load(); err != nil {
		// Just log or print the error for debugging, but continue
		// fmt.Printf("No .env file found or error loading: %v\n", err)
		return nil
	}
	return nil
}

// NewConfig creates a new Config from environment variables or defaults
func NewConfig() *Config {
	// Try to load .env file (ignore if not exists)
	_ = LoadEnv()

	// Parse timeout seconds
	timeoutStr := getEnv("MONGODB_TIMEOUT", "10")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 10 // Default if parsing fails
	}

	// Parse batch size
	batchSizeStr := getEnv("MONGODB_BATCH_SIZE", "1000")
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil || batchSize <= 0 {
		batchSize = 1000 // Default if parsing fails
	}

	return &Config{
		MongoURI:       getEnv("MONGODB_URI", "mongodb://mongodb:27017"),
		DatabaseName:   getEnv("MONGODB_DATABASE", "test_db"),
		TimeoutSeconds: timeout,
		BatchSize:      batchSize,
	}
}

// getEnv gets an environment variable value or returns a default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
