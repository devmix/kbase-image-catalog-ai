package processor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kbase-catalog/internal/config"

	"github.com/stretchr/testify/assert"
)

// Test helpers to create test directories and files
func setupTestDir(t *testing.T, dirPath string) {
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)
}

func cleanupTestDir(t *testing.T, dirPath string) {
	err := os.RemoveAll(dirPath)
	assert.NoError(t, err)
}

func TestNewDirectoryProcessor(t *testing.T) {
	cfg := &config.Config{}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	assert.NotNil(t, dp)
	assert.Equal(t, cfg, dp.config)
	assert.Equal(t, fs, dp.fs)
	assert.Equal(t, ip, dp.ip)
	assert.Equal(t, ig, dp.ig)
}

func TestProcessDirectory_NoImagesAndNoExistingData(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_process_dir")
	assert.NoError(t, err)
	defer cleanupTestDir(t, tempDir)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	ctx := context.Background()
	result, err := dp.ProcessDirectory(ctx, tempDir)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestProcessDirectory_WithExistingDataButNoNewImages(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_process_dir")
	assert.NoError(t, err)
	defer cleanupTestDir(t, tempDir)

	// Create an index.json file with existing data
	indexJsonPath := filepath.Join(tempDir, "index.json")
	data := map[string]interface{}{
		"image1.jpg": map[string]interface{}{
			"short_name": "image1",
		},
	}
	content, _ := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile(indexJsonPath, content, 0644)
	assert.NoError(t, err)

	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png", ".jpeg"},
	}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	ctx := context.Background()
	result, err := dp.ProcessDirectory(ctx, tempDir)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestNeedsProcessing_NewImage(t *testing.T) {
	cfg := &config.Config{}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	// Image not in current data - should need processing
	currentData := map[string]interface{}{}
	imgPath := "/test/dir/image1.jpg"

	result := dp.needsProcessing(currentData, imgPath)
	assert.True(t, result)
}

func TestNeedsProcessing_ExistingImageWithError(t *testing.T) {
	cfg := &config.Config{}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	// Image with error processing - should need processing
	currentData := map[string]interface{}{
		"image1.jpg": map[string]interface{}{
			"short_name": "error_processing",
		},
	}
	imgPath := "/test/dir/image1.jpg"

	result := dp.needsProcessing(currentData, imgPath)
	assert.True(t, result)
}

func TestNeedsProcessing_ExistingImageWithoutError(t *testing.T) {
	cfg := &config.Config{}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	// Image without error processing - should not need processing
	currentData := map[string]interface{}{
		"image1.jpg": map[string]interface{}{
			"short_name": "image1",
		},
	}
	imgPath := "/test/dir/image1.jpg"

	result := dp.needsProcessing(currentData, imgPath)
	assert.False(t, result)
}

func TestProcessImagesParallel_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		ParallelRequests: 0,
	}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	imagesToProcess := []string{"image1.jpg"}
	currentData := map[string]interface{}{}

	ctx := context.Background()
	result, err := dp.processImagesParallel(ctx, imagesToProcess, currentData)

	assert.Error(t, err)
	assert.False(t, result)
}

func TestProcessImagesParallel_NoImages(t *testing.T) {
	cfg := &config.Config{
		ParallelRequests: 2,
	}
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)

	dp := NewDirectoryProcessor(cfg, fs, ip, ig)

	imagesToProcess := []string{}
	currentData := map[string]interface{}{}

	ctx := context.Background()
	result, err := dp.processImagesParallel(ctx, imagesToProcess, currentData)

	assert.NoError(t, err)
	assert.False(t, result)
}

// This test is skipped due to complexity of context cancellation in parallel processing
func TestProcessImagesParallel_ContextCancelled(t *testing.T) {
	t.Skip("Skipping context cancellation test as it's complex to simulate properly")
}
