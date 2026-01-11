package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kbase-catalog/internal/config"
)

type IndexGenerator struct {
	config *config.Config
}

func NewIndexGenerator(cfg *config.Config) *IndexGenerator {
	return &IndexGenerator{
		config: cfg,
	}
}

func (ig *IndexGenerator) SaveIndexJson(indexJsonPath string, data map[string]interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(indexJsonPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write index.json: %w", err)
	}

	return nil
}

func (ig *IndexGenerator) GenerateCatalogIndexAsMarkdown(mdPath string, data map[string]interface{}) error {
	lines := []string{}
	lines = append(lines, "# Image Catalog")
	lines = append(lines, "| Image | Description |")
	lines = append(lines, "|---|---|")

	var sortedKeys []string
	for key := range data {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		info := data[key]
		if infoMap, ok := info.(map[string]interface{}); ok {
			shortName := key
			description := ""

			if sn, ok := infoMap["short_name"].(string); ok {
				shortName = sn
			}
			if desc, ok := infoMap["description"].(string); ok {
				description = desc
			}

			lines = append(lines, fmt.Sprintf("| [%s](%s) | %s |", shortName, key, description))
		}
	}

	content := strings.Join(lines, "\n")
	err := os.WriteFile(mdPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write index.md: %w", err)
	}

	return nil
}

func (ig *IndexGenerator) GenerateRootIndexAsMarkdown(rootPath string, subdirs []string) {
	rootMdPath := filepath.Join(rootPath, "index.md")

	lines := []string{}
	lines = append(lines, "# Directory List")

	if err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirName := filepath.Base(path)
			mdPath := filepath.Join(dirName, "index.md")
			lines = append(lines, fmt.Sprintf("- [%s](%s)", dirName, mdPath))
		}

		return nil
	}); err != nil {
		fmt.Printf("Error listing catalog root: %v\n", err)
	}

	content := strings.Join(lines, "\n")
	if err := os.WriteFile(rootMdPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing root index.md: %v\n", err)
	}
}

// GenerateGlobalIndex creates a global index of all catalogs with their metadata
func (ig *IndexGenerator) GenerateGlobalIndex(rootPath string, catalogData map[string]interface{}) error {
	globalIndexPath := filepath.Join(rootPath, "index.json")

	content, err := json.MarshalIndent(catalogData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal global index JSON: %w", err)
	}

	err = os.WriteFile(globalIndexPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write global index.json: %w", err)
	}

	return nil
}
