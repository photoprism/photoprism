## PhotoPrism — Customize Package

**Last Updated:** November 21, 2025

### Overview

The `customize` package defines user-facing configuration defaults for PhotoPrism’s Web UI, search, maps, imports, indexing, and feature flags. The settings are assembled by `NewDefaultSettings()` / `NewSettings()` and serialized through YAML so they can be stored or loaded at runtime.

### Feature Defaults

- Feature flags live in `FeatureSettings` and are initialized via the new `DefaultFeatures` variable.  
- `NewFeatures()` returns a copy of `DefaultFeatures`, letting callers mutate per-request or per-user state without modifying the shared defaults.

### Environment Overrides

- Set `PHOTOPRISM_DISABLE_FEATURES` to disable specific features at startup.  
- The value may be comma- or space-separated (case-insensitive); hyphens/underscores are ignored.  
- Tokens are inflected so singular/plural variants match (for example, `albums`, `album`, or `Album` all disable the Albums flag).

### Settings Lifecycle

- `NewDefaultSettings()` seeds UI, search, maps, imports, indexing, templates, downloads, and features from the defaults in this package.  
- `Settings.Load()` / `Save()` round-trip YAML configuration files.  
- `Settings.Propagate()` ensures required defaults (language, timezone, start page, map style) remain populated after loading.

### Testing

- Unit tests cover feature default copying, environment-based disabling, scope application, and ACL interactions.  
- Run `go test ./internal/config/customize/...` or the lints via `golangci-lint run ./internal/config/customize/...`.
