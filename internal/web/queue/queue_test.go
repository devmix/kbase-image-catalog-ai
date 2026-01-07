package queue

import (
	"testing"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskQueue(t *testing.T) {
	// Create a mock config
	mockConfig := &config.Config{}
	archivePath := "/tmp/test-archive"

	// Create a real processor for testing (we'll test it separately)
	realProcessor := &processor.CatalogProcessor{}

	queue := NewTaskQueue(mockConfig, realProcessor, archivePath)

	assert.NotNil(t, queue)
	assert.Equal(t, mockConfig, queue.config)
	assert.Equal(t, realProcessor, queue.processor)
	assert.Equal(t, archivePath, queue.archiveDir)
	assert.False(t, queue.isRunning)
	assert.NotNil(t, queue.tasks)
	assert.NotNil(t, queue.ctx)
	assert.NotNil(t, queue.cancel)
}

func TestTaskQueue_Start(t *testing.T) {
	// Create a mock config
	mockConfig := &config.Config{}
	archivePath := "/tmp/test-archive"

	// Create a real processor for testing
	realProcessor := &processor.CatalogProcessor{}

	queue := NewTaskQueue(mockConfig, realProcessor, archivePath)

	// Start the queue
	err := queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Try to start again - should not error but do nothing
	err = queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Stop the queue for clean up
	queue.Stop()
}

func TestTaskQueue_Stop(t *testing.T) {
	// Create a mock config
	mockConfig := &config.Config{}
	archivePath := "/tmp/test-archive"

	// Create a real processor for testing
	realProcessor := &processor.CatalogProcessor{}

	queue := NewTaskQueue(mockConfig, realProcessor, archivePath)

	// Start the queue first
	err := queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Stop the queue
	err = queue.Stop()
	assert.NoError(t, err)
	assert.False(t, queue.isRunning)

	// Try to stop again - should not error but do nothing
	err = queue.Stop()
	assert.NoError(t, err)
	assert.False(t, queue.isRunning)
}

func TestTaskQueue_AddTask(t *testing.T) {
	// Create a mock config
	mockConfig := &config.Config{}
	archivePath := "/tmp/test-archive"

	// Create a real processor for testing
	realProcessor := &processor.CatalogProcessor{}

	queue := NewTaskQueue(mockConfig, realProcessor, archivePath)

	// Try to add task when queue is not running - should return nil (no error)
	err := queue.AddTask("test-catalog", "manual")
	assert.NoError(t, err)

	// Start the queue
	err = queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Add a task when queue is running
	err = queue.AddTask("test-catalog", "manual")
	assert.NoError(t, err)

	// Stop the queue for clean up
	queue.Stop()
}

func TestTaskQueue_AddTask_WithFullChannel(t *testing.T) {
	// Create a mock config
	mockConfig := &config.Config{}
	archivePath := "/tmp/test-archive"

	// Create a real processor for testing
	realProcessor := &processor.CatalogProcessor{}

	queue := NewTaskQueue(mockConfig, realProcessor, archivePath)

	// Start the queue
	err := queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Fill up the channel (capacity is 100) by adding more tasks than capacity
	for i := 0; i < 105; i++ { // Add 105 tasks to exceed buffer
		err = queue.AddTask("test-catalog", "manual")
		assert.NoError(t, err)
	}

	// Stop the queue for clean up
	queue.Stop()
}
