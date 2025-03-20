package service

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/OTakumi/data-importer/internal/domain"
	"github.com/OTakumi/data-importer/internal/utils"
)

// ImporterService defines the interface for the importer service
type ImporterService interface {
	// ImportFile imports a single JSON file to MongoDB
	ImportFile(filePath string) (*domain.ImportResult, error)

	// ImportDirectory imports all JSON files in a directory to MongoDB
	ImportDirectory(dirPath string) ([]*domain.ImportResult, error)

	// ImportPath determines if the path is a file or directory and processes accordingly
	ImportPath(path string) (interface{}, error)
}

// DocumentRepository defines the interface for MongoDB operations
// This interface matches the existing Repository interface in the repository package
type DocumentRepository interface {
	// InsertDocuments inserts multiple documents into a collection
	InsertDocuments(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error)
	// Disconnect closes the connection to MongoDB
	Disconnect(ctx context.Context) error
}

// MongoImporter implements the ImporterService interface
type MongoImporter struct {
	fileUtils     utils.FileUtilsInterface // For file operations (インターフェースに変更)
	repo          DocumentRepository       // For MongoDB operations
	batchSize     int                      // Batch size for document imports
	ctx           context.Context          // Context for database operations
	removeIDField bool                     // Whether to remove _id fields during import
}

// NewMongoImporter creates a new MongoDB importer service
func NewMongoImporter(ctx context.Context, fileUtils utils.FileUtilsInterface, repo DocumentRepository, batchSize int) *MongoImporter {
	// Use a reasonable default batch size if not specified
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &MongoImporter{
		fileUtils: fileUtils,
		repo:      repo,
		batchSize: batchSize,
		ctx:       ctx,
	}
}

// ImportPath determines if the path is a file or directory and processes accordingly
func (m *MongoImporter) ImportPath(path string) (any, error) {
	// Check if path is a directory or file
	isDir, err := m.fileUtils.IsDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("error checking path %s: %w", path, err)
	}

	if isDir {
		return m.ImportDirectory(path)
	}

	return m.ImportFile(path)
}

// ImportFile imports a single JSON file to MongoDB
func (m *MongoImporter) ImportFile(filePath string) (*domain.ImportResult, error) {
	startTime := time.Now()
	result := &domain.ImportResult{
		FileName:       filepath.Base(filePath),
		CollectionName: utils.FilePathToCollectionName(filePath),
	}

	// Parse JSON file
	documents, err := m.fileUtils.ParseJSONFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("error parsing file %s: %w", filePath, err)
		return result, result.Error
	}

	// Convert to domain models
	var domainDocs []domain.Document
	for _, doc := range documents {
		// Use direct conversion since domain.Document is a map type alias
		domainDocs = append(domainDocs, domain.Document(doc))
	}

	// Clean documents by removing _id fields before import
	domainDocs = m.cleanDocuments(domainDocs)

	// Import documents in batches
	count, err := m.processBatches(domainDocs, result.CollectionName)
	if err != nil {
		result.Error = fmt.Errorf("error importing documents to collection %s: %w", result.CollectionName, err)
		return result, result.Error
	}

	// Update result
	result.InsertedCount = count
	result.Duration = time.Since(startTime)

	return result, nil
}

// ImportDirectory imports all JSON files in a directory to MongoDB
func (m *MongoImporter) ImportDirectory(dirPath string) ([]*domain.ImportResult, error) {
	// Find all JSON files in the directory
	jsonFiles, err := m.fileUtils.FindJSONFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error finding JSON files in directory %s: %w", dirPath, err)
	}

	if len(jsonFiles) == 0 {
		return nil, fmt.Errorf("no JSON files found in directory %s", dirPath)
	}

	// Process each file in parallel
	var wg sync.WaitGroup
	resultChan := make(chan *domain.ImportResult, len(jsonFiles))

	for _, file := range jsonFiles {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			result, _ := m.ImportFile(filePath)
			resultChan <- result
		}(file)
	}

	// Wait for all imports to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	var results []*domain.ImportResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Check if any imports failed
	var importErrors []error
	for _, result := range results {
		if result.Error != nil {
			importErrors = append(importErrors, result.Error)
		}
	}

	if len(importErrors) > 0 {
		// Return partial results with an error indicating some imports failed
		return results, fmt.Errorf("%d out of %d files failed to import", len(importErrors), len(jsonFiles))
	}

	return results, nil
}

// processBatches processes a slice of documents in batches
func (m *MongoImporter) processBatches(documents []domain.Document, collectionName string) (int, error) {
	// Call InsertDocuments and use the result
	result, err := m.repo.InsertDocuments(m.ctx, collectionName, documents)
	if err != nil {
		return 0, err
	}

	return result.InsertedCount, nil
}

// cleanDocuments removes _id fields from all documents to prevent MongoDB import errors
func (m *MongoImporter) cleanDocuments(documents []domain.Document) []domain.Document {
	// If removeIDField is false, return documents unchanged
	if !m.removeIDField {
		return documents
	}

	for i := range documents {
		// Remove the _id field
		delete(documents[i], "_id")

		// Optional: Process date fields with $date format
		for key, value := range documents[i] {
			if valueMap, ok := value.(map[string]any); ok {
				if dateStr, hasDate := valueMap["$date"]; hasDate {
					// Convert $date format to a regular date string
					if dateString, ok := dateStr.(string); ok {
						documents[i][key] = dateString
					}
				}
			}
		}
	}
	return documents
}
