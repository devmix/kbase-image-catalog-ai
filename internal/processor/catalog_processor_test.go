package processor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/llm"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryProcessor_NeedsProcessing(t *testing.T) {
	t.Run("New file should need processing", func(t *testing.T) {
		dp := &DirectoryProcessor{}
		currentData := make(map[string]interface{})

		result := dp.needsProcessing(currentData, "/test/image.jpg")
		assert.True(t, result)
	})

	t.Run("File with no error processing should be processed", func(t *testing.T) {
		dp := &DirectoryProcessor{}
		currentData := map[string]interface{}{
			"image.jpg": map[string]interface{}{
				"short_name":  "Test Image",
				"description": "This is a test image",
			},
		}

		result := dp.needsProcessing(currentData, "/test/image.jpg")
		assert.False(t, result)
	})

	t.Run("File with error processing should be reprocessed", func(t *testing.T) {
		dp := &DirectoryProcessor{}
		currentData := map[string]interface{}{
			"image.jpg": map[string]interface{}{
				"short_name":  "error_processing",
				"description": "Error processing file",
			},
		}

		result := dp.needsProcessing(currentData, "/test/image.jpg")
		assert.True(t, result)
	})
}

func TestImageProcessor_NeedsProcessing(t *testing.T) {
	t.Run("New file should need processing", func(t *testing.T) {
		ip := &ImageProcessor{}
		currentData := make(map[string]interface{})

		result := ip.needsProcessing(currentData, "/test/image.jpg")
		assert.True(t, result)
	})

	t.Run("File with no error processing should be processed", func(t *testing.T) {
		ip := &ImageProcessor{}
		currentData := map[string]interface{}{
			"image.jpg": map[string]interface{}{
				"short_name":  "Test Image",
				"description": "This is a test image",
			},
		}

		result := ip.needsProcessing(currentData, "/test/image.jpg")
		assert.False(t, result)
	})

	t.Run("File with error processing should be reprocessed", func(t *testing.T) {
		ip := &ImageProcessor{}
		currentData := map[string]interface{}{
			"image.jpg": map[string]interface{}{
				"short_name":  "error_processing",
				"description": "Error processing file",
			},
		}

		result := ip.needsProcessing(currentData, "/test/image.jpg")
		assert.True(t, result)
	})
}

func TestImageProcessor_ValidateResponse(t *testing.T) {
	t.Run("Valid response should return true", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "Test Image",
			Description: "This is a test image",
		}

		result := ValidateResponse(response)
		assert.True(t, result)
	})

	t.Run("Empty short name should return false", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "",
			Description: "This is a test image",
		}

		result := ValidateResponse(response)
		assert.False(t, result)
	})

	t.Run("Empty description should return false", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "Test Image",
			Description: "",
		}

		result := ValidateResponse(response)
		assert.False(t, result)
	})

	t.Run("Both empty should return false", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "",
			Description: "",
		}

		result := ValidateResponse(response)
		assert.False(t, result)
	})
}

func TestCatalogProcessor_NewCatalogProcessor(t *testing.T) {
	config := config.GetDefaultConfig()

	cp := NewCatalogProcessor(config, "/test/archive")

	assert.NotNil(t, cp)
	assert.NotNil(t, cp.config)
	assert.NotNil(t, cp.dp)
	assert.NotNil(t, cp.fs)
	assert.NotNil(t, cp.ip)
	assert.NotNil(t, cp.ig)
	assert.Equal(t, "/test/archive", cp.archiveDir)
}

func TestCatalogProcessor_ShouldExclude(t *testing.T) {
	t.Run("Should handle empty exclude filter", func(t *testing.T) {
		config := &config.Config{
			SupportedExtensions: []string{".jpg", ".png"},
			ExcludeFilter:       []string{},
		}

		cp := NewCatalogProcessor(config, "/test/archive")

		// With no filters, nothing should be excluded
		assert.False(t, cp.ShouldExclude("/any/path/temp"))
		assert.False(t, cp.ShouldExclude("/any/path/.git"))
	})
}

func TestCatalogProcessor_RebuildRootIndex(t *testing.T) {
	t.Run("Should handle empty directory", func(t *testing.T) {
		config := config.GetDefaultConfig()
		cp := NewCatalogProcessor(config, t.TempDir())

		// Test with an empty directory - should not error
		ctx := context.Background()
		err := cp.RebuildRootIndex(ctx)
		assert.NoError(t, err)
	})
}

