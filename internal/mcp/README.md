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
| `internal/api/mcp.go` | Gin HTTP handler with auth middleware, route registration |
| `internal/commands/mcp.go` | CLI command (`photoprism mcp serve`) using stdio transport |
| `internal/auth/acl/` | `ResourceMCP` constant and ACL grant rules (`GrantFullAccess` for admin; `GrantSearchAll` for manager in Pro/Portal and for the API client roles: client, instance, service, portal) |

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

### MCP Client Compatibility

Most MCP clients natively support Streamable HTTP with custom headers (`url` + `headers` in config). Clients that only support stdio-based servers in their config file require [`mcp-remote`](https://www.npmjs.com/package/mcp-remote) as a stdio-to-HTTP bridge.

**Direct HTTP config** (clients with native support):

```json
{
  "mcpServers": {
    "photoprism": {
      "url": "http://localhost:2342/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer <admin-token>"
      }
    }
  }
}
```

**stdio bridge** (clients without native HTTP support):

```json
{
  "mcpServers": {
    "photoprism": {
      "command": "npx",
      "args": [
        "-y", "mcp-remote",
        "http://localhost:2342/api/v1/mcp",
        "--header", "Authorization:${AUTH_HEADER}"
      ],
      "env": {
        "AUTH_HEADER": "Bearer <admin-token>"
      }
    }
  }
}
```

### Authorization

The HTTP endpoint uses PhotoPrism's existing ACL system:

- **Resource:** `ResourceMCP` (`"mcp"`)
- **Permission:** `ActionView` for read-only tools (handler-level check)
- **Grants:**
  - `RoleAdmin` → `GrantFullAccess` in every edition.
  - `RoleManager` → `GrantSearchAll` in Pro and Portal builds (the role does not exist in CE/Plus).
  - `RoleClient`, `RoleInstance`, `RoleService`, `RolePortal` → `GrantSearchAll` in every edition.
  - All other roles (`user`, `viewer`, `guest`, `visitor`, `contributor`, default) are denied.
- **Why `GrantSearchAll` for non-admins?** It includes `AccessAll`, `ActionView`, and `ActionSearch` — exactly what the read-only tools need — but excludes `ActionManage`/`ActionUpdate`/`ActionDelete`/`ActionCreate`. Any future write-capable MCP tool gated on those permissions will automatically be admin-only without needing per-tool checks.
- **Client tokens:** API client sessions must also include the `mcp` resource (or a wildcard) in their session scope; the ACL grant alone is not sufficient.
- **Auth model:** Request-level (every HTTP request runs through `Auth()`)
- **Public mode:** Blocked (returns 403)
- **Experimental gate:** the route only registers when `--experimental` is enabled; otherwise `/api/v1/mcp` returns 404.

### How Users Get Access

Regular user accounts (`RoleUser`, `RoleViewer`, etc.) are intentionally **not** in the `ResourceMCP` ACL. Regular users typically don't have shell access to the server, so they can't run the CLI commands themselves — and the prototype's tools only return static reference data, so there's no per-user information to authorize against. Access is therefore granted through admin-issued client tokens.

To onboard a user (or a CI job, IDE, etc.), an administrator runs the following on the PhotoPrism server:

```bash
./photoprism clients add \
  --name "Alice's IDE" \
  --scope mcp \
  --role client \
  --expires 2592000          # 30 days; use -1 for no expiry
```

The command prints a client ID and secret. Combine them into a bearer token (or pass them through the OAuth2 client-credentials flow) and paste the resulting value into the user's MCP client config — the same JSON snippets shown above under *MCP Client Compatibility* apply unchanged. Replace `<admin-token>` with the issued token.

To revoke access without disabling the user account, the administrator runs:

```bash
./photoprism clients remove <client-id>
```

> **Heads up:** `photoprism auth add --scope mcp <username>` creates an *app password* tied to a user account, but it currently does **not** grant MCP access — `RoleUser` is not in the `ResourceMCP` ACL. Use `photoprism clients add` for MCP integrations until that policy changes.

When MCP eventually grows tools that need user-scoped data (e.g. "list my albums"), the team will revisit the policy and likely add `RoleUser → GrantSearchAll` so the app-password path lights up. Until then, every MCP integration is an admin-provisioned client token tied to a named application.
