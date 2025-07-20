package transport

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jrossi/claude-code-sdk-golang/types"
)

// Mock pipes for testing streaming functions
type mockPipe struct {
	data []byte
	pos  int
	mu   sync.Mutex
}

func (m *mockPipe) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockPipe) Close() error {
	return nil
}

// TestDiscoverCLI tests the CLI discovery function comprehensively
func TestDiscoverCLI(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary home directory
	tempHome := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(tempHome, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		setup    func() func() // setup function returns cleanup function
		wantPath bool
		wantErr  string
	}{
		{
			name: "CLI found in PATH",
			setup: func() func() {
				// Create a fake CLI in a temp bin directory
				binDir := filepath.Join(tempDir, "bin")
				if err := os.MkdirAll(binDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				cliPath := filepath.Join(binDir, "claude")
				if runtime.GOOS == "windows" {
					cliPath += ".exe"
				}
				
				if err := os.WriteFile(cliPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
					t.Fatal(err)
				}
				
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
				
				return func() {
					os.Setenv("PATH", oldPath)
				}
			},
			wantPath: true,
		},
		{
			name: "CLI found in npm-global",
			setup: func() func() {
				// Create npm-global directory structure
				npmGlobalDir := filepath.Join(tempHome, ".npm-global", "bin")
				if err := os.MkdirAll(npmGlobalDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				cliPath := filepath.Join(npmGlobalDir, "claude")
				if err := os.WriteFile(cliPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
					t.Fatal(err)
				}
				
				// Mock user home directory
				oldHome := os.Getenv("HOME")
				if runtime.GOOS == "windows" {
					oldHome = os.Getenv("USERPROFILE")
					os.Setenv("USERPROFILE", tempHome)
				} else {
					os.Setenv("HOME", tempHome)
				}
				
				// Clear PATH to force discovery
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", "")
				
				return func() {
					if runtime.GOOS == "windows" {
						os.Setenv("USERPROFILE", oldHome)
					} else {
						os.Setenv("HOME", oldHome)
					}
					os.Setenv("PATH", oldPath)
				}
			},
			wantPath: true,
		},
		{
			name: "CLI found in local bin",
			setup: func() func() {
				// Create .local/bin directory structure
				localBinDir := filepath.Join(tempHome, ".local", "bin")
				if err := os.MkdirAll(localBinDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				cliPath := filepath.Join(localBinDir, "claude")
				if err := os.WriteFile(cliPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
					t.Fatal(err)
				}
				
				// Mock user home directory
				oldHome := os.Getenv("HOME")
				if runtime.GOOS == "windows" {
					oldHome = os.Getenv("USERPROFILE")
					os.Setenv("USERPROFILE", tempHome)
				} else {
					os.Setenv("HOME", tempHome)
				}
				
				// Clear PATH to force discovery
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", "")
				
				return func() {
					if runtime.GOOS == "windows" {
						os.Setenv("USERPROFILE", oldHome)
					} else {
						os.Setenv("HOME", oldHome)
					}
					os.Setenv("PATH", oldPath)
				}
			},
			wantPath: true,
		},
		{
			name: "CLI not found, Node.js not available",
			setup: func() func() {
				// Clear PATH completely and mock missing home
				oldPath := os.Getenv("PATH")
				oldHome := os.Getenv("HOME")
				if runtime.GOOS == "windows" {
					oldHome = os.Getenv("USERPROFILE")
					os.Setenv("USERPROFILE", "/nonexistent")
				} else {
					os.Setenv("HOME", "/nonexistent")
				}
				os.Setenv("PATH", "")
				
				return func() {
					os.Setenv("PATH", oldPath)
					if runtime.GOOS == "windows" {
						os.Setenv("USERPROFILE", oldHome)
					} else {
						os.Setenv("HOME", oldHome)
					}
				}
			},
			wantPath: false,
			wantErr:  "Node.js",
		},
		{
			name: "CLI not found, Node.js available",
			setup: func() func() {
				// Create a fake node binary
				nodeDir := filepath.Join(tempDir, "node_bin")
				if err := os.MkdirAll(nodeDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				nodePath := filepath.Join(nodeDir, "node")
				if runtime.GOOS == "windows" {
					nodePath += ".exe"
				}
				
				if err := os.WriteFile(nodePath, []byte("#!/bin/sh\necho node"), 0755); err != nil {
					t.Fatal(err)
				}
				
				// Set PATH to include node but not claude
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", nodeDir)
				
				// Mock missing home
				oldHome := os.Getenv("HOME")
				if runtime.GOOS == "windows" {
					oldHome = os.Getenv("USERPROFILE")
					os.Setenv("USERPROFILE", "/nonexistent")
				} else {
					os.Setenv("HOME", "/nonexistent")
				}
				
				return func() {
					os.Setenv("PATH", oldPath)
					if runtime.GOOS == "windows" {
						os.Setenv("USERPROFILE", oldHome)
					} else {
						os.Setenv("HOME", oldHome)
					}
				}
			},
			wantPath: false,
			wantErr:  "npm install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			config := &Config{
				Options: types.NewOptions(),
			}
			transport := NewSubprocessTransport(config)

			path, err := transport.discoverCLI()

			if tt.wantPath {
				if err != nil {
					t.Errorf("Expected to find CLI, got error: %v", err)
				}
				if path == "" {
					t.Error("Expected non-empty path")
				}
			} else {
				if err == nil {
					t.Error("Expected error, got none")
				} else if tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

// TestDiscoverCLIWindowsSpecific tests Windows-specific discovery paths
func TestDiscoverCLIWindowsSpecific(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tempDir, err := os.MkdirTemp("", "cli_test_windows")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		envVar  string
		subPath string
	}{
		{
			name:    "APPDATA npm path",
			envVar:  "APPDATA",
			subPath: "npm",
		},
		{
			name:    "PROGRAMFILES path",
			envVar:  "PROGRAMFILES",
			subPath: "nodejs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create directory structure
			baseDir := filepath.Join(tempDir, tt.name)
			cliDir := filepath.Join(baseDir, tt.subPath)
			if err := os.MkdirAll(cliDir, 0755); err != nil {
				t.Fatal(err)
			}

			cliPath := filepath.Join(cliDir, "claude.cmd")
			if err := os.WriteFile(cliPath, []byte("@echo off\necho test"), 0755); err != nil {
				t.Fatal(err)
			}

			// Set environment variable
			oldEnv := os.Getenv(tt.envVar)
			os.Setenv(tt.envVar, baseDir)
			defer os.Setenv(tt.envVar, oldEnv)

			// Clear PATH to force discovery
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", "")
			defer os.Setenv("PATH", oldPath)

			config := &Config{Options: types.NewOptions()}
			transport := NewSubprocessTransport(config)

			path, err := transport.discoverCLI()
			if err != nil {
				t.Errorf("Expected to find CLI, got error: %v", err)
			}
			if path != cliPath {
				t.Errorf("Expected path %q, got %q", cliPath, path)
			}
		})
	}
}

// TestStreamStdout tests the stdout streaming function
func TestStreamStdout(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectData []string
		expectErr  bool
		timeout    time.Duration
	}{
		{
			name:       "single line",
			input:      "test line\n",
			expectData: []string{"test line"},
			timeout:    time.Second,
		},
		{
			name:       "multiple lines",
			input:      "line1\nline2\nline3\n",
			expectData: []string{"line1", "line2", "line3"},
			timeout:    time.Second,
		},
		{
			name:       "empty lines ignored",
			input:      "line1\n\nline2\n\n",
			expectData: []string{"line1", "line2"},
			timeout:    time.Second,
		},
		{
			name:       "large line",
			input:      strings.Repeat("x", 100000) + "\n",
			expectData: []string{strings.Repeat("x", 100000)},
			timeout:    time.Second,
		},
		{
			name:    "no data",
			input:   "",
			timeout: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Options: types.NewOptions()}
			transport := NewSubprocessTransport(config)

			// Initialize channels
			transport.dataChan = make(chan []byte, 10)
			transport.errChan = make(chan error, 10)
			transport.doneChan = make(chan struct{})

			// Set up mock stdout
			transport.stdout = &mockPipe{data: []byte(tt.input)}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Start streaming in background
			go transport.streamStdout(ctx)

			// Collect data
			var receivedData []string
			var receivedErr error

		collectLoop:
			for {
				select {
				case data := <-transport.dataChan:
					if data != nil {
						receivedData = append(receivedData, string(data))
					} else {
						// Channel closed
						break collectLoop
					}
				case err := <-transport.errChan:
					receivedErr = err
					break collectLoop
				case <-ctx.Done():
					break collectLoop
				}
			}

			// Verify results
			if tt.expectErr && receivedErr == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectErr && receivedErr != nil {
				t.Errorf("Unexpected error: %v", receivedErr)
			}

			if len(receivedData) != len(tt.expectData) {
				t.Errorf("Expected %d lines, got %d: %v", len(tt.expectData), len(receivedData), receivedData)
			} else {
				for i, expected := range tt.expectData {
					if receivedData[i] != expected {
						t.Errorf("Line %d: expected %q, got %q", i, expected, receivedData[i])
					}
				}
			}
		})
	}
}

// TestStreamStderr tests the stderr streaming function
func TestStreamStderr(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		errContains string
	}{
		{
			name:        "single error line",
			input:       "Error: something went wrong\n",
			expectErr:   true,
			errContains: "something went wrong",
		},
		{
			name:        "multiple error lines",
			input:       "Warning: first issue\nError: second issue\n",
			expectErr:   true,
			errContains: "first issue",
		},
		{
			name:  "empty stderr",
			input: "",
		},
		{
			name:        "large stderr",
			input:       strings.Repeat("Error line\n", 1000),
			expectErr:   true,
			errContains: "Error line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Options: types.NewOptions()}
			transport := NewSubprocessTransport(config)

			// Initialize channels
			transport.dataChan = make(chan []byte, 10)
			transport.errChan = make(chan error, 10)
			transport.doneChan = make(chan struct{})

			// Set up mock stderr
			transport.stderr = &mockPipe{data: []byte(tt.input)}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// Start streaming in background
			go transport.streamStderr(ctx)

			// Wait for completion or timeout
			var receivedErr error
			select {
			case err := <-transport.errChan:
				receivedErr = err
			case <-ctx.Done():
				// Timeout is OK for empty stderr
			}

			// Verify results
			if tt.expectErr {
				if receivedErr == nil {
					t.Error("Expected error from stderr but got none")
				} else if !strings.Contains(receivedErr.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, receivedErr)
				}
			} else if receivedErr != nil {
				t.Errorf("Unexpected error from stderr: %v", receivedErr)
			}
		})
	}
}

