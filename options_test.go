package claudecode

import (
	"reflect"
	"testing"
)

func TestNewOptions(t *testing.T) {
	opts := NewOptions()

	// Test defaults
	if len(opts.AllowedTools) != 0 {
		t.Errorf("Expected AllowedTools to be empty, got %v", opts.AllowedTools)
	}
	if opts.MaxThinkingTokens != 8000 {
		t.Errorf("Expected MaxThinkingTokens to be 8000, got %d", opts.MaxThinkingTokens)
	}
	if len(opts.McpTools) != 0 {
		t.Errorf("Expected McpTools to be empty, got %v", opts.McpTools)
	}
	if opts.McpServers == nil {
		t.Error("Expected McpServers to be initialized")
	}
	if len(opts.McpServers) != 0 {
		t.Errorf("Expected McpServers to be empty, got %v", opts.McpServers)
	}
	if opts.ContinueConversation != false {
		t.Errorf("Expected ContinueConversation to be false, got %v", opts.ContinueConversation)
	}
	if len(opts.DisallowedTools) != 0 {
		t.Errorf("Expected DisallowedTools to be empty, got %v", opts.DisallowedTools)
	}

	// Test that optional fields are nil
	if opts.SystemPrompt != nil {
		t.Error("Expected SystemPrompt to be nil")
	}
	if opts.PermissionMode != nil {
		t.Error("Expected PermissionMode to be nil")
	}
	if opts.MaxTurns != nil {
		t.Error("Expected MaxTurns to be nil")
	}
}

func TestOptionsBuilderMethods(t *testing.T) {
	opts := NewOptions().
		WithSystemPrompt("You are a helpful assistant").
		WithAllowedTools("Read", "Write").
		WithPermissionMode(PermissionModeAcceptEdits).
		WithMaxTurns(5).
		WithModel("claude-3-sonnet").
		WithCwd("/tmp").
		WithContinueConversation().
		WithResume("session_123")

	// Test system prompt
	if opts.SystemPrompt == nil || *opts.SystemPrompt != "You are a helpful assistant" {
		t.Error("SystemPrompt not set correctly")
	}

	// Test allowed tools
	expected := []string{"Read", "Write"}
	if !reflect.DeepEqual(opts.AllowedTools, expected) {
		t.Errorf("Expected AllowedTools to be %v, got %v", expected, opts.AllowedTools)
	}

	// Test permission mode
	if opts.PermissionMode == nil || *opts.PermissionMode != PermissionModeAcceptEdits {
		t.Error("PermissionMode not set correctly")
	}

	// Test max turns
	if opts.MaxTurns == nil || *opts.MaxTurns != 5 {
		t.Error("MaxTurns not set correctly")
	}

	// Test model
	if opts.Model == nil || *opts.Model != "claude-3-sonnet" {
		t.Error("Model not set correctly")
	}

	// Test cwd
	if opts.Cwd == nil || *opts.Cwd != "/tmp" {
		t.Error("Cwd not set correctly")
	}

	// Test continue conversation
	if !opts.ContinueConversation {
		t.Error("ContinueConversation not set correctly")
	}

	// Test resume
	if opts.Resume == nil || *opts.Resume != "session_123" {
		t.Error("Resume not set correctly")
	}
}

func TestMcpServerConfigs(t *testing.T) {
	// Test StdioServerConfig
	stdioConfig := &StdioServerConfig{
		Command: "python",
		Args:    []string{"-m", "my_mcp_server"},
		Env:     map[string]string{"DEBUG": "1"},
	}
	if stdioConfig.ServerType() != "stdio" {
		t.Errorf("Expected stdio server type, got %s", stdioConfig.ServerType())
	}

	// Test SSEServerConfig
	sseConfig := &SSEServerConfig{
		URL:     "https://example.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	if sseConfig.ServerType() != "sse" {
		t.Errorf("Expected sse server type, got %s", sseConfig.ServerType())
	}

	// Test HTTPServerConfig
	httpConfig := &HTTPServerConfig{
		URL:     "https://api.example.com/mcp",
		Headers: map[string]string{"X-API-Key": "key123"},
	}
	if httpConfig.ServerType() != "http" {
		t.Errorf("Expected http server type, got %s", httpConfig.ServerType())
	}

	// Test adding MCP servers to options
	opts := NewOptions().
		AddMcpServer("stdio_server", stdioConfig).
		AddMcpServer("sse_server", sseConfig).
		AddMcpTool("filesystem").
		AddMcpTool("web_search")

	if len(opts.McpServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(opts.McpServers))
	}

	if opts.McpServers["stdio_server"] != stdioConfig {
		t.Error("stdio_server not added correctly")
	}

	if opts.McpServers["sse_server"] != sseConfig {
		t.Error("sse_server not added correctly")
	}

	expectedTools := []string{"filesystem", "web_search"}
	if !reflect.DeepEqual(opts.McpTools, expectedTools) {
		t.Errorf("Expected McpTools to be %v, got %v", expectedTools, opts.McpTools)
	}
}
