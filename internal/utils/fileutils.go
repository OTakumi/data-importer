package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem defines the interface for file operations
// This abstraction makes it easier to test file utilities
// without interacting with the actual file system
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	Walk(root string, fn filepath.WalkFunc) error
}

// RealFileSystem implements the FileSystem interface
// by delegating to the actual OS file functions
type RealFileSystem struct{}

// Stat wraps os.Stat to get file information
func (fs RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile wraps os.ReadFile to read file contents
func (fs RealFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// Walk wraps filepath.Walk to traverse directory trees
func (fs RealFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// FileUtils provides utility functions for file operations
// required by the MongoDB JSON importer
type FileUtils struct {
	fs FileSystem // The file system implementation to use
}

// NewFileUtils creates a new FileUtils instance with the given filesystem
// If nil is passed, it defaults to using the real file system
func NewFileUtils(fs FileSystem) *FileUtils {
	if fs == nil {
		fs = RealFileSystem{}
	}
	return &FileUtils{fs: fs}
}

// IsDirectory checks if the provided path is a directory
// Returns true if the path is a directory, false if it's a file
// Returns an error if the path doesn't exist or can't be accessed
func (fu *FileUtils) IsDirectory(path string) (bool, error) {
	fileInfo, err := fu.fs.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error checking path %s: %w", path, err)
	}
	return fileInfo.IsDir(), nil
}

// FindJSONFiles recursively finds all JSON files in the given directory
// Returns a slice of absolute paths to all JSON files in the directory tree
// Returns an error if the directory doesn't exist or can't be accessed
func (fu *FileUtils) FindJSONFiles(dirPath string) ([]string, error) {
	var jsonFiles []string

	// Walk through the directory tree
	err := fu.fs.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only add files with .json extension (case insensitive)
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".json" {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error finding JSON files in directory %s: %w", dirPath, err)
	}

	return jsonFiles, nil
}

// ParseJSONFile parses a JSON file into a slice of maps
// It handles two formats:
// 1. Array format: [{"key": "value"}, {"key": "value2"}]
// 2. Single object format: {"key": "value"}
// Returns a slice of maps representing JSON objects
// Returns an error if the file doesn't exist, can't be read, or contains invalid JSON
func (fu *FileUtils) ParseJSONFile(filePath string) ([]map[string]any, error) {
	// Read file content
	fileContent, err := fu.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Try to parse as array first
	var documents []map[string]any
	if err := json.Unmarshal(fileContent, &documents); err == nil {
		return documents, nil
	}

	// If parsing as array failed, try as single object
	var document map[string]any
	if err := json.Unmarshal(fileContent, &document); err != nil {
		return nil, fmt.Errorf("invalid JSON format in file %s: %w", filePath, err)
	}

	// Return single object in a slice
	return []map[string]any{document}, nil
}

// FilePathToCollectionName converts a file path to a collection name
// by extracting the file name without extension
// For example: "/path/to/users.json" becomes "users"
func FilePathToCollectionName(filePath string) string {
	fileName := filepath.Base(filePath)                         // Get the base filename from the path
	return strings.TrimSuffix(fileName, filepath.Ext(fileName)) // Remove the extension
}

// FileUtilsInterface defines the interface for file operations
type FileUtilsInterface interface {
	IsDirectory(path string) (bool, error)
	FindJSONFiles(dirPath string) ([]string, error)
	ParseJSONFile(filePath string) ([]map[string]interface{}, error)
}
