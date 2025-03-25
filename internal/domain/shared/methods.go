package shared

// MCP method names
const (
	// Core methods
	MethodInitialize = "initialize"
	MethodShutdown   = "shutdown"

	// Resource methods
	MethodListResources = "resources/list"
	MethodGetResource   = "resources/get"

	// Tool methods
	MethodListTools = "tools/list"
	MethodCallTool  = "tools/call"

	// Prompt methods
	MethodListPrompts = "prompts/list"
	MethodCallPrompt  = "prompts/call"
)

// InitializeParams represents parameters for the initialize method
type InitializeParams struct {
	ClientInfo ServerInfo            `json:"clientInfo"`
	Options    InitializationOptions `json:"options"`
}

// InitializeResult represents the result of the initialize method
type InitializeResult struct {
	ServerInfo   ServerInfo   `json:"serverInfo"`
	Capabilities Capabilities `json:"capabilities"`
}

// ListResourcesResult represents the result of the resources/list method
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// GetResourceParams represents parameters for the resources/get method
type GetResourceParams struct {
	URI string `json:"uri"`
}

// GetResourceResult represents the result of the resources/get method
type GetResourceResult struct {
	Content []Content `json:"content"`
}

// ListToolsResult represents the result of the tools/list method
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams represents parameters for the tools/call method
type CallToolParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

// CallToolResult represents the result of the tools/call method
type CallToolResult struct {
	Content []Content `json:"content"`
}

// ListPromptsResult represents the result of the prompts/list method
type ListPromptsResult struct {
	Prompts []Prompt `json:"prompts"`
}

// CallPromptParams represents parameters for the prompts/call method
type CallPromptParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CallPromptResult represents the result of the prompts/call method
type CallPromptResult struct {
	Content []Content `json:"content"`
}
