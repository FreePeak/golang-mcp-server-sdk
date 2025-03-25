package shared

import (
	"encoding/json"
	"testing"
)

func TestMethodConstants(t *testing.T) {
	// Verify that method constants are not empty
	if MethodInitialize == "" {
		t.Error("Expected MethodInitialize to be non-empty")
	}

	if MethodShutdown == "" {
		t.Error("Expected MethodShutdown to be non-empty")
	}

	if MethodListResources == "" {
		t.Error("Expected MethodListResources to be non-empty")
	}

	if MethodGetResource == "" {
		t.Error("Expected MethodGetResource to be non-empty")
	}

	if MethodListTools == "" {
		t.Error("Expected MethodListTools to be non-empty")
	}

	if MethodCallTool == "" {
		t.Error("Expected MethodCallTool to be non-empty")
	}

	if MethodListPrompts == "" {
		t.Error("Expected MethodListPrompts to be non-empty")
	}

	if MethodCallPrompt == "" {
		t.Error("Expected MethodCallPrompt to be non-empty")
	}
}

func TestInitializeParamsMarshaling(t *testing.T) {
	params := InitializeParams{
		ClientInfo: ServerInfo{
			Name:    "Test Client",
			Version: "1.0.0",
		},
		Options: InitializationOptions{
			Capabilities: Capabilities{
				Resources: &ResourcesCapability{},
			},
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal InitializeParams: %v", err)
	}

	var decoded InitializeParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InitializeParams: %v", err)
	}

	if decoded.ClientInfo.Name != "Test Client" {
		t.Errorf("Expected ClientInfo.Name to be 'Test Client', got '%s'", decoded.ClientInfo.Name)
	}

	if decoded.ClientInfo.Version != "1.0.0" {
		t.Errorf("Expected ClientInfo.Version to be '1.0.0', got '%s'", decoded.ClientInfo.Version)
	}

	if decoded.Options.Capabilities.Resources == nil {
		t.Error("Expected Options.Capabilities.Resources to be non-nil")
	}
}

func TestGetResourceParamsMarshaling(t *testing.T) {
	params := GetResourceParams{
		URI: "resource://test",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal GetResourceParams: %v", err)
	}

	var decoded GetResourceParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal GetResourceParams: %v", err)
	}

	if decoded.URI != "resource://test" {
		t.Errorf("Expected URI to be 'resource://test', got '%s'", decoded.URI)
	}
}

func TestCallToolParamsMarshaling(t *testing.T) {
	params := CallToolParams{
		Name:      "test-tool",
		Arguments: map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal CallToolParams: %v", err)
	}

	var decoded CallToolParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CallToolParams: %v", err)
	}

	if decoded.Name != "test-tool" {
		t.Errorf("Expected Name to be 'test-tool', got '%s'", decoded.Name)
	}

	args, ok := decoded.Arguments.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Arguments to be a map, got %T", decoded.Arguments)
	}

	value, exists := args["key"]
	if !exists {
		t.Error("Expected Arguments to have key 'key'")
	} else if value != "value" {
		t.Errorf("Expected Arguments['key'] to be 'value', got '%v'", value)
	}
}
