package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStdioServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		env     map[string]string
	}{
		{
			name:    "simple command",
			command: "node",
			args:    []string{"server.js"},
			env:     nil,
		},
		{
			name:    "complex command with env",
			command: "python",
			args:    []string{"-m", "mcp_server", "--port", "8080"},
			env:     map[string]string{"DEBUG": "1", "PORT": "8080"},
		},
		{
			name:    "minimal config",
			command: "mcp-server",
			args:    nil,
			env:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &StdioServerConfig{
				Command: tt.command,
				Args:    tt.args,
				Env:     tt.env,
			}
			
			if config.ServerType() != "stdio" {
				t.Errorf("StdioServerConfig.ServerType() = %v, want %v", config.ServerType(), "stdio")
			}
			if config.Command != tt.command {
				t.Errorf("StdioServerConfig.Command = %v, want %v", config.Command, tt.command)
			}
			if !reflect.DeepEqual(config.Args, tt.args) {
				t.Errorf("StdioServerConfig.Args = %v, want %v", config.Args, tt.args)
			}
			if !reflect.DeepEqual(config.Env, tt.env) {
				t.Errorf("StdioServerConfig.Env = %v, want %v", config.Env, tt.env)
			}
		})
	}
}

func TestStdioServerConfigJSON(t *testing.T) {
	config := &StdioServerConfig{
		Command: "node",
		Args:    []string{"server.js", "--port", "8080"},
		Env:     map[string]string{"DEBUG": "1"},
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal StdioServerConfig: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled StdioServerConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal StdioServerConfig: %v", err)
	}
	
	if !reflect.DeepEqual(*config, unmarshaled) {
		t.Errorf("Unmarshaled StdioServerConfig = %v, want %v", unmarshaled, *config)
	}
}

func TestSSEServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers map[string]string
	}{
		{
			name:    "simple SSE config",
			url:     "https://api.example.com/sse",
			headers: nil,
		},
		{
			name: "SSE config with headers",
			url:  "https://api.example.com/sse",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"User-Agent":    "Claude-SDK/1.0",
			},
		},
		{
			name:    "minimal SSE config",
			url:     "http://localhost:8080/sse",
			headers: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SSEServerConfig{
				URL:     tt.url,
				Headers: tt.headers,
			}
			
			if config.ServerType() != "sse" {
				t.Errorf("SSEServerConfig.ServerType() = %v, want %v", config.ServerType(), "sse")
			}
			if config.URL != tt.url {
				t.Errorf("SSEServerConfig.URL = %v, want %v", config.URL, tt.url)
			}
			if !reflect.DeepEqual(config.Headers, tt.headers) {
				t.Errorf("SSEServerConfig.Headers = %v, want %v", config.Headers, tt.headers)
			}
		})
	}
}

func TestSSEServerConfigJSON(t *testing.T) {
	config := &SSEServerConfig{
		URL: "https://api.example.com/sse",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal SSEServerConfig: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled SSEServerConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SSEServerConfig: %v", err)
	}
	
	if !reflect.DeepEqual(*config, unmarshaled) {
		t.Errorf("Unmarshaled SSEServerConfig = %v, want %v", unmarshaled, *config)
	}
}

func TestHTTPServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers map[string]string
	}{
		{
			name:    "simple HTTP config",
			url:     "https://api.example.com/mcp",
			headers: nil,
		},
		{
			name: "HTTP config with headers",
			url:  "https://api.example.com/mcp",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"Content-Type":  "application/json",
			},
		},
		{
			name:    "localhost HTTP config",
			url:     "http://localhost:3000/mcp",
			headers: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HTTPServerConfig{
				URL:     tt.url,
				Headers: tt.headers,
			}
			
			if config.ServerType() != "http" {
				t.Errorf("HTTPServerConfig.ServerType() = %v, want %v", config.ServerType(), "http")
			}
			if config.URL != tt.url {
				t.Errorf("HTTPServerConfig.URL = %v, want %v", config.URL, tt.url)
			}
			if !reflect.DeepEqual(config.Headers, tt.headers) {
				t.Errorf("HTTPServerConfig.Headers = %v, want %v", config.Headers, tt.headers)
			}
		})
	}
}

func TestHTTPServerConfigJSON(t *testing.T) {
	config := &HTTPServerConfig{
		URL: "https://api.example.com/mcp",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal HTTPServerConfig: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled HTTPServerConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal HTTPServerConfig: %v", err)
	}
	
	if !reflect.DeepEqual(*config, unmarshaled) {
		t.Errorf("Unmarshaled HTTPServerConfig = %v, want %v", unmarshaled, *config)
	}
}

