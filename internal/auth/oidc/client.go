package oidc

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zitadel/oidc/pkg/client"
	"github.com/zitadel/oidc/pkg/client/rp"
	utils "github.com/zitadel/oidc/pkg/http"
	"github.com/zitadel/oidc/pkg/oidc"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	RoleClaim = "photoprism_role"
	AdminRole = "photoprism_admin"
)

var log = event.Log

type Client struct {
	rp.RelyingParty
	debug bool
}

func NewClient(iss *url.URL, clientId, clientSecret, customScopes, siteUrl string, debug bool) (result *Client, err error) {
	log.Debugf("oidc: Provider Params: %s %s %s %s", iss.String(), clientId, clientSecret, siteUrl)

	u, err := url.Parse(siteUrl)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	u.Path = path.Join(u.Path, config.OidcRedirectUri)
	log.Debugf("oidc: redirect uri %s", u.String())

	var hashKey, encryptKey []byte

	if hashKey, err = rnd.RandomBytes(16); err != nil {
		log.Errorf("oidc: %q (create hash key)", err)
		return nil, err
	}

	if encryptKey, err = rnd.RandomBytes(16); err != nil {
		log.Errorf("oidc: %q (create encrypt key)", err)
		return nil, err
	}

	cookieHandler := utils.NewCookieHandler(hashKey, encryptKey, utils.WithUnsecure())
	httpClient := HttpClient(debug)

	clientOpt := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(
			rp.WithIssuedAtOffset(5 * time.Second),
		),
		rp.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, errorType string, errorDesc string, state string) {
			log.Errorf("oidc: %s: %s (state: %s)", errorType, errorDesc, state)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("oidc_error", fmt.Sprintf("oidc: %s", errorDesc))
		}),
	}

	discover, err := client.Discover(iss.String(), httpClient)

	if err != nil {
		log.Errorf("oidc: %q (discover)", err)
		return nil, err
	}

	for _, v := range discover.CodeChallengeMethodsSupported {
		if v == oidc.CodeChallengeMethodS256 {
			clientOpt = append(clientOpt, rp.WithPKCE(cookieHandler))
		}
	}

	scopes := strings.Split(strings.TrimSpace("openid email profile "+customScopes), " ")

	provider, err := rp.NewRelyingPartyOIDC(iss.String(), clientId, clientSecret, u.String(), scopes, clientOpt...)

	if err != nil {
		log.Errorf("oidc: %s (issuer)", err)
		return nil, err
	}

	log.Debugf("oidc: PKCE enabled %v", provider.IsPKCE())

	return &Client{
		provider,
		debug,
	}, nil
}

func state() string {
	return rnd.UUID()
}

func (c *Client) AuthCodeUrlHandler(ctx *gin.Context) {
	handle := rp.AuthURLHandler(state, c)
	handle(ctx.Writer, ctx.Request)
}

func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (userInfo oidc.UserInfo, tokens *oidc.Tokens, err error) {
	userinfoClosure := func(w http.ResponseWriter, r *http.Request, t *oidc.Tokens, state string, rp rp.RelyingParty, i oidc.UserInfo) {
		userInfo = i
		tokens = t
	}

	//you could also just take the access_token and id_token without calling the userinfo endpoint:
	//
	//tokeninfoClosure := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty) {
	//	log.Infof("IDTOKEN: %q\n\n" , tokens.IDToken)
	//	log.Infof("ACCESSTOKEN: %q\n\n" , tokens.AccessToken)
	//	log.Infof("REFRESHTOKEN: %q\n\n" , tokens.RefreshToken)
	//}

	handle := rp.CodeExchangeHandler(rp.UserinfoCallback(userinfoClosure), c)
	//handle := rp.CodeExchangeHandler(tokeninfoClosure, c)
	handle(ctx.Writer, ctx.Request)

	// log.Debugf("oidc: current request state: %v", ctx.Writer.Status())
	if sc := ctx.Writer.Status(); sc != 0 && sc != http.StatusOK {
		if oidcErr := ctx.Writer.Header().Get("oidc_error"); oidcErr == "" {
			return userInfo, tokens, errors.New("tailed to exchange the authentication code and retrieve the user information")
		} else {
			return userInfo, tokens, errors.New(oidcErr)
		}
	}

	return userInfo, tokens, nil
}
