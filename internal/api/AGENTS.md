# API Guidelines

**Last Updated:** April 9, 2026

## Handler Conventions

- Reuse limiter stacks such as `limiter.Auth` and `limiter.Login`, and use `limiter.AbortJSON` for `429` responses.
- Lean on `api.ClientIP`, `header.BearerToken`, and the shared `Abort*` helpers instead of duplicating request parsing.
- Compare secrets in constant time and set `Cache-Control: no-store` on sensitive responses.
- Register routes in `internal/server/routes.go`.
- New list endpoints should default to `count=100`, cap at `1000`, require `offset >= 0`, and document their parameters explicitly.
- When a test or fixture needs portal mode, set `PHOTOPRISM_NODE_ROLE=portal` together with `PHOTOPRISM_JOIN_TOKEN`.

## Roles, Scopes & Sessions

- Map user roles through `acl.ParseRole(s)` and `acl.UserRoles[...]`; map client roles through `acl.ClientRoles[...]`.
- Treat `RoleAliasNone` (`none`) and the empty string as `RoleNone`; default unknown client roles to `RoleClient`.
- When checking JWT or client scopes, use `acl.ScopePermits` and `acl.ScopeAttrPermits` instead of handwritten parsing.
- Use `AuthenticateAdmin`, `AuthenticateUser`, or `OAuthToken` helpers in tests rather than building auth flows manually.
- Build client-session requests with `entity.AddClientSession(..., authn.GrantClientCredentials, nil)` and `AuthenticatedRequest(...)`.
- Admin sessions may see `AdvertiseUrl` and `Database`; client and user sessions must not. `SiteUrl` and the client `storageNamespace` derived from it are safe to expose.

## Swagger & API Tests

- Annotate only routed handlers in `internal/api/*.go`; use full `/api/v1/...` paths and skip helper functions.
- Regenerate API docs with `make fmt-go swag-fmt swag` or `make swag-json`; `make swag` may fetch modules, so confirm network access first.
- While iterating, prefer focused runs such as `go test ./internal/api -run Cluster -count=1`.
- Isolate config paths with `t.TempDir()` and reuse `NewConfig`, `CliTestContext`, and `NewApiTest()` helpers.
- For negative permission checks, prefer OAuth client tokens over non-admin user fixtures.
- Register `CreateSession(router)` only once per test router; a second registration panics on duplicate routes.

