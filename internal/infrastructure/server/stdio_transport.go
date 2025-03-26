package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// StdioTransport implements a transport over standard input/output
type StdioTransport struct {
	reader     *bufio.Reader
	writer     *bufio.Writer
	closeCh    chan struct{}
	closeOnce  sync.Once
	isStarted  bool
	startMutex sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader:  bufio.NewReader(os.Stdin),
		writer:  bufio.NewWriter(os.Stdout),
		closeCh: make(chan struct{}),
	}
}

// NewStdioTransportWithIO creates a new stdio transport with custom reader and writer
func NewStdioTransportWithIO(reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader:  bufio.NewReader(reader),
		writer:  bufio.NewWriter(writer),
		closeCh: make(chan struct{}),
	}
}

// Start starts the transport
func (t *StdioTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	t.startMutex.Lock()
	if t.isStarted {
		t.startMutex.Unlock()
		return errors.New("transport already started")
	}
	t.isStarted = true
	t.startMutex.Unlock()

	go t.processMessages(ctx, handler)
	return nil
}

// Send sends a message
func (t *StdioTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error marshalling message")
	}

	// Add newline for message framing
	data = append(data, '\n')

	_, err = t.writer.Write(data)
	if err != nil {
		return errors.Wrap(err, "error writing message")
	}

	return t.writer.Flush()
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
	})
	return nil
}

// processMessages reads and processes incoming messages
func (t *StdioTransport) processMessages(ctx context.Context, handler transport.MessageHandler) {
	// Buffer for incomplete lines
	var buffer []byte

	// goroutine to handle context cancelation
	go func() {
		select {
		case <-ctx.Done():
			t.Close()
		case <-t.closeCh:
			// already closing
		}
	}()

	// Main read loop
	for {
		select {
		case <-t.closeCh:
			return
		default:
			// Continue with read
		}

		// Single byte buffer read to make it interruptible
		buf := make([]byte, 1024)
		n, err := t.reader.Read(buf)

		if n > 0 {
			// Add read bytes to buffer
			buffer = append(buffer, buf[:n]...)

			// Process complete lines
			for {
				// Look for newline
				i := bytes.IndexByte(buffer, '\n')
				if i < 0 {
					break // No complete line yet
				}

				// Extract line
				line := buffer[:i+1]
				buffer = buffer[i+1:]

				// Process the message
				var message json.RawMessage
				if err := json.Unmarshal(line, &message); err != nil {
					continue // Invalid JSON, skip
				}

				// Determine message type
				var basic struct {
					JSONRPC string      `json:"jsonrpc"`
					ID      interface{} `json:"id,omitempty"`
					Method  string      `json:"method,omitempty"`
				}
				if err := json.Unmarshal(message, &basic); err != nil {
					continue
				}

				var jsonRPCMessage shared.JSONRPCMessage
				if basic.Method != "" {
					if basic.ID != nil {
						// Request
						var request shared.JSONRPCRequest
						if err := json.Unmarshal(message, &request); err != nil {
							continue
						}
						jsonRPCMessage = request
					} else {
						// Notification
						var notification shared.JSONRPCNotification
						if err := json.Unmarshal(message, &notification); err != nil {
							continue
						}
						jsonRPCMessage = notification
					}
				} else {
					// Response
					var response shared.JSONRPCResponse
					if err := json.Unmarshal(message, &response); err != nil {
						continue
					}
					jsonRPCMessage = response
				}

				// Handle the message
				if handler != nil {
					if err := handler(ctx, jsonRPCMessage); err != nil {
						fmt.Fprintln(os.Stderr, "Error handling message:", err)
					}
				}
			}
		}

		// Handle errors
		if err != nil {
			if err == io.EOF {
				// EOF means we're done
				t.Close()
				return
			} else if errors.Is(err, context.Canceled) {
				// Context canceled
				return
			} else {
				// Other errors - might be temporary
				time.Sleep(10 * time.Millisecond)
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
