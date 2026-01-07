# ADR-0006: Implement Comprehensive Error Handling Strategy

- Status: accepted
- Date: 2026-01-08
- Deciders: Sergey Grachev

## Context

Error handling is critical for building reliable and maintainable software. The KBase Image Catalog project involves multiple complex operations (file I/O, network calls, AI processing, web serving) that can fail in various ways. A consistent error handling strategy ensures system resilience, good user experience, and effective debugging.

Key considerations:
- **Error Types**: Different kinds of errors (user input, system failures, network issues)
- **Error Propagation**: How errors flow through the system
- **User Experience**: How errors are presented to users
- **Debugging**: How errors are logged and tracked
- **Recovery**: System resilience and error recovery mechanisms
- **Testing**: How errors are tested and validated
- **Monitoring**: Error tracking and alerting
- **Performance**: Error handling overhead
- **Code Maintainability**: Clear error handling patterns
- **Integration**: Error handling across different components

## Decision

We will implement a **comprehensive error handling strategy** using Go's error handling idioms, custom error types, structured logging, and graceful degradation patterns.

## Consequences

### Positive Consequences

1. **System Resilience**: Graceful handling of failures prevents crashes
2. **Better Debugging**: Structured error information aids troubleshooting
3. **Improved User Experience**: Clear error messages and fallback behaviors
4. **Code Consistency**: Uniform error handling patterns throughout the codebase
5. **Effective Monitoring**: Easy tracking and alerting on error patterns
6. **Testability**: Clear error boundaries make testing easier
7. **Recovery Mechanisms**: Built-in retry and fallback logic
8. **Performance Optimization**: Error handling with minimal overhead
9. **Documentation**: Error types serve as implicit documentation
10. **Maintainability**: Clear error handling patterns make code easier to maintain

### Negative Consequences

1. **Code Verbosity**: Error handling can make code more verbose
2. **Complexity**: Multiple error types and handling patterns add complexity
3. **Performance Overhead**: Error creation and handling has some cost
4. **Learning Curve**: Team needs to understand error handling patterns
5. **Testing Overhead**: More comprehensive error testing required
6. **Documentation Effort**: Error types need documentation
7. **Error Propagation**: Complex error chains can be hard to trace
8. **Graceful Degradation**: Can mask underlying issues
9. **Error Recovery**: Recovery logic can be complex to implement
10. **Error Logging**: Too much logging can create noise

## Alternatives Considered

### 1. Panic-Based Error Handling
**Rejected** because:
- Violates Go's error handling philosophy
- Can cause application crashes
- Poor user experience
- Harder to test and recover from
- Not idiomatic Go practice

### 2. Minimal Error Handling (Return nil only)
**Rejected** because:
- Hides potential issues
- Makes debugging difficult
- Poor system reliability
- No recovery mechanisms
- Bad user experience

### 3. External Error Library (like pkg/errors)
**Rejected** because:
- Adds unnecessary dependencies
- Go 1.13+ built-in error wrapping is sufficient
- Can create dependency issues
- Less control over error types
- Overkill for current needs

### 4. Global Error Handler
**Rejected** because:
- Centralized error handling is inflexible
- Hard to customize per error type
- Can create bottlenecks
- Difficult to test
- Not idiomatic Go pattern

### 5. Exception-Based Approach
**Rejected** because:
- Not idiomatic Go
- Violates Go's error handling philosophy
- Can lead to unpredictable control flow
- Harder to reason about error paths
- Performance implications

## Related ADRs

- [ADR-0002: Use Go for Backend Development](./ADR-0002-use-go-for-backend.md) - Backend technology that implements error handling
- [ADR-0004: Integrate LLM Models for Image Analysis](./ADR-0004-llm-integration.md) - Error handling for AI API calls
- [ADR-0007: Focus on Performance Optimization](./ADR-0007-performance-optimization.md) - Error handling performance considerations
- [ADR-0008: Adopt Security-First Approach](./ADR-0008-security-approach.md) - Error handling security implications

## Implementation Details

### Error Type Hierarchy

