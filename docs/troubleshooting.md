# Troubleshooting

## Connection Issues

### "OBTRACE_URL environment variable is required"
Set the `OBTRACE_URL` environment variable to your Obtrace API base URL:
```bash
export OBTRACE_URL=https://api.obtrace.ai
```

### "OBTRACE_API_KEY environment variable is required"
Set the `OBTRACE_API_KEY` environment variable:
```bash
export OBTRACE_API_KEY=your-api-key
```

### HTTP 401 Unauthorized
- Verify your API key is valid and not expired.
- Check that the key has permissions for the requested operation.
- Ensure `OBTRACE_TENANT_ID` matches the tenant the key belongs to.

### HTTP 403 Forbidden
- The API key may not have access to the requested project.
- Write operations require appropriate permissions and `--enable-write` flag.

### TLS Errors
- Verify the CA certificate path is correct.
- For self-signed certificates, set `OBTRACE_TLS_INSECURE=true` (dev only).
- For mTLS, ensure both client cert and key are provided.

## Tool Issues

### "No results returned"
- Verify the time range covers the expected data window.
- Check that `project_id` matches where your data is ingested.
- For ClickHouse queries, use the correct table names: `otel_logs`, `otel_traces`, `otel_metrics`.

### Write tools not available
Write tools require the `--enable-write` flag:
```bash
mcp-obtrace --enable-write
```

### Specific category tools missing
Check that the category is not disabled. Use `--enabled-tools` to list specific categories:
```bash
mcp-obtrace --enabled-tools logs,traces,metrics,incidents
```

## Performance

### Slow queries
- Use narrower time ranges.
- Add `LIMIT` clauses to ClickHouse queries.
- Use specific filters (service, severity) to reduce data scanned.

### Timeout errors
The default HTTP timeout is 30 seconds. For large queries, consider narrowing the scope or using more specific filters.
