package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/jrossi/claude-code-sdk-golang/internal/types"
)

// SubprocessTransport implements Transport using Claude Code CLI subprocess.
type SubprocessTransport struct {
	config *Config
	cmd    *exec.Cmd

	// Channels for communication
	dataChan chan []byte
	errChan  chan error
	doneChan chan struct{}

	// Process pipes
	stdout io.ReadCloser
	stderr io.ReadCloser

	// State management
	connected bool
	streaming bool
	mu        sync.RWMutex // Protects state
}

// NewSubprocessTransport creates a new subprocess transport with the given configuration.
func NewSubprocessTransport(config *Config) *SubprocessTransport {
	return &SubprocessTransport{
		config:   config,
		dataChan: make(chan []byte, 100), // Buffered for performance
		errChan:  make(chan error, 10),
		doneChan: make(chan struct{}),
	}
}

// Connect establishes connection by discovering CLI and preparing the command.
func (st *SubprocessTransport) Connect(ctx context.Context) error {
	if st.connected {
		return nil
	}

	// Discover CLI path if not specified
	cliPath := st.config.CLIPath
	if cliPath == "" {
		var err error
		cliPath, err = st.discoverCLI()
		if err != nil {
			return err
		}
	}

	// Build command
	cmd, err := st.buildCommand(cliPath)
	if err != nil {
		return fmt.Errorf("failed to build command: %w", err)
	}

	st.cmd = cmd
	st.connected = true
	return nil
}

// Stream starts the subprocess and returns channels for receiving data and errors.
func (st *SubprocessTransport) Stream(ctx context.Context) (<-chan []byte, <-chan error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.streaming {
		// Already streaming, return existing channels
		return st.dataChan, st.errChan
	}

	if !st.connected || st.cmd == nil {
		// Send error and return
		go func() {
			st.errChan <- fmt.Errorf("connection error: not connected")
		}()
		return st.dataChan, st.errChan
	}

	// Set up process pipes
	var err error
	st.stdout, err = st.cmd.StdoutPipe()
	if err != nil {
		go func() {
			st.errChan <- fmt.Errorf("connection error: failed to create stdout pipe: %w", err)
		}()
		return st.dataChan, st.errChan
	}

	st.stderr, err = st.cmd.StderrPipe()
	if err != nil {
		go func() {
			st.errChan <- fmt.Errorf("connection error: failed to create stderr pipe: %w", err)
		}()
		return st.dataChan, st.errChan
	}

	// Start the process
	if err := st.cmd.Start(); err != nil {
		go func() {
			st.errChan <- fmt.Errorf("connection error: failed to start CLI process: %w", err)
		}()
		return st.dataChan, st.errChan
	}

	st.streaming = true

	// Start goroutines for streaming
	go st.streamStdout(ctx)
	go st.streamStderr(ctx)
	go st.waitForProcess(ctx)

	return st.dataChan, st.errChan
}

// Close terminates the subprocess and cleans up resources.
func (st *SubprocessTransport) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if !st.connected {
		return nil
	}

	st.connected = false
	st.streaming = false

	// Signal done to all goroutines
	select {
	case <-st.doneChan:
		// Already closed
	default:
		close(st.doneChan)
	}

	// Close pipes if they exist
	if st.stdout != nil {
		st.stdout.Close()
	}
	if st.stderr != nil {
		st.stderr.Close()
	}

	// Clean up command if it exists
	if st.cmd != nil && st.cmd.Process != nil {
		if err := st.cmd.Process.Kill(); err != nil {
			// Process might already be dead
		}
		st.cmd.Wait() // Clean up zombie
	}

	return nil
}

// IsConnected returns true if the transport is connected.
func (st *SubprocessTransport) IsConnected() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.connected
}

