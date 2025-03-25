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
	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server/client"
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

	// Client configuration registry
	clientRegistry *client.ConfigRegistry

	// Current client configuration
	currentClientConfig client.Config
}

// ServerOption represents a function that configures a server
type ServerOption func(*Server)

// WithResourceHandler returns an option to add a resource handler
func WithResourceHandler(handler handler.ResourceHandler) ServerOption {
	return func(s *Server) {
		s.resourceHandler = handler
		s.capabilities.Resources = &shared.ResourcesCapability{}
	}
}

// WithToolHandler returns an option to add a tool handler
func WithToolHandler(handler handler.ToolHandler) ServerOption {
	return func(s *Server) {
		s.toolHandler = handler
		s.capabilities.Tools = &shared.ToolsCapability{}
	}
}

// WithPromptHandler returns an option to add a prompt handler
func WithPromptHandler(handler handler.PromptHandler) ServerOption {
	return func(s *Server) {
		s.promptHandler = handler
		s.capabilities.Prompts = &shared.PromptsCapability{}
	}
}

// WithRequestHandler returns an option to add a custom request handler
func WithRequestHandler(method string, handler handler.RequestHandler) ServerOption {
	return func(s *Server) {
		s.requestHandlers[method] = handler
	}
}

// NewServer creates a new MCP server with the given options
func NewServer(name, version string, opts ...ServerOption) *Server {
	s := &Server{
		info: shared.ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities:    shared.Capabilities{},
		requestHandlers: make(map[string]handler.RequestHandler),
		clientRegistry:  client.NewConfigRegistry(),
	}

	// Register client configurations
	s.clientRegistry.Register(client.NewGenericConfig())
	s.clientRegistry.Register(client.NewCursorConfig())
	s.clientRegistry.Register(client.NewClaudeConfig())

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	return s
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

	// Register the message handler
	transport.SetCurrentHandler(s.handleMessage)

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
		fmt.Printf("Error: Invalid message type received, expected JSONRPCRequest\n")
		return errors.New("invalid message type")
	}

	fmt.Printf("Received request: Method=%s, ID=%v\n", req.Method, req.ID)

	if req.Method == shared.MethodInitialize {
		return s.handleInitialize(ctx, req)
	}

	s.mu.RLock()
	initialized := s.isInitialized
	s.mu.RUnlock()

	if !initialized {
		fmt.Printf("Error: Server not initialized, rejecting request for method: %s\n", req.Method)
		return s.sendErrorResponse(ctx, req, shared.InvalidRequest, "Server not initialized")
	}

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
	fmt.Printf("Processing initialize request\n")

	var params shared.InitializeParams
	if err := unmarshalParams(req.Params, &params); err != nil {
		fmt.Printf("Error unmarshalling initialize params: %v\n", err)
		return s.sendErrorResponse(ctx, req, shared.InvalidParams, "Invalid params")
	}

	fmt.Printf("Initialize: Client info - Name=%s, Version=%s\n",
		params.ClientInfo.Name, params.ClientInfo.Version)

	// Configure client-specific behavior
	clientType := client.DetectClientType(params.ClientInfo)
	fmt.Printf("Client type detected: %s\n", clientType)

	s.currentClientConfig = s.clientRegistry.GetConfig(clientType)

	// Apply client-specific configuration
	s.currentClientConfig.ConfigureServerInfo(&s.info)
	s.currentClientConfig.ConfigureCapabilities(&s.capabilities)

	s.mu.Lock()
	s.isInitialized = true
	s.mu.Unlock()
	fmt.Printf("Server initialized successfully\n")

	result := shared.InitializeResult{
		ServerInfo:   s.info,
		Capabilities: s.capabilities,
	}

	fmt.Printf("Sending initialize response with capabilities: %+v\n", s.capabilities)
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
	fmt.Printf("Processing tools/list request\n")
	var tools []shared.Tool

	if s.toolHandler != nil {
		fmt.Printf("Getting tools from tool handler\n")
		var err error
		tools, err = s.toolHandler.ListTools(ctx)
		if err != nil {
			fmt.Printf("Error getting tools from tool handler: %v\n", err)
			return s.sendMCPErrorResponse(ctx, req, err)
		}
		fmt.Printf("Got %d tools from tool handler\n", len(tools))
	} else {
		fmt.Printf("No tool handler registered\n")
	}

	// If we have a client configuration, add its default tools
	if s.currentClientConfig != nil {
		fmt.Printf("Getting default tools for client type: %s\n", s.currentClientConfig.GetClientType())
		defaultTools := s.currentClientConfig.GetDefaultTools()
		if len(defaultTools) > 0 {
			fmt.Printf("Adding %d default tools for client\n", len(defaultTools))
			tools = append(tools, defaultTools...)
		} else {
			fmt.Printf("No default tools for client\n")
		}
	} else {
		fmt.Printf("No client configuration available\n")
	}

	// If we don't have any tools, return an error
	if len(tools) == 0 {
		fmt.Printf("No tools available, returning error\n")
		return s.sendErrorResponse(ctx, req, shared.MethodNotFound, "Tools not supported")
	}

	result := shared.ListToolsResult{
		Tools: tools,
	}

	fmt.Printf("Returning %d tools\n", len(tools))
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