// MockCmd implements the exec.Cmd interface for testing
type MockCmd struct {
	ProcessState *os.ProcessState
	WaitError    error
	WaitDelay    time.Duration
	Process      *os.Process
}

func (m *MockCmd) Wait() error {
	if m.WaitDelay > 0 {
		time.Sleep(m.WaitDelay)
	}
	return m.WaitError
}

// TestWaitForProcess tests the process waiting function
func TestWaitForProcess(t *testing.T) {
	tests := []struct {
		name         string
		waitError    error
		waitDelay    time.Duration
		cancelAfter  time.Duration
		expectKilled bool
	}{
		{
			name:      "process exits normally",
			waitError: nil,
		},
		{
			name:      "process exits with error",
			waitError: fmt.Errorf("exit status 1"),
		},
		{
			name:         "process cancelled by context",
			waitDelay:    500 * time.Millisecond,
			cancelAfter:  100 * time.Millisecond,
			expectKilled: true,
		},
		{
			name:         "process cancelled by close",
			waitDelay:    500 * time.Millisecond,
			expectKilled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Options: types.NewOptions()}
			transport := NewSubprocessTransport(config)

			// Initialize channels
			transport.errChan = make(chan error, 10)
			transport.doneChan = make(chan struct{})

			// Create a real process for kill testing
			if tt.expectKilled {
				// Use a long-running command that we can kill
				cmd := exec.Command("sleep", "10")
				if runtime.GOOS == "windows" {
					cmd = exec.Command("ping", "-n", "10", "127.0.0.1")
				}
				err := cmd.Start()
				if err != nil {
					t.Skip("Cannot start test process:", err)
				}
				transport.cmd = cmd
			} else {
				// For non-killing tests, we'll use a mock command that exits normally
				// Since we can't mock Wait method directly, we'll test with a real command
				// that exits quickly
				cmd := exec.Command("true") // Command that always succeeds immediately
				if runtime.GOOS == "windows" {
					cmd = exec.Command("cmd", "/c", "exit", "0")
				}
				if tt.waitError != nil {
					cmd = exec.Command("false") // Command that always fails
					if runtime.GOOS == "windows" {
						cmd = exec.Command("cmd", "/c", "exit", "1")
					}
				}
				transport.cmd = cmd
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Start process monitoring
			done := make(chan struct{})
			go func() {
				transport.waitForProcess(ctx)
				close(done)
			}()

			// Handle different test scenarios
			if tt.cancelAfter > 0 {
				// Cancel context after delay
				time.AfterFunc(tt.cancelAfter, cancel)
			} else if tt.expectKilled && tt.cancelAfter == 0 {
				// Close transport after a short delay
				time.AfterFunc(100*time.Millisecond, func() {
					close(transport.doneChan)
				})
			}

			// Wait for completion
			select {
			case <-done:
				// Process monitoring completed
			case <-time.After(3 * time.Second):
				t.Fatal("Process monitoring timed out")
			}

			// Verify error channel is closed
			select {
			case _, ok := <-transport.errChan:
				if ok {
					t.Error("Error channel should be closed")
				}
			default:
				t.Error("Error channel should be closed and readable")
			}

			// For kill tests, verify process is dead
			if tt.expectKilled && transport.cmd.Process != nil {
				// Give process time to die
				time.Sleep(100 * time.Millisecond)
				err := transport.cmd.Process.Signal(os.Signal(nil))
				if err == nil {
					t.Error("Process should be dead but signal succeeded")
				}
			}
		})
	}
}

