package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

// DocsHandler implements a documentation resource handler
type DocsHandler struct {
	resources map[string]string
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{
		resources: map[string]string{
			"docs://introduction": `# MCP Calculator Server

This server provides basic calculator operations as MCP tools.

## Available Tools

- add: Add two numbers
- subtract: Subtract b from a
- multiply: Multiply two numbers
- divide: Divide a by b

## Usage

Each tool takes two parameters:
- a: First number
- b: Second number

Example:
{
  "name": "add",
  "arguments": {
    "a": 5,
    "b": 3
  }
}
`,
			"docs://tools/add": `# Add Tool

Adds two numbers and returns the result.

## Parameters

- a: First number (required)
- b: Second number (required)

## Example

Request:
{
  "name": "add",
  "arguments": {
    "a": 5,
    "b": 3
  }
}

Response:
{
  "content": [
    {
      "type": "text",
      "text": "8.000000"
    }
  ]
}
`,
			"docs://tools/subtract": `# Subtract Tool

Subtracts b from a and returns the result.

## Parameters

- a: First number (required)
- b: Second number (required)

## Example

Request:
{
  "name": "subtract",
  "arguments": {
    "a": 5,
    "b": 3
  }
}

Response:
{
  "content": [
    {
      "type": "text",
      "text": "2.000000"
    }
  ]
}
`,
		},
	}
}

// ListResources returns a list of available resources
func (h *DocsHandler) ListResources(ctx context.Context) ([]shared.Resource, error) {
	resources := make([]shared.Resource, 0, len(h.resources))

	for uri := range h.resources {
		name := uri
		if strings.HasPrefix(uri, "docs://") {
			name = strings.TrimPrefix(uri, "docs://")
		}

		resources = append(resources, shared.Resource{
			URI:         uri,
			Name:        name,
			Description: "Documentation resource",
		})
	}

	return resources, nil
}

// GetResource returns the content of a resource
func (h *DocsHandler) GetResource(ctx context.Context, uri string) ([]shared.Content, error) {
	content, exists := h.resources[uri]
	if !exists {
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("resource '%s' not found", uri), nil)
	}

	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: content,
		},
	}, nil
}
