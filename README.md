# MCP Server SDK

A Go implementation of the Model Context Protocol (MCP) server, allowing you to create MCP-compatible servers that can be used with AI assistants like Claude.

## Features

- Full implementation of the MCP protocol
- Support for resources, tools, and prompts
- Modular design using clean architecture
- Multiple transport options (stdio, HTTP)
- Easy-to-use API for creating MCP servers

## Installation

```bash
go get github.com/FreePeak/golang-mcp-server-sdk
```

## Quick Start

Here's a simple example of creating an MCP server with a calculator tool:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
    "github.com/FreePeak/golang-mcp-server-sdk/internal/usecases/calculator"
)

func main() {
    // Create a context that cancels on SIGINT/SIGTERM
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Set up signal handling
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        fmt.Println("\nShutting down...")
        cancel()
    }()

    // Create the MCP server
    srv := server.NewServer("calculator-server", "1.0.0")

    // Add the calculator tool handler
    srv.WithToolHandler(calculator.NewCalculatorHandler())

    // Use stdio transport
    err := srv.Connect(server.NewStdioTransport())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error connecting transport: %v\n", err)
        os.Exit(1)
    }

    // Start the server
    err = srv.Start(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
        os.Exit(1)
    }

    // Wait for context cancellation (from signal handler)
    <-ctx.Done()

    // Stop the server
    if err := srv.Stop(); err != nil {
        fmt.Fprintf(os.Stderr, "Error stopping server: %v\n", err)
    }
}
```

## Running the Example Server

The example calculator server can be run in two modes:

### Stdio mode (default)

```bash
go run cmd/server/main.go
```

This will start the server using the stdio transport, which can be integrated with tools like Claude.

### HTTP mode

```bash
go run cmd/server/main.go -http=:8080 -stdio=false
```

This will start the server on HTTP port 8080, allowing you to interact with it using HTTP requests or Server-Sent Events (SSE).

## Creating Your Own MCP Server

### 1. Define Handlers

First, implement the handlers for your server's capabilities:

#### Tool Handler Example

```go
type MyToolHandler struct{}

// ListTools returns a list of available tools
func (h *MyToolHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
    return []shared.Tool{
        {
            Name:        "myTool",
            Description: "A custom tool",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "input": map[string]interface{}{
                        "type": "string",
                    },
                },
                "required": []string{"input"},
            },
        },
    }, nil
}

// CallTool executes a tool with the given arguments
func (h *MyToolHandler) CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error) {
    if name != "myTool" {
        return nil, errors.New("tool not found")
    }

    args, ok := arguments.(map[string]interface{})
    if !ok {
        return nil, errors.New("invalid arguments")
    }

    input, ok := args["input"].(string)
    if !ok {
        return nil, errors.New("invalid input parameter")
    }

    return []shared.Content{
        shared.TextContent{
            Type: "text",
            Text: fmt.Sprintf("Processed: %s", input),
        },
    }, nil
}
```

### 2. Create Server with Handlers

```go
// Create the MCP server
srv := server.NewServer("my-server", "1.0.0")

// Add handlers
srv.WithToolHandler(NewMyToolHandler())
srv.WithResourceHandler(NewMyResourceHandler())
srv.WithPromptHandler(NewMyPromptHandler())

// Connect with the appropriate transport
err := srv.Connect(server.NewStdioTransport())
if err != nil {
    log.Fatalf("Error connecting transport: %v", err)
}

// Start the server
err = srv.Start(ctx)
if err != nil {
    log.Fatalf("Error starting server: %v", err)
}
```

## Transport Options

### Stdio Transport

The stdio transport uses standard input/output for communication, making it ideal for integration with desktop applications.

```go
srv.Connect(server.NewStdioTransport())
```

### HTTP Transport

The HTTP transport provides a web interface, supporting both standard HTTP requests and Server-Sent Events (SSE) for streaming.

```go
srv.Connect(server.NewHTTPTransport(":8080"))
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 