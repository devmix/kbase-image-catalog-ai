package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/llm"
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

// ProcessCatalog processes all directories in the catalog root
func (cp *CatalogProcessor) ProcessCatalog(ctx context.Context, catalogDir string) error {
	fmt.Printf("Starting scan in: %s\n", catalogDir)

	processedSubdirs := []string{}

	// Keep track of catalogs and their information for global index
	catalogs := make(map[string]interface{})
	hasFilter := cp.fs.HasFilter()

	err := filepath.Walk(catalogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if path == cp.archiveDir {
			return nil
		}

		relPath, err := filepath.Rel(catalogDir, path)
		if err != nil {
			return fmt.Errorf("Error transforming path to relative: %s: %v\n", path, err)
		}

		// Skip directories that match exclusion patterns
		if info.IsDir() && hasFilter {
			if err == nil && relPath != "." && cp.fs.ShouldExclude(relPath) {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			fmt.Printf("\n--> Processing directory: %s\n", strings.TrimPrefix(path, catalogDir+"/"))

			// Process the directory and get whether it was modified
			data, err := cp.dp.ProcessDirectory(ctx, path)
			if err != nil {
				fmt.Printf("Error processing directory %s: %v\n", path, err)
			} else if data != nil {
				processedSubdirs = append(processedSubdirs, path)
			}

			if data != nil {
				catalogs[relPath] = data
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	fmt.Printf("\nUpdating root index...\n")
	cp.ig.GenerateRootIndexAsMarkdown(cp.archiveDir, processedSubdirs)

	fmt.Printf("Generating global catalog index...\n")
	err = cp.ig.GenerateGlobalIndex(cp.archiveDir, catalogs)
	if err != nil {
		return fmt.Errorf("error generating global index: %w", err)
	}
	fmt.Printf("Done.\n")

	return nil
}

func (cp *CatalogProcessor) TestSingleImage(ctx context.Context, imagePath string) (*llm.LLMResponse, error) {
	return cp.ip.TestSingleImage(ctx, imagePath)
}

func (cp *CatalogProcessor) ShouldExclude(path string) bool {
	return cp.fs.ShouldExclude(path)
}
