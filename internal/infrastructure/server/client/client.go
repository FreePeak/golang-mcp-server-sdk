package client

import (
	"fmt"
	"strings"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// ClientType represents the type of client connecting to the server
type ClientType string

const (
	// ClientTypeCursor represents Cursor IDE client
	ClientTypeCursor ClientType = "cursor"
	// ClientTypeClaude represents Claude Desktop client
	ClientTypeClaude ClientType = "claude"
	// ClientTypeGeneric represents a generic client
	ClientTypeGeneric ClientType = "generic"
)

// Config defines the interface for client-specific configurations
type Config interface {
	// GetClientType returns the type of client this configuration is for
	GetClientType() ClientType

	// GetDefaultTools returns the default tools for this client
	GetDefaultTools() []shared.Tool

	// ConfigureServerInfo customizes server info for this client if needed
	ConfigureServerInfo(info *shared.ServerInfo)

	// ConfigureCapabilities customizes capabilities for this client if needed
	ConfigureCapabilities(capabilities *shared.Capabilities)
}

// ConfigRegistry is a registry of client configurations
type ConfigRegistry struct {
	configs map[ClientType]Config
}

// NewConfigRegistry creates a new client configuration registry
func NewConfigRegistry() *ConfigRegistry {
	return &ConfigRegistry{
		configs: make(map[ClientType]Config),
	}
}

// Register registers a client configuration
func (r *ConfigRegistry) Register(config Config) {
	r.configs[config.GetClientType()] = config
}

// GetConfig returns the configuration for a specific client type
func (r *ConfigRegistry) GetConfig(clientType ClientType) Config {
	config, exists := r.configs[clientType]
	if !exists {
		// Return a generic config if the specific one doesn't exist
		return NewGenericConfig()
	}
	return config
}

// DetectClientType attempts to detect the client type from client info
func DetectClientType(clientInfo shared.ServerInfo) ClientType {
	clientName := strings.ToLower(clientInfo.Name)

	fmt.Printf("Detecting client type for name: '%s'\n", clientInfo.Name)

	switch {
	case strings.Contains(clientName, "cursor"):
		fmt.Printf("Detected Cursor client\n")
		return ClientTypeCursor
	case clientName == "claude":
		fmt.Printf("Detected Claude client\n")
		return ClientTypeClaude
	default:
		fmt.Printf("Using generic client type for: %s\n", clientInfo.Name)
		return ClientTypeGeneric
	}
}
