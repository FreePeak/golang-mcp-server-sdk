package shared

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequestUnmarshal(t *testing.T) {
	jsonData := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "test.method",
		"params": {"key": "value"}
	}`

	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC to be '2.0', got '%s'", req.JSONRPC)
	}

	if req.ID != float64(1) {
		t.Errorf("Expected ID to be 1, got '%v'", req.ID)
	}

	if req.Method != "test.method" {
		t.Errorf("Expected Method to be 'test.method', got '%s'", req.Method)
	}

	// Check params
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		t.Fatalf("Params is not a map: %T", req.Params)
	}

	value, exists := params["key"]
	if !exists {
		t.Errorf("Expected params to have key 'key'")
	}

	if value != "value" {
		t.Errorf("Expected value to be 'value', got '%v'", value)
	}
}

func TestJSONRPCResponseMarshal(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]string{"status": "success"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc to be '2.0', got '%v'", parsed["jsonrpc"])
	}

	if parsed["id"] != float64(1) {
		t.Errorf("Expected id to be 1, got '%v'", parsed["id"])
	}

	result, ok := parsed["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", parsed["result"])
	}

	if result["status"] != "success" {
		t.Errorf("Expected result.status to be 'success', got '%v'", result["status"])
	}
}

func TestJSONRPCErrorMarshal(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Error: &JSONRPCError{
			Code:    -32000,
			Message: "Test error",
			Data:    "Additional info",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc to be '2.0', got '%v'", parsed["jsonrpc"])
	}

	if parsed["id"] != float64(1) {
		t.Errorf("Expected id to be 1, got '%v'", parsed["id"])
	}

	errObj, ok := parsed["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Error is not a map: %T", parsed["error"])
	}

	if errObj["code"] != float64(-32000) {
		t.Errorf("Expected error.code to be -32000, got '%v'", errObj["code"])
	}

	if errObj["message"] != "Test error" {
		t.Errorf("Expected error.message to be 'Test error', got '%v'", errObj["message"])
	}

	if errObj["data"] != "Additional info" {
		t.Errorf("Expected error.data to be 'Additional info', got '%v'", errObj["data"])
	}
}
