package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// JSONRPC constants
const (
	JSONRPCVersion = "2.0"
)

// Transport type
type TransportType string

const (
	// StdioTransport is the stdio transport type
	StdioTransport TransportType = "stdio"
	// HTTPTransport is the HTTP transport type
	HTTPTransport TransportType = "http"
)

// Message types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// EchoClient is a client for the echo server
type EchoClient struct {
	transportType TransportType
	httpClient    *http.Client
	httpURL       string
	stdinWriter   io.Writer
	stdoutReader  io.Reader
	mu            sync.Mutex
}

// NewEchoClient creates a new echo client
func NewEchoClient() *EchoClient {
	return &EchoClient{
		httpClient: &http.Client{},
	}
}

// UseHTTPTransport configures the client to use HTTP transport
func (c *EchoClient) UseHTTPTransport(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transportType = HTTPTransport
	c.httpURL = url
}

// UseStdioTransport configures the client to use stdio transport
func (c *EchoClient) UseStdioTransport() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transportType = StdioTransport
	c.stdinWriter = os.Stdout
	c.stdoutReader = os.Stdin
}

// UseCustomStdioTransport configures the client to use stdio transport with custom IO
func (c *EchoClient) UseCustomStdioTransport(writer io.Writer, reader io.Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transportType = StdioTransport
	c.stdinWriter = writer
	c.stdoutReader = reader
}

// Echo sends an echo request to the server
func (c *EchoClient) Echo(ctx context.Context, text string) (string, error) {
	// Create parameter object
	params := map[string]interface{}{
		"text": text,
	}

	// Marshal parameters to JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal parameters")
	}

	// Create request
	req := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		ID:      uuid.New().String(),
		Method:  "tools/call",
		Params: json.RawMessage(fmt.Sprintf(`{
			"name": "echo",
			"arguments": %s
		}`, paramsJSON)),
	}

	// Send request and get response
	var resp JSONRPCResponse
	if err := c.sendRequest(ctx, req, &resp); err != nil {
		return "", err
	}

	// Check for errors
	if resp.Error != nil {
		return "", fmt.Errorf("server error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
	}

	// Parse result
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", errors.Wrap(err, "failed to parse response")
	}

	if len(result.Content) == 0 {
		return "", errors.New("empty response")
	}

	return result.Content[0].Text, nil
}

// sendRequest sends a request to the server using the configured transport
func (c *EchoClient) sendRequest(ctx context.Context, req JSONRPCRequest, resp *JSONRPCResponse) error {
	c.mu.Lock()
	transportType := c.transportType
	c.mu.Unlock()

	switch transportType {
	case HTTPTransport:
		return c.sendHTTPRequest(ctx, req, resp)
	case StdioTransport:
		return c.sendStdioRequest(ctx, req, resp)
	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}
}

// sendHTTPRequest sends a request using HTTP transport
func (c *EchoClient) sendHTTPRequest(ctx context.Context, req JSONRPCRequest, resp *JSONRPCResponse) error {
	// Marshal request to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.httpURL, bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server returned non-OK status: %s", httpResp.Status)
	}

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	// Parse response JSON
	if err := json.Unmarshal(body, resp); err != nil {
		return errors.Wrap(err, "failed to parse response JSON")
	}

	return nil
}

// sendStdioRequest sends a request using stdio transport
func (c *EchoClient) sendStdioRequest(ctx context.Context, req JSONRPCRequest, resp *JSONRPCResponse) error {
	// Marshal request to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request")
	}

	// Add newline for message framing
	data = append(data, '\n')

	// Write request to stdin
	_, err = c.stdinWriter.Write(data)
	if err != nil {
		return errors.Wrap(err, "failed to write request")
	}

	// Read response from stdout
	// TODO: Implement proper message reading with a buffer and handling of multiple JSON messages
	// For the sake of this example, we assume a single response
	reader := io.LimitReader(c.stdoutReader, 10*1024) // Limit to 10KB
	responseData, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to read response")
	}

	// Parse response JSON
	if err := json.Unmarshal(responseData, resp); err != nil {
		return errors.Wrap(err, "failed to parse response JSON")
	}

	return nil
}
