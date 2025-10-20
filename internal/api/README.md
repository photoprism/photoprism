# API Package Guide

## Overview

The API package exposes PhotoPrism’s HTTP endpoints via Gin handlers. Each file under `internal/api` contains the handlers, request/response DTOs, and Swagger annotations for a specific feature area. Handlers remain thin: they validate input, enforce security or ACL checks, and delegate domain work to services in `internal/photoprism`, `internal/service`, or other internal packages. Keep exported types aligned with the REST schema and avoid embedding business logic directly in handlers.

## Routing & Wiring

- Register handlers in `internal/server/routes.go` by attaching them to the proper router group (for example, `APIv1 := router.Group(conf.BaseUri("/api/v1"), Api(conf))`).
- Group endpoints by resource to match existing patterns: sessions, cluster, photos, labels, files, downloads, metadata, and technical routes.
- Apply middleware stacks (`Api`, `AuthRequired`, `limiter.Auth`, etc.) at the router level to keep handlers focused on request handling.
- Use `conf.BaseUri()` when constructing route prefixes so configuration overrides propagate consistently.
- When new endpoints require feature toggles, gate them in the router rather than inside the handler so disabled routes remain undiscoverable.

## Handler Implementation Patterns

- Accept and return JSON using the shared response helpers. Set `header.ContentTypeJSON` and ensure responses include proper cache headers (`no-store` for sensitive payloads).
- Parse parameters with Gin binding and validate inputs before delegating work. For complex payloads, define dedicated request structs with validation tags.
- Use the shared download helpers (`safe.Download`, `avatar.SafeDownload`) when calling outward HTTP APIs to inherit timeout, size, and SSRF protections.
- Query and persist data through the corresponding services or repositories; avoid ad-hoc SQL or GORM usage in handlers when dedicated functions exist elsewhere.
- Surface pagination consistently with `count`, `offset`, and `limit` following the defaults (100 max 1000). Validate `offset >= 0` and clamp `count` to the allowed range.
- When responses need role-specific fields, build DTOs that redact sensitive data for non-admin roles so the handler stays deterministic.

## Security & Middleware

- Authenticate requests using the standard middleware (`AuthRequired`) and check roles via helpers in `internal/auth/acl` (`acl.ParseRole`, `acl.ScopePermits`, `acl.ScopeAttrPermits`).
- Never log secrets or tokens. Prefer structured logging through `event.Log` and redact sensitive values before logging.
- Enforce rate limiting with the shared limiters (`limiter.Auth`, `limiter.Login`) and respond with `limiter.AbortJSON` to maintain consistent 429 JSON payloads.
- Derive client IPs through `api.ClientIP` and extract bearer tokens with `header.BearerToken` or the helper setters. Use constant-time comparison for tokens and secrets.
- For downloads or proxy endpoints, validate URLs against allowed schemes (`http`, `https`) and reject private or loopback addresses unless explicitly required.

## Swagger Documentation 

- Annotate handlers with Swagger comments that include full `/api/v1/...` paths, request/response schemas, and security definitions. Only annotate routes that are externally accessible.
- Regenerate docs after adding or updating handlers: `make fmt-go swag-fmt swag`. This formats Go files, normalizes annotations, and updates `internal/api/swagger.json`. Do not edit the generated JSON manually.
- When adding new DTOs, keep field names aligned with the JSON schema and update client documentation if serialized names change.
- Use enum annotations sparingly and ensure they reflect actual runtime constraints to avoid misleading generated clients.

## Testing Strategy

- Build tests around the API harness (`NewApiTest`) to obtain a configured Gin router, config, and dependencies. This isolates filesystem paths and avoids polluting global state.
- Wrap requests with helper functions (for example, `PerformRequestJSON`, `PerformAuthenticatedRequest`) to capture status codes, headers, and payloads. Assert headers using constants from `pkg/http/header`.
- When handlers interact with the database, initialize fixtures through config helpers such as `config.NewTestConfig("api")` or `config.NewMinimalTestConfigWithDb("api", t.TempDir())` depending on fixture needs.
- Stub external dependencies (`httptest.Server`) for remote calls and set `AllowPrivate=true` explicitly when the test server binds to loopback addresses.
- Structure tests with table-driven subtests (`t.Run("CaseName", ...)`) and use PascalCase names. Provide cleanup functions (`t.Cleanup`) to remove temporary files or databases created during tests.

## Focused Test Runs

- Fast iteration: `go test ./internal/api -run '<Package|HandlerName>' -count=1`
- Cluster endpoints: `go test ./internal/api -run 'Cluster' -count=1`
- Downloads and zip streaming: `go test ./internal/api -run 'Download|Archive' -count=1`
- Combined CLI and API validation: pair `go test ./internal/commands -run 'Cluster' -count=1` with the matching API suite to ensure DTOs remain compatible.

## Preflight Checklist

- Format and regenerate documentation: `make fmt-go swag-fmt swag`
- Compile backend: `go build ./...`
- Execute targeted API suites: `go test ./internal/api -run '<Name>' -count=1`
- Run integration-heavy checks before release: `go test ./internal/service/cluster/registry -count=1` alongside relevant API routes to confirm cluster DTOs stay aligned.
- Verify that `photoprism show commands --json` reflects any new API-driven flags or outputs when CLI exposure changes.
