package services

import (
	"context"
	"encoding/json"
	"fmt"
	"kbase-catalog/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
)

// CatalogService handles catalog operations for the web server
type CatalogService struct {
	Config     *config.Config
	Processor  *processor.CatalogProcessor
	ArchiveDir string
}

// GetCatalogs returns list of all catalogs with extra information
func (cs *CatalogService) GetCatalogs(ctx context.Context) ([]map[string]interface{}, error) {
	catalogs := []map[string]interface{}{}
	archiveDir := cs.ArchiveDir

	if archiveDir == "" {
		archiveDir = "archive"
	}

	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		// If directory doesn't exist, create it and return empty list
		os.MkdirAll(archiveDir, 0755)
		return catalogs, nil
	}

	// First try to read the global index.json if it exists
	globalIndexPath := filepath.Join(archiveDir, "index.json")
	if utils.IsFileExists(globalIndexPath) {
		data, err := os.ReadFile(globalIndexPath)
		if err == nil {
			var globalIndexData map[string]interface{}
			if err := json.Unmarshal(data, &globalIndexData); err == nil {
				// Convert the global index data to the format expected by GetCatalogs
				for catalogName, catalogInfo := range globalIndexData {
					if catalogInfoMap, ok := catalogInfo.(map[string]interface{}); ok {
						catalogs = append(catalogs, map[string]interface{}{
							"name":       catalogName,
							"imageCount": int(catalogInfoMap["image_count"].(float64)),
							"lastUpdate": catalogInfoMap["last_update"],
						})
					}
				}
				return catalogs, nil
			}
		}
	}

	// If global index doesn't exist or has issues, fall back to the old method
	return cs.getCatalogsFallback(ctx)
}

// getCatalogsFallback is the original method for backward compatibility
func (cs *CatalogService) getCatalogsFallback(ctx context.Context) ([]map[string]interface{}, error) {
	catalogs := []map[string]interface{}{}
	archiveDir := cs.ArchiveDir

	// If directory doesn't exist, create it and return empty list
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		return catalogs, nil
	}

	// Read all subdirectories in archive
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, fmt.Errorf("error reading archive directory: %w", err)
	}

	for _, entry := range entries {
		// Skip the root directory itself and non-directories
		if !entry.IsDir() || entry.Name() == "." || entry.Name() == ".." {
			continue
		}

		path := filepath.Join(archiveDir, entry.Name())

		// Get image count and last update date
		imageCount, lastUpdate, err := cs.getCatalogInfo(path)
		if err != nil {
			// Log error but continue processing other catalogs
			fmt.Printf("Error getting catalog info for %s: %v\n", entry.Name(), err)
			continue // Continue with other catalogs even if one fails
		}

		if imageCount == 0 {
			continue // Skip empty catalogs or those with errors
		}

		catalogs = append(catalogs, map[string]interface{}{
			"name":       entry.Name(),
			"imageCount": imageCount,
			"lastUpdate": lastUpdate,
		})
	}

	return catalogs, nil
}

// GetCatalogImages returns all images in a catalog with their metadata
func (cs *CatalogService) GetCatalogImages(ctx context.Context, catalogName string) (map[string]interface{}, error) {
	archiveDir := cs.ArchiveDir

	if archiveDir == "" {
		archiveDir = "archive"
	}

	indexPath := filepath.Join(archiveDir, catalogName, "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return make(map[string]interface{}, 0), nil
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var indexData map[string]interface{}
	err = json.Unmarshal(data, &indexData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	return indexData, nil
}

// SearchCatalogs returns filtered catalogs based on search query
func (cs *CatalogService) SearchCatalogs(ctx context.Context, query string) ([]map[string]interface{}, error) {
	allCatalogs, err := cs.GetCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting catalogs for search: %w", err)
	}

	var filtered []map[string]interface{}

	// When no query is provided, return all catalogs
	if query == "" {
		filtered = allCatalogs
	} else {
		// Filter catalogs based on search query
		for _, catalog := range allCatalogs {
			name, _ := catalog["name"].(string)
			if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
				filtered = append(filtered, catalog)
			}
		}
	}

	return filtered, nil
}

// SearchCatalogImages returns filtered images in a catalog based on search query
func (cs *CatalogService) SearchCatalogImages(ctx context.Context, catalogName string, query string) (map[string]interface{}, error) {
	archiveDir := cs.ArchiveDir

	if archiveDir == "" {
		archiveDir = "archive"
	}

	indexPath := filepath.Join(archiveDir, catalogName, "index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("index file not found for catalog %s", catalogName)
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var indexData map[string]interface{}
	err = json.Unmarshal(data, &indexData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	// If no query provided, return all images
	if query == "" {
		return indexData, nil
	}

	// Filter images based on search query
	filteredData := make(map[string]interface{})

	for filename, value := range indexData {
		if dataMap, ok := value.(map[string]interface{}); ok {
			// Check if the query matches either the short name or description
			shortName := ""
			description := ""

			if sn, ok := dataMap["short_name"].(string); ok {
				shortName = sn
			}

			if desc, ok := dataMap["description"].(string); ok {
				description = desc
			}

			// If query matches either short name or description, include the image
			if strings.Contains(strings.ToLower(shortName), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(description), strings.ToLower(query)) {
				filteredData[filename] = dataMap
			}
		}
	}

	return filteredData, nil
}

// getCatalogInfo gets image count and last update date for a catalog directory
func (cs *CatalogService) getCatalogInfo(catalogPath string) (int, string, error) {
	// Count images in the catalog
	imageCount := 0
	lastUpdate := ""

	// Read index.json to get image information and update dates
	indexJsonPath := filepath.Join(catalogPath, "index.json")
	if _, err := os.Stat(indexJsonPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(indexJsonPath)
		if err != nil {
			return 0, "", err
		}

		var indexData map[string]interface{}
		err = json.Unmarshal(data, &indexData)
		if err != nil {
			return 0, "", err
		}

		// Count images and find the most recent update date
		for _, value := range indexData {
			if dataMap, ok := value.(map[string]interface{}); ok {
				imageCount++

				// Check for update_date field
				if updateDate, ok := dataMap["update_date"].(string); ok {
					// If this is the first date or it's more recent than current lastUpdate
					if lastUpdate == "" || updateDate > lastUpdate {
						lastUpdate = updateDate
					}
				}
			}
		}
	} else {
		// If index.json doesn't exist, scan directory for images
		entries, err := os.ReadDir(catalogPath)
		if err != nil {
			return 0, "", err
		}

		for _, entry := range entries {
			if !entry.IsDir() {

				// Skip files that match exclusion patterns
				if len(cs.Config.ExcludeFilter) > 0 {
					filePath := filepath.Join(catalogPath, entry.Name())
					if cs.Processor.ShouldExclude(filePath) {
						continue
					}
				}

				ext := strings.ToLower(filepath.Ext(entry.Name()))
				// Check if it's a supported image extension
				for _, supportedExt := range cs.Config.SupportedExtensions {
					if ext == strings.ToLower(supportedExt) {
						imageCount++
						break
					}
				}
			}
		}
	}

	return imageCount, lastUpdate, nil
}
