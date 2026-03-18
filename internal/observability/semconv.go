package observability

import "go.opentelemetry.io/otel/attribute"

// Semantic conventions for MCP tool spans.
const (
	// MCPToolName is the name of the MCP tool being invoked.
	MCPToolNameKey = "mcp.tool.name"

	// MCPServerName identifies the MCP server.
	MCPServerNameKey = "mcp.server.name"

	// MCPTransport identifies the transport mode (stdio, sse, streamable-http).
	MCPTransportKey = "mcp.transport"

	// ObtraceQueryType identifies the type of query being executed.
	ObtraceQueryTypeKey = "obtrace.query.type"
)

// Transport mode attributes.
var (
	TransportStdio         = attribute.String(MCPTransportKey, "stdio")
	TransportSSE           = attribute.String(MCPTransportKey, "sse")
	TransportStreamableHTTP = attribute.String(MCPTransportKey, "streamable-http")
)
