package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/wcollani/mcp-go-template/internal/mcp"
	"github.com/wcollani/mcp-go-template/internal/provider/hello"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	s := mcp.NewServer("mcp-go-template", "0.1.0")

	// Register providers here.
	// Each provider implements the Provider interface (internal/provider/provider.go).
	// See internal/provider/hello/ for a minimal working example.
	s.AddProvider(&hello.Provider{})

	// Add your providers:
	// s.AddProvider(grafana.NewProvider(os.Getenv("GRAFANA_URL"), os.Getenv("GRAFANA_API_KEY")))
	// s.AddProvider(myservice.NewProvider(os.Getenv("MY_SERVICE_URL")))

	if err := s.Run(context.Background()); err != nil {
		slog.Error("Server exited with error", "error", err)
		os.Exit(1)
	}
}
