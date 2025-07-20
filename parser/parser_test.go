package parser

import (
	"context"
	"github.com/jrossi/claude-code-sdk-golang/types"
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	// Test with default buffer size
	parser := NewParser(0)
	if parser.maxBufferSize != DefaultMaxBufferSize {
		t.Errorf("Expected default buffer size %d, got %d", DefaultMaxBufferSize, parser.maxBufferSize)
	}

	// Test with custom buffer size
	customSize := 2048
	parser = NewParser(customSize)
	if parser.maxBufferSize != customSize {
		t.Errorf("Expected custom buffer size %d, got %d", customSize, parser.maxBufferSize)
	}
}

func TestParseUserMessage(t *testing.T) {
	parser := NewParser(0)

	raw := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": "Hello, Claude!",
		},
	}

	msg, err := parser.parseUserMessage(raw)
	if err != nil {
		t.Fatalf("parseUserMessage failed: %v", err)
	}

	if msg.Content != "Hello, Claude!" {
		t.Errorf("Expected content 'Hello, Claude!', got '%s'", msg.Content)
	}

	if msg.Type() != "user" {
		t.Errorf("Expected type 'user', got '%s'", msg.Type())
	}
}

func TestParseTextBlock(t *testing.T) {
	parser := NewParser(0)

	block := map[string]any{
		"type": "text",
		"text": "This is a text response",
	}

	contentBlock, err := parser.parseContentBlock(block)
	if err != nil {
		t.Fatalf("parseContentBlock failed: %v", err)
	}

	textBlock, ok := contentBlock.(*types.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", contentBlock)
	}

	if textBlock.Text != "This is a text response" {
		t.Errorf("Expected text 'This is a text response', got '%s'", textBlock.Text)
	}

	if textBlock.Type() != "text" {
		t.Errorf("Expected type 'text', got '%s'", textBlock.Type())
	}
}

func TestParseToolUseBlock(t *testing.T) {
	parser := NewParser(0)

	block := map[string]any{
		"type": "tool_use",
		"id":   "tool_123",
		"name": "Read",
		"input": map[string]any{
			"file_path": "/path/to/file.txt",
		},
	}

	contentBlock, err := parser.parseContentBlock(block)
	if err != nil {
		t.Fatalf("parseContentBlock failed: %v", err)
	}

	toolBlock, ok := contentBlock.(*types.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected ToolUseBlock, got %T", contentBlock)
	}

	if toolBlock.ID != "tool_123" {
		t.Errorf("Expected ID 'tool_123', got '%s'", toolBlock.ID)
	}

	if toolBlock.Name != "Read" {
		t.Errorf("Expected name 'Read', got '%s'", toolBlock.Name)
	}

	if toolBlock.Type() != "tool_use" {
		t.Errorf("Expected type 'tool_use', got '%s'", toolBlock.Type())
	}

	filePath, ok := toolBlock.Input["file_path"].(string)
	if !ok || filePath != "/path/to/file.txt" {
		t.Error("Expected file_path input parameter")
	}
}

func TestParseToolResultBlock(t *testing.T) {
	parser := NewParser(0)

	block := map[string]any{
		"type":        "tool_result",
		"tool_use_id": "tool_123",
		"content":     "File contents here",
		"is_error":    false,
	}

	contentBlock, err := parser.parseContentBlock(block)
	if err != nil {
		t.Fatalf("parseContentBlock failed: %v", err)
	}

	resultBlock, ok := contentBlock.(*types.ToolResultBlock)
	if !ok {
		t.Fatalf("Expected ToolResultBlock, got %T", contentBlock)
	}

	if resultBlock.ToolUseID != "tool_123" {
		t.Errorf("Expected tool_use_id 'tool_123', got '%s'", resultBlock.ToolUseID)
	}

	if resultBlock.Content == nil || *resultBlock.Content != "File contents here" {
		t.Error("Expected content 'File contents here'")
	}

	if resultBlock.IsError == nil || *resultBlock.IsError != false {
		t.Error("Expected is_error false")
	}

	if resultBlock.Type() != "tool_result" {
		t.Errorf("Expected type 'tool_result', got '%s'", resultBlock.Type())
	}
}

