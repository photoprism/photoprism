# Config Package Guide

## Overview

PhotoPrism’s [runtime configuration](https://docs.photoprism.app/developer-guide/configuration/) is managed by this package. Fields are defined in [`options.go`](options.go) and then initialized with values from command-line flags, [environment variables](https://docs.photoprism.app/getting-started/config-options/), and [optional YAML files](https://docs.photoprism.app/getting-started/config-files/) (`storage/config/*.yml`).

## Sources and Precedence

PhotoPrism loads configuration in the following order:

1. **Built-in defaults** defined in this package.
2. **`defaults.yml`** — optional configuration defaults. PhotoPrism first checks `/etc/photoprism/defaults.yml` (or `.yaml`). If that file is missing or empty, it automatically falls back to `storage/config/defaults.yml` (respecting `.yml` / `.yaml` as well) under `PHOTOPRISM_CONFIG_PATH`. See [`defaults.yml`](https://docs.photoprism.app/getting-started/config-files/defaults/) if you package PhotoPrism for other environments and need to override the compiled defaults.
3. **Environment variables** prefixed with `PHOTOPRISM_…` and specified in [`flags.go`](flags.go) along with the CLI flags. This is the primary override mechanism in container environments.
4. **`options.yml`** — user-level configuration stored under `storage/config/options.yml` (or another directory controlled by `PHOTOPRISM_CONFIG_PATH`). Values here override both defaults and environment variables, see [`options.yml`](https://docs.photoprism.app/getting-started/config-files/).
5. **CLI flags** (for example `photoprism --cache-path=/tmp/cache`). Flags always win when a conflict exists.

The `PHOTOPRISM_CONFIG_PATH` variable controls where PhotoPrism looks for YAML files (defaults to `storage/config`).

> Any change to configuration (flags, env vars, YAML files) requires a restart. The Go process reads options during startup and does not watch for changes.

## Inspect Before Editing

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

## CLI Reference

- `photoprism help` (or `photoprism --help`) lists all subcommands and global flags.
- `photoprism show config` (alias `photoprism config`) renders every active option along with its current value. Pass `--json`, `--md`, `--tsv`, or `--csv` to change the output format.
- `photoprism show config-options` prints the description and default value for each option. Use this when updating [`flags.go`](flags.go).
- `photoprism show config-yaml` displays the configuration keys and their expected types in the [same structure that the YAML files use](https://docs.photoprism.app/getting-started/config-files/). It is a read-only helper meant to guide you when editing files under `storage/config`.
- Additional `show` subcommands document search filters, metadata tags, and supported thumbnail sizes; see [`internal/commands/show.go`](../commands/show.go) for the complete list.
