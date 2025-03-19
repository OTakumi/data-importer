# MongoDB JSON Importer Makefile

# 変数定義
BINARY_NAME=data-importer
BUILD_DIR=./build
MAIN_PATH=./cmd/importer
DOCKER_COMPOSE=docker-compose

# Go関連コマンド
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod
GOGET=$(GO) get
GOFMT=$(GO) fmt

# ディレクトリ構造作成
.PHONY: init
init:
	mkdir -p $(BUILD_DIR)
	mkdir -p cmd/importer
	mkdir -p internal/config
	mkdir -p internal/domain
	mkdir -p internal/repository
	mkdir -p internal/service
	mkdir -p internal/utils
	mkdir -p data
	mkdir -p testdata/valid
	mkdir -p testdata/invalid

# 依存関係の解決
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# ビルド
.PHONY: build
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# テスト実行
.PHONY: test
test:
	$(GOTEST) -v ./internal/...

# 統合テスト実行
.PHONY: test-integration
test-integration:
	$(GOTEST) -v ./tests/integration

# コードカバレッジレポート
.PHONY: coverage
coverage:
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# コードフォーマット
.PHONY: fmt
fmt:
	$(GOFMT) ./...

# クリーンアップ
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Dockerビルド
.PHONY: docker-build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker Compose起動
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up --build

# Docker Compose停止
.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

# 単一ファイルをインポート
.PHONY: import-file
import-file:
	$(DOCKER_COMPOSE) run --rm $(BINARY_NAME) /data/$(file)

# ディレクトリ内のすべてのファイルをインポート
.PHONY: import-dir
import-dir:
	$(DOCKER_COMPOSE) run --rm $(BINARY_NAME) /data

# ヘルプ表示
.PHONY: help
help:
	@echo "MongoDB JSON Importer - 利用可能なコマンド一覧"
	@echo "init              - プロジェクトディレクトリ構造を作成"
	@echo "deps              - 依存関係を解決"
	@echo "build             - アプリケーションをビルド"
	@echo "test              - テストを実行"
	@echo "coverage          - コードカバレッジレポートを生成"
	@echo "fmt               - コードをフォーマット"
	@echo "clean             - ビルドファイルを削除"
	@echo "docker-build      - Dockerイメージをビルド"
	@echo "docker-up         - Docker Composeでサービスを起動"
	@echo "docker-down       - Docker Composeでサービスを停止"
	@echo "import-file file=<ファイル名>  - 指定したJSONファイルをインポート"
	@echo "import-dir        - dataディレクトリ内の全JSONファイルをインポート"

# デフォルトターゲット
.DEFAULT_GOAL := help