// TestStreamFunctionIntegration tests streaming functions with context cancellation
func TestStreamFunctionIntegration(t *testing.T) {
	config := &Config{Options: types.NewOptions()}
	transport := NewSubprocessTransport(config)

	// Initialize channels
	transport.dataChan = make(chan []byte, 10)
	transport.errChan = make(chan error, 10)
	transport.doneChan = make(chan struct{})

	// Set up mock pipes with data
	stdoutData := "stdout line 1\nstdout line 2\n"
	stderrData := "stderr line 1\nstderr line 2\n"
	
	transport.stdout = &mockPipe{data: []byte(stdoutData)}
	transport.stderr = &mockPipe{data: []byte(stderrData)}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start both streaming functions
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transport.streamStdout(ctx)
	}()

	go func() {
		defer wg.Done()
		transport.streamStderr(ctx)
	}()

	// Collect results
	var stdoutLines []string
	var stderrErr error
	var dataChannelClosed bool

	for i := 0; i < 3; i++ { // Expect 2 stdout lines + 1 stderr error
		select {
		case data, ok := <-transport.dataChan:
			if !ok {
				dataChannelClosed = true
				break
			}
			if data != nil {
				stdoutLines = append(stdoutLines, string(data))
			}
		case err := <-transport.errChan:
			stderrErr = err
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for stream data")
		}
	}

	// Wait for goroutines to complete
	wg.Wait()

	// Verify results
	expectedStdoutLines := []string{"stdout line 1", "stdout line 2"}
	if len(stdoutLines) != len(expectedStdoutLines) {
		t.Errorf("Expected %d stdout lines, got %d: %v", len(expectedStdoutLines), len(stdoutLines), stdoutLines)
	}

	if stderrErr == nil {
		t.Error("Expected stderr error but got none")
	} else if !strings.Contains(stderrErr.Error(), "stderr line") {
		t.Errorf("Expected stderr error to contain stderr content, got: %v", stderrErr)
	}

	if !dataChannelClosed {
		t.Error("Data channel should be closed after stdout streaming completes")
	}
}

