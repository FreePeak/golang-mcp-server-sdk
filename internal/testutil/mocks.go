package testutil

import (
	"context"
	"sync"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// MockTransport implements transport.Transport for testing
type MockTransport struct {
	StartFunc  func(ctx context.Context, handler transport.MessageHandler) error
	SendFunc   func(ctx context.Context, message shared.JSONRPCMessage) error
	CloseFunc  func() error
	Messages   []shared.JSONRPCMessage
	mu         sync.Mutex
	Handler    transport.MessageHandler
	messagesCh chan shared.JSONRPCMessage
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		Messages:   make([]shared.JSONRPCMessage, 0),
		messagesCh: make(chan shared.JSONRPCMessage, 100),
	}
}

// Start implements Transport.Start
func (m *MockTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	m.Handler = handler
	if m.StartFunc != nil {
		return m.StartFunc(ctx, handler)
	}
	return nil
}

// Send implements Transport.Send
func (m *MockTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	m.mu.Lock()
	m.Messages = append(m.Messages, message)
	m.mu.Unlock()

	select {
	case m.messagesCh <- message:
	default:
	}

	if m.SendFunc != nil {
		return m.SendFunc(ctx, message)
	}
	return nil
}

// Close implements Transport.Close
func (m *MockTransport) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// SimulateIncomingMessage simulates an incoming message
func (m *MockTransport) SimulateIncomingMessage(ctx context.Context, message shared.JSONRPCMessage) error {
	if m.Handler != nil {
		return m.Handler(ctx, message)
	}
	return nil
}

// GetMessages returns all sent messages
func (m *MockTransport) GetMessages() []shared.JSONRPCMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Messages
}

// WaitForMessage waits for a message to be sent
func (m *MockTransport) WaitForMessage(ctx context.Context) (shared.JSONRPCMessage, error) {
	select {
	case msg := <-m.messagesCh:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// MockResourceHandler implements handler.ResourceHandler for testing
type MockResourceHandler struct {
	ListResourcesFunc func(ctx context.Context) ([]shared.Resource, error)
	GetResourceFunc   func(ctx context.Context, uri string) ([]shared.Content, error)
}

// ListResources implements ResourceHandler.ListResources
func (m *MockResourceHandler) ListResources(ctx context.Context) ([]shared.Resource, error) {
	if m.ListResourcesFunc != nil {
		return m.ListResourcesFunc(ctx)
	}
	return []shared.Resource{}, nil
}

// GetResource implements ResourceHandler.GetResource
func (m *MockResourceHandler) GetResource(ctx context.Context, uri string) ([]shared.Content, error) {
	if m.GetResourceFunc != nil {
		return m.GetResourceFunc(ctx, uri)
	}
	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: "Test content",
		},
	}, nil
}

// MockToolHandler implements handler.ToolHandler for testing
type MockToolHandler struct {
	ListToolsFunc func(ctx context.Context) ([]shared.Tool, error)
	CallToolFunc  func(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error)
}

// ListTools implements ToolHandler.ListTools
func (m *MockToolHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
	if m.ListToolsFunc != nil {
		return m.ListToolsFunc(ctx)
	}
	return []shared.Tool{}, nil
}

// CallTool implements ToolHandler.CallTool
func (m *MockToolHandler) CallTool(ctx context.Context, name string, arguments interface{}) ([]shared.Content, error) {
	if m.CallToolFunc != nil {
		return m.CallToolFunc(ctx, name, arguments)
	}
	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: "Test result",
		},
	}, nil
}

// MockPromptHandler implements handler.PromptHandler for testing
type MockPromptHandler struct {
	ListPromptsFunc func(ctx context.Context) ([]shared.Prompt, error)
	CallPromptFunc  func(ctx context.Context, name string, arguments map[string]interface{}) ([]shared.Content, error)
}

// ListPrompts implements PromptHandler.ListPrompts
func (m *MockPromptHandler) ListPrompts(ctx context.Context) ([]shared.Prompt, error) {
	if m.ListPromptsFunc != nil {
		return m.ListPromptsFunc(ctx)
	}
	return []shared.Prompt{}, nil
}

// CallPrompt implements PromptHandler.CallPrompt
func (m *MockPromptHandler) CallPrompt(ctx context.Context, name string, arguments map[string]interface{}) ([]shared.Content, error) {
	if m.CallPromptFunc != nil {
		return m.CallPromptFunc(ctx, name, arguments)
	}
	return []shared.Content{
		shared.TextContent{
			Type: "text",
			Text: "Test prompt result",
		},
	}, nil
}
