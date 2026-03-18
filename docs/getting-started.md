# Getting Started

## Prerequisites

- Go 1.24+ (for building from source)
- An Obtrace instance with API access
- An API key with appropriate permissions

## Installation

### From source

```bash
go install github.com/obtraceai/obtrace-mcp/cmd/mcp-obtrace@latest
```

### Docker

```bash
docker pull ghcr.io/obtraceai/mcp-obtrace:latest
```

## Quick Start

1. Set your Obtrace credentials:

```bash
export OBTRACE_URL=https://api.obtrace.ai
export OBTRACE_API_KEY=your-api-key
```

2. Run the server:

```bash
# Stdio mode (for Claude Desktop, Claude Code, etc.)
mcp-obtrace

# SSE mode (for web-based integrations)
mcp-obtrace --transport sse --addr :8000
```

3. Configure your MCP client to connect to the server.

## Claude Desktop Integration

Add this to your Claude Desktop MCP configuration (`~/.claude/mcp.json`):

```json
{
  "mcpServers": {
    "obtrace": {
      "command": "mcp-obtrace",
      "env": {
        "OBTRACE_URL": "https://api.obtrace.ai",
        "OBTRACE_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Verifying the Connection

Once connected, ask your AI assistant:

- "List my projects in Obtrace"
- "Show me errors in the last hour"
- "Search for traces with latency over 1 second"
