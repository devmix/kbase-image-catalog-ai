package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"

	"github.com/stretchr/testify/assert"
)

func TestCatalogService_GetCatalogInfo(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test catalog directory
	catalogPath := filepath.Join(tempDir, "test_catalog")
	err := os.MkdirAll(catalogPath, 0755)
	assert.NoError(t, err)

	// Create some test files including ones that should be excluded
	img1Path := filepath.Join(catalogPath, "test.jpg")
	tmpPath := filepath.Join(catalogPath, "temp_file.tmp")
	bakPath := filepath.Join(catalogPath, "backup.bak")

	// Create the files
	os.WriteFile(img1Path, []byte("fake image content"), 0644)
	os.WriteFile(tmpPath, []byte("temp file content"), 0644)
	os.WriteFile(bakPath, []byte("backup file content"), 0644)

	// Create a config with exclude filters
	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png"},
		ExcludeFilter:       []string{"**/*.tmp", "**/*.bak"},
	}

	// Test that the service can handle exclusion patterns correctly by creating
	// a real processor but not actually using it for processing
	processor := processor.NewCatalogProcessor(cfg, tempDir)

	// Create catalog service
	cs := &CatalogService{
		Config:    cfg,
		Processor: processor,
	}

	// Test that we can call getCatalogInfo without errors
	imageCount, _, err := cs.getCatalogInfo(catalogPath)
	assert.NoError(t, err)

	// Should find 1 image (the jpg file) since tmp and bak files are excluded
	assert.Equal(t, 1, imageCount)
}

func TestCatalogService_GetCatalogs(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test archive directory
	archiveDir := filepath.Join(tempDir, "archive")
	err := os.MkdirAll(archiveDir, 0755)
	assert.NoError(t, err)

	// Create a test catalog directory
	catalogPath := filepath.Join(archiveDir, "test_catalog")
	err = os.MkdirAll(catalogPath, 0755)
	assert.NoError(t, err)

	// Create some test files including ones that should be excluded
	img1Path := filepath.Join(catalogPath, "test.jpg")
	tmpPath := filepath.Join(catalogPath, "temp_file.tmp")

	// Create the files
	os.WriteFile(img1Path, []byte("fake image content"), 0644)
	os.WriteFile(tmpPath, []byte("temp file content"), 0644)

	// Create a config with exclude filters
	cfg := &config.Config{
		SupportedExtensions: []string{".jpg", ".png"},
		ExcludeFilter:       []string{"**/*.tmp"},
	}

	// Test that the service can handle exclusion patterns correctly by creating
	// a real processor but not actually using it for processing
	processor := processor.NewCatalogProcessor(cfg, tempDir)

	// Create catalog service
	cs := &CatalogService{
		Config:     cfg,
		Processor:  processor,
		ArchiveDir: archiveDir,
	}

	// Test that we can call GetCatalogs without errors
	catalogs, err := cs.GetCatalogs(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, catalogs)

	// Should find 1 catalog with 1 image (the jpg file)
	// The temp file should be excluded from the count
	assert.Len(t, catalogs, 1)

	// Check that the catalog has correct information
	catalog := catalogs[0]
	name, ok := catalog["name"].(string)
	assert.True(t, ok)
	assert.Equal(t, "test_catalog", name)
}
