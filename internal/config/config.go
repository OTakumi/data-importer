package config

import (
	"fmt"
	"net/url"
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
	// Check for custom .env file path
	envPath := os.Getenv("DOTENV_PATH")
	if envPath == "" {
		envPath = ".env"
	}

	// Load .env file if it exists
	// If file doesn't exist, godotenv returns an error but we can ignore it
	// as we'll fall back to OS environment variables
	if err := godotenv.Load(envPath); err != nil {
		// Just log or print the error for debugging, but continue
		fmt.Printf("No .env file found or error loading: %v\n", err)
		return nil
	}
	return nil
}

// BuildMongoURI builds MongoDB connection URI from individual components
func BuildMongoURI() string {
	// Check if MONGODB_URI is explicitly provided
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		return uri
	}

	// Get individual components
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")
	host := getEnv("MONGODB_HOST", "mongodb")
	port := getEnv("MONGODB_PORT", "27017")

	// Default format: mongodb://host:port
	mongoURI := fmt.Sprintf("mongodb://%s:%s", host, port)

	// If authentication is provided, add username and password
	if username != "" {
		// URL encode the password to handle special characters
		encodedPassword := url.QueryEscape(password)

		if encodedPassword != "" {
			// Format with authentication: mongodb://username:password@host:port
			mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s",
				username, encodedPassword, host, port)
		} else {
			// Username without password
			mongoURI = fmt.Sprintf("mongodb://%s@%s:%s",
				username, host, port)
		}
	}

	// Add authentication database if specified
	if authDB := os.Getenv("MONGODB_AUTH_DATABASE"); authDB != "" {
		mongoURI = fmt.Sprintf("%s/?authSource=%s", mongoURI, authDB)
	}

	// Add replica set name if specified
	if replicaSet := os.Getenv("MONGODB_REPLICA_SET"); replicaSet != "" {
		// Check if we already have query parameters
		if authDB := os.Getenv("MONGODB_AUTH_DATABASE"); authDB != "" {
			mongoURI = fmt.Sprintf("%s&replicaSet=%s", mongoURI, replicaSet)
		} else {
			mongoURI = fmt.Sprintf("%s/?replicaSet=%s", mongoURI, replicaSet)
		}
	}

	return mongoURI
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
		MongoURI:       BuildMongoURI(),
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
