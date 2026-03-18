# Configuration

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OBTRACE_URL` | Yes | - | Base URL of the Obtrace API |
| `OBTRACE_API_KEY` | Yes | - | API key for authentication |
| `OBTRACE_TENANT_ID` | No | - | Default tenant ID for scoping all queries |
| `OBTRACE_PROJECT_ID` | No | - | Default project ID for scoping queries |
| `OBTRACE_TLS_CA_CERT` | No | - | Path to custom CA certificate file |
| `OBTRACE_TLS_CLIENT_CERT` | No | - | Path to client certificate for mTLS |
| `OBTRACE_TLS_CLIENT_KEY` | No | - | Path to client key for mTLS |
| `OBTRACE_TLS_INSECURE` | No | `false` | Skip TLS certificate verification |

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--transport` | `stdio` | Transport mode: `stdio`, `sse`, `streamable-http` |
| `--addr` | `:8000` | Listen address for SSE and streamable-http |
| `--enabled-tools` | all | Comma-separated list of tool categories to enable |
| `--enable-write` | `false` | Enable write/mutating tools |
| `--disable-<category>` | `false` | Disable a specific tool category |
| `--version` | - | Print version and exit |

## Tool Categories

Enable or disable tool categories selectively:

```bash
# Only enable logs and traces
mcp-obtrace --enabled-tools logs,traces

# Enable all except replay
mcp-obtrace --disable-replay

# Enable write tools for incidents and dashboards
mcp-obtrace --enable-write --enabled-tools incidents,dashboards
```

## TLS / mTLS

For environments requiring mutual TLS:

```bash
export OBTRACE_TLS_CA_CERT=/path/to/ca.pem
export OBTRACE_TLS_CLIENT_CERT=/path/to/client.pem
export OBTRACE_TLS_CLIENT_KEY=/path/to/client-key.pem
mcp-obtrace
```
