package processor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kbase-catalog/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewFileScanner(t *testing.T) {
	cfg := &config.Config{}
	fs := NewFileScanner(cfg)

	assert.NotNil(t, fs)
	assert.Equal(t, cfg, fs.config)
}

func TestHasImages_EmptyDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_has_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result := fs.HasImages(tempDir)
	assert.False(t, result)
}

func TestHasImages_DirectoryWithImages(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_has_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create image files in the directory
	img1Path := filepath.Join(tempDir, "test.jpg")
	img2Path := filepath.Join(tempDir, "test.png")
	err = os.WriteFile(img1Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(img2Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result := fs.HasImages(tempDir)
	assert.True(t, result)
}

func TestHasImages_DirectoryWithNonImageFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_has_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create non-image files in the directory
	txtPath := filepath.Join(tempDir, "readme.txt")
	err = os.WriteFile(txtPath, []byte("some text"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result := fs.HasImages(tempDir)
	assert.False(t, result)
}

func TestHasImages_DirectoryWithUnsupportedExtension(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_has_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create file with unsupported extension
	txtPath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(txtPath, []byte("some text"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result := fs.HasImages(tempDir)
	assert.False(t, result)
}

func TestHasImages_NonExistentDirectory(t *testing.T) {
	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result := fs.HasImages("/non/existent/directory")
	assert.False(t, result)
}

func TestFindImagesToProcess_EmptyDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_find_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestFindImagesToProcess_DirectoryWithImages(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_find_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create image files in the directory
	img1Path := filepath.Join(tempDir, "test.jpg")
	img2Path := filepath.Join(tempDir, "test.png")
	img3Path := filepath.Join(tempDir, "test.jpeg")
	err = os.WriteFile(img1Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(img2Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(img3Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify all returned paths are to the created image files
	for _, path := range result {
		baseName := filepath.Base(path)
		assert.Contains(t, []string{"test.jpg", "test.png", "test.jpeg"}, baseName)
	}
}

func TestFindImagesToProcess_DirectoryWithMixedFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_find_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create image and non-image files in the directory
	img1Path := filepath.Join(tempDir, "test.jpg")
	txtPath := filepath.Join(tempDir, "readme.txt")
	img2Path := filepath.Join(tempDir, "test.png")

	err = os.WriteFile(img1Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(txtPath, []byte("some text"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(img2Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify all returned paths are to image files only
	for _, path := range result {
		baseName := filepath.Base(path)
		assert.Contains(t, []string{"test.jpg", "test.png"}, baseName)
	}
}

func TestFindImagesToProcess_IgnoresIndexFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_find_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create image and index files in the directory
	img1Path := filepath.Join(tempDir, "test.jpg")
	indexJsonPath := filepath.Join(tempDir, "index.json")
	indexMdPath := filepath.Join(tempDir, "index.md")

	err = os.WriteFile(img1Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(indexJsonPath, []byte("{}"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(indexMdPath, []byte("# Index"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Verify only the image file is returned
	assert.Contains(t, result, img1Path)
}

func TestFindImagesToProcess_DirectoryWithUppercaseExtensions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_find_images")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create image files with uppercase extensions in the directory
	img1Path := filepath.Join(tempDir, "test.JPG")
	img2Path := filepath.Join(tempDir, "test.PNG")

	err = os.WriteFile(img1Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(img2Path, []byte("fake image content"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)

	result, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify all returned paths are to the created image files
	for _, path := range result {
		baseName := filepath.Base(path)
		assert.Contains(t, []string{"test.JPG", "test.PNG"}, baseName)
	}
}

func TestLoadExistingData_ValidJsonFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_load_data")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create index.json file with data
	indexJsonPath := filepath.Join(tempDir, "index.json")
	data := map[string]interface{}{
		"image1.jpg": map[string]interface{}{
			"short_name": "image1",
		},
	}
	content, _ := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile(indexJsonPath, content, 0644)
	assert.NoError(t, err)

	cfg := &config.Config{}
	fs := NewFileScanner(cfg)

	result, err := fs.LoadExistingData(indexJsonPath)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "image1.jpg")
}

func TestLoadExistingData_InvalidJsonFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_load_data")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create index.json file with invalid JSON content
	indexJsonPath := filepath.Join(tempDir, "index.json")
	err = os.WriteFile(indexJsonPath, []byte("invalid json content"), 0644)
	assert.NoError(t, err)

	cfg := &config.Config{}
	fs := NewFileScanner(cfg)

	result, err := fs.LoadExistingData(indexJsonPath)
	assert.NoError(t, err)
	// Should return empty map when JSON is invalid
	assert.Empty(t, result)
}

func TestFileScanner_FilterExcludedFiles_Simple(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_filter_simple")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create test files including some that should be excluded
	img1Path := filepath.Join(tempDir, "test.jpg")
	tmpPath := filepath.Join(tempDir, "temp_file.tmp")
	bakPath := filepath.Join(tempDir, "backup.bak")

	// Create the files
	os.WriteFile(img1Path, []byte("fake image content"), 0644)
	os.WriteFile(tmpPath, []byte("temp file content"), 0644)
	os.WriteFile(bakPath, []byte("backup file content"), 0644)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png"},
		ExcludeFilter:       []string{"**/*.tmp", "**/*.bak"},
	}
	fs := NewFileScanner(cfg)

	// Test that we can create a FileScanner with filters
	assert.NotNil(t, fs)
	assert.Equal(t, cfg, fs.config)

	files := []string{img1Path, tmpPath, bakPath}
	filtered := fs.FilterExcludedFiles(files)

	assert.NotNil(t, filtered)
	assert.Len(t, filtered, 1)
	assert.Contains(t, filtered, img1Path)
}

func TestFileScanner_FilterWithEmptyConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_filter_empty")
	assert.NoError(t, err)
	defer cleanupFileScannerTestDir(t, tempDir)

	// Create test files
	img1Path := filepath.Join(tempDir, "test.jpg")
	tmpPath := filepath.Join(tempDir, "temp_file.tmp")

	// Create the files
	os.WriteFile(img1Path, []byte("fake image content"), 0644)
	os.WriteFile(tmpPath, []byte("temp file content"), 0644)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png"},
		ExcludeFilter:       []string{}, // Empty exclude
	}
	fs := NewFileScanner(cfg)

	// Should not panic and should create a valid scanner
	assert.NotNil(t, fs)

	// Test that it doesn't crash with empty exclude
	files := []string{img1Path, tmpPath}
	filtered := fs.FilterExcludedFiles(files)

	// With no filters, all files should be returned
	assert.Len(t, filtered, 2)
}

// Test helpers to create test directories and files
func cleanupFileScannerTestDir(t *testing.T, dirPath string) {
	err := os.RemoveAll(dirPath)
	assert.NoError(t, err)
}
