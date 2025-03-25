package server

import (
	"context"
	"fmt"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

// EchoHandler implements a simple echo tool handler for testing
type EchoHandler struct{}

// NewEchoHandler creates a new echo handler
func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

// ListTools returns the echo tool information
func (h *EchoHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
	return []shared.Tool{
		{
			Name:        "echo",
			Description: "Echoes back the input text",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text to echo back",
					},
				},
				"required": []string{"text"},
			},
		},
	}, nil
}

// CallTool executes the echo tool with the given arguments
func (h *EchoHandler) CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error) {
	if name != "echo" {
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("tool '%s' not found", name), nil)
	}

	// Parse arguments
	args, ok := arguments.(map[string]interface{})
	if !ok {
		return nil, mcperrors.NewInvalidInputError("invalid arguments", nil)
	}

	// Get the text to echo
	text, ok := args["text"].(string)
	if !ok {
		return nil, mcperrors.NewInvalidInputError("parameter 'text' must be a string", nil)
	}

	// Create a text content result
	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: text,
		},
	}, nil
}
