package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// TestableStdioTransport is a specialized version for testing
type TestableStdioTransport struct {
	inputBuf  *bytes.Buffer
	outputBuf *bytes.Buffer
	closeCh   chan struct{}
	closeOnce sync.Once
	handler   transport.MessageHandler
}

func NewTestableStdioTransport(input, output *bytes.Buffer) *TestableStdioTransport {
	return &TestableStdioTransport{
		inputBuf:  input,
		outputBuf: output,
		closeCh:   make(chan struct{}),
	}
}

func (t *TestableStdioTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	t.handler = handler
	go t.processMessages(ctx)
	return nil
}

func (t *TestableStdioTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Add newline for message framing
	data = append(data, '\n')

	_, err = t.outputBuf.Write(data)
	return err
}

func (t *TestableStdioTransport) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
	})
	return nil
}

func (t *TestableStdioTransport) processMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.closeCh:
			return
		default:
			line, err := t.readLine()
			if err != nil {
				if err == io.EOF {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return
			}

			// Process the message
			var message json.RawMessage
			if err := json.Unmarshal([]byte(line), &message); err != nil {
				continue
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

			if t.handler != nil {
				if err := t.handler(ctx, jsonRPCMessage); err != nil {
					// Just log the error
				}
			}
		}
	}
}

func (t *TestableStdioTransport) readLine() (string, error) {
	var line []byte
	var b byte
	var err error

	for {
		b, err = t.inputBuf.ReadByte()
		if err != nil {
			return "", err
		}

		line = append(line, b)
		if b == '\n' {
			return string(line), nil
		}
	}
}

func (t *TestableStdioTransport) WriteMessage(message string) {
	t.inputBuf.WriteString(message)
}

func TestStdioTransport(t *testing.T) {
	// Set up input and output buffers
	inputBuf := &bytes.Buffer{}
	outputBuf := &bytes.Buffer{}

	// Create testable transport
	tp := NewTestableStdioTransport(inputBuf, outputBuf)

	// Create context and message channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messagesCh := make(chan shared.JSONRPCMessage, 10)

	// Define message handler
	handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
		messagesCh <- message

		// Echo the request as a response
		if message.IsRequest() {
			req := message.(shared.JSONRPCRequest)
			response := shared.JSONRPCResponse{
				JSONRPC: shared.JSONRPCVersion,
				ID:      req.ID,
				Result:  req.Params,
			}
			return tp.Send(ctx, response)
		}

		return nil
	}

	// Start the transport
	if err := tp.Start(ctx, handler); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}

	// Write a request to the input buffer
	request := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      "test-id",
		Method:  "test-method",
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	requestJSON = append(requestJSON, '\n')

	_, err = inputBuf.Write(requestJSON)
	if err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Wait for the message to be received
	var receivedMessage shared.JSONRPCMessage
	select {
	case receivedMessage = <-messagesCh:
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Check that we received the correct message
	receivedRequest, ok := receivedMessage.(shared.JSONRPCRequest)
	if !ok {
		t.Fatalf("Expected JSONRPCRequest, got %T", receivedMessage)
	}

	if receivedRequest.JSONRPC != shared.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", shared.JSONRPCVersion, receivedRequest.JSONRPC)
	}

	if receivedRequest.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%v'", receivedRequest.ID)
	}

	if receivedRequest.Method != "test-method" {
		t.Errorf("Expected method 'test-method', got '%s'", receivedRequest.Method)
	}

	// Wait a bit for the response to be written
	time.Sleep(50 * time.Millisecond)

	// Read the response from the output buffer
	responseJSON, err := io.ReadAll(outputBuf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Parse the response
	var response shared.JSONRPCResponse
	err = json.Unmarshal(responseJSON, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check the response
	if response.JSONRPC != shared.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", shared.JSONRPCVersion, response.JSONRPC)
	}

	if response.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%v'", response.ID)
	}

	if response.Error != nil {
		t.Errorf("Unexpected error: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", response.Result)
	}

	if result["param1"] != "value1" {
		t.Errorf("Expected param1 'value1', got '%v'", result["param1"])
	}

	if int(result["param2"].(float64)) != 42 {
		t.Errorf("Expected param2 42, got %v", result["param2"])
	}
}

func TestStdioTransportFactory(t *testing.T) {
	factory := NewStdioTransportFactory()

	tp, err := factory.CreateTransport()
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	if tp == nil {
		t.Fatal("Expected non-nil transport")
	}

	stdioTp, ok := tp.(*StdioTransport)
	if !ok {
		t.Fatalf("Expected StdioTransport, got %T", tp)
	}

	if stdioTp.reader == nil {
		t.Error("Expected non-nil reader")
	}

	if stdioTp.writer == nil {
		t.Error("Expected non-nil writer")
	}
}

func TestStdioTransportInvalidJSON(t *testing.T) {
	// Set up input and output buffers
	inputBuf := &bytes.Buffer{}
	outputBuf := &bytes.Buffer{}

	// Create transport with custom IO
	tp := NewStdioTransportWithIO(inputBuf, outputBuf)

	// Create context and message channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handlerCalled := false
	handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
		handlerCalled = true
		return nil
	}

	// Start the transport
	if err := tp.Start(ctx, handler); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}

	// Write invalid JSON to the input buffer
	_, err := inputBuf.WriteString("invalid json\n")
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Wait a bit for the message to be processed
	time.Sleep(10 * time.Millisecond)

	// Check that the handler was not called
	if handlerCalled {
		t.Error("Expected handler not to be called for invalid JSON")
	}
}

