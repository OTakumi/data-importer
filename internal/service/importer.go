package service

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
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
	ImportPath(path string) (any, error)
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
func NewMongoImporterWithOptions(ctx context.Context, fileUtils utils.FileUtilsInterface, repo DocumentRepository, batchSize int, removeIDField bool) *MongoImporter {
	// Use a reasonable default batch size if not specified
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &MongoImporter{
		fileUtils:     fileUtils,
		repo:          repo,
		batchSize:     batchSize,
		ctx:           ctx,
		removeIDField: removeIDField,
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
	if !m.removeIDField {
		return documents
	}

	idCount := 0
	for i := range documents {
		if _, hasID := documents[i]["_id"]; hasID {
			idCount++

			// _idの種類も確認
			// fmt.Printf("Document %d has _id of type %T: %v\n",
			// 	i, documents[i]["_id"], documents[i]["_id"])

			delete(documents[i], "_id")

			// 削除後に確認
			// if _, stillHasID := documents[i]["_id"]; stillHasID {
			// 	fmt.Printf("WARNING: Document %d still has _id after deletion!\n", i)
			// }
		}

		// 各フィールドを再帰的に処理して日付を変換
		documents[i] = m.processDocumentDates(documents[i])
	}

	return documents
}

// processDocumentDates recursively processes all fields in a document
// converting date strings and MongoDB's $date format to time.Time objects
func (m *MongoImporter) processDocumentDates(doc domain.Document) domain.Document {
	for key, value := range doc {
		switch v := value.(type) {
		case map[string]interface{}:
			// $dateフィールドを持つオブジェクトをチェック
			if dateStr, ok := v["$date"]; ok {
				if ds, ok := dateStr.(string); ok {
					// 日付文字列をtime.Time型に変換
					t, err := parseDateTime(ds)
					if err == nil {
						// time.Time型をセット (MongoDB ドライバーが自動的に日付型として扱う)
						doc[key] = t
					} else {
						fmt.Printf("Warning: Failed to parse date string '%s': %v\n", ds, err)
					}
				}
			} else {
				// ネストされたマップを再帰的に処理
				doc[key] = m.processDocumentDates(v)
			}
		case string:
			// 文字列が日付形式かチェック
			if isDateString(v) {
				t, err := parseDateTime(v)
				if err == nil {
					doc[key] = t
				}
			}
		}
	}
	return doc
}

// parseDateTime parses a date string in various formats
func parseDateTime(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// isDateString checks if a string is in ISO date format
func isDateString(s string) bool {
	// ISO 8601形式の日付文字列をチェック
	matched, _ := regexp.MatchString(
		`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
		s)

	return matched
}