func TestMcpServerConfigInterface(t *testing.T) {
	var config McpServerConfig
	
	// Test StdioServerConfig implements McpServerConfig
	config = &StdioServerConfig{Command: "node"}
	if config.ServerType() != "stdio" {
		t.Errorf("StdioServerConfig.ServerType() = %v, want %v", config.ServerType(), "stdio")
	}
	
	// Test SSEServerConfig implements McpServerConfig
	config = &SSEServerConfig{URL: "https://example.com"}
	if config.ServerType() != "sse" {
		t.Errorf("SSEServerConfig.ServerType() = %v, want %v", config.ServerType(), "sse")
	}
	
	// Test HTTPServerConfig implements McpServerConfig
	config = &HTTPServerConfig{URL: "https://example.com"}
	if config.ServerType() != "http" {
		t.Errorf("HTTPServerConfig.ServerType() = %v, want %v", config.ServerType(), "http")
	}
}

func TestNewOptions(t *testing.T) {
	opts := NewOptions()
	
	// Test default values
	if opts.AllowedTools == nil {
		t.Error("NewOptions().AllowedTools should not be nil")
	}
	if len(opts.AllowedTools) != 0 {
		t.Errorf("NewOptions().AllowedTools length = %v, want 0", len(opts.AllowedTools))
	}
	if opts.MaxThinkingTokens != 8000 {
		t.Errorf("NewOptions().MaxThinkingTokens = %v, want 8000", opts.MaxThinkingTokens)
	}
	if opts.McpTools == nil {
		t.Error("NewOptions().McpTools should not be nil")
	}
	if len(opts.McpTools) != 0 {
		t.Errorf("NewOptions().McpTools length = %v, want 0", len(opts.McpTools))
	}
	if opts.McpServers == nil {
		t.Error("NewOptions().McpServers should not be nil")
	}
	if len(opts.McpServers) != 0 {
		t.Errorf("NewOptions().McpServers length = %v, want 0", len(opts.McpServers))
	}
	if opts.ContinueConversation != false {
		t.Errorf("NewOptions().ContinueConversation = %v, want false", opts.ContinueConversation)
	}
	if opts.DisallowedTools == nil {
		t.Error("NewOptions().DisallowedTools should not be nil")
	}
	if len(opts.DisallowedTools) != 0 {
		t.Errorf("NewOptions().DisallowedTools length = %v, want 0", len(opts.DisallowedTools))
	}
	
	// Test that optional fields are nil by default
	if opts.SystemPrompt != nil {
		t.Errorf("NewOptions().SystemPrompt = %v, want nil", opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != nil {
		t.Errorf("NewOptions().AppendSystemPrompt = %v, want nil", opts.AppendSystemPrompt)
	}
	if opts.PermissionMode != nil {
		t.Errorf("NewOptions().PermissionMode = %v, want nil", opts.PermissionMode)
	}
	if opts.Resume != nil {
		t.Errorf("NewOptions().Resume = %v, want nil", opts.Resume)
	}
	if opts.MaxTurns != nil {
		t.Errorf("NewOptions().MaxTurns = %v, want nil", opts.MaxTurns)
	}
	if opts.Model != nil {
		t.Errorf("NewOptions().Model = %v, want nil", opts.Model)
	}
	if opts.PermissionPromptToolName != nil {
		t.Errorf("NewOptions().PermissionPromptToolName = %v, want nil", opts.PermissionPromptToolName)
	}
	if opts.Cwd != nil {
		t.Errorf("NewOptions().Cwd = %v, want nil", opts.Cwd)
	}
}

func TestOptionsWithSystemPrompt(t *testing.T) {
	opts := NewOptions()
	prompt := "You are a helpful assistant."
	
	result := opts.WithSystemPrompt(prompt)
	
	// Test that it returns the same instance (method chaining)
	if result != opts {
		t.Error("WithSystemPrompt should return the same Options instance")
	}
	
	// Test that the system prompt is set
	if opts.SystemPrompt == nil {
		t.Error("SystemPrompt should not be nil after WithSystemPrompt")
	}
	if *opts.SystemPrompt != prompt {
		t.Errorf("SystemPrompt = %v, want %v", *opts.SystemPrompt, prompt)
	}
}

func TestOptionsWithAppendSystemPrompt(t *testing.T) {
	opts := NewOptions()
	appendPrompt := "Additional instructions."
	
	result := opts.WithAppendSystemPrompt(appendPrompt)
	
	// Test that it returns the same instance (method chaining)
	if result != opts {
		t.Error("WithAppendSystemPrompt should return the same Options instance")
	}
	
	// Test that the append system prompt is set
	if opts.AppendSystemPrompt == nil {
		t.Error("AppendSystemPrompt should not be nil after WithAppendSystemPrompt")
	}
	if *opts.AppendSystemPrompt != appendPrompt {
		t.Errorf("AppendSystemPrompt = %v, want %v", *opts.AppendSystemPrompt, appendPrompt)
	}
}

func TestOptionsWithAllowedTools(t *testing.T) {
	tests := []struct {
		name  string
		tools []string
	}{
		{"empty tools", []string{}},
		{"single tool", []string{"read_file"}},
		{"multiple tools", []string{"read_file", "write_file", "search_web"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithAllowedTools(tt.tools...)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithAllowedTools should return the same Options instance")
			}
			
			// Test that the allowed tools are set
			if !reflect.DeepEqual(opts.AllowedTools, tt.tools) {
				t.Errorf("AllowedTools = %v, want %v", opts.AllowedTools, tt.tools)
			}
		})
	}
}

func TestOptionsWithDisallowedTools(t *testing.T) {
	tests := []struct {
		name  string
		tools []string
	}{
		{"empty tools", []string{}},
		{"single tool", []string{"dangerous_tool"}},
		{"multiple tools", []string{"dangerous_tool", "risky_operation", "unsafe_command"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithDisallowedTools(tt.tools...)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithDisallowedTools should return the same Options instance")
			}
			
			// Test that the disallowed tools are set
			if !reflect.DeepEqual(opts.DisallowedTools, tt.tools) {
				t.Errorf("DisallowedTools = %v, want %v", opts.DisallowedTools, tt.tools)
			}
		})
	}
}

