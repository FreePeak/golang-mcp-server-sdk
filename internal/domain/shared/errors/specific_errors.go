package errors

import "fmt"

// ResourceNotFoundError indicates a resource was not found
type ResourceNotFoundError struct {
	URI string
}

// Error returns the error message
func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.URI)
}

// ResourceAccessDeniedError indicates access to a resource was denied
type ResourceAccessDeniedError struct {
	URI string
}

// Error returns the error message
func (e *ResourceAccessDeniedError) Error() string {
	return fmt.Sprintf("resource access denied: %s", e.URI)
}

// ToolNotFoundError indicates a tool was not found
type ToolNotFoundError struct {
	Name string
}

// Error returns the error message
func (e *ToolNotFoundError) Error() string {
	return fmt.Sprintf("tool not found: %s", e.Name)
}

// ToolExecutionError indicates a tool execution failed
type ToolExecutionError struct {
	Name  string
	Cause error
}

// Error returns the error message
func (e *ToolExecutionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("tool execution failed: %s: %v", e.Name, e.Cause)
	}
	return fmt.Sprintf("tool execution failed: %s", e.Name)
}

// Unwrap returns the underlying cause
func (e *ToolExecutionError) Unwrap() error {
	return e.Cause
}

// PromptNotFoundError indicates a prompt was not found
type PromptNotFoundError struct {
	Name string
}

// Error returns the error message
func (e *PromptNotFoundError) Error() string {
	return fmt.Sprintf("prompt not found: %s", e.Name)
}

// PromptExecutionError indicates a prompt execution failed
type PromptExecutionError struct {
	Name  string
	Cause error
}

// Error returns the error message
func (e *PromptExecutionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("prompt execution failed: %s: %v", e.Name, e.Cause)
	}
	return fmt.Sprintf("prompt execution failed: %s", e.Name)
}

// Unwrap returns the underlying cause
func (e *PromptExecutionError) Unwrap() error {
	return e.Cause
}
