## PhotoPrism MCP Server

**Last Updated:** October 20, 2026

### Current Capabilities

- **Transports:**
  - CLI: `photoprism mcp serve` (stdio, no auth; development and testing). The static-data contract documented under *Authorization* and *Extending the Tool Surface* below applies — write-capable tools MUST require a session token even on stdio.
  - HTTP: `POST/GET/DELETE /api/v1/mcp` (Streamable HTTP, authenticated). Disabled together with the stdio transport via `--disable-mcp` / `PHOTOPRISM_DISABLE_MCP` / `DisableMCP`: the HTTP route responds with the standard 404 and `photoprism mcp serve` refuses to start (pass `--disable-mcp=false` on the command line to override for ad-hoc development runs). The flag is also surfaced to the frontend through `ClientConfig.Disable.MCP` (`disable.mcp`), letting the UI hide MCP-related controls while the endpoint is off.
- **Authorization:** HTTP endpoint enforces the `ResourceMCP` ACL (admin plus the API client roles in every edition, user and viewer in Plus/Pro/Portal, manager in Pro/Portal) and the per-account MCP setting (`User.CanUseMCP`); anonymous access is permitted in public mode for the currently registered read-only tools. Outside public mode, each Streamable HTTP session is bound to the user (or client) that created it, so another account presenting its `Mcp-Session-Id` gets `403`.
- **Request Body Cap:** HTTP POST bodies are bounded at `MaxMCPRequestBytes` (currently `MaxMutationRequestBytes`, 256 KiB). Oversized requests receive the standard `413 Request Entity Too Large` response before the upstream SDK reads the body. Early rejection via `Content-Length` protects against large known-size payloads; other bodies are read through `http.MaxBytesReader` by the handler itself, which answers an overflow with the same audited `413`, refuses a body that is not a single JSON value with `400`, and passes the body on to the SDK only once it is read. JSON-RPC batches, which the SDK still accepts for protocol `2025-03-26`, are limited to `MaxMCPBatchMessages` (10) messages.
- **Session Timeout:** Streamable HTTP sessions idle out after `McpSessionTimeout` (5 minutes by default). Active clients renew the idle timer on every JSON-RPC request, so interactive IDE use is unaffected; sessions abandoned without the `DELETE` tear-down free up promptly instead of lingering.
- Read-only resources:
  - `photoprism://config-options`
  - `photoprism://search-filters`
  - `photoprism://cli-commands`
  - `photoprism://file-formats`
  - `photoprism://metadata-fields`
  - `photoprism://thumbnail-sizes`
- Read-only tools, each annotated with `readOnlyHint: true`:
  - `list_config_keys`
  - `find_search_filters`
  - `list_cli_commands`
  - `list_file_formats`
  - `list_metadata_fields`
  - `list_thumbnail_sizes`
- Library tools, registered on the library server only (authenticated, non-public sessions) and authorized per call:
  - `search_media`
  - `get_media`
  - `list_collections`
  - `get_collection`
  - `list_people`
  - `list_labels`

### Package layout

