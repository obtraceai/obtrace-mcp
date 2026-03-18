package tools

import (
	"context"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ChatToQueryParams are the parameters for the chat_to_query tool.
type ChatToQueryParams struct {
	Question  string `json:"question" required:"true" jsonschema:"description=Natural language question to convert into a query (e.g. 'show me errors in the checkout service in the last hour')"`
	QueryType string `json:"query_type" jsonschema:"description=Target query type,enum=logs|traces|metrics"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// RootCauseAnalysisParams are the parameters for the root_cause_analysis tool.
type RootCauseAnalysisParams struct {
	IncidentID string `json:"incident_id" jsonschema:"description=Incident ID to analyze"`
	TraceID    string `json:"trace_id" jsonschema:"description=Trace ID to analyze"`
	Service    string `json:"service" jsonschema:"description=Service name experiencing issues"`
	TimeFrom   string `json:"time_from" required:"true" jsonschema:"description=Start time of the issue window in RFC3339 format"`
	TimeTo     string `json:"time_to" required:"true" jsonschema:"description=End time of the issue window in RFC3339 format"`
	Symptoms   string `json:"symptoms" jsonschema:"description=Description of observed symptoms"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the analysis"`
}

// AutofixParams are the parameters for the autofix tool.
type AutofixParams struct {
	IncidentID string `json:"incident_id" jsonschema:"description=Incident ID to generate a fix for"`
	ErrorLog   string `json:"error_log" jsonschema:"description=Error log message or stack trace to analyze"`
	Service    string `json:"service" jsonschema:"description=Service name where the error occurred"`
	Language   string `json:"language" jsonschema:"description=Programming language of the affected code,enum=go|typescript|python|java|csharp|ruby|php"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the operation"`
}

// SummarizeParams are the parameters for the summarize tool.
type SummarizeParams struct {
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Service   string `json:"service" jsonschema:"description=Filter by service name"`
	Focus     string `json:"focus" jsonschema:"description=What to focus the summary on,enum=errors|latency|throughput|overall"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the summary"`
}

// AddAITools registers all AI-related tools with the MCP server.
func AddAITools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("chat_to_query", "Convert a natural language question into a ClickHouse query against Obtrace telemetry data. Returns the generated query and optionally its results.", handleChatToQuery),
		mcpobtrace.MustTool("root_cause_analysis", "Perform AI-powered root cause analysis by correlating logs, traces, and metrics around an incident or error. Returns probable root causes ranked by confidence.", handleRootCauseAnalysis),
		mcpobtrace.MustTool("summarize", "Generate an AI summary of system health, errors, latency, or throughput for a time window.", handleSummarize),
	}

	if enableWriteTools {
		tools = append(tools,
			mcpobtrace.MustTool("autofix", "Generate code fix suggestions for errors found in logs/traces. Uses AI to analyze the error context and propose patches.", handleAutofix),
		)
	}

	mcpobtrace.AddTools(s, tools)
}

func handleChatToQuery(ctx context.Context, params ChatToQueryParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"question": params.Question,
	}
	if params.QueryType != "" {
		body["query_type"] = params.QueryType
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/ai/chat-to-query", body)
}

func handleRootCauseAnalysis(ctx context.Context, params RootCauseAnalysisParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"time_from": params.TimeFrom,
		"time_to":   params.TimeTo,
	}
	if params.IncidentID != "" {
		body["incident_id"] = params.IncidentID
	}
	if params.TraceID != "" {
		body["trace_id"] = params.TraceID
	}
	if params.Service != "" {
		body["service"] = params.Service
	}
	if params.Symptoms != "" {
		body["symptoms"] = params.Symptoms
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/ai/root-cause", body)
}

func handleAutofix(ctx context.Context, params AutofixParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{}
	if params.IncidentID != "" {
		body["incident_id"] = params.IncidentID
	}
	if params.ErrorLog != "" {
		body["error_log"] = params.ErrorLog
	}
	if params.Service != "" {
		body["service"] = params.Service
	}
	if params.Language != "" {
		body["language"] = params.Language
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/ai/autofix", body)
}

func handleSummarize(ctx context.Context, params SummarizeParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"time_from": params.TimeFrom,
		"time_to":   params.TimeTo,
	}
	if params.Service != "" {
		body["service"] = params.Service
	}
	if params.Focus != "" {
		body["focus"] = params.Focus
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/ai/summarize", body)
}
