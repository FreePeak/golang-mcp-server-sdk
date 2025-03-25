package calculator

import (
	"context"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

func TestCalculatorHandlerListTools(t *testing.T) {
	handler := NewCalculatorHandler()

	ctx := context.Background()
	tools, err := handler.ListTools(ctx)

	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) != 4 {
		t.Fatalf("Expected 4 tools, got %d", len(tools))
	}

	// Check if all expected tools are present
	toolNames := map[string]bool{
		"add":      false,
		"subtract": false,
		"multiply": false,
		"divide":   false,
	}

	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for name, found := range toolNames {
		if !found {
			t.Errorf("Expected tool '%s' not found", name)
		}
	}
}

func TestCalculatorHandlerCallTool(t *testing.T) {
	handler := NewCalculatorHandler()
	ctx := context.Background()

	tests := []struct {
		name      string
		toolName  string
		args      map[string]interface{}
		wantValue string
		wantErr   bool
	}{
		{
			name:     "add",
			toolName: "add",
			args: map[string]interface{}{
				"a": float64(5),
				"b": float64(3),
			},
			wantValue: "8.000000",
			wantErr:   false,
		},
		{
			name:     "subtract",
			toolName: "subtract",
			args: map[string]interface{}{
				"a": float64(5),
				"b": float64(3),
			},
			wantValue: "2.000000",
			wantErr:   false,
		},
		{
			name:     "multiply",
			toolName: "multiply",
			args: map[string]interface{}{
				"a": float64(5),
				"b": float64(3),
			},
			wantValue: "15.000000",
			wantErr:   false,
		},
		{
			name:     "divide",
			toolName: "divide",
			args: map[string]interface{}{
				"a": float64(6),
				"b": float64(3),
			},
			wantValue: "2.000000",
			wantErr:   false,
		},
		{
			name:     "divide by zero",
			toolName: "divide",
			args: map[string]interface{}{
				"a": float64(6),
				"b": float64(0),
			},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:     "unknown tool",
			toolName: "unknown",
			args: map[string]interface{}{
				"a": float64(5),
				"b": float64(3),
			},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:     "missing parameter a",
			toolName: "add",
			args: map[string]interface{}{
				"b": float64(3),
			},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:     "missing parameter b",
			toolName: "add",
			args: map[string]interface{}{
				"a": float64(5),
			},
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := handler.CallTool(ctx, tt.toolName, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}

				// Check for specific error types
				if tt.toolName == "unknown" && !mcperrors.IsNotFound(err) {
					t.Errorf("Expected NotFoundError, got %T", err)
				} else if tt.toolName == "divide" && !mcperrors.IsInvalidInput(err) {
					t.Errorf("Expected InvalidInputError, got %T", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(content) != 1 {
				t.Fatalf("Expected 1 content item, got %d", len(content))
			}

			textContent, ok := content[0].(shared.TextContent)
			if !ok {
				t.Fatalf("Expected TextContent, got %T", content[0])
			}

			if textContent.Type != "text" {
				t.Errorf("Expected type 'text', got '%s'", textContent.Type)
			}

			if textContent.Text != tt.wantValue {
				t.Errorf("Expected value '%s', got '%s'", tt.wantValue, textContent.Text)
			}
		})
	}
}

func TestCalculatorHandlerInvalidArguments(t *testing.T) {
	handler := NewCalculatorHandler()
	ctx := context.Background()

	// Test with non-map arguments
	_, err := handler.CallTool(ctx, "add", "invalid")
	if err == nil {
		t.Error("Expected error for invalid arguments, got nil")
	}

	if !mcperrors.IsInvalidInput(err) {
		t.Errorf("Expected InvalidInputError, got %T", err)
	}

	// Test with wrong type arguments
	_, err = handler.CallTool(ctx, "add", map[string]interface{}{
		"a": "not a number",
		"b": float64(3),
	})

	if err == nil {
		t.Error("Expected error for invalid argument type, got nil")
	}

	if !mcperrors.IsInvalidInput(err) {
		t.Errorf("Expected InvalidInputError, got %T", err)
	}
}
