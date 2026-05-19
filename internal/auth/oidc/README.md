## PhotoPrism — OIDC Integration

**Last Updated:** February 22, 2026

### Overview

`internal/auth/oidc` implements PhotoPrism’s OpenID Connect (OIDC) Relying Party (RP) flow so users can sign in with third‑party identity providers. The package wraps the `zitadel/oidc` client to perform discovery, build the RP, redirect users to the provider, exchange codes for tokens, and retrieve profile claims in a predictable, testable way.

#### Constraints

- Relies on the provider’s `/.well-known/openid-configuration` for discovery and enforces `https` unless explicitly allowed via `insecure`.
- Uses random per-session cookie keys (16‑byte hash + encrypt) and the shared HTTP client defined in `http_client.go`.
- PKCE is enabled automatically when the provider advertises `S256`.
- Scopes default to `authn.OidcRequiredScopes` when none are supplied; scopes are cleaned via `clean.Scopes`.
- Token exchange uses the provider’s userinfo endpoint by default; errors are surfaced via Gin response headers (`oidc_error`) and audit logs.

#### Goals

- Provide a consistent RP client that can be reused by CLI, server routes, and tests.
- Keep redirect and code‑exchange handlers minimal while ensuring audit visibility and secure defaults.
- Allow editions (CE/Pro) to extend claim processing (e.g., groups, roles) without duplicating RP wiring.

#### Non-Goals

- Managing upstream identity provider configuration or enrollment.
- Implementing a full OIDC Provider; PhotoPrism acts only as a Relying Party.
- Handling every custom claim set; extension hooks should live beside claim parsing code.

### Package Layout (Code Map)

- `oidc.go` — package doc + logger.
- `client.go` — RP construction (`NewClient`), PKCE detection, auth redirect, code exchange + userinfo retrieval.
- `http_client.go` — shared HTTP client with TLS toggle and timeouts; helpers for tests in `http_client_test.go`.
- `redirect_url.go` — builds the redirect/callback URL from site config.
- `register.go` — provider registration glue; tests in `register_test.go`.
- `username.go` — derives usernames from claims; tests in `username_test.go`.
- `client_test.go`, `oidc_test.go` — happy-path and error-path coverage for discovery, auth URL, and code exchange.

### Related Packages & Entry Points

- `internal/server/routes.go` registers the OIDC auth and callback endpoints.
- `pkg/authn` defines required scopes and shared auth helpers.
- `internal/auth/acl` and private extension LDAP packages (`pro/internal/auth/ldap`, `portal/internal/auth/ldap`) handle role/group mapping; the planned OIDC group parsing will mirror this logic.
- `internal/config` provides OIDC options/flags (issuer, client ID/secret, scopes, insecure).
- `internal/event` supplies the logger used for audit and error reporting.

### Configuration & Safety

- Enforce `https` for issuers unless `insecure` is explicitly set (intended for dev/test).
- Cookie handler is created per client with fresh random keys to avoid reuse across restarts.
- Audit every provider/redirect/token error with sanitized messages; avoid logging secrets.
- Prefer explicit scopes from configuration; defaults request only the minimal set.

### Security Group Extension for Entra ID

The following features are supported by the current implementation:

- Reads security groups from the `groups` claim in ID or access tokens; accepts GUIDs or names (case-insensitive, sanitized via `NormalizeGroupID`).
- Optional required membership: `--oidc-group` (or `PHOTOPRISM_OIDC_GROUP`) lists one or more groups that must be present; login is rejected if none match. If the token signals overage via `_claim_names.groups` and contains no groups, login is denied with an audit entry explaining that membership could not be validated.
- Group-to-role mapping: `--oidc-group-role` (`GROUP=ROLE`, repeatable) assigns the first matching role; falls back to `--oidc-role` (default `guest`) when no mapping matches.
- Keeps app/directory roles (`roles`, `wids`) separate from security groups to avoid accidental privilege escalation.
- Claim name is configurable via `--oidc-group-claim` (default `groups`).

