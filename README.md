# MongoDB JSON Importer

GoでJSONファイルからMongoDBへデータをインポートするプログラムです。ファイル単体とディレクトリ両方に対応しています。レイヤードアーキテクチャに基づいて設計されています。

## 機能

- 単一のJSONファイルからMongoDBコレクションへのインポート
- ディレクトリ内の複数のJSONファイルを再帰的に処理
- コレクション名はファイル名から自動的に設定
- Dockerを使用したローカル環境での実行
- 拡張性の高いレイヤード設計

## プロジェクト構造

```
mongodb-importer/
├── cmd/
│   └── importer/
│       └── main.go       # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go     # 設定管理
│   ├── domain/
│   │   └── models.go     # ドメインモデル
│   ├── repository/
│   │   └── mongodb.go    # データアクセス層
│   ├── service/
│   │   └── importer.go   # ビジネスロジック層
│   └── utils/
│       └── fileutils.go  # ファイル操作ユーティリティ
├── Dockerfile
├── docker-compose.yaml
├── go.mod
└── README.md
```

## 必要条件

- Go 1.20以上
- Docker および Docker Compose

## セットアップと実行方法

### 1. リポジトリのクローン

```bash
git clone <リポジトリURL>
cd mongodb-importer
```

### 2. サンプルデータの準備

`data`ディレクトリを作成し、JSONファイルを配置します：

```bash
mkdir -p data
# サンプルJSONファイルをdata/users.jsonとして保存
```

### 3. Docker Composeで実行

```bash
docker-compose up --build
```

### 4. 手動実行（オプション）

```bash
# MongoDBコンテナが実行中であることを確認
docker-compose up -d mongodb

# 特定のJSONファイルを指定して実行
docker-compose run --rm mongodb-importer /data/users.json

# ディレクトリを指定して実行（/data内のすべてのJSONファイルが処理されます）
docker-compose run --rm mongodb-importer /data
```

## 環境変数

以下の環境変数を使って設定をカスタマイズできます：

- `MONGODB_URI`: MongoDB接続URI (デフォルト: `mongodb://mongodb:27017`)
- `MONGODB_DATABASE`: 使用するデータベース名 (デフォルト: `test_db`)

## 拡張方法

各レイヤーは適切に分離されているため、機能の追加や変更が容易です：

1. **設定の追加**: `config/config.go` を変更
2. **新しいリポジトリの追加**: `repository/` に新しいファイルを追加
3. **ビジネスロジックの拡張**: `service/` に機能を追加
4. **ユーティリティの拡張**: `utils/` に汎用的な関数を追加

## 注意事項

- JSONファイルは配列形式（複数のドキュメント）または単一オブジェクト形式に対応しています。
- 既存のコレクションにデータを追加する形で動作します（上書きはされません）。
- エラーが発生しても処理は継続され、エラーはログに記録されます。