func TestOptionsWithPermissionMode(t *testing.T) {
	tests := []struct {
		name string
		mode PermissionMode
	}{
		{"default mode", PermissionModeDefault},
		{"accept edits mode", PermissionModeAcceptEdits},
		{"bypass permissions mode", PermissionModeBypassPermissions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithPermissionMode(tt.mode)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithPermissionMode should return the same Options instance")
			}
			
			// Test that the permission mode is set
			if opts.PermissionMode == nil {
				t.Error("PermissionMode should not be nil after WithPermissionMode")
			}
			if *opts.PermissionMode != tt.mode {
				t.Errorf("PermissionMode = %v, want %v", *opts.PermissionMode, tt.mode)
			}
		})
	}
}

func TestOptionsWithMaxTurns(t *testing.T) {
	tests := []struct {
		name  string
		turns int
	}{
		{"single turn", 1},
		{"multiple turns", 5},
		{"many turns", 20},
		{"zero turns", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithMaxTurns(tt.turns)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithMaxTurns should return the same Options instance")
			}
			
			// Test that the max turns is set
			if opts.MaxTurns == nil {
				t.Error("MaxTurns should not be nil after WithMaxTurns")
			}
			if *opts.MaxTurns != tt.turns {
				t.Errorf("MaxTurns = %v, want %v", *opts.MaxTurns, tt.turns)
			}
		})
	}
}

func TestOptionsWithModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
	}{
		{"claude-3-sonnet", "claude-3-sonnet-20240229"},
		{"claude-3-opus", "claude-3-opus-20240229"},
		{"claude-3-haiku", "claude-3-haiku-20240307"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithModel(tt.model)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithModel should return the same Options instance")
			}
			
			// Test that the model is set
			if opts.Model == nil {
				t.Error("Model should not be nil after WithModel")
			}
			if *opts.Model != tt.model {
				t.Errorf("Model = %v, want %v", *opts.Model, tt.model)
			}
		})
	}
}

func TestOptionsWithCwd(t *testing.T) {
	tests := []struct {
		name string
		cwd  string
	}{
		{"absolute path", "/home/user/project"},
		{"relative path", "./project"},
		{"root path", "/"},
		{"temp path", "/tmp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithCwd(tt.cwd)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithCwd should return the same Options instance")
			}
			
			// Test that the cwd is set
			if opts.Cwd == nil {
				t.Error("Cwd should not be nil after WithCwd")
			}
			if *opts.Cwd != tt.cwd {
				t.Errorf("Cwd = %v, want %v", *opts.Cwd, tt.cwd)
			}
		})
	}
}

