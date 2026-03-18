package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListIncidentsParams are the parameters for the list_incidents tool.
type ListIncidentsParams struct {
	Status    string `json:"status" jsonschema:"description=Filter by status (open; acknowledged; resolved; closed),enum=open|acknowledged|resolved|closed"`
	Severity  string `json:"severity" jsonschema:"description=Filter by severity (critical; high; medium; low),enum=critical|high|medium|low"`
	Service   string `json:"service" jsonschema:"description=Filter by affected service"`
	TimeFrom  string `json:"time_from" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of incidents (default 50)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetIncidentParams are the parameters for the get_incident tool.
type GetIncidentParams struct {
	IncidentID string `json:"incident_id" required:"true" jsonschema:"description=The incident ID to retrieve"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetIncidentTimelineParams are the parameters for the get_incident_timeline tool.
type GetIncidentTimelineParams struct {
	IncidentID string `json:"incident_id" required:"true" jsonschema:"description=The incident ID"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// CreateIncidentParams are the parameters for the create_incident tool.
type CreateIncidentParams struct {
	Title       string `json:"title" required:"true" jsonschema:"description=Incident title"`
	Description string `json:"description" jsonschema:"description=Detailed description of the incident"`
	Severity    string `json:"severity" required:"true" jsonschema:"description=Severity level,enum=critical|high|medium|low"`
	Service     string `json:"service" jsonschema:"description=Affected service name"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to create the incident in"`
}

// UpdateIncidentParams are the parameters for the update_incident tool.
type UpdateIncidentParams struct {
	IncidentID string `json:"incident_id" required:"true" jsonschema:"description=The incident ID to update"`
	Status     string `json:"status" jsonschema:"description=New status,enum=open|acknowledged|resolved|closed"`
	Severity   string `json:"severity" jsonschema:"description=New severity,enum=critical|high|medium|low"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the update"`
}

// AddIncidentActivityParams are the parameters for the add_incident_activity tool.
type AddIncidentActivityParams struct {
	IncidentID string `json:"incident_id" required:"true" jsonschema:"description=The incident ID"`
	Message    string `json:"message" required:"true" jsonschema:"description=Activity message or note to add"`
	Type       string `json:"type" jsonschema:"description=Activity type (note; status_change; action),enum=note|status_change|action"`
	ProjectID  string `json:"project_id" jsonschema:"description=Project ID to scope the operation"`
}

// AddIncidentTools registers all incident-related tools with the MCP server.
func AddIncidentTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("list_incidents", "List incidents with optional filters for status, severity, service, and time range.", handleListIncidents),
		mcpobtrace.MustTool("get_incident", "Get full details of a specific incident including affected services, root cause, and current status.", handleGetIncident),
		mcpobtrace.MustTool("get_incident_timeline", "Get the activity timeline for an incident showing all status changes, notes, and actions.", handleGetIncidentTimeline),
	}

	if enableWriteTools {
		tools = append(tools,
			mcpobtrace.MustTool("create_incident", "Create a new incident with title, description, severity, and affected service.", handleCreateIncident),
			mcpobtrace.MustTool("update_incident", "Update an incident's status or severity.", handleUpdateIncident),
			mcpobtrace.MustTool("add_incident_activity", "Add a note, status change, or action to an incident's timeline.", handleAddIncidentActivity),
		)
	}

	mcpobtrace.AddTools(s, tools)
}

func handleListIncidents(ctx context.Context, params ListIncidentsParams) (*mcp.CallToolResult, error) {
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
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Severity != "" {
		q.Set("severity", params.Severity)
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

	return doGetRequest(ctx, cfg, "/api/v1/incidents?"+q.Encode())
}

func handleGetIncident(ctx context.Context, params GetIncidentParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/incidents/%s", url.PathEscape(params.IncidentID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleGetIncidentTimeline(ctx context.Context, params GetIncidentTimelineParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/incidents/%s/timeline", url.PathEscape(params.IncidentID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleCreateIncident(ctx context.Context, params CreateIncidentParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"title":    params.Title,
		"severity": params.Severity,
	}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if params.Service != "" {
		body["service"] = params.Service
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/incidents", body)
}

func handleUpdateIncident(ctx context.Context, params UpdateIncidentParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{}
	if params.Status != "" {
		body["status"] = params.Status
	}
	if params.Severity != "" {
		body["severity"] = params.Severity
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	path := fmt.Sprintf("/api/v1/incidents/%s", url.PathEscape(params.IncidentID))
	return doPutRequest(ctx, cfg, path, body)
}

func handleAddIncidentActivity(ctx context.Context, params AddIncidentActivityParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	activityType := params.Type
	if activityType == "" {
		activityType = "note"
	}

	body := map[string]any{
		"message": params.Message,
		"type":    activityType,
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	path := fmt.Sprintf("/api/v1/incidents/%s/activities", url.PathEscape(params.IncidentID))
	return doPostRequest(ctx, cfg, path, body)
}