| Package                      | Purpose                                                                                                                                                                          |
|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/mcp/`              | Core MCP logic: reference and library servers, per-tool authorization (`authorize.go`), data pipeline, resources, tools                                                          |
| `internal/api/mcp.go`        | Gin HTTP handler with auth middleware, account check, scope step-up, route registration                                                                                          |
| `internal/api/mcp_auth.go`   | `401` challenge naming the protected resource metadata, `TokenInfo` binding of MCP sessions to their owner                                                                       |
| `internal/api/mcp_grants.go` | Closes the MCP sessions of an app when its OAuth grant is revoked or its client is disabled                                                                                                                |
| `internal/commands/mcp.go`   | CLI command (`photoprism mcp serve`) using stdio transport                                                                                                                       |
| `internal/auth/acl/`         | `ResourceMCP` constant and ACL grant rules (`GrantFullAccess` for admin; `GrantSearchAll` for the API client roles: client, instance, service, portal; editions add their roles) |

### Scope

In scope for the current server:

- Reuse the existing internal reference data behind the `photoprism show` reports (see *Internal data sources*) instead of maintaining a parallel dataset.
- Keep outputs compact enough for LLM consumption.
- Authenticated remote access via the Streamable HTTP transport, plus a stdio transport for local development and testing.

Out of scope for the reference server (public mode and stdio), which must not regress without additional per-tool gates:

- Write-capable tools.
- Direct database access.
- Live PhotoPrism instance or API queries.
- Per-user state (albums, photos, sessions, settings).

The library server reads per-user library data with the session's permissions, through the search layer and the relation queries of `media.go`, `covers.go`, and `tool_people_labels.go`, which apply the same visibility rules; write-capable tools stay out of scope.

### Internal data sources

- Config options: `internal/config.Flags` plus `internal/config.OptionsReportSections`
- Search filters: `internal/form.Report(&form.SearchPhotos{})`
- CLI commands: the command list of the running edition, passed to `mcp.NewServer` in `mcp.Options` by `ServeMCP` (from the CLI app of the config) and by `photoprism mcp serve`, and converted with `internal/commands/catalog` as `photoprism show commands`
- File formats: `pkg/media.Report(fs.ExtensionList.Types(true), true, true, true)`, as `photoprism show file-formats`
- Metadata fields: `internal/meta.Report(&meta.Data{})` sorted by type and field, plus `meta.Docs`, as `photoprism show metadata`
- Thumbnail sizes: `internal/thumb.Report(thumb.Sizes.All(), false)`, as `photoprism show thumb-sizes`, with the sizes in `thumb.VideoSizes` (`photoprism show video-sizes`) marked as video sizes

The data is built once when the server is created. It comes from compile-time tables only, never from package variables that are set from the runtime config, such as the thumbnail size limits. The command list contains no runtime values either: hidden commands and flags are skipped, and a flag `default` is returned only for the standard scalar, path, and slice flag types, and omitted if the flag can also be set through an environment variable or a file, or its command path is not unique.

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

When the running config has `DisableMCP=true` (via `options.yml`, `PHOTOPRISM_DISABLE_MCP`, or `--disable-mcp`), the command refuses to start with `mcp serve disabled by config; pass --disable-mcp=false to override`. Pass the override on the command line for ad-hoc development runs without flipping the operator-facing setting:

```bash
./photoprism --disable-mcp=false mcp serve
```

### Run via HTTP

Start PhotoPrism:

```bash
./photoprism start
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
- Call `list_cli_commands` with `{"query":"users add"}`
- Call `list_file_formats` with `{"type":"raw"}`
- Call `list_metadata_fields` with `{"query":"exif:DateTimeOriginal"}`
- Call `list_thumbnail_sizes` with `{"type":"video"}`

### Available resources

`photoprism://config-options`

- JSON payload with `edition` and `items`
- Each item includes `section`, `environment`, `cli_flag`, `default`, and `description`

`photoprism://search-filters`

- JSON payload with `edition` and `items`
- Each item includes `filter`, `type`, `examples`, and `notes`

`photoprism://cli-commands`

- JSON payload with `edition` and `items`, in the order of `photoprism show commands`
- Each item includes `command` (the full path, e.g. `photoprism users add`), `usage`, `description`, `args_usage`, `aliases`, and `flags`
- Each flag includes `name`, `aliases`, `type`, `usage`, `required`, `env`, and `default` if the flag definition has a literal default
- Hidden commands, hidden flags, and the `--help` flag are not included; global options are covered by `photoprism://config-options`

`photoprism://file-formats`

- JSON payload with `edition` and `items`
- Each item includes `format`, `description`, `type` (lowercase, e.g. `image`, `raw`, `video`), and `extensions`

`photoprism://metadata-fields`

