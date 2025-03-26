package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases/calculator"
)

// SimpleToolHandler is a basic implementation of the ToolHandler interface
type SimpleToolHandler struct{}

// ListTools returns a list of available tools
func (h *SimpleToolHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
	return []shared.Tool{
		{
			Name:        "hello",
			Description: "Says hello to someone",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The name to greet",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "echo",
			Description: "Echoes back the input",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text to echo",
					},
				},
				"required": []string{"text"},
			},
		},
	}, nil
}

// CallTool executes a tool with the given arguments
func (h *SimpleToolHandler) CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error) {
	switch name {
	case "hello":
		args, ok := arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type")
		}

		nameArg, ok := args["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name argument must be a string")
		}

		return []shared.Content{
			shared.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Hello, %s!", nameArg),
			},
		}, nil

	case "echo":
		args, ok := arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type")
		}

		text, ok := args["text"].(string)
		if !ok {
			return nil, fmt.Errorf("text argument must be a string")
		}

		return []shared.Content{
			shared.TextContent{
				Type: "text",
				Text: text,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func main() {
	// Create a new MCP server
	mcp := server.NewServer(
		"example-server",
		"1.0.0",
		server.WithToolHandler(&SimpleToolHandler{}),
		server.WithToolHandler(calculator.NewCalculatorHandler()),
	)

	// Create a transport for the server
	httpTransport, err := server.NewHTTPTransportFactory(":8081").CreateTransport()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create transport: %v\n", err)
		os.Exit(1)
	}

	// Connect the server to the transport
	if err := mcp.Connect(httpTransport); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to transport: %v\n", err)
		os.Exit(1)
	}

	// Create a context with cancellation for server shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server
	if err := mcp.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MCP server started on :8081")
	fmt.Println("HTTP endpoint: http://localhost:8081")
	fmt.Println("SSE endpoint: http://localhost:8081/sse")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for termination signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down...")

	// Stop the server
	if err := mcp.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
	}
}
