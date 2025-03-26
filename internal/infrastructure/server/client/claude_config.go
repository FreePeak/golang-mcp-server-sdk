package client

import (
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// ClaudeConfig represents the configuration for Claude Desktop clients
type ClaudeConfig struct{}

// NewClaudeConfig creates a new Claude client configuration
func NewClaudeConfig() *ClaudeConfig {
	return &ClaudeConfig{}
}

// GetClientType returns the client type
func (c *ClaudeConfig) GetClientType() ClientType {
	return ClientTypeClaude
}

// GetDefaultTools returns the default tools for Claude Desktop
func (c *ClaudeConfig) GetDefaultTools() []shared.Tool {
	// Default tools for Claude Desktop
	// This can be expanded in the future
	return []shared.Tool{
		{
			Name:        "file_search",
			Description: "Search for files by name or pattern",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "File pattern to search for",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

// ConfigureServerInfo customizes server info for Claude client
func (c *ClaudeConfig) ConfigureServerInfo(info *shared.ServerInfo) {
	// No special customization needed for Claude clients
}

// ConfigureCapabilities customizes capabilities for Claude client
func (c *ClaudeConfig) ConfigureCapabilities(capabilities *shared.Capabilities) {
	// Claude supports all capabilities
	capabilities.Tools = &shared.ToolsCapability{}
	capabilities.Resources = &shared.ResourcesCapability{}
	capabilities.Prompts = &shared.PromptsCapability{}
}
