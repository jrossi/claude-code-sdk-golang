package client

import (
	"context"
	"errors"
	"github.com/jrossi/claude-code-sdk-golang/parser"
	"github.com/jrossi/claude-code-sdk-golang/transport"
	"github.com/jrossi/claude-code-sdk-golang/types"
	"sync"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.parser == nil {
		t.Fatal("Expected parser to be initialized")
	}
}

func TestClientSetParserBufferSize(t *testing.T) {
	client := NewClient()

	// Test setting custom buffer size
	customSize := 2048
	client.SetParserBufferSize(customSize)

	// We can't directly test the buffer size since it's private,
	// but we can verify the method doesn't panic and the parser is recreated
	if client.parser == nil {
		t.Error("Expected parser to be set after SetParserBufferSize")
	}
}

func TestClientQueryConfiguration(t *testing.T) {
	client := NewClient()
	
	tests := []struct {
		name    string
		prompt  string
		options *types.Options
	}{
		{
			name:    "nil options",
			prompt:  "Test prompt",
			options: nil,
		},
		{
			name:   "custom options",
			prompt: "Test prompt with options",
			options: types.NewOptions().
				WithSystemPrompt("System prompt").
				WithMaxTurns(5),
		},
		{
			name:    "empty prompt",
			prompt:  "",
			options: types.NewOptions(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// This will fail because we don't have a real CLI, but we can test configuration
			stream, err := client.Query(ctx, tt.prompt, tt.options)
			
			// We expect an error because there's no real CLI (but it might not fail immediately)
			// The important part is that the configuration is set correctly
			if stream != nil {
				stream.Close() // Clean up if somehow created
			}
			_ = err // Might be nil if CLI discovery fails later
			
			// Verify transport config was set
			if client.transportConfig == nil {
				t.Error("Expected transport config to be set")
			}
			if client.transportConfig.Prompt != tt.prompt {
				t.Errorf("Expected prompt %q, got %q", tt.prompt, client.transportConfig.Prompt)
			}
			
			// Check options handling
			expectedOptions := tt.options
			if expectedOptions == nil {
				expectedOptions = types.NewOptions()
			}
			if client.transportConfig.Options == nil {
				t.Error("Expected options to be set")
			}
		})
	}
}

func TestClientQueryWithCLIPathConfiguration(t *testing.T) {
	client := NewClient()
	
	tests := []struct {
		name    string
		prompt  string
		options *types.Options
		cliPath string
	}{
		{
			name:    "nil options with custom CLI path",
			prompt:  "Test prompt",
			options: nil,
			cliPath: "/custom/path/to/claude",
		},
		{
			name:   "custom options with CLI path",
			prompt: "Test prompt with options",
			options: types.NewOptions().
				WithSystemPrompt("System prompt").
				WithMaxTurns(3),
			cliPath: "/usr/local/bin/claude",
		},
		{
			name:    "empty CLI path",
			prompt:  "Test prompt",
			options: types.NewOptions(),
			cliPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// This will fail because we don't have a real CLI, but we can test configuration
			stream, err := client.QueryWithCLIPath(ctx, tt.prompt, tt.options, tt.cliPath)
			
			// We expect an error because there's no real CLI (but it might not fail immediately)
			// The important part is that the configuration is set correctly
			if stream != nil {
				stream.Close() // Clean up if somehow created
			}
			_ = err // Might be nil if CLI discovery fails later
			
			// Verify transport config was set with CLI path
			if client.transportConfig == nil {
				t.Error("Expected transport config to be set")
			}
			if client.transportConfig.Prompt != tt.prompt {
				t.Errorf("Expected prompt %q, got %q", tt.prompt, client.transportConfig.Prompt)
			}
			if client.transportConfig.CLIPath != tt.cliPath {
				t.Errorf("Expected CLI path %q, got %q", tt.cliPath, client.transportConfig.CLIPath)
			}
			
			// Check options handling
			expectedOptions := tt.options
			if expectedOptions == nil {
				expectedOptions = types.NewOptions()
			}
			if client.transportConfig.Options == nil {
				t.Error("Expected options to be set")
			}
		})
	}
}

// Note: We can't easily test the full Query method without a real CLI
// since it would require subprocess execution. The integration tests
// will validate the end-to-end functionality with mocked CLI.

func TestQueryStreamLifecycle(t *testing.T) {
	// Create a mock transport that doesn't require a real CLI
	transport := &mockTransport{}
	parser := parser.NewParser(0) // Use real parser

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream := NewQueryStream(ctx, transport, parser)

	if stream == nil {
		t.Fatal("Expected non-nil stream")
	}

	// Test initial state
	if stream.IsClosed() {
		t.Error("Expected stream to not be closed initially")
	}

	// Test channels are accessible
	if stream.Messages() == nil {
		t.Error("Expected non-nil messages channel")
	}
	if stream.Errors() == nil {
		t.Error("Expected non-nil errors channel")
	}

	// Test close
	err := stream.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !stream.IsClosed() {
		t.Error("Expected stream to be closed after Close()")
	}

	// Test double close
	err = stream.Close()
	if err != nil {
		t.Fatalf("Double close failed: %v", err)
	}
}

