package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/llm"

	"github.com/stretchr/testify/assert"
)

// TestImageProcessor_ProcessSingleImage tests the ProcessSingleImage function
func TestImageProcessor_ProcessSingleImage(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a mock image file
	testImagePath := filepath.Join(tempDir, "test_image.png")

	// Create a simple PNG image (10x10 red image)
	imgData := createTestImage(10, 10, 255, 0, 0) // Red image
	err := os.WriteFile(testImagePath, imgData, 0644)
	assert.NoError(t, err)

	// Create a mock server for LLM API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make(map[string]interface{})
		json.NewDecoder(r.Body).Decode(&body)

		// Mock successful response with valid JSON
		response := map[string]interface{}{
			"model": "test-model",
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"short_name": "Test Image", "description": "This is a test image."}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create config with mock API URL
	cfg := &config.Config{
		APIURL:  server.URL,
		Model:   "test-model",
		Timeout: 10,
		SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
	}

	processor := NewImageProcessor(cfg)

	t.Run("Successful processing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		currentData := make(map[string]interface{})

		processed, err := processor.ProcessSingleImage(ctx, testImagePath, currentData)
		assert.NoError(t, err)
		assert.True(t, processed)
		assert.Contains(t, currentData, "test_image.png")

		record := currentData["test_image.png"].(map[string]interface{})
		assert.Equal(t, "Test Image", record["short_name"])
		assert.Equal(t, "This is a test image.", record["description"])
	})

	t.Run("Should not process if already exists with valid data", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		currentData := make(map[string]interface{})
		currentData["test_image.png"] = map[string]interface{}{
			"short_name":    "Test Image",
			"description":   "This is a test image.",
			"original_name": "test_image.png",
			"vl_model":      "test-model",
			"update_date":   time.Now().Format(time.RFC3339),
		}

		processed, err := processor.ProcessSingleImage(ctx, testImagePath, currentData)
		assert.NoError(t, err)
		assert.False(t, processed) // Should not process since it already exists with valid data
	})

	t.Run("Should process if already exists but has error_processing status", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		currentData := make(map[string]interface{})
		currentData["test_image.png"] = map[string]interface{}{
			"short_name":    "error_processing",
			"description":   "Error processing file (retry will be attempted)",
			"original_name": "test_image.png",
			"vl_model":      "unknown",
			"update_date":   time.Now().Format(time.RFC3339),
		}

		processed, err := processor.ProcessSingleImage(ctx, testImagePath, currentData)
		assert.NoError(t, err)
		assert.True(t, processed) // Should process since it has error_processing status
	})
}

// TestImageProcessor_needsProcessing tests the needsProcessing function
func TestImageProcessor_needsProcessing(t *testing.T) {
	t.Run("Should need processing if file doesn't exist in data", func(t *testing.T) {
		currentData := make(map[string]interface{})
		result := NeedsProcessing(currentData, "/path/to/image.png")
		assert.True(t, result)
	})

	t.Run("Should not need processing if file exists with valid data", func(t *testing.T) {
		currentData := make(map[string]interface{})
		currentData["image.png"] = map[string]interface{}{
			"short_name":    "Test Image",
			"description":   "This is a test image.",
			"original_name": "image.png",
			"vl_model":      "test-model",
			"update_date":   time.Now().Format(time.RFC3339),
		}

		result := NeedsProcessing(currentData, "/path/to/image.png")
		assert.False(t, result)
	})

	t.Run("Should need processing if file exists but has error_processing status", func(t *testing.T) {
		currentData := make(map[string]interface{})
		currentData["image.png"] = map[string]interface{}{
			"short_name":    "error_processing",
			"description":   "Error processing file (retry will be attempted)",
			"original_name": "image.png",
			"vl_model":      "unknown",
			"update_date":   time.Now().Format(time.RFC3339),
		}

		result := NeedsProcessing(currentData, "/path/to/image.png")
		assert.True(t, result)
	})

	t.Run("Should need processing with invalid data type", func(t *testing.T) {
		currentData := make(map[string]interface{})
		currentData["image.png"] = "invalid_data_type"

		result := NeedsProcessing(currentData, "/path/to/image.png")
		assert.True(t, result)
	})
}

// TestImageProcessor_validateResponse tests the validateResponse function
func TestImageProcessor_validateResponse(t *testing.T) {
	t.Run("Should validate valid response", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "Test Image",
			Description: "This is a test image.",
		}

		result := ValidateResponse(response)
		assert.True(t, result)
	})

	t.Run("Should not validate response with empty short_name", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "",
			Description: "This is a test image.",
		}

		result := ValidateResponse(response)
		assert.False(t, result)
	})

	t.Run("Should not validate response with empty description", func(t *testing.T) {
		response := &llm.LLMResponse{
			ShortName:   "Test Image",
			Description: "",
		}

		result := ValidateResponse(response)
		assert.False(t, result)
	})

	t.Run("Should not validate nil response", func(t *testing.T) {
		var response *llm.LLMResponse

		result := ValidateResponse(response)
		assert.False(t, result)
	})
}

// TestImageProcessor_handleProcessingError tests the handleProcessingError function
func TestImageProcessor_handleProcessingError(t *testing.T) {
	t.Run("Should properly handle processing error", func(t *testing.T) {
		currentData := make(map[string]interface{})
		imgPath := "/path/to/test_image.png"

		HandleProcessingError(imgPath, currentData)

		assert.Contains(t, currentData, "test_image.png")

		record := currentData["test_image.png"].(map[string]interface{})
		assert.Equal(t, "error_processing", record["short_name"])
		assert.Equal(t, "Error processing file (retry will be attempted)", record["description"])
		assert.Equal(t, "unknown", record["vl_model"])
	})
}

// TestImageProcessor_TestSingleImage tests the TestSingleImage function
func TestImageProcessor_TestSingleImage(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a mock image file
	testImagePath := filepath.Join(tempDir, "test_image.png")

	// Create a simple PNG image (10x10 red image)
	imgData := createTestImage(10, 10, 255, 0, 0) // Red image
	err := os.WriteFile(testImagePath, imgData, 0644)
	assert.NoError(t, err)

	// Create a mock server for LLM API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make(map[string]interface{})
		json.NewDecoder(r.Body).Decode(&body)

		// Mock successful response with valid JSON
		response := map[string]interface{}{
			"model": "test-model",
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"short_name": "Test Image", "description": "This is a test image."}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create config with mock API URL
	cfg := &config.Config{
		APIURL:  server.URL,
		Model:   "test-model",
		Timeout: 10,
		SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
	}

	processor := NewImageProcessor(cfg)

	t.Run("Successful test", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := processor.TestSingleImage(ctx, testImagePath)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Test Image", response.ShortName)
		assert.Equal(t, "This is a test image.", response.Description)
	})

	t.Run("File not found error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := processor.TestSingleImage(ctx, "/non/existent/path/image.png")
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "file not found")
	})
}

// Helper function to create a simple test image
func createTestImage(width, height int, r, g, b uint8) []byte {
	// Create a simple image with specified color
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
