package client

import (
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// CursorConfig represents the configuration for Cursor IDE clients
type CursorConfig struct{}

// NewCursorConfig creates a new Cursor client configuration
func NewCursorConfig() *CursorConfig {
	return &CursorConfig{}
}

// GetClientType returns the client type
func (c *CursorConfig) GetClientType() ClientType {
	return ClientTypeCursor
}

// GetDefaultTools returns the default tools for Cursor IDE
func (c *CursorConfig) GetDefaultTools() []shared.Tool {
	return []shared.Tool{
		{
			Name:        "file_search",
			Description: "Fast file search based on fuzzy matching against file path. Use if you know part of the file path but don't know where it's located exactly.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Fuzzy filename to search for",
					},
					"explanation": map[string]interface{}{
						"type":        "string",
						"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
					},
				},
				"required": []string{"query", "explanation"},
			},
		},
		{
			Name:        "codebase_search",
			Description: "Find snippets of code from the codebase most relevant to the search query.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query to find relevant code.",
					},
					"explanation": map[string]interface{}{
						"type":        "string",
						"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
					},
					"target_directories": map[string]interface{}{
						"type":        "array",
						"description": "Glob patterns for directories to search over",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "grep_search",
			Description: "Fast text-based regex search that finds exact pattern matches within files or directories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The regex pattern to search for",
					},
					"case_sensitive": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the search should be case sensitive",
					},
					"include_pattern": map[string]interface{}{
						"type":        "string",
						"description": "Glob pattern for files to include (e.g. '*.ts' for TypeScript files)",
					},
					"exclude_pattern": map[string]interface{}{
						"type":        "string",
						"description": "Glob pattern for files to exclude",
					},
					"explanation": map[string]interface{}{
						"type":        "string",
						"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "list_dir",
			Description: "List the contents of a directory. The quick tool to use for discovery, before using more targeted tools.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"relative_workspace_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to list contents of, relative to the workspace root.",
					},
					"explanation": map[string]interface{}{
						"type":        "string",
						"description": "One sentence explanation as to why this tool is being used, and how it contributes to the goal.",
					},
				},
				"required": []string{"relative_workspace_path"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_file": map[string]interface{}{
						"type":        "string",
						"description": "The path of the file to read.",
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "The offset to start reading from.",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "The number of lines to read.",
					},
					"should_read_entire_file": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to read the entire file.",
					},
				},
				"required": []string{"target_file"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Edit a file at the specified path.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_file": map[string]interface{}{
						"type":        "string",
						"description": "The target file to modify.",
					},
					"instructions": map[string]interface{}{
						"type":        "string",
						"description": "A single sentence instruction describing what you are going to do for the sketched edit.",
					},
					"code_edit": map[string]interface{}{
						"type":        "string",
						"description": "Specify ONLY the precise lines of code that you wish to edit.",
					},
				},
				"required": []string{"target_file", "instructions", "code_edit"},
			},
		},
		{
			Name:        "run_terminal_cmd",
			Description: "Run a command in the terminal.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The terminal command to execute",
					},
					"explanation": map[string]interface{}{
						"type":        "string",
						"description": "One sentence explanation as to why this command needs to be run and how it contributes to the goal.",
					},
					"is_background": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the command should be run in the background",
					},
					"require_user_approval": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the user must approve the command before it is executed.",
					},
				},
				"required": []string{"command", "is_background", "require_user_approval"},
			},
		},
	}
}

// ConfigureServerInfo customizes server info for Cursor client
func (c *CursorConfig) ConfigureServerInfo(info *shared.ServerInfo) {
	// No special customization needed for Cursor clients
}

// ConfigureCapabilities customizes capabilities for Cursor client
func (c *CursorConfig) ConfigureCapabilities(capabilities *shared.Capabilities) {
	// Cursor needs tools capability
	capabilities.Tools = &shared.ToolsCapability{}
}
