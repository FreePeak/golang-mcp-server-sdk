package mocks

import (
	"context"
	"sync"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// MockTransport is a mock implementation of the Transport interface
type MockTransport struct {
	messages      []shared.JSONRPCMessage
	startCalled   bool
	closeCalled   bool
	handler       transport.MessageHandler
	mu            sync.Mutex
	messagesSent  []shared.JSONRPCMessage
	startErr      error
	sendErr       error
	closeErr      error
	triggerCond   *sync.Cond
	waitForResult bool
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	mt := &MockTransport{
		messages:     make([]shared.JSONRPCMessage, 0),
		messagesSent: make([]shared.JSONRPCMessage, 0),
	}
	mt.triggerCond = sync.NewCond(&mt.mu)
	return mt
}

// SetStartError sets the error to return from Start
func (m *MockTransport) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startErr = err
}

// SetSendError sets the error to return from Send
func (m *MockTransport) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = err
}

// SetCloseError sets the error to return from Close
func (m *MockTransport) SetCloseError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeErr = err
}

// SetWaitForResult sets whether to wait for results before proceeding
func (m *MockTransport) SetWaitForResult(wait bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.waitForResult = wait
}

// Start starts the mock transport
func (m *MockTransport) Start(ctx context.Context, handler transport.MessageHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCalled = true
	m.handler = handler

	if m.startErr != nil {
		return m.startErr
	}

	return nil
}

// Close closes the mock transport
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true

	if m.closeErr != nil {
		return m.closeErr
	}

	return nil
}

// Send records a sent message
func (m *MockTransport) Send(ctx context.Context, message shared.JSONRPCMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messagesSent = append(m.messagesSent, message)
	m.triggerCond.Broadcast()

	if m.sendErr != nil {
		return m.sendErr
	}

	return nil
}

// AddMessage adds a message to be processed
func (m *MockTransport) AddMessage(message shared.JSONRPCMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
}

// IsStartCalled checks if Start was called
func (m *MockTransport) IsStartCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.startCalled
}

// IsCloseCalled checks if Close was called
func (m *MockTransport) IsCloseCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeCalled
}

// GetMessagesSent gets all sent messages
func (m *MockTransport) GetMessagesSent() []shared.JSONRPCMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messagesSent
}

// ProcessMessage processes a message through the handler
func (m *MockTransport) ProcessMessage(ctx context.Context, message shared.JSONRPCMessage) error {
	m.mu.Lock()
	handler := m.handler
	waitForResult := m.waitForResult
	m.mu.Unlock()

	if handler == nil {
		return nil
	}

	err := handler(ctx, message)

	if waitForResult {
		// Wait for a response to be sent
		m.mu.Lock()
		originalLen := len(m.messagesSent)
		for len(m.messagesSent) <= originalLen {
			m.triggerCond.Wait()
		}
		m.mu.Unlock()
	}

	return err
}

// WaitForMessageCount waits until at least count messages have been sent
func (m *MockTransport) WaitForMessageCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for len(m.messagesSent) < count {
		m.triggerCond.Wait()
	}
}

// ClearMessagesSent clears the sent messages
func (m *MockTransport) ClearMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messagesSent = make([]shared.JSONRPCMessage, 0)
}
