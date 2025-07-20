// Package parser provides JSON streaming parsing for Claude Code CLI output.
package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jrossi/claude-code-sdk-golang/types"
	"strings"
)

const (
	// DefaultMaxBufferSize is the default maximum size for the JSON parsing buffer.
	DefaultMaxBufferSize = 1024 * 1024 // 1MB
)

// Parser handles streaming JSON parsing from Claude Code CLI output.
type Parser struct {
	// maxBufferSize limits the size of the internal buffer to prevent memory issues.
	maxBufferSize int

	// buffer accumulates partial JSON data until a complete message can be parsed.
	buffer []byte
}

// NewParser creates a new JSON parser with the specified maximum buffer size.
// If maxBufferSize is 0, DefaultMaxBufferSize will be used.
func NewParser(maxBufferSize int) *Parser {
	if maxBufferSize <= 0 {
		maxBufferSize = DefaultMaxBufferSize
	}

	return &Parser{
		maxBufferSize: maxBufferSize,
		buffer:        make([]byte, 0, 1024), // Start with 1KB capacity
	}
}

// ParseMessages processes a stream of raw bytes and returns parsed messages.
// This is the foundation - full implementation will be completed in Phase 4.
func (p *Parser) ParseMessages(ctx context.Context, data <-chan []byte) (<-chan types.Message, <-chan error) {
	msgChan := make(chan types.Message, 10)
	errChan := make(chan error, 5)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-data:
				if !ok {
					// Data channel closed, process remaining buffer if any
					if len(p.buffer) > 0 {
						if err := p.processRemainingBuffer(msgChan, errChan); err != nil {
							errChan <- err
						}
					}
					return
				}

				// Process the chunk (full implementation in Phase 4)
				if err := p.processChunk(chunk, msgChan, errChan); err != nil {
					errChan <- err
				}
			}
		}
	}()

	return msgChan, errChan
}

// processChunk handles a chunk of data from the CLI output.
// This implementation handles complex buffering scenarios including:
// - Multiple JSON objects on a single line
// - JSON objects split across multiple chunks
// - Mixed complete and partial JSON messages
// - Large JSON payloads that exceed single buffer reads
func (p *Parser) processChunk(chunk []byte, msgChan chan<- types.Message, errChan chan<- error) error {
	// Append new data to buffer
	p.buffer = append(p.buffer, chunk...)

	// Check buffer size limit to prevent memory exhaustion
	if len(p.buffer) > p.maxBufferSize {
		// Save the error data before clearing
		truncatedData := string(p.buffer[:100]) + "..."
		p.buffer = p.buffer[:0] // Clear buffer to recover

		return fmt.Errorf("JSON message exceeded maximum buffer size of %d bytes: buffer overflow: data starts with %q",
			p.maxBufferSize,
			truncatedData,
		)
	}

	// Process all complete JSON messages in the buffer
	return p.extractCompleteMessages(msgChan, errChan)
}