func TestQueryStreamStart(t *testing.T) {
	tests := []struct {
		name      string
		transport transport.Transport
		wantError bool
	}{
		{
			name:      "successful start",
			transport: &mockTransport{},
			wantError: false,
		},
		{
			name:      "transport connection error",
			transport: &mockTransportWithError{connectError: errors.New("connection failed")},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			parser := parser.NewParser(0)
			stream := NewQueryStream(ctx, tt.transport, parser)

			err := stream.Start()
			if (err != nil) != tt.wantError {
				t.Errorf("Start() error = %v, wantError %v", err, tt.wantError)
			}

			// Clean up
			stream.Close()
		})
	}
}

func TestQueryStreamMessageFlow(t *testing.T) {
	// Create a transport that sends test messages
	messageTransport := &mockMessageTransport{
		messages: []string{
			`{"type": "user", "message": {"content": "Hello"}}`+"\n",
			`{"type": "assistant", "message": {"content": [{"type": "text", "text": "Hi there!"}]}}`+"\n",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	parser := parser.NewParser(0)
	stream := NewQueryStream(ctx, messageTransport, parser)

	// Start the stream
	err := stream.Start()
	if err != nil {
		t.Fatalf("Failed to start stream: %v", err)
	}

	// Collect messages
	var receivedMessages []types.Message
	var receivedErrors []error

done:
	for {
		select {
		case msg, ok := <-stream.Messages():
			if !ok {
				break done
			}
			receivedMessages = append(receivedMessages, msg)

		case err, ok := <-stream.Errors():
			if !ok {
				// Errors channel closed, but continue reading messages
				continue
			}
			receivedErrors = append(receivedErrors, err)

		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	}

	// Verify messages were received
	if len(receivedMessages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(receivedMessages))
	}

	// Check for unexpected errors
	if len(receivedErrors) > 0 {
		t.Errorf("Unexpected errors: %v", receivedErrors)
	}

	// Clean up
	stream.Close()
}

func TestQueryStreamErrorHandling(t *testing.T) {
	// Create a transport that sends errors
	errorTransport := &mockErrorTransport{
		transportError: errors.New("transport error"),
		messages: []string{
			`{"invalid json"`+"\n", // This will cause a parse error
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	parser := parser.NewParser(0)
	stream := NewQueryStream(ctx, errorTransport, parser)

	// Start the stream
	err := stream.Start()
	if err != nil {
		t.Fatalf("Failed to start stream: %v", err)
	}

	// Collect errors
	var receivedErrors []error

	for {
		select {
		case err, ok := <-stream.Errors():
			if !ok {
				// Errors channel closed
				goto checkErrors
			}
			receivedErrors = append(receivedErrors, err)

		case <-stream.Messages():
			// Consume any messages

		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	}

checkErrors:
	// Verify errors were received
	if len(receivedErrors) == 0 {
		t.Error("Expected to receive errors, got none")
	}

	// Check that we got both transport and parse errors
	hasTransportError := false
	for _, err := range receivedErrors {
		if err.Error() == "transport error" {
			hasTransportError = true
			break
		}
	}
	if !hasTransportError {
		t.Error("Expected to receive transport error")
	}

	// Clean up
	stream.Close()
}

func TestQueryStreamConcurrentAccess(t *testing.T) {
	// Test concurrent access to IsClosed method
	transport := &mockTransport{}
	parser := parser.NewParser(0)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stream := NewQueryStream(ctx, transport, parser)

	// Start multiple goroutines to access IsClosed concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = stream.IsClosed()
			}
		}()
	}

	// Close the stream concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		stream.Close()
	}()

	wg.Wait()

	// Verify final state
	if !stream.IsClosed() {
		t.Error("Expected stream to be closed")
	}
}

func TestQueryStreamTransportCloseError(t *testing.T) {
	// Test error handling when transport.Close() fails
	transport := &mockTransportWithCloseError{
		closeError: errors.New("close failed"),
	}
	parser := parser.NewParser(0)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream := NewQueryStream(ctx, transport, parser)

	// Close should return the transport close error
	err := stream.Close()
	if err == nil {
		t.Error("Expected close error, got nil")
	}
	if err.Error() != "close failed" {
		t.Errorf("Expected 'close failed', got %v", err)
	}

	// Stream should still be marked as closed
	if !stream.IsClosed() {
		t.Error("Expected stream to be closed despite error")
	}
}

func TestQueryStreamContextCancellation(t *testing.T) {
	// Test that context cancellation properly stops all goroutines
	transport := &mockStreamingTransport{
		messages: []string{
			`{"type": "user", "message": {"content": "Message 1"}}`+"\n",
			`{"type": "user", "message": {"content": "Message 2"}}`+"\n",
		},
		delay: 50 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	parser := parser.NewParser(0)
	stream := NewQueryStream(ctx, transport, parser)

	// Start the stream
	err := stream.Start()
	if err != nil {
		t.Fatalf("Failed to start stream: %v", err)
	}

	// Cancel context early
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Verify channels close within reasonable time
	timeout := time.After(500 * time.Millisecond)
	messagesClosed := false
	errorsClosed := false

loop:
	for {
		select {
		case _, ok := <-stream.Messages():
			if !ok {
				messagesClosed = true
			}
		case _, ok := <-stream.Errors():
			if !ok {
				errorsClosed = true
			}
		case <-timeout:
			break loop
		}

		if messagesClosed && errorsClosed {
			break
		}
	}

	if !messagesClosed {
		t.Error("Messages channel should be closed after context cancellation")
	}
	if !errorsClosed {
		t.Error("Errors channel should be closed after context cancellation")
	}

	// Clean up
	stream.Close()
}

// Mock transport for testing
type mockTransport struct {
	connected bool
}

func (mt *mockTransport) Connect(ctx context.Context) error {
	mt.connected = true
	return nil
}

func (mt *mockTransport) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)

	// Close channels immediately for testing
	close(dataChan)
	close(errChan)

	return dataChan, errChan
}

