package repository

// TestMongoRepository パッケージはMongoDBリポジトリの機能テストを行います。
//
// テスト観点:
// 1. MockMongoRepositoryが正常に動作するか
// 2. InsertDocumentsメソッドが正しくドキュメントを挿入し、結果を返すか
// 3. 空のドキュメント配列の処理が適切に行われるか
// 4. エラーケースが適切に処理されるか
// 5. エラーのラップと取り出しが正常に機能するか

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"mongodb-importer/internal/config"
	"mongodb-importer/internal/domain"
)

// モックテスト（MongoDB接続なし）
func TestMockMongoRepository(t *testing.T) {
	ctx := context.Background()
	documents := []domain.Document{
		{"name": "test1", "value": 1},
		{"name": "test2", "value": 2},
	}

	// 成功ケース
	repo := &MockMongoRepository{
		InsertDocumentsFn: func(ctx context.Context, collectionName string, docs []domain.Document) (*domain.ImportResult, error) {
			return &domain.ImportResult{
				CollectionName: collectionName,
				InsertedCount:  len(docs),
				Error:          nil,
			}, nil
		},
	}

	result, err := repo.InsertDocuments(ctx, "testCollection", documents)
	if err != nil {
		t.Errorf("エラーは発生すべきではありません: %v", err)
	}
	if result.CollectionName != "testCollection" {
		t.Errorf("コレクション名が一致しません: expected=%s, got=%s", "testCollection", result.CollectionName)
	}
	if result.InsertedCount != 2 {
		t.Errorf("挿入されたドキュメント数が一致しません: expected=2, got=%d", result.InsertedCount)
	}

	// エラーケース
	expectedErr := errors.New("データベースエラー")
	repo = &MockMongoRepository{
		InsertDocumentsFn: func(ctx context.Context, collectionName string, docs []domain.Document) (*domain.ImportResult, error) {
			return nil, expectedErr
		},
	}

	_, err = repo.InsertDocuments(ctx, "testCollection", documents)
	if err != expectedErr {
		t.Errorf("エラーが一致しません: expected=%v, got=%v", expectedErr, err)
	}

	// 空のドキュメント配列
	repo = &MockMongoRepository{}
	result, err = repo.InsertDocuments(ctx, "emptyCollection", []domain.Document{})
	if err != nil {
		t.Errorf("空のドキュメント配列でエラーは発生すべきではありません: %v", err)
	}
	if result.InsertedCount != 0 {
		t.Errorf("挿入されたドキュメント数が一致しません: expected=0, got=%d", result.InsertedCount)
	}
}

// 統合テスト（MongoDB接続あり - テスト環境が準備されている場合のみ実行）
func TestMongoRepository_Integration(t *testing.T) {
	// MongoDB接続が利用可能かどうかをチェック
	// 接続が利用不可の場合はテストをスキップ
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.ClearCollections()

	mt.Run("insert_documents", func(mt *mtest.T) {
		// モックレスポンスの設定（バッチ処理に対応して複数のレスポンスを追加）
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// テスト用のMongoDBクライアントを使用した設定
		cfg := &config.Config{
			MongoURI:       "mongodb://localhost:27017",
			DatabaseName:   "test_db",
			TimeoutSeconds: 10,
		}

		// モックされたクライアントを使用するリポジトリを作成
		repo := &MongoRepository{
			client: mt.Client,
			db:     mt.Client.Database(cfg.DatabaseName),
		}

		// テストデータ
		documents := []domain.Document{
			{"name": "test1", "value": 1},
			{"name": "test2", "value": 2},
		}

		// ドキュメントの挿入
		result, err := repo.InsertDocuments(context.Background(), "testCollection", documents)
		if err != nil {
			t.Errorf("ドキュメント挿入でエラーが発生しました: %v", err)
		}
		if result.CollectionName != "testCollection" {
			t.Errorf("コレクション名が一致しません: expected=%s, got=%s", "testCollection", result.CollectionName)
		}
	})

	mt.Run("insert_large_batch", func(mt *mtest.T) {
		// 大量データのバッチ処理テスト
		// 複数回のInsertManyに対応するレスポンスを設定
		successResponse1 := mtest.CreateSuccessResponse()
		successResponse2 := mtest.CreateSuccessResponse()
		mt.AddMockResponses(successResponse1, successResponse2)

		// テスト用の設定
		cfg := &config.Config{
			MongoURI:       "mongodb://localhost:27017",
			DatabaseName:   "test_db",
			TimeoutSeconds: 10,
		}

		// モックされたクライアントを使用するリポジトリを作成
		repo := &MongoRepository{
			client: mt.Client,
			db:     mt.Client.Database(cfg.DatabaseName),
		}

		// バッチ処理をテストするための大量データ作成（テスト用に小さめの値を使用）
		largeDocuments := make([]domain.Document, 0, 2000)
		for i := 0; i < 2000; i++ {
			largeDocuments = append(largeDocuments, domain.Document{
				"index": i,
				"value": fmt.Sprintf("test value %d", i),
			})
		}

		// ドキュメントの挿入
		result, err := repo.InsertDocuments(context.Background(), "largeCollection", largeDocuments)
		if err != nil {
			t.Errorf("大量データ挿入でエラーが発生しました: %v", err)
		}
		if result.CollectionName != "largeCollection" {
			t.Errorf("コレクション名が一致しません: expected=%s, got=%s", "largeCollection", result.CollectionName)
		}
	})
}

// エラーケースのテスト
func TestRepositoryErrorHandling(t *testing.T) {
	// リポジトリエラーの作成と取り出し
	originalErr := errors.New("original error")
	repoErr := &domain.RepositoryError{
		Operation: "test operation",
		Err:       originalErr,
	}

	// Unwrap()メソッドのテスト
	err := errors.Unwrap(repoErr)
	if err != originalErr {
		t.Errorf("Unwrapされたエラーが一致しません: expected=%v, got=%v", originalErr, err)
	}

	// As()を使ったエラーの型チェック
	var targetErr *domain.RepositoryError
	if !errors.As(repoErr, &targetErr) {
		t.Errorf("errors.Asによる型チェックが失敗しました")
	}
	if targetErr != repoErr {
		t.Errorf("errors.Asで取得したエラーが一致しません")
	}

	// Is()を使ったエラーチェック
	if !errors.Is(repoErr, originalErr) {
		t.Errorf("errors.Isによるエラーチェックが失敗しました")
	}
}
