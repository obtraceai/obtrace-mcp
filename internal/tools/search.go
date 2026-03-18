package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GlobalSearchParams are the parameters for the global_search tool.
type GlobalSearchParams struct {
	Query     string `json:"query" required:"true" jsonschema:"description=Search query to match across logs; traces; metrics; dashboards; and incidents"`
	Types     string `json:"types" jsonschema:"description=Comma-separated list of types to search (logs;traces;metrics;dashboards;incidents). Default: all."`
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of results per type (default 10)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the search"`
}

// GenerateDeeplinkParams are the parameters for the generate_deeplink tool.
type GenerateDeeplinkParams struct {
	Type       string `json:"type" required:"true" jsonschema:"description=Type of resource to link to,enum=logs|traces|metrics|dashboard|incident|replay"`
	ResourceID string `json:"resource_id" jsonschema:"description=Resource ID (dashboard ID; incident ID; trace ID; session ID)"`
	Query      string `json:"query" jsonschema:"description=Pre-filled query for log/trace/metric views"`
	TimeFrom   string `json:"time_from" jsonschema:"description=Start time for the view"`
	TimeTo     string `json:"time_to" jsonschema:"description=End time for the view"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the link"`
}

// AddSearchTools registers all search-related tools with the MCP server.
func AddSearchTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("global_search", "Search across all Obtrace data types (logs, traces, metrics, dashboards, incidents) with a single query.", handleGlobalSearch),
		mcpobtrace.MustTool("generate_deeplink", "Generate a deep link URL to a specific view in the Obtrace UI (log query, trace detail, dashboard, incident, replay session).", handleGenerateDeeplink),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleGlobalSearch(ctx context.Context, params GlobalSearchParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("query", params.Query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.Types != "" {
		q.Set("types", params.Types)
	}
	if params.TimeFrom != "" {
		q.Set("time_from", params.TimeFrom)
	}
	if params.TimeTo != "" {
		q.Set("time_to", params.TimeTo)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/search?"+q.Encode())
}

func handleGenerateDeeplink(ctx context.Context, params GenerateDeeplinkParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"type": params.Type,
	}
	if params.ResourceID != "" {
		body["resource_id"] = params.ResourceID
	}
	if params.Query != "" {
		body["query"] = params.Query
	}
	if params.TimeFrom != "" {
		body["time_from"] = params.TimeFrom
	}
	if params.TimeTo != "" {
		body["time_to"] = params.TimeTo
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/deeplinks", body)
}
