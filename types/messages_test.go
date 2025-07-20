package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTextBlock(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"empty text", "", "text"},
		{"simple text", "Hello, world!", "text"},
		{"multiline text", "Line 1\nLine 2\nLine 3", "text"},
		{"unicode text", "Hello 世界 🌍", "text"},
		{"special characters", "!@#$%^&*(){}[]|\\:;\"'<>,.?/~`", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := &TextBlock{Text: tt.text}
			if got := tb.Type(); got != tt.want {
				t.Errorf("TextBlock.Type() = %v, want %v", got, tt.want)
			}
			if tb.Text != tt.text {
				t.Errorf("TextBlock.Text = %v, want %v", tb.Text, tt.text)
			}
		})
	}
}

func TestTextBlockJSON(t *testing.T) {
	tb := &TextBlock{Text: "test content"}
	
	// Test JSON marshaling
	data, err := json.Marshal(tb)
	if err != nil {
		t.Fatalf("Failed to marshal TextBlock: %v", err)
	}
	
	expected := `{"text":"test content"}`
	if string(data) != expected {
		t.Errorf("JSON marshal = %v, want %v", string(data), expected)
	}
	
	// Test JSON unmarshaling
	var unmarshaled TextBlock
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TextBlock: %v", err)
	}
	
	if !reflect.DeepEqual(*tb, unmarshaled) {
		t.Errorf("Unmarshaled TextBlock = %v, want %v", unmarshaled, *tb)
	}
}

func TestToolUseBlock(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		tool  string
		input map[string]any
	}{
		{
			name:  "simple tool use",
			id:    "tool_123",
			tool:  "read_file",
			input: map[string]any{"path": "/tmp/test.txt"},
		},
		{
			name:  "complex tool use",
			id:    "tool_456",
			tool:  "search_web",
			input: map[string]any{
				"query":   "golang testing",
				"limit":   10,
				"filters": []string{"recent", "relevant"},
				"options": map[string]any{"sort": "date", "lang": "en"},
			},
		},
		{
			name:  "empty input",
			id:    "tool_789",
			tool:  "list_files",
			input: map[string]any{},
		},
		{
			name:  "nil input",
			id:    "tool_000",
			tool:  "ping",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tub := &ToolUseBlock{
				ID:    tt.id,
				Name:  tt.tool,
				Input: tt.input,
			}
			
			if got := tub.Type(); got != "tool_use" {
				t.Errorf("ToolUseBlock.Type() = %v, want %v", got, "tool_use")
			}
			if tub.ID != tt.id {
				t.Errorf("ToolUseBlock.ID = %v, want %v", tub.ID, tt.id)
			}
			if tub.Name != tt.tool {
				t.Errorf("ToolUseBlock.Name = %v, want %v", tub.Name, tt.tool)
			}
			if !reflect.DeepEqual(tub.Input, tt.input) {
				t.Errorf("ToolUseBlock.Input = %v, want %v", tub.Input, tt.input)
			}
		})
	}
}

