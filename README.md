# Claude Code SDK for Go

A feature-complete Go SDK for [Claude Code](https://github.com/anthropics/claude-code) that provides 100% parity with the Python SDK. Built with Go's performance, type safety, and excellent concurrency model.

## Features

- 🚀 **High Performance** - Compiled Go binary with no interpreter overhead
- 🔒 **Type Safety** - Compile-time type checking for all API interactions
- ⚡ **Concurrency** - Native goroutines and channels for streaming responses
- 🛠️ **Complete API** - Full feature parity with Python SDK
- 🔧 **Extensible** - Public packages for advanced customization
- 📦 **Easy Install** - Single binary deployment

## Quick Start

### Installation

```bash
go get github.com/jrossi/claude-code-sdk-golang
```

### Prerequisites

Install the Claude Code CLI:
```bash
npm install -g @anthropic-ai/claude-code
export ANTHROPIC_API_KEY=your_api_key_here
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    claudecode "github.com/jrossi/claude-code-sdk-golang"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    stream, err := claudecode.Query(ctx, "What is 2 + 2?", nil)
    if err != nil {
        panic(err)
    }
    defer stream.Close()

    for {
        select {
        case message, ok := <-stream.Messages():
            if !ok {
                return // Stream completed
            }

            switch msg := message.(type) {
            case *claudecode.AssistantMessage:
                for _, block := range msg.Content {
                    if textBlock, ok := block.(*claudecode.TextBlock); ok {
                        fmt.Printf("Claude: %s\n", textBlock.Text)
                    }
                }
            case *claudecode.ResultMessage:
                if msg.TotalCostUSD != nil {
                    fmt.Printf("Cost: $%.4f\n", *msg.TotalCostUSD)
                }
            }

        case err := <-stream.Errors():
            if err != nil {
                fmt.Printf("Error: %v\n", err)
                return
            }

        case <-ctx.Done():
            fmt.Printf("Timeout: %v\n", ctx.Err())
            return
        }
    }
}
```

### Advanced Usage

```go
// Custom options
options := claudecode.NewOptions().
    WithSystemPrompt("You are a helpful assistant.").
    WithAllowedTools("Read", "Write").
    WithMaxTurns(5).
    WithPermissionMode(claudecode.PermissionModeAcceptEdits)

stream, err := claudecode.Query(ctx, "Create a hello.txt file", options)

// MCP server configuration
options = claudecode.NewOptions().
    AddMcpServer("filesystem", &claudecode.StdioServerConfig{
        Command: "npx",
        Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
    }).
    AddMcpTool("read_file").
    AddMcpTool("write_file")

// Custom CLI path
stream, err = claudecode.QueryWithCLIPath(ctx, prompt, options, "/custom/path/claude")
```

## Architecture

The SDK provides both high-level convenience APIs and low-level components for advanced use cases:

### High-Level API
- `claudecode.Query()` - Main entry point for most use cases
- `claudecode.QueryWithCLIPath()` - Custom CLI path support
- `claudecode.NewOptions()` - Fluent configuration builder

### Low-Level Components
```go
import (
    "github.com/jrossi/claude-code-sdk-golang/client"
    "github.com/jrossi/claude-code-sdk-golang/parser"
    "github.com/jrossi/claude-code-sdk-golang/transport"
    "github.com/jrossi/claude-code-sdk-golang/types"
)

// Custom client
client := client.NewClient()
client.SetParserBufferSize(2 * 1024 * 1024)

// Custom transport
transport := transport.NewSubprocessTransport(config)

// Custom parser
parser := parser.NewParser(bufferSize)
```

## Message Types

### Messages
- `UserMessage` - User input
- `AssistantMessage` - Claude's responses with content blocks
- `SystemMessage` - System notifications and metadata  
- `ResultMessage` - Final results with cost and usage information

### Content Blocks
- `TextBlock` - Text responses from Claude
- `ToolUseBlock` - Tool invocations with parameters
- `ToolResultBlock` - Tool execution results

## Configuration Options

- **System Prompts** - `WithSystemPrompt()`, `WithAppendSystemPrompt()`
- **Tools** - `WithAllowedTools()`, `WithDisallowedTools()`
- **Conversation** - `WithMaxTurns()`, `WithContinueConversation()`, `WithResume()`
- **Model** - `WithModel()`, `WithPermissionMode()`
- **MCP Servers** - `AddMcpServer()`, `AddMcpTool()`
- **Environment** - `WithCwd()`, custom CLI paths

## Error Handling

The SDK provides structured error types for comprehensive error handling:

```go
switch err := err.(type) {
case *claudecode.CLINotFoundError:
    fmt.Printf("CLI not found: %s\n", err.CLIPath)
case *claudecode.ConnectionError:
    fmt.Printf("Connection issue: %s\n", err.Message)
case *claudecode.ProcessError:
    fmt.Printf("Process failed (exit %d): %s\n", err.ExitCode, err.Message)
case *claudecode.JSONDecodeError:
    fmt.Printf("JSON parse error: %s\n", err.Line)
}
```

## Examples

See the [examples/](examples/) directory for complete examples:

- [`examples/quickstart.go`](examples/quickstart.go) - Basic usage patterns
- [`examples/advanced.go`](examples/advanced.go) - MCP servers and complex options
- [`examples/error_handling.go`](examples/error_handling.go) - Comprehensive error handling

Run examples:
```bash
go run examples/quickstart.go
go run examples/advanced.go
go run examples/error_handling.go
```

## Testing

Run unit tests:
```bash
go test ./...
```

Run integration tests (requires Claude Code CLI):
```bash
go test -tags=integration ./...
```

## Comparison with Python SDK

| Feature | Go SDK | Python SDK |
|---------|--------|------------|
| **Performance** | Compiled binary | Interpreted |
| **Type Safety** | Compile-time | Runtime |
| **Concurrency** | Goroutines + Channels | async/await |
| **Memory** | Lower overhead | Higher overhead |
| **Deployment** | Single binary | Python environment |
| **API** | 100% parity | ✓ |
| **Streaming** | Channels | Async iterators |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [Claude Code](https://github.com/anthropics/claude-code) - Official Claude Code CLI
- [Python SDK](https://github.com/anthropics/claude-code-sdk-python) - Original Python implementation
- [Documentation](https://docs.anthropic.com/en/docs/claude-code) - Claude Code documentation