// TestStreamContextCancellation tests that streaming respects context cancellation
func TestStreamContextCancellation(t *testing.T) {
	config := &Config{Options: types.NewOptions()}
	transport := NewSubprocessTransport(config)

	// Initialize channels
	transport.dataChan = make(chan []byte, 10)
	transport.errChan = make(chan error, 10)
	transport.doneChan = make(chan struct{})

	// Create pipes that block on read to test cancellation
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	
	transport.stdout = stdoutReader
	transport.stderr = stderrReader

	ctx, cancel := context.WithCancel(context.Background())

	// Start streaming functions
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transport.streamStdout(ctx)
	}()

	go func() {
		defer wg.Done()
		transport.streamStderr(ctx)
	}()

	// Cancel context after a short delay
	time.AfterFunc(100*time.Millisecond, cancel)

	// Close pipes to unblock reads
	time.AfterFunc(200*time.Millisecond, func() {
		stdoutWriter.Close()
		stderrWriter.Close()
	})

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - functions returned due to cancellation
	case <-time.After(time.Second):
		t.Fatal("Streaming functions did not respect context cancellation")
	}
}

// TestStreamDoneChanCancellation tests that streaming respects done channel
func TestStreamDoneChanCancellation(t *testing.T) {
	config := &Config{Options: types.NewOptions()}
	transport := NewSubprocessTransport(config)

	// Initialize channels
	transport.dataChan = make(chan []byte, 10)
	transport.errChan = make(chan error, 10)
	transport.doneChan = make(chan struct{})

	// Create pipes that block on read
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	
	transport.stdout = stdoutReader
	transport.stderr = stderrReader

	ctx := context.Background()

	// Start streaming functions
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transport.streamStdout(ctx)
	}()

	go func() {
		defer wg.Done()
		transport.streamStderr(ctx)
	}()

	// Close done channel after a short delay
	time.AfterFunc(100*time.Millisecond, func() {
		close(transport.doneChan)
	})

	// Close pipes to unblock reads
	time.AfterFunc(200*time.Millisecond, func() {
		stdoutWriter.Close()
		stderrWriter.Close()
	})

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - functions returned due to done channel
	case <-time.After(time.Second):
		t.Fatal("Streaming functions did not respect done channel")
	}
}

