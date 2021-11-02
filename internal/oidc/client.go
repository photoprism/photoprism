package oidc

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/caos/oidc/pkg/client"
	"github.com/caos/oidc/pkg/client/rp"
	"github.com/caos/oidc/pkg/oidc"
	"github.com/caos/oidc/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/httpclient"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const RedirectPath = "/auth/callback"

var log = event.Log

type Client struct {
	rp.RelyingParty
	debug bool
}

func NewClient(iss *url.URL, clientId, clientSecret, siteUrl string, debug bool) (result *Client, err error) {
	log.Debugf("oidc: Provider Params: %s %s %s %s", iss.String(), clientId, clientSecret, siteUrl)

	u, err := url.Parse(siteUrl)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/v1/", RedirectPath)
	log.Debugf("oidc: %s", u.String())

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
	httpClient := httpclient.Client(debug)

	options := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(
			rp.WithIssuedAtOffset(5 * time.Second),
		),
	}

	discover, err := client.Discover(iss.String(), httpClient)

	if err != nil {
		log.Errorf("oidc: %q (discover)", err)
		return nil, err
	}

	for _, v := range discover.CodeChallengeMethodsSupported {
		if v == oidc.CodeChallengeMethodS256 {
			options = append(options, rp.WithPKCE(cookieHandler))
		}
	}

	scopes := strings.Split("openid profile email", " ")

	provider, err := rp.NewRelyingPartyOIDC(iss.String(), clientId, clientSecret, u.String(), scopes, options...)

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

func (c *Client) AuthUrlHandler() http.HandlerFunc {
	return rp.AuthURLHandler(state, c)
}

func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (oidc.UserInfo, error) {
	var userinfo oidc.UserInfo

	userinfoClosure := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty, info oidc.UserInfo) {
		log.Infof("oidc: UserInfo: %s %s %s %s %s", info.GetEmail(), info.GetSubject(), info.GetNickname(), info.GetName(), info.GetPreferredUsername())
		log.Debugf("oidc: IDToken: %s", tokens.IDToken)
		log.Debugf("oidc: AToken: %s", tokens.AccessToken)
		log.Debugf("oidc: RToken: %s", tokens.RefreshToken)

		userinfo = info
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

	log.Debugf("oidc: current request state: %v", ctx.Writer.Status())
	if sc := ctx.Writer.Status(); sc != 0 && sc != http.StatusOK {
		return nil, errors.New("oidc: couldn't exchange auth code and thus not retrieve external user info")
	}

	return userinfo, nil
}

func (c *Client) IsAvailable() error {
	if c == nil {
		return errors.New("oidc: not initialized")
	}
	_, err := client.Discover(c.Issuer(), httpclient.Client(c.debug))
	if err != nil {
		return err
	}
	return nil
}
