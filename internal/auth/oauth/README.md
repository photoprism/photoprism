## OAuth Authorization Code Flow for Apps

**Last Updated:** October 18, 2026

Package `oauth` implements the protocol side of the OAuth 2.1 authorization code flow with PKCE that lets users connect apps and remote MCP clients to an instance. API handlers stay thin and call it; `internal/entity` provides the storage. The Portal's cluster OIDC provider is separate and keeps its own code.

### Building Blocks

| File           | Purpose                                                                                                                                                        |
|----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `authorize.go` | `ParseAuthorizeRequest`, `ValidateClient` (client ID and redirect URI, checked before an error may be redirected) and `Validate` (all other parameter bounds). |
| `pkce.go`      | `VerifyPKCE` and `ValidCodeChallenge`; only `S256` is accepted.                                                                                                |
| `redirect.go`  | `ValidRedirectURI`, `MatchRedirectURI` (exact match; any port for HTTP loopback URIs) and the response URL builders, which always add `iss` (RFC 9207).        |
| `code.go`      | `IssueCode` and `RedeemCode` for single-use codes valid for one minute, and `PurgeExpired`, which the session cleanup worker calls.                            |
| `secret.go`    | `HashSecret`, the SHA-256 hash under which codes, request IDs and refresh tokens are stored.                                                                   |

### Rules

- `ValidateClient` checks syntax only. An error may be redirected to the client only after its redirect URI matched the client's registration (`MatchRedirectURI`); before that, it is rendered as a page.
- Redirect URIs are HTTPS URLs with a host, HTTP URLs on a loopback host (`localhost`, `127.0.0.0/8`, `::1`), or private-use schemes in reverse domain notation followed by a single slash (`app.example:/callback`, RFC 8252). Fragments and user info are refused. The response builders refuse any other URI and an empty issuer, keep the well-formed query parameters of the registered URI, and override `code`, `state`, `iss` and the error parameters, removing them when the response has no value for them.
- Repeated authorization request parameters are refused, and values must not contain control, format or line separator characters.
- A code is consumed by the first redemption attempt, even if the client, redirect URI or verifier do not match. A row without a user or resource is never redeemed.
- A pending request becomes a code only while it is unexpired (`entity.AuthCode.Replace`).
- Raw codes are returned to the client only; `auth_codes` stores their hash.