func TestFileScanner_FindImagesToProcess(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test files
	img1Path := filepath.Join(tempDir, "test.jpg")
	img2Path := filepath.Join(tempDir, "test.png")
	img3Path := filepath.Join(tempDir, "test.txt")

	// Write some content to the files
	os.WriteFile(img1Path, []byte("image data"), 0644)
	os.WriteFile(img2Path, []byte("image data"), 0644)
	os.WriteFile(img3Path, []byte("text data"), 0644)

	// Create an index.json file to make sure it's filtered out
	indexJsonPath := filepath.Join(tempDir, "index.json")
	os.WriteFile(indexJsonPath, []byte("{}"), 0644)

	fs := NewFileScanner(config.GetDefaultConfig())

	images, err := fs.FindImagesToProcess(tempDir)
	assert.NoError(t, err)
	assert.Len(t, images, 2) // Only jpg and png files

	// Check that the files are in the right order (by name)
	assert.True(t, strings.HasSuffix(images[0], "test.png"))
	assert.True(t, strings.HasSuffix(images[1], "test.jpg"))
}

func TestFileScanner_LoadExistingData(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test index.json file
	indexJsonPath := filepath.Join(tempDir, "index.json")
	data := `{
		"image1.jpg": {
			"short_name": "Test Image 1",
			"description": "This is test image 1"
		},
		"image2.jpg": {
			"short_name": "Test Image 2", 
			"description": "This is test image 2"
		}
	}`

	os.WriteFile(indexJsonPath, []byte(data), 0644)

	fs := NewFileScanner(config.GetDefaultConfig())

	result, err := fs.LoadExistingData(indexJsonPath)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Check that the data was loaded correctly
	img1, ok1 := result["image1.jpg"]
	img2, ok2 := result["image2.jpg"]

	assert.True(t, ok1)
	assert.True(t, ok2)

	assert.Equal(t, "Test Image 1", img1.(map[string]interface{})["short_name"])
	assert.Equal(t, "This is test image 1", img1.(map[string]interface{})["description"])
	assert.Equal(t, "Test Image 2", img2.(map[string]interface{})["short_name"])
	assert.Equal(t, "This is test image 2", img2.(map[string]interface{})["description"])
}

func TestIndexGenerator_SaveIndexJson(t *testing.T) {
	tempDir := t.TempDir()

	indexJsonPath := filepath.Join(tempDir, "test.json")

	data := map[string]interface{}{
		"image1.jpg": map[string]interface{}{
			"short_name":    "Test Image 1",
			"description":   "This is test image 1",
			"original_name": "image1.jpg",
			"vl_model":      "test-model",
			"update_date":   time.Now().Format(time.RFC3339),
		},
	}

	ig := NewIndexGenerator(config.GetDefaultConfig())

	err := ig.SaveIndexJson(indexJsonPath, data)
	assert.NoError(t, err)

	// Check that the file was created
	content, err := os.ReadFile(indexJsonPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Image 1")
	assert.Contains(t, string(content), "This is test image 1")
}

func TestProcessImagesParallel_WithContextCancellation(t *testing.T) {
	// This test will run a short test with context cancellation
	config := config.GetDefaultConfig()

	// Create a mock processor to avoid real processing
	fs := NewFileScanner(config)
	ip := &ImageProcessor{config: config}
	ig := NewIndexGenerator(config)
	dp := NewDirectoryProcessor(config, fs, ip, ig)

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	imagesToProcess := []string{}
	currentData := make(map[string]interface{})

	// Test the parallel processing with a cancelled context
	newFilesFound, err := dp.processImagesParallel(ctx, imagesToProcess, currentData)
	assert.NoError(t, err)
	assert.False(t, newFilesFound)
}

func TestImageProcessor_HandleProcessingError(t *testing.T) {
	ip := &ImageProcessor{}
	currentData := make(map[string]interface{})

	imgPath := "/test/image.jpg"

	ip.handleProcessingError(imgPath, currentData)

	// Check that the error was recorded correctly
	imgKey := filepath.Base(imgPath)
	record, exists := currentData[imgKey]

	assert.True(t, exists)
	assert.Equal(t, "error_processing", record.(map[string]interface{})["short_name"])
	assert.Equal(t, "Error processing file (retry will be attempted)", record.(map[string]interface{})["description"])
}
