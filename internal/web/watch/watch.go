package watch

import (
	"context"
	"kbase-catalog/internal/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kbase-catalog/internal/web/queue"

	"github.com/fsnotify/fsnotify"
)

// CatalogWatcher monitors file system changes in the archive directory
type CatalogWatcher struct {
	watcher    *fsnotify.Watcher
	queue      *queue.TaskQueue
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
	archiveDir string
}

// NewCatalogWatcher creates a new catalog watcher
func NewCatalogWatcher(queue *queue.TaskQueue, archivePath string) (*CatalogWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Set default archive directory to "archive"
	return &CatalogWatcher{
		watcher:    watcher,
		queue:      queue,
		ctx:        ctx,
		cancel:     cancel,
		isRunning:  false,
		archiveDir: archivePath,
	}, nil
}

// Start starts the catalog watcher
func (cw *CatalogWatcher) Start() error {
	cw.isRunning = true

	// Add the archive directory and all subdirectories to watch
	err := cw.addDirectoriesToWatch(cw.archiveDir)
	if err != nil {
		log.Printf("Failed to add directories for watching: %v", err)
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-cw.watcher.Events:
				if !ok {
					return
				}

				// Only process write and create events to image files
				if event.Op&fsnotify.Chmod != fsnotify.Chmod {
					cw.handleFileChange(event.Name)
				}

			case err, ok := <-cw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)

			case <-cw.ctx.Done():
				cw.watcher.Close()
				return
			}
		}
	}()

	return nil
}

// Stop stops the catalog watcher
func (cw *CatalogWatcher) Stop() error {
	cw.cancel()
	cw.isRunning = false
	return cw.watcher.Close()
}

// addDirectoriesToWatch recursively adds all directories to watch for changes
func (cw *CatalogWatcher) addDirectoriesToWatch(rootDir string) error {
	// First, add the root directory itself
	err := cw.watcher.Add(rootDir)
	if err != nil {
		log.Printf("Failed to add root directory %s to watcher: %v", rootDir, err)
		return err
	}

	// Then recursively walk all subdirectories
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only add directories to watch
		if info.IsDir() && path != rootDir {
			err := cw.watcher.Add(path)
			if err != nil {
				log.Printf("Failed to add directory %s to watcher: %v", path, err)
				// Don't return error here - continue with other directories
			}
		}

		return nil
	})

	return err
}

// handleFileChange processes file system changes
func (cw *CatalogWatcher) handleFileChange(filePath string) {
	isDir := utils.IsDirectory(filePath)
	filePath, err := filepath.Rel(cw.archiveDir, filePath)
	if err != nil {
		log.Printf("Error getting relative path: %s", filePath)
		return
	}

	catalogName := filepath.Base(filePath)

	if !isDir {
		// Check if the file is an image file
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != "" {
			// Only process supported image extensions
			supportedExtensions := []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"}

			// Check if this is a file with a supported extension
			isImageFile := false
			for _, supportedExt := range supportedExtensions {
				if ext == supportedExt {
					isImageFile = true
					break
				}
			}

			if !isImageFile {
				return
			}

			// Extract catalog name from the file path
			// The path will be like "archive/collection1/image.jpg"
			parts := strings.Split(filePath, "/")

			// Make sure we have enough parts to extract the catalog name
			if len(parts) < 2 {
				log.Printf("Invalid file path structure: %s", filePath)
				return
			}

			catalogName = parts[0] // Get the second part which is the catalog name
		}
	}

	// Add reindex task to queue
	go func() {
		// Small delay to ensure file write is complete
		time.Sleep(200 * time.Millisecond)
		if err := cw.queue.AddTask(catalogName, "watcher"); err != nil {
			log.Printf("Failed to add reindex task for catalog %s: %v", catalogName, err)
		}
	}()
}
