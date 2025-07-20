package types

// ContentBlock represents a piece of content within a message.
// Implementations include TextBlock, ToolUseBlock, and ToolResultBlock.
type ContentBlock interface {
	Type() string
}

// TextBlock represents a text content block.
type TextBlock struct {
	Text string `json:"text"`
}

// Type returns the content block type identifier.
func (tb *TextBlock) Type() string {
	return "text"
}

// ToolUseBlock represents a tool use content block.
type ToolUseBlock struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Input contains arbitrary JSON data that varies by tool.
	// Using map[string]any here is necessary to handle dynamic tool parameters.
	Input map[string]any `json:"input"`
}

// Type returns the content block type identifier.
func (tub *ToolUseBlock) Type() string {
	return "tool_use"
}

// ToolResultBlock represents a tool result content block.
type ToolResultBlock struct {
	ToolUseID string  `json:"tool_use_id"`
	Content   *string `json:"content,omitempty"`
	IsError   *bool   `json:"is_error,omitempty"`
}

// Type returns the content block type identifier.
func (trb *ToolResultBlock) Type() string {
	return "tool_result"
}

// Message represents a message in the conversation.
// Implementations include UserMessage, AssistantMessage, SystemMessage, and ResultMessage.
type Message interface {
	Type() string
}

// UserMessage represents a message from the user.
type UserMessage struct {
	Content string `json:"content"`
}

// Type returns the message type identifier.
func (um *UserMessage) Type() string {
	return "user"
}

// AssistantMessage represents a message from the assistant with content blocks.
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
}

// Type returns the message type identifier.
func (am *AssistantMessage) Type() string {
	return "assistant"
}

// SystemMessage represents a system message with metadata.
type SystemMessage struct {
	Subtype string `json:"subtype"`
	// Data contains arbitrary JSON metadata that varies by system message type.
	// Using map[string]any here is necessary to handle dynamic system message data.
	Data map[string]any `json:"data,omitempty"`
}

// Type returns the message type identifier.
func (sm *SystemMessage) Type() string {
	return "system"
}

// ResultMessage represents a result message with cost and usage information.
type ResultMessage struct {
	Subtype       string   `json:"subtype"`
	DurationMs    int      `json:"duration_ms"`
	DurationAPIMs int      `json:"duration_api_ms"`
	IsError       bool     `json:"is_error"`
	NumTurns      int      `json:"num_turns"`
	SessionID     string   `json:"session_id"`
	TotalCostUSD  *float64 `json:"total_cost_usd,omitempty"`
	// Usage contains arbitrary JSON usage statistics that vary by API provider.
	// Using map[string]any here is necessary to handle dynamic usage metrics.
	Usage  map[string]any `json:"usage,omitempty"`
	Result *string        `json:"result,omitempty"`
}

// Type returns the message type identifier.
func (rm *ResultMessage) Type() string {
	return "result"
}

// PermissionMode defines the permission handling mode for tool execution.
type PermissionMode string

const (
	// PermissionModeDefault uses the CLI's default permission prompting behavior.
	PermissionModeDefault PermissionMode = "default"

	// PermissionModeAcceptEdits automatically accepts file edit operations.
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"

	// PermissionModeBypassPermissions allows all tools without prompting.
	// Use with caution as this bypasses all safety checks.
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
)
