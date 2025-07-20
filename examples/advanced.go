// Package main demonstrates advanced usage of the Claude Code SDK for Go.
// This example shows MCP server configuration, complex options, and error handling.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	claudecode "github.com/jrossi/claude-code-sdk-golang"
)

func mcpServerExample() error {
	fmt.Println("=== MCP Server Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Configure MCP servers for enhanced capabilities
	options := claudecode.NewOptions().
		WithSystemPrompt("You are an assistant with access to file operations and web search.").
		AddMcpServer("filesystem", &claudecode.StdioServerConfig{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		}).
		AddMcpServer("web_search", &claudecode.SSEServerConfig{
			URL: "https://api.search.example.com/mcp",
			Headers: map[string]string{
				"Authorization": "Bearer " + os.Getenv("SEARCH_API_KEY"),
			},
		}).
		AddMcpTool("read_file").
		AddMcpTool("write_file").
		AddMcpTool("search_web").
		WithMaxTurns(3).
		WithPermissionMode(claudecode.PermissionModeAcceptEdits)

	prompt := `Please help me with the following tasks:
1. Create a summary document in /tmp/summary.txt
2. Search for information about Go programming language features
3. Write the search results to the summary file`

	stream, err := claudecode.Query(ctx, prompt, options)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	return processAdvancedStream(stream, ctx)
}

func customCLIPathExample() error {
	fmt.Println("=== Custom CLI Path Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := claudecode.NewOptions().
		WithSystemPrompt("You are a helpful coding assistant.").
		WithModel("claude-3-haiku").
		WithCwd("/tmp")

	// Use custom CLI path if needed
	cliPath := os.Getenv("CLAUDE_CLI_PATH")
	if cliPath == "" {
		cliPath = "claude" // Use default discovery
	}

	var stream *claudecode.QueryStream
	var err error

	if cliPath == "claude" {
		stream, err = claudecode.Query(ctx, "What files are in the current directory?", options)
	} else {
		stream, err = claudecode.QueryWithCLIPath(ctx, "What files are in the current directory?", options, cliPath)
	}

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	return processAdvancedStream(stream, ctx)
}

func conversationResumptionExample() error {
	fmt.Println("=== Conversation Resumption Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// First conversation
	fmt.Println("Starting new conversation...")
	options := claudecode.NewOptions().
		WithSystemPrompt("You are a math tutor.").
		WithMaxTurns(2)

	stream, err := claudecode.Query(ctx, "What is the quadratic formula?", options)
	if err != nil {
		return fmt.Errorf("initial query failed: %w", err)
	}

	var sessionID string
	err = processStreamAndExtractSession(stream, ctx, &sessionID)
	stream.Close()
	if err != nil {
		return fmt.Errorf("failed to process initial conversation: %w", err)
	}

	if sessionID == "" {
		fmt.Println("No session ID found, skipping resumption example")
		return nil
	}

	// Resume conversation
	fmt.Printf("Resuming conversation with session ID: %s\n", sessionID)
	resumeOptions := claudecode.NewOptions().
		WithSystemPrompt("You are a math tutor.").
		WithResume(sessionID).
		WithContinueConversation()

	resumeStream, err := claudecode.Query(ctx, "Can you give me an example of using it?", resumeOptions)
	if err != nil {
		return fmt.Errorf("resume query failed: %w", err)
	}
	defer resumeStream.Close()

	return processAdvancedStream(resumeStream, ctx)
}

func processAdvancedStream(stream *claudecode.QueryStream, ctx context.Context) error {
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				fmt.Println()
				return nil
			}

			switch msg := message.(type) {
			case *claudecode.UserMessage:
				fmt.Printf("User: %s\n", msg.Content)

			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					switch b := block.(type) {
					case *claudecode.TextBlock:
						fmt.Printf("Claude: %s\n", b.Text)
					case *claudecode.ToolUseBlock:
						fmt.Printf("Claude is using tool: %s\n", b.Name)
						fmt.Printf("  ID: %s\n", b.ID)
						if len(b.Input) > 0 {
							fmt.Printf("  Input: %v\n", b.Input)
						}
					case *claudecode.ToolResultBlock:
						if b.IsError != nil && *b.IsError {
							fmt.Printf("Tool error for %s: %s\n", b.ToolUseID, safeStringValue(b.Content))
						} else {
							fmt.Printf("Tool result for %s: %s\n", b.ToolUseID, safeStringValue(b.Content))
						}
					}
				}

			case *claudecode.SystemMessage:
				fmt.Printf("System [%s]: %v\n", msg.Subtype, msg.Data)

			case *claudecode.ResultMessage:
				fmt.Printf("\n--- Result ---\n")
				fmt.Printf("Duration: %dms (API: %dms)\n", msg.DurationMs, msg.DurationAPIMs)
				fmt.Printf("Turns: %d\n", msg.NumTurns)
				fmt.Printf("Session: %s\n", msg.SessionID)
				if msg.TotalCostUSD != nil {
					fmt.Printf("Cost: $%.4f\n", *msg.TotalCostUSD)
				}
				if msg.Usage != nil {
					fmt.Printf("Usage: %v\n", msg.Usage)
				}
				if msg.IsError {
					fmt.Printf("Error: %s\n", safeStringValue(msg.Result))
				} else if msg.Result != nil {
					fmt.Printf("Result: %s\n", *msg.Result)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				return nil
			}
			
			// Handle different types of errors gracefully
			switch {
			case isToolError(err):
				fmt.Printf("Tool error (continuing): %v\n", err)
				continue
			case isPermissionError(err):
				fmt.Printf("Permission error (continuing): %v\n", err)
				continue
			default:
				return fmt.Errorf("stream error: %w", err)
			}

		case <-ctx.Done():
			return fmt.Errorf("context timeout: %w", ctx.Err())
		}
	}
}

