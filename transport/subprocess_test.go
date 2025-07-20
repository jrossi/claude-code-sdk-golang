package transport

import (
	"context"
	types2 "github.com/jrossi/claude-code-sdk-golang/types"
	"strings"
	"testing"
)

func TestNewSubprocessTransport(t *testing.T) {
	config := &Config{
		Prompt:  "Hello, world!",
		Options: types2.NewOptions(),
	}

	transport := NewSubprocessTransport(config)

	if transport == nil {
		t.Fatal("Expected non-nil transport")
	}

	if transport.IsConnected() {
		t.Error("Expected transport to not be connected initially")
	}
}

func TestCLIDiscovery(t *testing.T) {
	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: types2.NewOptions(),
		},
	}

	// Test CLI discovery (this will likely fail in CI but should not crash)
	_, err := transport.discoverCLI()
	if err != nil {
		// Expected in most test environments - just verify it's a proper error
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	}
}

func TestCommandBuilding(t *testing.T) {
	tests := []struct {
		name     string
		options  *types2.Options
		expected []string
	}{
		{
			name:    "basic options",
			options: types2.NewOptions(),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--print", "test prompt",
			},
		},
		{
			name: "with system prompt",
			options: types2.NewOptions().
				WithSystemPrompt("You are helpful"),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--system-prompt", "You are helpful",
				"--print", "test prompt",
			},
		},
		{
			name: "with tools",
			options: types2.NewOptions().
				WithAllowedTools("Read", "Write").
				WithDisallowedTools("Bash"),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--allowedTools", "Read,Write",
				"--disallowedTools", "Bash",
				"--print", "test prompt",
			},
		},
		{
			name: "with max turns and model",
			options: types2.NewOptions().
				WithMaxTurns(5).
				WithModel("claude-3-sonnet"),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--max-turns", "5",
				"--model", "claude-3-sonnet",
				"--print", "test prompt",
			},
		},
		{
			name: "with permission mode",
			options: types2.NewOptions().
				WithPermissionMode(types2.PermissionModeAcceptEdits),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--permission-mode", "acceptEdits",
				"--print", "test prompt",
			},
		},
		{
			name: "with continue conversation and resume",
			options: types2.NewOptions().
				WithContinueConversation().
				WithResume("session_123"),
			expected: []string{
				"--output-format", "stream-json", "--verbose",
				"--continue",
				"--resume", "session_123",
				"--print", "test prompt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &SubprocessTransport{
				config: &Config{
					Prompt:  "test prompt",
					Options: tt.options,
				},
			}

			cmd, err := transport.buildCommand("/fake/claude")
			if err != nil {
				t.Fatalf("buildCommand failed: %v", err)
			}

			args := cmd.Args[1:] // Skip the binary name

			// Check that all expected args are present
			for _, expected := range tt.expected {
				found := false
				for _, arg := range args {
					if arg == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected argument '%s' not found in %v", expected, args)
				}
			}

			// Verify CLAUDE_CODE_ENTRYPOINT environment variable is set
			hasEntrypoint := false
			for _, env := range cmd.Env {
				if strings.Contains(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go") {
					hasEntrypoint = true
					break
				}
			}
			if !hasEntrypoint {
				t.Error("Expected CLAUDE_CODE_ENTRYPOINT=sdk-go environment variable")
			}
		})
	}
}

func TestMcpServerConfigConversion(t *testing.T) {
	options := types2.NewOptions().
		AddMcpServer("stdio_server", &types2.StdioServerConfig{
			Command: "python",
			Args:    []string{"-m", "my_server"},
			Env:     map[string]string{"DEBUG": "1"},
		}).
		AddMcpServer("sse_server", &types2.SSEServerConfig{
			URL:     "https://example.com/mcp",
			Headers: map[string]string{"Authorization": "Bearer token"},
		}).
		AddMcpServer("http_server", &types2.HTTPServerConfig{
			URL:     "https://api.example.com/mcp",
			Headers: map[string]string{"X-API-Key": "key123"},
		})

	transport := &SubprocessTransport{
		config: &Config{
			Prompt:  "test",
			Options: options,
		},
	}

	converted := transport.convertMcpServers(options.McpServers)

	// Test stdio server conversion
	stdio, ok := converted["stdio_server"].(map[string]any)
	if !ok {
		t.Fatal("stdio_server not found or wrong type")
	}
	if stdio["type"] != "stdio" {
		t.Error("Expected stdio type")
	}
	if stdio["command"] != "python" {
		t.Error("Expected python command")
	}

	// Test SSE server conversion
	sse, ok := converted["sse_server"].(map[string]any)
	if !ok {
		t.Fatal("sse_server not found or wrong type")
	}
	if sse["type"] != "sse" {
		t.Error("Expected sse type")
	}
	if sse["url"] != "https://example.com/mcp" {
		t.Error("Expected correct SSE URL")
	}

	// Test HTTP server conversion
	http, ok := converted["http_server"].(map[string]any)
	if !ok {
		t.Fatal("http_server not found or wrong type")
	}
	if http["type"] != "http" {
		t.Error("Expected http type")
	}
	if http["url"] != "https://api.example.com/mcp" {
		t.Error("Expected correct HTTP URL")
	}
}

func TestTransportLifecycle(t *testing.T) {
	config := &Config{
		Prompt:  "test",
		Options: types2.NewOptions(),
		CLIPath: "/fake/claude", // Use fake path to avoid actual CLI requirement
	}

	transport := NewSubprocessTransport(config)

	// Test initial state
	if transport.IsConnected() {
		t.Error("Expected transport to not be connected initially")
	}

	// Test connect (should not fail with fake CLI path during command building)
	ctx := context.Background()
	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !transport.IsConnected() {
		t.Error("Expected transport to be connected after Connect()")
	}

	// Test close
	err = transport.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if transport.IsConnected() {
		t.Error("Expected transport to not be connected after Close()")
	}

	// Test double close (should not error)
	err = transport.Close()
	if err != nil {
		t.Fatalf("Double close failed: %v", err)
	}
}
