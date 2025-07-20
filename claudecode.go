// Package claudecode provides a Go SDK for interacting with Claude Code CLI.
//
// This package offers a Go-idiomatic interface for using Claude Code, matching
// the functionality of the Python SDK while leveraging Go's strengths like
// channels, context-based cancellation, and explicit error handling.
//
// Basic usage:
//
//	ctx := context.Background()
//	stream, err := claudecode.Query(ctx, "What is 2+2?", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stream.Close()
//
//	for {
//		select {
//		case msg, ok := <-stream.Messages():
//			if !ok {
//				return // Stream ended
//			}
//			// Process message
//		case err, ok := <-stream.Errors():
//			if !ok {
//				return // Stream ended
//			}
//			// Handle error
//		case <-ctx.Done():
//			return // Context cancelled
//		}
//	}
//
// Advanced usage with options:
//
//	options := claudecode.NewOptions().
//		WithSystemPrompt("You are a helpful assistant").
//		WithAllowedTools("Read", "Write").
//		WithPermissionMode(claudecode.PermissionModeAcceptEdits)
//
//	stream, err := claudecode.Query(ctx, "Create a hello.txt file", options)
//
// The SDK supports all Claude Code features including:
// - Tool usage with permission control
// - MCP (Model Context Protocol) server integration
// - Session continuation and resumption
// - Working directory specification
// - Model selection and configuration
package claudecode

import (
	"context"

	"github.com/jrossi/claude-code-sdk-golang/internal/client"
)

// defaultClient is the package-level client instance used by the Query function.
var defaultClient = client.NewClient()

// Query initiates a query to Claude Code and returns a stream for receiving messages.
// This is the main entry point for the SDK, providing a simple interface for most use cases.
//
// The function starts a Claude Code subprocess, sends the prompt, and returns a QueryStream
// that provides channels for receiving messages and errors.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - prompt: The user prompt to send to Claude
//   - options: Configuration options (can be nil for defaults)
//
// Returns:
//   - *QueryStream: Stream interface for receiving messages and errors
//   - error: Setup error if the query cannot be initiated
//
// The QueryStream must be closed when done to clean up resources:
//
//	stream, err := claudecode.Query(ctx, "Hello", nil)
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
// Context cancellation will terminate the Claude Code subprocess:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	stream, err := claudecode.Query(ctx, "Hello", nil)
func Query(ctx context.Context, prompt string, options *Options) (*QueryStream, error) {
	internal, err := defaultClient.Query(ctx, prompt, options)
	if err != nil {
		return nil, err
	}
	return wrapQueryStream(internal), nil
}

// QueryWithCLIPath initiates a query using a specific Claude Code CLI binary path.
// This is useful for testing or when the CLI is installed in a non-standard location.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - prompt: The user prompt to send to Claude
//   - options: Configuration options (can be nil for defaults)
//   - cliPath: Path to the Claude Code CLI binary
//
// Returns:
//   - *QueryStream: Stream interface for receiving messages and errors
//   - error: Setup error if the query cannot be initiated
//
// Example:
//
//	stream, err := claudecode.QueryWithCLIPath(
//		ctx,
//		"Hello",
//		nil,
//		"/usr/local/bin/claude"
//	)
func QueryWithCLIPath(ctx context.Context, prompt string, options *Options, cliPath string) (*QueryStream, error) {
	internal, err := defaultClient.QueryWithCLIPath(ctx, prompt, options, cliPath)
	if err != nil {
		return nil, err
	}
	return wrapQueryStream(internal), nil
}

// SetParserBufferSize configures the maximum buffer size for JSON parsing.
// This affects all subsequent queries made with the package-level Query function.
//
// The buffer size limits memory usage when processing large JSON responses from
// Claude Code. If a single JSON message exceeds this size, it will be rejected
// with a JSONDecodeError.
//
// Default buffer size is 1MB (1024*1024 bytes).
//
// Parameters:
//   - size: Maximum buffer size in bytes
//
// Example:
//
//	// Set buffer size to 5MB for large responses
//	claudecode.SetParserBufferSize(5 * 1024 * 1024)
func SetParserBufferSize(size int) {
	defaultClient.SetParserBufferSize(size)
}

// QueryStream provides a streaming interface for receiving messages from Claude Code.
// It wraps the internal client QueryStream to provide a clean public API.
type QueryStream struct {
	internal *client.QueryStream
}

// Messages returns a channel that receives parsed messages from Claude.
// The channel will be closed when the stream ends.
//
// Message types include:
//   - *UserMessage: User input (rarely seen in responses)
//   - *AssistantMessage: Claude's responses with content blocks
//   - *SystemMessage: System notifications and metadata
//   - *ResultMessage: Final results with cost and usage information
//
// Example:
//
//	for msg := range stream.Messages() {
//		switch m := msg.(type) {
//		case *claudecode.AssistantMessage:
//			for _, block := range m.Content {
//				if textBlock, ok := block.(*claudecode.TextBlock); ok {
//					fmt.Println("Claude:", textBlock.Text)
//				}
//			}
//		case *claudecode.ResultMessage:
//			fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
//		}
//	}
func (qs *QueryStream) Messages() <-chan Message {
	return qs.internal.Messages()
}

// Errors returns a channel that receives errors during streaming.
// The channel will be closed when the stream ends.
//
// Error types include:
//   - *CLINotFoundError: Claude Code CLI not installed or not found
//   - *ConnectionError: Communication issues with the CLI process
//   - *ProcessError: CLI process failed with non-zero exit code
//   - *JSONDecodeError: Malformed JSON from CLI output
//
// Errors are non-fatal unless they indicate complete failure. The stream
// may continue to produce messages even after reporting some errors.
//
// Example:
//
//	for err := range stream.Errors() {
//		if cliErr, ok := err.(*claudecode.CLINotFoundError); ok {
//			log.Fatal("Please install Claude Code:", cliErr.Message)
//		}
//		log.Printf("Stream error: %v", err)
//	}
func (qs *QueryStream) Errors() <-chan error {
	return qs.internal.Errors()
}

// Close terminates the stream and cleans up resources.
// It's safe to call Close multiple times.
//
// Close should always be called to ensure proper cleanup:
//
//	stream, err := claudecode.Query(ctx, "Hello", nil)
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
// After calling Close, the Messages() and Errors() channels will be closed
// and the underlying Claude Code subprocess will be terminated.
func (qs *QueryStream) Close() error {
	return qs.internal.Close()
}

// IsClosed returns true if the stream has been closed.
func (qs *QueryStream) IsClosed() bool {
	return qs.internal.IsClosed()
}

// wrapQueryStream wraps an internal QueryStream to provide the public API.
func wrapQueryStream(internal *client.QueryStream) *QueryStream {
	return &QueryStream{internal: internal}
}
