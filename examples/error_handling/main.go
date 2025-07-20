// Package main demonstrates comprehensive error handling patterns with the Claude Code SDK for Go.
// This example shows how to handle different types of errors gracefully.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	claudecode "github.com/jrossi/claude-code-sdk-golang"
)

func cliNotFoundExample() {
	fmt.Println("=== CLI Not Found Error Handling ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try with an invalid CLI path to demonstrate error handling
	stream, err := claudecode.QueryWithCLIPath(ctx, "Hello", nil, "/fake/nonexistent/claude")
	if err != nil {
		handleQueryError(err)
		return
	}
	defer stream.Close()

	// This shouldn't be reached with a fake path
	fmt.Println("Unexpectedly succeeded with fake CLI path")
}

func connectionTimeoutExample() {
	fmt.Println("=== Connection Timeout Error Handling ===")

	// Very short timeout to demonstrate timeout handling
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	stream, err := claudecode.Query(ctx, "What is the meaning of life?", nil)
	if err != nil {
		handleQueryError(err)
		return
	}
	defer stream.Close()

	// Process with timeout handling
	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				fmt.Println("Stream completed")
				return
			}
			fmt.Printf("Received message: %T\n", message)

		case err, ok := <-stream.Errors():
			if !ok {
				fmt.Println("Error stream completed")
				return
			}
			handleStreamError(err)

		case <-ctx.Done():
			fmt.Printf("Context timeout: %v\n", ctx.Err())
			return
		}
	}
}

func streamErrorHandlingExample() {
	fmt.Println("=== Stream Error Handling ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure with potentially problematic settings
	options := claudecode.NewOptions().
		WithAllowedTools("NonexistentTool").
		WithPermissionMode(claudecode.PermissionModeBypassPermissions)

	stream, err := claudecode.Query(ctx, "Use the NonexistentTool to do something", options)
	if err != nil {
		handleQueryError(err)
		return
	}
	defer stream.Close()

	errorCount := 0
	maxErrors := 3

	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				fmt.Println("Stream completed")
				return
			}

			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range msg.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			case *claudecode.ResultMessage:
				if msg.IsError {
					fmt.Printf("Result indicates error: %s\n", safeStringValue(msg.Result))
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				fmt.Println("Error stream completed")
				return
			}

			errorCount++
			fmt.Printf("Stream error #%d: %v\n", errorCount, err)

			// Demonstrate error classification and handling
			if shouldRetry(err) && errorCount < maxErrors {
				fmt.Println("Error appears recoverable, continuing...")
				continue
			}

			if errorCount >= maxErrors {
				fmt.Printf("Too many errors (%d), stopping\n", errorCount)
				return
			}

			handleStreamError(err)

		case <-ctx.Done():
			fmt.Printf("Context cancelled: %v\n", ctx.Err())
			return
		}
	}
}

func resourceCleanupExample() {
	fmt.Println("=== Resource Cleanup Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stream *claudecode.QueryStream
	var err error

	// Demonstrate proper resource cleanup with defer
	defer func() {
		if stream != nil {
			if closeErr := stream.Close(); closeErr != nil {
				fmt.Printf("Warning: Failed to close stream: %v\n", closeErr)
			} else {
				fmt.Println("Stream closed successfully")
			}
		}
	}()

	stream, err = claudecode.Query(ctx, "Hello!", nil)
	if err != nil {
		handleQueryError(err)
		return
	}

	// Simulate early return due to some condition
	fmt.Println("Simulating early return...")
	return // defer will ensure cleanup
}

func robustQueryExample() {
	fmt.Println("=== Robust Query with Retry Logic ===")

	maxRetries := 3
	baseTimeout := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Attempt %d/%d\n", attempt, maxRetries)

		// Increase timeout on each retry
		timeout := baseTimeout * time.Duration(attempt)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		success, err := attemptQuery(ctx, "What's the weather like?")
		cancel()

		if success {
			fmt.Println("Query succeeded!")
			return
		}

		if err != nil {
			fmt.Printf("Attempt %d failed: %v\n", attempt, err)

			if !shouldRetry(err) {
				fmt.Println("Error is not retriable, stopping")
				return
			}

			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("Waiting %v before retry...\n", waitTime)
				time.Sleep(waitTime)
			}
		}
	}

	fmt.Printf("All %d attempts failed\n", maxRetries)
}

