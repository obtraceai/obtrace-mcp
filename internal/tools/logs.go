package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// QueryLogsParams are the parameters for the query_logs tool.
type QueryLogsParams struct {
	Query     string `json:"query" required:"true" jsonschema:"description=ClickHouse SQL query to run against the logs table. Use otel_logs as the table name."`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format (e.g. 2024-01-01T00:00:00Z)"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of rows to return (default 100)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query. Overrides the default from config."`
}

// SearchLogsParams are the parameters for the search_logs tool.
type SearchLogsParams struct {
	Text      string   `json:"text" required:"true" jsonschema:"description=Free-text search query across log bodies"`
	Severity  []string `json:"severity" jsonschema:"description=Filter by severity levels (e.g. ERROR; WARN; INFO)"`
	Service   string   `json:"service" jsonschema:"description=Filter by service name"`
	TimeFrom  string   `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string   `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int      `json:"limit" jsonschema:"description=Maximum number of results (default 50)"`
	ProjectID string   `json:"project_id" jsonschema:"description=Project ID to scope the search"`
}

// ListLogServicesParams are the parameters for the list_log_services tool.
type ListLogServicesParams struct {
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// LogStatsParams are the parameters for the log_stats tool.
type LogStatsParams struct {
	Service   string `json:"service" jsonschema:"description=Filter by service name"`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	GroupBy   string `json:"group_by" jsonschema:"description=Group statistics by field (e.g. severity; service_name; resource)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// LogPatternsParams are the parameters for the log_patterns tool.
type LogPatternsParams struct {
	Service   string `json:"service" jsonschema:"description=Filter by service name"`
	Severity  string `json:"severity" jsonschema:"description=Filter by severity level"`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of patterns to return (default 20)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// AddLogTools registers all log-related tools with the MCP server.
func AddLogTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("query_logs", "Execute a ClickHouse SQL query against the logs table (otel_logs). Returns matching log entries with timestamp, severity, body, resource, and attributes.", handleQueryLogs),
		mcpobtrace.MustTool("search_logs", "Full-text search across log message bodies with optional filters for severity, service, and time range.", handleSearchLogs),
		mcpobtrace.MustTool("list_log_services", "List all service names that have emitted logs in the given time range.", handleListLogServices),
		mcpobtrace.MustTool("log_stats", "Get aggregated log statistics (counts by severity, service, etc.) for a time range.", handleLogStats),
		mcpobtrace.MustTool("log_patterns", "Discover recurring log patterns (templates) using automatic clustering.", handleLogPatterns),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleQueryLogs(ctx context.Context, params QueryLogsParams) (*mcp.CallToolResult, error) {
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

	return doGetRequest(ctx, cfg, "/api/v1/query/logs?"+q.Encode())
}

func handleSearchLogs(ctx context.Context, params SearchLogsParams) (*mcp.CallToolResult, error) {
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
	q.Set("text", params.Text)
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.Service != "" {
		q.Set("service", params.Service)
	}
	for _, sev := range params.Severity {
		q.Add("severity", sev)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/search/logs?"+q.Encode())
}

func handleListLogServices(ctx context.Context, params ListLogServicesParams) (*mcp.CallToolResult, error) {
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

	return doGetRequest(ctx, cfg, "/api/v1/logs/services?"+q.Encode())
}

func handleLogStats(ctx context.Context, params LogStatsParams) (*mcp.CallToolResult, error) {
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

	return doGetRequest(ctx, cfg, "/api/v1/logs/stats?"+q.Encode())
}

func handleLogPatterns(ctx context.Context, params LogPatternsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
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
	if params.Severity != "" {
		q.Set("severity", params.Severity)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/logs/patterns?"+q.Encode())
}

// doGetRequest is a shared helper for making GET requests to the Obtrace API.
func doGetRequest(ctx context.Context, cfg *mcpobtrace.ObtraceConfig, path string) (*mcp.CallToolResult, error) {
	client := mcpobtrace.ObtraceClientFromContext(ctx)

	reqURL := cfg.URL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if cfg.TenantID != "" {
		req.Header.Set("X-Obtrace-Tenant-ID", cfg.TenantID)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))), nil
	}

	// Pretty-print JSON if possible
	var prettyJSON json.RawMessage
	if err := json.Unmarshal(body, &prettyJSON); err == nil {
		formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err == nil {
			body = formatted
		}
	}

	return mcp.NewToolResultText(string(body)), nil
}
