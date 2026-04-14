## Cluster Operations

- Keep bootstrap code decoupled: avoid importing `internal/service/cluster/node/*` from `internal/config` or the cluster root, let nodes talk to the Portal over HTTP(S), and rely on constants from `internal/service/cluster/const.go`.
- Bootstrap refreshes node OAuth credentials on 401/403 responses (rotate secret + retry) and logs the refresh at info level. If the secret file cannot be written, the rotated value stays cached in memory so the current process can continue.
- Portal validation now accepts HTTP advertise URLs only for loopback hosts or cluster-internal domains (`*.svc`, `*.cluster.local`, `*.internal`); everything else must use HTTPS.
- Theme endpoint: `GET /api/v1/cluster/theme` streams a zip from `conf.ThemePath()`; only reinstall when `app.js` is missing and always use the header helpers in `pkg/http/header`.
- Registration flow: send `rotate=true` only for MySQL/MariaDB nodes without credentials, treat 401/403/404 as terminal, include `ClientID` + `ClientSecret` when renaming an existing node, and persist only newly generated secrets or DB settings.

### Registry & DTOs

- Use the client-backed registry (`NewClientRegistryWithConfig`)—the file-backed version is legacy.
- Nodes are keyed by UUID v7 (`/api/v1/cluster/nodes/{uuid}`), the registry interface stays UUID-first (`Get`, `FindByNodeUUID`, `FindByClientID`, `RotateSecret`, `DeleteAllByUUID`).
- CLI lookups resolve `uuid → ClientID → name`.
- DTOs normalize `Database.{Name,User,Driver,RotatedAt}` while exposing `ClientSecret` only during creation/rotation.
- `nodes rm --all-ids` cleans duplicate client rows.
- Admin responses may include `AdvertiseUrl`/`Database`; client/user sessions stay redacted.
- Registry files live under `conf.PortalConfigPath()/nodes/` (mode 0600).
- `ClientData` no longer stores `NodeUUID`.

### Provisioner & DSN

- Database/user names use UUID-based HMACs (`<prefix>d<hmac11>`, `<prefix>u<hmac11>` where the prefix defaults to `cluster_`).
- `BuildDSN` accepts a `driver` but falls back to MySQL format with a warning when unsupported.
- If Postgres provisioning is added, extend `BuildDSN` and `provisioner.DatabaseDriver` handling, add validations, and return `driver=postgres` consistently in API and CLI output.

### Sessions & Redaction

- Admin session (full view): `AuthenticateAdmin(app, router)`.
- User session: Create a non-admin test user (role=guest), set a password, then `AuthenticateUser`.
- Client session (redacted internal fields; `SiteUrl` visible):
  ```go
  s, _ := entity.AddClientSession("test-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
  token := s.AuthToken()
  r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/nodes", token)
  ```
- Admins see `AdvertiseUrl` and `Database`; client/user sessions don't. `SiteUrl` is safe to show to all roles.
- Client config includes `storageNamespace` (SHA-256 of `SiteUrl`) for browser storage scoping.

### Preflight Checklist

- `go build ./...`
- `make fmt-go swag-fmt swag`
- `go test ./internal/service/cluster/registry -count=1`
- `go test ./internal/api -run 'Cluster' -count=1`
- `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1`
- Tooling constraints: `make swag` may fetch modules, so confirm network access before running it.
