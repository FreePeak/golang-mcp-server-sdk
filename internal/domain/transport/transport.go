package transport

import (
	"context"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// MessageHandler is a function that handles incoming messages
type MessageHandler func(ctx context.Context, message shared.JSONRPCMessage) error

// Transport defines the interface for MCP transports
type Transport interface {
	// Start starts the transport with the given message handler
	Start(ctx context.Context, handler MessageHandler) error

	// Send sends a message through the transport
	Send(ctx context.Context, message shared.JSONRPCMessage) error

	// Close closes the transport
	Close() error
}

// TransportFactory creates transports
type TransportFactory interface {
	// CreateTransport creates a new transport instance
	CreateTransport() (Transport, error)
}
