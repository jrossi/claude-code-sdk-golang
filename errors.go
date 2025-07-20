// Package claudecode provides a Go SDK for interacting with Claude Code CLI.
package claudecode

import (
	"errors"
	"fmt"
)

// Sentinel errors for common cases
var (
	// ErrCLINotFound indicates that the Claude Code CLI is not installed or not found
	ErrCLINotFound = errors.New("claude code cli not found")

	// ErrCLIConnection indicates a connection error with the Claude Code CLI
	ErrCLIConnection = errors.New("cli connection error")

	// ErrJSONDecode indicates a JSON decoding error from CLI output
	ErrJSONDecode = errors.New("json decode error")

	// ErrStreamClosed indicates that the message stream has been closed
	ErrStreamClosed = errors.New("message stream closed")

	// ErrInvalidWorkingDirectory indicates an invalid working directory
	ErrInvalidWorkingDirectory = errors.New("invalid working directory")
)

// CLINotFoundError represents an error when Claude Code CLI is not found.
// It provides additional context about where the CLI was searched for.
type CLINotFoundError struct {
	Message string
	CLIPath string
	Err     error
}

func (e *CLINotFoundError) Error() string {
	if e.CLIPath != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.CLIPath)
	}
	return e.Message
}

func (e *CLINotFoundError) Unwrap() error {
	return e.Err
}

// ProcessError represents an error from a failed CLI process.
// It includes the exit code and stderr output for debugging.
type ProcessError struct {
	Message  string
	ExitCode int
	Stderr   string
	Err      error
}

func (e *ProcessError) Error() string {
	msg := fmt.Sprintf("%s (exit code: %d)", e.Message, e.ExitCode)
	if e.Stderr != "" {
		msg = fmt.Sprintf("%s\nError output: %s", msg, e.Stderr)
	}
	return msg
}

func (e *ProcessError) Unwrap() error {
	return e.Err
}

// JSONDecodeError represents an error when unable to decode JSON from CLI output.
// It preserves the original line and underlying error for debugging.
type JSONDecodeError struct {
	Line         string
	OriginalErr  error
	BufferLength int
}

func (e *JSONDecodeError) Error() string {
	truncated := e.Line
	if len(truncated) > 100 {
		truncated = truncated[:100] + "..."
	}
	return fmt.Sprintf("failed to decode JSON: %s", truncated)
}

func (e *JSONDecodeError) Unwrap() error {
	return e.OriginalErr
}

// ConnectionError represents a connection-related error with additional context.
type ConnectionError struct {
	Message string
	Err     error
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewCLINotFoundError creates a new CLINotFoundError with the given message and optional CLI path.
func NewCLINotFoundError(message, cliPath string) *CLINotFoundError {
	return &CLINotFoundError{
		Message: message,
		CLIPath: cliPath,
		Err:     ErrCLINotFound,
	}
}

// NewProcessError creates a new ProcessError with the given details.
func NewProcessError(message string, exitCode int, stderr string) *ProcessError {
	return &ProcessError{
		Message:  message,
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}

// NewJSONDecodeError creates a new JSONDecodeError with the given line and original error.
func NewJSONDecodeError(line string, originalErr error) *JSONDecodeError {
	return &JSONDecodeError{
		Line:        line,
		OriginalErr: originalErr,
	}
}

// NewConnectionError creates a new ConnectionError with the given message and underlying error.
func NewConnectionError(message string, err error) *ConnectionError {
	return &ConnectionError{
		Message: message,
		Err:     err,
	}
}