func TestParseAssistantMessage(t *testing.T) {
	parser := NewParser(0)

	raw := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{
					"type": "text",
					"text": "Hello! I can help you.",
				},
				map[string]any{
					"type": "tool_use",
					"id":   "tool_456",
					"name": "Write",
					"input": map[string]any{
						"file_path": "/tmp/output.txt",
						"content":   "test data",
					},
				},
			},
		},
	}

	msg, err := parser.parseAssistantMessage(raw)
	if err != nil {
		t.Fatalf("parseAssistantMessage failed: %v", err)
	}

	if len(msg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Type() != "assistant" {
		t.Errorf("Expected type 'assistant', got '%s'", msg.Type())
	}

	// Check first block (text)
	textBlock, ok := msg.Content[0].(*types.TextBlock)
	if !ok {
		t.Fatalf("Expected first block to be TextBlock, got %T", msg.Content[0])
	}
	if textBlock.Text != "Hello! I can help you." {
		t.Error("Text block content mismatch")
	}

	// Check second block (tool use)
	toolBlock, ok := msg.Content[1].(*types.ToolUseBlock)
	if !ok {
		t.Fatalf("Expected second block to be ToolUseBlock, got %T", msg.Content[1])
	}
	if toolBlock.Name != "Write" {
		t.Error("Tool block name mismatch")
	}
}

func TestParseResultMessage(t *testing.T) {
	parser := NewParser(0)

	raw := map[string]any{
		"type":            "result",
		"subtype":         "completion",
		"duration_ms":     1500.0,
		"duration_api_ms": 1200.0,
		"is_error":        false,
		"num_turns":       1.0,
		"session_id":      "session_123",
		"total_cost_usd":  0.0025,
		"usage": map[string]any{
			"input_tokens":  100,
			"output_tokens": 50,
		},
		"result": "Task completed successfully",
	}

	msg, err := parser.parseResultMessage(raw)
	if err != nil {
		t.Fatalf("parseResultMessage failed: %v", err)
	}

	if msg.Type() != "result" {
		t.Errorf("Expected type 'result', got '%s'", msg.Type())
	}

	if msg.Subtype != "completion" {
		t.Errorf("Expected subtype 'completion', got '%s'", msg.Subtype)
	}

	if msg.DurationMs != 1500 {
		t.Errorf("Expected duration_ms 1500, got %d", msg.DurationMs)
	}

	if msg.TotalCostUSD == nil || *msg.TotalCostUSD != 0.0025 {
		t.Error("Expected total_cost_usd 0.0025")
	}

	if msg.Result == nil || *msg.Result != "Task completed successfully" {
		t.Error("Expected result 'Task completed successfully'")
	}
}

func TestParseMessagesBasic(t *testing.T) {
	parser := NewParser(0)

	// Create a simple data channel
	dataChan := make(chan []byte, 2)

	// Send some JSON lines
	dataChan <- []byte(`{"type": "user", "message": {"content": "Hello"}}` + "\n")
	dataChan <- []byte(`{"type": "assistant", "message": {"content": [{"type": "text", "text": "Hi there!"}]}}` + "\n")
	close(dataChan)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	msgChan, errChan := parser.ParseMessages(ctx, dataChan)

	// Collect messages
	var messages []types.Message
	var errors []error

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				msgChan = nil
				break
			}
			messages = append(messages, msg)
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				break
			}
			errors = append(errors, err)
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}

		if msgChan == nil && errChan == nil {
			break
		}
	}

	if len(errors) > 0 {
		t.Fatalf("Unexpected errors: %v", errors)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Check first message (user)
	userMsg, ok := messages[0].(*types.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", messages[0])
	}
	if userMsg.Content != "Hello" {
		t.Error("User message content mismatch")
	}

	// Check second message (assistant)
	assistantMsg, ok := messages[1].(*types.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", messages[1])
	}
	if len(assistantMsg.Content) != 1 {
		t.Error("Assistant message should have 1 content block")
	}
}
