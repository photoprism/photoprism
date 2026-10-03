## Dummy OIDC Identity Provider

**Last Updated:** September 30, 2026

A minimal OpenID Connect provider for local development and tests, built on `zitadel/oidc/v3`. It accepts any client ID and secret and authenticates without user interaction. The `dummy-oidc` service in `compose.yaml` runs the published image, which `make docker-dummy-oidc` builds from this directory.

### Issuer

- Requests that reach the service directly (`http://dummy-oidc:9998`) use the issuer `http://dummy-oidc:9998`.
- Requests that Traefik forwards from `https://dummy-oidc.localssl.dev` carry `X-Forwarded-Proto: https`, so the provider answers with the issuer `https://dummy-oidc.localssl.dev` and matching endpoints. Use this address where the client requires an `https` issuer, as PhotoPrism does for `PHOTOPRISM_OIDC_URI`.

### Test Identities

An authorization request selects a test user with its `login_hint` parameter. Without one, the default identity is used; an unknown hint is refused with `invalid_request`.

| `login_hint` | Subject       | Username   | Email                    | Verified | Other Claims                                     |
|--------------|---------------|------------|--------------------------|----------|--------------------------------------------------|
| *(none)*     | `sub00000001` | `prefname` | `test@example.com`       | yes      | -                                                |
| `olivia`     | `sub00000002` | `olivia`   | `olivia@example.com`     | yes      | `groups`: `photoprism-admin`, `staff`            |
| `oscar`      | `sub00000003` | `oscar`    | `oscar@example.com`      | yes      | `groups`: `photoprism-user`                      |
| `otto`       | `sub00000004` | `otto`     | `otto@example.com`       | no       | -                                                |
| `ola`        | `sub00000005` | `ola`      | `ola@example.org`        | yes      | -                                                |
| `odette`     | `sub00000006` | `odette`   | `odette@example.com`     | yes      | group overage (`_claim_names`/`_claim_sources`)  |
| `petra`      | `sub00000007` | `petra`    | `petra@example.com`      | yes      | `pp_issuer_kind`: `portal`, `pp_role`: `manager` |
| `alice`      | `sub00000008` | `alice`    | `alice.oidc@example.com` | yes      | name of a local account in the test fixtures     |
| `nobody`     | `sub00000009` | -          | `nobody@example.com`     | no       | no claim to derive a username from               |

The claims are included in the ID token and returned by the userinfo endpoint. Identities are defined in `app/mock/identities.go`.

### Tests

```bash
cd docker/dummy/oidc/app && go test ./...
```
