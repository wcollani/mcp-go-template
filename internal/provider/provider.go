package provider

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Provider is the core interface every integration must implement.
// Resources expose read-only state; Tools expose actions/mutations.
type Provider interface {
	Name() string

	// Resources (read-only state)
	GetResources() ([]mcp.Resource, error)
	GetResourceContent(uri string) (string, error)
	GetResourceTemplates() ([]mcp.ResourceTemplate, error)

	// Prompts
	GetPrompts() ([]mcp.Prompt, error)
	GetPrompt(name string, arguments map[string]string) (*mcp.GetPromptResult, error)

	// Tools (actions / mutations)
	GetTools() ([]mcp.Tool, error)
	CallTool(name string, arguments map[string]interface{}) (*mcp.CallToolResult, error)
}
