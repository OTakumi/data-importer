///
/// discription: アプリケーション設定を保持するパッケージ
///

package config

import (
	"os"
)

// Config アプリケーション設定を保持する構造体
type Config struct {
	MongoURI       string
	DatabaseName   string
	TimeoutSeconds int
}

// NewConfig 環境変数またはデフォルト値から設定を読み込む
func NewConfig() *Config {
	return &Config{
		MongoURI:       getEnv("MONGODB_URI", "mongodb://mongodb:27017"),
		DatabaseName:   getEnv("MONGODB_DATABASE", "test_db"),
		TimeoutSeconds: 10,
	}
}

// getEnv 環境変数の値を取得、設定されていない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
