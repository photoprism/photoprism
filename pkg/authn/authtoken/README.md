## PhotoPrism — Auth Token Package

**Last Updated:** July 24, 2026

`pkg/authn/authtoken` signs and verifies **bunny.net-compatible Advanced (HMAC-SHA256) URL tokens** — a
signature over a request's path, expiry, and query parameters that lets header-less endpoints (loaded
through `<img src>` / `<a href download>` contexts, which can carry only a `?t=` query token and no
`Authorization` header, possibly through a CDN) authorize a request without a server-side lookup.

The token format matches bunny.net Advanced Token Authentication, so the same token PhotoPrism mints is
validated at the origin today and, with a shared signing key, at the CDN edge later. The package is a
leaf under `pkg/` with no `internal/*` import, and holds **no key material and no clock of its own** —
callers pass the signing key and the current time, keeping key storage, lifetime, and rotation policy in
the config layer and making tests deterministic.

### Token Format

```text
token = "HS256-" + base64url( HMAC_SHA256(key, signaturePath + expires + signingData + clientIP) )
```

- `signaturePath` — the URL path, or a path prefix (a directory token that authorizes everything under it).
- `expires` — Unix expiry (seconds); carried alongside the token and folded into the signature so it
  cannot be extended without re-signing.
- `signingData` — the request's query parameters **except** `token` and `expires`, sorted by key and
  joined `key=value&…` (see `SigningData`).
- `clientIP` — empty unless IP locking is used.

The signed values live in the URL, not inside the token (a verifier rebuilds the message from the request
and recomputes the MAC), so any parameter the signature covers is visible — never place a secret in one.
Encoding is unpadded, URL-safe base64 with the `HS256-` prefix.

### API

| Function                                                        | Purpose                                                        |
|-----------------------------------------------------------------|---------------------------------------------------------------|
| `Sign(key, signaturePath, expires, params, clientIP)`           | Returns the `HS256-…` token for a request.                    |
| `Verify(key, signaturePath, expires, params, clientIP, token, now)` | Returns nil when the signature matches and has not expired.   |
| `Valid(key, signaturePath, expires, params, clientIP, token, now)`  | Boolean wrapper around `Verify`.                              |
| `SigningData(params)`                                           | Renders the signed query parameters (sorted, minus token/expires). |

`Verify` returns a typed error: `ErrExpired` (valid signature, past its deadline), `ErrSignature`
(forged, tampered, or wrong key), or `ErrMalformed` (empty token). The comparison is constant-time.

### Examples

Mint a directory token authorizing a session's downloads for one hour:

```go
key := conf.TokenSigningKey() // shared per-instance secret, never sent to clients
expires := time.Now().Add(time.Hour).Unix()
token := authtoken.Sign(key, "/api/v1", expires, url.Values{"sid": {sess.ID}}, "")
// → carried in the URL as ?t=<expires>.<sid>.<token> (compact) or ?token=…&expires=…&sid=… (CDN form)
```

Verify a token rebuilt from a header-less request:

```go
err := authtoken.Verify(key, "/api/v1", expires, url.Values{"sid": {sid}}, "", token, time.Now().Unix())
switch {
case errors.Is(err, authtoken.ErrExpired):
	// expired — the client will refresh from the next config/response
case err != nil:
	// forged or malformed — reject
default:
	sess, _ := entity.FindSession(sid) // scope the response to what this session may see
}
```

### Usage in PhotoPrism

`internal/auth/tokens` exposes a generic `Signer` (key + signature path) and one package-level instance
per token kind, configured by `Config.Propagate` (a leaf configured like `thumb`/`dl`/`ttl`, so it never
imports `config`). The first consumer is the **download token**: the `Download` signer plus the thin
`SignDownload` / `VerifyDownload` wrappers that add the per-session `sid` parameter, compact
`<expires>.<sid>.<token>` encoding, and sliding expiry. A future preview kind slots in as a second
`Signer` instance beside `Download` without a new package — sharing the same signing key, since a token's
scope is bound in the signed message, not the key. The key is a 32-byte secret at `config/keys/signing.key`
(`Config.TokenSigningKey()`, `fs.ModeSecretFile`), regenerated if missing and always held in memory so
signing never fails on a read-only disk; the lifetime is `ttl.DownloadToken`
(one hour by default, overridable via `PHOTOPRISM_DOWNLOAD_TOKEN_MAXAGE`). Download tokens are **stateless** —
nothing is stored per session or user. `tokens.DownloadToken(sessionID)` returns the value clients pass as
`?t=` (unchanged API contract): the `public` placeholder in public mode, a configured static token
(`PHOTOPRISM_DOWNLOAD_TOKEN`) delivered to every client for permanent-cache URLs, otherwise a signed
`<expires>.<sid>.<token>` — falling back to the single coarse instance token for the sessionless public/share
configs. `internal/api/auth_tokens.go` validates it: `DownloadSession` resolves a signed token to its bound
session (public mode → the public session), and `InvalidDownloadToken` also accepts the coarse token via
`tokens.IsCoarseDownload` (constant-time). Download URLs never touch a CDN edge (they use the API path), so
they use the compact single-parameter encoding; the verbose `token=…&expires=…&sid=…` form is reserved for
the CDN-fronted preview endpoints.

### Testing

```bash
go test ./pkg/authn/authtoken -count=1
```

Cases cover the sign/verify round-trip, `SigningData` ordering, and rejection of a wrong key, tampered
path/params/expires, expiry, and malformed input (`sign_test.go`, `parse_test.go`).
