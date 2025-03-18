package service

import (
	"context"
	"errors"
	"testing"
	// "time"

	"github.com/OTakumi/data-importer/internal/domain"
	// "github.com/OTakumi/data-importer/internal/utils"
)

// MockFileUtils is a mock implementation of the file utilities for testing
type MockFileUtils struct {
	IsDirectoryFunc   func(path string) (bool, error)
	FindJSONFilesFunc func(dirPath string) ([]string, error)
	ParseJSONFileFunc func(filePath string) ([]map[string]interface{}, error)
}

// IsDirectory mocks the IsDirectory method
func (m *MockFileUtils) IsDirectory(path string) (bool, error) {
	return m.IsDirectoryFunc(path)
}

// FindJSONFiles mocks the FindJSONFiles method
func (m *MockFileUtils) FindJSONFiles(dirPath string) ([]string, error) {
	return m.FindJSONFilesFunc(dirPath)
}

// ParseJSONFile mocks the ParseJSONFile method
func (m *MockFileUtils) ParseJSONFile(filePath string) ([]map[string]interface{}, error) {
	return m.ParseJSONFileFunc(filePath)
}

// MockRepository is a mock implementation of the document repository for testing
type MockRepository struct {
	InsertDocumentsFunc func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error)
	DisconnectFunc      func(ctx context.Context) error
}

// InsertDocuments mocks the InsertDocuments method
func (m *MockRepository) InsertDocuments(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
	return m.InsertDocumentsFunc(ctx, collectionName, documents)
}

// Disconnect mocks the Disconnect method
func (m *MockRepository) Disconnect(ctx context.Context) error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc(ctx)
	}
	return nil
}

// TestImportFile tests the ImportFile method
func TestImportFile(t *testing.T) {
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name           string
		filePath       string
		mockFileUtils  *MockFileUtils
		mockRepo       *MockRepository
		expectedResult *domain.ImportResult
		expectError    bool
	}{
		{
			name:     "Successful import",
			filePath: "/data/users.json",
			mockFileUtils: &MockFileUtils{
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					return []map[string]interface{}{
						{"name": "John", "age": 30},
						{"name": "Jane", "age": 25},
					}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					if collectionName != "users" {
						t.Errorf("Expected collection name to be 'users', got '%s'", collectionName)
					}
					if len(documents) != 2 {
						t.Errorf("Expected 2 documents, got %d", len(documents))
					}
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(documents),
						Error:          nil,
					}, nil
				},
			},
			expectedResult: &domain.ImportResult{
				FileName:       "users.json",
				CollectionName: "users",
				InsertedCount:  2,
				Error:          nil,
			},
			expectError: false,
		},
		{
			name:     "Parse error",
			filePath: "/data/invalid.json",
			mockFileUtils: &MockFileUtils{
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					return nil, errors.New("parse error")
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					t.Error("InsertDocuments should not be called when parsing fails")
					return nil, nil
				},
			},
			expectedResult: &domain.ImportResult{
				FileName:       "invalid.json",
				CollectionName: "invalid",
				InsertedCount:  0,
				Error:          errors.New("error parsing file /data/invalid.json: parse error"),
			},
			expectError: true,
		},
		{
			name:     "Database error",
			filePath: "/data/users.json",
			mockFileUtils: &MockFileUtils{
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					return []map[string]interface{}{
						{"name": "John", "age": 30},
					}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					return nil, errors.New("database error")
				},
			},
			expectedResult: &domain.ImportResult{
				FileName:       "users.json",
				CollectionName: "users",
				InsertedCount:  0,
				Error:          errors.New("error importing documents to collection users: database error"),
			},
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with mocks
			importer := NewMongoImporter(ctx, tt.mockFileUtils, tt.mockRepo, 100)

			// Call the method
			result, err := importer.ImportFile(tt.filePath)

			// Check the error
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result properties
			if result.FileName != tt.expectedResult.FileName {
				t.Errorf("Expected filename %s, got %s", tt.expectedResult.FileName, result.FileName)
			}
			if result.CollectionName != tt.expectedResult.CollectionName {
				t.Errorf("Expected collection name %s, got %s", tt.expectedResult.CollectionName, result.CollectionName)
			}
			if result.InsertedCount != tt.expectedResult.InsertedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedResult.InsertedCount, result.InsertedCount)
			}

			// Check error messages if applicable
			if (result.Error != nil && tt.expectedResult.Error != nil) &&
				result.Error.Error() != tt.expectedResult.Error.Error() {
				t.Errorf("Expected error message %s, got %s", tt.expectedResult.Error.Error(), result.Error.Error())
			}
		})
	}
}

