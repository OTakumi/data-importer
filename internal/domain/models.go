package domain

import "time"

// Document MongoDBに保存するドキュメントを表す型
type Document map[string]any

// ImportResult インポート処理の結果を表す構造体
type ImportResult struct {
	FileName       string        // 処理されたファイル名（サービス層で使用）
	CollectionName string        // ドキュメントが挿入されたコレクション名
	InsertedCount  int           // 挿入されたドキュメントの数
	Duration       time.Duration // インポート処理にかかった時間（サービス層で使用）
	Error          error         // エラーが発生した場合のエラー情報
}

// RepositoryError リポジトリ層のエラーを表す構造体
type RepositoryError struct {
	Operation string
	Err       error
}

// Error エラーメッセージを返す
func (e *RepositoryError) Error() string {
	return e.Operation + ": " + e.Err.Error()
}

// Unwrap 元のエラーを返す
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// ServiceError サービス層のエラーを表す構造体
type ServiceError struct {
	Operation string
	Err       error
}

// Error エラーメッセージを返す
func (e *ServiceError) Error() string {
	return e.Operation + ": " + e.Err.Error()
}

// Unwrap 元のエラーを返す
func (e *ServiceError) Unwrap() error {
	return e.Err
}
