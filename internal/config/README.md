## PhotoPrism — Config Package

**Last Updated:** March 8, 2026

### Overview

PhotoPrism’s [runtime configuration](https://docs.photoprism.app/developer-guide/configuration/) is managed by this package. Fields are defined in [`options.go`](options.go) and then initialized with values from command-line flags, [environment variables](https://docs.photoprism.app/getting-started/config-options/), and [optional YAML files](https://docs.photoprism.app/getting-started/config-files/) (`storage/config/*.yml`).

Client config values are derived from the runtime configuration and exposed to the frontend via `GET /api/v1/config`. This includes a `storageNamespace` value (SHA-256 hash of `SiteUrl`) used by the browser to scope namespaced browser-storage keys on shared domains.

### Storage Namespace & Legacy Session Compatibility

- `storageNamespace` is deterministic per `SiteUrl` (`SHA-256(SiteUrl)`) and is used by the frontend storage wrappers to isolate data on shared domains.
- The preferred frontend/browser contract uses namespaced keys in the format `pp:<storageNamespace>:<key>`, with `pp:root:` as the fallback prefix when no namespace is available.
- Frontend reads namespaced keys first and then falls back to selected legacy global keys; when a legacy value is found, it is migrated to the active namespace on read.
- New mobile or web-view integrations should write the namespaced `session` preference flag plus both namespaced `session.token` and `session.id` keys when pre-populating authentication data.
- Writing only a token is not enough to restore an authenticated user session in current frontend logic, because session restore requires both token and session id.
- A namespaced `localStorage["pp:<storageNamespace>:session"]` value of `"true"` selects namespaced `sessionStorage` for the active session. Any other value uses namespaced `localStorage`.
- The OIDC callback bridge still honors the legacy unnamespaced `localStorage["session"] === "true"` preference during migration, but new integrations should not depend on that fallback.
- Older compatibility keys (`authToken` / `sessionId`) are only auto-migrated when both are present.

### Sources & Precedence

PhotoPrism loads configuration in the following order:

1. **Built-in defaults** defined in this package.
2. **`defaults.yml`** — optional configuration defaults. PhotoPrism first checks `/etc/photoprism/defaults.yml` (or `.yaml`). If that file is missing or empty, it automatically falls back to `storage/config/defaults.yml` (respecting `.yml` / `.yaml` as well) under `PHOTOPRISM_CONFIG_PATH`. See [`defaults.yml`](https://docs.photoprism.app/getting-started/config-files/defaults/) if you package PhotoPrism for other environments and need to override the compiled defaults.
3. **Environment variables** prefixed with `PHOTOPRISM_…` and specified in [`flags.go`](flags.go) along with the CLI flags. This is the primary override mechanism in container environments.
4. **`options.yml`** — user-level configuration stored under `storage/config/options.yml` (or another directory controlled by `PHOTOPRISM_CONFIG_PATH`). Values here override both defaults and environment variables, see [`options.yml`](https://docs.photoprism.app/getting-started/config-files/).
5. **CLI flags** (for example `photoprism --cache-path=/tmp/cache`). Flags always win when a conflict exists.

The `PHOTOPRISM_CONFIG_PATH` variable controls where PhotoPrism looks for YAML files (defaults to `storage/config`).

> Any change to configuration (flags, env vars, YAML files) requires a restart. The Go process reads options during startup and does not watch for changes.

### HTTP Hardening Defaults

- `PHOTOPRISM_HTTP_HEADER_TIMEOUT` / `--http-header-timeout` configures `http.Server.ReadHeaderTimeout` and defaults to `15s`.
- `PHOTOPRISM_HTTP_HEADER_BYTES` / `--http-header-bytes` configures `http.Server.MaxHeaderBytes` and defaults to `1048576` (1 MiB).
- `PHOTOPRISM_HTTP_IDLE_TIMEOUT` / `--http-idle-timeout` configures `http.Server.IdleTimeout` and defaults to `180s`.
- `ReadTimeout` and `WriteTimeout` remain disabled globally so large uploads/downloads are not interrupted by a one-size-fits-all timeout.

### Inspect Before Editing

Before changing environment variables or YAML files, run `photoprism config | grep -i <flag>` to confirm the current value of a flag, such as `site-url`, or `site` to show all related values:

```bash
photoprism config | grep -i site
```

Example output:

| Name        | Value                     |
|:------------|:--------------------------|
| site-url    | https://app.localssl.dev/ |
| site-https  | true                      |
| site-domain | app.localssl.dev          |
| site-author | @photoprism_app           |
| site-title  | PhotoPrism                |

### CLI Reference

- `photoprism help` (or `photoprism --help`) lists all subcommands and global flags.
- `photoprism show config` (alias `photoprism config`) renders every active option along with its current value. Pass `--json`, `--md`, `--tsv`, or `--csv` to change the output format. Portal-only rows (`portal-proxy`, `portal-proxy-uri`, `portal-config-path`, `portal-theme-path`) are included only when `node-role` is set to `portal`.
- `photoprism show config-options` prints the description and default value for each option. Use this when updating [`flags.go`](flags.go).
- `photoprism show config-yaml` displays the configuration keys and their expected types in the [same structure that the YAML files use](https://docs.photoprism.app/getting-started/config-files/). It is a read-only helper meant to guide you when editing files under `storage/config`.
- Additional `show` subcommands document search filters, metadata tags, and supported thumbnail sizes; see [`internal/commands/show.go`](../commands/show.go) for the complete list.
- Pro/Portal builds additionally expose `PHOTOPRISM_THEME_URL` / `--theme-url` (hidden in CE/Plus), which can bootstrap `config/theme/` from a secure ZIP download when no theme files are present yet. HTTP Basic credentials in the URL are supported for protected artifact endpoints and are redacted in config reports.