func TestOptionsWithContinueConversation(t *testing.T) {
	opts := NewOptions()
	
	// Verify initial state
	if opts.ContinueConversation != false {
		t.Errorf("Initial ContinueConversation = %v, want false", opts.ContinueConversation)
	}
	
	result := opts.WithContinueConversation()
	
	// Test that it returns the same instance (method chaining)
	if result != opts {
		t.Error("WithContinueConversation should return the same Options instance")
	}
	
	// Test that continue conversation is enabled
	if opts.ContinueConversation != true {
		t.Errorf("ContinueConversation = %v, want true", opts.ContinueConversation)
	}
}

func TestOptionsWithResume(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
	}{
		{"short session ID", "abc123"},
		{"long session ID", "session_1234567890abcdef"},
		{"uuid session ID", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.WithResume(tt.sessionID)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("WithResume should return the same Options instance")
			}
			
			// Test that the resume session ID is set
			if opts.Resume == nil {
				t.Error("Resume should not be nil after WithResume")
			}
			if *opts.Resume != tt.sessionID {
				t.Errorf("Resume = %v, want %v", *opts.Resume, tt.sessionID)
			}
		})
	}
}

func TestOptionsAddMcpServer(t *testing.T) {
	tests := []struct {
		name   string
		server string
		config McpServerConfig
	}{
		{
			name:   "stdio server",
			server: "filesystem",
			config: &StdioServerConfig{Command: "node", Args: []string{"server.js"}},
		},
		{
			name:   "sse server",
			server: "web_search",
			config: &SSEServerConfig{URL: "https://api.example.com/sse"},
		},
		{
			name:   "http server",
			server: "api_client",
			config: &HTTPServerConfig{URL: "https://api.example.com/mcp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			result := opts.AddMcpServer(tt.server, tt.config)
			
			// Test that it returns the same instance (method chaining)
			if result != opts {
				t.Error("AddMcpServer should return the same Options instance")
			}
			
			// Test that the MCP server is added
			if opts.McpServers == nil {
				t.Error("McpServers should not be nil after AddMcpServer")
			}
			config, exists := opts.McpServers[tt.server]
			if !exists {
				t.Errorf("Server %v not found in McpServers", tt.server)
			}
			if config != tt.config {
				t.Errorf("McpServers[%v] = %v, want %v", tt.server, config, tt.config)
			}
		})
	}
}

func TestOptionsAddMcpServerNilMap(t *testing.T) {
	opts := &Options{McpServers: nil}
	config := &StdioServerConfig{Command: "node"}
	
	result := opts.AddMcpServer("test", config)
	
	// Test that it returns the same instance (method chaining)
	if result != opts {
		t.Error("AddMcpServer should return the same Options instance")
	}
	
	// Test that the map is initialized and server is added
	if opts.McpServers == nil {
		t.Error("McpServers should not be nil after AddMcpServer")
	}
	if len(opts.McpServers) != 1 {
		t.Errorf("McpServers length = %v, want 1", len(opts.McpServers))
	}
	if opts.McpServers["test"] != config {
		t.Errorf("McpServers[test] = %v, want %v", opts.McpServers["test"], config)
	}
}

func TestOptionsAddMcpTool(t *testing.T) {
	tests := []struct {
		name  string
		tools []string
	}{
		{"single tool", []string{"read_file"}},
		{"multiple tools", []string{"read_file", "write_file", "search_web"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOptions()
			
			// Add tools one by one
			for _, tool := range tt.tools {
				result := opts.AddMcpTool(tool)
				
				// Test that it returns the same instance (method chaining)
				if result != opts {
					t.Error("AddMcpTool should return the same Options instance")
				}
			}
			
			// Test that all tools are added
			if !reflect.DeepEqual(opts.McpTools, tt.tools) {
				t.Errorf("McpTools = %v, want %v", opts.McpTools, tt.tools)
			}
		})
	}
}

