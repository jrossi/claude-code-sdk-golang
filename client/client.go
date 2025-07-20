// Package client provides the internal client implementation for coordinating
// transport and parser layers.
package client

import (
	"context"
	"github.com/jrossi/claude-code-sdk-golang/parser"
	transport2 "github.com/jrossi/claude-code-sdk-golang/transport"
	"github.com/jrossi/claude-code-sdk-golang/types"
)

// Client coordinates between transport and parser to provide Claude Code functionality.
type Client struct {
	// Configuration for transport
	transportConfig *transport2.Config

	// Parser for JSON messages
	parser *parser.Parser
}

// NewClient creates a new client with the given configuration.
func NewClient() *Client {
	return &Client{
		parser: parser.NewParser(0), // Use default buffer size
	}
}

// Query initiates a query to Claude Code and returns a QueryStream for receiving messages.
func (c *Client) Query(ctx context.Context, prompt string, options *types.Options) (*QueryStream, error) {
	// Set default options if none provided
	if options == nil {
		options = types.NewOptions()
	}

	// Create transport configuration
	c.transportConfig = &transport2.Config{
		Prompt:  prompt,
		Options: options,
		// CLIPath can be set later if needed
		// MaxBufferSize will use transport defaults
	}

	// Create subprocess transport
	subprocessTransport := transport2.NewSubprocessTransport(c.transportConfig)

	// Create query stream
	stream := NewQueryStream(ctx, subprocessTransport, c.parser)

	// Start the streaming process
	if err := stream.Start(); err != nil {
		return nil, err
	}

	return stream, nil
}

// QueryWithCLIPath initiates a query with a specific CLI path.
// This is useful for testing or when the CLI is installed in a non-standard location.
func (c *Client) QueryWithCLIPath(ctx context.Context, prompt string, options *types.Options, cliPath string) (*QueryStream, error) {
	// Set default options if none provided
	if options == nil {
		options = types.NewOptions()
	}

	// Create transport configuration with custom CLI path
	c.transportConfig = &transport2.Config{
		Prompt:  prompt,
		Options: options,
		CLIPath: cliPath,
	}

	// Create subprocess transport
	subprocessTransport := transport2.NewSubprocessTransport(c.transportConfig)

	// Create query stream
	stream := NewQueryStream(ctx, subprocessTransport, c.parser)

	// Start the streaming process
	if err := stream.Start(); err != nil {
		return nil, err
	}

	return stream, nil
}

// SetParserBufferSize configures the maximum buffer size for JSON parsing.
// This should be called before making queries.
func (c *Client) SetParserBufferSize(size int) {
	c.parser = parser.NewParser(size)
}
