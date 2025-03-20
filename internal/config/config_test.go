package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Clear all environment variables first
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGODB_TIMEOUT")
	os.Unsetenv("MONGODB_BATCH_SIZE")

	// Test default values
	config := NewConfig()
	if config.MongoURI != "mongodb://mongodb:27017" {
		t.Errorf("Expected default MongoURI to be 'mongodb://mongodb:27017', got '%s'", config.MongoURI)
	}
	if config.DatabaseName != "test_db" {
		t.Errorf("Expected default DatabaseName to be 'test_db', got '%s'", config.DatabaseName)
	}
	if config.TimeoutSeconds != 10 {
		t.Errorf("Expected default TimeoutSeconds to be 10, got %d", config.TimeoutSeconds)
	}
	if config.BatchSize != 1000 {
		t.Errorf("Expected default BatchSize to be 1000, got %d", config.BatchSize)
	}

	// Test environment variables
	os.Setenv("MONGODB_URI", "mongodb://custom:27017")
	os.Setenv("MONGODB_DATABASE", "custom_db")
	os.Setenv("MONGODB_TIMEOUT", "30")
	os.Setenv("MONGODB_BATCH_SIZE", "500")

	config = NewConfig()
	if config.MongoURI != "mongodb://custom:27017" {
		t.Errorf("Expected MongoURI to be 'mongodb://custom:27017', got '%s'", config.MongoURI)
	}
	if config.DatabaseName != "custom_db" {
		t.Errorf("Expected DatabaseName to be 'custom_db', got '%s'", config.DatabaseName)
	}
	if config.TimeoutSeconds != 30 {
		t.Errorf("Expected TimeoutSeconds to be 30, got %d", config.TimeoutSeconds)
	}
	if config.BatchSize != 500 {
		t.Errorf("Expected BatchSize to be 500, got %d", config.BatchSize)
	}

	// Test invalid timeout value
	os.Setenv("MONGODB_TIMEOUT", "invalid")
	config = NewConfig()
	if config.TimeoutSeconds != 10 {
		t.Errorf("Expected TimeoutSeconds to be default 10 when invalid value, got %d", config.TimeoutSeconds)
	}

	// Test invalid batch size value
	os.Setenv("MONGODB_TIMEOUT", "30") // Reset timeout to valid value
	os.Setenv("MONGODB_BATCH_SIZE", "invalid")
	config = NewConfig()
	if config.BatchSize != 1000 {
		t.Errorf("Expected BatchSize to be default 1000 when invalid value, got %d", config.BatchSize)
	}

	// Test negative batch size value
	os.Setenv("MONGODB_BATCH_SIZE", "-100")
	config = NewConfig()
	if config.BatchSize != 1000 {
		t.Errorf("Expected BatchSize to be default 1000 when negative value, got %d", config.BatchSize)
	}

	// Clean up
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGODB_TIMEOUT")
	os.Unsetenv("MONGODB_BATCH_SIZE")
}

func TestGetEnv(t *testing.T) {
	// Test when environment variable is not set
	os.Unsetenv("TEST_ENV")
	value := getEnv("TEST_ENV", "default_value")
	if value != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%s'", value)
	}

	// Test when environment variable is set
	os.Setenv("TEST_ENV", "test_value")
	value = getEnv("TEST_ENV", "default_value")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Clean up
	os.Unsetenv("TEST_ENV")
}

