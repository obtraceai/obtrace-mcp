package tools

import (
	"context"
	"fmt"
	"net/url"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListProjectsParams are the parameters for the list_projects tool.
type ListProjectsParams struct {
	Limit int `json:"limit" jsonschema:"description=Maximum number of projects (default 50)"`
}

// GetProjectParams are the parameters for the get_project tool.
type GetProjectParams struct {
	ProjectID string `json:"project_id" required:"true" jsonschema:"description=The project ID to retrieve"`
}

// ListAppsParams are the parameters for the list_apps tool.
type ListAppsParams struct {
	ProjectID string `json:"project_id" required:"true" jsonschema:"description=Project ID to list apps for"`
	Limit     int    `json:"limit" jsonschema:"description=Maximum number of apps (default 50)"`
}

// ListTeamsParams are the parameters for the list_teams tool.
type ListTeamsParams struct {
	Limit int `json:"limit" jsonschema:"description=Maximum number of teams (default 50)"`
}

// ListUsersParams are the parameters for the list_users tool.
type ListUsersParams struct {
	Limit int `json:"limit" jsonschema:"description=Maximum number of users (default 50)"`
}

// AddProjectTools registers all project/admin-related tools with the MCP server.
func AddProjectTools(s *server.MCPServer, enableWriteTools bool) {
	tools := []mcpobtrace.Tool{
		mcpobtrace.MustTool("list_projects", "List all projects accessible to the authenticated user.", handleListProjects),
		mcpobtrace.MustTool("get_project", "Get full details of a specific project including its apps, environments, and settings.", handleGetProject),
		mcpobtrace.MustTool("list_apps", "List all applications within a project.", handleListApps),
		mcpobtrace.MustTool("list_teams", "List all teams in the organization.", handleListTeams),
		mcpobtrace.MustTool("list_users", "List all users in the organization.", handleListUsers),
	}

	mcpobtrace.AddTools(s, tools)
}

func handleListProjects(ctx context.Context, params ListProjectsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limit))

	return doGetRequest(ctx, cfg, "/api/v1/projects?"+q.Encode())
}

func handleGetProject(ctx context.Context, params GetProjectParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(params.ProjectID))
	return doGetRequest(ctx, cfg, path)
}

func handleListApps(ctx context.Context, params ListAppsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limit))

	path := fmt.Sprintf("/api/v1/projects/%s/apps", url.PathEscape(params.ProjectID))
	return doGetRequest(ctx, cfg, path+"?"+q.Encode())
}

func handleListTeams(ctx context.Context, params ListTeamsParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limit))

	return doGetRequest(ctx, cfg, "/api/v1/teams?"+q.Encode())
}

func handleListUsers(ctx context.Context, params ListUsersParams) (*mcp.CallToolResult, error) {
	cfg, err := mcpobtrace.ObtraceConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", limit))

	return doGetRequest(ctx, cfg, "/api/v1/users?"+q.Encode())
}