// extractCompleteMessages processes the buffer to extract all complete JSON messages.
// This implements robust buffering that handles edge cases from the Python SDK tests:
// - Multiple JSON objects separated by newlines on the same line
// - JSON with embedded newlines in string values
// - Large JSON split across multiple reads
// - Mixed complete and partial JSON messages
func (p *Parser) extractCompleteMessages(msgChan chan<- types.Message, errChan chan<- error) error {
	// Handle multiple JSON objects that may be concatenated on single lines
	// Split by newlines first, but be careful about JSON with embedded newlines
	var processedBytes int
	var remainingBuffer []byte

	// Process line by line, but handle JSON that spans multiple lines
	for {
		if processedBytes >= len(p.buffer) {
			break
		}

		// Look for the next complete JSON object starting from current position
		start := processedBytes
		jsonStart := -1
		braceCount := 0
		inString := false
		escaped := false

		// Find the start of the next JSON object
		for i := start; i < len(p.buffer); i++ {
			b := p.buffer[i]

			if !inString && b == '{' {
				if jsonStart == -1 {
					jsonStart = i
				}
				braceCount++
			} else if !inString && b == '}' {
				braceCount--
				if braceCount == 0 && jsonStart != -1 {
					// Found complete JSON object
					jsonBytes := p.buffer[jsonStart : i+1]
					jsonStr := string(jsonBytes)

					// Try to parse this JSON object
					msg, err := p.parseMessage(jsonStr)
					if err != nil {
						// Send error but continue processing
						errChan <- fmt.Errorf("JSON decode error: %s: %w", jsonStr, err)
					} else if msg != nil {
						msgChan <- msg
					}

					processedBytes = i + 1
					jsonStart = -1
					braceCount = 0
					break
				}
			} else if b == '"' && !escaped {
				inString = !inString
			}

			// Handle escape sequences in strings
			if inString {
				escaped = !escaped && b == '\\'
			} else {
				escaped = false
			}
		}

		// If we didn't find a complete JSON object, we need more data
		if jsonStart != -1 && braceCount > 0 {
			// Incomplete JSON object - keep it in buffer
			remainingBuffer = p.buffer[jsonStart:]
			break
		}

		// If we didn't find any JSON start, skip non-JSON content
		if jsonStart == -1 {
			// Look for the next '{' or end of buffer
			found := false
			for i := processedBytes; i < len(p.buffer); i++ {
				if p.buffer[i] == '{' {
					processedBytes = i
					found = true
					break
				}
			}
			if !found {
				// No more JSON in buffer
				break
			}
		}
	}

	// Update buffer with any remaining incomplete JSON
	if len(remainingBuffer) > 0 {
		p.buffer = remainingBuffer
	} else {
		p.buffer = p.buffer[:0]
	}

	return nil
}

// processRemainingBuffer processes any remaining data in the buffer when input ends.
func (p *Parser) processRemainingBuffer(msgChan chan<- types.Message, errChan chan<- error) error {
	if len(p.buffer) == 0 {
		return nil
	}

	bufferStr := strings.TrimSpace(string(p.buffer))
	if bufferStr == "" {
		return nil
	}

	msg, err := p.parseMessage(bufferStr)
	if err != nil {
		return fmt.Errorf("JSON decode error: %s: %w", bufferStr, err)
	}

	if msg != nil {
		msgChan <- msg
	}

	return nil
}

// parseMessage parses a single JSON line into a Message.
func (p *Parser) parseMessage(line string) (types.Message, error) {
	// Parse as generic JSON first
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, err
	}

	// Determine message type
	msgType, ok := raw["type"].(string)
	if !ok {
		return nil, fmt.Errorf("message missing 'type' field")
	}

	// Parse based on type
	switch msgType {
	case "user":
		return p.parseUserMessage(raw)
	case "assistant":
		return p.parseAssistantMessage(raw)
	case "system":
		return p.parseSystemMessage(raw)
	case "result":
		return p.parseResultMessage(raw)
	default:
		// Unknown message type, skip silently for forward compatibility
		return nil, nil
	}
}

// parseUserMessage parses a user message from raw JSON data.
func (p *Parser) parseUserMessage(raw map[string]any) (*types.UserMessage, error) {
	message, ok := raw["message"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("user message missing 'message' field")
	}

	// Handle both string content and array content (for tool results)
	if contentStr, ok := message["content"].(string); ok {
		return &types.UserMessage{Content: contentStr}, nil
	}

	if contentArray, ok := message["content"].([]any); ok {
		// For tool result arrays, create a summary string
		return &types.UserMessage{Content: fmt.Sprintf("Tool results: %d items", len(contentArray))}, nil
	}

	return nil, fmt.Errorf("user message missing 'content' field")
}

