package queue

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
)

// ReindexTask represents a task to reindex a catalog
type ReindexTask struct {
	CatalogName string
	Source      string // "manual" or "watcher"
	CreatedAt   time.Time
}

// TaskQueue manages reindex tasks with concurrency control
type TaskQueue struct {
	tasks      chan *ReindexTask
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	processor  *processor.CatalogProcessor
	config     *config.Config
	isRunning  bool
	mutex      sync.RWMutex
	archiveDir string
}

// NewTaskQueue creates a new task queue for reindexing
func NewTaskQueue(cfg *config.Config, processor *processor.CatalogProcessor, archivePath string) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &TaskQueue{
		tasks:      make(chan *ReindexTask, 100), // Buffered channel with capacity of 100
		ctx:        ctx,
		cancel:     cancel,
		processor:  processor,
		config:     cfg,
		isRunning:  false,
		archiveDir: archivePath,
	}
}

// Start starts the task queue processing
func (q *TaskQueue) Start() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.isRunning {
		return nil // Already running
	}

	q.isRunning = true
	q.wg.Add(1)

	go func() {
		defer q.wg.Done()
		for {
			select {
			case task, ok := <-q.tasks:
				if !ok {
					return // Channel closed
				}

				// Process the reindex task
				q.processTask(task)

			case <-q.ctx.Done():
				return // Context cancelled
			}
		}
	}()

	return nil
}

// Stop stops the task queue processing
func (q *TaskQueue) Stop() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.isRunning {
		return nil // Already stopped
	}

	q.cancel()
	close(q.tasks)
	q.wg.Wait()
	q.isRunning = false

	return nil
}

// AddTask adds a reindex task to the queue
func (q *TaskQueue) AddTask(catalogName, source string) error {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if !q.isRunning {
		log.Printf("Task queue not running - cannot add task for catalog %s", catalogName)
		return nil // Queue not running
	}

	task := &ReindexTask{
		CatalogName: catalogName,
		Source:      source,
		CreatedAt:   time.Now(),
	}

	select {
	case q.tasks <- task:
		log.Printf("Added reindex task for catalog %s (source: %s)", catalogName, source)
		return nil
	default:
		// Channel is full, log warning but still add task
		log.Printf("Task queue is full - dropping task for catalog %s", catalogName)
		// For now we'll silently drop if full, but in a more robust implementation,
		// we might want to retry or queue with backoff
		return nil
	}
}

// processTask processes a single reindex task
func (q *TaskQueue) processTask(task *ReindexTask) {
	// TODO add rate limiting here and error handling for failed tasks

	// For now, just process the catalog directly
	catalogPath := filepath.Join(q.archiveDir, task.CatalogName)

	log.Printf("Processing reindex task for catalog %s (source: %s)", task.CatalogName, task.Source)

	err := q.processor.ProcessCatalog(q.ctx, catalogPath)
	if err != nil {
		// TODO retry or mark as failed
		// Log error but don't stop processing other tasks
		log.Printf("Failed to reindex catalog %s: %v", task.CatalogName, err)
	} else {
		log.Printf("Successfully reindexed catalog %s", task.CatalogName)
	}
}
