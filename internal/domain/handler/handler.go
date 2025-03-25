package handler

import (
	"context"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// RequestHandler defines a function that handles a specific request method
type RequestHandler func(ctx context.Context, params interface{}) (interface{}, error)

// ResourceHandler defines a handler for resources
type ResourceHandler interface {
	// ListResources returns a list of available resources
	ListResources(ctx context.Context) ([]shared.Resource, error)

	// GetResource returns the content of a specific resource
	GetResource(ctx context.Context, uri string) ([]shared.Content, error)
}

// ToolHandler defines a handler for tools
type ToolHandler interface {
	// ListTools returns a list of available tools
	ListTools(ctx context.Context) ([]shared.Tool, error)

	// CallTool executes a tool with the given arguments
	CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error)
}

// PromptHandler defines a handler for prompts
type PromptHandler interface {
	// ListPrompts returns a list of available prompts
	ListPrompts(ctx context.Context) ([]shared.Prompt, error)

	// CallPrompt executes a prompt with the given arguments
	CallPrompt(ctx context.Context, name string, arguments map[string]interface{}) ([]shared.Content, error)
}
