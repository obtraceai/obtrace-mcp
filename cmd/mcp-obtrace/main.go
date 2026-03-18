// Package main is the entry point for the Obtrace MCP server.
//
// It supports three transport modes:
//   - stdio: For local CLI and IDE integration (default)
//   - sse: Server-Sent Events over HTTP
//   - streamable-http: Streamable HTTP transport
//
// Usage:
//
//	mcp-obtrace [flags]
//
// Flags:
//
//	--transport      Transport mode: stdio, sse, streamable-http (default: stdio)
//	--addr           Listen address for sse/streamable-http (default: :8000)
//	--enabled-tools  Comma-separated list of tool categories to enable (default: all)
//	--enable-write   Enable write/mutating tools (default: false)
//	--version        Print version and exit
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mcpobtrace "github.com/obtraceai/obtrace-mcp"
	"github.com/obtraceai/obtrace-mcp/internal/tools"

	"github.com/mark3labs/mcp-go/server"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio, sse, streamable-http")
	addr := flag.String("addr", ":8000", "Listen address for sse/streamable-http transports")
	enabledTools := flag.String("enabled-tools", "", "Comma-separated list of tool categories to enable (default: all)")
	enableWrite := flag.Bool("enable-write", false, "Enable write/mutating tools (create, update, delete)")
	showVersion := flag.Bool("version", false, "Print version and exit")

	// Category-specific disable flags
	disableLogs := flag.Bool("disable-logs", false, "Disable log tools")
	disableTraces := flag.Bool("disable-traces", false, "Disable trace tools")
	disableMetrics := flag.Bool("disable-metrics", false, "Disable metric tools")
	disableDashboards := flag.Bool("disable-dashboards", false, "Disable dashboard tools")
	disableIncidents := flag.Bool("disable-incidents", false, "Disable incident tools")
	disableAlerts := flag.Bool("disable-alerts", false, "Disable alert tools")
	disableAI := flag.Bool("disable-ai", false, "Disable AI tools")
	disableProjects := flag.Bool("disable-projects", false, "Disable project/admin tools")
	disableReplay := flag.Bool("disable-replay", false, "Disable replay tools")
	disableSearch := flag.Bool("disable-search", false, "Disable search tools")

	flag.Parse()

	if *showVersion {
		fmt.Printf("mcp-obtrace %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Determine enabled categories
	enabled := resolveCategories(*enabledTools, map[mcpobtrace.Category]bool{
		mcpobtrace.CategoryLogs:       *disableLogs,
		mcpobtrace.CategoryTraces:     *disableTraces,
		mcpobtrace.CategoryMetrics:    *disableMetrics,
		mcpobtrace.CategoryDashboards: *disableDashboards,
		mcpobtrace.CategoryIncidents:  *disableIncidents,
		mcpobtrace.CategoryAlerts:     *disableAlerts,
		mcpobtrace.CategoryAI:         *disableAI,
		mcpobtrace.CategoryProjects:   *disableProjects,
		mcpobtrace.CategoryReplay:     *disableReplay,
		mcpobtrace.CategorySearch:     *disableSearch,
	})

	// Create MCP server
	s := server.NewMCPServer(
		"obtrace-mcp",
		version,
		server.WithInstructions(serverInstructions),
	)

	// Register tools by category
	if enabled[mcpobtrace.CategoryLogs] {
		tools.AddLogTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryTraces] {
		tools.AddTraceTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryMetrics] {
		tools.AddMetricTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryDashboards] {
		tools.AddDashboardTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryIncidents] {
		tools.AddIncidentTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryAlerts] {
		tools.AddAlertTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryAI] {
		tools.AddAITools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryProjects] {
		tools.AddProjectTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategoryReplay] {
		tools.AddReplayTools(s, *enableWrite)
	}
	if enabled[mcpobtrace.CategorySearch] {
		tools.AddSearchTools(s, *enableWrite)
	}

	// Run with the selected transport
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch *transport {
	case "stdio":
		if err := runStdio(ctx, s); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	case "sse":
		if err := runSSE(ctx, s, *addr); err != nil {
			log.Fatalf("sse server error: %v", err)
		}
	case "streamable-http":
		if err := runStreamableHTTP(ctx, s, *addr); err != nil {
			log.Fatalf("streamable-http server error: %v", err)
		}
	default:
		log.Fatalf("unknown transport: %s", *transport)
	}
}

func runStdio(ctx context.Context, s *server.MCPServer) error {
	ctxFunc, err := mcpobtrace.StdioContextFunc()
	if err != nil {
		return fmt.Errorf("initializing stdio context: %w", err)
	}

	stdioServer := server.NewStdioServer(s)
	stdioServer.SetContextFunc(ctxFunc)

	return stdioServer.Listen(ctx, os.Stdin, os.Stdout)
}

func runSSE(ctx context.Context, s *server.MCPServer, addr string) error {
	sseServer := server.NewSSEServer(s,
		server.WithBaseURL(fmt.Sprintf("http://localhost%s", addr)),
	)

	log.Printf("Starting SSE server on %s", addr)
	return sseServer.Start(addr)
}

func runStreamableHTTP(ctx context.Context, s *server.MCPServer, addr string) error {
	httpServer := server.NewStreamableHTTPServer(s)

	log.Printf("Starting Streamable HTTP server on %s", addr)
	return httpServer.Start(addr)
}

func resolveCategories(enabledStr string, disabled map[mcpobtrace.Category]bool) map[mcpobtrace.Category]bool {
	result := make(map[mcpobtrace.Category]bool)

	if enabledStr != "" {
		// Only enable explicitly listed categories
		for _, name := range strings.Split(enabledStr, ",") {
			cat := mcpobtrace.Category(strings.TrimSpace(name))
			result[cat] = true
		}
	} else {
		// Enable all categories by default
		for _, cat := range mcpobtrace.AllCategories() {
			result[cat] = true
		}
	}

	// Apply per-category disable flags
	for cat, isDisabled := range disabled {
		if isDisabled {
			delete(result, cat)
		}
	}

	return result
}

const serverInstructions = `You are connected to an Obtrace MCP server that provides access to an observability platform.

Available capabilities:
- **Logs**: Query, search, and analyze application logs stored in ClickHouse (table: otel_logs)
- **Traces**: Query distributed traces, search spans, get full trace trees, and analyze latency (table: otel_traces)
- **Metrics**: Query time-series metrics, list metric names/labels, and get statistics (table: otel_metrics)
- **Dashboards**: List, view, create, update, and delete observability dashboards
- **Incidents**: Manage incidents including creation, status updates, and timeline activities
- **Alerts**: Manage alert rules, view alert history, and configure contact points
- **AI**: Natural language to query conversion, root cause analysis, autofix suggestions, and system health summaries
- **Projects**: List projects, apps, teams, and users
- **Replay**: Browse and analyze session replay recordings
- **Search**: Global search across all data types and deep link generation

All data is scoped by tenant_id and project_id. Times should be in RFC3339 format.
Queries use ClickHouse SQL syntax. Table names: otel_logs, otel_traces, otel_metrics.
`
