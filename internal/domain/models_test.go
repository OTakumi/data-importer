package domain

// TestModels パッケージはドメインモデルの機能テストを行います。
//
// テスト観点:
// 1. RepositoryErrorのError()メソッドが適切なエラーメッセージを生成するか
// 2. RepositoryErrorのUnwrap()メソッドが元のエラーを正しく返すか
// 3. ImportResult構造体のフィールドが適切に設定され、アクセス可能か

import (
	"errors"
	"testing"
)

func TestRepositoryError(t *testing.T) {
	// エラーの作成
	originalErr := errors.New("original error")
	repoErr := &RepositoryError{
		Operation: "test operation",
		Err:       originalErr,
	}

	// Error()メソッドのテスト
	expectedErrMsg := "test operation: original error"
	if repoErr.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, repoErr.Error())
	}

	// Unwrap()メソッドのテスト
	unwrappedErr := repoErr.Unwrap()
	if unwrappedErr != originalErr {
		t.Errorf("Expected unwrapped error to be the original error")
	}
}

func TestImportResult(t *testing.T) {
	// ImportResultの作成と値のテスト
	result := ImportResult{
		CollectionName: "testCollection",
		InsertedCount:  5,
		Error:          nil,
	}

	if result.CollectionName != "testCollection" {
		t.Errorf("Expected CollectionName to be 'testCollection', got '%s'", result.CollectionName)
	}

	if result.InsertedCount != 5 {
		t.Errorf("Expected InsertedCount to be 5, got %d", result.InsertedCount)
	}

	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}

	// エラーがある場合のテスト
	testErr := errors.New("test error")
	result = ImportResult{
		CollectionName: "errorCollection",
		InsertedCount:  0,
		Error:          testErr,
	}

	if result.Error != testErr {
		t.Errorf("Expected Error to be 'test error'")
	}
}
