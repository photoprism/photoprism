## PhotoPrism — WebDAV Service Client

**Last Updated:** March 7, 2026

### Overview

`internal/service/webdav` contains the outbound WebDAV client used by PhotoPrism services and background workers. It wraps `github.com/emersion/go-webdav` with PhotoPrism-specific URL validation, control-operation timeouts, filesystem mapping, and error/logging behavior for remote uploads, downloads, and synchronization.

### Main Responsibilities

- Validate remote endpoints against `services-cidr` before opening outbound connections.
- Normalize remote paths and expose them as `pkg/fs.FileInfo` values.
- Support uploads, downloads, deletes, and remote directory creation.
- Enumerate remote directories for service folder browsing and sync refresh jobs.
- Exclude hidden dotfiles and entries inside hidden dot-directories because these are often lock files, partial uploads, or provider-managed metadata.

### Recursive Directory Discovery

`Client.Directories(dir, recursive, timeout)` is the main entry point for remote folder discovery.

- Fast path: when `recursive=true`, the client first performs a recursive `PROPFIND` using `Depth: infinity` through the upstream WebDAV library.
- Compatibility fallback: if the recursive request fails, the client retries discovery by walking the tree with repeated non-recursive `PROPFIND` requests using `Depth: 1`.
- Scope: this fallback is intentionally limited to directory enumeration. It does not change upload, download, delete, or file-listing semantics.

This behavior exists because some providers and appliances accept `Depth: 1` but reject `Depth: infinity`. The fallback keeps those servers usable for:

- service folder browsing via `entity.Service.Directories()`,
- sync refresh in `internal/workers/sync_refresh.go`.

### Timeout Behavior

Available timeout settings for `Service.AccTimeout` and `webdav.Timeout`:

| Setting | Value      | Effective Timeout |
|:--------|:-----------|:------------------|
| Default | `""`       | `60s`             |
| Medium  | `"medium"` | `60s`             |
| Low     | `"low"`    | `30s`             |
| High    | `"high"`   | `120s`            |
| None    | `"none"`   | no timeout        |

- `Timeout` values map to total HTTP request timeouts for non-transfer WebDAV calls such as directory discovery, file listing, directory creation, and delete operations.
- `Upload()` and `Download()` intentionally bypass the service timeout so long-running file transfers are not aborted by a total request deadline.
- Transfer requests still apply connection-level safeguards such as connect, TLS handshake, and pooled idle connection limits to avoid hanging before a transfer is established.
- In timeout-aware helper calls, `timeout=0` means "use the client's configured default timeout" (`c.timeout`), not "disable timeouts".
- A negative helper timeout means "do not override the current client/request timeout behavior"; this is used internally for legacy no-override call paths.
- Recursive directory discovery also applies the effective timeout as an overall traversal deadline, so iterative fallback walks do not run indefinitely.
- `MaxRequestDuration` is used for long-running recursive directory discovery, including the `Depth: 1` fallback.

### Logging

When a recursive `PROPFIND` fails, the client logs the failure and emits an informational message if it successfully switches to the iterative `Depth: 1` fallback. Successful fallback logs include the number of follow-up `PROPFIND` requests and the elapsed traversal time so operators can diagnose depth-limited servers without reducing the user-facing API response to only "could not connect".

### Package Layout

- `webdav.go` — package comment, timeout constants, and shared logger.
- `client.go` — outbound WebDAV client wrapper and compatibility fallback.
- `path.go` — shared path normalization helpers.
- `client_test.go` — unit tests, including a local `httptest` WebDAV fixture for depth-limited servers.

### Related Files

- [`internal/entity/service.go`](../../entity/service.go) — service-level directory discovery.
- [`internal/workers/sync_refresh.go`](../../workers/sync_refresh.go) — sync refresh that enumerates remote directories before file listing.
- [`scripts/dav-probe.sh`](../../../scripts/dav-probe.sh) — captures `PROPFIND` responses for troubleshooting remote server behavior.

### Testing

- Focused client tests: `go test ./internal/service/webdav -run 'TestClient_Directories' -count=1`
- Service-level regression checks: `go test ./internal/entity -run 'TestService_Directories' -count=1`

The local test server in `client_test.go` simulates both compliant servers and depth-1-only servers so the fallback can be validated without relying on the external dummy WebDAV container.
