package transport

import (
	"context"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// MessageHandler defines a function that processes incoming messages
type MessageHandler func(ctx context.Context, message shared.JSONRPCMessage) error

// Transport defines the interface for an MCP transport
type Transport interface {
	// Start begins processing messages, with the given handler function
	Start(ctx context.Context, handler MessageHandler) error

	// Send sends a message through the transport
	Send(ctx context.Context, message shared.JSONRPCMessage) error

	// Close stops the transport
	Close() error
}

// Factory creates transports
type Factory interface {
	// CreateTransport creates a new transport instance
	CreateTransport() (Transport, error)
}
