## PhotoPrism — Customize Package

**Last Updated:** October 23, 2026

### Overview

The `customize` package defines user-facing configuration defaults for PhotoPrism’s Web UI, search, maps, imports, indexing, and feature flags. The settings are assembled by `NewDefaultSettings()` / `NewSettings()` and serialized through YAML so they can be stored or loaded at runtime.

### Feature Defaults

- Feature flags live in `FeatureSettings` and are initialized via `DefaultFeatures`.
- `Export` (the Lightbox Export dialog) is enabled by default and can be turned off under Settings > General. Missing YAML fields keep the default; explicit `Export: false` and `Export: true` values survive save/load. It is independent of the Download setting and its content filters.
- `NewFeatures()` returns a copy of `DefaultFeatures`, letting callers mutate per-request or per-user state without modifying the shared defaults.

### Environment Overrides

- Set `PHOTOPRISM_DISABLE_FEATURES` to disable specific features at startup.  
- The value may be comma- or space-separated (case-insensitive); hyphens/underscores are ignored.  
- Tokens are inflected so singular/plural variants match (for example, `albums`, `album`, or `Album` all disable the Albums flag).

### Per-Role & Per-Scope Gating

- `Settings.ApplyACL(rules, role)` and `Settings.ApplyScope(scope)` filter selected feature flags for each session. Export requires both `photos.access_all` and `photos.download` grants and the `photos` scope, and the frontend's `canExport(config)` reads the resulting `features.export` flag. The downloads the dialog starts are authorized by the download and thumbnail endpoints like any other.
- The `Account` and `AppPasswords` flags require permission (and scope) to update the password (`ResourcePassword`/`ActionUpdate`), so a share-token visitor without an account does not see the account or app-password affordances.
- This per-session gating shapes only the client config the Web UI reads; server-side enforcement of disabled app passwords uses the global `AppPasswords` flag via `Config.DisableAppPasswords()`.

### Settings Lifecycle

- `NewDefaultSettings()` seeds UI, search, maps, imports, indexing, templates, downloads, and features from the defaults in this package.  
- `Settings.Load()` / `Save()` round-trip YAML configuration files.  
- `Settings.Propagate()` ensures required defaults (language, timezone, start page, map style) remain populated after loading.

### Testing

- Unit tests cover feature default copying, environment-based disabling, scope application, and ACL interactions.  
- Run `go test ./internal/config/customize/...` or the lints via `golangci-lint run ./internal/config/customize/...`.
