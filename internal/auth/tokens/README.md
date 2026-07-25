## PhotoPrism — URL Tokens

**Last Updated:** July 25, 2026

### Overview

`internal/auth/tokens` mints and validates the URL tokens PhotoPrism embeds in its header-less media
URLs, so endpoints loaded through `<img src>` / `<a href download>` contexts — which carry only a `?t=`
query parameter and no `Authorization` header, possibly through a CDN — can be authorized and scoped
without a server-side lookup.

The cryptography itself lives in `pkg/authn/authtoken`, a dependency-free leaf that implements the
bunny.net-compatible Advanced (HMAC-SHA256) token format (see its `README.md` for the wire format and
signing API). This package adds everything that depends on the running instance: the configured keys, the
per-kind policy, and the delivery rules.

#### Constraints

- A Propagate-configured leaf like `thumb`/`dl`/`ttl`: it imports neither `config` nor `get`, so
  `Config.Propagate` pushes settings in rather than the package pulling them.
- Holds no clock or key material of its own beyond what `Propagate` assigns.
- The zero-value `Signer` is unconfigured and refuses to sign or verify, so a missing key degrades to
  rejection rather than to a forgeable empty-key HMAC.
- Tokens are **stateless** — nothing is stored per session or user.

### Package Layout (Code Map)

- `tokens.go` — package doc.
- `signer.go` — the generic `Signer` (key + signature path), `Configured`, `Sign`, `Valid`, and the
  `KeyLen` / `maxTokenLen` bounds; tests in `signer_test.go`.
- `download.go` — the `Download` signer instance, the `DownloadToken` / `SignDownload` /
  `VerifyDownload` / `IsCoarseDownload` policy wrappers, and the `PublicMode` / `CoarseDownload`
  delivery settings; tests in `download_test.go`.
- `derive.go` — `Derive(key, purpose)` for the not-yet-signed preview token; tests in `derive_test.go`.

### Signing Key

`Config.TokenSigningKey()` returns a 32-byte instance secret stored at `config/keys/signing.key`
(`fs.ModeSecretFile`). It is regenerated when missing rather than backed up, and kept in memory even when
it cannot be persisted, so signing never fails on a read-only disk. A single key covers every token kind —
a token's scope is bound in the signed message, not in the key — and it never reaches clients.

`Config.Propagate` configures the package before anything mints or verifies a token:

```go
tokens.Download.Key = c.TokenSigningKey()
tokens.Download.SignaturePath = c.ApiUri()
tokens.PublicMode = c.Public()
tokens.CoarseDownload = c.DownloadToken()
```

### Download Tokens

The only signed kind today. `Download` signs over `Config.ApiUri()` — a directory token authorizing every
download URL — and the thin wrappers add the per-kind policy: the per-session `sid` parameter, the compact
`<expires>.<sid>.<token>` encoding, and the `ttl.DownloadToken` lifetime (one hour by default, floor of 15
minutes, overridable via `PHOTOPRISM_DOWNLOAD_TOKEN_MAXAGE`).

#### Delivery

`tokens.DownloadToken(sessionID)` returns the value clients pass as `?t=`:

- public mode → the `public` placeholder, which is never verified;
- a session → always a signed `<expires>.<sid>.<token>`, even when a static token is configured, because
  the token is a query parameter on a one-shot transfer rather than part of the URL path, so a per-session
  value costs no cache sharing;
- no session to sign for → the coarse token (`PHOTOPRISM_DOWNLOAD_TOKEN`), which is empty unless an
  operator configured one, so a default instance neither hands out nor accepts an unscoped capability.

`ClientSession` (`internal/config/client_config.go`) and `AddTokenHeaders`
(`internal/api/api_response_headers.go`) deliver it only to a session that already has its own preview
token: the download token is the higher-value credential, since it authorizes originals and a coarse one
is cross-accepted for previews. Search handlers pass the same value into viewer results so the client can
refresh it while browsing instead of polling.

#### Validation

`internal/api/auth_tokens.go` resolves the request side:

- `AuthDownload(c) (sess, valid)` is the merged gate the endpoints call — session-or-coarse authorization
  and the resolved session in one call, auditing a denial centrally like `AuthAny`.
- `InvalidDownloadToken(c)` is a thin wrapper for callers that need only the yes/no answer.
- `DownloadSession(c)` resolves a signed token to its bound session and memoizes it on the request
  context. Public mode resolves to the public session, and a Portal cluster JWT in a request header is
  accepted ahead of the `?t=` token (restricted to JWTs granting `acl.AccessAll` on files, since a
  transient JWT session cannot back a `?t=` token).
- A configured coarse token is accepted separately via `IsCoarseDownload` (constant-time) as an unscoped
  capability with no session, so the handler applies the by-design broad public access.

Download URLs never touch a CDN edge — they use the API path — so clients use the compact
single-parameter form. The verbose bunny.net edge form (`token=…&expires=…&sid=…`) is also accepted and is
reserved for the CDN-fronted preview endpoints.

Scoped consumers: `DownloadAlbum` (`internal/api/download_album.go`), `GetDownload`
(`internal/api/download.go`), `GetPhotoDownload` (`internal/api/photos.go`), and `ZipDownload`
(`internal/api/zip.go`).

### Preview Tokens

Not signed yet. `Derive(key, PurposePreview)` folds the signing key into the stable hex token that
`Config.PreviewToken()` publishes in preview URLs: HMAC-SHA256 keeps the key unrecoverable from the
published value, and the purpose separates it from other kinds so one cannot be replayed as another.
`Derive` takes the key as a parameter rather than reading `Download.Key`, so deriving the preview token in
`Propagate` has no ordering dependency on the signer configuration.

Signed preview tokens slot in as a second `Signer` instance beside `Download` without a new package, since
the per-kind policy lives in the wrappers rather than in the generic signer.

### Testing

```bash
go test ./internal/auth/tokens ./pkg/authn/authtoken -count=1
go test ./internal/api -run 'AuthDownload|DownloadSession|AddTokenHeaders' -count=1
go test ./internal/config -run 'TokenSigningKey|ClientSessionConfig|PreviewToken|DownloadToken' -count=1
```

Two gotchas when writing tests against token delivery:

- `TestConfig_ClientSessionConfig`-style tests need `c.Propagate()`, or the signer stays unconfigured and
  the coarse fallback is empty. Most session fixtures have no `PreviewToken` and therefore correctly
  assert an **empty** download token.
- `config.NewMinimalTestConfig` runs in public mode, so `PreviewToken` / `DownloadToken` short-circuit to
  the public placeholder. Clear `Public` and `Demo` when exercising authenticated behavior.
