package docs

import (
	"context"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

func TestDocsHandlerListResources(t *testing.T) {
	handler := NewDocsHandler()

	ctx := context.Background()
	resources, err := handler.ListResources(ctx)

	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("Expected 3 resources, got %d", len(resources))
	}

	// Check if all expected resources are present
	resourceURIs := map[string]bool{
		"docs://introduction":   false,
		"docs://tools/add":      false,
		"docs://tools/subtract": false,
	}

	for _, resource := range resources {
		resourceURIs[resource.URI] = true
	}

	for uri, found := range resourceURIs {
		if !found {
			t.Errorf("Expected resource '%s' not found", uri)
		}
	}
}

func TestDocsHandlerGetResource(t *testing.T) {
	handler := NewDocsHandler()
	ctx := context.Background()

	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "introduction",
			uri:     "docs://introduction",
			wantErr: false,
		},
		{
			name:    "tools/add",
			uri:     "docs://tools/add",
			wantErr: false,
		},
		{
			name:    "tools/subtract",
			uri:     "docs://tools/subtract",
			wantErr: false,
		},
		{
			name:    "non-existent resource",
			uri:     "docs://not-found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := handler.GetResource(ctx, tt.uri)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}

				if !mcperrors.IsNotFound(err) {
					t.Errorf("Expected NotFoundError, got %T", err)
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

			if textContent.Text == "" {
				t.Errorf("Expected non-empty text content")
			}
		})
	}
}

func TestDocsHandlerResourceNames(t *testing.T) {
	handler := NewDocsHandler()

	ctx := context.Background()
	resources, err := handler.ListResources(ctx)

	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	// Check that resource names are properly extracted from URIs
	for _, resource := range resources {
		expectedName := resource.URI
		if resource.URI == "docs://introduction" {
			expectedName = "introduction"
		} else if resource.URI == "docs://tools/add" {
			expectedName = "tools/add"
		} else if resource.URI == "docs://tools/subtract" {
			expectedName = "tools/subtract"
		}

		if resource.Name != expectedName {
			t.Errorf("Expected name '%s', got '%s'", expectedName, resource.Name)
		}
	}
}
