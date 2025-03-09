package domain

// Document MongoDBに保存するドキュメントを表す型
type Document map[string]interface{}

// ImportResult インポート処理の結果を表す構造体
type ImportResult struct {
	CollectionName string
	InsertedCount  int
	Error          error
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
