package oidc

import (
	"errors"
	"fmt"
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

const (
	RedirectPath  = "/auth/callback"
	OidcRoleClaim = "photoprism_role"
	OidcAdminRole = "photoprism_admin"
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
			options = append(options, rp.WithPKCE(cookieHandler))
		}
	}

	scopes := strings.Split(strings.TrimSpace("openid profile email "+customScopes), " ")

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
		log.Debugf("oidc: UserInfo: %s %s %s %s %s", info.GetEmail(), info.GetSubject(), info.GetNickname(), info.GetName(), info.GetPreferredUsername())

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
		err := ctx.Writer.Header().Get("oidc_error")
		if err == "" {
			return nil, errors.New("oidc: couldn't exchange auth code and thus not retrieve external user info (unknown error)")
		}
		return nil, errors.New(ctx.Writer.Header().Get("oidc_error"))
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

func UsernameFromUserInfo(userinfo oidc.UserInfo) (uname string) {
	if len(userinfo.GetPreferredUsername()) >= 4 {
		uname = userinfo.GetPreferredUsername()
	} else if len(userinfo.GetNickname()) >= 4 {
		uname = userinfo.GetNickname()
	} else if len(userinfo.GetName()) >= 4 {
		uname = strings.ReplaceAll(strings.ToLower(userinfo.GetName()), " ", "-")
	} else if len(userinfo.GetEmail()) >= 4 {
		uname = userinfo.GetEmail()
	} else {
		log.Error("oidc: no username found")
	}
	return uname
}

// HasRoleAdmin searches UserInfo claims for admin role.
// Returns true if role is present or false if claim was found but no role in there.
// Error will be returned if the role claim is not delivered at all.
func HasRoleAdmin(userinfo oidc.UserInfo) (bool, error) {
	claim := userinfo.GetClaim(OidcRoleClaim)
	return claimContainsProp(claim, OidcAdminRole)
}

func claimContainsProp(claim interface{}, property string) (bool, error) {
	switch t := claim.(type) {
	case nil:
		return false, errors.New("oidc: claim not found")
	case []interface{}:
		for _, value := range t {
			res, err := claimContainsProp(value, property)
			if err != nil {
				return false, err
			}
			if res {
				return res, nil
			}
		}
		return false, nil
	case interface{}:
		if value, ok := t.(string); ok {
			return value == property, nil
		} else {
			return false, errors.New("oidc: unexpected type")
		}
	//case string:
	//	return t == property, nil
	default:
		return false, errors.New("oidc: unexpected type")
	}
}
