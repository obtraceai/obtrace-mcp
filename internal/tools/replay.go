package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListReplaySessionsParams are the parameters for the list_replay_sessions tool.
type ListReplaySessionsParams struct {
	UserID    string `json:"user_id" jsonschema:"description=Filter by user ID"`
	Service   string `json:"service" jsonschema:"description=Filter by service/app name"`
	HasError  bool   `json:"has_error" jsonschema:"description=Filter to only sessions with errors"`
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of sessions (default 50)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetReplaySessionParams are the parameters for the get_replay_session tool.
type GetReplaySessionParams struct {
	SessionID string `json:"session_id" required:"true" jsonschema:"description=The replay session ID to retrieve"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetReplayEventsParams are the parameters for the get_replay_events tool.
type GetReplayEventsParams struct {
	SessionID string `json:"session_id" required:"true" jsonschema:"description=The replay session ID"`
	EventType string `json:"event_type" jsonschema:"description=Filter by event type (click; navigation; error; network; console),enum=click|navigation|error|network|console"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of events (default 200)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetReplayErrorsParams are the parameters for the get_replay_errors tool.
type GetReplayErrorsParams struct {
	SessionID string `json:"session_id" required:"true" jsonschema:"description=The replay session ID"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// AddReplayTools registers all replay-related tools with the MCP server.
func AddReplayTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("list_replay_sessions", "List session replay recordings with optional filters for user, service, errors, and time range.", handleListReplaySessions),
		mcpobtrace.MustTool("get_replay_session", "Get metadata and summary for a specific replay session including duration, page views, and error count.", handleGetReplaySession),
		mcpobtrace.MustTool("get_replay_events", "Get the event timeline for a replay session (clicks, navigations, network requests, console logs, errors).", handleGetReplayEvents),
		mcpobtrace.MustTool("get_replay_errors", "Get all errors that occurred during a replay session.", handleGetReplayErrors),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleListReplaySessions(ctx context.Context, params ListReplaySessionsParams) (*mcp.CallToolResult, error) {
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
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.UserID != "" {
		q.Set("user_id", params.UserID)
	}
	if params.Service != "" {
		q.Set("service", params.Service)
	}
	if params.HasError {
		q.Set("has_error", "true")
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

	return doGetRequest(ctx, cfg, "/api/v1/replay/sessions?"+q.Encode())
}

func handleGetReplaySession(ctx context.Context, params GetReplaySessionParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/replay/sessions/%s", url.PathEscape(params.SessionID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleGetReplayEvents(ctx context.Context, params GetReplayEventsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.EventType != "" {
		q.Set("event_type", params.EventType)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	path := fmt.Sprintf("/api/v1/replay/sessions/%s/events", url.PathEscape(params.SessionID))
	return doGetRequest(ctx, cfg, path+"?"+q.Encode())
}

func handleGetReplayErrors(ctx context.Context, params GetReplayErrorsParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/replay/sessions/%s/errors", url.PathEscape(params.SessionID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}
