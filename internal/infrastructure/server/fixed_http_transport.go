package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// FixedHTTPTransport implements a transport over HTTP with proper CORS handling
type FixedHTTPTransport struct {
	server      *http.Server
	clients     map[string]chan shared.JSONRPCMessage
	sessions    map[string]string // Maps request IDs to session IDs
	clientMutex sync.RWMutex
	closeCh     chan struct{}
	closeOnce   sync.Once
}

// NewFixedHTTPTransport creates a new HTTP transport with proper CORS handling
func NewFixedHTTPTransport(addr string) *FixedHTTPTransport {
	t := &FixedHTTPTransport{
		clients:  make(map[string]chan shared.JSONRPCMessage),
		sessions: make(map[string]string),
		closeCh:  make(chan struct{}),
	}

	// Create a custom handler that correctly handles OPTIONS requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// For SSE endpoint
		if r.URL.Path == "/sse" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			if r.Method == http.MethodGet {
				t.handleSSE(w, r)
				return
			}

			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For other endpoints (root path, etc.)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			t.handleRequest(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	t.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return t
}

// Start starts the transport
func (t *FixedHTTPTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	// Store the handler for use with incoming HTTP requests
	transport.SetCurrentHandler(handler)

	// Start the HTTP server
	go func() {
		fmt.Println("Starting Fixed HTTP server on", t.server.Addr)
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Start a goroutine to handle shutdown
	go func() {
		<-ctx.Done()
		t.Close()
	}()

	return nil
}

// Send sends a message to all connected clients
func (t *FixedHTTPTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error marshalling message")
	}

	// Check if this is a response to an initialize request
	var isInitializeResponse bool
	var requestID string
	if resp, ok := message.(shared.JSONRPCResponse); ok {
		requestID = resp.ID

		// Find which session this belongs to
		t.clientMutex.RLock()
		sessionID, exists := t.sessions[requestID]
		t.clientMutex.RUnlock()

		if exists {
			// Log that we're sending a message to a specific session
			fmt.Printf("Sending response ID %s to session %s: %s\n",
				requestID, sessionID, string(data))

			// Check if this is an initialize response - contains capabilities
			if resp.Result != nil {
				if initResult, ok := resp.Result.(shared.InitializeResult); ok {
					isInitializeResponse = true
					fmt.Printf("This is an initialize response with capabilities: %+v\n",
						initResult.Capabilities)
				}
			}
		} else {
			fmt.Printf("Sending response to ID %s (no session found): %s\n",
				requestID, string(data))
		}
	} else {
		fmt.Printf("Sending message: %s\n", string(data))
	}

	// Send to specific request ID first if it's a response
	if resp, ok := message.(shared.JSONRPCResponse); ok {
		t.clientMutex.RLock()
		ch, exists := t.clients[resp.ID]
		t.clientMutex.RUnlock()

		if exists {
			select {
			case ch <- message:
				fmt.Printf("Message sent directly to request channel: %s\n", resp.ID)
				// For non-initialize responses, we're done
				if !isInitializeResponse {
					return nil
				}
			default:
				fmt.Printf("Warning: Client channel for request %s is full\n", resp.ID)
			}
		}
	}

	// For initialize responses and other messages, also broadcast to SSE clients
	// so Cursor and other clients get notified about capabilities
	if isInitializeResponse {
		fmt.Println("Broadcasting initialize response to all SSE clients for tool discovery")
	}

	// Send to all SSE clients that match the session or to all clients for broadcasts
	t.clientMutex.RLock()
	defer t.clientMutex.RUnlock()

	sent := false
	for clientID, ch := range t.clients {
		// Skip direct request channels (they start with different prefix)
		if !strings.HasPrefix(clientID, "sse-") {
			continue
		}

		// For responses to specific requests, target only clients in the same session
		if resp, ok := message.(shared.JSONRPCResponse); ok && !isInitializeResponse {
			sessionID, exists := t.sessions[resp.ID]
			if !exists || !strings.Contains(clientID, sessionID) {
				continue
			}
		}

		select {
		case ch <- message:
			fmt.Printf("Message sent to SSE client: %s\n", clientID)
			sent = true
		default:
			fmt.Printf("Warning: SSE channel is full, skipping: %s\n", clientID)
		}
	}

	// For initialize responses, we expect at least one client
	if isInitializeResponse && !sent && len(t.clients) > 0 {
		// Don't error, just log a warning - initialization responses
		// don't need to be delivered to SSE clients immediately
		fmt.Println("Warning: No SSE clients connected yet to receive initialize response")
	}

	return nil
}

// Close closes the transport
func (t *FixedHTTPTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
		if t.server != nil {
			t.server.Shutdown(context.Background())
		}
	})
	return nil
}

