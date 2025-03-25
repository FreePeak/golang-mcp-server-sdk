package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/handler"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
	mcperrors "github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared/errors"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/transport"
)

// Server represents an MCP server
type Server struct {
	info         shared.ServerInfo
	capabilities shared.Capabilities
	transport    transport.Transport

	resourceHandler handler.ResourceHandler
	toolHandler     handler.ToolHandler
	promptHandler   handler.PromptHandler

	requestHandlers map[string]handler.RequestHandler
	isInitialized   bool
	mu              sync.RWMutex
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		info: shared.ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities:    shared.Capabilities{},
		requestHandlers: make(map[string]handler.RequestHandler),
	}
}

// WithResourceHandler adds a resource handler to the server
func (s *Server) WithResourceHandler(handler handler.ResourceHandler) *Server {
	s.resourceHandler = handler
	s.capabilities.Resources = &shared.ResourcesCapability{}
	return s
}

// WithToolHandler adds a tool handler to the server
func (s *Server) WithToolHandler(handler handler.ToolHandler) *Server {
	s.toolHandler = handler
	s.capabilities.Tools = &shared.ToolsCapability{}
	return s
}

// WithPromptHandler adds a prompt handler to the server
func (s *Server) WithPromptHandler(handler handler.PromptHandler) *Server {
	s.promptHandler = handler
	s.capabilities.Prompts = &shared.PromptsCapability{}
	return s
}

// SetRequestHandler registers a custom request handler
func (s *Server) SetRequestHandler(method string, handler handler.RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestHandlers[method] = handler
}

// Connect connects the server to a transport
func (s *Server) Connect(transport transport.Transport) error {
	s.transport = transport
	return nil
}

// Start starts the server with the given context
func (s *Server) Start(ctx context.Context) error {
	if s.transport == nil {
		return errors.New("no transport specified")
	}

	return s.transport.Start(ctx, s.handleMessage)
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.transport == nil {
		return nil
	}

	return s.transport.Close()
}

// handleMessage processes incoming JSON-RPC messages
func (s *Server) handleMessage(ctx context.Context, message shared.JSONRPCMessage) error {
	if message.IsResponse() || message.IsNotification() {
		// We only handle requests
		return nil
	}

	req, ok := message.(shared.JSONRPCRequest)
	if !ok {
		return errors.New("invalid message type")
	}

	if req.Method == shared.MethodInitialize {
		return s.handleInitialize(ctx, req)
	}

	s.mu.RLock()
	if !s.isInitialized {
		s.mu.RUnlock()
		return s.sendErrorResponse(ctx, req, shared.InvalidRequest, "Server not initialized")
	}
	s.mu.RUnlock()

	switch req.Method {
	case shared.MethodShutdown:
		return s.handleShutdown(ctx, req)
	case shared.MethodListResources:
		return s.handleListResources(ctx, req)
	case shared.MethodGetResource:
		return s.handleGetResource(ctx, req)
	case shared.MethodListTools:
		return s.handleListTools(ctx, req)
	case shared.MethodCallTool:
		return s.handleCallTool(ctx, req)
	case shared.MethodListPrompts:
		return s.handleListPrompts(ctx, req)
	case shared.MethodCallPrompt:
		return s.handleCallPrompt(ctx, req)
	default:
		// Check for custom handlers
		s.mu.RLock()
		handler, exists := s.requestHandlers[req.Method]
		s.mu.RUnlock()

		if exists {
			return s.handleCustomRequest(ctx, req, handler)
		}

		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Method not found")
	}
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(ctx context.Context, req shared.JSONRPCRequest) error {
	var params shared.InitializeParams
	if err := unmarshalParams(req.Params, &params); err != nil {
		return s.sendErrorResponse(ctx, req, shared.InvalidParams, "Invalid params")
	}

	s.mu.Lock()
	s.isInitialized = true
	s.mu.Unlock()

	result := shared.InitializeResult{
		ServerInfo:   s.info,
		Capabilities: s.capabilities,
	}

	return s.sendResponse(ctx, req, result)
}

// handleShutdown handles the shutdown method
func (s *Server) handleShutdown(ctx context.Context, req shared.JSONRPCRequest) error {
	s.mu.Lock()
	s.isInitialized = false
	s.mu.Unlock()

	return s.sendResponse(ctx, req, nil)
}

// handleListResources handles the resources/list method
func (s *Server) handleListResources(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.resourceHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Resources not supported")
	}

	resources, err := s.resourceHandler.ListResources(ctx)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.ListResourcesResult{
		Resources: resources,
	}

	return s.sendResponse(ctx, req, result)
}

