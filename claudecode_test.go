package claudecode

import (
	"context"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This will fail because there's no CLI, but we're testing the API structure
	stream, err := Query(ctx, "test prompt", nil)

	// We expect this to fail (CLI not found), but the error should be structured
	if err == nil {
		// If it somehow doesn't error, clean up the stream
		if stream != nil {
			stream.Close()
		}
		t.Skip("Unexpectedly succeeded - CLI might be available")
	}

	// Check that we get a proper error type
	if cliErr, ok := err.(*CLINotFoundError); ok {
		if cliErr.Message == "" {
			t.Error("Expected non-empty CLI error message")
		}
	} else if connErr, ok := err.(*ConnectionError); ok {
		// This is also acceptable - connection error due to missing CLI
		if connErr.Message == "" {
			t.Error("Expected non-empty connection error message")
		}
	} else {
		t.Errorf("Expected CLINotFoundError or ConnectionError, got %T: %v", err, err)
	}
}

func TestQueryWithCLIPath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	options := NewOptions().WithSystemPrompt("Test system prompt")

	// Test with fake CLI path
	stream, err := QueryWithCLIPath(ctx, "test prompt", options, "/fake/nonexistent/claude")

	// We expect this to fail (CLI not found), but the error should be structured
	if err == nil {
		// If it somehow doesn't error, clean up the stream
		if stream != nil {
			stream.Close()
		}
		t.Skip("Unexpectedly succeeded - CLI might exist at fake path")
	}

	// Check that we get a proper error type
	if cliErr, ok := err.(*CLINotFoundError); ok {
		if cliErr.CLIPath != "/fake/nonexistent/claude" {
			t.Errorf("Expected error to include CLI path, got: %v", err)
		}
	} else if connErr, ok := err.(*ConnectionError); ok {
		// This is also acceptable - connection error due to missing CLI
		if connErr.Message == "" {
			t.Error("Expected non-empty connection error message")
		}
	} else {
		t.Errorf("Expected CLINotFoundError or ConnectionError, got %T: %v", err, err)
	}
}

func TestSetParserBufferSize(t *testing.T) {
	// Test that SetParserBufferSize doesn't panic
	originalSize := 1024 * 1024 // Default size
	customSize := 2048

	// This should not panic
	SetParserBufferSize(customSize)

	// Reset to original for other tests
	SetParserBufferSize(originalSize)
}

func TestQueryStreamInterface(t *testing.T) {
	// Test that QueryStream satisfies the expected interface
	// Create a nil stream and test that the method signatures compile
	var stream *QueryStream

	// Check that the types are correct by assigning to variables of expected types
	var msgs <-chan Message
	var errs <-chan error
	var closed bool
	var closeErr error

	// Test method signature compatibility (without calling on nil pointer)
	if stream != nil {
		msgs = stream.Messages()
		errs = stream.Errors()
		closed = stream.IsClosed()
		closeErr = stream.Close()
	}

	// Suppress unused variable warnings
	_ = msgs
	_ = errs
	_ = closed
	_ = closeErr
}

func TestOptionsBuilder(t *testing.T) {
	// Test the fluent options builder API
	options := NewOptions().
		WithSystemPrompt("You are a helpful assistant").
		WithAllowedTools("Read", "Write", "Bash").
		WithPermissionMode(PermissionModeAcceptEdits).
		WithMaxTurns(5).
		WithModel("claude-3-sonnet").
		WithCwd("/tmp").
		WithContinueConversation().
		WithResume("session_123")

	// Verify the options were set
	if options.SystemPrompt == nil || *options.SystemPrompt != "You are a helpful assistant" {
		t.Error("SystemPrompt not set correctly")
	}

	if len(options.AllowedTools) != 3 {
		t.Errorf("Expected 3 allowed tools, got %d", len(options.AllowedTools))
	}

	if options.PermissionMode == nil || *options.PermissionMode != PermissionModeAcceptEdits {
		t.Error("PermissionMode not set correctly")
	}

	if options.MaxTurns == nil || *options.MaxTurns != 5 {
		t.Error("MaxTurns not set correctly")
	}

	if !options.ContinueConversation {
		t.Error("ContinueConversation not set correctly")
	}

	if options.Resume == nil || *options.Resume != "session_123" {
		t.Error("Resume not set correctly")
	}
}

