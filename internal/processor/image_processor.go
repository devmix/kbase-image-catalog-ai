package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kbase-catalog/internal/config"
	"kbase-catalog/internal/encoder"
	"kbase-catalog/internal/llm"
)

type ImageProcessor struct {
	config *config.Config
}

func NewImageProcessor(cfg *config.Config) *ImageProcessor {
	return &ImageProcessor{
		config: cfg,
	}
}

func (ip *ImageProcessor) ProcessSingleImage(ctx context.Context, imgPath string, currentData map[string]interface{}) (bool, error) {
	imgKey := filepath.Base(imgPath)
	record, exists := currentData[imgKey]

	if !ip.needsProcessing(currentData, imgPath) {
		return false, nil
	}

	var logMsg string
	if exists {
		if recordMap, ok := record.(map[string]interface{}); ok {
			if shortName, ok := recordMap["short_name"].(string); ok && shortName == "error_processing" {
				logMsg = fmt.Sprintf("RETRY: %s (previous attempt failed)", imgPath)
			} else {
				logMsg = fmt.Sprintf("Processing: %s", imgPath)
			}
		} else {
			logMsg = fmt.Sprintf("Processing: %s", imgPath)
		}
	} else {
		logMsg = fmt.Sprintf("Processing: %s", imgPath)
	}

	fmt.Printf("%s\n", logMsg)

	imageData, err := encoder.EncodeImageToBase64(imgPath)
	if err != nil {
		ip.handleProcessingError(imgPath, currentData)
		return true, fmt.Errorf("failed to encode image: %w", err)
	}

	client := llm.NewLLMClient(ip.config)
	llmResponse, model, err := client.AskLLM(ctx, imgPath, imageData)
	if err != nil {
		ip.handleProcessingError(imgPath, currentData)
		return true, fmt.Errorf("failed to process image with LLM: %w", err)
	}

	if llmResponse != nil && ValidateResponse(llmResponse) {
		currentData[imgKey] = map[string]interface{}{
			"short_name":    llmResponse.ShortName,
			"description":   llmResponse.Description,
			"original_name": filepath.Base(imgPath),
			"vl_model":      model,
			"update_date":   time.Now().Format(time.RFC3339),
		}
		fmt.Printf("  -> Successfully processed: %s\n", llmResponse.ShortName)
		return true, nil
	}

	ip.handleProcessingError(imgPath, currentData)
	return true, nil
}

func (ip *ImageProcessor) needsProcessing(currentData map[string]interface{}, imgPath string) bool {
	imgKey := filepath.Base(imgPath)
	record, exists := currentData[imgKey]

	if !exists {
		return true
	}

	if recordMap, ok := record.(map[string]interface{}); ok {
		if shortName, ok := recordMap["short_name"].(string); ok && shortName == "error_processing" {
			return true
		}
	}

	return false
}

// NeedsProcessing is a public wrapper for the internal needsProcessing function
func NeedsProcessing(currentData map[string]interface{}, imgPath string) bool {
	imgKey := filepath.Base(imgPath)
	record, exists := currentData[imgKey]

	if !exists {
		return true
	}

	// If record is not a map, treat it as if file needs processing
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		return true
	}

	if shortName, ok := recordMap["short_name"].(string); ok && shortName == "error_processing" {
		return true
	}

	return false
}

// ValidateResponse is a public wrapper for the internal validateResponse function
func ValidateResponse(response *llm.LLMResponse) bool {
	if response == nil {
		return false
	}
	return response.ShortName != "" && response.Description != ""
}

func (ip *ImageProcessor) handleProcessingError(imgPath string, currentData map[string]interface{}) {
	imgKey := filepath.Base(imgPath)
	currentData[imgKey] = map[string]interface{}{
		"short_name":    "error_processing",
		"description":   "Error processing file (retry will be attempted)",
		"original_name": filepath.Base(imgPath),
		"vl_model":      "unknown",
		"update_date":   time.Now().Format(time.RFC3339),
	}
	fmt.Printf("  -> Recognition error. Will be retried.\n")
}

// HandleProcessingError is a public wrapper for the internal handleProcessingError function
func HandleProcessingError(imgPath string, currentData map[string]interface{}) {
	imgKey := filepath.Base(imgPath)
	currentData[imgKey] = map[string]interface{}{
		"short_name":    "error_processing",
		"description":   "Error processing file (retry will be attempted)",
		"original_name": filepath.Base(imgPath),
		"vl_model":      "unknown",
		"update_date":   time.Now().Format(time.RFC3339),
	}
	fmt.Printf("  -> Recognition error. Will be retried.\n")
}

func (ip *ImageProcessor) TestSingleImage(ctx context.Context, imagePath string) (*llm.LLMResponse, error) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", imagePath)
	}

	fmt.Printf("Testing image: %s\n", imagePath)
	fmt.Printf("Directory: %s\n", filepath.Base(filepath.Dir(imagePath)))
	fmt.Printf("Filename: %s\n", filepath.Base(imagePath))

	imageData, err := encoder.EncodeImageToBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	client := llm.NewLLMClient(ip.config)
	llmResponse, model, err := client.AskLLM(ctx, imagePath, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to process image with LLM: %w", err)
	}

	if llmResponse != nil && llmResponse.ShortName != "" && llmResponse.Description != "" {
		fmt.Printf("\n✅ Successfully obtained result:\n")
		fmt.Printf("Short name: %s\n", llmResponse.ShortName)
		fmt.Printf("Description: %s\n", llmResponse.Description)
		fmt.Printf("Vision model: %s\n", model)
		return llmResponse, nil
	}

	return nil, fmt.Errorf("invalid LLM response")
}