#### Configuration Options

- `--oidc-group-claim` / `PHOTOPRISM_OIDC_GROUP_CLAIM`: claim to read (default `groups`).
- `--oidc-group` / `PHOTOPRISM_OIDC_GROUP`: comma- or multi-flag list of groups required for login (IDs or names accepted, normalized to lowercase alphanumerics/hyphen/underscore).
- `--oidc-group-role` / `PHOTOPRISM_OIDC_GROUP_ROLE`: mapping `GROUP=ROLE` (roles: `admin|manager|user|contributor|viewer|guest|none`). First match wins.
- `--oidc-role` / `PHOTOPRISM_OIDC_ROLE`: fallback role if no group mapping matches (defaults to `guest`).

#### Integration Guide for Entra ID

1. Register an app in Microsoft Entra ID (v2) or reuse your existing PhotoPrism registration. Note the instance ID and the application (client) ID.
2. Redirect URI: add [`https://{hostname}/api/v1/oidc/redirect`](https://docs.photoprism.app/getting-started/advanced/openid-connect/#redirect-url).
3. Token configuration → **Add optional claim** → **Token type** = ID (and Access if you prefer) → **Groups** → choose **Security groups**.
4. Under “Emit groups as”, pick **Group name** (cloud-only) or **sAMAccountName** / **DNSDomainName\sAMAccountName** for synced AD; this makes tokens carry human-friendly names instead of GUIDs.
5. If you keep **Group ID**, leave PhotoPrism config in GUID mode; if you emit names, set `PHOTOPRISM_OIDC_GROUP` / `PHOTOPRISM_OIDC_GROUP_ROLE` to those names (lowercase in config for consistency). When Microsoft signals group **overage** (too many groups to fit in the token), it sets `_claim_names.groups` and may omit groups entirely; PhotoPrism will currently block login if required groups are configured and no groups are present.
6. Grant admin consent for the chosen scopes (at minimum `openid profile email`, plus `offline_access` if you need refresh tokens).
7. Configure PhotoPrism (example `.env-oidc` with placeholder secrets):
   ```
   PHOTOPRISM_OIDC_URI="https://login.microsoftonline.com/f8b10857-a7f2-49ba-b73c-6f619715f574/v2.0"
   PHOTOPRISM_OIDC_CLIENT="11111111-2222-3333-4444-555555555555"
   PHOTOPRISM_OIDC_SECRET="asecure-random-oidc-client-secret"
   PHOTOPRISM_OIDC_GROUP_CLAIM="groups"
   PHOTOPRISM_OIDC_GROUP="photoprism-admins, photoprism-users"        # names or GUIDs
   PHOTOPRISM_OIDC_GROUP_ROLE="photoprism-admins=admin, photoprism-users=user"
   ```
8. Restart PhotoPrism; on login the service will:
   - Read groups from ID token, then fall back to userinfo if absent.
   - Deny login if required groups are configured but none are present (and overage is signaled).
   - Apply the first matching group→role mapping; otherwise assign the fallback role.

Please note:

- Entra ID security groups are only supported in PhotoPrism® Pro.
- If tokens still contain GUIDs, revisit Token configuration → Groups and change “Emit groups as” to a name format; reissue tokens by signing out/in. Names must be unique in your instance for deterministic mapping.
- Overage: when the `_claim_names.groups` marker is present and no groups are in the token, PhotoPrism cannot validate membership and will block login if `oidc-group` is set. (Graph-based resolution is described in the next section but is not yet implemented.)
- For mixed environments, you can supply both names and GUIDs in `oidc-group` / `oidc-group-role`; all entries are normalized and deduplicated.

#### Entra App Roles

As an alternative to security groups, we may use *Microsoft/Entra App Roles* to provide a more business-friendly option if needed:

- To implement this, PhotoPrism must read the `roles` claim, normalize it as with groups, and allow mapping by adding a new flag (e.g., `--oidc-role-claim=roles` or `--oidc-app-role=ROLE=photoprismRole`).
- This would require an estimated 80–150 lines of code (LOC), including wiring and tests, without introducing new dependencies.
- Once this feature is available, Entra admins can create app roles (e.g., `admin` or `viewer`) and assign them to users or groups in Entra.
- PhotoPrism would then receive readable role strings in tokens, eliminating the need to rely on security group names or GUIDs.

#### Microsoft Graph API

Support for the *Microsoft Graph API* is required to translate Entra security group GUIDs to display names and to fetch full membership lists when tokens omit groups:

- Resolve GUID → display name so `--oidc-group` / `--oidc-group-role` can use human-friendly group names while still matching IDs.
- Fetch memberships via Microsoft Graph when `_claim_names.groups` signals overage or when the token only carries IDs.
- Deduplicate and merge token groups with Graph results; continue to fall back gracefully if Graph is unavailable.

Implementation outline:

- Config: add flags/options such as `oidc-graph-lookup` (enable), `oidc-graph-timeout` (default ~3–5s), `oidc-graph-mode` (`client` for Client Credentials, `obo` for On-Behalf-Of), and optional scope override (default `https://graph.microsoft.com/.default`). Surface in flags, reports, and `options.yml`.
- Token acquisition:
  - Client Credentials flow using the existing OIDC client ID/secret against the instance token endpoint with Graph scope; requires admin-consented Application permission `Group.Read.All`.
  - On-Behalf-Of flow exchanging the user access token plus the same secret; requires Delegated `Group.Read.All` consent.
- Graph calls:
  - Prefer a single batch or `/v1.0/me/transitiveMemberOf?$select=id,displayName` to retrieve security groups; filter to `@odata.type` that ends with `group`.
  - Optionally fall back to `/v1.0/groups/{id}?$select=id,displayName` when only a few IDs need resolution.
- Processing: normalize both `id` and `displayName`, cache GUID→name mappings with a short TTL, merge into the existing group set, then apply required-group and group→role mapping logic.
- Testing: add httptest fixtures for token exchange and Graph responses, covering timeouts, 401/403, and partial data.

Impact:

- Allows administrators to configure PhotoPrism with recognizable group names instead of GUIDs.
- Makes log/debug output more readable and reduces reliance on Azure portal lookups for GUIDs.
- Provides a path to honor group-based access when tokens exceed size limits and omit groups by default.

#### Documentation & References

- Microsoft Entra ID: https://www.microsoft.com/en-us/security/business/identity-access/microsoft-entra-id
- Entra group claims: https://learn.microsoft.com/en-us/entra/identity-platform/access-token-claims-reference#groups-claim
- Entra app roles: https://learn.microsoft.com/en-us/entra/identity-platform/howto-add-app-roles-in-apps
- Group overage handling: https://learn.microsoft.com/en-us/entra/identity-platform/howto-add-app-roles-in-azure-ad-apps#group-overage-and-_claim_names
- Token customization guidance: https://learn.microsoft.com/en-us/entra/architecture/customize-tokens

### Operational Tips

- Always call `RedirectURL(siteUrl)` to build callbacks that respect reverse proxies and base URIs.
- Reuse `HttpClient(insecure)` so timeouts and TLS settings stay consistent.
- When adding claims processing, keep parsing isolated (e.g., new helper) and ensure failures do not block sign‑in unless required.

### Test Guidelines

- Unit tests: `go test ./internal/auth/oidc -count=1`
- Tests cover discovery failures, PKCE detection, redirect URL construction, username extraction, and code‑exchange error handling.
- For integration testing with a real IdP, set OIDC env vars in `compose.local.yaml`, start the dev server, and exercise `/auth/oidc` + callback.
