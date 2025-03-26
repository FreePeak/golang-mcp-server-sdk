package shared

// MCP-specific error codes
const (
	// Resource error codes
	ResourceNotFound     ErrorCode = -32100
	ResourceAccessDenied ErrorCode = -32101

	// Tool error codes
	ToolNotFound        ErrorCode = -32200
	ToolExecutionFailed ErrorCode = -32201

	// Prompt error codes
	PromptNotFound        ErrorCode = -32300
	PromptExecutionFailed ErrorCode = -32301
)
