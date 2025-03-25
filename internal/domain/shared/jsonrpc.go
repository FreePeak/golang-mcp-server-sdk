package shared

import (
	"encoding/json"
)

// JSONRPCVersion is the version of JSON-RPC to use
const JSONRPCVersion = "2.0"

// ErrorCode represents a JSON-RPC error code
type ErrorCode int

// Standard JSON-RPC error codes
const (
	ParseError     ErrorCode = -32700
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
	ServerError    ErrorCode = -32000
	// MCP-specific error codes
	NotFound ErrorCode = -32001
)

// JSONRPCMessage is the interface that all JSON-RPC messages implement
type JSONRPCMessage interface {
	// IsRequest returns true if the message is a request
	IsRequest() bool
	// IsResponse returns true if the message is a response
	IsResponse() bool
	// IsNotification returns true if the message is a notification
	IsNotification() bool
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsRequest returns true for requests
func (r JSONRPCRequest) IsRequest() bool {
	return true
}

// IsResponse returns false for requests
func (r JSONRPCRequest) IsResponse() bool {
	return false
}

// IsNotification returns false for requests
func (r JSONRPCRequest) IsNotification() bool {
	return false
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// IsRequest returns false for responses
func (r JSONRPCResponse) IsRequest() bool {
	return false
}

// IsResponse returns true for responses
func (r JSONRPCResponse) IsResponse() bool {
	return true
}

// IsNotification returns false for responses
func (r JSONRPCResponse) IsNotification() bool {
	return false
}

// JSONRPCNotification represents a JSON-RPC notification
type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsRequest returns false for notifications
func (n JSONRPCNotification) IsRequest() bool {
	return false
}

// IsResponse returns false for notifications
func (n JSONRPCNotification) IsResponse() bool {
	return false
}

// IsNotification returns true for notifications
func (n JSONRPCNotification) IsNotification() bool {
	return true
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

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
	case ServerError:
		return "Server error"
	case NotFound:
		return "Not found"
	default:
		return "Unknown error"
	}
}
