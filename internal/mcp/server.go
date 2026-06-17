package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wcollani/mcp-go-template/internal/provider"
)

type Server struct {
	mcpServer *mcp.Server
	providers []provider.Provider
}

func NewServer(name, version string) *Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    name,
			Version: version,
		},
		nil,
	)

	return &Server{
		mcpServer: s,
	}
}

// AddProvider registers all resources, resource templates, prompts, and tools
// from the given Provider with the underlying MCP server.
func (s *Server) AddProvider(p provider.Provider) {
	s.providers = append(s.providers, p)

	resources, err := p.GetResources()
	if err != nil {
		slog.Error("Failed to get resources from provider", "provider", p.Name(), "error", err)
		return
	}

	for _, res := range resources {
		r := res
		s.mcpServer.AddResource(&r, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			content, err := p.GetResourceContent(req.Params.URI)
			if err != nil {
				return nil, err
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      req.Params.URI,
						MIMEType: r.MIMEType,
						Text:     content,
					},
				},
			}, nil
		})
	}

	templates, err := p.GetResourceTemplates()
	if err == nil {
		for _, tmpl := range templates {
			t := tmpl
			s.mcpServer.AddResourceTemplate(&t, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				content, err := p.GetResourceContent(req.Params.URI)
				if err != nil {
					return nil, err
				}
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      req.Params.URI,
							MIMEType: t.MIMEType,
							Text:     content,
						},
					},
				}, nil
			})
		}
	}

	prompts, err := p.GetPrompts()
	if err == nil {
		for _, prompt := range prompts {
			pr := prompt
			s.mcpServer.AddPrompt(&pr, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return p.GetPrompt(req.Params.Name, req.Params.Arguments)
			})
		}
	}

	tools, err := p.GetTools()
	if err != nil {
		slog.Error("Failed to get tools from provider", "provider", p.Name(), "error", err)
		return
	}

	for _, tool := range tools {
		t := tool
		s.mcpServer.AddTool(&t, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args map[string]interface{}
			if req.Params.Arguments != nil {
				if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
					result := &mcp.CallToolResult{}
					result.SetError(err)
					return result, nil
				}
			}
			return p.CallTool(req.Params.Name, args)
		})
	}
}

// Providers returns the registered providers (useful for testing).
func (s *Server) Providers() []provider.Provider {
	return s.providers
}

// Run starts the MCP server. Transport is selected via the MCP_TRANSPORT env var:
//   - "sse" — Streamable HTTP on PORT (default 8080)
//   - anything else — stdio (default, works with Claude Desktop and most MCP clients)
func (s *Server) Run(ctx context.Context) error {
	transport := os.Getenv("MCP_TRANSPORT")
	if transport == "sse" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		mcpHandler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
			return s.mcpServer
		}, &mcp.SSEOptions{})

		bearerToken := os.Getenv("MCP_BEARER_TOKEN")
		// Restrict CORS to a specific origin in production via CORS_ALLOW_ORIGIN env var.
		corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "http://localhost:3000"
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bearerToken != "" && r.Method != "OPTIONS" {
				if r.Header.Get("Authorization") != "Bearer "+bearerToken {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id, Mcp-Protocol-Version")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			mcpHandler.ServeHTTP(w, r)
		})

		slog.Info("Starting MCP server via SSE", "port", port)
		return http.ListenAndServe(":"+port, handler)
	}

	slog.Info("Starting MCP server via stdio")
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}
