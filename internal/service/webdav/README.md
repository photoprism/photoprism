## PhotoPrism — WebDAV Service Client

**Last Updated:** October 16, 2026

### Overview

`internal/service/webdav` contains the outbound WebDAV client used by PhotoPrism services and background workers. It wraps `github.com/emersion/go-webdav` with PhotoPrism-specific URL validation, control-operation timeouts, filesystem mapping, and error/logging behavior for remote uploads, downloads, and synchronization.

### Main Responsibilities

- Validate remote endpoints against `services-cidr` before opening outbound connections.
- Normalize remote paths and expose them as `pkg/fs.FileInfo` values.
- Support uploads, downloads, deletes, and remote directory creation.
- Enumerate remote directories for service folder browsing and sync refresh jobs.
- Exclude hidden dotfiles and entries inside hidden dot-directories because these are often lock files, partial uploads, or provider-managed metadata. Other names that are hidden locally, such as `@eaDir` or `__MACOSX`, are transferred, and remote names with `$` or `~` are skipped because they are not portable. Paths are checked relative to the account URL, so its own path may contain a dot-directory, e.g. `/.dav/`.

### Recursive Directory Discovery

`Client.Directories(dir, recursive, timeout)` is the main entry point for remote folder discovery.

- Fast path: when `recursive=true`, the client first performs a recursive `PROPFIND` using `Depth: infinity` through the upstream WebDAV library.
- Compatibility fallback: if the recursive request fails, the client retries discovery by walking the tree with repeated non-recursive `PROPFIND` requests using `Depth: 1`.
- Folders are queued relative to the account URL, since servers return paths that include its path.
- Scope: this fallback is intentionally limited to directory enumeration. It does not change upload, download, delete, or file-listing semantics.

This behavior exists because some providers and appliances accept `Depth: 1` but reject `Depth: infinity`. The fallback keeps those servers usable for:

- service folder browsing via `entity.Service.Directories()`,
- sync refresh in `internal/workers/sync_refresh.go`.

### Transfer Path Policy

Directory discovery retains its existing hidden-name and traversal exclusions. Logical
transfer paths are also checked before listing, upload, download, directory creation, or
deletion: `SkipSyncPath` reports an excluded name and `UnsafeSyncPath` reports one carrying a
parent-directory segment. Reserved names use the shared `pkg/fs` set; other components that
start with a dot are omitted from sync as well, as are names containing a C0 control character or DEL. These are
checked after the parent-directory segment, so such a path still reports as unsafe. Endpoint prefixes and absolute local storage roots are not
passed as logical transfer paths.

Excluded entries are normal skips, not remote failures: they return `ErrSkipPath`, and sync
and share workers store an `ignore` disposition with empty error fields, so an ignored leading
batch still allows later eligible work to progress. A path carrying a parent-directory segment
returns `ErrUnsafePath` instead and is recorded as a transfer failure. The exclusion is checked
first, so a name that is both reserved and traversing reports the skip. Workers check original
relative source names before aliases or thumbnails can replace them, and recheck queued paths.
Existing traversal checks, safe joins, download-size limits, overwrite protection, timeouts,
and service network restrictions remain in force.

### Upload Responses

`Upload()` sends the `PUT` request itself with the file size as `Content-Length`. A `403 Forbidden` response returns an error wrapping `ErrForbidden`, so callers can tell a refused file from a failed transfer. Any other non-2xx status returns an error naming the status code, and so does a `301`, `302`, or `303` redirect, since the HTTP client follows it with a `GET` whose status does not describe the upload. A transfer that ends because too little data moved returns an error wrapping `ErrStalled`; `Download` does the same.

### Timeout Behavior

Available timeout settings for `Service.AccTimeout` and `webdav.Timeout`:

| Setting | Value      | Effective Timeout                 |
|:--------|:-----------|:----------------------------------|
| Default | `""`       | `60s`                             |
| Medium  | `"medium"` | `60s`                             |
| Low     | `"low"`    | `30s`                             |
| High    | `"high"`   | `120s`                            |
| None    | `"none"`   | 30 minutes (`MaxRequestDuration`) |

- `Timeout` values map to total HTTP request timeouts for non-transfer WebDAV calls such as directory discovery, file listing, directory creation, and delete operations.
- `Upload()` and `Download()` intentionally bypass the service timeout so long-running file transfers are not aborted by a total request deadline. Instead, a transfer ends when less than `transferStallBytes` (1 MiB) moves within `transferStallTimeout` (5 minutes), about 3.5 KB/s, so a remote that stops or nearly stops sending or reading ends the transfer; the error says the transfer stalled. Transfers have no total deadline: one that keeps moving at least that much data runs until it completes, however long the file takes. A download's wait for the response counts toward the stall timeout. An upload moves only while the socket send buffer drains, and once its body is sent, the wait for the response is 5 minutes plus one second per MiB of the file (`responseWait`), as servers may scan or move a large file before they answer.
- Transfer requests still apply connection-level safeguards such as connect, TLS handshake, and pooled idle connection limits to avoid hanging before a transfer is established.
- In timeout-aware helper calls, `timeout=0` means "use the client's configured default timeout" (`c.timeout`), not "disable timeouts". With the setting "None", these calls are bounded by `MaxRequestDuration` (30 minutes).
- A negative helper timeout means "do not override the current client/request timeout behavior"; no caller passes one at the moment.
- Recursive directory discovery also applies the effective timeout as an overall traversal deadline, so iterative fallback walks do not run indefinitely.
- `MaxRequestDuration` is used for long-running recursive directory discovery, including the `Depth: 1` fallback.

### Logging

When a recursive `PROPFIND` fails, the client logs the failure and emits an informational message if it successfully switches to the iterative `Depth: 1` fallback. Successful fallback logs include the number of follow-up `PROPFIND` requests and the elapsed traversal time so operators can diagnose depth-limited servers without reducing the user-facing API response to only "could not connect".

### Package Layout

- `webdav.go` — package comment, timeout constants, and shared logger.
- `client.go` — outbound WebDAV client wrapper and compatibility fallback.
- `path.go` — shared path normalization helpers.
- `stall.go` — stall timeout of uploads and downloads (`stallWatch`, `responseWait`); `stall_test.go` covers it with slow, stalling and trickling test servers.
- `client_test.go` — unit tests, including a local `httptest` WebDAV fixture for depth-limited servers.
- `transfer_policy_test.go` — exclusion and containment regressions for the transfer and listing calls.

### Related Files

- [`internal/entity/service.go`](../../entity/service.go) — service-level directory discovery.
- [`internal/workers/sync_refresh.go`](../../workers/sync_refresh.go) — sync refresh that enumerates remote directories before file listing.
- [`scripts/dav-probe.sh`](../../../scripts/dav-probe.sh) — captures `PROPFIND` responses for troubleshooting remote server behavior.

### Testing

- Focused client tests: `go test ./internal/service/webdav -run 'TestClient_Directories' -count=1`
- Upload responses: `go test ./internal/service/webdav -run 'TestClient_Upload' -count=1`
- Stall timeout and recovered panics: `go test ./internal/service/webdav -run 'TestStallWatch|Stall|RecoveredPanic|timeoutContext' -count=1`
- Service-level regression checks: `go test ./internal/entity -run 'TestService_Directories' -count=1`

The local test server in `client_test.go` simulates both compliant servers and depth-1-only servers so the fallback can be validated without relying on the external dummy WebDAV container.