```go
// Base error types
type KBaseError struct {
    Code        string
    Message     string
    Details     interface{}
    StackTrace  string
    Timestamp   time.Time
    Context     context.Context
}

// Specific error types
type ConfigError struct {
    KBaseError
    Field       string
    Value       interface{}
}

type FileNotFoundError struct {
    KBaseError
    Path        string
    IsDirectory bool
}

type NetworkError struct {
    KBaseError
    StatusCode  int
    URL         string
    Retryable   bool
}

type ProcessingError struct {
    KBaseError
    FileName    string
    FileSize    int64
    ProcessingStep string
}

type ValidationError struct {
    KBaseError
    Field       string
    Value       interface{}
    Constraint  string
}
```

### Error Handling Patterns

1. **Error Wrapping**: Use `%w` verb to wrap errors with context
2. **Error Checking**: Explicit error checking with `if err != nil`
3. **Error Creation**: Custom error types for different scenarios
4. **Error Recovery**: Retry mechanisms with exponential backoff
5. **Error Logging**: Structured logging with error context
6. **Error Propagation**: Pass errors up the call stack with context
7. **Error Transformation**: Convert errors to appropriate types
8. **Error Aggregation**: Collect multiple errors for batch processing

### Error Handling by Component

#### File System Operations
```go
func ScanDirectory(ctx context.Context, rootPath string) ([]string, error) {
    var imageFiles []string
    
    err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        if err != nil {
            return &FileNotFoundError{
                KBaseError: KBaseError{
                    Code:      "FILE_SCAN_ERROR",
                    Message:   "Failed to scan file",
                    Details:   map[string]interface{}{
                        "path": path,
                        "error": err.Error(),
                    },
                    Timestamp: time.Now(),
                    Context:   ctx,
                },
                Path:        path,
                IsDirectory: false,
            }
        }
        
        // Process files...
        return nil
    })
    
    return imageFiles, err
}
```

#### AI API Calls
```go
func AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (*LLMResponse, error) {
    // Validate input
    if len(imageData) == 0 {
        return nil, &ValidationError{
            KBaseError: KBaseError{
                Code:    "INVALID_INPUT",
                Message: "Image data cannot be empty",
                Details: map[string]interface{}{
                    "image_size": len(imageData),
                },
                Timestamp: time.Now(),
                Context:   ctx,
            },
            Field:     "image_data",
            Value:     len(imageData),
            Constraint: "non-empty",
        }
    }
    
    // Make API call with retries
    var response *LLMResponse
    var err error
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        response, err = callLLMAPI(ctx, imageData, prompt)
        if err == nil {
            return response, nil
        }
        
        // Check if error is retryable
        if isRetryableError(err) {
            time.Sleep(time.Duration(attempt+1) * retryDelay)
            continue
        }
        
        return nil, &NetworkError{
            KBaseError: KBaseError{
                Code:      "LLM_API_ERROR",
                Message:   "Failed to analyze image",
                Details:   map[string]interface{}{
                    "attempt": attempt + 1,
                    "error":   err.Error(),
                },
                Timestamp: time.Now(),
                Context:   ctx,
            },
            StatusCode: 0,
            URL:        apiURL,
            Retryable:  true,
        }
    }
    
    return nil, &ProcessingError{
        KBaseError: KBaseError{
            Code:      "PROCESSING_FAILURE",
            Message:   "Failed to process image after multiple attempts",
            Details:   map[string]interface{}{
                "total_attempts": maxRetries,
                "final_error":    err.Error(),
            },
            Timestamp: time.Now(),
            Context:   ctx,
        },
        FileName:    "unknown",
        FileSize:   int64(len(imageData)),
        ProcessingStep: "llm_analysis",
    }
}
```