func TestToolUseBlockJSON(t *testing.T) {
	tub := &ToolUseBlock{
		ID:   "tool_123",
		Name: "read_file",
		Input: map[string]any{
			"path":     "/tmp/test.txt",
			"encoding": "utf-8",
		},
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(tub)
	if err != nil {
		t.Fatalf("Failed to marshal ToolUseBlock: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled ToolUseBlock
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ToolUseBlock: %v", err)
	}
	
	if unmarshaled.ID != tub.ID {
		t.Errorf("Unmarshaled ID = %v, want %v", unmarshaled.ID, tub.ID)
	}
	if unmarshaled.Name != tub.Name {
		t.Errorf("Unmarshaled Name = %v, want %v", unmarshaled.Name, tub.Name)
	}
	if !reflect.DeepEqual(unmarshaled.Input, tub.Input) {
		t.Errorf("Unmarshaled Input = %v, want %v", unmarshaled.Input, tub.Input)
	}
}

func TestToolResultBlock(t *testing.T) {
	content := "File read successfully"
	isError := false
	isErrorTrue := true

	tests := []struct {
		name      string
		toolUseID string
		content   *string
		isError   *bool
	}{
		{
			name:      "successful result",
			toolUseID: "tool_123",
			content:   &content,
			isError:   &isError,
		},
		{
			name:      "error result",
			toolUseID: "tool_456",
			content:   nil,
			isError:   &isErrorTrue,
		},
		{
			name:      "minimal result",
			toolUseID: "tool_789",
			content:   nil,
			isError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trb := &ToolResultBlock{
				ToolUseID: tt.toolUseID,
				Content:   tt.content,
				IsError:   tt.isError,
			}
			
			if got := trb.Type(); got != "tool_result" {
				t.Errorf("ToolResultBlock.Type() = %v, want %v", got, "tool_result")
			}
			if trb.ToolUseID != tt.toolUseID {
				t.Errorf("ToolResultBlock.ToolUseID = %v, want %v", trb.ToolUseID, tt.toolUseID)
			}
			if !reflect.DeepEqual(trb.Content, tt.content) {
				t.Errorf("ToolResultBlock.Content = %v, want %v", trb.Content, tt.content)
			}
			if !reflect.DeepEqual(trb.IsError, tt.isError) {
				t.Errorf("ToolResultBlock.IsError = %v, want %v", trb.IsError, tt.isError)
			}
		})
	}
}

func TestToolResultBlockJSON(t *testing.T) {
	content := "Operation completed"
	isError := false
	
	trb := &ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   &content,
		IsError:   &isError,
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(trb)
	if err != nil {
		t.Fatalf("Failed to marshal ToolResultBlock: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled ToolResultBlock
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ToolResultBlock: %v", err)
	}
	
	if !reflect.DeepEqual(*trb, unmarshaled) {
		t.Errorf("Unmarshaled ToolResultBlock = %v, want %v", unmarshaled, *trb)
	}
}

func TestUserMessage(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty content", ""},
		{"simple content", "Hello, Claude!"},
		{"multiline content", "Line 1\nLine 2\nLine 3"},
		{"unicode content", "Hello 世界 🌍"},
		{"long content", "This is a very long message that contains multiple sentences and spans several lines to test how the UserMessage handles larger content blocks."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UserMessage{Content: tt.content}
			
			if got := um.Type(); got != "user" {
				t.Errorf("UserMessage.Type() = %v, want %v", got, "user")
			}
			if um.Content != tt.content {
				t.Errorf("UserMessage.Content = %v, want %v", um.Content, tt.content)
			}
		})
	}
}

func TestUserMessageJSON(t *testing.T) {
	um := &UserMessage{Content: "Test message"}
	
	// Test JSON marshaling
	data, err := json.Marshal(um)
	if err != nil {
		t.Fatalf("Failed to marshal UserMessage: %v", err)
	}
	
	expected := `{"content":"Test message"}`
	if string(data) != expected {
		t.Errorf("JSON marshal = %v, want %v", string(data), expected)
	}
	
	// Test JSON unmarshaling
	var unmarshaled UserMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal UserMessage: %v", err)
	}
	
	if !reflect.DeepEqual(*um, unmarshaled) {
		t.Errorf("Unmarshaled UserMessage = %v, want %v", unmarshaled, *um)
	}
}

