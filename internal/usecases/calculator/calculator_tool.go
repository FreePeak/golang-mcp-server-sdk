package calculator

import (
	"context"
	"fmt"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

// CalculatorHandler implements a basic calculator tool handler
type CalculatorHandler struct{}

// NewCalculatorHandler creates a new calculator handler
func NewCalculatorHandler() *CalculatorHandler {
	return &CalculatorHandler{}
}

// ListTools returns a list of available calculator tools
func (h *CalculatorHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
	return []shared.Tool{
		{
			Name:        "add",
			Description: "Add two numbers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "subtract",
			Description: "Subtract b from a",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "multiply",
			Description: "Multiply two numbers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "divide",
			Description: "Divide a by b",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
	}, nil
}

// CallTool executes a calculator tool with the given arguments
func (h *CalculatorHandler) CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error) {
	// Parse arguments
	args, ok := arguments.(map[string]interface{})
	if !ok {
		return nil, mcperrors.NewInvalidInputError("invalid arguments", nil)
	}

	// Get the two numbers
	a, ok := args["a"].(float64)
	if !ok {
		return nil, mcperrors.NewInvalidInputError("parameter 'a' must be a number", nil)
	}

	b, ok := args["b"].(float64)
	if !ok {
		return nil, mcperrors.NewInvalidInputError("parameter 'b' must be a number", nil)
	}

	var result float64

	// Execute the requested operation
	switch name {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, mcperrors.NewInvalidInputError("division by zero", nil)
		}
		result = a / b
	default:
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("tool '%s' not found", name), nil)
	}

	// Create a text content result
	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: fmt.Sprintf("%f", result),
		},
	}, nil
}