func TestMcpServerConfiguration(t *testing.T) {
	// Test MCP server configuration
	stdioConfig := &StdioServerConfig{
		Command: "python",
		Args:    []string{"-m", "my_mcp_server"},
		Env:     map[string]string{"DEBUG": "1"},
	}

	sseConfig := &SSEServerConfig{
		URL:     "https://example.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	httpConfig := &HTTPServerConfig{
		URL:     "https://api.example.com/mcp",
		Headers: map[string]string{"X-API-Key": "key123"},
	}

	// Test server type methods
	if stdioConfig.ServerType() != "stdio" {
		t.Error("Expected stdio server type")
	}
	if sseConfig.ServerType() != "sse" {
		t.Error("Expected sse server type")
	}
	if httpConfig.ServerType() != "http" {
		t.Error("Expected http server type")
	}

	// Test adding to options
	options := NewOptions().
		AddMcpServer("stdio_server", stdioConfig).
		AddMcpServer("sse_server", sseConfig).
		AddMcpTool("filesystem").
		AddMcpTool("web_search")

	if len(options.McpServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(options.McpServers))
	}

	if len(options.McpTools) != 2 {
		t.Errorf("Expected 2 MCP tools, got %d", len(options.McpTools))
	}
}

func TestMessageTypes(t *testing.T) {
	// Test that all message types implement the Message interface
	var messages []Message

	userMsg := &UserMessage{Content: "Hello"}
	assistantMsg := &AssistantMessage{Content: []ContentBlock{}}
	systemMsg := &SystemMessage{Subtype: "status", Data: map[string]any{}}
	resultMsg := &ResultMessage{
		Subtype:       "completion",
		DurationMs:    1000,
		DurationAPIMs: 800,
		IsError:       false,
		NumTurns:      1,
		SessionID:     "session_123",
	}

	messages = append(messages, userMsg, assistantMsg, systemMsg, resultMsg)

	// Test type methods
	if userMsg.Type() != "user" {
		t.Error("Expected user message type")
	}
	if assistantMsg.Type() != "assistant" {
		t.Error("Expected assistant message type")
	}
	if systemMsg.Type() != "system" {
		t.Error("Expected system message type")
	}
	if resultMsg.Type() != "result" {
		t.Error("Expected result message type")
	}
}

func TestContentBlockTypes(t *testing.T) {
	// Test that all content block types implement the ContentBlock interface
	var blocks []ContentBlock

	textBlock := &TextBlock{Text: "Hello, world!"}
	toolUseBlock := &ToolUseBlock{
		ID:   "tool_123",
		Name: "Read",
		Input: map[string]any{
			"file_path": "/path/to/file.txt",
		},
	}
	toolResultBlock := &ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   stringPtr("File contents"),
		IsError:   boolPtr(false),
	}

	blocks = append(blocks, textBlock, toolUseBlock, toolResultBlock)

	// Test type methods
	if textBlock.Type() != "text" {
		t.Error("Expected text block type")
	}
	if toolUseBlock.Type() != "tool_use" {
		t.Error("Expected tool_use block type")
	}
	if toolResultBlock.Type() != "tool_result" {
		t.Error("Expected tool_result block type")
	}
}

func TestErrorTypes(t *testing.T) {
	// Test error type creation and methods
	cliErr := NewCLINotFoundError("CLI not found", "/fake/path")
	if cliErr.CLIPath != "/fake/path" {
		t.Error("CLI path not set correctly")
	}

	processErr := NewProcessError("Process failed", 1, "error output")
	if processErr.ExitCode != 1 {
		t.Error("Exit code not set correctly")
	}
	if processErr.Stderr != "error output" {
		t.Error("Stderr not set correctly")
	}

	jsonErr := NewJSONDecodeError("invalid json", nil)
	if jsonErr.Line != "invalid json" {
		t.Error("JSON line not set correctly")
	}

	connErr := NewConnectionError("Connection failed", nil)
	if connErr.Message != "Connection failed" {
		t.Error("Connection error message not set correctly")
	}
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
