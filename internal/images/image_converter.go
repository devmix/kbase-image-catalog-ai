package images

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"

	"kbase-catalog/internal/config"

	"github.com/chai2010/webp"
)

// ImageConverter handles image conversion to WebP format
type ImageConverter struct {
	config *config.Config
}

// NewImageConverter creates a new instance of ImageConverter
func NewImageConverter(cfg *config.Config) *ImageConverter {
	return &ImageConverter{
		config: cfg,
	}
}

// ConvertImages converts images in the specified directory to WebP format
func (ic *ImageConverter) ConvertImages(ctx context.Context, inputDir, originDir string, quality int) error {
	fmt.Printf("Converting images in: %s\n", inputDir)

	// Find all image files
	imageFiles, err := ic.findImageFiles(inputDir)
	if err != nil {
		return fmt.Errorf("failed to find image files: %w", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("No image files found.")
		return nil
	}

	fmt.Printf("Found %d image files\n", len(imageFiles))

	convertedCount := 0
	movedCount := 0

	for _, imagePath := range imageFiles {
		fmt.Printf("Converting: %s\n", imagePath)

		// Generate output path (replace extension with .webp)
		outputPath := imagePath[:len(imagePath)-len(filepath.Ext(imagePath))] + ".webp"

		// Check if output file already exists
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("  Warning: %s already exists.\n", outputPath)
		} else {
			// Convert image to WebP format
			err = ic.convertToWebP(imagePath, outputPath, quality)
			if err != nil {
				fmt.Printf("  Error converting %s to WebP: %v\n", imagePath, err)
				continue
			}

			fmt.Printf("  Converted to: %s\n", outputPath)
			convertedCount++
		}

		// Move original file
		movedPath, err := ic.moveOriginalFile(imagePath, originDir)
		if err != nil {
			fmt.Printf("Error moving original %s: %v\n", imagePath, err)
			continue
		}

		if movedPath != "" {
			fmt.Printf("  Moved original to: %s\n", movedPath)
			movedCount++
		}
	}

	fmt.Println("\nConversion completed!")
	fmt.Printf("Converted: %d files\n", convertedCount)
	fmt.Printf("Moved originals: %d files\n", movedCount)

	return nil
}

// findImageFiles recursively finds all image files in the root directory
func (ic *ImageConverter) findImageFiles(rootDir string) ([]string, error) {
	var imageFiles []string

	// Use configured extensions or fall back to default if not set
	var extensions map[string]bool
	if len(ic.config.ConvertImageExtensions) > 0 {
		extensions = make(map[string]bool)
		for _, ext := range ic.config.ConvertImageExtensions {
			extensions[ext] = true
		}
	} else {
		return imageFiles, nil
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has a supported extension
		ext := strings.ToLower(filepath.Ext(path))
		if extensions[ext] {
			imageFiles = append(imageFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return imageFiles, nil
}

// moveOriginalFile moves the original file to the origin directory structure
func (ic *ImageConverter) moveOriginalFile(originalPath, originDir string) (string, error) {
	// Get parent directory name
	parentDir := filepath.Base(filepath.Dir(originalPath))

	// Create destination path
	destinationDir := filepath.Join(originDir, parentDir)
	err := os.MkdirAll(destinationDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move file using os.Rename (which is the equivalent of shutil.move in Python)
	destinationPath := filepath.Join(destinationDir, filepath.Base(originalPath))

	fmt.Printf("  Moving original to: %s\n", destinationPath)

	// Try to use os.Rename first (fastest method)
	err = os.Rename(originalPath, destinationPath)
	if err != nil {
		// If rename fails due to cross-device link error, copy and remove
		if isCrossDeviceError(err) {
			fmt.Printf("  Cross-device link detected. Copying instead of moving.\n")

			// Copy the file
			err = copyFile(originalPath, destinationPath)
			if err != nil {
				return "", fmt.Errorf("failed to copy original file: %w", err)
			}

			// Remove the original file after successful copy
			err = os.Remove(originalPath)
			if err != nil {
				return "", fmt.Errorf("failed to remove original file after copying: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to move original file: %w", err)
		}
	}

	return destinationPath, nil
}

// isCrossDeviceError checks if an error is a cross-device link error
func isCrossDeviceError(err error) bool {
	if err == nil {
		return false
	}

	// Check for the specific error message that indicates cross-device link issue
	errStr := err.Error()
	return strings.Contains(errStr, "invalid cross-device link") ||
		strings.Contains(errStr, "cross-device link not permitted")
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get the file permissions from source
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create destination file with same permissions
	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file content
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// convertToWebP converts an image file to WebP format
func (ic *ImageConverter) convertToWebP(inputPath, outputPath string, quality int) error {
	// Open the input image file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode the input image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Open the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode the image as WebP
	err = webp.Encode(outFile, img, &webp.Options{Quality: float32(quality)})
	if err != nil {
		return fmt.Errorf("failed to encode WebP: %w", err)
	}

	return nil
}
