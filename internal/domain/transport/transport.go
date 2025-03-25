package transport

import (
	"context"
	"sync"

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

var (
	currentHandler MessageHandler
	handlerMutex   sync.RWMutex
)

// SetCurrentHandler sets the current message handler
func SetCurrentHandler(handler MessageHandler) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	currentHandler = handler
}

// GetCurrentHandler gets the current message handler
func GetCurrentHandler() MessageHandler {
	handlerMutex.RLock()
	defer handlerMutex.RUnlock()
	return currentHandler
}
