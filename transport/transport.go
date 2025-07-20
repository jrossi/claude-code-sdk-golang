// Package transport provides abstractions for communicating with Claude Code CLI.
package transport

import (
	"context"
	"github.com/jrossi/claude-code-sdk-golang/types"
	"io"
)

// Transport represents an abstract communication channel with Claude Code CLI.
// Implementations handle the lifecycle of connecting to, streaming from, and
// disconnecting from the CLI process.
type Transport interface {
	// Connect establishes a connection to the Claude Code CLI.
	// The context can be used to cancel the connection attempt.
	Connect(ctx context.Context) error

	// Stream returns channels for receiving data and errors from the CLI.
	// The data channel receives raw bytes from the CLI's stdout.
	// The error channel receives any errors that occur during streaming.
	// Both channels will be closed when streaming ends.
	Stream(ctx context.Context) (<-chan []byte, <-chan error)

	// Close terminates the connection and cleans up resources.
	// It should be safe to call multiple times.
	Close() error

	// IsConnected returns true if the transport is currently connected.
	IsConnected() bool
}

// Config contains configuration for creating a transport.
type Config struct {
	// Prompt is the user prompt to send to Claude.
	Prompt string

	// Options contains query options.
	Options *types.Options

	// CLIPath specifies the path to the Claude Code CLI binary.
	// If empty, the transport will attempt to discover it.
	CLIPath string

	// Timeout specifies the maximum time to wait for CLI operations.
	// If zero, a reasonable default will be used.
	Timeout string

	// MaxBufferSize specifies the maximum size for internal buffers.
	// If zero, a reasonable default will be used.
	MaxBufferSize int

	// Stdout and Stderr can be set for testing to capture CLI output.
	// In normal operation, these should be nil.
	Stdout io.Writer
	Stderr io.Writer
}