// discoverCLI attempts to find the Claude Code CLI binary.
func (st *SubprocessTransport) discoverCLI() (string, error) {
	// First try which/where command
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// Common installation paths to check
	var searchPaths []string

	homeDir, err := os.UserHomeDir()
	if err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, ".npm-global", "bin", "claude"),
			filepath.Join(homeDir, ".local", "bin", "claude"),
			filepath.Join(homeDir, "node_modules", ".bin", "claude"),
			filepath.Join(homeDir, ".yarn", "bin", "claude"),
		)
	}

	// System paths
	searchPaths = append(searchPaths,
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude", // macOS with Homebrew on Apple Silicon
	)

	// Windows-specific paths
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			searchPaths = append(searchPaths,
				filepath.Join(appData, "npm", "claude.cmd"),
			)
		}
		if programFiles := os.Getenv("PROGRAMFILES"); programFiles != "" {
			searchPaths = append(searchPaths,
				filepath.Join(programFiles, "nodejs", "claude.cmd"),
			)
		}
	}

	// Check each path
	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	// Check if Node.js is installed
	if _, err := exec.LookPath("node"); err != nil {
		return "", fmt.Errorf("CLI not found: Claude Code requires Node.js, which is not installed.\n\n" +
			"Install Node.js from: https://nodejs.org/\n" +
			"\nAfter installing Node.js, install Claude Code:\n" +
			"  npm install -g @anthropic-ai/claude-code")
	}

	return "", fmt.Errorf("CLI not found: Claude Code not found. Install with:\n" +
		"  npm install -g @anthropic-ai/claude-code\n" +
		"\nIf already installed locally, try:\n" +
		"  export PATH=\"$HOME/node_modules/.bin:$PATH\"\n" +
		"\nOr specify the path when creating transport")
}

