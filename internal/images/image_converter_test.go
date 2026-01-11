package images

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"kbase-catalog/internal/config"

	"github.com/stretchr/testify/assert"
)

// TestImageConverter_ConvertImages tests the ConvertImages function
func TestImageConverter_ConvertImages(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a mock image file
	testImagePath := filepath.Join(tempDir, "test_image.png")

	// Create a simple PNG image (10x10 red image)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red image
		}
	}

	// Encode to PNG and write to file
	file, err := os.Create(testImagePath)
	assert.NoError(t, err)
	defer file.Close()

	err = png.Encode(file, img)
	assert.NoError(t, err)

	// Create a temporary directory for origin files
	originDir := filepath.Join(tempDir, "origin")

	// Create config with default settings
	cfg := &config.Config{
		ConvertImageExtensions: []string{".png", ".jpg", ".jpeg"},
	}

	processor := NewImageConverter(cfg)

	t.Run("Successful conversion and move", func(t *testing.T) {
		ctx := context.Background()

		err := processor.ConvertImages(ctx, tempDir, originDir, 80)
		assert.NoError(t, err)

		// Check if WebP file was created
		webpPath := testImagePath[:len(testImagePath)-len(filepath.Ext(testImagePath))] + ".webp"
		_, err = os.Stat(webpPath)
		assert.NoError(t, err, "WebP file should be created")

		// Check if original file was moved
		movedPath := filepath.Join(originDir, filepath.Base(filepath.Dir(testImagePath)), filepath.Base(testImagePath))
		_, err = os.Stat(movedPath)
		assert.NoError(t, err, "Original file should be moved")
	})
}

// TestImageConverter_findImageFiles tests the findImageFiles function
func TestImageConverter_findImageFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create some test image files
	testImage1 := filepath.Join(tempDir, "test1.png")
	testImage2 := filepath.Join(tempDir, "test2.jpg")
	testImage3 := filepath.Join(tempDir, "test3.txt")

	// Create simple files
	os.WriteFile(testImage1, []byte("png content"), 0644)
	os.WriteFile(testImage2, []byte("jpg content"), 0644)
	os.WriteFile(testImage3, []byte("txt content"), 0644)

	// Create config with default settings
	cfg := &config.Config{
		ConvertImageExtensions: []string{".png", ".jpg"},
	}

	processor := NewImageConverter(cfg)

	t.Run("Should find only image files with correct extensions", func(t *testing.T) {
		files, err := processor.findImageFiles(tempDir)
		assert.NoError(t, err)
		assert.Len(t, files, 2)
		assert.Contains(t, files, testImage1)
		assert.Contains(t, files, testImage2)
	})
}
