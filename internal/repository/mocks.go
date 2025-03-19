/// Description: テスト用のモック

package repository

import (
	"context"

	"github.com/OTakumi/data-importer/internal/domain"
)

// MockMongoRepository はMongoRepositoryのモック実装です
// テスト用途に使用されます
type MockMongoRepository struct {
	InsertDocumentsFn func(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error)
	DisconnectFn      func(ctx context.Context) error
}

// InsertDocuments はInsertDocumentsのモック実装です
func (m *MockMongoRepository) InsertDocuments(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
	if m.InsertDocumentsFn != nil {
		return m.InsertDocumentsFn(ctx, collectionName, documents)
	}
	// デフォルトの実装
	return &domain.ImportResult{
		CollectionName: collectionName,
		InsertedCount:  len(documents),
		Error:          nil,
	}, nil
}

// Disconnect はDisconnectのモック実装です
func (m *MockMongoRepository) Disconnect(ctx context.Context) error {
	if m.DisconnectFn != nil {
		return m.DisconnectFn(ctx)
	}
	// デフォルトの実装
	return nil
}