// sendResponse sends a success response
func (s *Server) sendResponse(ctx context.Context, req shared.JSONRPCRequest, result interface{}) error {
	response := shared.JSONRPCResponse{
		JSONRPC: shared.JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}

	fmt.Printf("Sending response for method %s with ID: %v\n", req.Method, req.ID)

	if s.transport == nil {
		errMsg := "transport not initialized"
		fmt.Printf("Error sending response: %s\n", errMsg)
		return errors.New(errMsg)
	}

	if err := s.transport.Send(ctx, response); err != nil {
		fmt.Printf("Error sending response: %v\n", err)
		return err
	}

	fmt.Printf("Response sent successfully\n")
	return nil
}

// sendErrorResponse sends an error response
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

// sendMCPErrorResponse sends an error response based on an MCP error
func (s *Server) sendMCPErrorResponse(ctx context.Context, req shared.JSONRPCRequest, err error) error {
	// Check if it's a known error type
	var code shared.ErrorCode
	var message string

	switch e := err.(type) {
	case *mcperrors.ResourceNotFoundError:
		code = shared.ResourceNotFound
		message = e.Error()
	case *mcperrors.ResourceAccessDeniedError:
		code = shared.ResourceAccessDenied
		message = e.Error()
	case *mcperrors.ToolNotFoundError:
		code = shared.ToolNotFound
		message = e.Error()
	case *mcperrors.ToolExecutionError:
		code = shared.ToolExecutionFailed
		message = e.Error()
	case *mcperrors.PromptNotFoundError:
		code = shared.PromptNotFound
		message = e.Error()
	case *mcperrors.PromptExecutionError:
		code = shared.PromptExecutionFailed
		message = e.Error()
	case *mcperrors.MCPError:
		// Handle generic MCP errors
		switch e.Type {
		case mcperrors.ErrorTypeNotFound:
			code = shared.InvalidParams
			message = fmt.Sprintf("Not found: %s", e.Message)
		case mcperrors.ErrorTypeInvalidInput:
			code = shared.InvalidParams
			message = fmt.Sprintf("Invalid input: %s", e.Message)
		case mcperrors.ErrorTypeUnauthorized:
			code = shared.InvalidRequest
			message = fmt.Sprintf("Unauthorized: %s", e.Message)
		default:
			code = shared.InternalError
			message = fmt.Sprintf("Internal error: %s", e.Message)
		}
	default:
		// Unknown error type
		code = shared.InternalError
		message = fmt.Sprintf("Internal error: %v", err)
	}

	return s.sendErrorResponse(ctx, req, shared.InternalError, err.Error())
}

// unmarshalParams unmarshals the params field into the target struct
func unmarshalParams(params interface{}, target interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// generateID generates a unique ID for a request
func generateID() string {
	return uuid.New().String()
}
