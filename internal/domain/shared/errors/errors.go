package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrorType defines the type of error
type ErrorType string

const (
	// ErrorTypeNotFound indicates a resource, tool, or prompt not found
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeInvalidInput indicates invalid input parameters
	ErrorTypeInvalidInput ErrorType = "invalid_input"
	// ErrorTypeUnauthorized indicates unauthorized access
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal ErrorType = "internal"
)

// MCPError represents an MCP-specific error
type MCPError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *MCPError) Unwrap() error {
	return e.Cause
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, cause error) *MCPError {
	return &MCPError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Cause:   cause,
	}
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(message string, cause error) *MCPError {
	return &MCPError{
		Type:    ErrorTypeInvalidInput,
		Message: message,
		Cause:   cause,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, cause error) *MCPError {
	return &MCPError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *MCPError {
	return &MCPError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Type == ErrorTypeInvalidInput
	}
	return false
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Type == ErrorTypeUnauthorized
	}
	return false
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Type == ErrorTypeInternal
	}
	return false
}
