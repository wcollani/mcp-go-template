package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewTool constructs an mcp.Tool with the given name, description, and a
// minimal JSON Schema object for the input. The inputSchema map should follow
// JSON Schema draft conventions — e.g.:
//
//	mcp.NewTool("search_items", "Search for items", map[string]interface{}{
//	    "query": map[string]interface{}{
//	        "type":        "string",
//	        "description": "The search term",
//	        "required":    true,
//	    },
//	})
//
// To define a tool with no parameters, pass nil or an empty map for inputSchema.
// Mark required fields with "required": true at the property level (non-standard
// shorthand) — NewTool promotes them to the top-level required array automatically.
func NewTool(name, description string, inputSchema map[string]interface{}) *mcp.Tool {
	// Shallow copy so required-key removal never mutates the caller's map.
	properties := make(map[string]interface{}, len(inputSchema))
	for k, v := range inputSchema {
		properties[k] = v
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	var required []string
	for k, v := range properties {
		if propMap, ok := v.(map[string]interface{}); ok {
			if req, ok := propMap["required"].(bool); ok && req {
				required = append(required, k)
				// Copy before mutating so we don't modify the caller's map.
				propCopy := make(map[string]interface{}, len(propMap))
				for pk, pv := range propMap {
					if pk != "required" {
						propCopy[pk] = pv
					}
				}
				properties[k] = propCopy
			}
		}
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return &mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}
}

// TextResult returns a successful *mcp.CallToolResult with a single TextContent block.
func TextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// ErrorResult returns a *mcp.CallToolResult signaling a tool-level error.
// Per the MCP spec, tool errors should be returned inside CallToolResult —
// not as protocol-level errors — so the LLM can see the failure and self-correct.
func ErrorResult(err error) *mcp.CallToolResult {
	result := &mcp.CallToolResult{}
	result.SetError(err)
	return result
}
