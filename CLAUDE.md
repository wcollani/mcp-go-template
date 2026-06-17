# mcp-go-template Coding Guidelines (CLAUDE.md)

A Go template for MCP servers. Fork or clone this to skip boilerplate when building a new MCP integration.

## Build and Run

```bash
go build ./cmd/server          # build
go run ./cmd/server            # stdio mode (Claude Desktop, most clients)
MCP_TRANSPORT=sse go run ./cmd/server   # SSE mode on :8080
```

## Testing and Linting

```bash
go test -v -race -cover ./...
golangci-lint run
```

CI runs both on every push and PR via GitHub-hosted runners. No self-hosted runner needed — this is a public template repo.

## Adding a Provider

1. Copy `internal/provider/hello/` as your starting point
2. Implement the `Provider` interface (`Name()`, `GetResources()`, `GetTools()`)
3. Register it in `cmd/server/main.go` alongside the existing providers
4. Add a test file following the `hello_test.go` pattern

## Coding Rules

- **toolbuilder only:** Use `internal/mcp/toolbuilder.go` (`NewTool`, `TextResult`, `ErrorResult`) for all tool construction — do not hand-roll `mcp.Tool` structs
- **No stdout under stdio transport:** All logs go to `os.Stderr` via `slog`. Writing to `os.Stdout` corrupts the JSON-RPC stream
- **Fail-fast timeouts:** Wrap all external calls in `context.WithTimeout` (10s default)
- **Graceful degradation:** Provider init failures must log and continue — do not crash the server

## Releasing

This is a template, not a versioned library. Users fork it; there's no release cycle. Keep the `main` branch in a buildable, documented state at all times.
