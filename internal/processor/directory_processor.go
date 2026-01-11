package processor

import (
	"context"
	"fmt"
	"kbase-catalog/internal/utils"
	"os"
	"path/filepath"
	"sync"
	"time"

	"kbase-catalog/internal/config"
)

// DirectoryProcessor handles processing of individual directories
type DirectoryProcessor struct {
	config *config.Config
	mutex  sync.RWMutex
	fs     *FileScanner
	ip     *ImageProcessor
	ig     *IndexGenerator
}

// NewDirectoryProcessor creates a new instance of DirectoryProcessor
func NewDirectoryProcessor(cfg *config.Config, fs *FileScanner, ip *ImageProcessor, ig *IndexGenerator) *DirectoryProcessor {
	return &DirectoryProcessor{
		config: cfg,
		fs:     fs,
		ip:     ip,
		ig:     ig,
	}
}

// ProcessDirectory processes all images in a directory
func (dp *DirectoryProcessor) ProcessDirectory(ctx context.Context, dirPath string) (map[string]interface{}, error) {
	fmt.Printf("Processing directory: %s\n", dirPath)

	indexJsonPath := filepath.Join(dirPath, "index.json")
	indexMdPath := filepath.Join(dirPath, "index.md")

	currentData, err := dp.fs.LoadExistingData(indexJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing data: %w", err)
	}

	imagesToProcess, err := dp.fs.FindImagesToProcess(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find images: %w", err)
	}

	if len(imagesToProcess) == 0 && len(currentData) == 0 {
		return nil, nil
	}

	// Find all files that exist in the directory
	existingFiles := make(map[string]bool)
	for _, imgPath := range imagesToProcess {
		if imgPath == "index.json" || imgPath == "index.md" {
			continue
		}
		baseName := filepath.Base(imgPath)
		existingFiles[baseName] = true
	}

	// Remove entries from currentData for files that no longer exist
	hasChanges := false
	for key := range currentData {
		// Skip index files (they're not images)
		if key == "index.json" || key == "index.md" {
			continue
		}

		// If the file doesn't exist anymore, remove it from data
		if !existingFiles[key] {
			delete(currentData, key)
			hasChanges = true
		}
	}

	// Process new or updated images
	if len(imagesToProcess) != 0 {
		if dp.config.ParallelRequests > 1 {
			hasChanges, err = dp.processImagesParallel(ctx, imagesToProcess, currentData)
			if err != nil {
				return nil, fmt.Errorf("failed to process images in parallel: %w", err)
			}
		} else {
			for _, imgPath := range imagesToProcess {
				if imgPath == "index.json" || imgPath == "index.md" {
					continue
				}

				processed, err := dp.ip.ProcessSingleImage(ctx, imgPath, currentData)
				if err != nil {
					fmt.Printf("Error processing image %s: %v\n", imgPath, err)
					continue
				}
				if processed {
					hasChanges = true
				}
			}
		}
	}

	// Save index files only if we have data to save or if there was a change
	if hasChanges || !utils.IsFileExists(indexJsonPath) {
		// If no images exist in directory, remove the index files
		if len(currentData) == 0 {
			// Remove old files if they exist
			if utils.IsFileExists(indexJsonPath) {
				os.Remove(indexJsonPath)
			}
			if utils.IsFileExists(indexMdPath) {
				os.Remove(indexMdPath)
			}
			return nil, nil
		}
	}

	if err := dp.saveIndexJson(indexJsonPath, currentData); err != nil {
		return nil, fmt.Errorf("failed to save index.json: %w", err)
	}

	if len(currentData) > 0 {
		// Only regenerate markdown if there's data and index.json exists
		err := dp.generateCatalogIndexAsMarkdown(indexMdPath, currentData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate markdown index: %w", err)
		}
	}

	catalogData := dp.createCatalogData(currentData)

	return catalogData, nil
}

func (dp *DirectoryProcessor) createCatalogData(currentData map[string]interface{}) map[string]interface{} {
	if len(currentData) == 0 {
		return nil
	}
	catalogData := make(map[string]interface{})
	catalogData["image_count"] = len(currentData)
	lastUpdate := time.Now()
	for _, value := range currentData {
		if meta, ok := value.(map[string]interface{}); !ok {
			currentDate := meta["update_date"]
			if currentDate == nil {
				continue
			}
			if imageUpdated, err := time.Parse(time.RFC3339, currentDate.(string)); err == nil {
				if lastUpdate.Unix() < imageUpdated.Unix() {
					lastUpdate = imageUpdated
				}
			}
		}
	}
	catalogData["last_update"] = lastUpdate.Format(time.RFC3339)
	return catalogData
}

// processImagesParallel processes images in parallel
func (dp *DirectoryProcessor) processImagesParallel(ctx context.Context, imagesToProcess []string, currentData map[string]interface{}) (bool, error) {
	if len(imagesToProcess) == 0 {
		return false, nil
	}

	// Validate config parameter
	if dp.config.ParallelRequests <= 0 {
		return false, fmt.Errorf("invalid ParallelRequests configuration: %d", dp.config.ParallelRequests)
	}

	fmt.Printf("Processing %d images in parallel (max %d concurrent requests)\n", len(imagesToProcess), dp.config.ParallelRequests)

	var filteredImages []string
	for _, imgPath := range imagesToProcess {
		if dp.needsProcessing(currentData, imgPath) {
			filteredImages = append(filteredImages, imgPath)
		}
	}

	if len(filteredImages) == 0 {
		return false, nil
	}

	results := make(chan bool, len(filteredImages))
	errors := make(chan error, len(filteredImages))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, dp.config.ParallelRequests)

	for _, imgPath := range filteredImages {
		wg.Add(1)

		// Create a copy of the image path for closure capture
		imgPathCopy := imgPath

		go func(path string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			case semaphore <- struct{}{}:
				defer func() {
					// Ensure we release the semaphore even if goroutine exits early
					select {
					case <-semaphore:
					default:
					}
				}()
			}

			processed, err := dp.ip.ProcessSingleImage(ctx, path, currentData)
			if err != nil {
				errors <- fmt.Errorf("error processing %s: %w", path, err)
				return
			}
			results <- processed
		}(imgPathCopy)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	newFilesFound := false
	for result := range results {
		if result {
			newFilesFound = true
		}
	}

	for err := range errors {
		fmt.Printf("Parallel processing error: %v\n", err)
		newFilesFound = true
	}

	return newFilesFound, nil
}

// needsProcessing checks if an image needs processing
func (dp *DirectoryProcessor) needsProcessing(currentData map[string]interface{}, imgPath string) bool {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	imgKey := filepath.Base(imgPath)
	record, exists := currentData[imgKey]

	if !exists {
		return true
	}

	if recordMap, ok := record.(map[string]interface{}); ok {
		if shortName, ok := recordMap["short_name"].(string); ok && shortName == "error_processing" {
			return true
		}
	}

	return false
}

// saveIndexJson saves the index data to JSON file
func (dp *DirectoryProcessor) saveIndexJson(indexJsonPath string, data map[string]interface{}) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	return dp.ig.SaveIndexJson(indexJsonPath, data)
}

// generateCatalogIndexAsMarkdown generates markdown index from data
func (dp *DirectoryProcessor) generateCatalogIndexAsMarkdown(mdPath string, data map[string]interface{}) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	return dp.ig.GenerateCatalogIndexAsMarkdown(mdPath, data)
}
