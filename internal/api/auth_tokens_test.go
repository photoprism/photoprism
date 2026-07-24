package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// downloadCtx builds a gin context for a GET request carrying the given query string.
func downloadCtx(query string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/dl/x?"+query, nil)
	return c
}

func TestDownloadSession(t *testing.T) {
	conf := get.Config()

	t.Run("PublicModeReturnsPublicSession", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePublic)
		assert.NotNil(t, DownloadSession(downloadCtx("t=whatever")))
	})

	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	t.Run("SignedTokenResolvesSession", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice")
		got := DownloadSession(downloadCtx("t=" + tokens.SignDownload(sess.ID)))
		assert.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
	})
	t.Run("VerboseSignedTokenResolvesSession", func(t *testing.T) {
		// The CDN-facing verbose form (token=…&expires=…&sid=…) signs the same message as the compact
		// "?t=" value, so a compact token split into query params resolves to the same session.
		sess := entity.SessionFixtures.Get("alice")
		parts := strings.SplitN(tokens.SignDownload(sess.ID), ".", 3)
		q := "token=" + parts[2] + "&expires=" + parts[0] + "&sid=" + parts[1]
		got := DownloadSession(downloadCtx(q))
		assert.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
	})
	t.Run("ForgedSignedTokenReturnsNil", func(t *testing.T) {
		assert.Nil(t, DownloadSession(downloadCtx("t=1784883674.ad041bd1d789b2926104c07bc481bd6dec898650351b2b4d9269223db960d4bc.HS256-forgedsignaturevalue")))
	})
	t.Run("CoarseOrUnknownTokenReturnsNil", func(t *testing.T) {
		// A coarse (static/instance) token is not session-bound, so it resolves to no session.
		orig := tokens.CoarseDownload
		tokens.CoarseDownload = "coarse-instance-token"
		defer func() { tokens.CoarseDownload = orig }()
		assert.Nil(t, DownloadSession(downloadCtx("t=coarse-instance-token")))
		assert.Nil(t, DownloadSession(downloadCtx("t=totally-unknown")))
	})
	t.Run("NonJwtHeaderNotAcceptedForDownload", func(t *testing.T) {
		// Header auth on downloads is restricted to cluster JWTs; a regular session bearer/X-Auth-Token
		// does not resolve here (it must use a "?t=" token). With no "?t=" the request has no session.
		sess := entity.SessionFixtures.Get("alice")
		c := downloadCtx("")
		c.Request.Header.Set("X-Auth-Token", sess.AuthToken())
		assert.Nil(t, DownloadSession(c))
	})
	t.Run("NonJwtHeaderFallsBackToQueryToken", func(t *testing.T) {
		// A non-JWT header must not shadow a valid "?t=" token — the token still resolves the session.
		sess := entity.SessionFixtures.Get("alice")
		c := downloadCtx("t=" + tokens.SignDownload(sess.ID))
		c.Request.Header.Set("X-Auth-Token", sess.AuthToken())
		got := DownloadSession(c)
		assert.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
	})
	t.Run("BasicAuthHeaderFallsBackToQueryToken", func(t *testing.T) {
		// A request behind a basic-auth reverse proxy carries "Authorization: Basic …" on every request.
		// That is not a bearer token, so it must not route to header auth — the signed "?t=" token still
		// resolves the session.
		sess := entity.SessionFixtures.Get("alice")
		c := downloadCtx("t=" + tokens.SignDownload(sess.ID))
		c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")
		got := DownloadSession(c)
		assert.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
	})
}

func TestVerifyDownloadParams(t *testing.T) {
	sess := entity.SessionFixtures.Get("alice")
	parts := strings.SplitN(tokens.SignDownload(sess.ID), ".", 3)

	t.Run("Valid", func(t *testing.T) {
		id, ok := verifyDownloadParams(parts[0], parts[1], parts[2])
		assert.True(t, ok)
		assert.Equal(t, sess.ID, id)
	})
	t.Run("NonNumericExpires", func(t *testing.T) {
		id, ok := verifyDownloadParams("not-a-number", parts[1], parts[2])
		assert.False(t, ok)
		assert.Empty(t, id)
	})
	t.Run("ForgedToken", func(t *testing.T) {
		id, ok := verifyDownloadParams(parts[0], parts[1], "HS256-forgedsignaturevalue")
		assert.False(t, ok)
		assert.Empty(t, id)
	})
}

func TestInvalidDownloadToken(t *testing.T) {
	conf := get.Config()

	t.Run("PublicModeAcceptsAnyToken", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePublic)
		assert.False(t, InvalidDownloadToken(downloadCtx("t=whatever")))
	})

	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	t.Run("SignedTokenAccepted", func(t *testing.T) {
		v := tokens.SignDownload(entity.SessionFixtures.Get("alice").ID)
		assert.False(t, InvalidDownloadToken(downloadCtx("t="+v)))
	})
	t.Run("CoarseTokenAccepted", func(t *testing.T) {
		// The coarse instance token is not session-bound but stays valid, so static-token links keep working.
		orig := tokens.CoarseDownload
		tokens.CoarseDownload = "coarse-xyz"
		defer func() { tokens.CoarseDownload = orig }()
		assert.False(t, InvalidDownloadToken(downloadCtx("t=coarse-xyz")))
	})
	t.Run("UnknownTokenRejected", func(t *testing.T) {
		orig := tokens.CoarseDownload
		tokens.CoarseDownload = "coarse-xyz"
		defer func() { tokens.CoarseDownload = orig }()
		assert.True(t, InvalidDownloadToken(downloadCtx("t=totally-unknown")))
	})
}

func TestAuthDownload(t *testing.T) {
	conf := get.Config()

	t.Run("PublicModeValidWithSession", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePublic)
		sess, valid := AuthDownload(downloadCtx("t=anything"))
		assert.True(t, valid)
		assert.NotNil(t, sess)
	})

	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	t.Run("SignedTokenReturnsSession", func(t *testing.T) {
		s := entity.SessionFixtures.Get("alice")
		sess, valid := AuthDownload(downloadCtx("t=" + tokens.SignDownload(s.ID)))
		assert.True(t, valid)
		if assert.NotNil(t, sess) {
			assert.Equal(t, s.ID, sess.ID)
		}
	})
	t.Run("CoarseTokenValidNoSession", func(t *testing.T) {
		orig := tokens.CoarseDownload
		tokens.CoarseDownload = "coarse-abc"
		defer func() { tokens.CoarseDownload = orig }()
		sess, valid := AuthDownload(downloadCtx("t=coarse-abc"))
		assert.True(t, valid)
		assert.Nil(t, sess)
	})
	t.Run("UnknownTokenInvalid", func(t *testing.T) {
		sess, valid := AuthDownload(downloadCtx("t=nope"))
		assert.False(t, valid)
		assert.Nil(t, sess)
	})
}