- JSON payload with `edition`, `items`, and `docs`
- Each item includes `field`, `type`, and the `exiftool`, `xmp`, and `dcmi` tags it is read from
- Each `docs` entry links a metadata `namespace` to its documentation `url`

`photoprism://thumbnail-sizes`

- JSON payload with `edition` and `items`
- Each item includes `name`, `width`, `height`, `aspect_ratio`, `available`, `usage`, and `video`, which marks the standard video sizes

### Available tools

`list_config_keys`

- Inputs: `section`, `query`, `edition`, `limit`
- Returns matching config rows with environment variables, CLI flags, defaults, descriptions, and a conservative `edition_support` hint
- Validation rejects unsupported `edition` values

`find_search_filters`

- Inputs: `query`, `type`, `limit`
- Returns matching search filters with examples and notes
- Validation rejects unsupported filter `type` values

The following tools return `items`, `total`, `limit`, `offset`, `truncated`, and an optional `warnings` array. `limit` defaults to 20 and is capped at 50, `offset` must not be negative, and free-text inputs are capped at 200 characters. An offset past the end returns no items with the correct `total`.

`list_cli_commands`

- Inputs: `query` (command paths, aliases, usage, descriptions, and flag names), `limit`, `offset`
- Returns the items of `photoprism://cli-commands`
- Returns no items and a warning if the server was created without a command list

`list_file_formats`

- Inputs: `query` (format names, extensions, and descriptions), `type` (a format type such as `image`, `raw`, or `video`, or `all`), `limit`, `offset`
- Returns the items of `photoprism://file-formats`
- Validation rejects unsupported `type` values and names the accepted ones

`list_metadata_fields`

- Inputs: `query` (field and tag names), `limit`, `offset`
- Returns the items of `photoprism://metadata-fields`

`list_thumbnail_sizes`

- Inputs: `type` (`image`, `video`, or `all`; default `all`)
- Returns `items` and `total`; `video` returns only the standard video sizes, while `image` and `all` return every thumbnail size
- Validation rejects unsupported `type` values

The library tools read the library with the permissions of the session, through the same search functions and feature settings as the matching REST endpoints. They are not available in public mode or with `photoprism mcp serve`, and a token needs the scope of the tool's resource besides `mcp`.

`search_media` (declaration: `photos`, `search`)

- Inputs: `query` (search filters, see `find_search_filters`), `type` (`image`, `raw`, `live`, `video`, `animated`, `audio`, `vector`, `document`, combined with `|`), `collection` (album UID), `person` (person UID), `taken_after`, `taken_before` (`YYYY-MM-DD`; on or after, and before the date), `order` (`newest` default, `oldest`, `added`, `edited`, `relevance`), `limit` (default 20, max 50), `offset`, `detail` (`brief` default, `full`)
- Returns `items`, `limit`, `offset`, `next_offset`, `truncated`, and optional `warnings`, without `total`; each item is a media with `uid`, `type`, `title`, `taken_at`, `taken_at_local`, `place` (`name`, `country`), and `favorite`, and with `detail=full` also `caption`, `camera`, `lens`, `labels`, `people`, and `collections` (each collection type only with the token scope of its resource)
- Runs `search.UserPhotos` like `GET /api/v1/photos?merged=true`: the web app's file selection, so file filters such as `codec:avc1` find the same media, with the feature settings of `search.ApplyPhotoFeatures`, as the photos endpoint applies them; structured inputs take precedence over the same filters in `query`
- `offset` is a row position of the search result, not a media count: pass back `next_offset`, which is omitted on the last page. Each page reads one row before `offset` and a window of `3 * limit` rows (at most 150), skips the rows of a media that began before `offset`, and reads on in windows that double in size (at most 5,000 rows) until it has `limit` media and has seen where the next one begins (`pageMedia`); it stops when the request is canceled. The offered orders keep the rows of a media adjacent, which `TestSearchMediaOrders` checks
- `detail=full` loads labels, people, and collections with one query per kind for the page and keeps those the session may see: labels without private ones if `search.LabelsPublicOnly` applies, people as `search.UserSubjects` returns them, and collections as `search.UserAlbums` returns them per collection type. Labels and people require full access to pictures and collections view access to albums, including the `albums` scope of a token, as in the photo details