// handleGetResource handles the resources/get method
func (s *Server) handleGetResource(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.resourceHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Resources not supported")
	}

	var params shared.GetResourceParams
	if err := unmarshalParams(req.Params, &params); err != nil {
		return s.sendErrorResponse(ctx, req, shared.InvalidParams, "Invalid params")
	}

	content, err := s.resourceHandler.GetResource(ctx, params.URI)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.GetResourceResult{
		Content: content,
	}

	return s.sendResponse(ctx, req, result)
}

// handleListTools handles the tools/list method
func (s *Server) handleListTools(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.toolHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Tools not supported")
	}

	tools, err := s.toolHandler.ListTools(ctx)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.ListToolsResult{
		Tools: tools,
	}

	return s.sendResponse(ctx, req, result)
}

// handleCallTool handles the tools/call method
func (s *Server) handleCallTool(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.toolHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Tools not supported")
	}

	var params shared.CallToolParams
	if err := unmarshalParams(req.Params, &params); err != nil {
		return s.sendErrorResponse(ctx, req, shared.InvalidParams, "Invalid params")
	}

	content, err := s.toolHandler.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.CallToolResult{
		Content: content,
	}

	return s.sendResponse(ctx, req, result)
}

// handleListPrompts handles the prompts/list method
func (s *Server) handleListPrompts(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.promptHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Prompts not supported")
	}

	prompts, err := s.promptHandler.ListPrompts(ctx)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.ListPromptsResult{
		Prompts: prompts,
	}

	return s.sendResponse(ctx, req, result)
}

// handleCallPrompt handles the prompts/call method
func (s *Server) handleCallPrompt(ctx context.Context, req shared.JSONRPCRequest) error {
	if s.promptHandler == nil {
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Prompts not supported")
	}

	var params shared.CallPromptParams
	if err := unmarshalParams(req.Params, &params); err != nil {
		return s.sendErrorResponse(ctx, req, shared.InvalidParams, "Invalid params")
	}

	content, err := s.promptHandler.CallPrompt(ctx, params.Name, params.Arguments)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	result := shared.CallPromptResult{
		Content: content,
	}

	return s.sendResponse(ctx, req, result)
}

// handleCustomRequest handles a custom request method
func (s *Server) handleCustomRequest(ctx context.Context, req shared.JSONRPCRequest, handler handler.RequestHandler) error {
	result, err := handler(ctx, req.Params)
	if err != nil {
		return s.sendMCPErrorResponse(ctx, req, err)
	}

	return s.sendResponse(ctx, req, result)
}

// sendResponse sends a JSON-RPC response
func (s *Server) sendResponse(ctx context.Context, req shared.JSONRPCRequest, result interface{}) error {
	response := shared.JSONRPCResponse{
		JSONRPC: shared.JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}

	return s.transport.Send(ctx, response)
}

// sendErrorResponse sends a JSON-RPC error response
func (s *Server) sendErrorResponse(ctx context.Context, req shared.JSONRPCRequest, code shared.ErrorCode, message string) error {
	response := shared.JSONRPCResponse{
		JSONRPC: shared.JSONRPCVersion,
		ID:      req.ID,
		Error: &shared.JSONRPCError{
			Code:    int(code),
			Message: message,
		},
	}

	return s.transport.Send(ctx, response)
}

// sendMCPErrorResponse maps MCP errors to appropriate JSON-RPC errors
func (s *Server) sendMCPErrorResponse(ctx context.Context, req shared.JSONRPCRequest, err error) error {
	var code shared.ErrorCode
	var message string

	var mcpErr *mcperrors.MCPError
	if errors.As(err, &mcpErr) {
		switch mcpErr.Type {
		case mcperrors.ErrorTypeNotFound:
			code = shared.InvalidParams
			message = fmt.Sprintf("Not found: %s", mcpErr.Message)
		case mcperrors.ErrorTypeInvalidInput:
			code = shared.InvalidParams
			message = fmt.Sprintf("Invalid input: %s", mcpErr.Message)
		case mcperrors.ErrorTypeUnauthorized:
			code = shared.InvalidRequest
			message = fmt.Sprintf("Unauthorized: %s", mcpErr.Message)
		default:
			code = shared.InternalError
			message = fmt.Sprintf("Internal error: %s", mcpErr.Message)
		}
	} else {
		code = shared.InternalError
		message = fmt.Sprintf("Internal error: %v", err)
	}

	return s.sendErrorResponse(ctx, req, code, message)
}

// unmarshalParams unmarshals request parameters
func unmarshalParams(params interface{}, target interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}
