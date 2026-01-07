# ADR-0004: Integrate LLM Models for Image Analysis

- Status: accepted
- Date: 2026-01-08
- Deciders: Sergey Grachev

## Context

The core functionality of KBase Image Catalog relies on analyzing images and generating descriptive metadata. This requires integration with AI/ML models capable of understanding image content and generating natural language descriptions.

Key considerations:
- **Model Performance**: Accuracy and quality of image analysis
- **API Compatibility**: Integration ease with existing infrastructure
- **Cost Efficiency**: Computational and financial costs
- **Latency**: Response time for image analysis
- **Scalability**: Ability to handle large volumes of images
- **Reliability**: Uptime and error handling requirements
- **Customization**: Ability to fine-tune for specific use cases
- **Privacy**: Data handling and security considerations
- **Maintenance**: Long-term model updates and support
- **Ecosystem**: Available tools and community support

## Decision

We will integrate **LLM (Large Language Model) APIs** for image analysis, specifically focusing on vision-language models like LLaVA and Qwen-VL that can process both images and text to generate descriptive metadata.

## Consequences

### Positive Consequences

1. **High-Quality Analysis**: LLMs provide sophisticated understanding of image content
2. **Natural Language Descriptions**: Generate human-readable metadata
3. **Versatility**: Handle diverse image types (photos, art, screenshots, diagrams)
4. **Contextual Understanding**: Recognize relationships between objects and scenes
5. **Scalability**: Cloud-based models can handle variable workloads
6. **Continuous Improvement**: Models are regularly updated and improved
7. **Cost-Effective**: Pay-as-you-go pricing models for variable usage
8. **Easy Integration**: RESTful APIs simplify integration
9. **Multiple Model Support**: Ability to switch between different models
10. **Future-Proof**: Leverage ongoing advancements in AI capabilities

### Negative Consequences

1. **Network Dependency**: Requires internet connection for API calls
2. **Latency Issues**: Response times can be slower than local processing
3. **Cost Concerns**: API costs can add up with large volumes
4. **Privacy Risks**: Images may be processed by third-party services
5. **Rate Limiting**: API rate limits can affect processing throughput
6. **Model Availability**: Dependent on third-party service availability
7. **Configuration Complexity**: Requires careful prompt engineering
8. **Error Handling**: Complex error scenarios from API failures
9. **Data Transfer**: Large image files need to be uploaded
10. **Vendor Lock-in**: Potential dependency on specific AI providers

## Alternatives Considered

### 1. Custom Computer Vision Models
**Rejected** because:
- High development and training costs
- Requires significant ML expertise
- Limited to specific use cases without generalization
- Ongoing maintenance and model updates needed
- Higher infrastructure requirements

### 2. Local Open-Source Models
**Rejected** because:
- Higher computational requirements for local inference
- Less accurate than state-of-the-art cloud models
- More complex deployment and maintenance
- Limited to available hardware resources
- Slower response times

### 3. Traditional Computer Vision (OpenCV, etc.)
**Rejected** because:
- Limited to object detection and classification
- Cannot generate natural language descriptions
- Requires extensive custom development
- Poor understanding of complex scenes and relationships
- Limited to pre-trained capabilities

### 4. Third-Party Image Analysis Services
**Rejected** because:
- Less specialized for our use case
- Limited customization options
- Potentially higher costs
- Less control over processing quality
- Integration complexity with multiple services

### 5. Hybrid Approach (Local + Cloud)
**Rejected** because:
- Increased architectural complexity
- Higher development and maintenance overhead
- Difficult to optimize for both approaches
- Potential duplication of effort
- More points of failure

## Related ADRs

- [ADR-0002: Use Go for Backend Development](./ADR-0002-use-go-for-backend.md) - Backend technology that implements LLM integration
- [ADR-0003: Use HTMX for Web Interface](./ADR-0003-htmx-for-web-interface.md) - UI that displays LLM analysis results
- [ADR-0006: Implement Comprehensive Error Handling Strategy](./ADR-0006-error-handling.md) - Error handling for API failures
- [ADR-0007: Focus on Performance Optimization](./ADR-0007-performance-optimization.md) - Performance considerations for API calls

