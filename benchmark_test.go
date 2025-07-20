package claudecode

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkOptionsBuilder benchmarks the options builder pattern
func BenchmarkOptionsBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOptions().
			WithSystemPrompt("You are a helpful assistant").
			WithAllowedTools("Read", "Write", "Bash").
			WithDisallowedTools("Execute").
			WithPermissionMode(PermissionModeAcceptEdits).
			WithMaxTurns(5).
			WithModel("claude-3-sonnet").
			WithCwd("/tmp").
			WithContinueConversation().
			WithResume("session_123").
			AddMcpServer("stdio", &StdioServerConfig{
				Command: "python",
				Args:    []string{"-m", "server"},
			}).
			AddMcpTool("filesystem")
	}
}

// BenchmarkOptionsBuilderSimple benchmarks simple options creation
func BenchmarkOptionsBuilderSimple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOptions().WithSystemPrompt("Test prompt")
	}
}

// BenchmarkNewOptions benchmarks options creation
func BenchmarkNewOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOptions()
	}
}

// BenchmarkMessageTypeChecking benchmarks message type checking
func BenchmarkMessageTypeChecking(b *testing.B) {
	messages := []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{}},
		&SystemMessage{Subtype: "test"},
		&ResultMessage{Subtype: "test"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range messages {
			_ = msg.Type()
		}
	}
}

// BenchmarkContentBlockTypeChecking benchmarks content block type checking
func BenchmarkContentBlockTypeChecking(b *testing.B) {
	blocks := []ContentBlock{
		&TextBlock{Text: "test content"},
		&ToolUseBlock{ID: "1", Name: "test", Input: map[string]any{"key": "value"}},
		&ToolResultBlock{ToolUseID: "1", Content: stringPtr("result")},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, block := range blocks {
			_ = block.Type()
		}
	}
}

// BenchmarkErrorCreation benchmarks error creation
func BenchmarkErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCLINotFoundError("CLI not found", "/path/to/cli")
		_ = NewConnectionError("Connection failed", nil)
		_ = NewProcessError("Process failed", 1, "stderr")
		_ = NewJSONDecodeError("invalid json", nil)
	}
}

// BenchmarkSetParserBufferSize benchmarks parser buffer size setting
func BenchmarkSetParserBufferSize(b *testing.B) {
	sizes := []int{1024, 2048, 4096, 8192, 16384}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetParserBufferSize(sizes[i%len(sizes)])
	}
}

// BenchmarkQueryStreamWrapper benchmarks stream wrapper creation
func BenchmarkQueryStreamWrapper(b *testing.B) {
	// Create a nil internal stream for benchmarking wrapper overhead
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapQueryStream(nil)
	}
}

// BenchmarkMcpServerConfiguration benchmarks MCP server configuration
func BenchmarkMcpServerConfiguration(b *testing.B) {
	stdioConfig := &StdioServerConfig{
		Command: "python",
		Args:    []string{"-m", "server"},
		Env:     map[string]string{"DEBUG": "1"},
	}
	
	sseConfig := &SSEServerConfig{
		URL:     "https://example.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := NewOptions()
		opts.AddMcpServer("stdio", stdioConfig)
		opts.AddMcpServer("sse", sseConfig)
		opts.AddMcpTool("tool1")
		opts.AddMcpTool("tool2")
		
		// Check server types
		_ = stdioConfig.ServerType()
		_ = sseConfig.ServerType()
	}
}

// BenchmarkComplexOptionsCreation benchmarks creating complex options with all features
func BenchmarkComplexOptionsCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOptions().
			WithSystemPrompt("You are a helpful coding assistant").
			WithAppendSystemPrompt("Please be concise and accurate").
			WithAllowedTools("Read", "Write", "Bash", "Search", "Edit").
			WithDisallowedTools("Execute", "Delete").
			WithPermissionMode(PermissionModeBypassPermissions).
			WithMaxTurns(10).
			WithModel("claude-3-opus-20240229").
			WithCwd("/Users/developer/project").
			WithContinueConversation().
			WithResume("session_abcd1234").
			AddMcpServer("filesystem", &StdioServerConfig{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				Env:     map[string]string{"NODE_ENV": "production", "DEBUG": "mcp:*"},
			}).
			AddMcpServer("web_search", &SSEServerConfig{
				URL: "https://api.search.example.com/mcp",
				Headers: map[string]string{
					"Authorization": "Bearer sk-1234567890abcdef",
					"User-Agent":    "claude-code-sdk-go/1.0",
					"Accept":        "application/json",
				},
			}).
			AddMcpServer("database", &HTTPServerConfig{
				URL: "https://db.example.com/mcp/v1",
				Headers: map[string]string{
					"Authorization": "Bearer db-token-xyz",
					"Content-Type":  "application/json",
				},
			}).
			AddMcpTool("read_file").
			AddMcpTool("write_file").
			AddMcpTool("search_web").
			AddMcpTool("query_database").
			AddMcpTool("execute_sql")
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Test typical SDK usage patterns that allocate memory
		opts := NewOptions()
		
		// String pointer allocations
		opts.WithSystemPrompt("Test prompt")
		opts.WithModel("claude-3-sonnet")
		opts.WithCwd("/tmp")
		opts.WithResume("session_123")
		
		// Slice allocations
		opts.WithAllowedTools("tool1", "tool2", "tool3")
		opts.WithDisallowedTools("bad_tool")
		
		// Map allocations
		opts.AddMcpServer("server", &StdioServerConfig{
			Command: "test",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY": "value"},
		})
		
		// Message creation
		_ = &UserMessage{Content: "test"}
		_ = &AssistantMessage{Content: []ContentBlock{
			&TextBlock{Text: "response"},
			&ToolUseBlock{
				ID:   "tool_1",
				Name: "test_tool",
				Input: map[string]any{
					"param1": "value1",
					"param2": 42,
				},
			},
		}}
		
		// Error creation
		_ = NewCLINotFoundError("test error", "/path")
	}
}

// BenchmarkConcurrentOptionsCreation benchmarks concurrent options creation
func BenchmarkConcurrentOptionsCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewOptions().
				WithSystemPrompt("Concurrent test").
				WithAllowedTools("Read", "Write").
				WithMaxTurns(5)
		}
	})
}

// BenchmarkParserBufferSizeUpdates benchmarks frequent parser buffer size updates
func BenchmarkParserBufferSizeUpdates(b *testing.B) {
	sizes := []int{1024, 2048, 4096, 8192, 16384, 32768, 65536}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			SetParserBufferSize(size)
		}
	}
}

// BenchmarkQueryContextCreation benchmarks context creation patterns typical for Query calls
func BenchmarkQueryContextCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cancel() // Clean up immediately for benchmark
		_ = ctx
	}
}

// BenchmarkLargeMessageCreation benchmarks creating large messages
func BenchmarkLargeMessageCreation(b *testing.B) {
	// Create large content for realistic benchmarking
	largeText := ""
	for i := 0; i < 1000; i++ {
		largeText += "This is a large message content for benchmarking purposes. "
	}
	
	largeInput := make(map[string]any)
	for i := 0; i < 100; i++ {
		largeInput[fmt.Sprintf("param_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{Text: largeText},
				&ToolUseBlock{
					ID:    "large_tool",
					Name:  "process_data",
					Input: largeInput,
				},
			},
		}
	}
}

