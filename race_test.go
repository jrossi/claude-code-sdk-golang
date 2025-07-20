package claudecode

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentOptionsBuilder tests concurrent access to options builder
func TestConcurrentOptionsBuilder(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				opts := NewOptions()
				opts.WithSystemPrompt("Test prompt")
				opts.WithAllowedTools("tool1", "tool2")
				opts.WithMaxTurns(5)
				opts.WithModel("claude-3-sonnet")
				opts.AddMcpTool("mcp_tool")
				
				// Verify some basic properties
				if len(opts.AllowedTools) != 2 {
					t.Errorf("Goroutine %d: expected 2 allowed tools, got %d", id, len(opts.AllowedTools))
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentSetParserBufferSize tests concurrent calls to SetParserBufferSize
func TestConcurrentSetParserBufferSize(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 100
	
	sizes := []int{1024, 2048, 4096, 8192, 16384}
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				size := sizes[(id*numOperations+j)%len(sizes)]
				SetParserBufferSize(size)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Reset to default
	SetParserBufferSize(1024 * 1024)
}

// TestConcurrentErrorCreation tests concurrent error creation
func TestConcurrentErrorCreation(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 100
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Create different types of errors concurrently
				_ = NewCLINotFoundError("CLI not found", "/path/to/cli")
				_ = NewConnectionError("Connection failed", nil)
				_ = NewProcessError("Process failed", 1, "stderr output")
				_ = NewJSONDecodeError("invalid json line", nil)
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentQueryCalls tests concurrent Query calls (will fail gracefully)
func TestConcurrentQueryCalls(t *testing.T) {
	const numGoroutines = 10
	const numOperations = 5
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				
				// These will likely fail due to CLI not being available, but should not race
				stream, err := Query(ctx, "test prompt", NewOptions())
				if err == nil && stream != nil {
					stream.Close()
				}
				
				cancel()
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentMessageCreation tests concurrent message and content block creation
func TestConcurrentMessageCreation(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Create various message types
				userMsg := &UserMessage{Content: "test user message"}
				
				assistantMsg := &AssistantMessage{
					Content: []ContentBlock{
						&TextBlock{Text: "test response"},
						&ToolUseBlock{
							ID:   "tool_1",
							Name: "test_tool",
							Input: map[string]any{
								"param": "value",
							},
						},
						&ToolResultBlock{
							ToolUseID: "tool_1",
							Content:   stringPtr("tool result"),
							IsError:   boolPtr(false),
						},
					},
				}
				
				systemMsg := &SystemMessage{
					Subtype: "status",
					Data: map[string]any{
						"status": "active",
						"id":     id,
					},
				}
				
				resultMsg := &ResultMessage{
					Subtype:       "completion",
					DurationMs:    1000,
					DurationAPIMs: 800,
					IsError:       false,
					NumTurns:      1,
					SessionID:     "session_123",
				}
				
				// Test type methods concurrently
				if userMsg.Type() != "user" {
					t.Errorf("Goroutine %d: expected user type", id)
				}
				if assistantMsg.Type() != "assistant" {
					t.Errorf("Goroutine %d: expected assistant type", id)
				}
				if systemMsg.Type() != "system" {
					t.Errorf("Goroutine %d: expected system type", id)
				}
				if resultMsg.Type() != "result" {
					t.Errorf("Goroutine %d: expected result type", id)
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentMcpServerConfiguration tests concurrent MCP server configuration
func TestConcurrentMcpServerConfiguration(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 100
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				opts := NewOptions()
				
				// Add multiple server types concurrently
				stdioConfig := &StdioServerConfig{
					Command: "python",
					Args:    []string{"-m", "server"},
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
				
				opts.AddMcpServer("stdio", stdioConfig)
				opts.AddMcpServer("sse", sseConfig)
				opts.AddMcpServer("http", httpConfig)
				
				opts.AddMcpTool("tool1")
				opts.AddMcpTool("tool2")
				opts.AddMcpTool("tool3")
				
				// Verify server types
				if stdioConfig.ServerType() != "stdio" {
					t.Errorf("Goroutine %d: expected stdio server type", id)
				}
				if sseConfig.ServerType() != "sse" {
					t.Errorf("Goroutine %d: expected sse server type", id)
				}
				if httpConfig.ServerType() != "http" {
					t.Errorf("Goroutine %d: expected http server type", id)
				}
				
				// Verify configuration
				if len(opts.McpServers) != 3 {
					t.Errorf("Goroutine %d: expected 3 MCP servers, got %d", id, len(opts.McpServers))
				}
				if len(opts.McpTools) != 3 {
					t.Errorf("Goroutine %d: expected 3 MCP tools, got %d", id, len(opts.McpTools))
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentPermissionModeAccess tests concurrent access to permission mode constants
func TestConcurrentPermissionModeAccess(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Read permission mode constants concurrently
				modes := []PermissionMode{
					PermissionModeDefault,
					PermissionModeAcceptEdits,
					PermissionModeBypassPermissions,
				}
				
				opts := NewOptions()
				for _, mode := range modes {
					opts.WithPermissionMode(mode)
					
					if opts.PermissionMode == nil {
						t.Errorf("Goroutine %d: permission mode should not be nil", id)
					}
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentInterfaceImplementations tests concurrent interface method calls
func TestConcurrentInterfaceImplementations(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 1000
	
	var wg sync.WaitGroup
	
	// Create instances of all types that implement interfaces
	messages := []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{}},
		&SystemMessage{Subtype: "test"},
		&ResultMessage{Subtype: "test"},
	}
	
	blocks := []ContentBlock{
		&TextBlock{Text: "test"},
		&ToolUseBlock{ID: "1", Name: "test"},
		&ToolResultBlock{ToolUseID: "1"},
	}
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Test Message interface implementations
				for _, msg := range messages {
					_ = msg.Type()
				}
				
				// Test ContentBlock interface implementations
				for _, block := range blocks {
					_ = block.Type()
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestConcurrentWrapperCreation tests concurrent QueryStream wrapper creation
func TestConcurrentWrapperCreation(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Create wrapper with nil internal stream
				wrapper := wrapQueryStream(nil)
				if wrapper == nil {
					t.Errorf("Goroutine %d: wrapper should not be nil", id)
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// TestDataRace tests for potential data races in shared state
func TestDataRace(t *testing.T) {
	const numGoroutines = 100
	
	var wg sync.WaitGroup
	
	// Test concurrent access to package-level functionality
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Multiple operations that might have shared state
			SetParserBufferSize(1024 * (id + 1))
			
			opts := NewOptions().
				WithSystemPrompt("Concurrent test").
				WithAllowedTools("tool1", "tool2")
			
			_ = opts.AllowedTools
			_ = opts.SystemPrompt
			
			// Create errors
			err1 := NewCLINotFoundError("test", "/path")
			err2 := NewConnectionError("test", nil)
			
			_ = err1.Error()
			_ = err2.Error()
			
			// Test interface implementations
			msg := &UserMessage{Content: "test"}
			_ = msg.Type()
			
			block := &TextBlock{Text: "test"}
			_ = block.Type()
		}(i)
	}
	
	wg.Wait()
}

// TestMemoryStressWithConcurrency tests memory allocation under concurrent load
func TestMemoryStressWithConcurrency(t *testing.T) {
	const numGoroutines = 20
	const numOperations = 500
	
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				// Allocate and immediately release to test GC behavior
				opts := NewOptions()
				
				// Add many servers and tools
				for k := 0; k < 10; k++ {
					opts.AddMcpServer(fmt.Sprintf("server_%d_%d", i, k), &StdioServerConfig{
						Command: "test",
						Args:    []string{"arg1", "arg2"},
						Env:     map[string]string{"KEY": "value"},
					})
					opts.AddMcpTool(fmt.Sprintf("tool_%d_%d", i, k))
				}
				
				// Create large messages
				largeContent := make([]ContentBlock, 100)
				for k := range largeContent {
					largeContent[k] = &TextBlock{Text: fmt.Sprintf("Large content %d", k)}
				}
				
				_ = &AssistantMessage{Content: largeContent}
			}
		}(i)
	}
	
	wg.Wait()
}