// +build integration

// Package claudecode contains integration tests that require the Claude Code CLI to be installed.
// Run with: go test -tags=integration ./...
package claudecode

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestRealCLIBasicQuery tests a basic query with the real CLI
func TestRealCLIBasicQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if API key is available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := Query(ctx, "What is 2+2? Please respond with just the number.", nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer stream.Close()

	var receivedAssistantMessage bool
	var receivedResultMessage bool

	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				// Stream completed
				if !receivedAssistantMessage {
					t.Error("Expected to receive at least one assistant message")
				}
				if !receivedResultMessage {
					t.Error("Expected to receive a result message")
				}
				return
			}

			switch msg := message.(type) {
			case *AssistantMessage:
				receivedAssistantMessage = true
				t.Logf("Received assistant message with %d content blocks", len(msg.Content))

				// Check for text content
				hasText := false
				for _, block := range msg.Content {
					if textBlock, ok := block.(*TextBlock); ok {
						hasText = true
						t.Logf("Assistant text: %s", textBlock.Text)
					}
				}
				if !hasText {
					t.Error("Expected assistant message to contain text")
				}

			case *ResultMessage:
				receivedResultMessage = true
				t.Logf("Result: duration=%dms, turns=%d, session=%s", 
					msg.DurationMs, msg.NumTurns, msg.SessionID)

				if msg.IsError {
					t.Errorf("Result indicates error: %s", safeStringPtr(msg.Result))
				}

				if msg.SessionID == "" {
					t.Error("Expected non-empty session ID")
				}

			case *UserMessage:
				t.Logf("User message: %s", msg.Content)

			case *SystemMessage:
				t.Logf("System message [%s]: %v", msg.Subtype, msg.Data)

			default:
				t.Logf("Received message of type: %T", message)
			}

		case err, ok := <-stream.Errors():
			if !ok {
				// Error stream completed
				return
			}
			t.Fatalf("Stream error: %v", err)

		case <-ctx.Done():
			t.Fatalf("Test timeout: %v", ctx.Err())
		}
	}
}

// TestRealCLIWithOptions tests query with various options
func TestRealCLIWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	options := NewOptions().
		WithSystemPrompt("You are a helpful assistant that responds concisely.").
		WithMaxTurns(1).
		WithModel("claude-3-haiku")

	stream, err := Query(ctx, "Explain quantum computing in one sentence.", options)
	if err != nil {
		t.Fatalf("Query with options failed: %v", err)
	}
	defer stream.Close()

	var messageCount int
	var resultReceived bool

	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				if messageCount == 0 {
					t.Error("Expected to receive at least one message")
				}
				if !resultReceived {
					t.Error("Expected to receive a result message")
				}
				return
			}

			messageCount++
			t.Logf("Message %d: %T", messageCount, message)

			switch msg := message.(type) {
			case *AssistantMessage:
				if len(msg.Content) == 0 {
					t.Error("Assistant message should have content")
				}

			case *ResultMessage:
				resultReceived = true
				if msg.NumTurns != 1 {
					t.Errorf("Expected 1 turn, got %d", msg.NumTurns)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				return
			}
			t.Fatalf("Stream error: %v", err)

		case <-ctx.Done():
			t.Fatalf("Test timeout: %v", ctx.Err())
		}
	}
}

// TestRealCLIWithTools tests using tools (if available)
func TestRealCLIWithTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	options := NewOptions().
		WithAllowedTools("Read", "Write").
		WithSystemPrompt("You are a helpful file assistant.").
		WithCwd(tmpDir).
		WithPermissionMode(PermissionModeAcceptEdits)

	prompt := "Create a file called test.txt with the content 'Hello from integration test' and then read it back to verify."

	stream, err := Query(ctx, prompt, options)
	if err != nil {
		t.Fatalf("Query with tools failed: %v", err)
	}
	defer stream.Close()

	var sawToolUse bool
	var sawToolResult bool

	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				t.Logf("Tool use observed: %v", sawToolUse)
				t.Logf("Tool result observed: %v", sawToolResult)
				return
			}

			switch msg := message.(type) {
			case *AssistantMessage:
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *TextBlock:
						t.Logf("Claude: %s", b.Text)
					case *ToolUseBlock:
						sawToolUse = true
						t.Logf("Tool use: %s (ID: %s)", b.Name, b.ID)
						t.Logf("Tool input: %v", b.Input)
					case *ToolResultBlock:
						sawToolResult = true
						if b.IsError != nil && *b.IsError {
							t.Logf("Tool error for %s: %s", b.ToolUseID, safeStringPtr(b.Content))
						} else {
							t.Logf("Tool result for %s: %s", b.ToolUseID, safeStringPtr(b.Content))
						}
					}
				}

			case *ResultMessage:
				if msg.IsError {
					t.Errorf("Result indicates error: %s", safeStringPtr(msg.Result))
				}
				t.Logf("Total cost: $%.4f", safeFloat64Ptr(msg.TotalCostUSD))
			}

		case err, ok := <-stream.Errors():
			if !ok {
				return
			}
			// Tool-related errors might be expected in some environments
			t.Logf("Stream error (might be expected): %v", err)

		case <-ctx.Done():
			t.Fatalf("Test timeout: %v", ctx.Err())
		}
	}
}

// TestRealCLIErrorHandling tests error scenarios
func TestRealCLIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InvalidCLIPath", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := QueryWithCLIPath(ctx, "test", nil, "/fake/invalid/path")
		if err == nil {
			t.Error("Expected error with invalid CLI path")
		}

		t.Logf("Got expected error: %v", err)
	})

	t.Run("VeryShortTimeout", func(t *testing.T) {
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			t.Skip("ANTHROPIC_API_KEY not set, skipping timeout test")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		stream, err := Query(ctx, "test", nil)
		if err != nil {
			// Immediate error is fine
			t.Logf("Got immediate error (expected): %v", err)
			return
		}
		defer stream.Close()

		// Should timeout quickly
		select {
		case <-stream.Messages():
			t.Log("Received message despite short timeout")
		case err := <-stream.Errors():
			t.Logf("Got stream error (expected): %v", err)
		case <-ctx.Done():
			t.Log("Context timeout as expected")
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected faster timeout")
		}
	})
}

// TestRealCLIStreamLifecycle tests proper stream management
func TestRealCLIStreamLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping lifecycle test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := Query(ctx, "Just say 'hello'", nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Test that stream starts open
	if stream.IsClosed() {
		t.Error("Stream should not be closed initially")
	}

	// Test channels are accessible
	if stream.Messages() == nil {
		t.Error("Messages channel should not be nil")
	}
	if stream.Errors() == nil {
		t.Error("Errors channel should not be nil")
	}

	// Process at least one message
	select {
	case message := <-stream.Messages():
		if message != nil {
			t.Logf("Received message: %T", message)
		}
	case err := <-stream.Errors():
		t.Fatalf("Unexpected error: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for first message")
	}

	// Test close
	err = stream.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !stream.IsClosed() {
		t.Error("Stream should be closed after Close()")
	}

	// Test double close
	err = stream.Close()
	if err != nil {
		t.Fatalf("Double close failed: %v", err)
	}
}

// Helper functions

func safeStringPtr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func safeFloat64Ptr(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}