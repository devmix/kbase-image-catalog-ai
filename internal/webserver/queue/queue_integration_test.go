package queue

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"kbase-catalog/internal/config"
	"kbase-catalog/internal/processor"
)

// Integration test to verify that the task queue processes tasks correctly
func TestTaskQueue_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-queue-integration")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock config
	mockConfig := &config.Config{}

	// Create a real processor for testing
	realProcessor := processor.NewCatalogProcessor(mockConfig, tempDir)

	queue := NewTaskQueue(mockConfig, realProcessor, tempDir)

	// Start the queue
	err = queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Add a task to the queue
	err = queue.AddTask("test-catalog", "manual")
	assert.NoError(t, err)

	// Give it a moment to process
	// Note: This is a basic test as we cannot easily verify that the actual processing happened
	// without creating a full integration test with a real catalog structure

	// Stop the queue
	err = queue.Stop()
	assert.NoError(t, err)
	assert.False(t, queue.isRunning)
}

// Test that the task queue can handle multiple concurrent tasks
func TestTaskQueue_MultipleTasks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-queue-multiple")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock config
	mockConfig := &config.Config{}

	// Create a real processor for testing
	realProcessor := processor.NewCatalogProcessor(mockConfig, tempDir)

	queue := NewTaskQueue(mockConfig, realProcessor, tempDir)

	// Start the queue
	err = queue.Start()
	assert.NoError(t, err)
	assert.True(t, queue.isRunning)

	// Add multiple tasks to the queue
	for i := 0; i < 10; i++ {
		err = queue.AddTask("test-catalog-"+string(rune(i+'0')), "manual")
		assert.NoError(t, err)
	}

	// Stop the queue
	err = queue.Stop()
	assert.NoError(t, err)
	assert.False(t, queue.isRunning)
}
