package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestOAuthRevoke(t *testing.T) {
	const tokenPath = "/api/v1/oauth/token"
	const revokePath = "/api/v1/oauth/revoke"

	t.Run("ClientSuccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		OAuthRevoke(router)

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(data.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, authToken)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
	t.Run("FormDataAccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		OAuthRevoke(router)

		createData := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(createData.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeData := url.Values{
			"token":           {authToken},
			"token_type_hint": {form.AccessToken},
		}

		revokeToken, _ := http.NewRequest("POST", revokePath, strings.NewReader(revokeData.Encode()))
		revokeToken.Header.Add(header.ContentType, header.ContentTypeForm)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
	t.Run("FormDataSessionId", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		OAuthRevoke(router)

		createData := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(createData.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		sessId := gjson.Get(createResp.Body.String(), "session_id").String()

		revokeData := url.Values{
			"token":           {sessId},
			"token_type_hint": {form.SessionID},
		}

		revokeToken, _ := http.NewRequest("POST", revokePath, strings.NewReader(revokeData.Encode()))
		revokeToken.Header.Add(header.ContentType, header.ContentTypeForm)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusForbidden, revokeResp.Code)
	})
	t.Run("FormDataRefId", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		OAuthRevoke(router)

		createData := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(createData.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		sessId := gjson.Get(createResp.Body.String(), "session_id").String()

		revokeData := url.Values{
			"token":           {rnd.RefID(sessId)},
			"token_type_hint": {form.RefID},
		}

		revokeToken, _ := http.NewRequest("POST", revokePath, strings.NewReader(revokeData.Encode()))
		revokeToken.Header.Add(header.ContentType, header.ContentTypeForm)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusForbidden, revokeResp.Code)
	})
	t.Run("WrongTokenTypeHint", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		OAuthRevoke(router)

		createData := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(createData.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeData := url.Values{
			"token":           {authToken},
			"token_type_hint": {form.SessionID},
		}

		revokeToken, _ := http.NewRequest("POST", revokePath, strings.NewReader(revokeData.Encode()))
		revokeToken.Header.Add(header.ContentType, header.ContentTypeForm)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusUnauthorized, revokeResp.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthRevoke(router)

		sess := entity.SessionFixtures.Get("alice_token")

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, sess.AuthToken())

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusForbidden, revokeResp.Code)
	})
	t.Run("UserSuccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		OAuthToken(router)
		OAuthRevoke(router)

		data := url.Values{
			"grant_type":  {"password"},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(data.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)
		createToken.Header.Add(header.XAuthToken, sessId)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, authToken)
		createToken.Header.Add(header.XAuthToken, sessId)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
}
