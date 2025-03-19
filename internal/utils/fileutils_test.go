package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// MockFileInfo implements os.FileInfo interface for testing
// It provides the minimum implementation needed for our tests
type MockFileInfo struct {
	isDir bool
	name  string
}

// Name returns the base name of the file
func (m MockFileInfo) Name() string { return m.name }

// Size returns the length in bytes - not used in our tests
func (m MockFileInfo) Size() int64 { return 0 }

// Mode returns the file mode bits - not used in our tests
func (m MockFileInfo) Mode() os.FileMode { return 0 }

// ModTime returns the modification time - not used in our tests
func (m MockFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir reports whether the file is a directory
func (m MockFileInfo) IsDir() bool { return m.isDir }

// Sys returns the underlying data source - not used in our tests
func (m MockFileInfo) Sys() interface{} { return nil }

// MockFileSystem implements FileSystem interface for testing
// It provides an in-memory representation of a file system
type MockFileSystem struct {
	files map[string][]byte // Map of file paths to their content
	dirs  map[string]bool   // Map of directory paths to a bool indicating they exist
}

// NewMockFileSystem creates a new instance of MockFileSystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// Stat returns FileInfo for a file or directory
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	// Check if it's a directory
	if isDir, ok := m.dirs[name]; ok && isDir {
		return MockFileInfo{isDir: true, name: filepath.Base(name)}, nil
	}

	// Check if it's a file
	if _, ok := m.files[name]; ok {
		return MockFileInfo{isDir: false, name: filepath.Base(name)}, nil
	}

	return nil, os.ErrNotExist
}

// ReadFile returns the content of a file
func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	content, exists := m.files[filename]
	if !exists {
		return nil, os.ErrNotExist
	}
	return content, nil
}

// Walk simulates traversing a directory tree
func (m *MockFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	// Check if root exists
	rootInfo, err := m.Stat(root)
	if err != nil {
		return fn(root, nil, err)
	}

	// Call fn on root
	err = fn(root, rootInfo, nil)
	if err != nil {
		return err
	}

	// If root is not a directory, we're done
	if !rootInfo.IsDir() {
		return nil
	}

	// Process all files and directories
	paths := []string{}
	rootWithSep := root
	if !strings.HasSuffix(rootWithSep, string(filepath.Separator)) {
		rootWithSep += string(filepath.Separator)
	}

	// Collect all paths under root
	for path := range m.dirs {
		if strings.HasPrefix(path, rootWithSep) && path != root {
			paths = append(paths, path)
		}
	}

	for path := range m.files {
		if strings.HasPrefix(path, rootWithSep) {
			paths = append(paths, path)
		}
	}

	// Sort paths to simulate natural traversal order
	sort.Strings(paths)

	// Call function on all paths
	for _, path := range paths {
		pathInfo, err := m.Stat(path)
		if err != nil {
			if err := fn(path, nil, err); err != nil && err != filepath.SkipDir {
				return err
			}
			continue
		}

		err = fn(path, pathInfo, nil)
		if err != nil {
			if err == filepath.SkipDir {
				// Skip this directory or file
				continue
			}
			return err
		}
	}

	return nil
}

// AddFile adds a file to the mock filesystem
func (m *MockFileSystem) AddFile(path string, content []byte) {
	m.files[path] = content

	// Ensure parent directories exist
	dir := filepath.Dir(path)
	m.ensureParentDirs(dir)
}

// AddDirectory adds a directory to the mock filesystem
func (m *MockFileSystem) AddDirectory(path string) {
	m.dirs[path] = true

	// Ensure parent directories exist
	if path != "/" && path != "." {
		dir := filepath.Dir(path)
		m.ensureParentDirs(dir)
	}
}

// ensureParentDirs makes sure all parent directories of a path exist
func (m *MockFileSystem) ensureParentDirs(dir string) {
	if dir == "/" || dir == "." {
		return
	}

	m.dirs[dir] = true

	// Recursively ensure parent's parent exists
	parent := filepath.Dir(dir)
	if parent != dir {
		m.ensureParentDirs(parent)
	}
}

