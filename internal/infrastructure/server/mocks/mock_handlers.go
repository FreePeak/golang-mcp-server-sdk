package mocks

import (
	"context"
	"fmt"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
)

// MockResourceHandler is a mock implementation of ResourceHandler
type MockResourceHandler struct {
	resources        []shared.Resource
	resourceContents map[string][]shared.Content
	listError        error
	getErrors        map[string]error
}

// NewMockResourceHandler creates a new mock resource handler
func NewMockResourceHandler() *MockResourceHandler {
	return &MockResourceHandler{
		resources:        make([]shared.Resource, 0),
		resourceContents: make(map[string][]shared.Content),
		getErrors:        make(map[string]error),
	}
}

// AddResource adds a resource to the mock handler
func (m *MockResourceHandler) AddResource(uri, name, description string, content []shared.Content) {
	m.resources = append(m.resources, shared.Resource{
		URI:         uri,
		Name:        name,
		Description: description,
	})
	m.resourceContents[uri] = content
}

// SetListError sets the error to return from ListResources
func (m *MockResourceHandler) SetListError(err error) {
	m.listError = err
}

// SetGetError sets the error to return from GetResource for a specific URI
func (m *MockResourceHandler) SetGetError(uri string, err error) {
	m.getErrors[uri] = err
}

// ListResources returns the list of resources
func (m *MockResourceHandler) ListResources(ctx context.Context) ([]shared.Resource, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.resources, nil
}

// GetResource returns the content for a resource
func (m *MockResourceHandler) GetResource(ctx context.Context, uri string) ([]shared.Content, error) {
	if err, ok := m.getErrors[uri]; ok && err != nil {
		return nil, err
	}

	content, ok := m.resourceContents[uri]
	if !ok {
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("resource %s not found", uri), nil)
	}

	return content, nil
}

// MockToolHandler is a mock implementation of ToolHandler
type MockToolHandler struct {
	tools       []shared.Tool
	toolResults map[string][]shared.Content
	listError   error
	callErrors  map[string]error
	calledTools map[string][]interface{}
}

// NewMockToolHandler creates a new mock tool handler
func NewMockToolHandler() *MockToolHandler {
	return &MockToolHandler{
		tools:       make([]shared.Tool, 0),
		toolResults: make(map[string][]shared.Content),
		callErrors:  make(map[string]error),
		calledTools: make(map[string][]interface{}),
	}
}

// AddTool adds a tool to the mock handler
func (m *MockToolHandler) AddTool(name, description string, inputSchema interface{}, result []shared.Content) {
	m.tools = append(m.tools, shared.Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	})
	m.toolResults[name] = result
}

// SetListError sets the error to return from ListTools
func (m *MockToolHandler) SetListError(err error) {
	m.listError = err
}

// SetCallError sets the error to return from CallTool for a specific tool
func (m *MockToolHandler) SetCallError(name string, err error) {
	m.callErrors[name] = err
}

// GetCalledArgs returns the arguments a tool was called with
func (m *MockToolHandler) GetCalledArgs(name string) []interface{} {
	return m.calledTools[name]
}

// ListTools returns the list of tools
func (m *MockToolHandler) ListTools(ctx context.Context) ([]shared.Tool, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tools, nil
}

// CallTool calls a tool and returns the result
func (m *MockToolHandler) CallTool(ctx context.Context, name string, args interface{}) ([]shared.Content, error) {
	if err, ok := m.callErrors[name]; ok && err != nil {
		return nil, err
	}

	// Record the call
	m.calledTools[name] = append(m.calledTools[name], args)

	result, ok := m.toolResults[name]
	if !ok {
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("tool %s not found", name), nil)
	}

	return result, nil
}

// MockPromptHandler is a mock implementation of PromptHandler
type MockPromptHandler struct {
	prompts       []shared.Prompt
	promptResults map[string][]shared.Content
	listError     error
	callErrors    map[string]error
	calledPrompts map[string][]map[string]interface{}
}

// NewMockPromptHandler creates a new mock prompt handler
func NewMockPromptHandler() *MockPromptHandler {
	return &MockPromptHandler{
		prompts:       make([]shared.Prompt, 0),
		promptResults: make(map[string][]shared.Content),
		callErrors:    make(map[string]error),
		calledPrompts: make(map[string][]map[string]interface{}),
	}
}

// AddPrompt adds a prompt to the mock handler
func (m *MockPromptHandler) AddPrompt(name, description string, args []shared.PromptArgument, result []shared.Content) {
	m.prompts = append(m.prompts, shared.Prompt{
		Name:        name,
		Description: description,
		Arguments:   args,
	})
	m.promptResults[name] = result
}

// SetListError sets the error to return from ListPrompts
func (m *MockPromptHandler) SetListError(err error) {
	m.listError = err
}

// SetCallError sets the error to return from CallPrompt for a specific prompt
func (m *MockPromptHandler) SetCallError(name string, err error) {
	m.callErrors[name] = err
}

// GetCalledArgs returns the arguments a prompt was called with
func (m *MockPromptHandler) GetCalledArgs(name string) []map[string]interface{} {
	return m.calledPrompts[name]
}

// ListPrompts returns the list of prompts
func (m *MockPromptHandler) ListPrompts(ctx context.Context) ([]shared.Prompt, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.prompts, nil
}

// CallPrompt calls a prompt and returns the result
func (m *MockPromptHandler) CallPrompt(ctx context.Context, name string, args map[string]interface{}) ([]shared.Content, error) {
	if err, ok := m.callErrors[name]; ok && err != nil {
		return nil, err
	}

	// Record the call
	m.calledPrompts[name] = append(m.calledPrompts[name], args)

	result, ok := m.promptResults[name]
	if !ok {
		return nil, mcperrors.NewNotFoundError(fmt.Sprintf("prompt %s not found", name), nil)
	}

	return result, nil
}
