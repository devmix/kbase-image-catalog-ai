package llm

import (
	"context"
	"encoding/json"
	"kbase-catalog/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLLMClient_AskLLM(t *testing.T) {
	// Create a mock server to simulate LLM API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body := make(map[string]interface{})
		json.NewDecoder(r.Body).Decode(&body)

		// Verify request structure
		assert.Equal(t, "test-model", body["model"])

		// Mock successful response
		response := map[string]interface{}{
			"model": "test-model",
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"short_name": "Test Image", "description": "This is a test image."}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create config with mock API URL
	client := &LLMClient{
		config: &config.Config{
			APIURL:  server.URL,
			Model:   "test-model",
			Timeout: 10,
			SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Test successful response
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, model, err := client.AskLLM(ctx, "/test/image.jpg", "data:image/jpeg;base64,test-data")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Test Image", response.ShortName)
	assert.Equal(t, "This is a test image.", response.Description)
	assert.Equal(t, "test-model", model)
}

func TestLLMClient_AskLLM_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &LLMClient{
		config: &config.Config{
			APIURL:  server.URL,
			Model:   "test-model",
			Timeout: 10,
			SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, model, err := client.AskLLM(ctx, "/test/image.jpg", "data:image/jpeg;base64,test-data")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "", model)
}

func TestLLMClient_AskLLM_InvalidResponse(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"invalid": json}`))
	}))
	defer server.Close()

	client := &LLMClient{
		config: &config.Config{
			APIURL:  server.URL,
			Model:   "test-model",
			Timeout: 10,
			SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, model, err := client.AskLLM(ctx, "/test/image.jpg", "data:image/jpeg;base64,test-data")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "", model)
}

func TestLLMClient_AskLLM_InvalidJSONResponse(t *testing.T) {
	// Create a mock server that returns valid JSON but invalid content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"invalid_json": `,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &LLMClient{
		config: &config.Config{
			APIURL:  server.URL,
			Model:   "test-model",
			Timeout: 10,
			SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, model, err := client.AskLLM(ctx, "/test/image.jpg", "data:image/jpeg;base64,test-data")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "", model)
}

func TestLLMClient_AskLLM_InvalidContentFormat(t *testing.T) {
	// Create a mock server that returns valid JSON but with invalid content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"model": "test-model",
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": `{"short_name": "", "description": ""}`,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &LLMClient{
		config: &config.Config{
			APIURL:  server.URL,
			Model:   "test-model",
			Timeout: 10,
			SystemPrompt: `You are a helpful assistant specialized in image analysis.
You must respond in valid JSON format ONLY, without any extra text.
The JSON must contain two keys:
1. "short_name": a short, descriptive name for the image.
2. "description": a detailed description of the image in English.`,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, model, err := client.AskLLM(ctx, "/test/image.jpg", "data:image/jpeg;base64,test-data")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	// Should have empty content but not fail
	assert.Equal(t, "", response.ShortName)
	assert.Equal(t, "", response.Description)
	assert.Equal(t, "test-model", model)
}
