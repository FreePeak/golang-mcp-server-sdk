package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

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
	handler     transport.MessageHandler
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
	t.handler = handler

	// Start heartbeat goroutine
	go t.sendHeartbeats(ctx)

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
	fmt.Printf("New SSE connection from: %s\n", r.RemoteAddr)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a unique client ID
	clientID := fmt.Sprintf("%p", r)
	fmt.Printf("Assigned client ID: %s\n", clientID)

	// Create message channel for this client
	messageCh := make(chan shared.JSONRPCMessage, 100)

	// Register the client
	t.clientMutex.Lock()
	t.clients[clientID] = messageCh
	clientCount := len(t.clients)
	t.clientMutex.Unlock()
	fmt.Printf("Client registered. Total clients: %d\n", clientCount)

	// Clean up on connection close
	defer func() {
		fmt.Printf("SSE connection closing for client: %s\n", clientID)
		t.clientMutex.Lock()
		delete(t.clients, clientID)
		newClientCount := len(t.clients)
		t.clientMutex.Unlock()
		fmt.Printf("Client unregistered. Total clients: %d\n", newClientCount)
		close(messageCh)
	}()

	// Prepare a flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		fmt.Printf("Error: Streaming not supported for client: %s\n", clientID)
		return
	}

	// Initial keepalive to ensure connection is established
	fmt.Fprintf(w, "data: %s\n\n", `{"type":"connect","status":"ok"}`)
	flusher.Flush()
	fmt.Printf("Sent initial keepalive message to client: %s\n", clientID)

	// Send messages to the client
	for {
		select {
		case <-t.closeCh:
			fmt.Printf("Transport closing for client: %s\n", clientID)
			return
		case <-r.Context().Done():
			fmt.Printf("Client context done for client: %s\n", clientID)
			return
		case msg, ok := <-messageCh:
			if !ok {
				fmt.Printf("Message channel closed for client: %s\n", clientID)
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling message for client %s: %v\n", clientID, err)
				continue
			}

			fmt.Printf("Sending message to client %s: %s\n", clientID, string(data))
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
		// Pass the message to the handler
		if t.handler != nil {
			if err := t.handler(ctx, message); err != nil {
				fmt.Printf("Error handling message: %v\n", err)
			}
		}
	}
}

// sendHeartbeats periodically sends heartbeat messages to all connected clients
func (t *HTTPTransport) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.closeCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.clientMutex.RLock()
			clientCount := len(t.clients)
			t.clientMutex.RUnlock()

			if clientCount > 0 {
				fmt.Printf("Sending heartbeat to %d clients\n", clientCount)
				heartbeat := shared.JSONRPCNotification{
					JSONRPC: shared.JSONRPCVersion,
					Method:  "system/heartbeat",
					Params: map[string]interface{}{
						"timestamp": time.Now().Unix(),
					},
				}

				// Don't use t.Send to avoid recursive lock
				t.clientMutex.RLock()
				for id, ch := range t.clients {
					select {
					case ch <- heartbeat:
						// Message sent successfully
					default:
						fmt.Printf("Client %s heartbeat channel full, skipping\n", id)
					}
				}
				t.clientMutex.RUnlock()
			}
		}
	}
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