func TestAssistantMessage(t *testing.T) {
	tests := []struct {
		name    string
		content []ContentBlock
	}{
		{
			name:    "empty content",
			content: []ContentBlock{},
		},
		{
			name: "single text block",
			content: []ContentBlock{
				&TextBlock{Text: "Hello!"},
			},
		},
		{
			name: "multiple content blocks",
			content: []ContentBlock{
				&TextBlock{Text: "I'll help you read that file."},
				&ToolUseBlock{
					ID:    "tool_123",
					Name:  "read_file",
					Input: map[string]any{"path": "/tmp/test.txt"},
				},
			},
		},
		{
			name: "complex content blocks",
			content: []ContentBlock{
				&TextBlock{Text: "Here's the result:"},
				&ToolResultBlock{
					ToolUseID: "tool_123",
					Content:   stringPtr("File contents here"),
					IsError:   boolPtr(false),
				},
				&TextBlock{Text: "The file was read successfully."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := &AssistantMessage{Content: tt.content}
			
			if got := am.Type(); got != "assistant" {
				t.Errorf("AssistantMessage.Type() = %v, want %v", got, "assistant")
			}
			if len(am.Content) != len(tt.content) {
				t.Errorf("AssistantMessage.Content length = %v, want %v", len(am.Content), len(tt.content))
			}
		})
	}
}

func TestSystemMessage(t *testing.T) {
	tests := []struct {
		name    string
		subtype string
		data    map[string]any
	}{
		{
			name:    "simple system message",
			subtype: "info",
			data:    map[string]any{"message": "Session started"},
		},
		{
			name:    "complex system message",
			subtype: "tool_status",
			data: map[string]any{
				"tool_id":     "tool_123",
				"status":      "completed",
				"duration_ms": 1500,
				"metadata":    map[string]any{"size": 1024, "encoding": "utf-8"},
			},
		},
		{
			name:    "minimal system message",
			subtype: "ping",
			data:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &SystemMessage{
				Subtype: tt.subtype,
				Data:    tt.data,
			}
			
			if got := sm.Type(); got != "system" {
				t.Errorf("SystemMessage.Type() = %v, want %v", got, "system")
			}
			if sm.Subtype != tt.subtype {
				t.Errorf("SystemMessage.Subtype = %v, want %v", sm.Subtype, tt.subtype)
			}
			if !reflect.DeepEqual(sm.Data, tt.data) {
				t.Errorf("SystemMessage.Data = %v, want %v", sm.Data, tt.data)
			}
		})
	}
}

func TestResultMessage(t *testing.T) {
	totalCost := 0.0025
	result := "Query completed successfully"
	
	tests := []struct {
		name          string
		subtype       string
		durationMs    int
		durationAPIMs int
		isError       bool
		numTurns      int
		sessionID     string
		totalCostUSD  *float64
		usage         map[string]any
		result        *string
	}{
		{
			name:          "successful result",
			subtype:       "completion",
			durationMs:    5000,
			durationAPIMs: 3000,
			isError:       false,
			numTurns:      3,
			sessionID:     "session_123",
			totalCostUSD:  &totalCost,
			usage: map[string]any{
				"input_tokens":  1500,
				"output_tokens": 800,
				"cache_hits":    5,
			},
			result: &result,
		},
		{
			name:          "error result",
			subtype:       "error",
			durationMs:    1000,
			durationAPIMs: 500,
			isError:       true,
			numTurns:      1,
			sessionID:     "session_456",
			totalCostUSD:  nil,
			usage:         nil,
			result:        nil,
		},
		{
			name:          "minimal result",
			subtype:       "timeout",
			durationMs:    30000,
			durationAPIMs: 0,
			isError:       true,
			numTurns:      0,
			sessionID:     "session_789",
			totalCostUSD:  nil,
			usage:         map[string]any{},
			result:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := &ResultMessage{
				Subtype:       tt.subtype,
				DurationMs:    tt.durationMs,
				DurationAPIMs: tt.durationAPIMs,
				IsError:       tt.isError,
				NumTurns:      tt.numTurns,
				SessionID:     tt.sessionID,
				TotalCostUSD:  tt.totalCostUSD,
				Usage:         tt.usage,
				Result:        tt.result,
			}
			
			if got := rm.Type(); got != "result" {
				t.Errorf("ResultMessage.Type() = %v, want %v", got, "result")
			}
			if rm.Subtype != tt.subtype {
				t.Errorf("ResultMessage.Subtype = %v, want %v", rm.Subtype, tt.subtype)
			}
			if rm.DurationMs != tt.durationMs {
				t.Errorf("ResultMessage.DurationMs = %v, want %v", rm.DurationMs, tt.durationMs)
			}
			if rm.IsError != tt.isError {
				t.Errorf("ResultMessage.IsError = %v, want %v", rm.IsError, tt.isError)
			}
			if !reflect.DeepEqual(rm.TotalCostUSD, tt.totalCostUSD) {
				t.Errorf("ResultMessage.TotalCostUSD = %v, want %v", rm.TotalCostUSD, tt.totalCostUSD)
			}
		})
	}
}

