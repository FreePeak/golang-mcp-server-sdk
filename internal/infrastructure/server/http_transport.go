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
	clients     map[string]chan shared.JSONRPCMessage // key is sessionID
	clientMutex sync.RWMutex
	closeCh     chan struct{}
	closeOnce   sync.Once
	handler     transport.MessageHandler
	baseURL     string
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(host string, port int) *HTTPTransport {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	t := &HTTPTransport{
		clients: make(map[string]chan shared.JSONRPCMessage),
		closeCh: make(chan struct{}),
		baseURL: baseURL,
	}

	mux := http.NewServeMux()
	// SSE endpoint for establishing connection
	mux.HandleFunc("/mcp/sse", t.handleSSE)
	// Message endpoint for receiving client messages
	mux.HandleFunc("/mcp/message", t.handleRequest)

	addr := fmt.Sprintf("%s:%d", host, port)
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

	fmt.Printf("Starting MCP server at %s\n", t.baseURL)
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// handleRequest handles incoming HTTP requests
func (t *HTTPTransport) handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== Incoming Request ===\n")
	fmt.Printf("Method: %s\n", r.Method)
	fmt.Printf("URL: %s\n", r.URL.String())
	fmt.Printf("RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Printf("Headers: %+v\n", r.Header)

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		fmt.Printf("Handling OPTIONS preflight request\n")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		fmt.Printf("Invalid method: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("sessionid")
	if sessionID == "" {
		fmt.Printf("Missing sessionid parameter\n")
		http.Error(w, "Missing sessionid parameter", http.StatusBadRequest)
		return
	}
	fmt.Printf("SessionID: %s\n", sessionID)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %v\n", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("Request Body: %s\n", string(body))

	var message json.RawMessage
	if err := json.Unmarshal(body, &message); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var basic struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id,omitempty"`
		Method  string      `json:"method,omitempty"`
	}
	if err := json.Unmarshal(message, &basic); err != nil {
		fmt.Printf("Error unmarshaling basic JSON-RPC: %v\n", err)
		http.Error(w, "Invalid JSON-RPC message", http.StatusBadRequest)
		return
	}

	fmt.Printf("JSON-RPC Version: %s\n", basic.JSONRPC)
	fmt.Printf("Method: %s\n", basic.Method)
	fmt.Printf("ID: %v\n", basic.ID)

	if basic.JSONRPC != shared.JSONRPCVersion {
		fmt.Printf("Invalid JSON-RPC version: %s (expected: %s)\n", basic.JSONRPC, shared.JSONRPCVersion)
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
				fmt.Printf("Error unmarshaling JSON-RPC request: %v\n", err)
				http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
				return
			}
			jsonRPCMessage = request
			fmt.Printf("Parsed Request: %+v\n", request)

			// Handle initialize request specially
			if request.Method == "initialize" {
				fmt.Printf("Handling initialize request\n")
				response := shared.JSONRPCResponse{
					JSONRPC: shared.JSONRPCVersion,
					ID:      request.ID,
					Result: map[string]interface{}{
						"serverInfo": shared.ServerInfo{
							Name:    "golang-mcp-server",
							Version: "1.0.0",
							Metadata: map[string]interface{}{
								"transport": "sse",
								"baseUrl":   t.baseURL,
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

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				fmt.Printf("Sent initialize response: %+v\n", response)
				return
			}

			// Handle tools/list request
			if request.Method == "tools/list" {
				fmt.Printf("Handling tools/list request\n")
				response := shared.JSONRPCResponse{
					JSONRPC: shared.JSONRPCVersion,
					ID:      request.ID,
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

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				fmt.Printf("Sent tools/list response: %+v\n", response)
				return
			}
		} else {
			// Notification
			var notification shared.JSONRPCNotification
			if err := json.Unmarshal(message, &notification); err != nil {
				fmt.Printf("Error unmarshaling JSON-RPC notification: %v\n", err)
				http.Error(w, "Invalid JSON-RPC notification", http.StatusBadRequest)
				return
			}
			jsonRPCMessage = notification
			fmt.Printf("Parsed Notification: %+v\n", notification)
		}
	} else {
		// Response
		var response shared.JSONRPCResponse
		if err := json.Unmarshal(message, &response); err != nil {
			fmt.Printf("Error unmarshaling JSON-RPC response: %v\n", err)
			http.Error(w, "Invalid JSON-RPC response", http.StatusBadRequest)
			return
		}
		jsonRPCMessage = response
		fmt.Printf("Parsed Response: %+v\n", response)
	}

	// Queue the message for processing
	fmt.Printf("Processing message: %+v\n", jsonRPCMessage)
	if err := t.processMessage(r.Context(), jsonRPCMessage); err != nil {
		fmt.Printf("Error processing message: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response for requests
	if jsonRPCMessage.IsRequest() {
		fmt.Printf("Sending accepted response for request\n")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case <-t.closeCh:
		// Transport closed
		http.Error(w, "Server shutting down", http.StatusServiceUnavailable)
	}
	fmt.Printf("=== End Request ===\n\n")
}

// handleSSE handles Server-Sent Events connections
func (t *HTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== New SSE Connection ===\n")
	fmt.Printf("Method: %s\n", r.Method)
	fmt.Printf("URL: %s\n", r.URL.String())
	fmt.Printf("RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Printf("Headers: %+v\n", r.Header)

	if r.Method != http.MethodGet {
		fmt.Printf("Invalid method: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Get or generate session ID
	sessionID := r.URL.Query().Get("sessionid")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	fmt.Printf("New SSE connection from: %s with session ID: %s\n", r.RemoteAddr, sessionID)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

	// Create message channel for this client if it doesn't exist
	t.clientMutex.Lock()
	if _, exists := t.clients[sessionID]; !exists {
		t.clients[sessionID] = make(chan shared.JSONRPCMessage, 100)
	}
	messageCh := t.clients[sessionID]
	clientCount := len(t.clients)
	t.clientMutex.Unlock()
	fmt.Printf("Client registered with session ID: %s. Total clients: %d\n", sessionID, clientCount)

	// Clean up when the connection is closed
	defer func() {
		fmt.Printf("SSE connection closing for session ID: %s\n", sessionID)
		t.clientMutex.Lock()
		if ch, exists := t.clients[sessionID]; exists {
			delete(t.clients, sessionID)
			close(ch)
		}
		t.clientMutex.Unlock()
		fmt.Printf("Client unregistered. Total clients: %d\n", len(t.clients))
	}()

	// Set up flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		fmt.Printf("Error: Streaming not supported for session ID: %s\n", sessionID)
		return
	}

	fmt.Printf("\n=== Sending Initial Connection Event ===\n")
	// Send initial connection event
	initEvent := map[string]interface{}{
		"type": "connection",
		"data": map[string]interface{}{
			"sessionId": sessionID,
			"serverInfo": map[string]interface{}{
				"name":    "golang-mcp-server",
				"version": "1.0.0",
				"metadata": map[string]interface{}{
					"transport": "sse",
					"baseUrl":   t.baseURL,
					"endpoints": map[string]string{
						"sse":     "/mcp/sse",
						"message": "/mcp/message",
					},
				},
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
		},
	}

	data, _ := json.Marshal(initEvent)
	fmt.Printf("Connection Event Data: %s\n", string(data))
	fmt.Fprintf(w, "event: connection\ndata: %s\n\n", string(data))
	flusher.Flush()
	fmt.Printf("=== End Initial Connection Event ===\n")

	fmt.Printf("\n=== Starting Message Loop ===\n")
	// Wait for messages
	for {
		select {
		case <-t.closeCh:
			fmt.Printf("Transport closing for session ID: %s\n", sessionID)
			return
		case <-r.Context().Done():
			fmt.Printf("Client context done for session ID: %s\n", sessionID)
			return
		case msg, ok := <-messageCh:
			if !ok {
				fmt.Printf("Message channel closed for session ID: %s\n", sessionID)
				return
			}

			var eventType string
			var eventData interface{}

			switch m := msg.(type) {
			case shared.JSONRPCNotification:
				eventType = m.Method
				eventData = m.Params
				fmt.Printf("Sending notification event: %s\n", m.Method)
			case shared.JSONRPCRequest:
				eventType = m.Method
				eventData = map[string]interface{}{
					"id":     m.ID,
					"method": m.Method,
					"params": m.Params,
				}
				fmt.Printf("Sending request event: %s\n", m.Method)
			case shared.JSONRPCResponse:
				eventType = "response"
				eventData = m
				fmt.Printf("Sending response event for ID: %v\n", m.ID)
			default:
				eventType = "message"
				eventData = m
				fmt.Printf("Sending unknown event type\n")
			}

			data, err := json.Marshal(eventData)
			if err != nil {
				fmt.Printf("Error marshaling message for session ID %s: %v\n", sessionID, err)
				continue
			}

			fmt.Printf("Event Data: %s\n", string(data))
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(data))
			flusher.Flush()
		}
	}
}

// Send sends a message to all connected clients
func (t *HTTPTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	fmt.Printf("\n=== Sending Message ===\n")
	fmt.Printf("Message: %+v\n", message)

	// Marshal message to JSON - this is just for error checking
	if _, err := json.Marshal(message); err != nil {
		fmt.Printf("Error marshaling message: %v\n", err)
		return errors.Wrap(err, "error marshalling message")
	}

	t.clientMutex.RLock()
	defer t.clientMutex.RUnlock()

	for sessionID, ch := range t.clients {
		select {
		case ch <- message:
			fmt.Printf("Message sent to session ID: %s\n", sessionID)
		default:
			fmt.Printf("Channel full for session ID %s, skipping message\n", sessionID)
		}
	}
	fmt.Printf("=== End Send ===\n\n")
	return nil
}

// processMessage processes an incoming message
func (t *HTTPTransport) processMessage(ctx context.Context, message shared.JSONRPCMessage) error {
	select {
	case <-t.closeCh:
		return nil
	default:
		if t.handler != nil {
			if err := t.handler(ctx, message); err != nil {
				fmt.Printf("Error handling message: %v\n", err)
				return err
			}
		}
		return nil
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
				for sessionID, ch := range t.clients {
					select {
					case ch <- heartbeat:
						// Message sent successfully
					default:
						fmt.Printf("Session ID %s heartbeat channel full, skipping\n", sessionID)
					}
				}
				t.clientMutex.RUnlock()
			}
		}
	}
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

// HTTPTransportFactory creates HTTP transports
type HTTPTransportFactory struct {
	host string
	port int
}

// NewHTTPTransportFactory creates a new HTTP transport factory
func NewHTTPTransportFactory(host string, port int) *HTTPTransportFactory {
	return &HTTPTransportFactory{
		host: host,
		port: port,
	}
}

// CreateTransport creates a new HTTP transport
func (f *HTTPTransportFactory) CreateTransport() (transport.Transport, error) {
	return NewHTTPTransport(f.host, f.port), nil
}
