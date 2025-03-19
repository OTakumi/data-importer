package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/OTakumi/data-importer/internal/config"
	"github.com/OTakumi/data-importer/internal/domain"
)

// Repository データアクセスのインターフェース
type Repository interface {
	InsertDocuments(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error)
	Disconnect(ctx context.Context) error
}

// MongoRepository MongoDBとの接続を管理するリポジトリ
type MongoRepository struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoRepository MongoDBリポジトリの新しいインスタンスを作成する
func NewMongoRepository(ctx context.Context, cfg *config.Config) (*MongoRepository, error) {
	// 接続オプションの設定
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	// MongoDBに接続
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, &domain.RepositoryError{
			Operation: "MongoDB接続",
			Err:       err,
		}
	}

	// 接続確認
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, &domain.RepositoryError{
			Operation: "MongoDB接続確認",
			Err:       err,
		}
	}

	// データベースの取得
	db := client.Database(cfg.DatabaseName)

	return &MongoRepository{
		client: client,
		db:     db,
	}, nil
}

// InsertDocuments 指定したコレクションに複数のドキュメントをバッチ処理で挿入する
func (r *MongoRepository) InsertDocuments(ctx context.Context, collectionName string, documents []domain.Document) (*domain.ImportResult, error) {
	if len(documents) == 0 {
		return &domain.ImportResult{
			CollectionName: collectionName,
			InsertedCount:  0,
			Error:          nil,
		}, nil
	}

	// コレクションの取得
	collection := r.db.Collection(collectionName)

	// バッチサイズを設定（パフォーマンスとメモリ使用量のバランスを取る）
	batchSize := 1000
	totalBatches := (len(documents) + batchSize - 1) / batchSize // 切り上げ除算
	totalInserted := 0

	// バッチ処理
	for i := 0; i < len(documents); i += batchSize {
		// 現在のバッチの終了位置を計算
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		// 現在のバッチを取得
		batch := documents[i:end]

		// インターフェースのスライスに変換
		var interfaceSlice []interface{}
		for _, doc := range batch {
			interfaceSlice = append(interfaceSlice, doc)
		}

		// バッチをInsertManyで挿入
		result, err := collection.InsertMany(ctx, interfaceSlice)
		if err != nil {
			return nil, &domain.RepositoryError{
				Operation: fmt.Sprintf("コレクション %s へのドキュメント挿入（バッチ %d/%d）",
					collectionName, i/batchSize+1, totalBatches),
				Err: err,
			}
		}

		totalInserted += len(result.InsertedIDs)

		// バッチ処理の進捗をログに出力（大量データのデバッグに役立つ）
		if totalBatches > 1 {
			fmt.Printf("コレクション %s: バッチ %d/%d 完了（%d件挿入）\n",
				collectionName, i/batchSize+1, totalBatches, len(result.InsertedIDs))
		}
	}

	// 結果の作成
	return &domain.ImportResult{
		CollectionName: collectionName,
		InsertedCount:  totalInserted,
		Error:          nil,
	}, nil
}

// Disconnect MongoDBとの接続を切断する
func (r *MongoRepository) Disconnect(ctx context.Context) error {
	if r.client != nil {
		if err := r.client.Disconnect(ctx); err != nil {
			return &domain.RepositoryError{
				Operation: "MongoDB切断",
				Err:       err,
			}
		}
	}
	return nil
}
