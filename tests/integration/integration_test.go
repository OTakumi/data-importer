package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OTakumi/data-importer/internal/config"
	"github.com/OTakumi/data-importer/internal/repository"
	"github.com/OTakumi/data-importer/internal/service"
	"github.com/OTakumi/data-importer/internal/utils"
)

// TestIntegration is an integration test that tests the full import process
// This test requires a MongoDB instance to be available
// You can configure the MongoDB connection by setting environment variables or creating a .env.test file
func TestIntegration(t *testing.T) {
	// Load test specific .env file if it exists
	testEnvFilePath := ".env.test"
	if _, err := os.Stat(testEnvFilePath); err == nil {
		os.Setenv("DOTENV_PATH", testEnvFilePath)
	}

	// For CI/CD environments, allow setting a test-specific MongoDB URI
	// if testURI := os.Getenv("TEST_MONGODB_URI"); testURI != "" {
	// 	os.Setenv("MONGODB_URI", testURI)
	// }

	// Get test data paths
	testDataDir := findTestDataDir(t)
	usersArrayPath := filepath.Join(testDataDir, "users_array.json")
	productObjectPath := filepath.Join(testDataDir, "product_object.json")
	invalidJSONPath := filepath.Join(testDataDir, "invalid.json")

	// Verify that test data files exist
	if _, err := os.Stat(usersArrayPath); os.IsNotExist(err) {
		t.Skipf("Test data file %s not found. Skipping test.", usersArrayPath)
	}
	if _, err := os.Stat(productObjectPath); os.IsNotExist(err) {
		t.Skipf("Test data file %s not found. Skipping test.", productObjectPath)
	}

	// Initialize config (this will load from .env.test or environment variables)
	cfg := config.NewConfig()

	// Override database name for tests to avoid affecting production data
	testDBName := "test_db_integration"
	if os.Getenv("TEST_MONGODB_DATABASE") != "" {
		testDBName = os.Getenv("TEST_MONGODB_DATABASE")
	}
	cfg.DatabaseName = testDBName

	t.Logf("Using MongoDB URI: %s, Database: %s", cfg.MongoURI, cfg.DatabaseName)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Initialize MongoDB repository
	repo, err := repository.NewMongoRepository(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Disconnect(context.Background()); err != nil {
			t.Logf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize file utilities
	fileUtils := utils.NewFileUtils(nil) // Use actual file system

	// Initialize importer service with batch size from config
	importer := service.NewMongoImporter(ctx, fileUtils, repo, cfg.BatchSize)

	// Run subtests
	t.Run("ImportArrayJSON", func(t *testing.T) {
		testImportArrayJSON(t, importer, usersArrayPath)
	})

	t.Run("ImportObjectJSON", func(t *testing.T) {
		testImportObjectJSON(t, importer, productObjectPath)
	})

	t.Run("ImportInvalidJSON", func(t *testing.T) {
		testImportInvalidJSON(t, importer, invalidJSONPath)
	})

	t.Run("ImportDirectory", func(t *testing.T) {
		testImportDirectory(t, importer, testDataDir)
	})
}

// findTestDataDir searches for the testdata directory
func findTestDataDir(t *testing.T) string {
	// Try different possible locations for the test data
	possiblePaths := []string{
		"testdata",          // If run from project root
		"../../testdata",    // If run from tests/integration
		"../testdata",       // If run from tests
		"../../../testdata", // Other possible location
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			t.Logf("Found test data directory: %s", absPath)
			return path
		}
	}

	// If testdata directory is not found, create a temporary one
	tempDir, err := os.MkdirTemp("", "testdata-*")
	if err != nil {
		t.Fatalf("Failed to create temporary test data directory: %v", err)
	}
	t.Logf("Created temporary test data directory: %s", tempDir)

	// This is not ideal, but allows tests to continue
	// In a real scenario, you would populate this directory with test files
	t.Logf("Warning: Using temporary test data directory. Tests may be skipped.")
	return tempDir
}

// testImportArrayJSON tests importing an array format JSON file
func testImportArrayJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err != nil {
		t.Fatalf("Failed to import array JSON: %v", err)
	}

	if result.InsertedCount <= 0 {
		t.Errorf("No documents were inserted: %+v", result)
	}

	t.Logf("Array JSON import result: %d documents inserted (collection: %s)",
		result.InsertedCount, result.CollectionName)
}

// testImportObjectJSON tests importing a single object JSON file
func testImportObjectJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err != nil {
		t.Fatalf("Failed to import object JSON: %v", err)
	}

	if result.InsertedCount != 1 {
		t.Errorf("Expected 1 document to be inserted, but got %d: %+v",
			result.InsertedCount, result)
	}

	t.Logf("Object JSON import result: %d documents inserted (collection: %s)",
		result.InsertedCount, result.CollectionName)
}

// testImportInvalidJSON tests importing an invalid JSON file
func testImportInvalidJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err == nil {
		t.Errorf("Import of invalid JSON file succeeded unexpectedly: %+v", result)
	}

	t.Logf("Invalid JSON file import failed as expected: %v", err)
}

// testImportDirectory tests importing a directory
func testImportDirectory(t *testing.T, importer *service.MongoImporter, dirPath string) {
	results, err := importer.ImportDirectory(dirPath)
	if err != nil {
		// Errors are expected (invalid JSON file is included)
		t.Logf("Directory import had partial errors: %v", err)
	}

	// Check results
	var successCount, errorCount int
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	t.Logf("Directory import results: %d succeeded, %d failed, total %d files",
		successCount, errorCount, len(results))

	// Ensure at least one success
	if successCount == 0 {
		t.Errorf("No files were successfully imported in directory import")
	}
}
