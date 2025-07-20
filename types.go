package claudecode

import "github.com/jrossi/claude-code-sdk-golang/internal/types"

// Re-export message and content block types from internal package
type (
	// ContentBlock represents a piece of content within a message.
	ContentBlock = types.ContentBlock

	// TextBlock represents a text content block.
	TextBlock = types.TextBlock

	// ToolUseBlock represents a tool use content block.
	ToolUseBlock = types.ToolUseBlock

	// ToolResultBlock represents a tool result content block.
	ToolResultBlock = types.ToolResultBlock

	// Message represents a message in the conversation.
	Message = types.Message

	// UserMessage represents a message from the user.
	UserMessage = types.UserMessage

	// AssistantMessage represents a message from the assistant with content blocks.
	AssistantMessage = types.AssistantMessage

	// SystemMessage represents a system message with metadata.
	SystemMessage = types.SystemMessage

	// ResultMessage represents a result message with cost and usage information.
	ResultMessage = types.ResultMessage
)
