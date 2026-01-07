package errors

import (
	"context"
	"time"
)

// Base error types
type BaseError struct {
	error
	Code       string
	Message    string
	Details    interface{}
	StackTrace string
	Timestamp  time.Time
	Context    context.Context
}

// Specific error types
type ConfigError struct {
	BaseError
	Field string
	Value interface{}
}

type FileNotFoundError struct {
	BaseError
	Path        string
	IsDirectory bool
}

type NetworkError struct {
	BaseError
	StatusCode int
	URL        string
	Retryable  bool
}

type ProcessingError struct {
	BaseError
	FileName       string
	FileSize       int64
	ProcessingStep string
}

type ValidationError struct {
	BaseError
	Field      string
	Value      interface{}
	Constraint string
}

type WebServerError struct {
	BaseError
}
