package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"kbase-catalog/internal/config"
)

type LLMResponse struct {
	ShortName   string `json:"short_name"`
	Description string `json:"description"`
}

type LLMClient struct {
	config *config.Config
	client *http.Client
}

func NewLLMClient(cfg *config.Config) *LLMClient {
	return &LLMClient{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (c *LLMClient) AskLLM(ctx context.Context, imagePath string, imageData string) (*LLMResponse, string, error) {
	payload := map[string]interface{}{
		"model": c.config.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": c.config.SystemPrompt,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Analyze this image and provide a short name and description.",
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageData,
						},
					},
				},
			},
		},
		"stream": false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request to LLM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("LLM API returned status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal LLM response: %w", err)
	}

	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, "", fmt.Errorf("unexpected response format from LLM API")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("unexpected message format in LLM response")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, "", fmt.Errorf("unexpected content format in LLM response")
	}

	var llmResponse LLMResponse
	err = json.Unmarshal([]byte(content), &llmResponse)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse LLM response as JSON: %w", err)
	}

	modelName := ""
	if model, ok := response["model"].(string); ok {
		modelName = model
	}

	return &llmResponse, modelName, nil
}
