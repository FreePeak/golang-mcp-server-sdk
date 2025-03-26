package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// StdioTransport implements a transport over standard input/output
type StdioTransport struct {
	reader    *bufio.Reader
	writer    *bufio.Writer
	handler   transport.MessageHandler
	closeCh   chan struct{}
	closeOnce sync.Once
	writeMu   sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader:  bufio.NewReader(os.Stdin),
		writer:  bufio.NewWriter(os.Stdout),
		closeCh: make(chan struct{}),
	}
}

// Start starts the transport
func (t *StdioTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	t.handler = handler

	// Start reading messages
	go t.readMessages(ctx)

	return nil
}

// Send sends a message through the transport
func (t *StdioTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error marshalling message")
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if _, err := t.writer.Write(data); err != nil {
		return errors.Wrap(err, "error writing message")
	}
	if err := t.writer.WriteByte('\n'); err != nil {
		return errors.Wrap(err, "error writing newline")
	}
	if err := t.writer.Flush(); err != nil {
		return errors.Wrap(err, "error flushing writer")
	}

	return nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
	})
	return nil
}

// readMessages reads messages from stdin
func (t *StdioTransport) readMessages(ctx context.Context) {
	for {
		select {
		case <-t.closeCh:
			return
		case <-ctx.Done():
			return
		default:
			line, err := t.reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				}
				return
			}

			var message json.RawMessage
			if err := json.Unmarshal(line, &message); err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshalling message: %v\n", err)
				continue
			}

			var basic struct {
				JSONRPC string      `json:"jsonrpc"`
				ID      interface{} `json:"id,omitempty"`
				Method  string      `json:"method,omitempty"`
			}
			if err := json.Unmarshal(message, &basic); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid JSON-RPC message: %v\n", err)
				continue
			}

			if basic.JSONRPC != shared.JSONRPCVersion {
				fmt.Fprintf(os.Stderr, "Invalid JSON-RPC version\n")
				continue
			}

			// Parse the message based on its type
			var jsonRPCMessage shared.JSONRPCMessage
			if basic.Method != "" {
				if basic.ID != nil {
					// Request
					var request shared.JSONRPCRequest
					if err := json.Unmarshal(message, &request); err != nil {
						fmt.Fprintf(os.Stderr, "Invalid JSON-RPC request: %v\n", err)
						continue
					}
					jsonRPCMessage = request

					// Handle initialize request specially
					if request.Method == "initialize" {
						response := shared.JSONRPCResponse{
							JSONRPC: shared.JSONRPCVersion,
							ID:      request.ID,
							Result: map[string]interface{}{
								"serverInfo": shared.ServerInfo{
									Name:    "golang-mcp-server",
									Version: "1.0.0",
								},
								"capabilities": shared.Capabilities{
									Tools: &shared.ToolsCapability{},
								},
							},
						}

						if err := t.Send(ctx, response); err != nil {
							fmt.Fprintf(os.Stderr, "Error sending initialize response: %v\n", err)
						}
						continue
					}
				} else {
					// Notification
					var notification shared.JSONRPCNotification
					if err := json.Unmarshal(message, &notification); err != nil {
						fmt.Fprintf(os.Stderr, "Invalid JSON-RPC notification: %v\n", err)
						continue
					}
					jsonRPCMessage = notification
				}
			} else {
				// Response
				var response shared.JSONRPCResponse
				if err := json.Unmarshal(message, &response); err != nil {
					fmt.Fprintf(os.Stderr, "Invalid JSON-RPC response: %v\n", err)
					continue
				}
				jsonRPCMessage = response
			}

			// Handle the message
			if t.handler != nil {
				if err := t.handler(ctx, jsonRPCMessage); err != nil {
					fmt.Fprintf(os.Stderr, "Error handling message: %v\n", err)
				}
			}
		}
	}
}

// StdioTransportFactory creates stdio transports
type StdioTransportFactory struct{}

// NewStdioTransportFactory creates a new stdio transport factory
func NewStdioTransportFactory() *StdioTransportFactory {
	return &StdioTransportFactory{}
}

// CreateTransport creates a new stdio transport
func (f *StdioTransportFactory) CreateTransport() (transport.Transport, error) {
	return NewStdioTransport(), nil
}
