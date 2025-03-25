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

// MCPError represents an error in the MCP protocol
type MCPError struct {
	Code    ErrorCode
	Message string
	Data    interface{}
}

// Error returns the error message
func (e *MCPError) Error() string {
	return e.Message
}

// ErrorCode represents the type of error
type ErrorCode int

const (
	// NotFound represents a resource/tool/prompt not found error
	NotFound ErrorCode = iota
	// InvalidInput represents an invalid input error
	InvalidInput
	// InternalError represents an internal server error
	InternalError
	// Unauthorized represents an unauthorized error
	Unauthorized
)

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, data interface{}) *MCPError {
	return &MCPError{
		Code:    NotFound,
		Message: message,
		Data:    data,
	}
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(message string, data interface{}) *MCPError {
	return &MCPError{
		Code:    InvalidInput,
		Message: message,
		Data:    data,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, data interface{}) *MCPError {
	return &MCPError{
		Code:    InternalError,
		Message: message,
		Data:    data,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, data interface{}) *MCPError {
	return &MCPError{
		Code:    Unauthorized,
		Message: message,
		Data:    data,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	var mcpErr *MCPError
	if ok := (err).(*MCPError) == mcpErr; !ok {
		return &MCPError{
			Code:    InternalError,
			Message: fmt.Sprintf("%s: %v", message, err),
		}
	}

	mcpErr = err.(*MCPError)
	return &MCPError{
		Code:    mcpErr.Code,
		Message: fmt.Sprintf("%s: %s", message, mcpErr.Message),
		Data:    mcpErr.Data,
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Code == NotFound
	}
	return false
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Code == InvalidInput
	}
	return false
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Code == Unauthorized
	}
	return false
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr.Code == InternalError
	}
	return false
}