`get_media` (declaration: `photos`, `view`)

- Inputs: `uid` (media UID from `search_media`)
- Returns the `detail=full` fields of `search_media` plus `width` and `height` (primary file), `duration` (seconds), `location` (`lat`, `lng`, `altitude`) if the session may view places, `keywords` of the picture details if it has full access to pictures, `files` (`name` without path, `type`, `size`, `primary`), `quality`, `private`, `archived`, and `review` (quality 0 to 2)
- Uses the checks of `GET /api/v1/photos/:uid` (`search.PhotoVisibleToSession`, `query.PhotoPreloadByUID`, `Photo.RedactForSession`) and the relation filters of `search_media`; a UID the session cannot see gets the same not-found error as one that does not exist

`list_collections` (declaration: `albums`, `search`) and `get_collection` (`albums`, `view`)

- `list_collections` inputs: `type` (`album` default, `folder`, `moment`, `month`, `state`, or `all`), `query` (text in titles and locations, not search filters), `limit`, `offset`; returns `items`, `total`, `limit`, `offset`, `truncated`, and optional `warnings`
- `get_collection` input: `uid` (collection UID from `list_collections`)
- Each collection has `uid`, `type`, `title`, `description`, `media_count`, `cover` (media UID, if the session may see the picture), `year`, `month`, and `day` (months and moments), `created_at`, and `updated_at`
- Runs `search.UserAlbums` per type, like `GET /api/v1/albums?type=<type>`, so each type is checked by its own resource (`search.AlbumTypeResource`), and a token also needs the scope of that resource: `albums`, `folders`, `moments`, `calendar` (months), or `places` (states). With `all`, types the role or token may not see are left out, with a warning naming a missing scope. `get_collection` checks a collection by its own type and reports one the session cannot see as not found

`list_people` (declaration: `people`, `search`) and `list_labels` (`labels`, `search`)

- Inputs: `query` (text in names and aliases of people, or in label names), `limit`, `offset`; `list_labels` also `all` (include labels with low priority or a single media)
- Return `items`, `total`, `limit`, `offset`, `truncated`, and optional `warnings`; people have `uid`, `name`, `media_count`, and `cover` (media UID, if the session may see the picture), labels `uid`, `name`, `slug` (for the `label:` filter), `media_count`, and `categories` (names, without private ones for sessions that see public labels only)
- Run `search.UserSubjects` with type `person` and at least one file, like the people page (`GET /api/v1/subjects?type=person&files=1`), and `search.UserLabels` after `entity.UpdateLabelCountsIfNeeded`, like `GET /api/v1/labels`, so private and hidden people and private labels follow the same rules

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
  - `RoleManager` → `GrantSearchAll` in Pro and Portal builds (the role does not exist in CE/Plus); `RoleClusterAdmin` → `GrantFullAccess` in Portal.
  - `RoleUser`, `RoleViewer` → `GrantSearchAll` in Plus, Pro, and Portal (CE has neither role; Plus builds without membership keep `user`).
  - `RoleClient`, `RoleInstance`, `RoleService`, `RolePortal` → `GrantSearchAll` in every edition.
  - All other roles (`guest`, `visitor`, `contributor`, default) are denied.
