package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	utils "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Client represents an OpenID Connect (OIDC) Relying Party Client.
type Client struct {
	rp.RelyingParty
	insecure bool
	prompt   []string
}

// NewClient creates a new OpenID Connect (OIDC) Relying Party (RP) client using the provided discovery URL,
// client credentials, scopes, authorization prompt, and site URL.
func NewClient(issuerUri *url.URL, oidcClient, oidcSecret, oidcScopes, oidcPrompt, siteUrl string, insecure bool) (result *Client, err error) {
	if issuerUri == nil {
		err = errors.New("issuer uri required")
		event.AuditErr([]string{"oidc", "provider", status.Error(err)})
		return nil, errors.New("issuer uri required")
	} else if !insecure && issuerUri.Scheme != "https" {
		err = errors.New("issuer uri must use https")
		event.AuditErr([]string{"oidc", "provider", status.Error(err)})
		return nil, err
	}

	// Get redirect URL based on site URL.
	redirectUrl, urlErr := RedirectURL(siteUrl)

	if urlErr != nil {
		event.AuditErr([]string{"oidc", "redirect url", status.Error(urlErr)})
		return nil, urlErr
	}

	// Generate cryptographic keys.
	var hashKey, encryptKey []byte

	if hashKey, err = rnd.RandomBytes(16); err != nil {
		event.AuditErr([]string{"oidc", "hash key", status.Error(err)})
		return nil, err
	}

	if encryptKey, err = rnd.RandomBytes(16); err != nil {
		event.AuditErr([]string{"oidc", "encrypt key", status.Error(err)})
		return nil, err
	}

	// Create cookie handler. The short-lived state (CSRF defense) and PKCE
	// code_verifier cookies keep the Secure attribute on HTTPS deployments; it is
	// only dropped when running insecurely (HTTP issuer / relaxed TLS), gated by the
	// same flag that already permits a non-HTTPS issuer.
	var cookieOpts []utils.CookieHandlerOpt
	if insecure {
		cookieOpts = append(cookieOpts, utils.WithUnsecure())
	}
	// Scope the cookies to the OIDC endpoints under the instance base path instead
	// of the library default Path=/, so they survive to the callback without relying
	// on a shared-domain reverse proxy rewriting the Set-Cookie path.
	cookieOpts = append(cookieOpts, utils.WithPath(CookiePath(siteUrl)))
	cookieHandler := utils.NewCookieHandler(hashKey, encryptKey, cookieOpts...)

	// Create HTTP client.
	httpClient := HttpClient(insecure)

	// Set OIDC Relying Party client options.
	clientOpt := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(
			rp.WithIssuedAtOffset(5*time.Second),
			// Accept EdDSA — the PhotoPrism Portal OIDC OP signs ID tokens with
			// Ed25519 — alongside the default RS256 and the other common IdP
			// algorithms; the verifier otherwise rejects EdDSA-signed ID tokens
			// with "signature algorithm not supported".
			rp.WithSupportedSigningAlgorithms("RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512", "EdDSA"),
			// Disable the library's strict nonce check: its default expects an empty
			// nonce and would reject every value we echo. PhotoPrism validates the
			// nonce itself on the callback (see CheckNonce and README "Nonce Handling").
			rp.WithNonce(nil),
		),
		rp.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, errorType string, errorDesc string, state string) {
			event.AuditErr([]string{"oidc", "%s", "%s (state %s)"}, errorType, errorDesc, state)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("oidc_error", fmt.Sprintf("oidc: %s", errorDesc))
		}),
	}

	// Perform service discovery through the standardized /.well-known/openid-configuration endpoint.
	discover, err := client.Discover(context.TODO(), issuerUri.String(), httpClient)

	if err != nil {
		event.AuditErr([]string{"oidc", "provider", "service discovery", status.Error(err)})
		return nil, err
	}

	// If possible, use Proof of Key Code Exchange (PKCE).
	for _, v := range discover.CodeChallengeMethodsSupported {
		if v == oidc.CodeChallengeMethodS256 {
			clientOpt = append(clientOpt, rp.WithPKCE(cookieHandler))
		}
	}

	// Set default scopes if no scopes were specified.
	if oidcScopes == "" {
		oidcScopes = authn.OidcRequiredScopes
	}

	event.AuditDebug([]string{"oidc", "provider", "scopes", oidcScopes})

	// Parse scopes into string slice.
	scopes := clean.Scopes(oidcScopes)

	// Create RelyingParty provider.
	provider, err := rp.NewRelyingPartyOIDC(context.TODO(), issuerUri.String(), oidcClient, oidcSecret, redirectUrl, scopes, clientOpt...)

	if err != nil {
		event.AuditErr([]string{"oidc", "provider", status.Error(err)})
		return nil, err
	}

	if provider.IsPKCE() {
		event.AuditDebug([]string{"oidc", "provider", "pkce", "enabled"})
	} else {
		event.AuditDebug([]string{"oidc", "provider", "pkce", "disabled"})
	}

	// Validate the configured authorization prompt and drop unsupported values, so
	// a typo can never break the redirect to the identity provider.
	prompt, invalidPrompt := ParsePrompt(oidcPrompt)

	if len(invalidPrompt) > 0 {
		event.AuditWarn([]string{"oidc", "provider", "unsupported prompt %s", status.Skipped}, clean.Log(strings.Join(invalidPrompt, " ")))
	}

	// Return OIDC Client with RelyingParty provider.
	return &Client{
		RelyingParty: provider,
		insecure:     insecure,
		prompt:       prompt,
	}, nil
}