// buildCommand constructs the CLI command with all options.
func (st *SubprocessTransport) buildCommand(cliPath string) (*exec.Cmd, error) {
	args := []string{"--output-format", "stream-json", "--verbose"}

	opts := st.config.Options
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	// System prompts
	if opts.SystemPrompt != nil {
		args = append(args, "--system-prompt", *opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != nil {
		args = append(args, "--append-system-prompt", *opts.AppendSystemPrompt)
	}

	// Tools
	if len(opts.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(opts.AllowedTools, ","))
	}
	if len(opts.DisallowedTools) > 0 {
		args = append(args, "--disallowedTools", strings.Join(opts.DisallowedTools, ","))
	}

	// Conversation control
	if opts.MaxTurns != nil {
		args = append(args, "--max-turns", strconv.Itoa(*opts.MaxTurns))
	}
	if opts.ContinueConversation {
		args = append(args, "--continue")
	}
	if opts.Resume != nil {
		args = append(args, "--resume", *opts.Resume)
	}

	// Model and permissions
	if opts.Model != nil {
		args = append(args, "--model", *opts.Model)
	}
	if opts.PermissionMode != nil {
		args = append(args, "--permission-mode", string(*opts.PermissionMode))
	}
	if opts.PermissionPromptToolName != nil {
		args = append(args, "--permission-prompt-tool", *opts.PermissionPromptToolName)
	}

	// MCP configuration
	if len(opts.McpServers) > 0 {
		mcpConfig := map[string]any{
			"mcpServers": st.convertMcpServers(opts.McpServers),
		}
		mcpJSON, err := json.Marshal(mcpConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal MCP config: %w", err)
		}
		args = append(args, "--mcp-config", string(mcpJSON))
	}

	// Add the prompt
	args = append(args, "--print", st.config.Prompt)

	// Create command
	cmd := exec.Command(cliPath, args...)

	// Set working directory if specified
	if opts.Cwd != nil {
		cmd.Dir = *opts.Cwd
	}

	// Set environment
	cmd.Env = append(os.Environ(), "CLAUDE_CODE_ENTRYPOINT=sdk-go")

	return cmd, nil
}

// convertMcpServers converts Go MCP server configs to the format expected by CLI.
func (st *SubprocessTransport) convertMcpServers(servers map[string]types.McpServerConfig) map[string]any {
	result := make(map[string]any)

	for name, server := range servers {
		switch s := server.(type) {
		case *types.StdioServerConfig:
			config := map[string]any{
				"type":    "stdio",
				"command": s.Command,
			}
			if len(s.Args) > 0 {
				config["args"] = s.Args
			}
			if len(s.Env) > 0 {
				config["env"] = s.Env
			}
			result[name] = config

		case *types.SSEServerConfig:
			config := map[string]any{
				"type": "sse",
				"url":  s.URL,
			}
			if len(s.Headers) > 0 {
				config["headers"] = s.Headers
			}
			result[name] = config

		case *types.HTTPServerConfig:
			config := map[string]any{
				"type": "http",
				"url":  s.URL,
			}
			if len(s.Headers) > 0 {
				config["headers"] = s.Headers
			}
			result[name] = config
		}
	}

	return result
}

// streamStdout reads from stdout and sends data to the data channel.
func (st *SubprocessTransport) streamStdout(ctx context.Context) {
	defer func() {
		// Close channels when done streaming
		close(st.dataChan)
	}()

	if st.stdout == nil {
		return
	}

	scanner := bufio.NewScanner(st.stdout)
	// Set a larger buffer for scanner to handle large JSON lines
	const maxCapacity = 1024 * 1024 // 1MB per line
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for {
		select {
		case <-ctx.Done():
			return
		case <-st.doneChan:
			return
		default:
			// Continue scanning
		}

		if !scanner.Scan() {
			// Check for error or EOF
			if err := scanner.Err(); err != nil {
				select {
				case st.errChan <- fmt.Errorf("connection error: error reading stdout: %w", err):
				case <-ctx.Done():
				case <-st.doneChan:
				}
			}
			return
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Make a copy of the line data since scanner reuses the buffer
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)

		// Send the line to the data channel
		select {
		case st.dataChan <- lineCopy:
		case <-ctx.Done():
			return
		case <-st.doneChan:
			return
		}
	}
}

// streamStderr reads from stderr and collects error output.
func (st *SubprocessTransport) streamStderr(ctx context.Context) {
	if st.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(st.stderr)
	var stderrLines []string
	const maxStderrSize = 10 * 1024 * 1024 // 10MB max stderr
	var totalSize int

	for {
		select {
		case <-ctx.Done():
			return
		case <-st.doneChan:
			return
		default:
			// Continue scanning
		}

		if !scanner.Scan() {
			// EOF or error - process stderr content
			if len(stderrLines) > 0 {
				stderrOutput := strings.Join(stderrLines, "\n")
				// Send stderr as a connection error (non-blocking)
				select {
				case st.errChan <- fmt.Errorf("connection error: CLI stderr output: %s", stderrOutput):
				case <-ctx.Done():
				case <-st.doneChan:
				}
			}
			return
		}

		line := scanner.Text()
		lineSize := len(line)

		// Enforce memory limit
		if totalSize+lineSize > maxStderrSize {
			stderrLines = append(stderrLines, fmt.Sprintf("[stderr truncated after %d bytes]", totalSize))
			break
		}

		stderrLines = append(stderrLines, line)
		totalSize += lineSize
	}
}

// waitForProcess waits for the subprocess to complete and handles exit codes.
func (st *SubprocessTransport) waitForProcess(ctx context.Context) {
	defer func() {
		// Close error channel when process monitoring is done
		close(st.errChan)
	}()

	if st.cmd == nil {
		return
	}

	// Wait for process in a separate goroutine to allow context cancellation
	processErrChan := make(chan error, 1)
	go func() {
		processErrChan <- st.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context cancelled, kill the process
		if st.cmd.Process != nil {
			st.cmd.Process.Kill()
		}
		<-processErrChan // Wait for process to actually exit
		return

	case <-st.doneChan:
		// Transport closed, kill the process
		if st.cmd.Process != nil {
			st.cmd.Process.Kill()
		}
		<-processErrChan // Wait for process to actually exit
		return

	case err := <-processErrChan:
		// Process completed naturally
		if err != nil {
			// Process failed
			if exitErr, ok := err.(*exec.ExitError); ok {
				processErr := fmt.Errorf("process error: CLI process failed with exit code %d: %w", exitErr.ExitCode(), err)

				// Send process error (non-blocking)
				select {
				case st.errChan <- processErr:
				case <-ctx.Done():
				case <-st.doneChan:
				}
			} else {
				// Other error
				select {
				case st.errChan <- fmt.Errorf("connection error: process wait failed: %w", err):
				case <-ctx.Done():
				case <-st.doneChan:
				}
			}
		}
		// Process completed successfully (exit code 0) - no error to send
		return
	}
}
