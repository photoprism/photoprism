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
}

func NewClient(iss *url.URL, clientId, clientSecret, siteUrl string, debug bool) *Client {
	log.Debugf("Provider Params: %s %s %s %s", iss.String(), clientId, clientSecret, siteUrl)

	u, err := url.Parse(siteUrl)
	if err != nil {
		log.Error(err)
	}
	u.Path = path.Join(u.Path, "/api/v1/", RedirectPath)
	log.Debugf(u.String())

	hashKey, err := rnd.RandomBytes(16)
	encryptKey, err := rnd.RandomBytes(16)
	if err != nil {
		log.Errorf("oidc intialization: %q", err)
		return nil
	}

	cookieHandler := utils.NewCookieHandler(hashKey, encryptKey, utils.WithUnsecure())
	httpClient := httpclient.Client(debug)

	options := []rp.Option{
		rp.WithHTTPClient(httpClient),
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}

	discover, err := client.Discover(iss.String(), httpClient)
	if err != nil {
		log.Errorf("oidc intialization: %q", err)
		return nil
	}
	for _, v := range discover.CodeChallengeMethodsSupported {
		if v == oidc.CodeChallengeMethodS256 {
			//options = append(options, rp.WithPKCE(cookieHandler))
		}
	}

	scopes := strings.Split("openid profile email", " ")
	//scopes := strings.Split("openid profile email photoprism", " ")

	provider, err := rp.NewRelyingPartyOIDC(iss.String(), clientId, clientSecret, u.String(), scopes, options...)
	if err != nil {
		log.Errorf("oidc intialization: %s", err)
		return nil
	}
	log.Debugf("PKCE enabled: %v", provider.IsPKCE())

	return &Client{
		provider,
	}
}

func state() string {
	return rnd.UUID()
}

func (c *Client) AuthUrlHandler() http.HandlerFunc {
	return rp.AuthURLHandler(state, c)
}

//var tempstate string
//
//func (c *Client) AuthUrl() string {
//	tempstate = state()
//	return rp.AuthURL(tempstate, c)
//}

//func (c *Client) Available() bool {
//	return c.RelyingParty != nil
//}

func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (oidc.UserInfo, error) {
	var userinfo oidc.UserInfo

	userinfoClosure := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty, info oidc.UserInfo) {
		log.Infof("UserInfo: %s %s %s %s %s", info.GetEmail(), info.GetSubject(), info.GetNickname(), info.GetName(), info.GetPreferredUsername())
		log.Debugf("IDToken: %s", tokens.IDToken)
		log.Debugf("AToken: %s", tokens.AccessToken)
		log.Debugf("RToken: %s", tokens.RefreshToken)

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

	log.Debugf("current request state: %v", ctx.Writer.Status())
	if sc := ctx.Writer.Status(); sc != 0 && sc != http.StatusOK {
		return nil, errors.New("oidc: couldn't exchange auth code and thus not retrieve external user info")
	}

	return userinfo, nil
}