// parseAssistantMessage parses an assistant message from raw JSON data.
func (p *Parser) parseAssistantMessage(raw map[string]any) (*types.AssistantMessage, error) {
	message, ok := raw["message"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("assistant message missing 'message' field")
	}

	contentArray, ok := message["content"].([]any)
	if !ok {
		return nil, fmt.Errorf("assistant message missing 'content' array")
	}

	var contentBlocks []types.ContentBlock
	for _, blockData := range contentArray {
		block, ok := blockData.(map[string]any)
		if !ok {
			continue
		}

		contentBlock, err := p.parseContentBlock(block)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content block: %w", err)
		}

		if contentBlock != nil {
			contentBlocks = append(contentBlocks, contentBlock)
		}
	}

	return &types.AssistantMessage{Content: contentBlocks}, nil
}

// parseContentBlock parses a content block from raw JSON data.
func (p *Parser) parseContentBlock(block map[string]any) (types.ContentBlock, error) {
	blockType, ok := block["type"].(string)
	if !ok {
		return nil, fmt.Errorf("content block missing 'type' field")
	}

	switch blockType {
	case "text":
		text, ok := block["text"].(string)
		if !ok {
			return nil, fmt.Errorf("text block missing 'text' field")
		}
		return &types.TextBlock{Text: text}, nil

	case "tool_use":
		id, ok := block["id"].(string)
		if !ok {
			return nil, fmt.Errorf("tool_use block missing 'id' field")
		}
		name, ok := block["name"].(string)
		if !ok {
			return nil, fmt.Errorf("tool_use block missing 'name' field")
		}
		input, ok := block["input"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool_use block missing 'input' field")
		}
		return &types.ToolUseBlock{ID: id, Name: name, Input: input}, nil

	case "tool_result":
		toolUseID, ok := block["tool_use_id"].(string)
		if !ok {
			return nil, fmt.Errorf("tool_result block missing 'tool_use_id' field")
		}

		result := &types.ToolResultBlock{ToolUseID: toolUseID}

		if content, exists := block["content"]; exists && content != nil {
			if contentStr, ok := content.(string); ok {
				result.Content = &contentStr
			}
		}

		if isError, exists := block["is_error"]; exists && isError != nil {
			if isErrorBool, ok := isError.(bool); ok {
				result.IsError = &isErrorBool
			}
		}

		return result, nil

	default:
		// Unknown content block type, skip for forward compatibility
		return nil, nil
	}
}

// parseSystemMessage parses a system message from raw JSON data.
func (p *Parser) parseSystemMessage(raw map[string]any) (*types.SystemMessage, error) {
	subtype, ok := raw["subtype"].(string)
	if !ok {
		return nil, fmt.Errorf("system message missing 'subtype' field")
	}

	return &types.SystemMessage{
		Subtype: subtype,
		Data:    raw, // Include all raw data for system messages
	}, nil
}

// parseResultMessage parses a result message from raw JSON data.
func (p *Parser) parseResultMessage(raw map[string]any) (*types.ResultMessage, error) {
	subtype, ok := raw["subtype"].(string)
	if !ok {
		return nil, fmt.Errorf("result message missing 'subtype' field")
	}

	result := &types.ResultMessage{Subtype: subtype}

	// Parse required integer fields
	if val, ok := raw["duration_ms"].(float64); ok {
		result.DurationMs = int(val)
	}
	if val, ok := raw["duration_api_ms"].(float64); ok {
		result.DurationAPIMs = int(val)
	}
	if val, ok := raw["is_error"].(bool); ok {
		result.IsError = val
	}
	if val, ok := raw["num_turns"].(float64); ok {
		result.NumTurns = int(val)
	}
	if val, ok := raw["session_id"].(string); ok {
		result.SessionID = val
	}

	// Parse optional fields
	if val, ok := raw["total_cost_usd"].(float64); ok {
		result.TotalCostUSD = &val
	}
	if val, ok := raw["usage"].(map[string]any); ok {
		result.Usage = val
	}
	if val, ok := raw["result"].(string); ok {
		result.Result = &val
	}

	return result, nil
}
