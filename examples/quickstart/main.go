// Package main demonstrates basic usage of the Claude Code SDK for Go.
// This example mirrors the functionality of the Python SDK's quick_start.py.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	claudecode "github.com/jrossi/claude-code-sdk-golang"
)

func basicExample() error {
	fmt.Println("=== Basic Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := claudecode.Query(ctx, "What is 2 + 2?", nil)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	// Process messages from the stream
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				// Messages channel closed
				fmt.Println()
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
				if msg.TotalCostUSD != nil && *msg.TotalCostUSD > 0 {
					fmt.Printf("Cost: $%.4f\n", *msg.TotalCostUSD)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				// Errors channel closed
				return nil
			}
			return fmt.Errorf("stream error: %w", err)

		case <-ctx.Done():
			return fmt.Errorf("context timeout: %w", ctx.Err())
		}
	}
}

func withOptionsExample() error {
	fmt.Println("=== With Options Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := claudecode.NewOptions().
		WithSystemPrompt("You are a helpful assistant that explains things simply.").
		WithMaxTurns(1)

	stream, err := claudecode.Query(ctx, "Explain what Go is in one sentence.", options)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	// Process messages from the stream
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				// Messages channel closed
				fmt.Println()
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
				if msg.TotalCostUSD != nil && *msg.TotalCostUSD > 0 {
					fmt.Printf("Cost: $%.4f\n", *msg.TotalCostUSD)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				// Errors channel closed
				return nil
			}
			return fmt.Errorf("stream error: %w", err)

		case <-ctx.Done():
			return fmt.Errorf("context timeout: %w", ctx.Err())
		}
	}
}

func withToolsExample() error {
	fmt.Println("=== With Tools Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	options := claudecode.NewOptions().
		WithAllowedTools("Read", "Write").
		WithSystemPrompt("You are a helpful file assistant.")

	stream, err := claudecode.Query(ctx, "Create a file called hello.txt with 'Hello, World!' in it", options)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer stream.Close()

	// Process messages from the stream
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				// Messages channel closed
				fmt.Println()
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
				if msg.TotalCostUSD != nil && *msg.TotalCostUSD > 0 {
					fmt.Printf("\nCost: $%.4f\n", *msg.TotalCostUSD)
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				// Errors channel closed
				return nil
			}
			return fmt.Errorf("stream error: %w", err)

		case <-ctx.Done():
			return fmt.Errorf("context timeout: %w", ctx.Err())
		}
	}
}

func main() {
	// Check if Claude Code CLI is available
	if err := basicExample(); err != nil {
		log.Printf("Basic example failed: %v", err)
		if isConnectionError(err) {
			printInstallationInstructions()
			os.Exit(1)
		}
	}

	if err := withOptionsExample(); err != nil {
		log.Printf("Options example failed: %v", err)
		if isConnectionError(err) {
			return // Already printed instructions
		}
	}

	if err := withToolsExample(); err != nil {
		log.Printf("Tools example failed: %v", err)
		if isConnectionError(err) {
			return // Already printed instructions
		}
	}

	fmt.Println("All examples completed successfully!")
}

// isConnectionError checks if the error is related to CLI connection issues
func isConnectionError(err error) bool {
	// Check for common CLI-related error patterns
	errStr := err.Error()
	return contains(errStr, "CLI not found") ||
		contains(errStr, "connection error") ||
		contains(errStr, "executable file not found")
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

func printInstallationInstructions() {
	fmt.Println(`Claude Code CLI not found. To run these examples, you need to install the Claude Code CLI:

1. Install Node.js from: https://nodejs.org/
2. Install Claude Code CLI:
   npm install -g @anthropic-ai/claude-code

3. Set up your API key:
   export ANTHROPIC_API_KEY=your_api_key_here

4. Run this example again:
   go run examples/quickstart.go

For more information, visit: https://github.com/anthropics/claude-code`)
}