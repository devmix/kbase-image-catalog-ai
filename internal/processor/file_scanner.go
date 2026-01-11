package processor

import (
	"encoding/json"
	"fmt"
	"kbase-catalog/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"kbase-catalog/internal/config"

	"github.com/moby/patternmatcher"
)

type FileScanner struct {
	config  *config.Config
	exclude *patternmatcher.PatternMatcher
}

func NewFileScanner(cfg *config.Config) *FileScanner {
	var matcher *patternmatcher.PatternMatcher = nil
	if len(cfg.ExcludeFilter) != 0 {
		m, err := patternmatcher.New(cfg.ExcludeFilter)
		if err != nil {
			panic(err)
		}
		matcher = m
	}

	return &FileScanner{
		config:  cfg,
		exclude: matcher,
	}
}

func (fs *FileScanner) HasImages(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			for _, supportedExt := range fs.config.SupportedExtensions {
				if ext == strings.ToLower(supportedExt) {
					return true
				}
			}
		}
	}

	return false
}

func (fs *FileScanner) FindImagesToProcess(dirPath string) ([]string, error) {
	var images []string

	for _, ext := range fs.config.SupportedExtensions {
		pattern := filepath.Join(dirPath, "*"+ext)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to find files with extension %s: %w", ext, err)
		}
		images = append(images, matches...)

		patternUpper := filepath.Join(dirPath, "*"+strings.ToUpper(ext[1:]))
		matchesUpper, err := filepath.Glob(patternUpper)
		if err != nil {
			return nil, fmt.Errorf("failed to find files with uppercase extension %s: %w", ext, err)
		}
		images = append(images, matchesUpper...)
	}

	var filteredImages []string
	for _, img := range images {
		baseName := filepath.Base(img)
		if baseName != "index.json" && baseName != "index.md" {
			filteredImages = append(filteredImages, img)
		}
	}

	// Apply exclusion patterns
	if len(fs.config.ExcludeFilter) > 0 {
		filteredImages = fs.FilterExcludedFiles(filteredImages)
	}

	return filteredImages, nil
}

func (fs *FileScanner) LoadExistingData(indexJsonPath string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	if utils.IsFileExists(indexJsonPath) {
		content, err := os.ReadFile(indexJsonPath)
		if err != nil {
			return data, fmt.Errorf("failed to read index.json: %w", err)
		}

		err = json.Unmarshal(content, &data)
		if err != nil {
			fmt.Printf("Error reading %s, creating new data.\n", indexJsonPath)
			return make(map[string]interface{}), nil
		}
	}

	return data, nil
}

func (fs *FileScanner) ShouldExclude(file string) bool {
	if fs.exclude == nil {
		return false // When no exclude, don't exclude anything
	}
	matched, _ := fs.exclude.MatchesOrParentMatches(file)
	return matched
}

func (fs *FileScanner) FilterExcludedFiles(files []string) []string {
	var result []string
	for _, file := range files {
		if !fs.ShouldExclude(file) {
			result = append(result, file)
		}
	}
	return result
}
