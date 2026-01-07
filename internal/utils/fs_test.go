package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_is_directory")
	assert.NoError(t, err)
	defer func() {
		os.RemoveAll(tempDir)
	}()

	// Create a test file in the directory
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Test with existing directory
	result := IsDirectory(tempDir)
	assert.True(t, result)

	// Test with existing file
	result = IsDirectory(testFile)
	assert.False(t, result)

	// Test with non-existent path
	result = IsDirectory("/non/existent/path")
	assert.False(t, result)
}

func TestIsFileExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_is_file_exists")
	assert.NoError(t, err)
	defer func() {
		os.RemoveAll(tempDir)
	}()

	// Create a test file in the directory
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Test with existing file
	result := IsFileExists(testFile)
	assert.True(t, result)

	// Test with non-existent file
	result = IsFileExists("/non/existent/file.txt")
	assert.False(t, result)

	// Test with directory path (should return false as it's not a file)
	result = IsFileExists(tempDir)
	assert.False(t, result)
}
