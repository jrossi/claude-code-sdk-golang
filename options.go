package claudecode

import "github.com/jrossi/claude-code-sdk-golang/internal/types"

// Re-export types from internal package to maintain clean public API
type (
	// McpServerConfig represents configuration for an MCP (Model Context Protocol) server.
	McpServerConfig = types.McpServerConfig

	// StdioServerConfig represents an MCP server that communicates via stdio.
	StdioServerConfig = types.StdioServerConfig

	// SSEServerConfig represents an MCP server that communicates via Server-Sent Events.
	SSEServerConfig = types.SSEServerConfig

	// HTTPServerConfig represents an MCP server that communicates via HTTP.
	HTTPServerConfig = types.HTTPServerConfig

	// Options contains configuration options for Claude Code queries.
	Options = types.Options

	// PermissionMode defines the permission handling mode for tool execution.
	PermissionMode = types.PermissionMode
)

// Re-export permission mode constants
const (
	// PermissionModeDefault uses the CLI's default permission prompting behavior.
	PermissionModeDefault = types.PermissionModeDefault

	// PermissionModeAcceptEdits automatically accepts file edit operations.
	PermissionModeAcceptEdits = types.PermissionModeAcceptEdits

	// PermissionModeBypassPermissions allows all tools without prompting.
	// Use with caution as this bypasses all safety checks.
	PermissionModeBypassPermissions = types.PermissionModeBypassPermissions
)

// Re-export constructor function
var NewOptions = types.NewOptions
