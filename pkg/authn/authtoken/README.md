## PhotoPrism — Auth Token Package

**Last Updated:** July 25, 2026

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

| Function                                                            | Purpose                                                            |
|---------------------------------------------------------------------|--------------------------------------------------------------------|
| `Sign(key, signaturePath, expires, params, clientIP)`               | Returns the `HS256-…` token for a request.                         |
| `Verify(key, signaturePath, expires, params, clientIP, token, now)` | Returns nil when the signature matches and has not expired.        |
| `Valid(key, signaturePath, expires, params, clientIP, token, now)`  | Boolean wrapper around `Verify`.                                   |
| `SigningData(params)`                                               | Renders the signed query parameters (sorted, minus token/expires). |

`Verify` returns a typed error: `ErrExpired` (valid signature, past its deadline), `ErrSignature`
(forged, tampered, or wrong key), or `ErrMalformed` (empty token). The comparison is constant-time.

### Examples

Mint a directory token authorizing a session's downloads for one hour:

```go
// key is a caller-supplied instance secret that never reaches clients.
expires := time.Now().Add(time.Hour).Unix()
token := authtoken.Sign(key, "/api/v1", expires, url.Values{"sid": {sessionID}}, "")
// → carried in the URL as ?t=<expires>.<sid>.<token> (compact) or ?token=…&expires=…&sid=… (CDN form)
```

Verify a token rebuilt from a header-less request:

```go
err := authtoken.Verify(key, "/api/v1", expires, url.Values{"sid": {sid}}, "", token, time.Now().Unix())
switch {
case errors.Is(err, authtoken.ErrExpired):
	// expired — the caller may hand the client a fresh token
case err != nil:
	// forged or malformed — reject
default:
	// authentic: sid identifies the session the request is scoped to
}
```

### Usage in PhotoPrism

The app-level wiring lives in `internal/auth/tokens` — the signing key, the per-kind signers, the delivery
rules, and the request-side gates. See `internal/auth/tokens/README.md`; that layer is deliberately not
described here, since this package must stay free of any dependency on it.

### Testing

```bash
go test ./pkg/authn/authtoken -count=1
```

Cases cover the sign/verify round-trip, `SigningData` ordering, and rejection of a wrong key, tampered
path/params/expires, expiry, and malformed input (`sign_test.go`, `parse_test.go`). `golden_test.go` pins
the wire format to a frozen token and cross-checks it against a from-scratch HMAC written to the
documented bunny.net algorithm, so a change to the message layout cannot silently diverge from what an
edge would validate.
