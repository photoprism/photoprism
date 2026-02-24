## PhotoPrism — Glossary

**Last Updated:** February 23, 2026

### Purpose & Scope

- This is the single source of truth for terminology used across `specs/` and related docs.
- Define terms once here and reference this file instead of redefining the same terms in multiple documents.
- Keep technical/API contract names unchanged where required, even when user-facing wording differs.

### Canonical Terms

- **admin** — user or client with elevated authorization scopes/roles.
- **AdvertiseUrl** — internal/service URL that cluster peers use to reach an instance or service.
- **app** — intentional identifier in names/examples (for example `app.js`, `app.kubernetes.io/*`, `photoprism-app`); not the preferred generic runtime role term.
- **client** — OAuth/API client identity and credentials (`ClientID`, `ClientSecret`), and broadly a caller of an API.
- **cluster domain** — DNS domain used to derive cluster defaults (for example portal/instance URLs).
- **cluster UUID** — stable cluster identifier used by provisioning and cluster metadata.
- **instance** — PhotoPrism runtime with role `instance` (a cluster member serving UI/API/media features).
- **Join Token** — bootstrap bearer token used for initial registration (`/api/v1/cluster/nodes/register`).
- **node** — technical identifier used in API/config contracts (for example `/api/v1/cluster/nodes`, `NodeName`, `PHOTOPRISM_NODE_*`, `config/node/...` paths).
- **portal** — PhotoPrism runtime with role `portal`, providing cluster control-plane APIs and routing.
- **service** — PhotoPrism runtime with role `service` (non-instance cluster member focused on service workloads).
- **SiteUrl** — canonical public URL/origin for an instance.
- **tenant** — shared-domain routing ownership label used in path-based URLs such as `/i/<tenant>/...`; typically maps to a registered instance name.

### Writing Rules

- Use **instance**/**instances** for cluster runtime behavior and role language.
- Use **tenant**/**tenants** for shared-domain path ownership and routing semantics.
- Keep **node** where a field name, endpoint, flag/env var, config path, or code contract explicitly requires it.
- Keep **app** only where it is an intentional identifier/example.
- When a term appears ambiguous, link or refer back to this glossary.

### Change Management

- Update this file first when introducing, renaming, or clarifying core terminology.
- When terminology changes, update dependent docs to match these definitions.
