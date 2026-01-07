package services

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Read all subdirectories in archive
	err := filepath.Walk(archiveDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == archiveDir {
			return nil
		}

		// Only process directories (not files)
		if info.IsDir() {
			relPath, _ := filepath.Rel(archiveDir, path)

			// Get image count and last update date
			imageCount, lastUpdate, err := cs.getCatalogInfo(path)
			if err != nil {
				// Log error but continue processing other catalogs
				fmt.Printf("Error getting catalog info for %s: %v\n", relPath, err)
			}

			if imageCount == 0 {
				return nil // Skip empty catalogs or those with errors
			}

			catalogs = append(catalogs, map[string]interface{}{
				"name":       relPath,
				"imageCount": imageCount,
				"lastUpdate": lastUpdate,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error reading archive directory: %w", err)
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
