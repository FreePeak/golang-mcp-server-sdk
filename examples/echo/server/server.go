package server

import (
	"context"
	"fmt"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
)

// TransportType represents the transport type
type TransportType string

const (
	// StdioTransport is the stdio transport type
	StdioTransport TransportType = "stdio"
	// HTTPTransport is the HTTP transport type
	HTTPTransport TransportType = "http"
)

// EchoServer wraps the MCP server with the echo tool
type EchoServer struct {
	mcpServer *server.Server
	transport transport.Transport
}

// NewEchoServer creates a new echo server
func NewEchoServer() *EchoServer {
	// Create the MCP server with our echo tool
	mcpServer := server.NewServer("echo-server", "1.0.0")
	mcpServer.WithToolHandler(NewEchoHandler())

	return &EchoServer{
		mcpServer: mcpServer,
	}
}

// UseTransport configures the server to use the specified transport
func (s *EchoServer) UseTransport(transportType TransportType, addr string) error {
	var transport transport.Transport

	switch transportType {
	case StdioTransport:
		transport = server.NewStdioTransport()
		fmt.Fprintf(os.Stderr, "Echo server using stdio transport\n")
	case HTTPTransport:
		transport = server.NewHTTPTransport(addr)
		fmt.Fprintf(os.Stderr, "Echo server using HTTP transport on %s\n", addr)
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}

	s.transport = transport
	return s.mcpServer.Connect(transport)
}

// Start starts the echo server
func (s *EchoServer) Start(ctx context.Context) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured, call UseTransport first")
	}

	return s.mcpServer.Start(ctx)
}

// Stop stops the echo server
func (s *EchoServer) Stop() error {
	return s.mcpServer.Stop()
}
