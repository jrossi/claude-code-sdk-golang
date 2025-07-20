package types

// McpServerConfig represents configuration for an MCP (Model Context Protocol) server.
// Different server types (stdio, SSE, HTTP) implement this interface.
type McpServerConfig interface {
	// ServerType returns the server type identifier ("stdio", "sse", or "http").
	ServerType() string
}

// StdioServerConfig represents an MCP server that communicates via stdio.
type StdioServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// ServerType returns "stdio" as the server type identifier.
func (s *StdioServerConfig) ServerType() string {
	return "stdio"
}

// SSEServerConfig represents an MCP server that communicates via Server-Sent Events.
type SSEServerConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ServerType returns "sse" as the server type identifier.
func (s *SSEServerConfig) ServerType() string {
	return "sse"
}

// HTTPServerConfig represents an MCP server that communicates via HTTP.
type HTTPServerConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ServerType returns "http" as the server type identifier.
func (s *HTTPServerConfig) ServerType() string {
	return "http"
}

// Options contains configuration options for Claude Code queries.
type Options struct {
	// AllowedTools specifies which tools Claude is allowed to use.
	// If empty, default tool restrictions apply.
	AllowedTools []string `json:"allowedTools,omitempty"`

	// MaxThinkingTokens sets the maximum number of tokens for internal reasoning.
	// Default is 8000.
	MaxThinkingTokens int `json:"maxThinkingTokens,omitempty"`

	// SystemPrompt sets a custom system prompt for the conversation.
	SystemPrompt *string `json:"systemPrompt,omitempty"`

	// AppendSystemPrompt appends additional text to the existing system prompt.
	AppendSystemPrompt *string `json:"appendSystemPrompt,omitempty"`

	// McpTools specifies which MCP (Model Context Protocol) tools to enable.
	McpTools []string `json:"mcpTools,omitempty"`

	// McpServers configures MCP servers by name.
	McpServers map[string]McpServerConfig `json:"mcpServers,omitempty"`

	// PermissionMode controls how tool permissions are handled.
	PermissionMode *PermissionMode `json:"permissionMode,omitempty"`

	// ContinueConversation indicates whether to continue an existing conversation.
	ContinueConversation bool `json:"continueConversation,omitempty"`

	// Resume specifies a session ID to resume from.
	Resume *string `json:"resume,omitempty"`

	// MaxTurns limits the number of conversation turns.
	MaxTurns *int `json:"maxTurns,omitempty"`

	// DisallowedTools specifies which tools Claude is explicitly not allowed to use.
	DisallowedTools []string `json:"disallowedTools,omitempty"`

	// Model specifies which Claude model to use.
	Model *string `json:"model,omitempty"`

	// PermissionPromptToolName specifies which tool to use for permission prompts.
	PermissionPromptToolName *string `json:"permissionPromptToolName,omitempty"`

	// Cwd sets the working directory for the Claude Code session.
	Cwd *string `json:"cwd,omitempty"`
}

// NewOptions creates a new Options instance with sensible defaults.
func NewOptions() *Options {
	return &Options{
		AllowedTools:         []string{},
		MaxThinkingTokens:    8000,
		McpTools:             []string{},
		McpServers:           make(map[string]McpServerConfig),
		ContinueConversation: false,
		DisallowedTools:      []string{},
	}
}

// WithSystemPrompt sets the system prompt for the options.
func (o *Options) WithSystemPrompt(prompt string) *Options {
	o.SystemPrompt = &prompt
	return o
}

// WithAppendSystemPrompt sets the append system prompt for the options.
func (o *Options) WithAppendSystemPrompt(prompt string) *Options {
	o.AppendSystemPrompt = &prompt
	return o
}

// WithAllowedTools sets the allowed tools for the options.
func (o *Options) WithAllowedTools(tools ...string) *Options {
	o.AllowedTools = tools
	return o
}

// WithDisallowedTools sets the disallowed tools for the options.
func (o *Options) WithDisallowedTools(tools ...string) *Options {
	o.DisallowedTools = tools
	return o
}

// WithPermissionMode sets the permission mode for the options.
func (o *Options) WithPermissionMode(mode PermissionMode) *Options {
	o.PermissionMode = &mode
	return o
}

// WithMaxTurns sets the maximum number of turns for the options.
func (o *Options) WithMaxTurns(turns int) *Options {
	o.MaxTurns = &turns
	return o
}

// WithModel sets the model for the options.
func (o *Options) WithModel(model string) *Options {
	o.Model = &model
	return o
}

// WithCwd sets the working directory for the options.
func (o *Options) WithCwd(cwd string) *Options {
	o.Cwd = &cwd
	return o
}

// WithContinueConversation enables conversation continuation.
func (o *Options) WithContinueConversation() *Options {
	o.ContinueConversation = true
	return o
}

// WithResume sets the session ID to resume from.
func (o *Options) WithResume(sessionID string) *Options {
	o.Resume = &sessionID
	return o
}

// AddMcpServer adds an MCP server configuration.
func (o *Options) AddMcpServer(name string, config McpServerConfig) *Options {
	if o.McpServers == nil {
		o.McpServers = make(map[string]McpServerConfig)
	}
	o.McpServers[name] = config
	return o
}

// AddMcpTool adds an MCP tool to the enabled tools list.
func (o *Options) AddMcpTool(tool string) *Options {
	o.McpTools = append(o.McpTools, tool)
	return o
}
