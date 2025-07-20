package client

import (
	"context"
	"github.com/jrossi/claude-code-sdk-golang/parser"
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
