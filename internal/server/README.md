## PhotoPrism — HTTP Server

**Last Updated:** November 22, 2025

### Overview

`internal/server` wires Gin, middleware, and configuration into the PhotoPrism HTTP/HTTPS/WebDAV servers. It owns startup/shutdown orchestration, route registration, and helpers for recovery/logging. Subpackages (`process`, `limits`, etc.) are kept lightweight so CLI commands and workers can embed the same server behavior without duplicating boilerplate.

#### Context & Constraints

- Uses the configured `config.Config` to decide TLS, AutoTLS, Unix sockets, proxies, compression, and trusted headers.
- Middleware must stay small and deterministic because it runs on every request; heavy logic belongs in handlers.
- Panics are recovered by `Recovery()` which logs stack traces and returns 500.
- Startup supports mutually exclusive endpoints: Unix socket, HTTPS with certs, AutoTLS (with redirect listener), or plain HTTP.

#### Goals

- Provide a single entrypoint (`Start`) that configures listeners, middleware, and routes consistently.
- Keep health/readiness endpoints lightweight and cache-safe.
- Ensure redirect and TLS listeners include sensible timeouts.

#### Non-Goals

- Managing Docker/Traefik lifecycle (handled by compose files).
- Serving static files directly; templates are loaded via Gin and routed by `routes_webapp.go`.

### Package Layout (Code Map)

- `start.go` — main startup flow, listener selection (HTTP/HTTPS/AutoTLS/Unix socket), graceful shutdown.
- `routes_webapp.go` — Web UI routes and shared method helpers (`MethodsGetHead`).
- `recovery.go` — panic recovery middleware with stack trace logging.
- `logger.go` — request logging middleware (enabled in debug mode).
- `security.go` — security headers and trusted proxy/platform handling.
- `webdav_*.go` & tests — WebDAV handlers and regression tests for overwrite, traversal, and metadata flags.
- `process/` — light wrappers for server process metadata.

### Related Packages

- `internal/api` — registers REST endpoints consumed by `registerRoutes`.
- `internal/config` — supplies HTTP/TLS/socket settings, compression, proxies, and base URI paths.
- `internal/server/process` — exposes process ID for logging.
- `pkg/http/header` — shared HTTP header constants used by health endpoints.

### Configuration & Safety Notes

- Compression: only gzip is enabled; brotli requests log a notice.
- Trusted proxies/platform headers are read from config; misconfiguration may expose client IP spoofing—keep the list tight.
- AutoTLS: uses `autocert` and spins up a redirect listener with explicit read/write timeouts; ensure ports 80/443 are reachable.
- Unix sockets: optional `force` query removes stale sockets; permissions can be set via `mode` query.
- Health endpoints (`/livez`, `/health`, `/healthz`, `/readyz`) return `Cache-Control: no-store` and `Access-Control-Allow-Origin: *`.

### Testing

- Lint & unit tests: `golangci-lint run ./internal/server...` and `go test ./internal/server/...`
- WebDAV behaviors are covered by `webdav_*_test.go`; they rely on temp directories and in-memory routers.

### Operational Tips

- Prefer `Start` with context cancellation so graceful shutdown is triggered (`server.Close()`).
- When adding routes, register them in `registerRoutes` and reuse `MethodsGetHead` for safe verbs.
- Keep middleware light; log or enforce security at the edge (Traefik) when possible, but maintain server-side defaults for defense in depth.