// handleRequest handles HTTP POST requests
func (t *FixedHTTPTransport) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from request
	sessionID := ""
	cookie, err := r.Cookie("mcp-session")
	if err == nil {
		sessionID = cookie.Value
	}

	// If no session cookie exists, create one
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%s-%d", r.RemoteAddr, time.Now().UnixNano())
		http.SetCookie(w, &http.Cookie{
			Name:     "mcp-session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   r.TLS != nil,
			MaxAge:   3600, // 1 hour
		})
	}

	// Read and parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received request from session %s: %s\n", sessionID, string(body))

	// Parse the JSON-RPC request
	var request shared.JSONRPCRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	// Special handling for initialize
	isInitialize := request.Method == shared.MethodInitialize

	// For initialize requests, track the session in both directions
	if isInitialize {
		fmt.Printf("Initialize request from session: %s (ID: %s)\n", sessionID, request.ID)
	}

	// Associate this request ID with the session
	t.clientMutex.Lock()
	t.sessions[request.ID] = sessionID

	// Create a response channel for this request
	responseID := request.ID
	interceptCh := make(chan shared.JSONRPCMessage, 1)
	t.clients[responseID] = interceptCh
	t.clientMutex.Unlock()

	// Clean up when done
	defer func() {
		t.clientMutex.Lock()
		delete(t.clients, responseID)
		t.clientMutex.Unlock()
	}()

	// Create a context with timeout
	timeout := 10 * time.Second // increase timeout for complex operations
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Set up interceptor to capture the response
	var (
		response shared.JSONRPCMessage
		received = make(chan struct{})
	)

	go func() {
		for {
			select {
			case msg := <-interceptCh:
				// Check if this is a response to our request
				if resp, ok := msg.(shared.JSONRPCResponse); ok && resp.ID == responseID {
					response = resp
					close(received)
					return
				}
			case <-ctx.Done():
				// Context done, exit
				return
			}
		}
	}()

	// Process the request through the handler
	currentHandler := transport.GetCurrentHandler()
	if currentHandler == nil {
		http.Error(w, "Server not initialized", http.StatusInternalServerError)
		return
	}

	// Send the request to be processed
	if err := currentHandler(ctx, request); err != nil {
		fmt.Printf("Error handling request: %v\n", err)

		// Create an error response
		errorResp := shared.JSONRPCResponse{
			JSONRPC: shared.JSONRPCVersion,
			ID:      request.ID,
			Error: &shared.JSONRPCError{
				Code:    int(shared.InternalError),
				Message: err.Error(),
			},
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(errorResp); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Wait for the response or timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-received:
		// Response captured by interceptor
	case <-timer.C:
		// Create a default response on timeout
		fmt.Printf("Timeout waiting for response to request ID: %s\n", responseID)
		response = shared.JSONRPCResponse{
			JSONRPC: shared.JSONRPCVersion,
			ID:      request.ID,
			Result:  nil, // Empty result
		}
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// handleSSE handles Server-Sent Events connections
func (t *FixedHTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from request
	sessionID := ""
	cookie, err := r.Cookie("mcp-session")
	if err == nil {
		sessionID = cookie.Value
	}

	// If no session cookie exists, create one based on remote address
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%s-%d", r.RemoteAddr, time.Now().UnixNano())
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a unique client ID for this SSE connection
	clientID := fmt.Sprintf("sse-%s", sessionID)

	// Create a channel for this client
	clientCh := make(chan shared.JSONRPCMessage, 10)

	// Add the client to the client list
	t.clientMutex.Lock()
	t.clients[clientID] = clientCh
	t.clientMutex.Unlock()

	fmt.Printf("SSE client connected: %s (session: %s)\n", clientID, sessionID)

	// Clean up when the connection is closed
	defer func() {
		t.clientMutex.Lock()
		delete(t.clients, clientID)
		// Note: We don't delete from sessions map to maintain session state
		t.clientMutex.Unlock()
		fmt.Printf("SSE client disconnected: %s\n", clientID)
	}()

	// Set up flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Write initial SSE response to establish connection
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	// Create a context that's cancelled when the client disconnects
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Set up a ticker for keepalive messages
	ticker := time.NewTicker(10 * time.Second) // reduced to 10 seconds for better reactivity
	defer ticker.Stop()

	// Message loop
	for {
		select {
		case msg := <-clientCh:
			// Marshal the message to JSON
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling message: %v\n", err)
				continue
			}

			// Write the SSE message in the correct format
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			fmt.Printf("Sent message to SSE client %s\n", clientID)
		case <-ticker.C:
			// Send a keepalive comment
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-t.closeCh:
			// Transport closed
			return
		case <-ctx.Done():
			// Client disconnected
			return
		}
	}
}
