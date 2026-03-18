package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// QueryMetricsParams are the parameters for the query_metrics tool.
type QueryMetricsParams struct {
	Query     string `json:"query" required:"true" jsonschema:"description=ClickHouse SQL query against the metrics table (otel_metrics). Select from otel_metrics."`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of rows (default 100)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListMetricNamesParams are the parameters for the list_metric_names tool.
type ListMetricNamesParams struct {
	Prefix    string `json:"prefix" jsonschema:"description=Filter metric names by prefix"`
	Service   string `json:"service" jsonschema:"description=Filter by service name"`
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetMetricMetadataParams are the parameters for the get_metric_metadata tool.
type GetMetricMetadataParams struct {
	MetricName string `json:"metric_name" required:"true" jsonschema:"description=The metric name to get metadata for"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListMetricLabelNamesParams are the parameters for the list_metric_label_names tool.
type ListMetricLabelNamesParams struct {
	MetricName string `json:"metric_name" required:"true" jsonschema:"description=The metric name to list label names for"`
	TimeFrom   string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo     string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListMetricLabelValuesParams are the parameters for the list_metric_label_values tool.
type ListMetricLabelValuesParams struct {
	MetricName string `json:"metric_name" required:"true" jsonschema:"description=The metric name"`
	LabelName  string `json:"label_name" required:"true" jsonschema:"description=The label name to list values for"`
	TimeFrom   string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo     string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// MetricStatsParams are the parameters for the metric_stats tool.
type MetricStatsParams struct {
	MetricName string `json:"metric_name" required:"true" jsonschema:"description=The metric name to get statistics for"`
	Service    string `json:"service" jsonschema:"description=Filter by service name"`
	TimeFrom   string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo     string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Step       string `json:"step" jsonschema:"description=Aggregation step/interval (e.g. 1m; 5m; 1h). Default 1m."`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// AddMetricTools registers all metric-related tools with the MCP server.
func AddMetricTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("query_metrics", "Execute a ClickHouse SQL query against the metrics table (otel_metrics). Returns matching metric data points.", handleQueryMetrics),
		mcpobtrace.MustTool("list_metric_names", "List available metric names, optionally filtered by prefix or service.", handleListMetricNames),
		mcpobtrace.MustTool("get_metric_metadata", "Get metadata for a specific metric (type, description, unit).", handleGetMetricMetadata),
		mcpobtrace.MustTool("list_metric_label_names", "List all label/attribute names for a given metric.", handleListMetricLabelNames),
		mcpobtrace.MustTool("list_metric_label_values", "List all values for a specific label of a metric.", handleListMetricLabelValues),
		mcpobtrace.MustTool("metric_stats", "Get time-series statistics for a metric (min, max, avg, sum, count) over a time range with configurable step.", handleMetricStats),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleQueryMetrics(ctx context.Context, params QueryMetricsParams) (*mcp.CallToolResult, error) {
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

	return doGetRequest(ctx, cfg, "/api/v1/query/metrics?"+q.Encode())
}

func handleListMetricNames(ctx context.Context, params ListMetricNamesParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	if params.Prefix != "" {
		q.Set("prefix", params.Prefix)
	}
	if params.Service != "" {
		q.Set("service", params.Service)
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

	return doGetRequest(ctx, cfg, "/api/v1/metrics/names?"+q.Encode())
}

func handleGetMetricMetadata(ctx context.Context, params GetMetricMetadataParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/metrics/%s/metadata", url.PathEscape(params.MetricName))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleListMetricLabelNames(ctx context.Context, params ListMetricLabelNamesParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/metrics/%s/labels", url.PathEscape(params.MetricName))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleListMetricLabelValues(ctx context.Context, params ListMetricLabelValuesParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/metrics/%s/labels/%s/values",
		url.PathEscape(params.MetricName),
		url.PathEscape(params.LabelName))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleMetricStats(ctx context.Context, params MetricStatsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	step := params.Step
	if step == "" {
		step = "1m"
	}

	q := url.Values{}
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	q.Set("step", step)
	if params.Service != "" {
		q.Set("service", params.Service)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	path := fmt.Sprintf("/api/v1/metrics/%s/stats", url.PathEscape(params.MetricName))
	return doGetRequest(ctx, cfg, path+"?"+q.Encode())
}
