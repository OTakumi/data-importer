# MongoDB JSONインポーター

JSONファイルからMongoDBへデータを効率的にインポートするためのGoツールです。ファイル単位またはディレクトリ単位の処理に対応し、大量データのバッチ処理を行うことができます。

## 機能

- 単一のJSONファイルからMongoDBコレクションへのインポート
- ディレクトリ内の複数のJSONファイルを再帰的に処理
- ファイル名をコレクション名として自動設定
- 配列形式の複数ドキュメントと単一オブジェクト形式の両方に対応
- ドキュメントのバッチ処理による効率的なインポート
- 環境変数または.envファイルによる柔軟な設定
- MongoDB固有の_idフィールドを自動的に除去してインポートエラーを防止
- $date形式の日付フィールドを標準形式に変換
- Docker環境での簡単な実行

## 必要条件

- Go 1.20以上
- MongoDB (ローカルまたはリモート)
- Docker および Docker Compose (オプション)

## インストール

```bash
# リポジトリのクローン
git clone https://github.com/OTakumi/data-importer.git
cd data-importer

# ビルド
make build
```

## 使い方

### コマンドライン

```bash
# 単一のJSONファイルをインポート
./data-importer path/to/file.json

# ディレクトリ内のすべてのJSONファイルをインポート
./data-importer path/to/directory

# カスタム環境設定ファイルを使用
./data-importer -env=custom.env path/to/file.json

# ヘルプを表示
./data-importer --help
```

### Docker環境での実行

```bash
# ビルドして実行
make docker-build
make docker-up

# JSONファイルをインポート
make import-file file=path/to/file.json

# ディレクトリをインポート
make import-dir dir=path/to/directory

# 停止
make docker-down
```

## 環境設定

環境変数または.envファイルを使用して設定をカスタマイズできます。

### 環境変数

#### MongoDB接続設定（オプション1：個別コンポーネント）
- `MONGODB_USERNAME`: MongoDBのユーザー名
- `MONGODB_PASSWORD`: MongoDBのパスワード
- `MONGODB_HOST`: MongoDBのホスト名（デフォルト: "mongodb"）
- `MONGODB_PORT`: MongoDBのポート番号（デフォルト: "27017"）
- `MONGODB_AUTH_DATABASE`: 認証用データベース名（オプション）
- `MONGODB_REPLICA_SET`: レプリカセット名（オプション）

#### MongoDB接続設定（オプション2：URIを直接指定）
- `MONGODB_URI`: MongoDB接続URI（デフォルト: `mongodb://mongodb:27017`）
- `MONGODB_DATABASE`: 使用するデータベース名（デフォルト: `test_db`）

#### アプリケーション設定
- `MONGODB_TIMEOUT`: タイムアウト秒数（デフォルト: `10`）
- `MONGODB_BATCH_SIZE`: バッチサイズ（デフォルト: `1000`）

### .envファイル

以下のような.envファイルを作成することで、環境変数の設定が可能です：

```
# MongoDB接続設定（オプション1：個別コンポーネント）
MONGODB_USERNAME=admin
MONGODB_PASSWORD=password123
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_AUTH_DATABASE=admin

# MongoDB接続設定（オプション2：URIを直接指定）
#MONGODB_URI=mongodb://admin:password123@localhost:27017/?authSource=admin

# データベースと設定
MONGODB_DATABASE=import_db
MONGODB_TIMEOUT=30
MONGODB_BATCH_SIZE=500
```

## テスト

```bash
# すべてのテストを実行
make test

# 統合テストを実行
make integration-test

# カバレッジレポートを生成
make coverage
```

統合テストを実行する場合は、`.env.test`ファイルを作成するか、環境変数を設定してください。詳細は[統合テストのREADME](tests/integration/README.md)を参照してください。

## プロジェクト構造

```
data-importer/
├── cmd/
│   └── importer/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go            # 設定管理
│   ├── domain/
│   │   └── models.go            # ドメインモデル
│   ├── repository/
│   │   └── mongodb.go           # データアクセス層
│   ├── service/
│   │   └── importer.go          # ビジネスロジック層
│   └── utils/
│       └── fileutils.go         # ファイル操作ユーティリティ
├── tests/
│   ├── integration/             # 統合テスト
│   └── testdata/                # テスト用データファイル
├── Dockerfile
├── docker-compose.yaml
├── .env.sample                  # 環境変数設定サンプル
├── Makefile
└── README.md
```

## 今後の開発計画

### 短期的な改善点
1. **ストリーミング処理の実装**
   - 巨大JSONファイルのメモリ効率を向上
   - JSONストリーミングパーサーの導入

2. **並列処理の最適化**
   - 複数ファイルの並列処理効率の向上
   - リソース使用量の最適化

3. **詳細な進捗表示**
   - リアルタイム進捗バーの実装
   - 処理速度や統計情報の表示

### 中長期的な拡張計画
1. **データ変換機能**
   - JSONスキーマの変換機能
   - フィールド名やデータ型の変換

2. **他のデータソースへの対応**
   - CSV、XML、YAMLなど他の形式への対応
   - さまざまなデータベースへのエクスポート機能

3. **GUIインターフェースの追加**
   - ウェブベースの管理インターフェース
   - ドラッグアンドドロップによるファイル操作

## ライセンス

[MIT License](LICENSE)
