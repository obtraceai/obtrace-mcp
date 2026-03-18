# Tools Reference

## Logs

### `query_logs`
Execute a ClickHouse SQL query against the logs table (`otel_logs`).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | ClickHouse SQL query |
| `time_from` | string | Yes | Start time (RFC3339) |
| `time_to` | string | Yes | End time (RFC3339) |
| `limit` | integer | No | Max rows (default 100) |
| `project_id` | string | No | Project scope override |

### `search_logs`
Full-text search across log message bodies.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `text` | string | Yes | Search query |
| `severity` | string[] | No | Severity filter (ERROR, WARN, INFO) |
| `service` | string | No | Service name filter |
| `time_from` | string | Yes | Start time (RFC3339) |
| `time_to` | string | Yes | End time (RFC3339) |
| `limit` | integer | No | Max results (default 50) |
| `project_id` | string | No | Project scope override |

### `list_log_services`
List all service names that emitted logs.

### `log_stats`
Aggregated log statistics (counts by severity, service).

### `log_patterns`
Discover recurring log patterns using automatic clustering.

---

## Traces

### `query_traces`
Execute a ClickHouse SQL query against the traces table (`otel_traces`).

### `get_trace`
Get the full trace tree for a specific trace ID.

### `search_traces`
Search for traces by service, operation, latency, and status.

### `list_trace_services`
List services that emitted trace spans.

### `list_trace_operations`
List operations for a given service.

### `trace_stats`
Aggregated trace statistics (p50/p95/p99 latency, error rates).

---

## Metrics

### `query_metrics`
Execute a ClickHouse SQL query against the metrics table (`otel_metrics`).

### `list_metric_names`
List available metric names with optional prefix/service filter.

### `get_metric_metadata`
Get metadata for a metric (type, description, unit).

### `list_metric_label_names`
List label names for a metric.

### `list_metric_label_values`
List values for a specific label of a metric.

### `metric_stats`
Time-series statistics with configurable aggregation step.

---

## Dashboards

### `list_dashboards`
List dashboards with optional search.

### `get_dashboard`
Get full dashboard details.

### `get_dashboard_panels`
Get panel configurations and queries.

### `create_dashboard` (write)
Create a new dashboard.

### `update_dashboard` (write)
Update a dashboard.

### `delete_dashboard` (write)
Delete a dashboard.

---

## Incidents

### `list_incidents`
List incidents with status/severity/service filters.

### `get_incident`
Get full incident details.

### `get_incident_timeline`
Get the activity timeline for an incident.

### `create_incident` (write)
Create a new incident.

### `update_incident` (write)
Update incident status or severity.

### `add_incident_activity` (write)
Add a note or action to the incident timeline.

---

## Alerts

### `list_alert_rules`
List alert rules with status/severity filters.

### `get_alert_rule`
Get full alert rule details.

### `list_alert_history`
View alert firing/resolution history.

### `list_contact_points`
List notification contact points.

### `create_alert_rule` (write)
Create a new alert rule.

### `update_alert_rule` (write)
Update an alert rule.

### `delete_alert_rule` (write)
Delete an alert rule.

---

## AI

### `chat_to_query`
Convert natural language to ClickHouse queries.

### `root_cause_analysis`
AI-powered root cause analysis correlating logs, traces, and metrics.

### `summarize`
Generate system health summaries for a time window.

### `autofix` (write)
Generate code fix suggestions for errors.

---

## Projects

### `list_projects`
List accessible projects.

### `get_project`
Get project details.

### `list_apps`
List apps in a project.

### `list_teams`
List teams.

### `list_users`
List users.

---

## Replay

### `list_replay_sessions`
List session replay recordings.

### `get_replay_session`
Get replay session metadata.

### `get_replay_events`
Get the event timeline for a session.

### `get_replay_errors`
Get errors from a replay session.

---

## Search

### `global_search`
Search across all Obtrace data types.

### `generate_deeplink`
Generate deep link URLs to the Obtrace UI.
