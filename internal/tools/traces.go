package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// QueryTracesParams are the parameters for the query_traces tool.
type QueryTracesParams struct {
	Query     string `json:"query" required:"true" jsonschema:"description=ClickHouse SQL query against the traces table (otel_traces). Select from otel_traces."`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of rows (default 100)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetTraceParams are the parameters for the get_trace tool.
type GetTraceParams struct {
	TraceID   string `json:"trace_id" required:"true" jsonschema:"description=The trace ID to retrieve"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// SearchTracesParams are the parameters for the search_traces tool.
type SearchTracesParams struct {
	Service    string `json:"service" jsonschema:"description=Filter by service name"`
	Operation  string `json:"operation" jsonschema:"description=Filter by operation/span name"`
	MinLatency string `json:"min_latency" jsonschema:"description=Minimum span duration (e.g. 500ms; 1s; 2m)"`
	MaxLatency string `json:"max_latency" jsonschema:"description=Maximum span duration"`
	Status     string `json:"status" jsonschema:"description=Filter by span status (OK; ERROR; UNSET),enum=OK|ERROR|UNSET"`
	TimeFrom   string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo     string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit      int    `json:"limit" jsonschema:"description=Maximum number of traces (default 50)"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListTraceServicesParams are the parameters for the list_trace_services tool.
type ListTraceServicesParams struct {
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListTraceOperationsParams are the parameters for the list_trace_operations tool.
type ListTraceOperationsParams struct {
	Service   string `json:"service" required:"true" jsonschema:"description=Service name to list operations for"`
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// TraceStatsParams are the parameters for the trace_stats tool.
type TraceStatsParams struct {
	Service   string `json:"service" jsonschema:"description=Filter by service name"`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	GroupBy   string `json:"group_by" jsonschema:"description=Group statistics by field (e.g. service_name; operation; status)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// AddTraceTools registers all trace-related tools with the MCP server.
func AddTraceTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("query_traces", "Execute a ClickHouse SQL query against the traces table (otel_traces). Returns matching spans with trace_id, span_id, operation, duration, status, and attributes.", handleQueryTraces),
		mcpobtrace.MustTool("get_trace", "Get the full trace tree (all spans) for a specific trace ID, showing the complete request flow across services.", handleGetTrace),
		mcpobtrace.MustTool("search_traces", "Search for traces matching criteria like service, operation, latency range, and status.", handleSearchTraces),
		mcpobtrace.MustTool("list_trace_services", "List all service names that have emitted trace spans in the given time range.", handleListTraceServices),
		mcpobtrace.MustTool("list_trace_operations", "List all operation/span names for a given service.", handleListTraceOperations),
		mcpobtrace.MustTool("trace_stats", "Get aggregated trace statistics (p50/p95/p99 latency, error rates, throughput) for a time range.", handleTraceStats),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleQueryTraces(ctx context.Context, params QueryTracesParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("query", params.Query)
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/query/traces?"+q.Encode())
}

func handleGetTrace(ctx context.Context, params GetTraceParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	path := fmt.Sprintf("/api/v1/traces/%s", url.PathEscape(params.TraceID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleSearchTraces(ctx context.Context, params SearchTracesParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.Service != "" {
		q.Set("service", params.Service)
	}
	if params.Operation != "" {
		q.Set("operation", params.Operation)
	}
	if params.MinLatency != "" {
		q.Set("min_latency", params.MinLatency)
	}
	if params.MaxLatency != "" {
		q.Set("max_latency", params.MaxLatency)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/search/traces?"+q.Encode())
}

func handleListTraceServices(ctx context.Context, params ListTraceServicesParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	if params.TimeFrom != "" {
		q.Set("time_from", params.TimeFrom)
	}
	if params.TimeTo != "" {
		q.Set("time_to", params.TimeTo)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/traces/services?"+q.Encode())
}

func handleListTraceOperations(ctx context.Context, params ListTraceOperationsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("service", params.Service)
	if params.TimeFrom != "" {
		q.Set("time_from", params.TimeFrom)
	}
	if params.TimeTo != "" {
		q.Set("time_to", params.TimeTo)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/traces/operations?"+q.Encode())
}

func handleTraceStats(ctx context.Context, params TraceStatsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	if params.Service != "" {
		q.Set("service", params.Service)
	}
	if params.GroupBy != "" {
		q.Set("group_by", params.GroupBy)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/traces/stats?"+q.Encode())
}
