package shared

import (
	"testing"
)

func TestTextContent(t *testing.T) {
	content := TextContent{
		Type: "text",
		Text: "test content",
	}

	if content.GetType() != "text" {
		t.Errorf("Expected type 'text', got '%s'", content.GetType())
	}
}

func TestImageContent(t *testing.T) {
	content := ImageContent{
		Type: "image",
		URL:  "https://example.com/image.png",
	}

	if content.GetType() != "image" {
		t.Errorf("Expected type 'image', got '%s'", content.GetType())
	}
}

func TestEmbeddedResource(t *testing.T) {
	content := EmbeddedResource{
		Type: "resource",
		URI:  "resource://example",
	}

	if content.GetType() != "resource" {
		t.Errorf("Expected type 'resource', got '%s'", content.GetType())
	}
}

func TestCapabilities(t *testing.T) {
	// Create capabilities with all options
	capabilities := Capabilities{
		Resources: &ResourcesCapability{},
		Tools:     &ToolsCapability{},
		Prompts:   &PromptsCapability{},
	}

	if capabilities.Resources == nil {
		t.Error("Expected non-nil Resources capability")
	}

	if capabilities.Tools == nil {
		t.Error("Expected non-nil Tools capability")
	}

	if capabilities.Prompts == nil {
		t.Error("Expected non-nil Prompts capability")
	}

	// Create capabilities with some options
	partialCapabilities := Capabilities{
		Resources: &ResourcesCapability{},
	}

	if partialCapabilities.Resources == nil {
		t.Error("Expected non-nil Resources capability")
	}

	if partialCapabilities.Tools != nil {
		t.Error("Expected nil Tools capability")
	}

	if partialCapabilities.Prompts != nil {
		t.Error("Expected nil Prompts capability")
	}
}
