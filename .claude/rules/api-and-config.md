## API & Config Changes

- Respect precedence: `options.yml` overrides CLI/env values, which override defaults.
- Adding a new option: update `internal/config/options.go` (yaml/flag tags), register in `internal/config/flags.go`, expose a getter, surface it in `*config.Report()`, and write generated values back to `options.yml`. Use `CliTestContext` in `internal/config/test.go` to exercise new flags.
- Adding a `customize.FeatureSettings` flag: a new field defaults to `true` via reflection (`features_default.go`) and is operator-disableable through `PHOTOPRISM_DISABLE_FEATURES` — no new CLI option needed. It cascades: update the full-struct literals in `internal/config/customize/{acl,scope}_test.go` and `internal/config/client_config_test.go` (the longest field name re-aligns every literal via gofmt), and `testdata/settings.yml` self-updates via `TestSettings_Save`. If the flag is only meaningful for accounts/roles, gate it per-session in `customize.Settings.ApplyACL` / `ApplyScope` (e.g. `Account`/`AppPasswords` require `ResourcePassword`/`ActionUpdate`); that shapes the Web UI client config only — enforce server-side behavior on the global flag via a `Config.DisableX()` helper.
- Identify an app-password credential by its session, not its token: `(*entity.Session).IsApplication()` (auth provider `application`) covers every grant that mints one (`password`/`session`/`cli`). The token format and grant type vary, so don't gate on `rnd.IsAppPassword` or `GrantType`.
- For `options.yml` writes, prefer config-owned persistence helpers: `Config.SaveOptionsPatch(...)` for generic merges, `Config.SaveClusterOptionsUpdate(...)` for cluster-managed metadata.
- Use `pkg/fs.ConfigFilePath` for config filenames so existing `.yml` files stay valid and new installs can adopt `.yaml` transparently.
- Use the public accessors on `*config.Config` (e.g. `JWKSUrl()`, `SetJWKSUrl()`) instead of mutating `Config.Options()` directly; reserve raw option tweaks for test fixtures.
- New metadata sources (e.g. `SrcOllama`, `SrcOpenAI`) must be defined in both `internal/entity/src.go` and the frontend lookup tables (`frontend/src/common/util.js`).
- Config init order: load `options.yml` (`c.initSettings()`), run `EarlyExt().InitEarly(c)`, connect/register the DB, then `Ext().Init(c)`.
- Favor explicit CLI flags: check `c.cliCtx.IsSet("<flag>")` before overriding user-supplied values.
- Database helpers: reuse `conf.Db()` / `conf.Database*()`, avoid GORM `WithContext`, quote MySQL identifiers, and reject unsupported drivers early.

## Handler Conventions

- Reuse limiter stacks (`limiter.Auth`, `limiter.Login`) and `limiter.AbortJSON` for 429s. Lean on `api.ClientIP`, `header.BearerToken`, and `Abort*` helpers.
- Compare secrets with constant-time checks; set `Cache-Control: no-store` on sensitive responses.
- Register routes in `internal/server/routes.go`. New list endpoints default `count=100` (max 1000) and `offset≥0`; document parameters explicitly.
- Set portal mode via `PHOTOPRISM_NODE_ROLE=portal` plus `PHOTOPRISM_JOIN_TOKEN` when needed.

## API Shape Checklist

When renaming or adding fields:
- Field casing: **TitleCase** (`UUID`, `Name`, `SiteUrl`) for fields backed by a DB entity (mirror the entity/model), **camelCase** (`storageNamespace`, `redirectUri`) for generated/artificial payloads (client config, session, action/RPC bodies). A filtered/computed entity projection stays TitleCase; an action payload stays camelCase but MAY TitleCase its single entity-identity field (e.g. `UUID`). See `specs/common/field-casing.md`.
- Update DTOs in `internal/service/cluster/response.go` and any mappers.
- Update handlers and regenerate Swagger: `make fmt-go swag-fmt swag`.
- Update tests (search/replace old field names) and examples in `specs/`.
- Quick grep: `rg -n 'oldField|newField' -S` across code, tests, and specs.

## Testing Helpers

- Isolate config paths with `t.TempDir()`; reuse `NewConfig`, `CliTestContext`, and `NewApiTest()` harnesses.
- Authenticate via `AuthenticateAdmin`, `AuthenticateUser`, or `OAuthToken`. Toggle auth with `conf.SetAuthMode(config.AuthModePasswd)`.
- Prefer OAuth client tokens over non-admin fixtures for negative permission checks.

## Roles & ACL

- Map roles via the shared tables: users through `acl.ParseRole(s)` / `acl.UserRoles[...]`, clients through `acl.ClientRoles[...]`.
- Treat `RoleAliasNone` ("none") and an empty string as `RoleNone`; default unknown client roles to `RoleClient`.
- Build CLI role help from `Roles.CliUsageString()` (e.g. `acl.ClientRoles.CliUsageString()`) — never hand-maintain role lists.
- For JWT/client scope checks, use the shared helpers (`acl.ScopePermits` / `acl.ScopeAttrPermits`).