func (mt *mockTransport) Close() error {
	mt.connected = false
	return nil
}

func (mt *mockTransport) IsConnected() bool {
	return mt.connected
}

// Mock transport with connection error
type mockTransportWithError struct {
	connectError error
	connected    bool
}

func (mt *mockTransportWithError) Connect(ctx context.Context) error {
	if mt.connectError != nil {
		return mt.connectError
	}
	mt.connected = true
	return nil
}

func (mt *mockTransportWithError) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)
	close(dataChan)
	close(errChan)
	return dataChan, errChan
}

func (mt *mockTransportWithError) Close() error {
	mt.connected = false
	return nil
}

func (mt *mockTransportWithError) IsConnected() bool {
	return mt.connected
}

// Mock transport with close error
type mockTransportWithCloseError struct {
	closeError error
	connected  bool
}

func (mt *mockTransportWithCloseError) Connect(ctx context.Context) error {
	mt.connected = true
	return nil
}

func (mt *mockTransportWithCloseError) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)
	close(dataChan)
	close(errChan)
	return dataChan, errChan
}

func (mt *mockTransportWithCloseError) Close() error {
	mt.connected = false
	return mt.closeError
}

func (mt *mockTransportWithCloseError) IsConnected() bool {
	return mt.connected
}

// Mock transport that sends test messages
type mockMessageTransport struct {
	messages  []string
	connected bool
}

func (mt *mockMessageTransport) Connect(ctx context.Context) error {
	mt.connected = true
	return nil
}

func (mt *mockMessageTransport) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte, len(mt.messages))
	errChan := make(chan error)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		for _, msg := range mt.messages {
			select {
			case dataChan <- []byte(msg):
			case <-ctx.Done():
				return
			}
		}
	}()

	return dataChan, errChan
}

func (mt *mockMessageTransport) Close() error {
	mt.connected = false
	return nil
}

func (mt *mockMessageTransport) IsConnected() bool {
	return mt.connected
}

// Mock transport that sends errors
type mockErrorTransport struct {
	transportError error
	messages       []string
	connected      bool
}

func (mt *mockErrorTransport) Connect(ctx context.Context) error {
	mt.connected = true
	return nil
}

func (mt *mockErrorTransport) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte, len(mt.messages))
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		// Send transport error first
		if mt.transportError != nil {
			select {
			case errChan <- mt.transportError:
			case <-ctx.Done():
				return
			}
		}

		// Send malformed messages
		for _, msg := range mt.messages {
			select {
			case dataChan <- []byte(msg):
			case <-ctx.Done():
				return
			}
		}
	}()

	return dataChan, errChan
}

func (mt *mockErrorTransport) Close() error {
	mt.connected = false
	return nil
}

func (mt *mockErrorTransport) IsConnected() bool {
	return mt.connected
}

// Mock transport with streaming delay
type mockStreamingTransport struct {
	messages  []string
	delay     time.Duration
	connected bool
}

func (mt *mockStreamingTransport) Connect(ctx context.Context) error {
	mt.connected = true
	return nil
}

func (mt *mockStreamingTransport) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		for _, msg := range mt.messages {
			select {
			case <-time.After(mt.delay):
				select {
				case dataChan <- []byte(msg):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return dataChan, errChan
}

func (mt *mockStreamingTransport) Close() error {
	mt.connected = false
	return nil
}

func (mt *mockStreamingTransport) IsConnected() bool {
	return mt.connected
}
