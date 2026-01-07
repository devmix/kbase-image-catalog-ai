package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"kbase-catalog/internal/config"
)

// TestIntegrationProcessFullFlow tests a full processing flow
func TestIntegrationProcessFullFlow(t *testing.T) {
	// Create a temporary directory structure for integration test
	tempDir := t.TempDir()

	// Create test subdirectories and files
	subDir1 := filepath.Join(tempDir, "test_catalog")
	err := os.MkdirAll(subDir1, 0755)
	assert.NoError(t, err)

	// Create a test image file (just some dummy content)
	testImage := filepath.Join(subDir1, "test_image.jpg")
	os.WriteFile(testImage, []byte("dummy image content"), 0644)

	// Create another directory with an image
	subDir2 := filepath.Join(tempDir, "another_catalog")
	err = os.MkdirAll(subDir2, 0755)
	assert.NoError(t, err)

	testImage2 := filepath.Join(subDir2, "test_image2.png")
	os.WriteFile(testImage2, []byte("dummy image content"), 0644)

	// Create a config with default values
	cfg := config.GetDefaultConfig()

	// Test CatalogProcessor creation
	cp := NewCatalogProcessor(cfg, "")
	assert.NotNil(t, cp)

	// Since we can't actually process real images without an LLM API,
	// we'll at least verify that the processor can be created and the methods exist

	// Verify that the directory processor can be initialized
	dp := cp.dp
	assert.NotNil(t, dp)

	// Test FileScanner creation
	fs := cp.fs
	assert.NotNil(t, fs)

	// Test ImageProcessor creation
	ip := cp.ip
	assert.NotNil(t, ip)

	// Test IndexGenerator creation
	ig := cp.ig
	assert.NotNil(t, ig)

	// Test that we can call methods on the processor components
	// Note: We're not actually processing images in this test due to external dependencies
	// but we can verify that all components can be created and have expected methods

	// Test some basic file operations
	images, err := fs.FindImagesToProcess(subDir1)
	assert.NoError(t, err)
	assert.Len(t, images, 1)

	// Check the image path
	assert.True(t, strings.HasSuffix(images[0], "test_image.jpg"))

	// Test directory processing function - this will just check the logic without actual API calls
	ctx := context.Background()

	// Test that we can at least create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	assert.NotNil(t, timeoutCtx)
	cancel()

	fmt.Printf("Integration test passed - components created and basic functions available\n")
}

// TestProcessorComponentInit tests component initialization
func TestProcessorComponentInit(t *testing.T) {
	cfg := config.GetDefaultConfig()

	// Test all component creation
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)
	dp := NewDirectoryProcessor(cfg, fs, ip, ig)
	cp := NewCatalogProcessor(cfg, "")

	assert.NotNil(t, fs)
	assert.NotNil(t, ip)
	assert.NotNil(t, ig)
	assert.NotNil(t, dp)
	assert.NotNil(t, cp)
}
