package tools

import (
	"bytes"
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

// ListDashboardsParams are the parameters for the list_dashboards tool.
type ListDashboardsParams struct {
	Query     string `json:"query" jsonschema:"description=Search query to filter dashboards by name or description"`
	ProjectID string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of dashboards to return (default 50)"`
}

// GetDashboardParams are the parameters for the get_dashboard tool.
type GetDashboardParams struct {
	DashboardID string `json:"dashboard_id" required:"true" jsonschema:"description=The dashboard ID or UID to retrieve"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// GetDashboardPanelsParams are the parameters for the get_dashboard_panels tool.
type GetDashboardPanelsParams struct {
	DashboardID string `json:"dashboard_id" required:"true" jsonschema:"description=The dashboard ID or UID"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to scope the query"`
}

// CreateDashboardParams are the parameters for the create_dashboard tool.
type CreateDashboardParams struct {
	Name        string `json:"name" required:"true" jsonschema:"description=Dashboard name"`
	Description string `json:"description" jsonschema:"description=Dashboard description"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to create the dashboard in"`
}

// UpdateDashboardParams are the parameters for the update_dashboard tool.
type UpdateDashboardParams struct {
	DashboardID string `json:"dashboard_id" required:"true" jsonschema:"description=The dashboard ID or UID to update"`
	Name        string `json:"name" jsonschema:"description=New dashboard name"`
	Description string `json:"description" jsonschema:"description=New dashboard description"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to scope the update"`
}

// DeleteDashboardParams are the parameters for the delete_dashboard tool.
type DeleteDashboardParams struct {
	DashboardID string `json:"dashboard_id" required:"true" jsonschema:"description=The dashboard ID or UID to delete"`
	ProjectID   string `json:"project_id" jsonschema:"description=Project ID to scope the deletion"`
}

// AddDashboardTools registers all dashboard-related tools with the MCP server.
func AddDashboardTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("list_dashboards", "List dashboards with optional search query. Returns dashboard ID, name, description, and metadata.", handleListDashboards),
		mcpobtrace.MustTool("get_dashboard", "Get full dashboard details including all panel configurations and queries.", handleGetDashboard),
		mcpobtrace.MustTool("get_dashboard_panels", "Get the panel configurations and queries for a specific dashboard.", handleGetDashboardPanels),
	}

	if enableWriteTools {
		tools = append(tools,
			mcpobtrace.MustTool("create_dashboard", "Create a new dashboard with the given name and description.", handleCreateDashboard),
			mcpobtrace.MustTool("update_dashboard", "Update an existing dashboard's name or description.", handleUpdateDashboard),
			mcpobtrace.MustTool("delete_dashboard", "Delete a dashboard by its ID.", handleDeleteDashboard),
		)
	}

	mcpobtrace.AddTools(s, tools)
}

func handleListDashboards(ctx context.Context, params ListDashboardsParams) (*mcp.CallToolResult, error) {
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
	if params.Query != "" {
		q.Set("query", params.Query)
	}
	if projectID != "" {
		q.Set("project_id", projectID)
	}

	return doGetRequest(ctx, cfg, "/api/v1/dashboards?"+q.Encode())
}

func handleGetDashboard(ctx context.Context, params GetDashboardParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/dashboards/%s", url.PathEscape(params.DashboardID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleGetDashboardPanels(ctx context.Context, params GetDashboardPanelsParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/dashboards/%s/panels", url.PathEscape(params.DashboardID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doGetRequest(ctx, cfg, path)
}

func handleCreateDashboard(ctx context.Context, params CreateDashboardParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID := params.ProjectID
	if projectID == "" {
		projectID = cfg.ProjectID
	}

	body := map[string]any{
		"name":        params.Name,
		"description": params.Description,
	}
	if projectID != "" {
		body["project_id"] = projectID
	}

	return doPostRequest(ctx, cfg, "/api/v1/dashboards", body)
}

func handleUpdateDashboard(ctx context.Context, params UpdateDashboardParams) (*mcp.CallToolResult, error) {
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
	if projectID != "" {
		body["project_id"] = projectID
	}

	path := fmt.Sprintf("/api/v1/dashboards/%s", url.PathEscape(params.DashboardID))
	return doPutRequest(ctx, cfg, path, body)
}

func handleDeleteDashboard(ctx context.Context, params DeleteDashboardParams) (*mcp.CallToolResult, error) {
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

	path := fmt.Sprintf("/api/v1/dashboards/%s", url.PathEscape(params.DashboardID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	return doDeleteRequest(ctx, cfg, path)
}

// doPostRequest is a shared helper for making POST requests to the Obtrace API.
func doPostRequest(ctx context.Context, cfg *mcpobtrace.ObtraceConfig, path string, body any) (*mcp.CallToolResult, error) {
	return doMutatingRequest(ctx, cfg, http.MethodPost, path, body)
}

// doPutRequest is a shared helper for making PUT requests to the Obtrace API.
func doPutRequest(ctx context.Context, cfg *mcpobtrace.ObtraceConfig, path string, body any) (*mcp.CallToolResult, error) {
	return doMutatingRequest(ctx, cfg, http.MethodPut, path, body)
}

// doDeleteRequest is a shared helper for making DELETE requests to the Obtrace API.
func doDeleteRequest(ctx context.Context, cfg *mcpobtrace.ObtraceConfig, path string) (*mcp.CallToolResult, error) {
	return doMutatingRequest(ctx, cfg, http.MethodDelete, path, nil)
}

// doMutatingRequest handles POST/PUT/DELETE requests.
func doMutatingRequest(ctx context.Context, cfg *mcpobtrace.ObtraceConfig, method, path string, body any) (*mcp.CallToolResult, error) {
	client := mcpobtrace.ObtraceClientFromContext(ctx)

	reqURL := cfg.URL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if cfg.TenantID != "" {
		req.Header.Set("X-Obtrace-Tenant-ID", cfg.TenantID)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))), nil
	}

	// Pretty-print JSON if possible
	var prettyJSON json.RawMessage
	if err := json.Unmarshal(respBody, &prettyJSON); err == nil {
		formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
		if err == nil {
			respBody = formatted
		}
	}

	return mcp.NewToolResultText(string(respBody)), nil
}
