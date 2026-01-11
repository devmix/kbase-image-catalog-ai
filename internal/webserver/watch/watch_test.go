package watch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCatalogWatcher(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-archive")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test creating a new catalog watcher
	watcher, err := NewCatalogWatcher(nil, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.Equal(t, tempDir, watcher.archiveDir)
	assert.False(t, watcher.isRunning)

	// Test with empty archive path
	watcher2, err := NewCatalogWatcher(nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, watcher2)
	assert.Equal(t, "", watcher2.archiveDir)
}

func TestCatalogWatcher_Start(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-archive")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	watcher, err := NewCatalogWatcher(nil, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	// Test starting the watcher
	err = watcher.Start()
	assert.NoError(t, err)
	assert.True(t, watcher.isRunning)

	// Try to start again - should not error but do nothing
	err = watcher.Start()
	assert.NoError(t, err)
	assert.True(t, watcher.isRunning)

	// Clean up
	watcher.Stop()
}

func TestCatalogWatcher_Stop(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-archive")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	watcher, err := NewCatalogWatcher(nil, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	// Start the watcher first
	err = watcher.Start()
	assert.NoError(t, err)
	assert.True(t, watcher.isRunning)

	// Test stopping the watcher
	err = watcher.Stop()
	assert.NoError(t, err)
	assert.False(t, watcher.isRunning)

	// Try to stop again - should not error but do nothing
	err = watcher.Stop()
	assert.NoError(t, err)
	assert.False(t, watcher.isRunning)
}

func TestCatalogWatcher_addDirectoriesToWatch(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-archive")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some test subdirectories
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(tempDir, "subdir2", "nested")

	err = os.MkdirAll(subDir1, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(subDir2, 0755)
	assert.NoError(t, err)

	watcher, err := NewCatalogWatcher(nil, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	// Test adding directories to watch
	err = watcher.addDirectoriesToWatch(tempDir)
	assert.NoError(t, err)

	// Note: We can't easily test the actual fsnotify.Watcher behavior in unit tests,
	// but we can at least check that it doesn't panic or return an error for valid paths.
}

func TestCatalogWatcher_handleFileChange(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-archive")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some test files and directories
	subDir1 := filepath.Join(tempDir, "collection1")
	err = os.MkdirAll(subDir1, 0755)
	assert.NoError(t, err)

	// Create a test image file
	testImageFile := filepath.Join(subDir1, "test.png")
	_, err = os.Create(testImageFile)
	assert.NoError(t, err)

	watcher, err := NewCatalogWatcher(nil, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	// Since we can't easily test the actual AddTask functionality in this context,
	// just make sure that handleFileChange doesn't panic
	watcher.handleFileChange(testImageFile)

	// Test with non-image file - should not call AddTask
	nonImageFile := filepath.Join(subDir1, "test.txt")
	_, err = os.Create(nonImageFile)
	assert.NoError(t, err)

	// Just ensure no panic occurs
	watcher.handleFileChange(nonImageFile)

	// Test with invalid path - should not panic
	invalidPath := filepath.Join(tempDir, "nonexistent", "test.png")
	watcher.handleFileChange(invalidPath)
}
