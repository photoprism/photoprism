# Config Guidelines

**Last Updated:** April 9, 2026

## Option Precedence & Wiring

- Respect config precedence: `options.yml` overrides CLI and environment values, which override defaults.
- When adding a new option, update `internal/config/options.go` for YAML and flag tags, register the flag in `internal/config/flags.go`, expose a getter, surface it in `*config.Report()`, and persist generated values by setting `c.options.OptionsYaml` before writing `options.yml`.
- Exercise new flags with `CliTestContext` from `internal/config/test.go`.

## Config Persistence

- For `options.yml` writes, use config-owned helpers instead of ad-hoc YAML handling: `Config.SaveOptionsPatch(...)` for generic merges and `Config.SaveClusterOptionsUpdate(...)` for cluster-managed updates.
- Use `pkg/fs.ConfigFilePath` when you need a config filename so existing `.yml` files stay valid while new installs may adopt `.yaml`.
- In Go code, use public `*config.Config` accessors such as `Config.JWKSUrl()`, `Config.SetJWKSUrl()`, and `Config.ClusterUUID()` instead of mutating `Config.Options()` directly; reserve raw option mutation for test fixtures.

## CLI Override Rules & DB Helpers

- Favor explicit CLI flags: check `c.cliCtx.IsSet("<flag>")` before overriding user-supplied values.
- Follow the `ClusterUUID` pattern for generated values: `options.yml`, then CLI or environment overrides, then a generated value persisted back to disk.
- Reuse `conf.Db()` and `conf.Database*()` helpers, avoid GORM `WithContext`, quote MySQL identifiers, and reject unsupported drivers early.