// TestSubprocessTransportNilPipeHandling tests behavior with nil pipes
func TestSubprocessTransportNilPipeHandling(t *testing.T) {
	config := &Config{Options: types.NewOptions()}
	transport := NewSubprocessTransport(config)

	// Initialize channels
	transport.dataChan = make(chan []byte, 10)
	transport.errChan = make(chan error, 10)
	transport.doneChan = make(chan struct{})

	// Set pipes to nil
	transport.stdout = nil
	transport.stderr = nil

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start streaming functions with nil pipes
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transport.streamStdout(ctx)
	}()

	go func() {
		defer wg.Done()
		transport.streamStderr(ctx)
	}()

	// Wait for completion
	wg.Wait()

	// Verify data channel is closed (from streamStdout)
	select {
	case _, ok := <-transport.dataChan:
		if ok {
			t.Error("Data channel should be closed when stdout is nil")
		}
	default:
		t.Error("Data channel should be closed and readable")
	}

	// Verify no errors from stderr (should exit early when nil)
	select {
	case err := <-transport.errChan:
		t.Errorf("Unexpected error from nil stderr: %v", err)
	default:
		// Expected - no error should be sent for nil stderr
	}
}

// TestWaitForProcessNilCmd tests waitForProcess with nil cmd
func TestWaitForProcessNilCmd(t *testing.T) {
	config := &Config{Options: types.NewOptions()}
	transport := NewSubprocessTransport(config)

	// Initialize channels
	transport.errChan = make(chan error, 10)
	transport.doneChan = make(chan struct{})

	// Set cmd to nil
	transport.cmd = nil

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start process monitoring
	done := make(chan struct{})
	go func() {
		transport.waitForProcess(ctx)
		close(done)
	}()

	// Should complete quickly with nil cmd
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("waitForProcess should exit quickly with nil cmd")
	}

	// Verify error channel is closed
	select {
	case _, ok := <-transport.errChan:
		if ok {
			t.Error("Error channel should be closed")
		}
	default:
		t.Error("Error channel should be closed and readable")
	}
}