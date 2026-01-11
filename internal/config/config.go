package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	APIURL              string   `yaml:"api_url"`
	Model               string   `yaml:"model"`
	Timeout             int      `yaml:"timeout"`
	SystemPrompt        string   `yaml:"system_prompt"`
	SupportedExtensions []string `yaml:"supported_extensions"`
	ExcludeFilter       []string `yaml:"exclude_filter"`
	ParallelRequests    int      `yaml:"parallel_requests"`
	MaxRetries          int      `yaml:"max_retries"`
	RetryDelay          int      `yaml:"retry_delay"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read YAML file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func GetDefaultConfig() *Config {
	return &Config{
		APIURL:  "http://localhost:1234/v1/chat/completions",
		Model:   "llava-v1.5-7b",
		Timeout: 60,
		SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.

Example output format:
{"short_name": "Sunset on the beach", "description": "The image shows a sunset at sea..."}`,
		SupportedExtensions: []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"},
		ExcludeFilter:       []string{},
		ParallelRequests:    3,
		MaxRetries:          3,
		RetryDelay:          5,
	}
}

func validateConfig(config *Config) error {
	if config.APIURL == "" {
		return fmt.Errorf("api_url is required")
	}
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if config.ParallelRequests <= 0 {
		return fmt.Errorf("parallel_requests must be positive")
	}
	if config.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative")
	}
	return nil
}

func (c *Config) WriteToFile(configPath string) error {
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}
