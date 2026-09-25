package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestOAuthToken_TokenLimit(t *testing.T) {
	const tokenPath = "/api/v1/oauth/token" // #nosec G101 test constant, not a credential
	const revokePath = "/api/v1/oauth/revoke"

	// Session IDs that outrank any generated one, so the token minted below is retained
	// only because it was the one this request issued.
	highIDs := []string{strings.Repeat("f", 64), strings.Repeat("e", 64)}

	t.Run("KeepsTheIssuedToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		dropIsolatedClientSessions(t)

		client := entity.FindClientByUID(isolatedClientID)

		if client == nil {
			t.Fatal("client fixture not found")
		}

		// Fill the client's token budget so the request below has to evict a session, and
		// stamp the fillers with the second the request will run in, since sessions created
		// in the same second are the case the token's own creation must not evict it.
		second := entity.Now().Add(time.Second)

		for _, id := range highIDs {
			sess := client.NewSession(nil, "client_credentials")
			sess.ID = id

			if err := sess.Create(); err != nil {
				t.Fatal(err)
			} else if err = sess.Updates(entity.Values{"created_at": second}); err != nil {
				t.Fatal(err)
			}
		}

		time.Sleep(time.Until(second))

		OAuthToken(router)
		OAuthRevoke(router)

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {isolatedClientID},
			"client_secret": {isolatedClientSecret},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(data.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()
		assert.NotEmpty(t, authToken)

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, authToken)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
}
