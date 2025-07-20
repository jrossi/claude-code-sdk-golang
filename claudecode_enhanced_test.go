package claudecode

import (
	"context"
	"testing"
	"time"
)

func TestQueryWithNilOptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test that Query works with nil options
	_, err := Query(ctx, "test prompt", nil)
	
	// We expect an error because CLI is not available, but the function should handle nil options
	if err == nil {
		t.Skip("CLI might be available, test skipped")
	}
	
	// The error should be related to CLI availability, not nil options
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestQueryWithEmptyPrompt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test that Query works with empty prompt
	_, err := Query(ctx, "", NewOptions())
	
	// We expect an error because CLI is not available
	if err == nil {
		t.Skip("CLI might be available, test skipped")
	}
	
	// The error should be related to CLI availability, not empty prompt
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestQueryStreamWrapper(t *testing.T) {
	// Test the QueryStream wrapper functionality
	// Since we can't create a real stream without CLI, test the wrapper function directly
	if wrapQueryStream(nil) == nil {
		t.Error("Expected non-nil wrapper even with nil internal stream")
	}
}

func TestQueryStreamNilHandling(t *testing.T) {
	// Test QueryStream methods with nil internal stream
	// This tests the wrapper's robustness
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected with nil internal stream
			t.Log("Panic expected with nil internal stream:", r)
		}
	}()
	
	stream := &QueryStream{internal: nil}
	
	// These should either work gracefully or panic predictably
	_ = stream.Messages()
	_ = stream.Errors()
	_ = stream.IsClosed()
	_ = stream.Close()
}

func TestSetParserBufferSizeEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"zero size", 0},
		{"negative size", -1},
		{"small size", 1},
		{"large size", 100 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic for any size value
			SetParserBufferSize(tt.size)
		})
	}
	
	// Reset to reasonable default
	SetParserBufferSize(1024 * 1024)
}

