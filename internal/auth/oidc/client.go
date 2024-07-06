package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	utils "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Client represents an OpenID Connect (OIDC) Relying Party Client.
type Client struct {
	rp.RelyingParty
	insecure bool
}

// NewClient creates and returns a new OpenID Connect (OIDC) Relying Party Client based on the specified parameters.
func NewClient(issuerUri *url.URL, oidcClient, oidcSecret, oidcScopes, siteUrl string, insecure bool) (result *Client, err error) {
	if issuerUri == nil {
		err = errors.New("issuer uri required")
		event.AuditErr([]string{"oidc", "provider", "%s"}, err)
		return nil, errors.New("issuer uri required")
	} else if insecure == false && issuerUri.Scheme != "https" {
		err = errors.New("issuer uri must use https")
		event.AuditErr([]string{"oidc", "provider", "%s"}, err)
		return nil, err
	}

	// Get redirect URL based on site URL.
	redirectUrl, urlErr := RedirectURL(siteUrl)

	if urlErr != nil {
		event.AuditErr([]string{"oidc", "redirect url", "%s"}, err)
		return nil, err
	}

	// Generate cryptographic keys.
	var hashKey, encryptKey []byte

	if hashKey, err = rnd.RandomBytes(16); err != nil {
		event.AuditErr([]string{"oidc", "hash key", "%s"}, err)
		return nil, err
	}

	if encryptKey, err = rnd.RandomBytes(16); err != nil {
		event.AuditErr([]string{"oidc", "encrypt key", "%s"}, err)
		return nil, err
	}

	// Create cookie handler.
	cookieHandler := utils.NewCookieHandler(hashKey, encryptKey, utils.WithUnsecure())

	// Create HTTP client.
	httpClient := HttpClient(insecure)

	// Set OIDC Relying Party client options.
	clientOpt := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(
			rp.WithIssuedAtOffset(5 * time.Second),
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
		event.AuditErr([]string{"oidc", "provider", "service discovery", "%s"}, err)
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
		oidcScopes = "openid email profile"
	}

	event.AuditDebug([]string{"oidc", "provider", "scopes", oidcScopes})

	// Parse scopes into string slice.
	scopes := clean.Scopes(oidcScopes)

	// Create RelyingParty provider.
	provider, err := rp.NewRelyingPartyOIDC(context.TODO(), issuerUri.String(), oidcClient, oidcSecret, redirectUrl, scopes, clientOpt...)

	if err != nil {
		event.AuditErr([]string{"oidc", "provider", "%s"}, err)
		return nil, err
	}

	if provider.IsPKCE() {
		event.AuditDebug([]string{"oidc", "provider", "pkce", "enabled"})
	} else {
		event.AuditDebug([]string{"oidc", "provider", "pkce", "disabled"})
	}

	// Return OIDC Client with RelyingParty provider.
	return &Client{
		provider,
		insecure,
	}, nil
}

// AuthCodeUrlHandler redirects a browser to the login page of the configured OIDC identity provider.
func (c *Client) AuthCodeUrlHandler(ctx *gin.Context) {
	handle := rp.AuthURLHandler(rnd.State, c)
	handle(ctx.Writer, ctx.Request)
}

// CodeExchangeUserInfo verifies a redirect auth request and returns the user information and tokens if successful.
func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (userInfo *oidc.UserInfo, tokens *oidc.Tokens[*oidc.IDTokenClaims], err error) {
	getInfo := func(w http.ResponseWriter, r *http.Request, t *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, i *oidc.UserInfo) {
		userInfo = i
		tokens = t
	}

	// It would also be possible to directly get the user info from the oidc.IDTokenClaims
	// without performing a request to the userinfo endpoint of the OIDC identity provider.
	handle := rp.CodeExchangeHandler(rp.UserinfoCallback(getInfo), c)

	handle(ctx.Writer, ctx.Request)

	if sc := ctx.Writer.Status(); sc != 0 && sc != http.StatusOK {
		if oidcErr := ctx.Writer.Header().Get("oidc_error"); oidcErr == "" {
			return userInfo, tokens, errors.New("failed to exchange token for user info")
		} else {
			return userInfo, tokens, errors.New(oidcErr)
		}
	}

	return userInfo, tokens, nil
}
