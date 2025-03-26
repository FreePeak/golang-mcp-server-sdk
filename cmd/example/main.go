package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
)

// handleCalculate handles the calculate tool request
func handleCalculate(params map[string]interface{}) (interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	a, ok := params["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("a must be a number")
	}

	b, ok := params["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("b must be a number")
	}

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return result, nil
}

// createMessageHandler creates a message handler with access to the transport
func createMessageHandler(transport transport.Transport) transport.MessageHandler {
	return func(ctx context.Context, message shared.JSONRPCMessage) error {
		switch msg := message.(type) {
		case shared.JSONRPCRequest:
			switch msg.Method {
			case "initialize":
				response := shared.JSONRPCResponse{
					JSONRPC: shared.JSONRPCVersion,
					ID:      msg.ID,
					Result: map[string]interface{}{
						"serverInfo": shared.ServerInfo{
							Name:    "golang-mcp-server",
							Version: "1.0.0",
							Metadata: map[string]interface{}{
								"transport": "sse",
								"baseUrl":   "http://127.0.0.1:8080",
								"endpoints": map[string]string{
									"sse":     "/mcp/sse",
									"message": "/mcp/message",
								},
							},
						},
						"capabilities": shared.Capabilities{
							Tools: &shared.ToolsCapability{
								ListChanged: true,
							},
						},
					},
				}
				return transport.Send(ctx, response)

			case "tools/list":
				response := shared.JSONRPCResponse{
					JSONRPC: shared.JSONRPCVersion,
					ID:      msg.ID,
					Result: map[string]interface{}{
						"tools": []map[string]interface{}{
							{
								"name":        "calculate",
								"version":     "1.0.0",
								"description": "Perform basic arithmetic calculations",
								"status":      "active",
								"category":    "math",
								"capabilities": map[string]interface{}{
									"streaming": false,
									"async":     false,
								},
								"inputSchema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"operation": map[string]interface{}{
											"type":        "string",
											"description": "The arithmetic operation to perform",
											"enum":        []string{"add", "subtract", "multiply", "divide"},
										},
										"a": map[string]interface{}{
											"type":        "number",
											"description": "First number",
										},
										"b": map[string]interface{}{
											"type":        "number",
											"description": "Second number",
										},
									},
									"required": []string{"operation", "a", "b"},
								},
							},
						},
					},
				}

				// Send the response with the correct event type
				eventData := map[string]interface{}{
					"type": "tools/list",
					"data": response,
				}
				data, _ := json.Marshal(eventData)
				return transport.Send(ctx, shared.JSONRPCNotification{
					JSONRPC: shared.JSONRPCVersion,
					Method:  "tools/list",
					Params:  data,
				})

			case "tools/call":
				if msg.Params == nil {
					return fmt.Errorf("missing params")
				}

				toolParams, ok := msg.Params.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid params format")
				}

				name, ok := toolParams["name"].(string)
				if !ok {
					return fmt.Errorf("missing tool name")
				}

				args, ok := toolParams["arguments"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid arguments format")
				}

				var result interface{}
				var err error

				switch name {
				case "calculate":
					result, err = handleCalculate(args)
				default:
					err = fmt.Errorf("unknown tool: %s", name)
				}

				if err != nil {
					errorResponse := shared.JSONRPCResponse{
						JSONRPC: shared.JSONRPCVersion,
						ID:      msg.ID,
						Error: &shared.JSONRPCError{
							Code:    -32603,
							Message: err.Error(),
						},
					}
					return transport.Send(ctx, errorResponse)
				}

				response := shared.JSONRPCResponse{
					JSONRPC: shared.JSONRPCVersion,
					ID:      msg.ID,
					Result: map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": fmt.Sprintf("%v", result),
							},
						},
					},
				}
				return transport.Send(ctx, response)

			default:
				return fmt.Errorf("unknown method: %s", msg.Method)
			}

		case shared.JSONRPCNotification:
			if msg.Method == "initialized" {
				// Client has completed initialization
				fmt.Println("Client initialized")
				return nil
			}
			return fmt.Errorf("unknown notification: %s", msg.Method)

		default:
			return fmt.Errorf("unexpected message type: %T", message)
		}
	}
}

func main() {
	// Parse command line flags
	transportType := flag.String("transport", "stdio", "Transport type (stdio or sse)")
	host := flag.String("host", "localhost", "Host to listen on (for SSE transport)")
	port := flag.Int("port", 8080, "Port to listen on (for SSE transport)")
	flag.Parse()

	// Create transport factory based on type
	var factory transport.TransportFactory
	switch *transportType {
	case "stdio":
		factory = server.NewStdioTransportFactory()
	case "sse":
		factory = server.NewHTTPTransportFactory(*host, *port)
	default:
		fmt.Fprintf(os.Stderr, "Invalid transport type: %s\n", *transportType)
		os.Exit(1)
	}

	// Create transport
	transport, err := factory.CreateTransport()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create transport: %v\n", err)
		os.Exit(1)
	}

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal")
		cancel()
	}()

	fmt.Printf("Starting MCP server with %s transport\n", *transportType)
	if *transportType == "sse" {
		fmt.Printf("Listening on http://%s:%d\n", *host, *port)
	}

	// Start the transport with the message handler
	if err := transport.Start(ctx, createMessageHandler(transport)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start transport: %v\n", err)
		os.Exit(1)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Cleanup
	if err := transport.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}
}
