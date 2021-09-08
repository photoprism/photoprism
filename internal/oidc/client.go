package oidc

import (
	"errors"
	"github.com/caos/oidc/pkg/client/rp"
	"github.com/caos/oidc/pkg/oidc"
	"github.com/caos/oidc/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/rnd"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var log = event.Log

type Client struct {
	rp.RelyingParty
}

const (
	RedirectPath = "/auth/callback"
	hashKey = "hashkey012345678"
	encrKey = "encrkey012345678"
)

func NewClient(iss *url.URL, clientId, clientSecret, siteUrl string) *Client {
	u, err := url.Parse(siteUrl)
	if err != nil {
		log.Error(err)
	}
	u.Path = path.Join(u.Path, "/api/v1/", RedirectPath)
	log.Debugf(u.String())

	cookieHandler := utils.NewCookieHandler([]byte(hashKey), []byte(encrKey), utils.WithUnsecure())

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}

	scopes := strings.Split("openid profile email", " ")
	//scopes := strings.Split("openid profile email photoprism", " ")

	log.Debugf("Provider Params: %s %s %s %s", iss.String(), clientId, clientSecret, siteUrl)

	provider, err := rp.NewRelyingPartyOIDC(iss.String(), clientId, clientSecret, u.String(), scopes, options...)
	if err != nil {
		log.Errorf("error creating provider: %s", err.Error())
	}
	log.Debug(provider)

	return &Client{
		provider,
	}
}

var state = func() string {
	return rnd.UUID()
}

func (c *Client) AuthUrlHandler() http.HandlerFunc {
	return rp.AuthURLHandler(state, c)
}

var tempstate string

func (c *Client) AuthUrl() string {
	tempstate = state()
	return rp.AuthURL(tempstate, c)
}

func (c *Client) CodeExchangeUserInfo(ctx *gin.Context) (oidc.UserInfo, error) {
	var userinfo oidc.UserInfo

	userinfoClosure := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty, info oidc.UserInfo) {
		log.Infof("UserInfo: %s %s %s %s %s", info.GetEmail(), info.GetSubject(), info.GetNickname(), info.GetName(), info.GetPreferredUsername())
		log.Debugf("IDToken: %s", tokens.IDToken)
		log.Debugf("AToken: %s", tokens.AccessToken)
		log.Debugf("RToken: %s", tokens.RefreshToken)

		userinfo = info
	}

	handle := rp.CodeExchangeHandler(rp.UserinfoCallback(userinfoClosure), c)
	handle(ctx.Writer, ctx.Request)

	log.Debugf("current request state: %v", ctx.Writer.Status())
	if ctx.Writer.Status() != 0 && ctx.Writer.Status() != http.StatusOK {
		return nil, errors.New("oidc: couldn't exchange auth code and thus not retrieve external user info")
	}

	return userinfo, nil
}