func TestResultMessageJSON(t *testing.T) {
	totalCost := 0.0025
	result := "Success"
	
	rm := &ResultMessage{
		Subtype:       "completion",
		DurationMs:    5000,
		DurationAPIMs: 3000,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session_123",
		TotalCostUSD:  &totalCost,
		Usage: map[string]any{
			"input_tokens":  1500,
			"output_tokens": 800,
		},
		Result: &result,
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(rm)
	if err != nil {
		t.Fatalf("Failed to marshal ResultMessage: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled ResultMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ResultMessage: %v", err)
	}
	
	if unmarshaled.Subtype != rm.Subtype {
		t.Errorf("Unmarshaled Subtype = %v, want %v", unmarshaled.Subtype, rm.Subtype)
	}
	if unmarshaled.DurationMs != rm.DurationMs {
		t.Errorf("Unmarshaled DurationMs = %v, want %v", unmarshaled.DurationMs, rm.DurationMs)
	}
	if unmarshaled.IsError != rm.IsError {
		t.Errorf("Unmarshaled IsError = %v, want %v", unmarshaled.IsError, rm.IsError)
	}
}

func TestPermissionModeConstants(t *testing.T) {
	tests := []struct {
		name string
		mode PermissionMode
		want string
	}{
		{"default mode", PermissionModeDefault, "default"},
		{"accept edits mode", PermissionModeAcceptEdits, "acceptEdits"},
		{"bypass permissions mode", PermissionModeBypassPermissions, "bypassPermissions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.want {
				t.Errorf("PermissionMode = %v, want %v", string(tt.mode), tt.want)
			}
		})
	}
}

func TestContentBlockInterface(t *testing.T) {
	var block ContentBlock
	
	// Test TextBlock implements ContentBlock
	block = &TextBlock{Text: "test"}
	if block.Type() != "text" {
		t.Errorf("TextBlock.Type() = %v, want %v", block.Type(), "text")
	}
	
	// Test ToolUseBlock implements ContentBlock
	block = &ToolUseBlock{ID: "1", Name: "test", Input: nil}
	if block.Type() != "tool_use" {
		t.Errorf("ToolUseBlock.Type() = %v, want %v", block.Type(), "tool_use")
	}
	
	// Test ToolResultBlock implements ContentBlock
	block = &ToolResultBlock{ToolUseID: "1"}
	if block.Type() != "tool_result" {
		t.Errorf("ToolResultBlock.Type() = %v, want %v", block.Type(), "tool_result")
	}
}

func TestMessageInterface(t *testing.T) {
	var message Message
	
	// Test UserMessage implements Message
	message = &UserMessage{Content: "test"}
	if message.Type() != "user" {
		t.Errorf("UserMessage.Type() = %v, want %v", message.Type(), "user")
	}
	
	// Test AssistantMessage implements Message
	message = &AssistantMessage{Content: []ContentBlock{}}
	if message.Type() != "assistant" {
		t.Errorf("AssistantMessage.Type() = %v, want %v", message.Type(), "assistant")
	}
	
	// Test SystemMessage implements Message
	message = &SystemMessage{Subtype: "test"}
	if message.Type() != "system" {
		t.Errorf("SystemMessage.Type() = %v, want %v", message.Type(), "system")
	}
	
	// Test ResultMessage implements Message
	message = &ResultMessage{Subtype: "test"}
	if message.Type() != "result" {
		t.Errorf("ResultMessage.Type() = %v, want %v", message.Type(), "result")
	}
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}