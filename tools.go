package mcpobtrace

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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

// MustTool creates a Tool from a typed handler function. It panics if the handler
// signature is invalid. The handler must be a function of the form:
//
//	func(ctx context.Context, params T) (*mcp.CallToolResult, error)
//
// where T is a struct whose fields are used to generate the JSON schema.
func MustTool[T any](name, description string, handler func(ctx context.Context, params T) (*mcp.CallToolResult, error), opts ...mcp.ToolOption) Tool {
	schema := generateSchema[T]()

	allOpts := []mcp.ToolOption{mcp.WithToolInputSchema(schema)}
	allOpts = append(allOpts, opts...)

	tool := mcp.NewTool(name, allOpts...)
	tool.Description = description

	wrappedHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params T
		data, err := json.Marshal(request.Params.Arguments)
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

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(data),
				},
			},
		}, nil
	}, opts...)
}

// generateSchema generates a JSON schema from a struct type's tags.
func generateSchema[T any]() mcp.ToolInputSchema {
	schema := mcp.ToolInputSchema{
		Type:       "object",
		Properties: make(map[string]map[string]any),
	}

	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := range t.NumField() {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]

		prop := make(map[string]any)

		// Determine type from Go type
		switch field.Type.Kind() {
		case reflect.String:
			prop["type"] = "string"
		case reflect.Int, reflect.Int32, reflect.Int64:
			prop["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			prop["type"] = "number"
		case reflect.Bool:
			prop["type"] = "boolean"
		case reflect.Slice:
			prop["type"] = "array"
			if field.Type.Elem().Kind() == reflect.String {
				prop["items"] = map[string]any{"type": "string"}
			}
		default:
			prop["type"] = "object"
		}

		// jsonschema tag for description and enum
		if jsTag := field.Tag.Get("jsonschema"); jsTag != "" {
			parts := strings.Split(jsTag, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "description=") {
					prop["description"] = strings.TrimPrefix(part, "description=")
				} else if strings.HasPrefix(part, "enum=") {
					enumStr := strings.TrimPrefix(part, "enum=")
					prop["enum"] = strings.Split(enumStr, "|")
				}
			}
		}

		schema.Properties[name] = prop

		// Check required tag
		if reqTag := field.Tag.Get("required"); reqTag == "true" {
			schema.Required = append(schema.Required, name)
		}
	}

	return schema
}

// AddTools registers a slice of Tools with the MCP server.
func AddTools(s *server.MCPServer, tools []Tool) {
	for _, t := range tools {
		s.AddTool(t.Tool, t.Handler)
	}
}
