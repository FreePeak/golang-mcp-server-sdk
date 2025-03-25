package errors

import (
	"fmt"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	// Ensure error types are correctly defined
	if ErrorTypeNotFound != "not_found" {
		t.Errorf("Expected ErrorTypeNotFound to be 'not_found', got '%s'", ErrorTypeNotFound)
	}

	if ErrorTypeInvalidInput != "invalid_input" {
		t.Errorf("Expected ErrorTypeInvalidInput to be 'invalid_input', got '%s'", ErrorTypeInvalidInput)
	}

	if ErrorTypeUnauthorized != "unauthorized" {
		t.Errorf("Expected ErrorTypeUnauthorized to be 'unauthorized', got '%s'", ErrorTypeUnauthorized)
	}

	if ErrorTypeInternal != "internal" {
		t.Errorf("Expected ErrorTypeInternal to be 'internal', got '%s'", ErrorTypeInternal)
	}
}

func TestNewNotFoundError(t *testing.T) {
	msg := "resource not found"
	cause := fmt.Errorf("original error")
	err := NewNotFoundError(msg, cause)

	if err.Type != ErrorTypeNotFound {
		t.Errorf("Expected Type to be '%s', got '%s'", ErrorTypeNotFound, err.Type)
	}

	if err.Message != msg {
		t.Errorf("Expected Message to be '%s', got '%s'", msg, err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected Cause to be '%v', got '%v'", cause, err.Cause)
	}

	if !IsNotFound(err) {
		t.Error("IsNotFound should return true for not found errors")
	}

	if IsInvalidInput(err) {
		t.Error("IsInvalidInput should return false for not found errors")
	}
}

func TestNewInvalidInputError(t *testing.T) {
	msg := "invalid input"
	cause := fmt.Errorf("original error")
	err := NewInvalidInputError(msg, cause)

	if err.Type != ErrorTypeInvalidInput {
		t.Errorf("Expected Type to be '%s', got '%s'", ErrorTypeInvalidInput, err.Type)
	}

	if err.Message != msg {
		t.Errorf("Expected Message to be '%s', got '%s'", msg, err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected Cause to be '%v', got '%v'", cause, err.Cause)
	}

	if !IsInvalidInput(err) {
		t.Error("IsInvalidInput should return true for invalid input errors")
	}

	if IsNotFound(err) {
		t.Error("IsNotFound should return false for invalid input errors")
	}
}

func TestErrorWithoutCause(t *testing.T) {
	msg := "error without cause"
	err := &MCPError{
		Type:    ErrorTypeInternal,
		Message: msg,
		Cause:   nil,
	}

	expected := fmt.Sprintf("%s: %s", ErrorTypeInternal, msg)
	if err.Error() != expected {
		t.Errorf("Expected error message to be '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorWithCause(t *testing.T) {
	msg := "error with cause"
	cause := fmt.Errorf("original error")
	err := &MCPError{
		Type:    ErrorTypeInternal,
		Message: msg,
		Cause:   cause,
	}

	expected := fmt.Sprintf("%s: %s: %v", ErrorTypeInternal, msg, cause)
	if err.Error() != expected {
		t.Errorf("Expected error message to be '%s', got '%s'", expected, err.Error())
	}
}

func TestIsErrorTypeCheckers(t *testing.T) {
	// Create different error types
	notFoundErr := NewNotFoundError("not found", nil)
	invalidInputErr := NewInvalidInputError("invalid input", nil)
	unauthorizedErr := NewUnauthorizedError("unauthorized", nil)
	internalErr := NewInternalError("internal", nil)

	// Regular error
	regularErr := fmt.Errorf("regular error")

	// Test not found checks
	if !IsNotFound(notFoundErr) {
		t.Error("IsNotFound should return true for not found errors")
	}
	if IsNotFound(invalidInputErr) || IsNotFound(unauthorizedErr) || IsNotFound(internalErr) || IsNotFound(regularErr) {
		t.Error("IsNotFound should return false for other error types")
	}

	// Test invalid input checks
	if !IsInvalidInput(invalidInputErr) {
		t.Error("IsInvalidInput should return true for invalid input errors")
	}
	if IsInvalidInput(notFoundErr) || IsInvalidInput(unauthorizedErr) || IsInvalidInput(internalErr) || IsInvalidInput(regularErr) {
		t.Error("IsInvalidInput should return false for other error types")
	}

	// Test unauthorized checks
	if !IsUnauthorized(unauthorizedErr) {
		t.Error("IsUnauthorized should return true for unauthorized errors")
	}
	if IsUnauthorized(notFoundErr) || IsUnauthorized(invalidInputErr) || IsUnauthorized(internalErr) || IsUnauthorized(regularErr) {
		t.Error("IsUnauthorized should return false for other error types")
	}

	// Test internal checks
	if !IsInternal(internalErr) {
		t.Error("IsInternal should return true for internal errors")
	}
	if IsInternal(notFoundErr) || IsInternal(invalidInputErr) || IsInternal(unauthorizedErr) || IsInternal(regularErr) {
		t.Error("IsInternal should return false for other error types")
	}
}
