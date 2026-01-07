package encoder

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeImageToBase64(t *testing.T) {
	t.Run("Valid PNG image", func(t *testing.T) {
		// Create a temporary PNG file for testing
		tempDir := t.TempDir()
		testImagePath := filepath.Join(tempDir, "test.png")

		// Create a simple PNG image (10x10 red image)
		img := createTestImage(10, 10, 255, 0, 0) // Red image
		err := os.WriteFile(testImagePath, img, 0644)
		assert.NoError(t, err)

		result, err := EncodeImageToBase64(testImagePath)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Check that the result starts with the expected data URI prefix
		assert.Contains(t, result, "data:image/png;base64,")

		// Verify it's valid base64 by attempting to decode it
		decoded, err := decodeBase64String(result)
		assert.NoError(t, err)
		assert.NotEmpty(t, decoded)
	})

	t.Run("Valid JPG image", func(t *testing.T) {
		// Create a temporary JPG file for testing
		tempDir := t.TempDir()
		testImagePath := filepath.Join(tempDir, "test.jpg")

		// Create a simple JPG image (10x10 blue image)
		img := createTestImage(10, 10, 0, 0, 255) // Blue image
		err := os.WriteFile(testImagePath, img, 0644)
		assert.NoError(t, err)

		result, err := EncodeImageToBase64(testImagePath)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Check that the result starts with the expected data URI prefix
		assert.Contains(t, result, "data:image/png;base64,")

		// Verify it's valid base64 by attempting to decode it
		decoded, err := decodeBase64String(result)
		assert.NoError(t, err)
		assert.NotEmpty(t, decoded)
	})

	t.Run("File not found", func(t *testing.T) {
		result, err := EncodeImageToBase64("/non/existent/path/image.png")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to open image file")
	})

	t.Run("Invalid image format", func(t *testing.T) {
		// Create a temporary file with invalid content
		tempDir := t.TempDir()
		testImagePath := filepath.Join(tempDir, "invalid.txt")
		err := os.WriteFile(testImagePath, []byte("not an image"), 0644)
		assert.NoError(t, err)

		result, err := EncodeImageToBase64(testImagePath)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to decode image")
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

// Helper function to decode the base64 part of a data URI
func decodeBase64String(dataURI string) ([]byte, error) {
	// Extract base64 part from data URI (after the comma)
	parts := strings.Split(dataURI, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URI format")
	}

	return base64.StdEncoding.DecodeString(parts[1])
}
