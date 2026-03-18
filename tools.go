package mcpobtrace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool represents a registered MCP tool with its handler.
type Tool struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// Category represents a group of related tools.
type Category string

const (
	CategoryLogs       Category = "logs"
	CategoryTraces     Category = "traces"
	CategoryMetrics    Category = "metrics"
	CategoryDashboards Category = "dashboards"
	CategoryIncidents  Category = "incidents"
	CategoryAlerts     Category = "alerts"
	CategoryAI         Category = "ai"
	CategoryProjects   Category = "projects"
	CategoryReplay     Category = "replay"
	CategorySearch     Category = "search"
	CategoryAdmin      Category = "admin"
)

// AllCategories returns all available tool categories.
func AllCategories() []Category {
	return []Category{
		CategoryLogs,
		CategoryTraces,
		CategoryMetrics,
		CategoryDashboards,
		CategoryIncidents,
		CategoryAlerts,
		CategoryAI,
		CategoryProjects,
		CategoryReplay,
		CategorySearch,
		CategoryAdmin,
	}
}

// MustTool creates a Tool from a typed handler function.
// The type parameter T is used to auto-generate the JSON schema via invopop/jsonschema.
//
// The handler must be a function of the form:
//
//	func(ctx context.Context, params T) (*mcp.CallToolResult, error)
func MustTool[T any](name, description string, handler func(ctx context.Context, params T) (*mcp.CallToolResult, error), opts ...mcp.ToolOption) Tool {
	allOpts := []mcp.ToolOption{mcp.WithDescription(description), mcp.WithInputSchema[T]()}
	allOpts = append(allOpts, opts...)

	tool := mcp.NewTool(name, allOpts...)

	wrappedHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params T
		data, err := json.Marshal(request.GetArguments())
		if err != nil {
			return nil, fmt.Errorf("marshaling arguments: %w", err)
		}
		if err := json.Unmarshal(data, &params); err != nil {
			return nil, fmt.Errorf("unmarshaling arguments into %T: %w", params, err)
		}
		return handler(ctx, params)
	}

	return Tool{
		Tool:    tool,
		Handler: wrappedHandler,
	}
}

// ConvertTool creates a Tool from a handler that returns a generic result type.
// The result is marshaled to JSON and returned as text content.
func ConvertTool[T any, R any](name, description string, handler func(ctx context.Context, params T) (R, error), opts ...mcp.ToolOption) Tool {
	return MustTool(name, description, func(ctx context.Context, params T) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, params)
		if err != nil {
			return nil, err
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshaling result: %w", err)
		}

		return mcp.NewToolResultText(string(data)), nil
	}, opts...)
}

// AddTools registers a slice of Tools with the MCP server.
func AddTools(s *server.MCPServer, tools []Tool) {
	for _, t := range tools {
		s.AddTool(t.Tool, t.Handler)
	}
}