func attemptQuery(ctx context.Context, prompt string) (bool, error) {
	stream, err := claudecode.Query(ctx, prompt, nil)
	if err != nil {
		return false, err
	}
	defer stream.Close()

	for {
		select {
		case message, ok := <-stream.Messages():
			if !ok {
				return true, nil // Successfully completed
			}

			// Process message (simplified)
			switch msg := message.(type) {
			case *claudecode.AssistantMessage:
				fmt.Printf("Received response with %d blocks\n", len(msg.Content))
			case *claudecode.ResultMessage:
				if msg.IsError {
					return false, fmt.Errorf("result error: %s", safeStringValue(msg.Result))
				}
			}

		case err, ok := <-stream.Errors():
			if !ok {
				return true, nil // Error stream closed, but might have succeeded
			}
			return false, err

		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
}

// Error classification functions

func handleQueryError(err error) {
	fmt.Printf("Query error: %v\n", err)

	// Type assertion for specific error types
	var cliErr *claudecode.CLINotFoundError
	var connErr *claudecode.ConnectionError
	var procErr *claudecode.ProcessError
	var jsonErr *claudecode.JSONDecodeError

	switch {
	case errors.As(err, &cliErr):
		fmt.Printf("CLI not found at path: %s\n", cliErr.CLIPath)
		fmt.Println("Suggestion: Install Claude Code CLI or check PATH")

	case errors.As(err, &connErr):
		fmt.Printf("Connection issue: %s\n", connErr.Message)
		if connErr.Err != nil {
			fmt.Printf("Underlying error: %v\n", connErr.Err)
		}

	case errors.As(err, &procErr):
		fmt.Printf("Process error (exit code %d): %s\n", procErr.ExitCode, procErr.Message)
		if procErr.Stderr != "" {
			fmt.Printf("Stderr: %s\n", procErr.Stderr)
		}

	case errors.As(err, &jsonErr):
		fmt.Printf("JSON decode error on line: %s\n", jsonErr.Line)
		if jsonErr.OriginalErr != nil {
			fmt.Printf("Parse error: %v\n", jsonErr.OriginalErr)
		}

	default:
		fmt.Printf("Unknown error type: %T\n", err)
	}
}

func handleStreamError(err error) {
	fmt.Printf("Stream error: %v\n", err)

	// Classify and handle stream errors
	switch {
	case isTransientError(err):
		fmt.Println("This appears to be a transient error")
	case isConfigurationError(err):
		fmt.Println("This appears to be a configuration error")
	case isPermissionError(err):
		fmt.Println("This appears to be a permission error")
	default:
		fmt.Println("Unknown stream error type")
	}
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry certain types of errors
	if isConfigurationError(err) || isPermissionError(err) {
		return false
	}

	// Retry transient errors and timeouts
	return isTransientError(err) || errors.Is(err, context.DeadlineExceeded)
}

func isTransientError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "timeout") ||
		contains(errStr, "connection refused") ||
		contains(errStr, "temporary failure") ||
		contains(errStr, "service unavailable")
}

func isConfigurationError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "CLI not found") ||
		contains(errStr, "invalid argument") ||
		contains(errStr, "unknown option") ||
		contains(errStr, "invalid configuration")
}

func isPermissionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "permission denied") ||
		contains(errStr, "access denied") ||
		contains(errStr, "unauthorized") ||
		contains(errStr, "forbidden")
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
		fn   func()
	}{
		{"CLI Not Found", cliNotFoundExample},
		{"Connection Timeout", connectionTimeoutExample},
		{"Stream Error Handling", streamErrorHandlingExample},
		{"Resource Cleanup", resourceCleanupExample},
		{"Robust Query with Retry", robustQueryExample},
	}

	fmt.Println("Claude Code SDK - Error Handling Examples")
	fmt.Println("=========================================")

	for _, example := range examples {
		fmt.Printf("\nRunning %s example...\n", example.name)
		example.fn()
		fmt.Printf("--- %s example completed ---\n", example.name)
	}

	fmt.Println("\nAll error handling examples completed!")
	fmt.Println("Note: Some errors are expected and demonstrate proper handling patterns.")
}