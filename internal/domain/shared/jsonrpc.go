package shared

// JSONRPCVersion is the version of JSON-RPC used by MCP
const JSONRPCVersion = "2.0"

// JSONRPCMessage represents a JSON-RPC message
type JSONRPCMessage interface {
	IsRequest() bool
	IsResponse() bool
	IsNotification() bool
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// IsRequest returns true for request messages
func (r JSONRPCRequest) IsRequest() bool {
	return true
}

// IsResponse returns false for request messages
func (r JSONRPCRequest) IsResponse() bool {
	return false
}

// IsNotification returns false for request messages
func (r JSONRPCRequest) IsNotification() bool {
	return false
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// IsRequest returns false for response messages
func (r JSONRPCResponse) IsRequest() bool {
	return false
}

// IsResponse returns true for response messages
func (r JSONRPCResponse) IsResponse() bool {
	return true
}

// IsNotification returns false for response messages
func (r JSONRPCResponse) IsNotification() bool {
	return false
}

// JSONRPCNotification represents a JSON-RPC notification
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// IsRequest returns false for notification messages
func (n JSONRPCNotification) IsRequest() bool {
	return false
}

// IsResponse returns false for notification messages
func (n JSONRPCNotification) IsResponse() bool {
	return false
}

// IsNotification returns true for notification messages
func (n JSONRPCNotification) IsNotification() bool {
	return true
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorCode defines standard JSON-RPC error codes
type ErrorCode int

const (
	// ParseError indicates an error while parsing the JSON text
	ParseError ErrorCode = -32700
	// InvalidRequest indicates the JSON was not a valid Request object
	InvalidRequest ErrorCode = -32600
	// MethodNotFound indicates the method does not exist / is not available
	MethodNotFound ErrorCode = -32601
	// InvalidParams indicates invalid method parameters
	InvalidParams ErrorCode = -32602
	// InternalError indicates an internal JSON-RPC error
	InternalError ErrorCode = -32603
)

// ErrorMessage returns a standard error message for a given error code
func ErrorMessage(code ErrorCode) string {
	switch code {
	case ParseError:
		return "Parse error"
	case InvalidRequest:
		return "Invalid request"
	case MethodNotFound:
		return "Method not found"
	case InvalidParams:
		return "Invalid params"
	case InternalError:
		return "Internal error"
	default:
		return "Unknown error"
	}
}
