package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Test valid config
	validConfigContent := `
api_url: "http://localhost:1234/v1/chat/completions"
model: "test-model"
timeout: 60
system_prompt: |
  Test system prompt
supported_extensions:
  - ".png"
  - ".jpg"
exclude_filter:
  - "*/temp/*"
  - "*/tmp/*"
  - "*.tmp"
  - "*.bak"
  - ".git"
parallel_requests: 3
max_retries: 3
retry_delay: 5
`

	err := os.WriteFile(configPath, []byte(validConfigContent), 0644)
	assert.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "http://localhost:1234/v1/chat/completions", config.APIURL)
	assert.Equal(t, "test-model", config.Model)
	assert.Equal(t, 60, config.Timeout)
	assert.Equal(t, 3, config.ParallelRequests)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.RetryDelay)
	assert.Equal(t, []string{"*/temp/*", "*/tmp/*", "*.tmp", "*.bak", ".git"}, config.ExcludeFilter)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/non/existent/path/config.yaml")
	assert.Error(t, err)
}

func TestValidateConfig(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := &Config{
			APIURL:              "http://localhost:1234/v1/chat/completions",
			Model:               "test-model",
			Timeout:             60,
			ParallelRequests:    3,
			MaxRetries:          3,
			RetryDelay:          5,
			SupportedExtensions: []string{".png", ".jpg"},
			SystemPrompt:        "Test prompt",
		}

		err := validateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Missing API URL", func(t *testing.T) {
		config := &Config{
			Model:               "test-model",
			Timeout:             60,
			ParallelRequests:    3,
			MaxRetries:          3,
			RetryDelay:          5,
			SupportedExtensions: []string{".png", ".jpg"},
			SystemPrompt:        "Test prompt",
		}

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api_url is required")
	})

	t.Run("Invalid timeout", func(t *testing.T) {
		config := &Config{
			APIURL:              "http://localhost:1234/v1/chat/completions",
			Model:               "test-model",
			Timeout:             0,
			ParallelRequests:    3,
			MaxRetries:          3,
			RetryDelay:          5,
			SupportedExtensions: []string{".png", ".jpg"},
			SystemPrompt:        "Test prompt",
		}

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout must be positive")
	})

	t.Run("Invalid parallel requests", func(t *testing.T) {
		config := &Config{
			APIURL:              "http://localhost:1234/v1/chat/completions",
			Model:               "test-model",
			Timeout:             60,
			ParallelRequests:    0,
			MaxRetries:          3,
			RetryDelay:          5,
			SupportedExtensions: []string{".png", ".jpg"},
			SystemPrompt:        "Test prompt",
		}

		err := validateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parallel_requests must be positive")
	})
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "http://localhost:1234/v1/chat/completions", config.APIURL)
	assert.Equal(t, "llava-v1.5-7b", config.Model)
	assert.Equal(t, 60, config.Timeout)
	assert.Equal(t, 3, config.ParallelRequests)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.RetryDelay)
}
