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

// TestIntegration は統合テストを実行する
// このテストは実際のMongoDBに接続するため、環境変数で設定されたMongoDBが利用可能である必要がある
// テスト実行前に以下のコマンドでDockerコンテナを起動しておくことを推奨:
// $ docker-compose up -d mongodb
func TestIntegration(t *testing.T) {
	// テスト環境のチェック
	// テスト用のMongoDBが利用可能かどうかを確認
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // デフォルト値
	}

	// テストデータのパスを設定
	testDataDir := "../testdata"
	usersArrayPath := filepath.Join(testDataDir, "users_array.json")
	productObjectPath := filepath.Join(testDataDir, "product_object.json")
	invalidJSONPath := filepath.Join(testDataDir, "invalid.json")

	// テストデータファイルが存在するか確認
	if _, err := os.Stat(usersArrayPath); os.IsNotExist(err) {
		t.Skipf("テストデータファイル %s が見つかりません。テストをスキップします。", usersArrayPath)
	}
	if _, err := os.Stat(productObjectPath); os.IsNotExist(err) {
		t.Skipf("テストデータファイル %s が見つかりません。テストをスキップします。", productObjectPath)
	}

	// コンテキストの作成（タイムアウト付き）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 設定の初期化
	cfg := &config.Config{
		MongoURI:       mongoURI,
		DatabaseName:   "test_db_integration",
		TimeoutSeconds: 30,
	}

	// MongoDBリポジトリの初期化
	repo, err := repository.NewMongoRepository(ctx, cfg)
	if err != nil {
		t.Fatalf("MongoDBへの接続に失敗しました: %v", err)
	}
	defer func() {
		if err := repo.Disconnect(context.Background()); err != nil {
			t.Logf("MongoDBの切断中にエラーが発生しました: %v", err)
		}
	}()

	// ファイルユーティリティの初期化
	fileUtils := utils.NewFileUtils(nil) // 実際のファイルシステムを使用

	// インポータサービスの初期化
	importer := service.NewMongoImporter(ctx, fileUtils, repo, 100) // テスト用に小さめのバッチサイズ

	// サブテストの実行
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

// testImportArrayJSON は配列形式のJSONファイルのインポートをテストする
func testImportArrayJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err != nil {
		t.Fatalf("配列JSONのインポートに失敗しました: %v", err)
	}

	if result.InsertedCount <= 0 {
		t.Errorf("ドキュメントが挿入されませんでした: %+v", result)
	}

	t.Logf("配列JSONのインポート結果: %d件のドキュメントが挿入されました (コレクション: %s)",
		result.InsertedCount, result.CollectionName)
}

// testImportObjectJSON は単一オブジェクト形式のJSONファイルのインポートをテストする
func testImportObjectJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err != nil {
		t.Fatalf("オブジェクトJSONのインポートに失敗しました: %v", err)
	}

	if result.InsertedCount != 1 {
		t.Errorf("期待される挿入件数は1件ですが、%d件が挿入されました: %+v",
			result.InsertedCount, result)
	}

	t.Logf("オブジェクトJSONのインポート結果: %d件のドキュメントが挿入されました (コレクション: %s)",
		result.InsertedCount, result.CollectionName)
}

// testImportInvalidJSON は不正なJSONファイルのインポートをテストする
func testImportInvalidJSON(t *testing.T, importer *service.MongoImporter, filePath string) {
	result, err := importer.ImportFile(filePath)
	if err == nil {
		t.Errorf("不正なJSONファイルのインポートが成功してしまいました: %+v", result)
	}

	t.Logf("不正なJSONファイルのインポートは想定通り失敗しました: %v", err)
}

// testImportDirectory はディレクトリのインポートをテストする
func testImportDirectory(t *testing.T, importer *service.MongoImporter, dirPath string) {
	results, err := importer.ImportDirectory(dirPath)
	if err != nil {
		// エラーが発生することは想定内（不正なJSONファイルが含まれているため）
		t.Logf("ディレクトリインポートで部分的なエラーが発生しました: %v", err)
	}

	// 結果の確認
	var successCount, errorCount int
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	t.Logf("ディレクトリインポート結果: %d件成功, %d件失敗, 合計%d件のファイル",
		successCount, errorCount, len(results))

	// 少なくとも1つの成功があるか確認
	if successCount == 0 {
		t.Errorf("ディレクトリインポートで1件も成功していません")
	}
}
