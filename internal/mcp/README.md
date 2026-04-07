## PhotoPrism MCP Prototype

**Last Updated:** April 7, 2026

### Current capabilities

- **Transports:**
  - CLI: `photoprism mcp serve` (stdio, no auth)
  - HTTP: `POST/GET/DELETE /api/v1/mcp` (Streamable HTTP, authenticated)
- **Authentication:** HTTP endpoint requires admin role via `ResourceMCP` ACL
- **Feature gate:** HTTP endpoint requires `--experimental` flag
- Read-only resources:
  - `photoprism://config-options`
  - `photoprism://search-filters`
- Read-only tools:
  - `list_config_keys`
  - `find_search_filters`

### Package layout

| Package | Purpose |
|---------|---------|
| `internal/mcp/` | Core MCP logic: server factory, data pipeline, resources, tools |
| `internal/api/mcp_serve.go` | Gin HTTP handler with auth middleware, route registration |
| `internal/commands/mcp.go` | CLI command (`photoprism mcp serve`) using stdio transport |
| `internal/auth/acl/` | `ResourceMCP` constant and admin-only grant rules |

### Goals and non-goals

Goals:

- Prove the MCP model works end-to-end inside the PhotoPrism codebase
- Reuse internal reference data instead of maintaining a separate copy
- Keep outputs concise enough for LLM use
- Provide authenticated remote access via Streamable HTTP transport

Non-Goals:

- No write-capable tools
- No direct database access
- No live PhotoPrism instance or API queries
- No non-admin access

### Internal data sources

- Config options: `internal/config.Flags` plus `internal/config.OptionsReportSections`
- Search filters: `internal/form.Report(&form.SearchPhotos{})`

### Run locally (stdio)

Build the CLI:

```bash
go build ./cmd/photoprism
```

Start the MCP server over stdio:

```bash
./photoprism mcp serve
```

The process waits for an MCP client on stdin/stdout. Logs are written to stderr so the MCP message stream stays valid.

### Run via HTTP

Start PhotoPrism with experimental mode enabled:

```bash
./photoprism --experimental start
```

The MCP endpoint is available at `/api/v1/mcp`. Authenticate with an admin token:

```bash
# Initialize session
curl -X POST http://localhost:2342/api/v1/mcp \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"curl","version":"1.0"}}}'
```

### Test with MCP inspector

Stdio transport:

```bash
npx @modelcontextprotocol/inspector ./photoprism mcp serve
```

Useful smoke tests:

- List resources
- Read `photoprism://config-options`
- Read `photoprism://search-filters`
- Call `list_config_keys` with `{"query":"http","limit":3}`
- Call `find_search_filters` with `{"query":"Berlin","limit":5}`

### Available resources

`photoprism://config-options`

- JSON payload with `edition` and `items`
- Each item includes `section`, `environment`, `cli_flag`, `default`, and `description`

`photoprism://search-filters`

- JSON payload with `edition` and `items`
- Each item includes `filter`, `type`, `examples`, and `notes`

### Available tools

`list_config_keys`

- Inputs: `section`, `query`, `edition`, `limit`
- Returns matching config rows with environment variables, CLI flags, defaults, descriptions, and a conservative `edition_support` hint
- Validation rejects unsupported `edition` values

`find_search_filters`

- Inputs: `query`, `type`, `limit`
- Returns matching search filters with examples and notes
- Validation rejects unsupported filter `type` values

### Authorization

The HTTP endpoint uses PhotoPrism's existing ACL system:

- **Resource:** `ResourceMCP` (`"mcp"`)
- **Permission:** `ActionView` for read-only tools
- **Roles:** Admin-only (`GrantFullAccess`)
- **Auth model:** Request-level (every HTTP request runs through `Auth()`)
- **Public mode:** Blocked (returns 403)