func TestLoadEnv(t *testing.T) {
	// Create a temporary directory for test .env file
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test .env file
	envFilePath := filepath.Join(tempDir, ".env")
	envContent := `
MONGODB_URI=mongodb://envfile:27017
MONGODB_DATABASE=envfile_db
MONGODB_TIMEOUT=20
MONGODB_BATCH_SIZE=250
`
	err = os.WriteFile(envFilePath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Change working directory to temp directory to test .env loading
	originalDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Clear environment variables
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGODB_TIMEOUT")
	os.Unsetenv("MONGODB_BATCH_SIZE")

	// Load .env file
	err = LoadEnv()
	if err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	// Create config and check values
	config := NewConfig()

	// Check if values from .env file were loaded
	if config.MongoURI != "mongodb://envfile:27017" {
		t.Errorf("Expected MongoURI from .env to be 'mongodb://envfile:27017', got '%s'", config.MongoURI)
	}
	if config.DatabaseName != "envfile_db" {
		t.Errorf("Expected DatabaseName from .env to be 'envfile_db', got '%s'", config.DatabaseName)
	}
	if config.TimeoutSeconds != 20 {
		t.Errorf("Expected TimeoutSeconds from .env to be 20, got %d", config.TimeoutSeconds)
	}
	if config.BatchSize != 250 {
		t.Errorf("Expected BatchSize from .env to be 250, got %d", config.BatchSize)
	}

	// Test non-existent .env file
	os.Remove(envFilePath)
	err = LoadEnv()
	if err != nil {
		t.Errorf("Expected no error for non-existent .env file, got %v", err)
	}

	// Clean up
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGODB_TIMEOUT")
	os.Unsetenv("MONGODB_BATCH_SIZE")
}

func TestBuildMongoURI(t *testing.T) {
	// Clear all MongoDB related environment variables first
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_USERNAME")
	os.Unsetenv("MONGODB_PASSWORD")
	os.Unsetenv("MONGODB_HOST")
	os.Unsetenv("MONGODB_PORT")
	os.Unsetenv("MONGODB_AUTH_DATABASE")
	os.Unsetenv("MONGODB_REPLICA_SET")

	// Test 1: Default values (no environment variables set)
	uri := BuildMongoURI()
	expected := "mongodb://mongodb:27017"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 2: Custom host and port
	os.Setenv("MONGODB_HOST", "custom-host")
	os.Setenv("MONGODB_PORT", "12345")
	uri = BuildMongoURI()
	expected = "mongodb://custom-host:12345"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 3: With username and password
	os.Setenv("MONGODB_USERNAME", "testuser")
	os.Setenv("MONGODB_PASSWORD", "testpass")
	uri = BuildMongoURI()
	expected = "mongodb://testuser:testpass@custom-host:12345"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 4: With special characters in password
	os.Setenv("MONGODB_PASSWORD", "test@pass:123")
	uri = BuildMongoURI()
	expected = "mongodb://testuser:test%40pass%3A123@custom-host:12345"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 5: With authentication database
	os.Setenv("MONGODB_PASSWORD", "testpass") // Reset to simple password
	os.Setenv("MONGODB_AUTH_DATABASE", "admin")
	uri = BuildMongoURI()
	expected = "mongodb://testuser:testpass@custom-host:12345/?authSource=admin"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 6: With replica set
	os.Setenv("MONGODB_REPLICA_SET", "rs0")
	uri = BuildMongoURI()
	expected = "mongodb://testuser:testpass@custom-host:12345/?authSource=admin&replicaSet=rs0"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 7: Just replica set (no auth db)
	os.Unsetenv("MONGODB_AUTH_DATABASE")
	uri = BuildMongoURI()
	expected = "mongodb://testuser:testpass@custom-host:12345/?replicaSet=rs0"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 8: Username without password
	os.Unsetenv("MONGODB_PASSWORD")
	os.Unsetenv("MONGODB_REPLICA_SET")
	uri = BuildMongoURI()
	expected = "mongodb://testuser@custom-host:12345"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Test 9: Full URI takes precedence
	os.Setenv("MONGODB_URI", "mongodb://direct-uri:27017")
	uri = BuildMongoURI()
	expected = "mongodb://direct-uri:27017"
	if uri != expected {
		t.Errorf("Expected URI '%s', got '%s'", expected, uri)
	}

	// Clean up
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_USERNAME")
	os.Unsetenv("MONGODB_PASSWORD")
	os.Unsetenv("MONGODB_HOST")
	os.Unsetenv("MONGODB_PORT")
	os.Unsetenv("MONGODB_AUTH_DATABASE")
	os.Unsetenv("MONGODB_REPLICA_SET")
}
