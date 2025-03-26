package client

import (
	"github.com/FreePeak/golang-mcp-server-sdk/internal/domain/shared"
)

// GenericConfig represents a generic client configuration
type GenericConfig struct{}

// NewGenericConfig creates a new generic client configuration
func NewGenericConfig() *GenericConfig {
	return &GenericConfig{}
}

// GetClientType returns the client type
func (c *GenericConfig) GetClientType() ClientType {
	return ClientTypeGeneric
}

// GetDefaultTools returns the default tools for a generic client
func (c *GenericConfig) GetDefaultTools() []shared.Tool {
	// Generic client has no default tools
	return []shared.Tool{}
}

// ConfigureServerInfo customizes server info for a generic client
func (c *GenericConfig) ConfigureServerInfo(info *shared.ServerInfo) {
	// No customization needed for generic clients
}

// ConfigureCapabilities customizes capabilities for a generic client
func (c *GenericConfig) ConfigureCapabilities(capabilities *shared.Capabilities) {
	// No customization needed for generic clients
}
