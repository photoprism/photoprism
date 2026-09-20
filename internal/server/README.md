## PhotoPrism — HTTP Server

**Last Updated:** September 17, 2026

### Overview

`internal/server` wires Gin, middleware, and configuration into the PhotoPrism HTTP/HTTPS/WebDAV servers. It owns startup/shutdown orchestration, route registration, and helpers for recovery/logging. Subpackages (`process`, `limits`, etc.) are kept lightweight so CLI commands and workers can embed the same server behavior without duplicating boilerplate.

#### Constraints

- Uses the configured `config.Config` to decide TLS, AutoTLS, Unix sockets, proxies, compression, and trusted headers.
- Middleware must stay small and deterministic because it runs on every request; heavy logic belongs in handlers.
- Panics are recovered by `Recovery()`, which returns 500 and writes a debug-level entry holding the stack trace, the caller, the request method, its route template and an allowlist of headers.
- Startup supports mutually exclusive endpoints: Unix socket, HTTPS with certs, AutoTLS (with redirect listener), or plain HTTP.

#### Goals

- Provide a single entrypoint (`Start`) that configures listeners, middleware, and routes consistently.
- Keep health/readiness endpoints lightweight and cache-safe.
- Ensure redirect and TLS listeners include sensible header and idle limits.

#### Non-Goals

- Managing Docker/Traefik lifecycle (handled by compose files).
- Serving static files directly; templates are loaded via Gin and routed by `routes_webapp.go`.

### Package Layout (Code Map)

- `start.go` — main startup flow, listener selection (HTTP/HTTPS/AutoTLS/Unix socket), graceful shutdown.
- `routes_webapp.go` — Web UI routes and shared method helpers (`MethodsGetHead`).
- `static_precompressed.go` — `PrecompressedStatic` handler that serves bundled `/static/*` assets from precompressed siblings emitted by `frontend/scripts/precompress.js`; the same handler accepts operator-supplied siblings for `/c/static/*` and falls back to identity when none exist. Range requests always serve identity, and `http.ServeContent` continues to handle `Last-Modified` + `If-Modified-Since` revalidation for both encoded and identity responses (the handler does not set an `ETag`, so `If-None-Match` is latent rather than active).
- `recovery.go` — panic recovery middleware with stack trace logging.
- `logger.go` — request logging middleware (enabled in debug mode).
- `security.go` — security headers and trusted proxy/platform handling.
- `webdav_*.go` & tests — WebDAV handlers and regression tests for overwrite, traversal, and metadata flags.
- `webdav_path.go` — shared helper to classify built-in and path-proxied WebDAV routes.
- `webdav_destination.go` — shared collection and account-path validation for edition COPY/MOVE handlers.
- `process/` — light wrappers for server process metadata.

### Related Packages

- `internal/api` — registers REST endpoints consumed by `registerRoutes`.
- `internal/config` — supplies HTTP/TLS/socket settings, compression, proxies, and base URI paths.
- `internal/server/process` — exposes process ID for logging.
- `pkg/http/header` — shared HTTP header constants used by health endpoints.

### Configuration & Safety Notes

- Compression: configured via `PHOTOPRISM_HTTP_COMPRESSION` / `--http-compression` as a comma-separated preference list. Supported tokens are `zstd`, `gzip`, and `none` (empty value also disables compression). The default ships as `zstd,gzip` so capable clients receive zstd while everyone else falls back to gzip; unknown tokens are ignored with a startup warning.
- Bundled frontend assets under `/static/*` are served with precompressed `.zst` / `.gz` siblings produced at build time by `frontend/scripts/precompress.js` (the npm `postbuild` hook for `make build-js`), selected via `PrecompressedStatic` in `static_precompressed.go`. Custom static assets under `/c/static/*` go through the same handler so extensions and operators *may* ship precompressed siblings alongside their files; without siblings the route serves identity. The runtime middleware bypasses both routes so it never re-encodes an already-encoded body and so `PHOTOPRISM_HTTP_COMPRESSION=none` consistently disables every encoded code path on these routes.
- Trusted proxies/platform headers are read from config; keep the list tight.
- If no trusted proxy ranges are configured (or the configured ranges are invalid), proxy trust is disabled and client IP resolution falls back to the TCP peer address.
- HTTP hardening defaults:
  - `ReadHeaderTimeout` is configured via `PHOTOPRISM_HTTP_HEADER_TIMEOUT` / `--http-header-timeout` (default `15s`).
  - `MaxHeaderBytes` is configured via `PHOTOPRISM_HTTP_HEADER_BYTES` / `--http-header-bytes` (default `1 MiB`).
  - `IdleTimeout` is configured via `PHOTOPRISM_HTTP_IDLE_TIMEOUT` / `--http-idle-timeout` (default `180s`).
  - Global `ReadTimeout` / `WriteTimeout` remain disabled to avoid breaking large transfers.