// TestImportDirectory tests the ImportDirectory method
func TestImportDirectory(t *testing.T) {
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name            string
		dirPath         string
		mockFileUtils   *MockFileUtils
		mockRepo        *MockRepository
		expectedResults []*domain.ImportResult
		expectError     bool
	}{
		{
			name:    "Successful directory import",
			dirPath: "/data",
			mockFileUtils: &MockFileUtils{
				FindJSONFilesFunc: func(dirPath string) ([]string, error) {
					return []string{"/data/users.json", "/data/products.json"}, nil
				},
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					if filePath == "/data/users.json" {
						return []map[string]interface{}{
							{"name": "User1"},
							{"name": "User2"},
						}, nil
					}
					return []map[string]interface{}{
						{"id": 1, "name": "Product1"},
						{"id": 2, "name": "Product2"},
						{"id": 3, "name": "Product3"},
					}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(documents),
						Error:          nil,
					}, nil
				},
			},
			expectedResults: []*domain.ImportResult{
				{
					FileName:       "users.json",
					CollectionName: "users",
					InsertedCount:  2,
					Error:          nil,
				},
				{
					FileName:       "products.json",
					CollectionName: "products",
					InsertedCount:  3,
					Error:          nil,
				},
			},
			expectError: false,
		},
		{
			name:    "No JSON files found",
			dirPath: "/empty",
			mockFileUtils: &MockFileUtils{
				FindJSONFilesFunc: func(dirPath string) ([]string, error) {
					return []string{}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					t.Error("InsertDocuments should not be called when no files are found")
					return nil, nil
				},
			},
			expectedResults: nil,
			expectError:     true,
		},
		{
			name:    "Error finding JSON files",
			dirPath: "/invalid",
			mockFileUtils: &MockFileUtils{
				FindJSONFilesFunc: func(dirPath string) ([]string, error) {
					return nil, errors.New("directory error")
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					t.Error("InsertDocuments should not be called when finding files fails")
					return nil, nil
				},
			},
			expectedResults: nil,
			expectError:     true,
		},
		{
			name:    "Partial success with some failures",
			dirPath: "/mixed",
			mockFileUtils: &MockFileUtils{
				FindJSONFilesFunc: func(dirPath string) ([]string, error) {
					return []string{"/mixed/valid.json", "/mixed/invalid.json"}, nil
				},
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					if filePath == "/mixed/valid.json" {
						return []map[string]interface{}{{"valid": true}}, nil
					}
					return nil, errors.New("parse error")
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(documents),
						Error:          nil,
					}, nil
				},
			},
			expectedResults: []*domain.ImportResult{
				{
					FileName:       "valid.json",
					CollectionName: "valid",
					InsertedCount:  1,
					Error:          nil,
				},
				{
					FileName:       "invalid.json",
					CollectionName: "invalid",
					InsertedCount:  0,
					Error:          errors.New("error parsing file /mixed/invalid.json: parse error"),
				},
			},
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with mocks
			importer := NewMongoImporter(ctx, tt.mockFileUtils, tt.mockRepo, 100)

			// Call the method
			results, err := importer.ImportDirectory(tt.dirPath)

			// Check the error
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For the "No JSON files found" case, we expect no results
			if tt.name == "No JSON files found" || tt.name == "Error finding JSON files" {
				if results != nil {
					t.Errorf("Expected nil results, got %v", results)
				}
				return
			}

			// Check results count
			if len(results) != len(tt.expectedResults) {
				t.Errorf("Expected %d results, got %d", len(tt.expectedResults), len(results))
				return
			}

			// We can't guarantee the order of results due to parallel processing,
			// so we'll just check the total count of documents
			totalActualCount := 0
			totalExpectedCount := 0
			for _, r := range results {
				totalActualCount += r.InsertedCount
			}
			for _, r := range tt.expectedResults {
				totalExpectedCount += r.InsertedCount
			}

			if totalActualCount != totalExpectedCount {
				t.Errorf("Expected total document count %d, got %d", totalExpectedCount, totalActualCount)
			}
		})
	}
}

