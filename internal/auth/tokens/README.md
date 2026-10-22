## PhotoPrism — URL Tokens

**Last Updated:** October 21, 2026

### Overview

`internal/auth/tokens` mints and validates the URL tokens PhotoPrism embeds in its header-less media URLs, so endpoints loaded through `<img src>` / `<a href download>` contexts — which carry only a `?t=` query parameter or a path segment and no `Authorization` header, possibly through a CDN — can be authorized and scoped: a signed token names its session, which the request side resolves, while per-user and static tokens are matched by value.

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
- `preview.go` — the `Preview` signer instance and the `PreviewToken` / `PreviewExpires` / `SignPreview` / `VerifyPreview` / `ParseCompact` wrappers for the session-bound preview tokens of OAuth2 clients without an account; tests in `preview_test.go`.
- `derive.go` — `Derive(key, purpose)` for the instance preview token; tests in `derive_test.go`.
- `oauth.go` — the `OAuth` signer instance and the `SignCode` / `VerifyCode` / `SignRefresh` / `VerifyRefresh` wrappers for the authorization codes and refresh tokens of the OAuth app flow, built on `Signer.SignValue` / `VerifyValue` and `SplitValue`; tests in `oauth_test.go`.

### Signing Key

`Config.TokenSigningKey()` returns a 32-byte instance secret stored at `config/keys/signing.key`
(`fs.ModeSecretFile`). It is regenerated when missing rather than backed up, and kept in memory even when
it cannot be persisted, so signing never fails on a read-only disk. A single key covers every token kind —
URL tokens bind their scope in the signed message, and OAuth values use a subkey derived per purpose —
and it never reaches clients. Refresh tokens live for months, so the key must survive restarts and be
shared by replicas: if `config/keys` cannot be written, is not restored with a backup, or differs
between replicas, apps connected with the OAuth app flow must be approved again, and their old grants
stay listed until the idle purge. `tokens.OAuth` always uses this instance key, also once URL tokens can be signed with a CDN key.

`Config.Propagate` configures the package before anything mints or verifies a token:

```go
tokens.Download.Key = c.TokenSigningKey()
tokens.Download.SignaturePath = c.ApiUri()
tokens.Preview.Key = c.TokenSigningKey()
tokens.Preview.SignaturePath = c.ApiUri()
tokens.OAuth.Key = c.TokenSigningKey()
tokens.PublicMode = c.Public()
tokens.CoarseDownload = c.DownloadToken()
```

### OAuth Codes & Refresh Tokens

`OAuth` appends an expiry and a MAC to the random values the OAuth app flow issues: a code becomes `<random>.<expires>.<mac>` and a refresh token `<GrantUID>.<random>.<expires>.<mac>`, where `<expires>` is a Unix time in canonical decimal form (as in the compact URL tokens) and the MAC is 16 bytes of HMAC-SHA256 in hex over everything before it, keyed by a subkey derived from the signing key for the purpose (`oauth-code`, `oauth-refresh`), so neither kind verifies as the other or as a URL token. `internal/auth/oauth` verifies the MAC and the expiry before it looks up a code or grant and stores the SHA-256 hash of the full value. A value is valid through the second of its expiry; `VerifyValue` returns `ErrInvalidValue` for a wrong shape or MAC, and the value with `ErrExpiredValue` once a value with a valid MAC has expired. Codes expire after `oauth.CodeTTL` and refresh tokens `oauth.RefreshTokenTTL` after issue, an hour after the grant's idle limit; the signed expiry only filters, while code lifetime, grant rotation and idle expiry stay in the database. A new signing key invalidates every outstanding code and refresh token, so connected apps sign in again (see Signing Key).

### Download Tokens

`Download` signs over `Config.ApiUri()` — a directory token authorizing every
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

`ClientSession` (`internal/config/client_config.go`) and `AddTokenHeaders` (`internal/api/api_response_headers.go`) deliver it only to a session that has a preview token of its own or receives signed preview tokens (see *Preview Tokens*): the download token is the higher-value credential, since it authorizes originals and a coarse one is cross-accepted for previews. Search handlers pass the same value into viewer results so the client can refresh it while browsing instead of polling.

`ClientSession` is also the only place a **preview** token is assigned to a client config. The base builders (`ClientUser`, `ClientShare`) carry neither token, so an account session without a preview token of its own receives neither; a client session without an account receives each token its scope (and, for previews, its client role) permits, and public mode assigns `entity.TokenPublic` explicitly.

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

Accounts, app passwords and OAuth app tokens carry their account's random preview token, which `entity.PreviewToken` matches by value. `Derive(key, PurposePreview)` folds the signing key into the stable hex token that `Config.PreviewToken()` returns: HMAC-SHA256 keeps the key unrecoverable from the value, and the purpose separates it from other kinds so one cannot be replayed as another. It is registered under `entity.TokenConfig` and accepted on the preview endpoints, but no client config delivers it, so an operator who pins one with `--preview-token` builds those URLs themselves. `Derive` takes the key as a parameter rather than reading `Download.Key`, so deriving the preview token in `Propagate` has no ordering dependency on the signer configuration.

Sessions of OAuth2 clients without an account (`Session.SignedPreview`) receive a signed token instead and store none. `Preview` signs the compact `<expires>.<sid>.<token>` form over the API path, with a `kind` parameter that is signed but not transmitted, so a download token never verifies as a preview token and the reverse. The expiry is the session's own, so the value stays the same for the whole session; a session without an expiry gets the end of the next UTC day (`PreviewExpires`), so its value changes once a day and is always valid for at least one more.

`internal/api/auth_tokens.go` verifies it: `InvalidPreviewToken(c, resources)` takes a value of the compact shape from the `:token` segment, or from `?t=` when the segment is `signed`, resolves its session through `PreviewSession` (memoized on the request), and refuses it unless the client is enabled and the session's scope and client role cover one of the endpoint's resources. Other values, including a `?t=` read because the segment cleaned to empty, are matched by value, with no session lookup. Such clients load previews from the API: `Config.ClientOAuth` sets their `contentUri` and `videoUri` to `ApiUri()`, and `SessionContentUri` does the same for the viewer URLs.

### Testing

```bash
go test ./internal/auth/tokens ./pkg/authn/authtoken -count=1
go test ./internal/api -run 'AuthDownload|DownloadSession|AddTokenHeaders|Preview|SessionContentUri' -count=1
go test ./internal/config -run 'TokenSigningKey|ClientSessionConfig|ClientOAuth|PreviewToken|DownloadToken' -count=1
```

Two gotchas when writing tests against token delivery:

- `TestConfig_ClientSessionConfig`-style tests need `c.Propagate()`, or the signer stays unconfigured and the coarse fallback is empty. Most session fixtures carry a `PreviewToken` and therefore assert a **non-empty**, signed download token; the user-less client and cluster fixtures carry none. A client session without an account takes its client's scope through `SetClient`, so build it from a client with the scope under test rather than editing the session's scope. Check the fixture before writing the assertion — `assert.NotEmpty` on a token proves nothing about *which* token was delivered.
- `config.NewMinimalTestConfig` runs in public mode, so `PreviewToken` / `DownloadToken` short-circuit to
  the public placeholder. Clear `Public` and `Demo` when exercising authenticated behavior.