- **Account setting:** `auth_users.mcp` (`User.MCP`: `-1` off, `0` default, `1` on) is read only through `User.CanUseMCP()`, which also requires an active, registered account and the `ResourceMCP` grant above. After the ACL check, the handler refuses every session that acts as a user (login session, app password, app access token, API client bound to a user) whose account fails `CanUseMCP()` with `403`, on every request, so after a change through the Admin UI or REST, open MCP sessions stop working with their next request; a change through the CLI takes effect as the server's session cache expires. Clients without a user, cluster JWTs, and the public session are not affected. The consent page offers no scope to such an account, and changing the setting revokes the account's derived sessions and app grants like other privilege changes made by an admin who manages all accounts. Admins set it in the user dialogs of Plus, Pro, and Portal or with `photoprism users add|mod --mcp[=false]`.
- **Why `GrantSearchAll` for non-admins?** It includes `AccessAll`, `ActionView`, and `ActionSearch` — exactly what the read-only tools need — but excludes `ActionManage`/`ActionUpdate`/`ActionDelete`/`ActionCreate`. The handler checks `ActionView` on `ResourceMCP` only, so the grant does not limit what an individual tool can do: each tool on the library server declares the resource and action of its data, which the per-tool check below enforces, and the tool applies the same visibility rules as the matching REST endpoint. A client token then needs both the `mcp` scope and the scope of that resource.
- **Per-tool check:** `AuthorizeTool` (`authorize.go`) decides each call from the tool's `ToolAccess` declaration, the session's role (`Session.Grants`, client and user role), the account setting, and the token scope (`mcp` plus the declared scope, each with the declared action). The `authorizeTools` middleware of the library server refuses calls it does not allow with an error result naming the tool, audited, and lists only the tools the role and account may use; a tool the token lacks the scope for stays listed, and `ServeMCP` answers its `tools/call`, also inside a JSON-RPC batch, with `403` and `WWW-Authenticate: Bearer error="insufficient_scope", scope="..."` (with `resource_metadata` while the OAuth app flow is enabled), naming the app scopes the token has and the missing ones, so OAuth clients re-authorize for it without losing the others. The middleware decides; the step-up only asks in advance. The session reaches the middleware and the tools in `TokenInfo.Extra` (`SessionKey`), in process only.
- **Client tokens:** API client sessions must also include the `mcp` resource (or a wildcard) in their session scope; the ACL grant alone is not sufficient.
- **Auth model:** request-level. The handler runs `AuthAudience(c, conf.OAuthResourceUrl(), acl.ResourceMCP, acl.ActionView)`, i.e. `Auth()` that also accepts OAuth app access tokens issued for this endpoint, followed by `s.Abort(c)`, which writes the matching status code (`401` unauthenticated, `403` ACL deny, `429` rate-limited) and returns `true` so the handler can `return` early.
- **Public mode:** anonymous access is permitted. In public mode, `api.Session()` returns the default public session (effectively admin), so `Auth(...)` passes and the reference server's tools are reachable without a token. The handler's factory returns the reference server (`NewServer`) in public mode and the library server (`NewLibraryServer`) otherwise. This is an intentional, narrow allowance for demo deployments (`demo.photoprism.app`); it is safe only because every tool of the reference server returns static reference metadata derived from the sources listed under *Internal data sources* — no database access, no per-user state, no secrets, no mutations. **Any tool that touches per-user state, the database, or mutates anything MUST be registered on the library server only, with a declaration**. See *Extending the Tool Surface* below.
- **Stdio transport:** no authentication for the currently registered read-only tools — the trust boundary is the OS process pipe between client and server, and the spawning shell already has full PhotoPrism access on the host. The same static-data contract that makes public-mode HTTP safe makes stdio safe. **Any future write-capable, destructive, or per-user tool MUST require a session token on stdio too** (for example via `PHOTOPRISM_MCP_TOKEN` or a `--session` flag, validated through the same `Auth(...)` path the HTTP handler uses), so a misbehaving or prompt-injected MCP client cannot destroy data the user did not explicitly authorize. See *Extending the Tool Surface* below.

### Rate Limiting

The MCP handler does not install a custom rate limiter — there is no per-endpoint bucket. Coverage depends on the edition:

| Build  | Generic per-IP HTTP limiter? | Notes                                                                                                                                                                                                                                                        |
|--------|------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| CE     | no                           | Only the authentication check protects the endpoint. In public mode all callers are anonymous, so CE deployments that run with `--public` have no per-request throttle in front of MCP.                                                                      |
| Plus   | no                           | Same as CE. The IPS middleware exists but only consumes tokens on known scanner/exploit paths, so it does not throttle MCP traffic.                                                                                                                          |
| Pro    | yes                          | `pro/internal/server/register.go` calls `router.Use(limiter.Middleware(limiter.NewLimit(rate.Every(secOpt.RequestInterval), secOpt.RequestLimit)))` when both options are set. The limiter is per client IP and applies to every API endpoint, MCP included. |
| Portal | yes                          | Same wiring as Pro in `portal/internal/server/register.go`.                                                                                                                                                                                                  |

A per-endpoint limiter (via `limiter.Auth` / `limiter.Login` / `limiter.AbortJSON`) is only worth adding when MCP grows write-capable tools or endpoints that warrant stricter throttling than the generic IP limiter — for example, anything that mutates state or that triggers expensive backend work.

CE and Plus deployments that expose the endpoint to untrusted networks should enforce per-IP request limits at the reverse proxy. The application-level `McpSessionTimeout` (5 minutes) and `MaxMCPRequestBytes` (256 KiB) bound the per-session memory footprint and per-request allocation, but they do not throttle request frequency; a proxy rule (for example nginx `limit_req_zone` or Traefik `rateLimit` middleware) is the recommended companion control in builds without a generic HTTP limiter.

### Scope Plumbing

`mcp` is the canonical scope token for `ResourceMCP`. The relevant pieces:

- **Sanitization:** `pkg/clean.Scope` lowercases the input and parses it through `pkg/list.ParseAttr`. There is no allowlist of valid scope tokens, so `--scope mcp`, `--scope "mcp metrics"`, and `--scope "*"` are all accepted by `clients add` and `auth add` without any registry update.
- **Authorization:** `internal/auth/acl.ScopePermits` checks the parsed attribute list against the resource string. Because `acl.ResourceMCP.String() == "mcp"`, the existing `attr.Contains(...)` path matches a session that holds the `mcp` token. See `TestScopePermits/MCPScope` in `internal/auth/acl/scope_test.go` for the canonical assertions (admin token, mixed scopes, case-insensitivity, deny on unrelated scopes).
- **Cluster JWTs:** instance-side validation runs through `Config.JWTAllowedScopes()` (`internal/api/api_auth_jwt.go`). The default allowlist is `DefaultJWTAllowedScopes = "config cluster vision metrics mcp users"` in `internal/config/config_cluster.go`, so portal-issued JWTs with `scope=mcp` are accepted out of the box. Operators that override the list via `--jwt-scope` / `PHOTOPRISM_JWT_SCOPE` need to include `mcp` themselves.

### Extending the Tool Surface

Anonymous access in public mode is only safe as long as every registered tool returns static reference metadata. Before adding a new tool, confirm it fits the existing contract:

- No database reads or writes.
- No per-user state (albums, photos, sessions, settings).
- No filesystem, network, or subprocess side effects.
- No access to secrets or runtime config values (only flag schema/defaults are allowed).

If a proposed tool does **not** fit that contract, register it on the library server only (`registerLibraryTools`, called by `NewLibraryServer`), never on the server `NewServer` returns:

1. **Declare it.** Register it through `AddTool` with a `ToolAccess` that names the ACL resource and action the call needs (and `Scope` if the token scope differs from the resource name). `sdkmcp.AddTool` is not called anywhere else; `TestLibraryServerDeclarations` fails for an undeclared tool or an action other than view, search, or download.
2. **Let the middleware decide.** `authorizeTools` checks each `tools/call` with `AuthorizeTool` and hides the tools the caller's role or account cannot use from `tools/list`; `ServeMCP` answers a call whose token lacks the scope with `403 insufficient_scope` before the SDK runs it. Both read the session from `RequestSession(req)`, which the HTTP handler puts in the token info of every request.
3. **Keep resources and prompts static.** Only `tools/call` is authorized per item; resources, templates, and prompts pass the middleware, so the library server offers only the static resources of `NewServer`.
4. **Apply the data's own rules in the tool.** Read the caller with `RequestSession(req)` and call the same search function as the REST endpoint with that session, so the tool returns what the endpoint returns for the same session.

