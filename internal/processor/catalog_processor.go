package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/llm"
	"kbase-catalog/internal/utils"
)

// CatalogProcessor handles processing of the catalog directory structure
type CatalogProcessor struct {
	config     *config.Config
	dp         *DirectoryProcessor
	fs         *FileScanner
	ip         *ImageProcessor
	ig         *IndexGenerator
	archiveDir string
}

// NewCatalogProcessor creates a new instance of CatalogProcessor
func NewCatalogProcessor(cfg *config.Config, archiveDir string) *CatalogProcessor {
	fs := NewFileScanner(cfg)
	ip := NewImageProcessor(cfg)
	ig := NewIndexGenerator(cfg)
	return &CatalogProcessor{
		config:     cfg,
		dp:         NewDirectoryProcessor(cfg, fs, ip, ig),
		fs:         fs,
		ip:         ip,
		ig:         ig,
		archiveDir: archiveDir,
	}
}

// ProcessImagesCatalog processes images in the single catalog directory
func (cp *CatalogProcessor) ProcessImagesCatalog(ctx context.Context, catalogDir string) error {
	fmt.Printf("Starting scan in: %s\n", catalogDir)

	if cp.fs.ShouldExclude(catalogDir) {
		return nil
	}

	fmt.Printf("\n--> Processing directory: %s\n", strings.TrimPrefix(catalogDir, catalogDir+"/"))

	data, err := cp.dp.ProcessDirectory(ctx, catalogDir)
	if err != nil {
		return fmt.Errorf("Error processing directory %s: %v\n", catalogDir, err)
	}

	err = cp.mergeWithRooIndex(catalogDir, err, data)
	if err != nil {
		return fmt.Errorf("Error merging with root index: %v\n", err)
	}

	return nil
}

// mergeWithRooIndex merges catalog data with the root index
func (cp *CatalogProcessor) mergeWithRooIndex(catalogDir string, err error, data map[string]interface{}) error {
	// Load existing root index data
	rootIndexPath := filepath.Join(cp.archiveDir, "index.json")
	var catalogData map[string]interface{}
	if utils.IsFileExists(rootIndexPath) {
		catalogData, err = cp.fs.LoadExistingData(rootIndexPath)
		if err != nil {
			return fmt.Errorf("failed to load existing data: %v", err)
		}
	} else {
		catalogData = make(map[string]interface{})
	}

	catalogName := filepath.Base(catalogDir)

	catalogData[catalogName] = data

	// Generate the global index with updated information
	err = cp.ig.GenerateGlobalJsonIndex(cp.archiveDir, catalogData)
	if err != nil {
		fmt.Printf("Warning: Failed to update root index: %v\n", err)
	}

	// Also update markdown index if needed
	err = cp.ig.GenerateGlobalMarkdownIndex(cp.archiveDir, catalogData)
	if err != nil {
		fmt.Printf("Warning: Failed to update root markdown index: %v\n", err)
	}
	return nil
}

// RebuildRootIndex rebuilds the root index.json file that aggregates all catalogs
func (cp *CatalogProcessor) RebuildRootIndex(ctx context.Context) error {
	rootPath := cp.archiveDir

	fmt.Printf("Rebuilding root index in: %s\n", rootPath)

	catalogData := make(map[string]interface{})

	err := cp.readCatalogDirectories(rootPath, catalogData)
	if err != nil {
		return fmt.Errorf("failed to read catalog directories: %w", err)
	}

	// Generate the global index
	err = cp.ig.GenerateGlobalJsonIndex(rootPath, catalogData)
	if err != nil {
		return fmt.Errorf("failed to generate global index: %w", err)
	}

	// Generate the global markdown
	err = cp.ig.GenerateGlobalMarkdownIndex(rootPath, catalogData)
	if err != nil {
		return fmt.Errorf("failed to generate global index: %w", err)
	}

	fmt.Printf("Root index rebuilt successfully\n")

	return nil
}

// readCatalogDirectories recursively reads directories and collects catalog data
func (cp *CatalogProcessor) readCatalogDirectories(rootPath string, catalogData map[string]interface{}) error {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip if it's the root path itself
		if entry.Name() == "" {
			continue
		}

		path := filepath.Join(rootPath, entry.Name())

		// Skip excluded paths
		if cp.fs.ShouldExclude(path) {
			continue
		}

		// Only process directories
		if !entry.IsDir() {
			continue
		}

		// Look for index.json in the directory to get catalog metadata
		indexJsonPath := filepath.Join(path, "index.json")
		if !utils.IsFileExists(indexJsonPath) {
			// Directory doesn't have an index.json, skip it
			continue
		}

		// Load the existing data from index.json
		data, err := cp.fs.LoadExistingData(indexJsonPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load index.json for %s: %v\n", path, err)
			continue
		}

		// Extract catalog information from the data
		catalogName := entry.Name()
		if len(data) > 0 {
			// Get the first entry to extract metadata (we don't actually use it, but it's here for completeness)
			for _, value := range data {
				if _, ok := value.(map[string]interface{}); ok {
					break
				}
			}

			catalogInfo := make(map[string]interface{})

			// Add basic info
			catalogInfo["name"] = catalogName
			catalogInfo["image_count"] = len(data)

			// Get last update time if available
			lastUpdate := time.Now()
			for _, value := range data {
				if meta, ok := value.(map[string]interface{}); ok {
					if currentDate, exists := meta["update_date"]; exists {
						if imageUpdated, err := time.Parse(time.RFC3339, currentDate.(string)); err == nil {
							if lastUpdate.Unix() < imageUpdated.Unix() {
								lastUpdate = imageUpdated
							}
						}
					}
				}
			}
			catalogInfo["last_update"] = lastUpdate.Format(time.RFC3339)

			// Add the catalog info to our main map
			catalogData[catalogName] = catalogInfo
		} else {
			// Empty directory, add basic info
			catalogInfo := map[string]interface{}{
				"name":        catalogName,
				"image_count": 0,
				"last_update": time.Now().Format(time.RFC3339),
			}
			catalogData[catalogName] = catalogInfo
		}
	}

	return nil
}

func (cp *CatalogProcessor) TestSingleImage(ctx context.Context, imagePath string) (*llm.LLMResponse, error) {
	return cp.ip.TestSingleImage(ctx, imagePath)
}

func (cp *CatalogProcessor) ShouldExclude(path string) bool {
	return cp.fs.ShouldExclude(path)
}

func (cp *CatalogProcessor) ProcessCatalog(ctx context.Context) error {
	rootPath := cp.archiveDir

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		catalogName := entry.Name()
		if catalogName == "" || !entry.IsDir() {
			continue
		}

		path := filepath.Join(rootPath, catalogName)

		if err := cp.ProcessImagesCatalog(ctx, path); err != nil {
			log.Printf("Failed to reindex catalog %s: %v", catalogName, err)
		} else {
			log.Printf("Successfully reindexed catalog %s", catalogName)
		}
	}

	return nil
}