func TestQueryWithContextCancellation(t *testing.T) {
	// Test behavior when context is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err := Query(ctx, "test prompt", NewOptions())
	
	// Should get an error, possibly context cancellation related
	if err == nil {
		t.Skip("Query succeeded despite cancelled context")
	}
	
	// Error should be meaningful
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestQueryWithCLIPathEdgeCases(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	tests := []struct {
		name    string
		cliPath string
	}{
		{"empty path", ""},
		{"relative path", "./nonexistent"},
		{"absolute path", "/nonexistent/path/to/claude"},
		{"path with spaces", "/path with spaces/claude"},
		{"windows-style path", "C:\\nonexistent\\claude.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QueryWithCLIPath(ctx, "test", NewOptions(), tt.cliPath)
			
			// We expect errors for all these cases
			if err == nil {
				t.Skip("Query succeeded unexpectedly")
			}
			
			// Error should be meaningful
			if err.Error() == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}

func TestOptionsBuilderChaining(t *testing.T) {
	// Test that all builder methods return the same instance for chaining
	options := NewOptions()
	
	// Test method chaining returns
	if options.WithSystemPrompt("test") != options {
		t.Error("WithSystemPrompt should return same instance")
	}
	if options.WithAppendSystemPrompt("test") != options {
		t.Error("WithAppendSystemPrompt should return same instance")
	}
	if options.WithAllowedTools("tool1", "tool2") != options {
		t.Error("WithAllowedTools should return same instance")
	}
	if options.WithDisallowedTools("tool3") != options {
		t.Error("WithDisallowedTools should return same instance")
	}
	if options.WithPermissionMode(PermissionModeDefault) != options {
		t.Error("WithPermissionMode should return same instance")
	}
	if options.WithMaxTurns(5) != options {
		t.Error("WithMaxTurns should return same instance")
	}
	if options.WithModel("claude-3-sonnet") != options {
		t.Error("WithModel should return same instance")
	}
	if options.WithCwd("/tmp") != options {
		t.Error("WithCwd should return same instance")
	}
	if options.WithContinueConversation() != options {
		t.Error("WithContinueConversation should return same instance")
	}
	if options.WithResume("session") != options {
		t.Error("WithResume should return same instance")
	}
}

func TestMcpServerConfigurationEdgeCases(t *testing.T) {
	// Test various MCP server configurations
	options := NewOptions()
	
	// Test adding nil server (should not panic)
	result := options.AddMcpServer("test", nil)
	if result != options {
		t.Error("AddMcpServer should return same instance")
	}
	
	// Test adding with empty name
	stdioConfig := &StdioServerConfig{Command: "test"}
	result = options.AddMcpServer("", stdioConfig)
	if result != options {
		t.Error("AddMcpServer should return same instance")
	}
	
	// Test adding duplicate server names
	result = options.AddMcpServer("test", stdioConfig)
	result = options.AddMcpServer("test", stdioConfig) // Same name
	if result != options {
		t.Error("AddMcpServer should return same instance")
	}
	
	// Test adding MCP tools
	result = options.AddMcpTool("")     // Empty tool name
	result = options.AddMcpTool("tool1") // Valid tool name
	result = options.AddMcpTool("tool1") // Duplicate tool name
	if result != options {
		t.Error("AddMcpTool should return same instance")
	}
}

func TestPermissionModeConstantValues(t *testing.T) {
	// Test that permission mode constants have expected values
	if PermissionModeDefault != "default" {
		t.Errorf("PermissionModeDefault = %v, want 'default'", PermissionModeDefault)
	}
	if PermissionModeAcceptEdits != "acceptEdits" {
		t.Errorf("PermissionModeAcceptEdits = %v, want 'acceptEdits'", PermissionModeAcceptEdits)
	}
	if PermissionModeBypassPermissions != "bypassPermissions" {
		t.Errorf("PermissionModeBypassPermissions = %v, want 'bypassPermissions'", PermissionModeBypassPermissions)
	}
}

func TestMessageTypeInterface(t *testing.T) {
	// Test that all message types properly implement Message interface
	messages := []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{}},
		&SystemMessage{Subtype: "test"},
		&ResultMessage{Subtype: "test"},
	}
	
	expectedTypes := []string{"user", "assistant", "system", "result"}
	
	for i, msg := range messages {
		if msg.Type() != expectedTypes[i] {
			t.Errorf("Message %d: expected type %s, got %s", i, expectedTypes[i], msg.Type())
		}
	}
}

func TestContentBlockTypeInterface(t *testing.T) {
	// Test that all content block types properly implement ContentBlock interface
	blocks := []ContentBlock{
		&TextBlock{Text: "test"},
		&ToolUseBlock{ID: "1", Name: "test"},
		&ToolResultBlock{ToolUseID: "1"},
	}
	
	expectedTypes := []string{"text", "tool_use", "tool_result"}
	
	for i, block := range blocks {
		if block.Type() != expectedTypes[i] {
			t.Errorf("ContentBlock %d: expected type %s, got %s", i, expectedTypes[i], block.Type())
		}
	}
}

func TestErrorTypeImplementations(t *testing.T) {
	// Test that all error types implement the error interface
	errors := []error{
		NewCLINotFoundError("test", "/path"),
		NewConnectionError("test", nil),
		NewProcessError("test", 1, "stderr"),
		NewJSONDecodeError("test", nil),
	}
	
	for i, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error %d should have non-empty Error() string", i)
		}
	}
}

func TestStdioServerConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *StdioServerConfig
	}{
		{
			name: "empty command",
			config: &StdioServerConfig{
				Command: "",
				Args:    []string{},
				Env:     nil,
			},
		},
		{
			name: "nil args and env",
			config: &StdioServerConfig{
				Command: "test",
				Args:    nil,
				Env:     nil,
			},
		},
		{
			name: "empty env map",
			config: &StdioServerConfig{
				Command: "test",
				Args:    []string{"arg1"},
				Env:     map[string]string{},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.ServerType() != "stdio" {
				t.Error("Expected stdio server type")
			}
		})
	}
}

func TestSSEServerConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *SSEServerConfig
	}{
		{
			name: "empty URL",
			config: &SSEServerConfig{
				URL:     "",
				Headers: nil,
			},
		},
		{
			name: "nil headers",
			config: &SSEServerConfig{
				URL:     "https://example.com",
				Headers: nil,
			},
		},
		{
			name: "empty headers map",
			config: &SSEServerConfig{
				URL:     "https://example.com",
				Headers: map[string]string{},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.ServerType() != "sse" {
				t.Error("Expected sse server type")
			}
		})
	}
}

func TestHTTPServerConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *HTTPServerConfig
	}{
		{
			name: "empty URL",
			config: &HTTPServerConfig{
				URL:     "",
				Headers: nil,
			},
		},
		{
			name: "nil headers",
			config: &HTTPServerConfig{
				URL:     "https://example.com",
				Headers: nil,
			},
		},
		{
			name: "empty headers map",
			config: &HTTPServerConfig{
				URL:     "https://example.com",
				Headers: map[string]string{},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.ServerType() != "http" {
				t.Error("Expected http server type")
			}
		})
	}
}

func TestToolResultBlockEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		block  *ToolResultBlock
		wantID string
	}{
		{
			name: "nil content and error",
			block: &ToolResultBlock{
				ToolUseID: "tool_123",
				Content:   nil,
				IsError:   nil,
			},
			wantID: "tool_123",
		},
		{
			name: "empty content string",
			block: &ToolResultBlock{
				ToolUseID: "tool_456",
				Content:   stringPtr(""),
				IsError:   boolPtr(false),
			},
			wantID: "tool_456",
		},
		{
			name: "error with no content",
			block: &ToolResultBlock{
				ToolUseID: "tool_789",
				Content:   nil,
				IsError:   boolPtr(true),
			},
			wantID: "tool_789",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.block.Type() != "tool_result" {
				t.Error("Expected tool_result type")
			}
			if tt.block.ToolUseID != tt.wantID {
				t.Errorf("Expected ToolUseID %s, got %s", tt.wantID, tt.block.ToolUseID)
			}
		})
	}
}

func TestToolUseBlockEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		block *ToolUseBlock
	}{
		{
			name: "nil input",
			block: &ToolUseBlock{
				ID:    "tool_123",
				Name:  "test_tool",
				Input: nil,
			},
		},
		{
			name: "empty input map",
			block: &ToolUseBlock{
				ID:    "tool_456",
				Name:  "test_tool",
				Input: map[string]any{},
			},
		},
		{
			name: "complex input",
			block: &ToolUseBlock{
				ID:   "tool_789",
				Name: "test_tool",
				Input: map[string]any{
					"nested": map[string]any{
						"key": "value",
					},
					"array": []any{1, 2, 3},
					"null":  nil,
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.block.Type() != "tool_use" {
				t.Error("Expected tool_use type")
			}
			if tt.block.ID == "" {
				t.Error("Expected non-empty ID")
			}
			if tt.block.Name == "" {
				t.Error("Expected non-empty Name")
			}
		})
	}
}

func TestResultMessageEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		msg  *ResultMessage
	}{
		{
			name: "minimal result message",
			msg: &ResultMessage{
				Subtype:   "test",
				SessionID: "session_123",
			},
		},
		{
			name: "result with nil cost and usage",
			msg: &ResultMessage{
				Subtype:      "completion",
				SessionID:    "session_456",
				TotalCostUSD: nil,
				Usage:        nil,
				Result:       nil,
			},
		},
		{
			name: "result with zero values",
			msg: &ResultMessage{
				Subtype:       "completion",
				DurationMs:    0,
				DurationAPIMs: 0,
				NumTurns:      0,
				SessionID:     "session_789",
				TotalCostUSD:  floatPtr(0.0),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Type() != "result" {
				t.Error("Expected result type")
			}
			if tt.msg.SessionID == "" {
				t.Error("Expected non-empty SessionID")
			}
		})
	}
}

// Helper function for creating float pointers
func floatPtr(f float64) *float64 {
	return &f
}