When the new tool needs an authenticated stdio session (any write-capable, destructive, or per-user tool), apply the same gate to `mcp serve`:

5. **Spawn-time session token.** Accept the token via `PHOTOPRISM_MCP_TOKEN` or a `--session` CLI flag in `internal/commands/mcp.go`, resolve it through the same `Auth(...)` path the HTTP handler uses (`acl.ResourceMCP` + the appropriate `Action*`), and pass the resolved session into the factory the same way the HTTP handler does. No token → static-data server (current behavior, unchanged); valid token → library server with the session in the token info. Only the gate keying differs from the HTTP handler.

Cross-cutting MCP primitives that protect *every* transport (and SHOULD be applied whenever they fit):

- **Roots.** Server-enforce the client's `initialize`-time `roots` for any tool that walks the catalog (albums, folders, files). Never trust the client to keep itself in scope.
- **`destructiveHint` annotation.** Mark high-blast-radius tools (`delete_album`, `wipe_index`, `reset_users`) with `annotations.destructiveHint: true` so MCP-aware UIs warn the human before approving the call.
- **`elicitation/create` for confirmation.** For destructive calls, ask the host to elicit human consent before mutating — the MCP-native equivalent of the `--yes` flag the destructive CLI commands already require.

Whichever path you pick, **add a test in `internal/mcp/server_test.go` that fails if the restricted tool shows up in `tools/list` or is callable over the public path or an unauthenticated stdio session**, and update *Available tools* above.

### How Users Get Access

Every account whose role may use `mcp` (admins in every edition; users, viewers, and managers in the editions that have them) can connect an MCP client, unless an admin turned MCP off for the account (`photoprism users mod --mcp=false <username>`, or the **MCP** switch in the user dialogs) or for the instance (`--disable-mcp`).

API clients get access through admin-issued client tokens, for example for a CI job or an IDE:

```bash
./photoprism clients add \
  --name "Alice's IDE" \
  --scope "mcp read photos albums people labels" \
  --role client \
  --expires 2592000          # 30 days; use -1 for no expiry
```

The command prints a client ID and secret. Combine them into a bearer token (or pass them through the OAuth2 client-credentials flow) and paste the resulting value into the MCP client config — the same JSON snippets shown above under *MCP Client Compatibility* apply unchanged. Replace `<admin-token>` with the issued token. A client gets the library tools whose resource scopes it holds; with `mcp` alone, only the reference tools. Resource scopes apply to the whole API, not only to MCP, and `read` limits the token to read operations; add `folders`, `moments`, or `calendar` for those collection types, and `places` only if the client should receive coordinates. A client added for a user account (`photoprism clients add [flags] <username>`) also follows that account's MCP setting. In PhotoPrism Pro, admins can also add, edit, disable and delete clients in the **Clients** tab of the admin settings, where managers can view them.

To revoke access without disabling the user account, the administrator runs:

```bash
./photoprism clients rm <client-id>
```

An app password needs `mcp` plus the scope of each tool's resource, e.g. `photoprism auth add --scope "mcp read photos albums people labels" <username>`, with the same rules as for clients; with `mcp` alone, it reaches the reference tools only.

Accounts that may use MCP can also connect an MCP client such as `claude.ai` by signing in to PhotoPrism and approving it (OAuth authorization code flow with PKCE). The app is registered with `photoprism clients add --redirect-uri <uri>` (in Pro also in the **Clients** tab, which lists the connected accounts of each app) or, where `--oauth-apps` allows it, identified by its client metadata URL; the connection is listed under Settings > Account > Connected Apps, where it can be disconnected.
