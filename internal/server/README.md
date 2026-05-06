## PhotoPrism — HTTP Server

**Last Updated:** May 3, 2026

### Overview

`internal/server` wires Gin, middleware, and configuration into the PhotoPrism HTTP/HTTPS/WebDAV servers. It owns startup/shutdown orchestration, route registration, and helpers for recovery/logging. Subpackages (`process`, `limits`, etc.) are kept lightweight so CLI commands and workers can embed the same server behavior without duplicating boilerplate.

#### Constraints

- Uses the configured `config.Config` to decide TLS, AutoTLS, Unix sockets, proxies, compression, and trusted headers.
- Middleware must stay small and deterministic because it runs on every request; heavy logic belongs in handlers.
- Panics are recovered by `Recovery()` which logs stack traces and returns 500.
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
  - Built-in security middleware skips browser-document headers (`Content-Security-Policy`, `X-Frame-Options`) on `/originals` and `/import` paths.
  - PROPFIND `207 Multi-Status` responses normalize XML media type to `application/xml; charset=utf-8`.
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
