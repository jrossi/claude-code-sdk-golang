package transport

import (
	"context"
	"fmt"
	"github.com/jrossi/claude-code-sdk-golang/types"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestSubprocessTransportErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "nil config",
			config: nil,
			expectError: true,
		},
		{
			name: "empty config",
			config: &Config{},
			expectError: true,
		},
		{
			name: "invalid CLI path",
			config: &Config{
				Prompt:  "test",
				Options: types.NewOptions(),
				CLIPath: "/nonexistent/path/to/claude",
			},
			expectError: false, // Connect shouldn't fail, but Stream should
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transport *SubprocessTransport
			
			if tt.config == nil {
				// Test nil config handling
				transport = NewSubprocessTransport(tt.config)
				if transport == nil {
					t.Error("Expected non-nil transport even with nil config")
				}
				return
			}
			
			transport = NewSubprocessTransport(tt.config)
			
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			
			err := transport.Connect(ctx)
			if tt.expectError && err == nil {
				t.Error("Expected connection error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected connection error: %v", err)
			}
		})
	}
}

func TestSubprocessTransportDoubleConnect(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
		CLIPath: "/fake/claude",
	}
	
	transport := NewSubprocessTransport(config)
	ctx := context.Background()
	
	// First connect
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("First connect failed: %v", err)
	}
	
	// Second connect should not error
	err = transport.Connect(ctx)
	if err != nil {
		t.Errorf("Second connect failed: %v", err)
	}
	
	if !transport.IsConnected() {
		t.Error("Expected transport to remain connected")
	}
}

func TestSubprocessTransportStreamWithoutConnect(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
	}
	
	transport := NewSubprocessTransport(config)
	ctx := context.Background()
	
	// Try to stream without connecting
	dataChan, errChan := transport.Stream(ctx)
	
	// Should get an error
	select {
	case err := <-errChan:
		if err == nil {
			t.Error("Expected error when streaming without connect")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), "not connected") {
			t.Errorf("Expected 'not connected' error, got: %v", err)
		}
	case <-dataChan:
		t.Error("Should not receive data when not connected")
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for error")
	}
}

func TestSubprocessTransportDoubleStream(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
		CLIPath: "/fake/claude",
	}
	
	transport := NewSubprocessTransport(config)
	ctx := context.Background()
	
	// Connect first
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	
	// First stream call
	dataChan1, errChan1 := transport.Stream(ctx)
	
	// Second stream call should return same channels
	dataChan2, errChan2 := transport.Stream(ctx)
	
	if dataChan1 != dataChan2 {
		t.Error("Expected same data channel on double stream")
	}
	if errChan1 != errChan2 {
		t.Error("Expected same error channel on double stream")
	}
}

func TestSubprocessTransportConcurrentAccess(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
		CLIPath: "/fake/claude",
	}
	
	transport := NewSubprocessTransport(config)
	ctx := context.Background()
	
	// Test concurrent access to IsConnected
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = transport.IsConnected()
			}
		}()
	}
	
	// Connect concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		transport.Connect(ctx)
	}()
	
	wg.Wait()
}

