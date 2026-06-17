package hello

import (
	"context"
	"testing"
)

func TestProvider_Name(t *testing.T) {
	p := &Provider{}
	if p.Name() != "hello" {
		t.Errorf("expected name 'hello', got %s", p.Name())
	}
}

func TestProvider_GetResources(t *testing.T) {
	p := &Provider{}
	resources, err := p.GetResources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].URI != "hello://world" {
		t.Errorf("expected URI 'hello://world', got %s", resources[0].URI)
	}
}

func TestProvider_GetResourceContent(t *testing.T) {
	p := &Provider{}
	content, err := p.GetResourceContent(context.Background(), "hello://world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty content")
	}

	_, err = p.GetResourceContent(context.Background(), "unknown://resource")
	if err == nil {
		t.Error("expected error for unknown URI, got nil")
	}
}

func TestProvider_CallTool_Greet(t *testing.T) {
	p := &Provider{}

	result, err := p.CallTool(context.Background(), "greet", map[string]interface{}{"name": "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result, got error")
	}

	result, err = p.CallTool(context.Background(), "greet", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result with default name, got error")
	}
}

func TestProvider_CallTool_Unknown(t *testing.T) {
	p := &Provider{}
	result, err := p.CallTool(context.Background(), "does_not_exist", nil)
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unknown tool")
	}
}
