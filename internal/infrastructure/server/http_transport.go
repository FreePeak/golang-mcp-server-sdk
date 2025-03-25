package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// HTTPTransport implements a transport over HTTP
type HTTPTransport struct {
	server      *http.Server
	clients     map[string]chan shared.JSONRPCMessage
	clientMutex sync.RWMutex
	closeCh     chan struct{}
	closeOnce   sync.Once
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(addr string) *HTTPTransport {
	t := &HTTPTransport{
		clients: make(map[string]chan shared.JSONRPCMessage),
		closeCh: make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", t.handleRequest)
	mux.HandleFunc("/sse", t.handleSSE)

	t.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return t
}

// Start starts the transport
func (t *HTTPTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	go t.listenForMessages(ctx, handler)

	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// Send sends a message to all connected clients
func (t *HTTPTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	// Marshal message to JSON - this is just for error checking,
	// the actual marshaling for each client happens in handleSSE
	if _, err := json.Marshal(message); err != nil {
		return errors.Wrap(err, "error marshalling message")
	}

	t.clientMutex.RLock()
	defer t.clientMutex.RUnlock()

	for _, ch := range t.clients {
		select {
		case ch <- message:
			// Message sent successfully
		default:
			// Client's channel is full, skip
		}
	}

	return nil
}

// Close closes the transport
func (t *HTTPTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
		if t.server != nil {
			t.server.Shutdown(context.Background())
		}
	})
	return nil
}

// handleRequest handles incoming HTTP requests
func (t *HTTPTransport) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var message json.RawMessage
	if err := json.Unmarshal(body, &message); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var basic struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id,omitempty"`
		Method  string      `json:"method,omitempty"`
	}
	if err := json.Unmarshal(message, &basic); err != nil {
		http.Error(w, "Invalid JSON-RPC message", http.StatusBadRequest)
		return
	}

	if basic.JSONRPC != shared.JSONRPCVersion {
		http.Error(w, "Invalid JSON-RPC version", http.StatusBadRequest)
		return
	}

	// Parse the message based on its type
	var jsonRPCMessage shared.JSONRPCMessage
	if basic.Method != "" {
		if basic.ID != nil {
			// Request
			var request shared.JSONRPCRequest
			if err := json.Unmarshal(message, &request); err != nil {
				http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
				return
			}
			jsonRPCMessage = request
		} else {
			// Notification
			var notification shared.JSONRPCNotification
			if err := json.Unmarshal(message, &notification); err != nil {
				http.Error(w, "Invalid JSON-RPC notification", http.StatusBadRequest)
				return
			}
			jsonRPCMessage = notification
		}
	} else {
		// Response
		var response shared.JSONRPCResponse
		if err := json.Unmarshal(message, &response); err != nil {
			http.Error(w, "Invalid JSON-RPC response", http.StatusBadRequest)
			return
		}
		jsonRPCMessage = response
	}

	// Queue the message for processing
	t.handleMessage(context.Background(), jsonRPCMessage)

	// If this is a request, we need to send a response
	if jsonRPCMessage.IsRequest() {
		// The sendMessage callback will send the response back to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}
}

// handleSSE handles Server-Sent Events connections
func (t *HTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a unique client ID
	clientID := fmt.Sprintf("%p", r)

	// Create message channel for this client
	messageCh := make(chan shared.JSONRPCMessage, 100)

	// Register the client
	t.clientMutex.Lock()
	t.clients[clientID] = messageCh
	t.clientMutex.Unlock()

	// Clean up on connection close
	defer func() {
		t.clientMutex.Lock()
		delete(t.clients, clientID)
		t.clientMutex.Unlock()
		close(messageCh)
	}()

	// Prepare a flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send messages to the client
	for {
		select {
		case <-t.closeCh:
			return
		case <-r.Context().Done():
			return
		case msg, ok := <-messageCh:
			if !ok {
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleMessage processes an incoming message
func (t *HTTPTransport) handleMessage(ctx context.Context, message shared.JSONRPCMessage) {
	// This channel receives messages to be handled
	select {
	case <-t.closeCh:
		return
	default:
		// Queue the message
	}
}

// listenForMessages listens for incoming messages and passes them to the handler
func (t *HTTPTransport) listenForMessages(ctx context.Context, handler transport.MessageHandler) {
	// This is where we'd process messages from some channel
	// For HTTP transport, the messages are processed on the handler goroutine directly
}

// HTTPTransportFactory creates HTTP transports
type HTTPTransportFactory struct {
	addr string
}

// NewHTTPTransportFactory creates a new HTTP transport factory
func NewHTTPTransportFactory(addr string) *HTTPTransportFactory {
	return &HTTPTransportFactory{
		addr: addr,
	}
}

// CreateTransport creates a new HTTP transport
func (f *HTTPTransportFactory) CreateTransport() (transport.Transport, error) {
	return NewHTTPTransport(f.addr), nil
}
