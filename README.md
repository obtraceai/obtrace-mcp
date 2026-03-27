# obtrace-mcp

Model Context Protocol (MCP) server for the Obtrace observability platform.

Exposes logs, traces, metrics, dashboards, incidents, alerts, AI analysis, session replay, and admin capabilities as MCP tools consumable by LLM agents and IDE integrations.

## Install

### Binary

```bash
go install github.com/obtraceai/obtrace-mcp/cmd/mcp-obtrace@latest
```

### Docker

```bash
docker pull ghcr.io/obtraceai/obtrace-mcp:latest
```

### From source

```bash
git clone https://github.com/obtraceai/obtrace-mcp.git
cd obtrace-mcp
make build
```

## Configuration

Set environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `OBTRACE_URL` | Yes | Base URL of the Obtrace API (e.g. `https://api.obtrace.ai`) |
| `OBTRACE_API_KEY` | Yes | API key for authentication |
| `OBTRACE_TENANT_ID` | No | Default tenant ID for scoping |
| `OBTRACE_PROJECT_ID` | No | Default project ID for scoping |
| `OBTRACE_TLS_CA_CERT` | No | Path to custom CA certificate |
| `OBTRACE_TLS_CLIENT_CERT` | No | Path to client certificate (mTLS) |
| `OBTRACE_TLS_CLIENT_KEY` | No | Path to client key (mTLS) |
| `OBTRACE_TLS_INSECURE` | No | Set to `true` to skip TLS verification |

## Usage

### Stdio (CLI / IDE integration)

```bash
export OBTRACE_URL=https://api.obtrace.ai
export OBTRACE_API_KEY=your-api-key

mcp-obtrace
```

### SSE (HTTP Server-Sent Events)

```bash
mcp-obtrace --transport sse --addr :8000
```

### Streamable HTTP

```bash
mcp-obtrace --transport streamable-http --addr :8000
```

### Docker Compose

```bash
cp .env.example .env
# Edit .env with your credentials
docker compose up
```

## Claude Desktop / Claude Code

Add to your MCP configuration:

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

## Tool Categories

| Category | Tools | Description |
|----------|-------|-------------|
| `logs` | `query_logs`, `search_logs`, `list_log_services`, `log_stats`, `log_patterns` | Query and analyze application logs |
| `traces` | `query_traces`, `get_trace`, `search_traces`, `list_trace_services`, `list_trace_operations`, `trace_stats` | Distributed tracing queries and analysis |
| `metrics` | `query_metrics`, `list_metric_names`, `get_metric_metadata`, `list_metric_label_names`, `list_metric_label_values`, `metric_stats` | Time-series metric queries |
| `dashboards` | `list_dashboards`, `get_dashboard`, `get_dashboard_panels`, `create_dashboard`*, `update_dashboard`*, `delete_dashboard`* | Dashboard management |
| `incidents` | `list_incidents`, `get_incident`, `get_incident_timeline`, `create_incident`*, `update_incident`*, `add_incident_activity`* | Incident management |
| `alerts` | `list_alert_rules`, `get_alert_rule`, `list_alert_history`, `list_contact_points`, `create_alert_rule`*, `update_alert_rule`*, `delete_alert_rule`* | Alert rule management |
| `ai` | `chat_to_query`, `root_cause_analysis`, `summarize`, `autofix`* | AI-powered observability analysis |
| `projects` | `list_projects`, `get_project`, `list_apps`, `list_teams`, `list_users` | Project and org administration |
| `replay` | `list_replay_sessions`, `get_replay_session`, `get_replay_events`, `get_replay_errors` | Session replay browsing |
| `search` | `global_search`, `generate_deeplink` | Cross-type search and deep linking |

*Write tools (marked with *) require `--enable-write` flag.

## Flags

```
--transport        Transport mode: stdio, sse, streamable-http (default: stdio)
--addr             Listen address for sse/streamable-http (default: :8000)
--enabled-tools    Comma-separated list of tool categories to enable (default: all)
--enable-write     Enable write/mutating tools (default: false)
--disable-<cat>    Disable a specific category (e.g. --disable-replay)
--version          Print version and exit
```

## Development

```bash
make build         # Build binary
make test          # Run tests
make lint          # Run linter
make run           # Build and run (stdio)
make run-sse       # Build and run (SSE on :8000)
make image         # Build Docker image
```

## Docs

- `docs/index.md`
- `docs/getting-started.md`
- `docs/tools-reference.md`
- `docs/configuration.md`
- `docs/security.md`