func TestOptionsJSON(t *testing.T) {
	systemPrompt := "You are a helpful assistant."
	maxTurns := 5
	permissionMode := PermissionModeAcceptEdits
	
	opts := NewOptions().
		WithSystemPrompt(systemPrompt).
		WithMaxTurns(maxTurns).
		WithPermissionMode(permissionMode).
		WithAllowedTools("read_file", "write_file").
		AddMcpTool("read_file")
	
	// Test JSON marshaling (without MCP servers due to interface marshaling complexity)
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal Options: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled Options
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Options: %v", err)
	}
	
	// Verify core fields
	if *unmarshaled.SystemPrompt != systemPrompt {
		t.Errorf("Unmarshaled SystemPrompt = %v, want %v", *unmarshaled.SystemPrompt, systemPrompt)
	}
	if *unmarshaled.MaxTurns != maxTurns {
		t.Errorf("Unmarshaled MaxTurns = %v, want %v", *unmarshaled.MaxTurns, maxTurns)
	}
	if *unmarshaled.PermissionMode != permissionMode {
		t.Errorf("Unmarshaled PermissionMode = %v, want %v", *unmarshaled.PermissionMode, permissionMode)
	}
	if !reflect.DeepEqual(unmarshaled.AllowedTools, opts.AllowedTools) {
		t.Errorf("Unmarshaled AllowedTools = %v, want %v", unmarshaled.AllowedTools, opts.AllowedTools)
	}
	if !reflect.DeepEqual(unmarshaled.McpTools, opts.McpTools) {
		t.Errorf("Unmarshaled McpTools = %v, want %v", unmarshaled.McpTools, opts.McpTools)
	}
}

func TestOptionsJSONMarshalingWithMcpServers(t *testing.T) {
	opts := NewOptions().
		AddMcpServer("filesystem", &StdioServerConfig{Command: "node"}).
		AddMcpServer("web_search", &SSEServerConfig{URL: "https://api.example.com"})
	
	// Test that JSON marshaling works (even if unmarshaling of interfaces is complex)
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal Options with MCP servers: %v", err)
	}
	
	// Verify that the JSON contains the expected structure
	var rawJSON map[string]any
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}
	
	mcpServers, exists := rawJSON["mcpServers"]
	if !exists {
		t.Error("mcpServers field not found in JSON")
	}
	
	serversMap, ok := mcpServers.(map[string]any)
	if !ok {
		t.Error("mcpServers should be a map")
	}
	
	if len(serversMap) != 2 {
		t.Errorf("Expected 2 MCP servers in JSON, got %d", len(serversMap))
	}
}

func TestOptionsChaining(t *testing.T) {
	// Test that all builder methods can be chained together
	opts := NewOptions().
		WithSystemPrompt("Test prompt").
		WithAppendSystemPrompt("Additional prompt").
		WithAllowedTools("tool1", "tool2").
		WithDisallowedTools("bad_tool").
		WithPermissionMode(PermissionModeAcceptEdits).
		WithMaxTurns(10).
		WithModel("claude-3-sonnet").
		WithCwd("/tmp").
		WithContinueConversation().
		WithResume("session_123").
		AddMcpServer("test", &StdioServerConfig{Command: "test"}).
		AddMcpTool("mcp_tool")
	
	// Verify all options were set correctly
	if opts.SystemPrompt == nil || *opts.SystemPrompt != "Test prompt" {
		t.Error("SystemPrompt not set correctly in chain")
	}
	if opts.AppendSystemPrompt == nil || *opts.AppendSystemPrompt != "Additional prompt" {
		t.Error("AppendSystemPrompt not set correctly in chain")
	}
	if len(opts.AllowedTools) != 2 {
		t.Error("AllowedTools not set correctly in chain")
	}
	if len(opts.DisallowedTools) != 1 {
		t.Error("DisallowedTools not set correctly in chain")
	}
	if opts.PermissionMode == nil || *opts.PermissionMode != PermissionModeAcceptEdits {
		t.Error("PermissionMode not set correctly in chain")
	}
	if opts.MaxTurns == nil || *opts.MaxTurns != 10 {
		t.Error("MaxTurns not set correctly in chain")
	}
	if opts.Model == nil || *opts.Model != "claude-3-sonnet" {
		t.Error("Model not set correctly in chain")
	}
	if opts.Cwd == nil || *opts.Cwd != "/tmp" {
		t.Error("Cwd not set correctly in chain")
	}
	if !opts.ContinueConversation {
		t.Error("ContinueConversation not set correctly in chain")
	}
	if opts.Resume == nil || *opts.Resume != "session_123" {
		t.Error("Resume not set correctly in chain")
	}
	if len(opts.McpServers) != 1 {
		t.Error("McpServers not set correctly in chain")
	}
	if len(opts.McpTools) != 1 {
		t.Error("McpTools not set correctly in chain")
	}
}