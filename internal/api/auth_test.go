package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/httpclient"

	"github.com/stretchr/testify/assert"
)

func TestAuthEndpoints(t *testing.T) {
	t.Run("successful oidc authentication", func(t *testing.T) {
		app, router, _ := NewApiTestWithOIDC()
		AuthEndpoints(router)

		// Step 1a: Request AuthURL
		log.Debug("Requesting OIDC AuthURL...")
		r := PerformRequest(app, http.MethodGet, "/api/v1/auth/external")
		assert.Equal(t, http.StatusFound, r.Code)

		// Step 1b: Redirect user agent to OP and save state cookie
		l := r.Header().Get("Location")
		log.Debug("Requesting AuthCode from OP: ", l)
		cookies := r.Header().Values("Set-Cookie")
		log.Debug("Cookies: ", cookies)
		assert.Contains(t, l, "authorize")

		var l2 string
		cl := httpclient.Client(true)
		cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if strings.Contains(req.URL.String(), "localhost") {

				l2 = req.URL.RequestURI()
				return http.ErrUseLastResponse
			}
			return nil
		}

		_, err := cl.Get(l)
		if err != nil {
			t.Error(err)
		}
		log.Debug(l2)
		log.Debug("Successful")

		// Step 2a: OP redirects user agent back to PhotoPrism
		// Step 2b: PhotoPrism redeems AuthCode and fetches tokens from OP
		log.Debug("Redeem AuthCode...")
		r3 := PerformRequestWithCookie(app, http.MethodGet, l2, strings.Join(cookies, "; "))

		assert.Equal(t, http.StatusOK, r3.Code)
		log.Debug("Successful")
	})

	t.Run("oidc authentication: missing cookie", func(t *testing.T) {
		app, router, _ := NewApiTestWithOIDC()
		AuthEndpoints(router)

		// Step 1a: Request AuthURL
		log.Debug("Requesting OIDC AuthURL...")
		r := PerformRequest(app, http.MethodGet, "/api/v1/auth/external")
		assert.Equal(t, r.Code, http.StatusFound)

		// Step 1b: Redirect user agent to OP and save state cookie
		l := r.Header().Get("Location")
		log.Debug("Requesting AuthCode from OP: ", l)
		cookie := ""
		assert.Contains(t, l, "authorize")
		var l2 string
		cl := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if strings.Contains(req.URL.String(), "localhost") {

					l2 = req.URL.RequestURI()
					return http.ErrUseLastResponse
				}
				return nil
			},
		}
		_, err := cl.Get(l)
		if err != nil {
			t.Error(err)
		}
		log.Debug(l2)
		log.Debug("Successful")

		// Step 2a: OP redirects user agent back to PhotoPrism
		// Step 2b: PhotoPrism redeems AuthCode and fetches tokens from OP
		log.Debug("Redeem AuthCode...")
		r3 := PerformRequestWithCookie(app, http.MethodGet, l2, cookie)

		assert.Equal(t, http.StatusUnauthorized, r3.Code)
		log.Debug("Successful")
	})
}
