package shared

// ServerInfo contains information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializationOptions contains options for initializing the server
type InitializationOptions struct {
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities represents the server's capabilities
type Capabilities struct {
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// ResourcesCapability indicates support for resources
type ResourcesCapability struct{}

// ToolsCapability indicates support for tools
type ToolsCapability struct{}

// PromptsCapability indicates support for prompts
type PromptsCapability struct{}

// Resource represents a resource exposed by the server
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Tool represents a tool exposed by the server
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema"`
}

// Content represents content returned by tools
type Content interface {
	GetType() string
}

// TextContent represents text content
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// GetType returns the content type
func (t TextContent) GetType() string {
	return t.Type
}

// ImageContent represents image content
type ImageContent struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// GetType returns the content type
func (i ImageContent) GetType() string {
	return i.Type
}

// EmbeddedResource represents a reference to a resource
type EmbeddedResource struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// GetType returns the content type
func (e EmbeddedResource) GetType() string {
	return e.Type
}

// Prompt represents a prompt exposed by the server
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	Schema      interface{} `json:"schema"`
}