## Implementation Details

### Supported Models

1. **LLaVA v1.5-7b**: Open-source vision-language model
2. **Qwen-VL-8B-Instruct**: Alibaba's vision-language model
3. **OpenAI GPT-4V**: Commercial alternative with high accuracy

### API Integration Architecture

```go
// LLM Client Interface
type LLMClient interface {
    AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (*LLMResponse, error)
    GetModelInfo() ModelInfo
    IsAvailable() bool
}

// HTTP Client with retry logic
type HTTPClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
    maxRetries int
    retryDelay time.Duration
}
```

### Image Processing Pipeline

1. **Image Validation**: Check file format, size, and content
2. **Image Preprocessing**: Resize, format conversion if needed
3. **Base64 Encoding**: Convert image to base64 for API transport
4. **Prompt Construction**: Build system prompt with user requirements
5. **API Call**: Send request to LLM service
6. **Response Parsing**: Extract structured metadata from JSON response
7. **Error Handling**: Retry on failures with exponential backoff
8. **Result Storage**: Save metadata to index files

### Configuration Options

```yaml
api_url: "http://192.168.1.7:1234/v1/chat/completions"
model: "llava-v1.5-7b"
timeout: 60
system_prompt: |-
  You are a helpful assistant specialized in image analysis.
  You must respond in valid JSON format ONLY, without any extra text.
  The JSON must contain two keys:
  1. "short_name": a short, descriptive name for the image.
  2. "description": a detailed description of the image in English.

  Example output format:
  {"short_name": "Sunset on the beach", "description": "The image shows a sunset at sea..."}
supported_extensions:
  - ".png"
  - ".jpg"
  - ".jpeg"
  - ".webp"
  - ".gif"
  - ".bmp"
parallel_requests: 3
max_retries: 3
retry_delay: 5
```

### Performance Optimizations

1. **Connection Pooling**: Reuse HTTP connections for multiple requests
2. **Parallel Processing**: Concurrent API calls for multiple images
3. **Request Batching**: Group similar images for efficiency
4. **Caching**: Cache results for identical images
5. **Timeout Management**: Configurable timeouts for different operations
6. **Retry Logic**: Exponential backoff for failed requests
7. **Rate Limiting**: Respect API rate limits
8. **Memory Management**: Efficient image data handling

## Error Handling Strategy

1. **Network Errors**: Retry with exponential backoff
2. **API Errors**: Fall back to alternative models
3. **Timeout Errors**: Increase timeout and retry
4. **Rate Limiting**: Implement queuing and backoff
5. **Invalid Responses**: Validate and retry malformed responses
6. **Authentication Errors**: Alert for API key issues
7. **Service Unavailable**: Implement circuit breaker pattern

## Cost Optimization

1. **Model Selection**: Use appropriate models for different image types
2. **Batch Processing**: Process similar images together
3. **Caching**: Avoid duplicate processing
4. **Monitoring**: Track API usage and costs
5. **Alternative Models**: Have fallback options for cost optimization
6. **Quality Settings**: Adjust prompt complexity based on needs

## Future Considerations

- **Local Model Support**: Plan for offline processing capabilities
- **Model Fine-tuning**: Consider custom training for specific domains
- **Multi-language Support**: Add support for multiple languages
- **Advanced Features**: Object detection, OCR, scene understanding
- **Cost Management**: Implement budget tracking and alerts
- **Performance Monitoring**: Track API response times and success rates

## Testing Strategy

1. **Unit Tests**: Test individual components (HTTP client, response parsing)
2. **Integration Tests**: Test end-to-end image processing pipeline
3. **Performance Tests**: Measure API response times and throughput
4. **Error Scenarios**: Test various failure conditions
5. **Model Comparison**: Test different models for quality and performance
6. **Load Testing**: Test under high concurrency