func TestStdioTransportClose(t *testing.T) {
	// Create transport
	tp := NewStdioTransport()

	// Create context and message channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the transport with a handler that blocks until canceled
	started := make(chan struct{})
	handlerCalled := make(chan struct{})

	handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
		close(handlerCalled)

		// Block until canceled
		<-ctx.Done()
		return nil
	}

	// Start transport in a goroutine
	go func() {
		if err := tp.Start(ctx, handler); err != nil {
			t.Errorf("Failed to start transport: %v", err)
		}
		close(started)
	}()

	// Wait a bit for the transport to start
	time.Sleep(10 * time.Millisecond)

	// Close the transport
	if err := tp.Close(); err != nil {
		t.Fatalf("Failed to close transport: %v", err)
	}

	// Wait for the transport to finish
	select {
	case <-started:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for transport to close")
	}
}

func TestStdioTransportWithIO(t *testing.T) {
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)

	transport := NewStdioTransportWithIO(input, output)

	// Create a test request
	request := shared.JSONRPCRequest{
		JSONRPC: shared.JSONRPCVersion,
		ID:      1,
		Method:  "test",
		Params:  map[string]string{"key": "value"},
	}

	// Marshal the request to JSON and add newline
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	requestJSON = append(requestJSON, '\n')

	// Write to input buffer
	_, err = input.Write(requestJSON)
	if err != nil {
		t.Fatalf("Failed to write to input buffer: %v", err)
	}

	// Message handler that will receive the request
	var receivedMessage shared.JSONRPCMessage
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
		receivedMessage = message
		wg.Done()
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the transport
	if err := transport.Start(ctx, handler); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}

	// Wait for message to be processed or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Message was processed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message to be processed")
	}

	// Verify the message was received correctly
	receivedRequest, ok := receivedMessage.(shared.JSONRPCRequest)
	if !ok {
		t.Fatalf("Expected JSONRPCRequest, got %T", receivedMessage)
	}

	if receivedRequest.JSONRPC != shared.JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", shared.JSONRPCVersion, receivedRequest.JSONRPC)
	}

	if receivedRequest.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", receivedRequest.ID)
	}

	if receivedRequest.Method != "test" {
		t.Errorf("Expected Method 'test', got %s", receivedRequest.Method)
	}

	// Test sending a response
	response := shared.JSONRPCResponse{
		JSONRPC: shared.JSONRPCVersion,
		ID:      1,
		Result:  map[string]string{"status": "success"},
	}

	// Send the response
	if err := transport.Send(ctx, response); err != nil {
		t.Fatalf("Failed to send response: %v", err)
	}

	// Verify output
	outputStr := output.String()

	// Check that output contains expected JSON properties
	if !strings.Contains(outputStr, "\"jsonrpc\":\"2.0\"") {
		t.Errorf("Output missing jsonrpc field: %s", outputStr)
	}

	if !strings.Contains(outputStr, "\"id\":1") {
		t.Errorf("Output missing id field: %s", outputStr)
	}

	if !strings.Contains(outputStr, "\"result\"") {
		t.Errorf("Output missing result field: %s", outputStr)
	}

	// Close the transport
	if err := transport.Close(); err != nil {
		t.Fatalf("Failed to close transport: %v", err)
	}
}

func TestStdioTransportStartMultipleTimes(t *testing.T) {
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)

	transport := NewStdioTransportWithIO(input, output)

	handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the transport first time
	if err := transport.Start(ctx, handler); err != nil {
		t.Fatalf("Failed to start transport first time: %v", err)
	}

	// Try to start again, should fail
	err := transport.Start(ctx, handler)
	if err == nil {
		t.Fatal("Expected error when starting transport twice, got nil")
	}

	// Close the transport
	if err := transport.Close(); err != nil {
		t.Fatalf("Failed to close transport: %v", err)
	}
}

func TestStdioTransportEOF(t *testing.T) {
	// Create a reader that's empty to simulate EOF
	input := strings.NewReader("")
	output := bytes.NewBuffer(nil)

	transport := NewStdioTransportWithIO(input, output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to signal when the handler is called
	handlerCalled := make(chan struct{})

	// Start the transport in a goroutine
	startDone := make(chan struct{})
	go func() {
		// Use a simple handler that just closes the handlerCalled channel
		handler := func(ctx context.Context, message shared.JSONRPCMessage) error {
			close(handlerCalled)
			return nil
		}

		err := transport.Start(ctx, handler)
		if err != nil {
			t.Errorf("Failed to start transport: %v", err)
		}
		close(startDone)
	}()

	// Wait for the transport to start
	select {
	case <-startDone:
		// Transport started
	case <-time.After(100 * time.Millisecond):
		// Give it some time to start
	}

	// Give the transport time to process the empty reader and detect EOF
	time.Sleep(100 * time.Millisecond)

	// Now close the transport explicitly
	if err := transport.Close(); err != nil {
		t.Fatalf("Failed to close transport: %v", err)
	}

	// Wait for context to be canceled or timeout
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be canceled")
	case <-time.After(100 * time.Millisecond):
		// Success - transport closed without hanging
	}
}
