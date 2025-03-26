package server

import (
	"context"
	"testing"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server/mocks"
)

func TestServerCreation(t *testing.T) {
	server := NewServer("test-server", "1.0.0")

	if server.info.Name != "test-server" {
		t.Errorf("Expected server name to be 'test-server', got '%s'", server.info.Name)
	}

	if server.info.Version != "1.0.0" {
		t.Errorf("Expected server version to be '1.0.0', got '%s'", server.info.Version)
	}

	// Verify handlers are not set
	if server.resourceHandler != nil {
		t.Error("Expected resourceHandler to be nil")
	}

	if server.toolHandler != nil {
		t.Error("Expected toolHandler to be nil")
	}

	if server.promptHandler != nil {
		t.Error("Expected promptHandler to be nil")
	}
}

func TestWithHandlers(t *testing.T) {
	resourceHandler := mocks.NewMockResourceHandler()
	toolHandler := mocks.NewMockToolHandler()
	promptHandler := mocks.NewMockPromptHandler()

	// Create server with handler options
	server := NewServer("test-server", "1.0.0",
		WithResourceHandler(resourceHandler),
		WithToolHandler(toolHandler),
		WithPromptHandler(promptHandler),
	)

	// Verify handlers are set
	if server.resourceHandler != resourceHandler {
		t.Error("Expected resourceHandler to be set")
	}

	if server.toolHandler != toolHandler {
		t.Error("Expected toolHandler to be set")
	}

	if server.promptHandler != promptHandler {
		t.Error("Expected promptHandler to be set")
	}

	// Verify capabilities
	if server.capabilities.Resources == nil {
		t.Error("Expected Resources capability to be set")
	}

	if server.capabilities.Tools == nil {
		t.Error("Expected Tools capability to be set")
	}

	if server.capabilities.Prompts == nil {
		t.Error("Expected Prompts capability to be set")
	}
}

func TestConnectAndStart(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	transport := mocks.NewMockTransport()

	// Connect to transport
	if err := server.Connect(transport); err != nil {
		t.Fatalf("Failed to connect to transport: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !transport.IsStartCalled() {
		t.Error("Expected transport.Start to be called")
	}

	// Stop server
	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if !transport.IsCloseCalled() {
		t.Error("Expected transport.Close to be called")
	}
}

func TestStartWithNoTransport(t *testing.T) {
	server := NewServer("test-server", "1.0.0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server without transport
	err := server.Start(ctx)
	if err == nil {
		t.Fatal("Expected error starting server without transport")
	}
}

func TestInitialize(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	transport := mocks.NewMockTransport()

	// Don't wait for results - the test was hanging here
	transport.SetWaitForResult(false)

	if err := server.Connect(transport); err != nil {
		t.Fatalf("Failed to connect to transport: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Create initialize request
	initRequest := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      1,
		Method:  shared.MethodInitialize,
		Params: shared.InitializeParams{
			ClientInfo: shared.ServerInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Options: shared.InitializationOptions{},
		},
	}

	// Process request
	if err := transport.ProcessMessage(ctx, initRequest); err != nil {
		t.Fatalf("Failed to process initialize message: %v", err)
	}

	// Give the server a moment to process the message and send the response
	// (since we're not waiting for results synchronously)
	transport.WaitForMessageCount(1)

	// Check response
	messages := transport.GetMessagesSent()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	response, ok := messages[0].(shared.JSONRPCResponse)
	if !ok {
		t.Fatalf("Expected JSONRPCResponse, got %T", messages[0])
	}

	if response.ID != 1 {
		t.Errorf("Expected ID to be 1, got %v", response.ID)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	// The response.Result could be either a map[string]interface{} or a shared.InitializeResult
	// Let's handle both cases
	var serverInfo shared.ServerInfo

	switch result := response.Result.(type) {
	case shared.InitializeResult:
		serverInfo = result.ServerInfo
	case map[string]interface{}:
		serverInfoMap, ok := result["serverInfo"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected serverInfo to be a map, got %T", result["serverInfo"])
		}

		nameVal, ok := serverInfoMap["name"]
		if !ok {
			t.Fatalf("Expected serverInfo.name to exist")
		}

		versionVal, ok := serverInfoMap["version"]
		if !ok {
			t.Fatalf("Expected serverInfo.version to exist")
		}

		serverInfo = shared.ServerInfo{
			Name:    nameVal.(string),
			Version: versionVal.(string),
		}
	default:
		t.Fatalf("Expected result to be a map or InitializeResult, got %T", response.Result)
	}

	if serverInfo.Name != "test-server" {
		t.Errorf("Expected serverInfo.name to be 'test-server', got '%v'", serverInfo.Name)
	}

	if serverInfo.Version != "1.0.0" {
		t.Errorf("Expected serverInfo.version to be '1.0.0', got '%v'", serverInfo.Version)
	}
}

func TestResourceMethods(t *testing.T) {
	resourceHandler := mocks.NewMockResourceHandler()
	resourceHandler.AddResource("resource://test", "Test Resource", "A test resource", []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: "Test content",
		},
	})

	server := NewServer("test-server", "1.0.0",
		WithResourceHandler(resourceHandler),
	)

	transport := mocks.NewMockTransport()
	transport.SetWaitForResult(false)

	if err := server.Connect(transport); err != nil {
		t.Fatalf("Failed to connect to transport: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Initialize server
	initRequest := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      1,
		Method:  shared.MethodInitialize,
		Params: shared.InitializeParams{
			ClientInfo: shared.ServerInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			Options: shared.InitializationOptions{},
		},
	}

	if err := transport.ProcessMessage(ctx, initRequest); err != nil {
		t.Fatalf("Failed to process initialize message: %v", err)
	}

	// Wait for the initialize response
	transport.WaitForMessageCount(1)

	// Reset transport messages
	transport.ClearMessagesSent()

	// Test resources/list
	listResourcesRequest := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      2,
		Method:  shared.MethodListResources,
	}

	if err := transport.ProcessMessage(ctx, listResourcesRequest); err != nil {
		t.Fatalf("Failed to process list resources message: %v", err)
	}

	transport.WaitForMessageCount(1)

	messages := transport.GetMessagesSent()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	listResponse, ok := messages[0].(shared.JSONRPCResponse)
	if !ok {
		t.Fatalf("Expected JSONRPCResponse, got %T", messages[0])
	}

	if listResponse.ID != 2 {
		t.Errorf("Expected ID to be 2, got %v", listResponse.ID)
	}

	if listResponse.Error != nil {
		t.Errorf("Expected no error, got %v", listResponse.Error)
	}

	// Reset transport messages
	transport.ClearMessagesSent()

	// Test resources/get
	getResourceRequest := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      3,
		Method:  shared.MethodGetResource,
		Params: shared.GetResourceParams{
			URI: "resource://test",
		},
	}

	if err := transport.ProcessMessage(ctx, getResourceRequest); err != nil {
		t.Fatalf("Failed to process get resource message: %v", err)
	}

	transport.WaitForMessageCount(1)

	messages = transport.GetMessagesSent()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	getResponse, ok := messages[0].(shared.JSONRPCResponse)
	if !ok {
		t.Fatalf("Expected JSONRPCResponse, got %T", messages[0])
	}

	if getResponse.ID != 3 {
		t.Errorf("Expected ID to be 3, got %v", getResponse.ID)
	}

	if getResponse.Error != nil {
		t.Errorf("Expected no error, got %v", getResponse.Error)
	}
}