#### Web Server Error Handling
```go
func (s *Server) handleCatalogDetail(w http.ResponseWriter, r *http.Request) {
    catalogName := strings.TrimPrefix(r.URL.Path, "/catalog/")
    
    if catalogName == "" {
        s.renderError(w, r, &ValidationError{
            KBaseError: KBaseError{
                Code:    "INVALID_REQUEST",
                Message: "Catalog name is required",
                Details: map[string]interface{}{
                    "path": r.URL.Path,
                },
                Timestamp: time.Now(),
                Context:   r.Context(),
            },
            Field:     "catalog_name",
            Value:     "",
            Constraint: "non-empty",
        }, http.StatusBadRequest)
        return
    }
    
    indexData, err := s.catalogService.GetCatalogImages(r.Context(), catalogName)
    if err != nil {
        s.renderError(w, r, &FileNotFoundError{
            KBaseError: KBaseError{
                Code:    "CATALOG_NOT_FOUND",
                Message: "Catalog not found",
                Details: map[string]interface{}{
                    "catalog": catalogName,
                    "error":   err.Error(),
                },
                Timestamp: time.Now(),
                Context:   r.Context(),
            },
            Path:        catalogName,
            IsDirectory: true,
        }, http.StatusNotFound)
        return
    }
    
    // Render template...
}
```

### Error Recovery Strategies

1. **Retry with Exponential Backoff**: For transient failures
2. **Circuit Breaker Pattern**: For service dependencies
3. **Graceful Degradation**: Provide fallback functionality
4. **Timeout Management**: Prevent hanging operations
5. **Rate Limiting**: Handle API rate limits
6. **Fallback Services**: Alternative processing paths
7. **Batch Processing**: Retry failed batches separately
8. **User Notification**: Inform users of issues

### Error Logging and Monitoring

```go
// Structured logging
func logError(ctx context.Context, err error) {
    if kbaseErr, ok := err.(*KBaseError); ok {
        log.WithFields(log.Fields{
            "error_code":    kbaseErr.Code,
            "error_message": kbaseErr.Message,
            "details":       kbaseErr.Details,
            "timestamp":     kbaseErr.Timestamp,
            "context":       ctx,
        }).Error(kbaseErr.Message)
    } else {
        log.WithFields(log.Fields{
            "error":     err.Error(),
            "timestamp": time.Now(),
            "context":   ctx,
        }).Error("Unexpected error")
    }
}

// Error metrics
func trackError(err error) {
    if kbaseErr, ok := err.(*KBaseError); ok {
        metrics.IncrementCounter("errors_total", map[string]string{
            "code": kbaseErr.Code,
        })
    } else {
        metrics.IncrementCounter("errors_total", map[string]string{
            "code": "UNKNOWN",
        })
    }
}
```

### Error Testing Strategy

1. **Unit Tests**: Test individual error handling functions
2. **Integration Tests**: Test error propagation through components
3. **Error Scenarios**: Test various failure conditions
4. **Recovery Tests**: Test retry and fallback mechanisms
5. **Performance Tests**: Measure error handling overhead
6. **Load Tests**: Test error handling under high load

## Error Handling Guidelines

### When to Create Custom Errors

1. **Different Recovery Strategies**: When errors need different handling
2. **User-Facing Messages**: When different error messages are needed
3. **Error Classification**: When errors need to be categorized
4. **Context Information**: When additional context is helpful
5. **Retry Logic**: When errors have different retry characteristics

### Error Message Guidelines

1. **User-Friendly**: Clear and understandable to non-technical users
2. **Actionable**: Tell users what they can do to fix the issue
3. **Concise**: Provide essential information without overload
4. **Consistent**: Use similar phrasing for similar errors
5. **Professional**: Maintain professional tone

### Error Logging Guidelines

1. **Structured**: Use consistent format for all error logs
2. **Contextual**: Include relevant context information
3. **Performance**: Avoid expensive operations in error paths
4. **Security**: Don't log sensitive information
5. **Correlation**: Include request IDs for traceability

## Future Considerations

- **Error Aggregation**: Implement error aggregation for batch operations
- **Advanced Retry Logic**: Implement more sophisticated retry strategies
- **Error Analytics**: Add error analytics and reporting
- **Automated Recovery**: Implement automated recovery mechanisms
- **Error Budgets**: Implement error budget tracking for SLOs

## Maintenance

- **Regular Review**: Periodically review error handling effectiveness
- **Error Metrics**: Monitor error rates and patterns
- **User Feedback**: Collect user feedback on error messages
- **Performance Impact**: Assess error handling performance impact
- **Code Quality**: Maintain error handling code quality