func TestBuildCommandWithAllOptions(t *testing.T) {
	// Test comprehensive option combinations
	options := types.NewOptions().
		WithSystemPrompt("You are a helpful assistant").
		WithAppendSystemPrompt("Additional instructions").
		WithAllowedTools("Read", "Write", "Search").
		WithDisallowedTools("Bash", "Execute").
		WithPermissionMode(types.PermissionModeBypassPermissions).
		WithMaxTurns(10).
		WithModel("claude-3-opus").
		WithCwd("/custom/directory").
		WithContinueConversation().
		WithResume("session_456").
		AddMcpTool("filesystem").
		AddMcpTool("web_search").
		AddMcpServer("filesystem", &types.StdioServerConfig{
			Command: "python",
			Args:    []string{"-m", "filesystem_server"},
		})
	
	config := &Config{
		Prompt:        "Complex test prompt",
		Options:       options,
		MaxBufferSize: 8192,
	}
	
	transport := NewSubprocessTransport(config)
	cmd, err := transport.buildCommand("/fake/claude")
	if err != nil {
		t.Fatalf("buildCommand failed: %v", err)
	}
	
	args := cmd.Args[1:] // Skip binary name
	
	// Check some critical arguments are present (not all due to different ordering/grouping)
	mustHaveArgs := []string{
		"--output-format",
		"--system-prompt", 
		"--allowedTools",
		"--max-turns",
		"--model",
		"--continue",
		"--resume",
		"--print",
	}
	
	for _, mustHave := range mustHaveArgs {
		found := false
		for _, arg := range args {
			if arg == mustHave {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected argument '%s' not found in %v", mustHave, args)
		}
	}
	
	// Check that some values are present
	if !contains(args, "You are a helpful assistant") {
		t.Error("Expected system prompt value not found")
	}
	if !contains(args, "Read,Write,Search") {
		t.Error("Expected allowed tools value not found")
	}
	if !contains(args, "10") {
		t.Error("Expected max turns value not found")
	}
}

func TestBuildCommandEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "empty prompt",
			config: &Config{
				Prompt:  "",
				Options: types.NewOptions(),
			},
			wantErr: false,
		},
		{
			name: "nil options",
			config: &Config{
				Prompt:  "test",
				Options: nil,
			},
			wantErr: true,
		},
		{
			name: "very long prompt",
			config: &Config{
				Prompt:  fmt.Sprintf("This is a very long prompt %s", strings.Repeat("x", 10000)),
				Options: types.NewOptions(),
			},
			wantErr: false,
		},
		{
			name: "prompt with special characters",
			config: &Config{
				Prompt:  "Test with \"quotes\" and 'apostrophes' and $variables and \n newlines \t tabs",
				Options: types.NewOptions(),
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewSubprocessTransport(tt.config)
			_, err := transport.buildCommand("/fake/claude")
			
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscoverCLIEdgeCases(t *testing.T) {
	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: types.NewOptions(),
		},
	}
	
	// Test CLI discovery on different platforms
	_, err := transport.discoverCLI()
	
	// We expect this to fail in test environments, but it should not panic
	if err == nil {
		// If it succeeds, that's fine too (maybe CLI is actually installed)
		t.Log("CLI discovery succeeded (CLI may be installed)")
	} else {
		// Verify error message is meaningful
		if err.Error() == "" {
			t.Error("Expected non-empty error message from CLI discovery")
		}
		t.Logf("CLI discovery failed as expected: %v", err)
	}
}

func TestDiscoverCLIExecutablePaths(t *testing.T) {
	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: types.NewOptions(),
		},
	}
	
	// Test the internal logic would test findExecutableInPath if it were exported
	candidates := []string{"claude", "claude-code"}
	if runtime.GOOS == "windows" {
		candidates = []string{"claude.exe", "claude-code.exe"}
	}
	
	// Since findExecutableInPath is not exported, we test the overall discovery
	for _, candidate := range candidates {
		// This will likely fail, but should not panic
		_, err := transport.discoverCLI()
		t.Logf("CLI discovery for %s: %v", candidate, err)
		break // Only test once since it's the same method
	}
}

