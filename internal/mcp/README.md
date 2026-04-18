## PhotoPrism MCP Server

**Last Updated:** April 13, 2026

> See `specs/platform/mcp.md` for the canonical specification, including the rationale for the user-access policy and the role/grant matrix per edition.

### Current Capabilities

- **Transports:**
  - CLI: `photoprism mcp serve` (stdio, no auth; development and testing)
  - HTTP: `POST/GET/DELETE /api/v1/mcp` (Streamable HTTP, authenticated)
- **Authorization:** HTTP endpoint enforces the `ResourceMCP` ACL (admin plus the API client roles in every edition, manager in Pro/Portal); anonymous access is permitted in public mode for the currently registered read-only tools.
- **Feature gate:** HTTP endpoint requires the `--experimental` flag; absent that flag `/api/v1/mcp` returns 404.
- Read-only resources:
  - `photoprism://config-options`
  - `photoprism://search-filters`
- Read-only tools:
  - `list_config_keys`
  - `find_search_filters`

### Package layout

| Package                    | Purpose                                                                                                                                                                              |
|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/mcp/`            | Core MCP logic: server factory, data pipeline, resources, tools                                                                                                                      |
| `internal/api/mcp.go`      | Gin HTTP handler with auth middleware, route registration                                                                                                                            |
| `internal/commands/mcp.go` | CLI command (`photoprism mcp serve`) using stdio transport                                                                                                                           |
| `internal/auth/acl/`       | `ResourceMCP` constant and ACL grant rules (`GrantFullAccess` for admin; `GrantSearchAll` for manager in Pro/Portal and for the API client roles: client, instance, service, portal) |

### Scope

In scope for the current server:

- Reuse existing internal reference data (`config.Flags`, `config.OptionsReportSections`, `form.Report(&form.SearchPhotos{})`) instead of maintaining a parallel dataset.
- Keep outputs compact enough for LLM consumption.
- Authenticated remote access via the Streamable HTTP transport, plus a stdio transport for local development and testing.

Out of scope for the current server (must not regress without additional per-tool gates):

- Write-capable tools.
- Direct database access.
- Live PhotoPrism instance or API queries.
- Per-user state (albums, photos, sessions, settings).

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
- **Auth model:** request-level. The handler runs `Auth(c, acl.ResourceMCP, acl.ActionView)` followed by `s.Abort(c)`, which writes the matching status code (`401` unauthenticated, `403` ACL deny, `429` rate-limited) and returns `true` so the handler can `return` early.
- **Public mode:** anonymous access is permitted. In public mode, `api.Session()` returns the default public session (effectively admin), so `Auth(...)` passes and the currently registered read-only tools are reachable without a token. This is an intentional, narrow allowance for demo deployments (`demo.photoprism.app`); it is safe only because every registered tool today returns static reference metadata derived from `config.Flags` and `form.Report(&form.SearchPhotos{})` — no database access, no per-user state, no secrets, no mutations. **Any future tool that touches per-user state, the database, or mutates anything MUST NOT be registered on this server without an additional per-tool check**. See *Extending the Tool Surface* below.
- **Experimental gate:** the route only registers when `--experimental` is enabled; otherwise `/api/v1/mcp` returns 404. Public-mode anonymous access therefore also requires the operator to explicitly opt into experimental features.

### Rate Limiting

The MCP handler does not install a custom rate limiter — there is no per-endpoint bucket. Coverage depends on the edition:

| Build  | Generic per-IP HTTP limiter? | Notes                                                                                                                                                                                                                                                        |
|--------|------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| CE     | no                           | Only the experimental gate and the admin/client auth check protect the endpoint. In public mode all callers are anonymous, so CE deployments that run with `--public --experimental` have no per-request throttle in front of MCP.                           |
| Plus   | no                           | Same as CE. The IPS middleware exists but only consumes tokens on known scanner/exploit paths, so it does not throttle MCP traffic.                                                                                                                          |
| Pro    | yes                          | `pro/internal/server/register.go` calls `router.Use(limiter.Middleware(limiter.NewLimit(rate.Every(secOpt.RequestInterval), secOpt.RequestLimit)))` when both options are set. The limiter is per client IP and applies to every API endpoint, MCP included. |
| Portal | yes                          | Same wiring as Pro in `portal/internal/server/register.go`.                                                                                                                                                                                                  |

A per-endpoint limiter (via `limiter.Auth` / `limiter.Login` / `limiter.AbortJSON`) is only worth adding when MCP grows write-capable tools or endpoints that warrant stricter throttling than the generic IP limiter — for example, anything that mutates state or that triggers expensive backend work.

### Scope Plumbing

`mcp` is the canonical scope token for `ResourceMCP`. The relevant pieces:

- **Sanitization:** `pkg/clean.Scope` lowercases the input and parses it through `pkg/list.ParseAttr`. There is no allowlist of valid scope tokens, so `--scope mcp`, `--scope "mcp metrics"`, and `--scope "*"` are all accepted by `clients add` and `auth add` without any registry update.
- **Authorization:** `internal/auth/acl.ScopePermits` checks the parsed attribute list against the resource string. Because `acl.ResourceMCP.String() == "mcp"`, the existing `attr.Contains(...)` path matches a session that holds the `mcp` token. See `TestScopePermits/MCPScope` in `internal/auth/acl/scope_test.go` for the canonical assertions (admin token, mixed scopes, case-insensitivity, deny on unrelated scopes).
- **Cluster JWTs:** instance-side validation runs through `Config.JWTAllowedScopes()` (`internal/api/api_auth_jwt.go`). The default allowlist is `DefaultJWTAllowedScopes = "config cluster vision metrics mcp"` in `internal/config/config_cluster.go`, so portal-issued JWTs with `scope=mcp` are accepted out of the box. Operators that override the list via `--jwt-scope` / `PHOTOPRISM_JWT_SCOPE` need to include `mcp` themselves.

### Extending the Tool Surface

Anonymous access in public mode is only safe as long as every registered tool returns static reference metadata. Before adding a new tool, confirm it fits the existing contract:

- No database reads or writes.
- No per-user state (albums, photos, sessions, settings).
- No filesystem, network, or subprocess side effects.
- No access to secrets or runtime config values (only flag schema/defaults are allowed).

If a proposed tool does **not** fit that contract, do not register it on the default `*sdkmcp.Server`. Instead, take one of these paths (ordered by how much work they are):

1. **Two servers, one factory.** `internal/api/mcp.go` already passes a factory to `sdkmcp.NewStreamableHTTPHandler`. Build a second server with the full tool set and return it from the factory only when the request is non-public and authenticated; keep the default server restricted to the public-safe tools. Policy becomes declarative at construction time, and a missing tool in the public server surfaces as a standard "tool not found" from the SDK rather than leaking its existence.
2. **Per-tool context checks.** `sdkmcp.AddTool` closures receive a `*sdkmcp.CallToolRequest`; stash caller context on the MCP session at `initialize` and reject inside the tool closure when public mode is active or the ACL deny list applies. Use this when the same tool has different output per caller (e.g. admin sees raw values, client sees redacted).
3. **SDK middleware.** For cross-cutting concerns such as per-tool rate limits or structured audit entries, wire an `sdkmcp.Middleware` that inspects the JSON-RPC method and tool name before dispatch.

Whichever path you pick, **add a test in `internal/mcp/server_test.go` that fails if the restricted tool shows up in `tools/list` or is callable over the public path**, and update the *Available Tools* table below.

### How Users Get Access

Regular user accounts (`RoleUser`, `RoleViewer`, etc.) are intentionally **not** in the `ResourceMCP` ACL. Regular users typically don't have shell access to the server, so they can't run the CLI commands themselves — and the currently registered tools only return static reference data, so there's no per-user information to authorize against. Access is therefore granted through admin-issued client tokens.

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

> **Heads up:** `photoprism auth add --scope mcp <username>` creates an *app password* tied to a user account, but it currently does **not** grant MCP access — `RoleUser` is not in the `ResourceMCP` ACL. Use `photoprism clients add` for MCP integrations until that policy changes. The reasoning is documented in `specs/platform/mcp.md` under *User Access Model* (deliberate hold, not an oversight).

When MCP eventually grows tools that need user-scoped data (e.g. "list my albums"), the team will revisit the policy and likely add `RoleUser → GrantSearchAll` so the app-password path lights up. Until then, every MCP integration is an admin-provisioned client token tied to a named application.