// AuthURLHandler redirects a browser to the login page of the configured OIDC identity provider.
func (c *Client) AuthURLHandler(ctx *gin.Context) {
	// Send a per-request nonce (stored in a signed, encrypted cookie) so the
	// provider reflects it back in the ID token; see README "Nonce Handling" for
	// the Cognito rationale. On generation or cookie failure, fall back to a
	// nonce-less redirect.
	var urlParams []rp.URLParamOpt

	if nonce, nonceErr := Nonce(); nonceErr != nil {
		event.AuditWarn([]string{"oidc", "nonce", status.Error(nonceErr)})
	} else if cookieErr := c.CookieHandler().SetCookie(ctx.Writer, NonceCookie, nonce); cookieErr != nil {
		event.AuditWarn([]string{"oidc", "nonce cookie", status.Error(cookieErr)})
	} else {
		urlParams = append(urlParams, rp.WithURLParam(nonceParam, nonce))
	}

	// Ask the provider to re-prompt (e.g. login or select_account) when configured,
	// so a previously rejected user is no longer silently re-authenticated via SSO.
	if len(c.prompt) > 0 {
		urlParams = append(urlParams, rp.WithPromptURLParam(c.prompt...))
	}

	handle := rp.AuthURLHandler(rnd.State, c, urlParams...)
	handle(ctx.Writer, ctx.Request)
}

// codeExchangeRecorder captures the OIDC code-exchange handler's status and
// headers while discarding its body, so a failure (e.g. a missing state cookie)
// is reported to the caller instead of being written as a raw error to the real
// response. The caller renders a branded page in its place.
type codeExchangeRecorder struct {
	header http.Header
	status int
}

func (w *codeExchangeRecorder) Header() http.Header { return w.header }
func (w *codeExchangeRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return len(b), nil
}
func (w *codeExchangeRecorder) WriteHeader(status int) { w.status = status }

// CodeExchangeUserInfo verifies a redirect auth request and returns the user information and tokens if successful.
func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (userInfo *oidc.UserInfo, tokens *oidc.Tokens[*oidc.IDTokenClaims], err error) {
	getInfo := func(w http.ResponseWriter, r *http.Request, t *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, i *oidc.UserInfo) {
		userInfo = i
		tokens = t
	}

	// Read and clear the per-request nonce cookie set on the authorization redirect
	// so the ID token's nonce claim can be validated once the exchange succeeds.
	expectedNonce, _ := c.CookieHandler().CheckCookie(ctx.Request, NonceCookie)
	c.CookieHandler().DeleteCookie(ctx.Writer, NonceCookie)

	// It would also be possible to directly get the user info from the oidc.IDTokenClaims
	// without performing a request to the userinfo endpoint of the OIDC identity provider.
	handle := rp.CodeExchangeHandler(rp.UserinfoCallback(getInfo), c)

	// Run the exchange against a recorder so a failure isn't written as a raw,
	// unbranded error to the browser; the caller renders a branded page instead.
	rec := &codeExchangeRecorder{header: make(http.Header)}
	handle(rec, ctx.Request)

	if sc := rec.status; sc != 0 && sc != http.StatusOK {
		if oidcErr := rec.header.Get("oidc_error"); oidcErr == "" {
			err = errors.New("failed to exchange token for user info")
		} else {
			err = errors.New(oidcErr)
		}

		event.SystemError([]string{"oidc", "code exchange", "status %d", "%s"}, sc, err.Error())

		return userInfo, tokens, err
	}

	// Validate the ID token's nonce against the value sent for this request,
	// tolerating a provider that omits the nonce on a session-resumed token.
	if err = CheckNonce(expectedNonce, tokens); err != nil {
		event.SystemError([]string{"oidc", "code exchange", "%s"}, err.Error())

		return nil, nil, err
	}

	// Propagate any cookies the handler set on success (e.g. clearing the
	// single-use state cookie) to the real response.
	for _, ck := range rec.header.Values("Set-Cookie") {
		ctx.Writer.Header().Add("Set-Cookie", ck)
	}

	return userInfo, tokens, nil
}