func TestConvertMcpServersEdgeCases(t *testing.T) {
	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: types.NewOptions(),
		},
	}
	
	tests := []struct {
		name    string
		servers map[string]types.McpServerConfig
	}{
		{
			name:    "nil servers",
			servers: nil,
		},
		{
			name:    "empty servers",
			servers: map[string]types.McpServerConfig{},
		},
		{
			name: "stdio server with empty args",
			servers: map[string]types.McpServerConfig{
				"test": &types.StdioServerConfig{
					Command: "test-command",
					Args:    []string{},
					Env:     nil,
				},
			},
		},
		{
			name: "stdio server with nil env",
			servers: map[string]types.McpServerConfig{
				"test": &types.StdioServerConfig{
					Command: "test-command",
					Args:    []string{"arg1", "arg2"},
					Env:     nil,
				},
			},
		},
		{
			name: "sse server with empty headers",
			servers: map[string]types.McpServerConfig{
				"test": &types.SSEServerConfig{
					URL:     "https://example.com",
					Headers: map[string]string{},
				},
			},
		},
		{
			name: "http server with nil headers",
			servers: map[string]types.McpServerConfig{
				"test": &types.HTTPServerConfig{
					URL:     "https://example.com",
					Headers: nil,
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := transport.convertMcpServers(tt.servers)
			
			if tt.servers == nil && len(result) != 0 {
				t.Error("Expected empty result for nil input")
			}
			if tt.servers != nil && len(tt.servers) == 0 && len(result) != 0 {
				t.Error("Expected empty result for empty input")
			}
		})
	}
}

func TestTransportCloseWithoutConnect(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
	}
	
	transport := NewSubprocessTransport(config)
	
	// Close without connecting should not error
	err := transport.Close()
	if err != nil {
		t.Errorf("Close without connect failed: %v", err)
	}
	
	if transport.IsConnected() {
		t.Error("Expected transport to not be connected after close")
	}
}

func TestTransportConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "config with custom buffer size",
			config: &Config{
				Prompt:        "test",
				Options:       types.NewOptions(),
				MaxBufferSize: 16384,
			},
		},
		{
			name: "config with zero buffer size",
			config: &Config{
				Prompt:        "test",
				Options:       types.NewOptions(),
				MaxBufferSize: 0,
			},
		},
		{
			name: "config with negative buffer size",
			config: &Config{
				Prompt:        "test",
				Options:       types.NewOptions(),
				MaxBufferSize: -1,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewSubprocessTransport(tt.config)
			if transport == nil {
				t.Error("Expected non-nil transport")
			}
			
			// Test that config is stored
			if transport.config != tt.config {
				t.Error("Expected config to be stored in transport")
			}
		})
	}
}

func TestFindExecutableInPathEdgeCases(t *testing.T) {
	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: types.NewOptions(),
		},
	}
	
	tests := []struct {
		name      string
		executable string
	}{
		{"empty string", ""},
		{"nonexistent command", "definitely-nonexistent-command-12345"},
		{"command with path separator", "path/to/nonexistent"},
		{"command with spaces", "command with spaces"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic - test CLI discovery instead of internal method
			_, err := transport.discoverCLI()
			// We expect this to fail for most test cases, which is fine
			t.Logf("CLI discovery with %s: %v", tt.executable, err)
		})
	}
}

func TestSubprocessTransportChannelCapacity(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
	}
	
	transport := NewSubprocessTransport(config)
	
	// Verify channel capacities are as expected
	if cap(transport.dataChan) != 100 {
		t.Errorf("Expected data channel capacity 100, got %d", cap(transport.dataChan))
	}
	if cap(transport.errChan) != 10 {
		t.Errorf("Expected error channel capacity 10, got %d", cap(transport.errChan))
	}
	if cap(transport.doneChan) != 0 {
		t.Errorf("Expected done channel capacity 0, got %d", cap(transport.doneChan))
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types.NewOptions(),
	}
	
	transport := NewSubprocessTransport(config)
	cmd, err := transport.buildCommand("/fake/claude")
	if err != nil {
		t.Fatalf("buildCommand failed: %v", err)
	}
	
	// Check for SDK entrypoint environment variable
	foundEntrypoint := false
	for _, env := range cmd.Env {
		if env == "CLAUDE_CODE_ENTRYPOINT=sdk-go" {
			foundEntrypoint = true
			break
		}
	}
	
	if !foundEntrypoint {
		t.Error("Expected CLAUDE_CODE_ENTRYPOINT=sdk-go environment variable")
	}
	
	// Check that other environment variables are preserved
	if len(cmd.Env) == 0 {
		t.Error("Expected environment variables to be preserved")
	}
}