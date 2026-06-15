# mcp-go-template

A Go template for MCP (Model Context Protocol) servers. Clone this to skip the boilerplate and start writing integrations immediately.

Built from the patterns I extracted after writing 18 MCP providers in [homelab-mcp](https://github.com/wcollani/homelab-mcp).

## What's included

- **`Provider` interface** — one interface for resources, tools, and prompts
- **`toolbuilder` helpers** — `NewTool`, `TextResult`, `ErrorResult` for ergonomic tool construction
- **Transport selection** — stdio (Claude Desktop, most clients) or SSE via `MCP_TRANSPORT=sse`
- **Working `hello` provider** — a minimal example with a resource and a tool, ready to copy
- **Multi-stage Dockerfile** — scratch image, ~10MB binary
- **GitHub Actions CI** — lint + test on every push

## Quick start

```bash
git clone https://github.com/wcollani/mcp-go-template
cd mcp-go-template
go run ./cmd/server        # stdio mode
# or
MCP_TRANSPORT=sse go run ./cmd/server  # SSE mode on :8080
```

## Adding a provider

### 1. Create the package

Copy `internal/provider/hello/` as your starting point:

```bash
cp -r internal/provider/hello internal/provider/myservice
```

### 2. Implement the interface

```go
// internal/provider/myservice/myservice.go
package myservice

import (
    "fmt"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    mcphelper "github.com/wcollani/mcp-go-template/internal/mcp"
)

type Provider struct {
    baseURL string
    apiKey  string
}

func NewProvider(baseURL, apiKey string) *Provider {
    return &Provider{baseURL: baseURL, apiKey: apiKey}
}

func (p *Provider) Name() string { return "myservice" }

func (p *Provider) GetResources() ([]mcp.Resource, error) {
    return []mcp.Resource{
        {
            URI:         "myservice://status",
            Name:        "My Service Status",
            Description: "Current status of My Service",
            MIMEType:    "application/json",
        },
    }, nil
}

func (p *Provider) GetResourceContent(uri string) (string, error) {
    if uri == "myservice://status" {
        // fetch from p.baseURL and return JSON string
        return `{"status": "ok"}`, nil
    }
    return "", fmt.Errorf("resource not found: %s", uri)
}

func (p *Provider) GetResourceTemplates() ([]mcp.ResourceTemplate, error) {
    return []mcp.ResourceTemplate{}, nil
}

func (p *Provider) GetPrompts() ([]mcp.Prompt, error) { return []mcp.Prompt{}, nil }
func (p *Provider) GetPrompt(name string, _ map[string]string) (*mcp.GetPromptResult, error) {
    return nil, fmt.Errorf("prompt not found: %s", name)
}

func (p *Provider) GetTools() ([]mcp.Tool, error) {
    return []mcp.Tool{
        *mcphelper.NewTool(
            "restart_myservice",
            "Restart the My Service process.",
            nil, // no parameters
        ),
        *mcphelper.NewTool(
            "search_myservice",
            "Search My Service by query string.",
            map[string]interface{}{
                "query": map[string]interface{}{
                    "type":        "string",
                    "description": "Search term",
                    "required":    true,
                },
            },
        ),
    }, nil
}

func (p *Provider) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
    switch name {
    case "restart_myservice":
        // call p.baseURL/restart
        return mcphelper.TextResult("My Service restarted successfully."), nil
    case "search_myservice":
        query, _ := args["query"].(string)
        // call p.baseURL/search?q=query
        return mcphelper.TextResult(fmt.Sprintf("Results for: %s", query)), nil
    default:
        return mcphelper.ErrorResult(fmt.Errorf("tool not found: %s", name)), nil
    }
}
```

### 3. Register it in main.go

```go
import "github.com/wcollani/mcp-go-template/internal/provider/myservice"

s.AddProvider(myservice.NewProvider(
    os.Getenv("MY_SERVICE_URL"),
    os.Getenv("MY_SERVICE_API_KEY"),
))
```

### 4. Test it

```bash
go test ./...
```

## Tool construction

`NewTool` handles JSON Schema generation from a flat map. Mark required parameters with `"required": true` at the property level — the helper promotes them to the top-level `required` array:

```go
mcphelper.NewTool(
    "tool_name",
    "Human-readable description for the LLM.",
    map[string]interface{}{
        "required_param": map[string]interface{}{
            "type":        "string",
            "description": "This field is required",
            "required":    true,   // promoted to top-level required[]
        },
        "optional_param": map[string]interface{}{
            "type":        "integer",
            "description": "This field is optional",
        },
    },
)
```

Return helpers:

```go
mcphelper.TextResult("success message")    // IsError = false
mcphelper.ErrorResult(err)                 // IsError = true, LLM sees the error and can retry
```

**Always use `ErrorResult` for tool-level errors** — not `return nil, err`. Protocol-level errors terminate the session; tool-level errors let the LLM self-correct.

## Transport

| `MCP_TRANSPORT` | Behavior |
|---|---|
| unset / any other value | stdio — works with Claude Desktop, MCP Inspector, most clients |
| `sse` | Streamable HTTP on `PORT` (default `8080`) — works with remote agents |

```bash
# stdio (default)
go run ./cmd/server

# SSE
MCP_TRANSPORT=sse PORT=9000 go run ./cmd/server
```

## Docker

```bash
docker build -t mcp-go-template .
docker run --env-file .env mcp-go-template
```

Copy `.env.example` to `.env` and fill in your service URLs.

## Resources vs Tools

The `Provider` interface draws a clear line:

- **Resources** (`GetResources`, `GetResourceContent`) — read-only state snapshots. Use for dashboards, status pages, config data. The LLM can read these without taking an action.
- **Tools** (`GetTools`, `CallTool`) — actions and mutations. Use for API calls that change state or trigger side effects.

This distinction matters: MCP clients can present resources differently from tools, and some clients restrict tool execution while allowing resource reads.
