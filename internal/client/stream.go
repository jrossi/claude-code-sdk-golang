package client

import (
	"context"
	"sync"

	"github.com/jrossi/claude-code-sdk-golang/internal/parser"
	"github.com/jrossi/claude-code-sdk-golang/internal/transport"
	"github.com/jrossi/claude-code-sdk-golang/internal/types"
)

// QueryStream provides a streaming interface for receiving messages from Claude Code.
// It coordinates between the transport layer (subprocess) and parser layer (JSON parsing).
type QueryStream struct {
	// Transport handles subprocess communication
	transport transport.Transport

	// Parser handles JSON message parsing
	parser *parser.Parser

	// Channels for message streaming
	messages chan types.Message
	errors   chan error

	// Lifecycle management
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	closeMutex sync.Mutex
}

// NewQueryStream creates a new query stream with the given transport and parser.
func NewQueryStream(ctx context.Context, transport transport.Transport, parser *parser.Parser) *QueryStream {
	// Create a cancellable context for this stream
	streamCtx, cancel := context.WithCancel(ctx)

	return &QueryStream{
		transport: transport,
		parser:    parser,
		messages:  make(chan types.Message, 50), // Buffered for performance
		errors:    make(chan error, 20),         // Buffered for error reporting
		ctx:       streamCtx,
		cancel:    cancel,
	}
}

// Start begins the streaming process by connecting transport and starting parsing.
func (qs *QueryStream) Start() error {
	// Connect to the CLI
	if err := qs.transport.Connect(qs.ctx); err != nil {
		return err
	}

	// Start streaming from transport
	rawData, transportErrors := qs.transport.Stream(qs.ctx)

	// Start parsing the raw data
	parsedMessages, parseErrors := qs.parser.ParseMessages(qs.ctx, rawData)

	// Start goroutines to merge the streams
	go qs.mergeMessages(parsedMessages)
	go qs.mergeErrors(transportErrors, parseErrors)

	return nil
}

// Messages returns a channel that receives parsed messages from Claude.
// The channel will be closed when the stream ends.
func (qs *QueryStream) Messages() <-chan types.Message {
	return qs.messages
}

// Errors returns a channel that receives errors during streaming.
// The channel will be closed when the stream ends.
func (qs *QueryStream) Errors() <-chan error {
	return qs.errors
}

// Close terminates the stream and cleans up resources.
// It's safe to call Close multiple times.
func (qs *QueryStream) Close() error {
	qs.closeMutex.Lock()
	defer qs.closeMutex.Unlock()

	if qs.closed {
		return nil
	}

	qs.closed = true

	// Cancel the context to signal all goroutines
	qs.cancel()

	// Close the transport
	if err := qs.transport.Close(); err != nil {
		return err
	}

	// Don't close channels here - let the merge goroutines close them
	// when they detect context cancellation

	return nil
}

// IsClosed returns true if the stream has been closed.
func (qs *QueryStream) IsClosed() bool {
	qs.closeMutex.Lock()
	defer qs.closeMutex.Unlock()
	return qs.closed
}

// mergeMessages forwards parsed messages to the messages channel.
func (qs *QueryStream) mergeMessages(parsedMessages <-chan types.Message) {
	defer func() {
		// When parsing is done, close messages channel
		close(qs.messages)
	}()

	for {
		select {
		case <-qs.ctx.Done():
			return
		case msg, ok := <-parsedMessages:
			if !ok {
				// Parsed messages channel closed
				return
			}

			// Forward the message (non-blocking)
			select {
			case qs.messages <- msg:
			case <-qs.ctx.Done():
				return
			}
		}
	}
}

// mergeErrors forwards errors from both transport and parser to the errors channel.
func (qs *QueryStream) mergeErrors(transportErrors, parseErrors <-chan error) {
	defer func() {
		// When both error sources are done, close errors channel
		close(qs.errors)
	}()

	// Track if channels are still open
	transportOpen := true
	parseOpen := true

	for transportOpen || parseOpen {
		select {
		case <-qs.ctx.Done():
			return

		case err, ok := <-transportErrors:
			if !ok {
				transportOpen = false
				break
			}
			// Forward transport error (non-blocking)
			select {
			case qs.errors <- err:
			case <-qs.ctx.Done():
				return
			}

		case err, ok := <-parseErrors:
			if !ok {
				parseOpen = false
				break
			}
			// Forward parse error (non-blocking)
			select {
			case qs.errors <- err:
			case <-qs.ctx.Done():
				return
			}
		}
	}
}
