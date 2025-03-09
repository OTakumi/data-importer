package config

// TestConfig パッケージは設定管理機能のテストを行います。
//
// テスト観点:
// 1. 環境変数が設定されていない場合にデフォルト値が正しく使用されるか
// 2. 環境変数が設定されている場合にその値が正しく読み込まれるか
// 3. getEnv関数が環境変数の有無に応じて適切な値を返すか

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")

	// デフォルト値のテスト
	config := NewConfig()
	if config.MongoURI != "mongodb://mongodb:27017" {
		t.Errorf("Expected default MongoURI to be 'mongodb://mongodb:27017', got '%s'", config.MongoURI)
	}
	if config.DatabaseName != "test_db" {
		t.Errorf("Expected default DatabaseName to be 'test_db', got '%s'", config.DatabaseName)
	}
	if config.TimeoutSeconds != 10 {
		t.Errorf("Expected default TimeoutSeconds to be 10, got %d", config.TimeoutSeconds)
	}

	// 環境変数を設定してテスト
	os.Setenv("MONGODB_URI", "mongodb://custom:27017")
	os.Setenv("MONGODB_DATABASE", "custom_db")

	config = NewConfig()
	if config.MongoURI != "mongodb://custom:27017" {
		t.Errorf("Expected MongoURI to be 'mongodb://custom:27017', got '%s'", config.MongoURI)
	}
	if config.DatabaseName != "custom_db" {
		t.Errorf("Expected DatabaseName to be 'custom_db', got '%s'", config.DatabaseName)
	}

	// 環境変数をクリア
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
}

func TestGetEnv(t *testing.T) {
	// 環境変数が設定されていない場合のテスト
	os.Unsetenv("TEST_ENV")
	value := getEnv("TEST_ENV", "default_value")
	if value != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%s'", value)
	}

	// 環境変数が設定されている場合のテスト
	os.Setenv("TEST_ENV", "test_value")
	value = getEnv("TEST_ENV", "default_value")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// 環境変数をクリア
	os.Unsetenv("TEST_ENV")
}