// TestIsDirectory tests the IsDirectory function
func TestIsDirectory(t *testing.T) {
	// Setup mock filesystem with test data
	mockFS := NewMockFileSystem()
	mockFS.AddDirectory("/test_dir")
	mockFS.AddFile("/test_file.json", []byte(`{"key":"value"}`))

	fu := NewFileUtils(mockFS)

	// Test cases
	tests := []struct {
		name           string
		path           string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "Directory path",
			path:           "/test_dir",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "File path",
			path:           "/test_file.json",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "Non-existent path",
			path:           "/non_existent",
			expectedResult: false,
			expectError:    true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fu.IsDirectory(tt.path)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result only if no error expected
			if !tt.expectError && result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// TestFindJSONFiles tests the FindJSONFiles function
func TestFindJSONFiles(t *testing.T) {
	// Setup mock filesystem with test data
	mockFS := NewMockFileSystem()
	mockFS.AddDirectory("/root")
	mockFS.AddDirectory("/root/dir1")
	mockFS.AddDirectory("/root/dir2")
	mockFS.AddFile("/root/file1.json", []byte("{}"))
	mockFS.AddFile("/root/file2.txt", []byte("text"))
	mockFS.AddFile("/root/dir1/file3.json", []byte("{}"))
	mockFS.AddFile("/root/dir2/file4.json", []byte("{}"))
	mockFS.AddFile("/root/dir2/file5.txt", []byte("text"))

	fu := NewFileUtils(mockFS)

	// Test cases
	tests := []struct {
		name          string
		dirPath       string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:          "Root directory",
			dirPath:       "/root",
			expectedFiles: []string{"/root/file1.json", "/root/dir1/file3.json", "/root/dir2/file4.json"},
			expectError:   false,
		},
		{
			name:          "Subdirectory",
			dirPath:       "/root/dir1",
			expectedFiles: []string{"/root/dir1/file3.json"},
			expectError:   false,
		},
		{
			name:          "Non-existent directory",
			dirPath:       "/non_existent",
			expectedFiles: nil,
			expectError:   true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := fu.FindJSONFiles(tt.dirPath)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result only if no error expected
			if !tt.expectError {
				// Sort both slices for comparison
				sortedExpected := make([]string, len(tt.expectedFiles))
				copy(sortedExpected, tt.expectedFiles)
				sort.Strings(sortedExpected)

				sortedActual := make([]string, len(files))
				copy(sortedActual, files)
				sort.Strings(sortedActual)

				if !reflect.DeepEqual(sortedActual, sortedExpected) {
					t.Errorf("Expected files %v, got %v", sortedExpected, sortedActual)
				}
			}
		})
	}
}

// TestParseJSONFile tests the ParseJSONFile function
func TestParseJSONFile(t *testing.T) {
	// Setup mock filesystem with test data
	mockFS := NewMockFileSystem()

	// Array format JSON
	arrayJSON := []byte(`[{"id":1,"name":"Item 1"},{"id":2,"name":"Item 2"}]`)
	mockFS.AddFile("/array.json", arrayJSON)

	// Single object format JSON
	objectJSON := []byte(`{"id":1,"name":"Single Item"}`)
	mockFS.AddFile("/object.json", objectJSON)

	// Invalid JSON
	invalidJSON := []byte(`{"id":1,"name":"Broken JSON"`)
	mockFS.AddFile("/invalid.json", invalidJSON)

	fu := NewFileUtils(mockFS)

	// Test cases
	tests := []struct {
		name         string
		filePath     string
		expectedDocs []map[string]interface{}
		expectError  bool
	}{
		{
			name:     "Array format JSON",
			filePath: "/array.json",
			expectedDocs: []map[string]interface{}{
				{"id": float64(1), "name": "Item 1"},
				{"id": float64(2), "name": "Item 2"},
			},
			expectError: false,
		},
		{
			name:     "Single object format JSON",
			filePath: "/object.json",
			expectedDocs: []map[string]interface{}{
				{"id": float64(1), "name": "Single Item"},
			},
			expectError: false,
		},
		{
			name:         "Invalid JSON",
			filePath:     "/invalid.json",
			expectedDocs: nil,
			expectError:  true,
		},
		{
			name:         "Non-existent file",
			filePath:     "/non_existent.json",
			expectedDocs: nil,
			expectError:  true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := fu.ParseJSONFile(tt.filePath)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result only if no error expected
			if !tt.expectError && !reflect.DeepEqual(docs, tt.expectedDocs) {
				t.Errorf("Expected documents %v, got %v", tt.expectedDocs, docs)
			}
		})
	}
}

// TestFilePathToCollectionName tests the FilePathToCollectionName function
func TestFilePathToCollectionName(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		filePath       string
		expectedResult string
	}{
		{
			name:           "Simple filename",
			filePath:       "users.json",
			expectedResult: "users",
		},
		{
			name:           "Path with directory",
			filePath:       "/data/products.json",
			expectedResult: "products",
		},
		{
			name:           "Filename with multiple dots",
			filePath:       "order.items.json",
			expectedResult: "order.items",
		},
		{
			name:           "Filename with no extension",
			filePath:       "metadata",
			expectedResult: "metadata",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilePathToCollectionName(tt.filePath)

			if result != tt.expectedResult {
				t.Errorf("Expected result %q, got %q", tt.expectedResult, result)
			}
		})
	}
}
