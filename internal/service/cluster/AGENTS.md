# Cluster Guidelines

**Last Updated:** April 9, 2026

## Bootstrap & Registration

- Keep bootstrap code decoupled: do not import `internal/service/cluster/node/*` from `internal/config` or the cluster root; nodes talk to the Portal over HTTP(S) and use `internal/service/cluster/const.go`.
- On `401` or `403`, bootstrap refreshes node OAuth credentials by rotating the secret and retrying; log that at info level. If the secret file cannot be written, keep the rotated value in memory.
- Portal validation may accept HTTP advertise URLs only for loopback or cluster-internal domains such as `*.svc`, `*.cluster.local`, and `*.internal`; all other advertise URLs must use HTTPS.
- Registration flow: send `rotate=true` only for MySQL or MariaDB nodes without credentials, treat `401`, `403`, and `404` as terminal, include `ClientID` plus `ClientSecret` when renaming an existing node, and persist only newly generated secrets or DB settings.
- Config init order for cluster-aware startup is: load `options.yml` with `c.initSettings()`, run `EarlyExt().InitEarly(c)`, connect or register the DB, then invoke `Ext().Init(c)`.

## Registry, DTOs & Provisioning

- Use `NewClientRegistryWithConfig`; the file-backed registry is legacy.
- Nodes are keyed by UUID v7 at `/api/v1/cluster/nodes/{uuid}`. Keep the registry interface UUID-first: `Get`, `FindByNodeUUID`, `FindByClientID`, `RotateSecret`, and `DeleteAllByUUID`.
- CLI lookups should resolve `uuid -> ClientID -> name`.
- DTOs normalize `Database.{Name,User,Driver,RotatedAt}` and expose `ClientSecret` only during creation or rotation.
- `nodes rm --all-ids` must clean duplicate client rows.
- Registry files live under `conf.PortalConfigPath()/nodes/` with mode `0600`, and `ClientData` no longer stores `NodeUUID`.
- Database and user names use UUID-based HMACs in `<prefix>d<hmac11>` and `<prefix>u<hmac11>` form; the prefix defaults to `cluster_` and may be overridden only by the portal-only `database-provision-prefix` flag.
- `BuildDSN` accepts a `driver` but falls back to MySQL format with a warning when the driver is unsupported.
- If Postgres provisioning is added, extend both `BuildDSN` and `provisioner.DatabaseDriver`, add validations, and return `driver=postgres` consistently in API and CLI output.

## Cluster API & Theme Changes

- When renaming or adding cluster response fields, update DTOs in `internal/service/cluster/response.go`, handlers, Swagger, tests, specs, and grep for old and new field names.
- The theme endpoint `GET /api/v1/cluster/theme` streams a zip from `conf.ThemePath()`. Reinstall only when `app.js` is missing and use the shared helpers in `pkg/http/header`.
- Admin responses may include `AdvertiseUrl` and `Database`; client and user sessions must remain redacted.

## Cluster Tests

- Generate OAuth client IDs with `rnd.GenerateUID(entity.ClientUID)` and node UUIDs with `rnd.UUIDv7()`; treat `node.uuid` as required in responses.
- Cluster registry tests under `internal/service/cluster/registry` intentionally use a full `config.TestConfig()` because they persist `entity.Client` rows. Do not switch them to minimal config helpers unless the tests stop touching the database.
- Exercise Portal endpoints with `httptest`, guard extraction paths with `pkg/fs.Unzip` size caps, and confirm admin-only fields disappear for client or user sessions.
- Portal proxy URI validation must use the Portal test environment with `NODES=2` and verify both instance routes when changing `PHOTOPRISM_PORTAL_PROXY_URI` or matching node `PHOTOPRISM_SITE_URL` prefixes; use `PORTAL_TEST_ENV_ARGS=--proxy-uri=/instance/` to regenerate consistent `.env` values.
- Before `make -C portal test-start`, run a full rebuild with `make -C portal test-env NODES=2`; avoid `--no-build` refreshes unless you are intentionally validating env-only changes.

## Cluster Preflight

- `go build ./...`
- `make fmt-go swag-fmt swag`
- `go test ./internal/service/cluster/registry -count=1`
- `go test ./internal/api -run 'Cluster' -count=1`
- `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1`
