# Security

## Authentication

The MCP server authenticates to the Obtrace API using an API key passed via the `OBTRACE_API_KEY` environment variable. The key is sent as a `Bearer` token in the `Authorization` header.

- Never hardcode API keys in configuration files or code.
- Use environment variables or secret managers.
- Use distinct API keys per environment (dev, staging, prod).

## Write Tool Gating

Write/mutating tools (create, update, delete operations) are **disabled by default**. They must be explicitly enabled with the `--enable-write` flag:

```bash
mcp-obtrace --enable-write
```

This prevents accidental modifications when the MCP server is used in read-only contexts.

## TLS

For production deployments:

- Use TLS with a valid CA certificate.
- Consider mTLS for additional authentication.
- Never use `OBTRACE_TLS_INSECURE=true` in production.

## Scoping

All API requests are scoped by:

- **Tenant ID**: Set via `OBTRACE_TENANT_ID` or per-tool `project_id` parameter.
- **Project ID**: Set via `OBTRACE_PROJECT_ID` or per-tool `project_id` parameter.

This ensures that MCP tools can only access data within the configured scope.

## Best Practices

1. Run the MCP server with the minimum required tool categories enabled.
2. Use read-only API keys when write tools are not needed.
3. Monitor API key usage via audit logs.
4. Rotate API keys regularly.
5. Use network segmentation to restrict MCP server access to authorized clients.