- WebDAV response behavior:
  - Successful Basic and token authentication uses the entity-owned one-minute user cache. Account saves invalidate that user's WebDAV and general session entries together; other users stay cached. Path edits retain the normal privilege-change and session-revocation policy. Authentication snapshots the cache generation before lookup and skips stale cache writes after an intervening invalidation, without interrupting in-flight requests.
  - Edition COPY/MOVE handlers validate destinations with `WebDAVDestinationStatus` before serving the operation. Collection prefixes and configured account paths apply to URL-decoded, canonical destinations; source-path and role checks remain in the edition handlers.
  - Built-in security middleware skips browser-document headers (`Content-Security-Policy`, `X-Frame-Options`) on `/originals` and `/import` paths.
  - PROPFIND `207 Multi-Status` responses normalize XML media type to `application/xml; charset=utf-8`.
  - Request errors go to the console-only system log (`event.System*`), not the browser log stream, since `x/net/webdav` embeds absolute server paths in its messages. A `MKCOL` on an existing collection is a benign sync-client probe: it returns 405 and is logged at debug rather than as an error.
  - `LOCK` lifetimes are capped at `mutex.WebDAVMaxLockLifetime` (default one hour) so a client cannot mint locks that never expire: `WebDAVClampLockTimeout` clamps the requested `Timeout` header and the `webdavLockSystem` wrapper enforces the same bound on the stored lock.
  - A request path component containing a backslash is refused before anything is created, renamed, or copied: `400` for `PUT`, `MKCOL`, and `MOVE`/`COPY` destinations (`WebDAVSeparatorInName`). `LOCK` is declined in `webDAVFileSystem.OpenFile`, which is where it would create its placeholder, and upstream reports that as `500`. Names that already contain one stay readable and removable.
  - A `PUT` declaring a length above `conf.OriginalsLimitBytes()` returns `413` before the destination is opened. A body of unknown length stays legal and is bounded as it is read, carrying the upstream handler's own status; any partial file it leaves is removed by `WebDAVRemovePartialUpload`.
  - `LOCK`, `PROPFIND` and `PROPPATCH` bodies are bounded at `api.MaxWebDAVMetadataRequestBytes` (128 KiB), since those are the methods the handler parses into memory. A declared length above the bound returns `413`; any other body is bounded as it is read.
- AutoTLS: uses `autocert` and spins up a redirect listener; ensure ports 80/443 are reachable.
- Unix sockets: optional `force` query removes stale sockets; permissions can be set via `mode` query.
- Health endpoints (`/livez`, `/health`, `/healthz`, `/readyz`) return `Cache-Control: no-store` and `Access-Control-Allow-Origin: *`.

### Testing

- Lint & unit tests: `golangci-lint run ./internal/server...` and `go test ./internal/server/...`
- WebDAV behaviors are covered by `webdav_*_test.go`; they rely on temp directories and in-memory routers, including PROPFIND `207` XML/header assertions and path classification checks.

### Operational Tips

- Prefer `Start` with context cancellation so graceful shutdown is triggered (`server.Close()`).
- When adding routes, register them in `registerRoutes` and reuse `MethodsGetHead` for safe verbs.
- Keep middleware light; log or enforce security at the edge (Traefik) when possible, but maintain server-side defaults for defense in depth.

### WebDAV Credential Actions

Token and app-password requests resolve the session even when the authenticated account
is cached. Every request applies the credential scope and both account/client WebDAV grants.
GET, HEAD, and POST require download; PROPFIND requires view, or upload for an explicit
Depth 0 probe of the target within the upload path (including its root). If neither an
account base nor upload path is configured, only the mount root may be probed. Broader or omitted
depths still require view; target-only probes expose properties, not file contents or children. PUT, MKCOL, and MOVE require
upload; DELETE requires delete; PROPPATCH, LOCK, and UNLOCK require update. COPY requires
both download and update. OPTIONS requires WebDAV admission without a read/write action.
Core and edition handlers share `WebDAVMethodPermissions`; request-level probe admission
is handled by `WebDAVRequestPermits`. Existing account eligibility, edition filesystem
permissions, and path checks remain mandatory.
Ordinary Basic passwords retain their account authentication and cache behavior.

### WebDAV Path Policy

Every mount uses the reserved administrative names defined by `pkg/fs.ReservedPathNames`,
and `pkg/fs.ReservedPathPatterns`, including `.env.*`, `.*ignore`, `.*_history`, `.bash_history-*.tmp`, and `.*.cnf` variants,
plus the component suffixes in `pkg/fs.ReservedPathSuffixes`.
Ignore-file names matching `.*ignore` remain visible under `ReservedPathPolicy{AllowIgnoreNames: true}`
and use the managed-file write policy below. Independent reserved-name/pattern and
logical ancestor restrictions still apply. Matching is case-insensitive and applies at every path component.
A backslash in any component is refused outright rather than cleaned, so a name is never rewritten
into one that differs from the name the handler opens.
Entries with reserved logical components are hidden from reads and directory listings,
and cannot be written. Ordinary dot files and eligible linked originals remain supported;
the configured mount root's own ancestry is not part of its served namespace.

Directory MOVE/DELETE and overwrite destinations are inspected before mutation, with a
100,000-entry and 128-level preflight limit. Protected entries, inspection errors, cancellation,
or limit exhaustion refuse the operation. COPY uses the visible source view and does not
copy reserved children. Filesystem operations enforce the same policy after HTTP preflight.

YAML writes, `.*ignore` mutations, and the `X-Favorite` upload header require effective
photo `FullAccess` authority from both client and account. They retain the existing WebDAV action scope. Without full photo authority, an otherwise
permitted upload succeeds but its favorite flag does not create a YAML sidecar. This covers direct
writes, renames, and affected logical descendants. YAML/ignore-file reads and ordinary JSON transfers
retain their prior admission. An explicit COPY Depth 0 can create an empty collection
without copying its managed-file descendants.

Name checks do not resolve operator-managed symlinks. Existing filesystem mappings are trusted;
WebDAV provides no operation for creating filesystem symlinks. Normal filesystem errors and
permissions still apply when following an operator-configured path.
