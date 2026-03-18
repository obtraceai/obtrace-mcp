package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListAlertRulesParams are the parameters for the list_alert_rules tool.
type ListAlertRulesParams struct {
	Status    string `json:"status" jsonschema:"description=Filter by status (active; inactive; firing; pending),enum=active|inactive|firing|pending"`
	Severity  string `json:"severity" jsonschema:"description=Filter by severity,enum=critical|high|medium|low"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of rules (default 50)"`
}

// GetAlertRuleParams are the parameters for the get_alert_rule tool.
type GetAlertRuleParams struct {
	RuleID    string `json:"rule_id" required:"true" jsonschema:"description=The alert rule ID to retrieve"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// ListAlertHistoryParams are the parameters for the list_alert_history tool.
type ListAlertHistoryParams struct {
	RuleID    string `json:"rule_id" jsonschema:"description=Filter by alert rule ID"`
	Status    string `json:"status" jsonschema:"description=Filter by alert status (firing; resolved),enum=firing|resolved"`
	TimeFrom  string `json:"time_from" required:"true" jsonschema:"description=Start time in RFC3339 format"`
	TimeTo    string `json:"time_to" required:"true" jsonschema:"description=End time in RFC3339 format"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of entries (default 100)"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// CreateAlertRuleParams are the parameters for the create_alert_rule tool.
type CreateAlertRuleParams struct {
	Name        string `json:"name" required:"true" jsonschema:"description=Alert rule name"`
	Description string `json:"description" jsonschema:"description=Alert rule description"`
	Query       string `json:"query" required:"true" jsonschema:"description=The query/condition expression for the alert"`
	Severity    string `json:"severity" required:"true" jsonschema:"description=Alert severity,enum=critical|high|medium|low"`
	Interval    string `json:"interval" jsonschema:"description=Evaluation interval (e.g. 1m; 5m). Default 1m."`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to create the rule in"`
}

// UpdateAlertRuleParams are the parameters for the update_alert_rule tool.
type UpdateAlertRuleParams struct {
	RuleID      string `json:"rule_id" required:"true" jsonschema:"description=The alert rule ID to update"`
	Name        string `json:"name" jsonschema:"description=New rule name"`
	Description string `json:"description" jsonschema:"description=New rule description"`
	Query       string `json:"query" jsonschema:"description=New query/condition expression"`
	Severity    string `json:"severity" jsonschema:"description=New severity,enum=critical|high|medium|low"`
	Enabled     bool   `json:"enabled" jsonschema:"description=Whether the rule is enabled"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to scope the update"`
}

// DeleteAlertRuleParams are the parameters for the delete_alert_rule tool.
type DeleteAlertRuleParams struct {
	RuleID    string `json:"rule_id" required:"true" jsonschema:"description=The alert rule ID to delete"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the deletion"`
}

// ListContactPointsParams are the parameters for the list_contact_points tool.
type ListContactPointsParams struct {
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// AddAlertTools registers all alerting-related tools with the MCP server.
func AddAlertTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("list_alert_rules", "List alert rules with optional filters for status and severity.", handleListAlertRules),
		mcpobtrace.MustTool("get_alert_rule", "Get full details of a specific alert rule including its query, conditions, and notification settings.", handleGetAlertRule),
		mcpobtrace.MustTool("list_alert_history", "List alert firing/resolution history for a time range.", handleListAlertHistory),
		mcpobtrace.MustTool("list_contact_points", "List configured notification contact points (email, Slack, PagerDuty, etc.).", handleListContactPoints),
	}

	if enableWriteTools {
		tools = append(tools,
			mcpobtrace.MustTool("create_alert_rule", "Create a new alert rule with query, severity, and evaluation interval.", handleCreateAlertRule),
			mcpobtrace.MustTool("update_alert_rule", "Update an existing alert rule's properties.", handleUpdateAlertRule),
			mcpobtrace.MustTool("delete_alert_rule", "Delete an alert rule by its ID.", handleDeleteAlertRule),
		)
	}

	mcpobtrace.AddTools(s, tools)
}

func handleListAlertRules(ctx context.Context, params ListAlertRulesParams) (*mcp.CallToolResult, error) {
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
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/alerts/rules?"+q.Encode())
}

func handleGetAlertRule(ctx context.Context, params GetAlertRuleParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/alerts/rules/%s", url.PathEscape(params.RuleID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleListAlertHistory(ctx context.Context, params ListAlertHistoryParams) (*mcp.CallToolResult, error) {
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
	q.Set("time_from", params.TimeFrom)
	q.Set("time_to", params.TimeTo)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if params.RuleID != "" {
		q.Set("rule_id", params.RuleID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/alerts/history?"+q.Encode())
}

func handleListContactPoints(ctx context.Context, params ListContactPointsParams) (*mcp.CallToolResult, error) {
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

	return doGetRequest(ctx, cfg, "/api/v1/alerts/contacts?"+q.Encode())
}

func handleCreateAlertRule(ctx context.Context, params CreateAlertRuleParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	interval := params.Interval
	if interval == "" {
		interval = "1m"
	}

	body := map[string]any{
		"name":     params.Name,
		"query":    params.Query,
		"severity": params.Severity,
		"interval": interval,
	}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/alerts/rules", body)
}

func handleUpdateAlertRule(ctx context.Context, params UpdateAlertRuleParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{}
	if params.Name != "" {
		body["name"] = params.Name
	}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if params.Query != "" {
		body["query"] = params.Query
	}
	if params.Severity != "" {
		body["severity"] = params.Severity
	}
	body["enabled"] = params.Enabled
	if projectID != "" {
		body["project_id"] = projectID
	}

	path := fmt.Sprintf("/api/v1/alerts/rules/%s", url.PathEscape(params.RuleID))
	return doPutRequest(ctx, cfg, path, body)
}

func handleDeleteAlertRule(ctx context.Context, params DeleteAlertRuleParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/alerts/rules/%s", url.PathEscape(params.RuleID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doDeleteRequest(ctx, cfg, path)
}
