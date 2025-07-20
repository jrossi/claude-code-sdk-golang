package claudecode

import (
	types2 "github.com/jrossi/claude-code-sdk-golang/types"
)

// Re-export types from internal package to maintain clean public API
type (
	// McpServerConfig represents configuration for an MCP (Model Context Protocol) server.
	McpServerConfig = types2.McpServerConfig

	// StdioServerConfig represents an MCP server that communicates via stdio.
	StdioServerConfig = types2.StdioServerConfig

	// SSEServerConfig represents an MCP server that communicates via Server-Sent Events.
	SSEServerConfig = types2.SSEServerConfig

	// HTTPServerConfig represents an MCP server that communicates via HTTP.
	HTTPServerConfig = types2.HTTPServerConfig

	// Options contains configuration options for Claude Code queries.
	Options = types2.Options

	// PermissionMode defines the permission handling mode for tool execution.
	PermissionMode = types2.PermissionMode
)

// Re-export permission mode constants
const (
	// PermissionModeDefault uses the CLI's default permission prompting behavior.
	PermissionModeDefault = types2.PermissionModeDefault

	// PermissionModeAcceptEdits automatically accepts file edit operations.
	PermissionModeAcceptEdits = types2.PermissionModeAcceptEdits

	// PermissionModeBypassPermissions allows all tools without prompting.
	// Use with caution as this bypasses all safety checks.
	PermissionModeBypassPermissions = types2.PermissionModeBypassPermissions
)

// Re-export constructor function
var NewOptions = types2.NewOptions