func processStreamAndExtractSession(stream *claudecode.QueryStream, ctx context.Context, sessionID *string) error {
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				return nil
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			case *claudecode.ResultMessage:
				*sessionID = msg.SessionID
				if msg.TotalCostUSD != nil && *msg.TotalCostUSD > 0 {
					fmt.Printf("Cost: $%.4f\n", *msg.TotalCostUSD)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				return nil
			}
			return fmt.Errorf("stream error: %w", err)

		case <-ctx.Done():
			return fmt.Errorf("context timeout: %w", ctx.Err())
		}
	}
}

func isToolError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "tool") && (contains(errStr, "error") || contains(errStr, "failed"))
}

func isPermissionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "permission") || contains(errStr, "denied") || contains(errStr, "unauthorized")
}

func safeStringValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	examples := []struct {
		name string
		fn   func() error
	}{
		{"MCP Server", mcpServerExample},
		{"Custom CLI Path", customCLIPathExample},
		{"Conversation Resumption", conversationResumptionExample},
	}

	for _, example := range examples {
		fmt.Printf("Running %s example...\n", example.name)
		if err := example.fn(); err != nil {
			log.Printf("%s example failed: %v", example.name, err)
			
			// If it's a connection error, print instructions and exit
			if isConnectionError(err) {
				printInstallationInstructions()
				return
			}
		}
		fmt.Println()
	}

	fmt.Println("All advanced examples completed!")
}

func isConnectionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "CLI not found") ||
		contains(errStr, "connection error") ||
		contains(errStr, "executable file not found")
}

func printInstallationInstructions() {
	fmt.Println(`
Claude Code CLI not found. To run these examples, you need to install the Claude Code CLI:

1. Install Node.js from: https://nodejs.org/
2. Install Claude Code CLI:
   npm install -g @anthropic-ai/claude-code

3. Set up your API key:
   export ANTHROPIC_API_KEY=your_api_key_here

4. (Optional) For MCP servers, install MCP server packages:
   npm install -g @modelcontextprotocol/server-filesystem

5. Run this example again:
   go run examples/advanced.go

For more information, visit: https://github.com/anthropics/claude-code
`)
}