// TestImportPath tests the ImportPath method
func TestImportPath(t *testing.T) {
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		path          string
		mockFileUtils *MockFileUtils
		mockRepo      *MockRepository
		expectError   bool
		expectDir     bool
	}{
		{
			name: "Path is a file",
			path: "/data/users.json",
			mockFileUtils: &MockFileUtils{
				IsDirectoryFunc: func(path string) (bool, error) {
					return false, nil
				},
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					return []map[string]interface{}{{"name": "User1"}}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(documents),
						Error:          nil,
					}, nil
				},
			},
			expectError: false,
			expectDir:   false,
		},
		{
			name: "Path is a directory",
			path: "/data",
			mockFileUtils: &MockFileUtils{
				IsDirectoryFunc: func(path string) (bool, error) {
					return true, nil
				},
				FindJSONFilesFunc: func(dirPath string) ([]string, error) {
					return []string{"/data/file1.json"}, nil
				},
				ParseJSONFileFunc: func(filePath string) ([]map[string]interface{}, error) {
					return []map[string]interface{}{{"key": "value"}}, nil
				},
			},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(documents),
						Error:          nil,
					}, nil
				},
			},
			expectError: false,
			expectDir:   true,
		},
		{
			name: "Error checking path",
			path: "/invalid/path",
			mockFileUtils: &MockFileUtils{
				IsDirectoryFunc: func(path string) (bool, error) {
					return false, errors.New("path error")
				},
			},
			mockRepo:    &MockRepository{},
			expectError: true,
			expectDir:   false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with mocks
			importer := NewMongoImporter(ctx, tt.mockFileUtils, tt.mockRepo, 100)

			// Call the method
			result, err := importer.ImportPath(tt.path)

			// Check the error
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Skip further checks if error is expected
			if tt.expectError {
				return
			}

			// Check result type
			if tt.expectDir {
				// Result should be a slice of ImportResult
				if _, ok := result.([]*domain.ImportResult); !ok {
					t.Errorf("Expected result to be []*ImportResult, got %T", result)
				}
			} else {
				// Result should be a single ImportResult
				if _, ok := result.(*domain.ImportResult); !ok {
					t.Errorf("Expected result to be *ImportResult, got %T", result)
				}
			}
		})
	}
}

// TestProcessBatches tests the processBatches method
func TestProcessBatches(t *testing.T) {
	ctx := context.Background()

	// Create test data with 250 documents
	documents := make([]domain.Document, 250)
	for i := 0; i < 250; i++ {
		// Use map conversion since domain.Document is a map type alias
		documents[i] = domain.Document{"id": i, "name": "Test"}
	}

	// Test cases
	tests := []struct {
		name          string
		batchSize     int
		documents     []domain.Document
		mockRepo      *MockRepository
		expectedCount int
		expectError   bool
	}{
		{
			name:      "Process all documents successfully",
			batchSize: 100,
			documents: documents,
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, docs []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  len(docs),
						Error:          nil,
					}, nil
				},
			},
			expectedCount: 250,
			expectError:   false,
		},
		{
			name:      "Database error",
			batchSize: 100,
			documents: documents,
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, docs []domain.Document) (*domain.ImportResult, error) {
					return nil, errors.New("database error")
				},
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:      "Empty document list",
			batchSize: 100,
			documents: []domain.Document{},
			mockRepo: &MockRepository{
				InsertDocumentsFunc: func(ctx context.Context, collectionName string, docs []domain.Document) (*domain.ImportResult, error) {
					return &domain.ImportResult{
						CollectionName: collectionName,
						InsertedCount:  0,
						Error:          nil,
					}, nil
				},
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with mocks
			importer := NewMongoImporter(ctx, &MockFileUtils{}, tt.mockRepo, tt.batchSize)

			// Call the method directly (it's private, but we can access it in tests)
			count, err := importer.processBatches(tt.documents, "test_collection")

			// Check the error
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check count
			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}
