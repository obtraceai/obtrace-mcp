// Package observability provides OpenTelemetry instrumentation for the MCP server.
package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/obtraceai/obtrace-mcp"

// Tracer returns the package-level tracer.
func Tracer() trace.Tracer {
	return otel.Tracer(instrumentationName)
}

// ToolCallAttrs returns common attributes for a tool call span.
func ToolCallAttrs(toolName string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("mcp.tool.name", toolName),
		attribute.String("mcp.server.name", "obtrace-mcp"),
	}
}

// TraceToolCall wraps a tool call with a span.
func TraceToolCall(ctx context.Context, toolName string, fn func(ctx context.Context) error) error {
	ctx, span := Tracer().Start(ctx, fmt.Sprintf("mcp.tool.%s", toolName),
		trace.WithAttributes(ToolCallAttrs(toolName)...),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(attribute.Float64("mcp.tool.duration_ms", float64(duration.Milliseconds())))

	if err != nil {
		span.RecordError(err)
	}

	return err
}
