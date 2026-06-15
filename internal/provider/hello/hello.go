package hello

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcphelper "github.com/wcollani/mcp-go-template/internal/mcp"
)

// Provider is a minimal example provider that demonstrates resources and tools.
// Copy this package as a starting point for a new integration.
type Provider struct{}

func (h *Provider) Name() string { return "hello" }

func (h *Provider) GetResources() ([]mcp.Resource, error) {
	return []mcp.Resource{
		{
			URI:         "hello://world",
			Name:        "Hello World",
			Description: "A simple hello world resource",
			MIMEType:    "text/plain",
		},
	}, nil
}

func (h *Provider) GetResourceContent(uri string) (string, error) {
	if uri == "hello://world" {
		return "Hello from your MCP server!", nil
	}
	return "", fmt.Errorf("resource not found: %s", uri)
}

func (h *Provider) GetResourceTemplates() ([]mcp.ResourceTemplate, error) {
	return []mcp.ResourceTemplate{}, nil
}

func (h *Provider) GetPrompts() ([]mcp.Prompt, error) { return []mcp.Prompt{}, nil }
func (h *Provider) GetPrompt(name string, arguments map[string]string) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("prompt not found: %s", name)
}

func (h *Provider) GetTools() ([]mcp.Tool, error) {
	return []mcp.Tool{
		*mcphelper.NewTool(
			"greet",
			"Return a greeting for the given name.",
			map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "The name to greet",
					"required":    true,
				},
			},
		),
	}, nil
}

func (h *Provider) CallTool(name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if name == "greet" {
		who, _ := arguments["name"].(string)
		if who == "" {
			who = "world"
		}
		return mcphelper.TextResult(fmt.Sprintf("Hello, %s!", who)), nil
	}
	return mcphelper.ErrorResult(fmt.Errorf("tool not found: %s", name)), nil
}
