package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// previewThumbResources are the resources a thumbnail or video request authorizes against.
var previewThumbResources = acl.Resources{acl.ResourcePhotos, acl.ResourceFiles}

// withPreviewAuth enables token validation for the test and restores public mode afterward.
func withPreviewAuth(t *testing.T) *config.Config {
	t.Helper()
	conf := get.Config()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	require.True(t, tokens.Preview.Configured())
	return conf
}

// newPreviewClientSession stores a client without an account with the scope and a session of it, and
// returns both with a signed preview token of the session, minted whether or not it would be delivered.
func newPreviewClientSession(t *testing.T, scope string) (*entity.Client, *entity.Session, string) {
	t.Helper()
	return newPreviewRoleClientSession(t, scope, acl.RoleClient.String())
}

// newPreviewRoleClientSession is newPreviewClientSession for a client with the role.
func newPreviewRoleClientSession(t *testing.T, scope, role string) (*entity.Client, *entity.Session, string) {
	t.Helper()

	client, err := entity.AddClient(form.Client{ClientName: "Preview Client " + clean.Scope(scope) + " " + role, AuthScope: scope, ClientRole: role})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Purge() })

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
	sess := client.NewSession(ctx, authn.GrantClientCredentials)
	require.NoError(t, sess.Save())
	t.Cleanup(func() { _ = sess.Delete() })
	require.True(t, sess.SignedPreview())

	token := tokens.PreviewToken(sess.ID, sess.ExpiresAt())
	require.NotEmpty(t, token)

	return client, sess, token
}

// previewCtx returns a request context with the token path segment and the raw query.
func previewCtx(segment, query string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/t/hash/x/tile_500?"+query, nil)
	c.Params = gin.Params{{Key: "token", Value: segment}}
	return c
}

// changeByte returns the token with its byte at index i replaced by another allowed character.
func changeByte(token string, i int) string {
	b := []byte(token)
	if b[i] == 'a' {
		b[i] = 'b'
	} else {
		b[i] = 'a'
	}
	return string(b)
}

func TestInvalidPreviewToken(t *testing.T) {
	conf := withPreviewAuth(t)
	_, sess, signed := newPreviewClientSession(t, "*")

	alice := entity.NewSession(conf.SessionMaxAge(), 0).SetUser(entity.UserFixtures.Pointer("alice"))
	require.NoError(t, alice.Create())
	t.Cleanup(func() { _ = alice.Delete() })
	userToken := alice.PreviewToken
	require.NotEmpty(t, userToken)

	t.Run("SignedPath", func(t *testing.T) {
		c := previewCtx(signed, "")
		assert.False(t, InvalidPreviewToken(c, previewThumbResources))
		assert.Equal(t, sess.ID, PreviewSession(c).ID)
	})
	t.Run("SignedQuery", func(t *testing.T) {
		assert.False(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, "t="+url.QueryEscape(signed)), previewThumbResources))
	})
	t.Run("UserToken", func(t *testing.T) {
		c := previewCtx(userToken, "")
		assert.False(t, InvalidPreviewToken(c, previewThumbResources))
		_, resolved := c.Get(previewSessionKey)
		assert.False(t, resolved, "no session lookup")
	})
	t.Run("InstanceToken", func(t *testing.T) {
		// The test config was propagated in public mode, so the instance token is registered here.
		entity.PreviewToken.Set("test_instance", conf.PreviewToken())
		t.Cleanup(func() { entity.PreviewToken.Unset("test_instance") })
		assert.False(t, InvalidPreviewToken(previewCtx(conf.PreviewToken(), ""), previewThumbResources))
	})
	t.Run("ChangedByte", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(changeByte(signed, len(signed)-2), ""), previewThumbResources))
		assert.True(t, InvalidPreviewToken(previewCtx(changeByte(signed, strings.LastIndex(signed, ".")-1), ""), previewThumbResources))
	})
	t.Run("DotsRemoved", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(strings.ReplaceAll(signed, ".", ""), ""), previewThumbResources))
	})
	t.Run("Slash", func(t *testing.T) {
		withSlash := signed[:len(signed)-2] + "/" + signed[len(signed)-1:]
		assert.NotContains(t, clean.UrlToken(withSlash), "/")
		assert.True(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, "t="+url.QueryEscape(withSlash)), previewThumbResources))
	})
	t.Run("DownloadToken", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(tokens.SignDownload(sess.ID), ""), previewThumbResources))
	})
	t.Run("OtherSession", func(t *testing.T) {
		_, other, _ := newPreviewClientSession(t, "*")
		parts := strings.SplitN(signed, ".", 3)
		assert.True(t, InvalidPreviewToken(previewCtx(parts[0]+"."+other.ID+"."+parts[2], ""), previewThumbResources))
	})
	t.Run("LeadingZeroExpiry", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx("0"+signed, ""), previewThumbResources))
	})
	t.Run("CoarseDownloadToken", func(t *testing.T) {
		prev := tokens.CoarseDownload
		tokens.CoarseDownload = "coarse-preview-test"
		t.Cleanup(func() { tokens.CoarseDownload = prev })
		assert.False(t, InvalidPreviewToken(previewCtx("coarse-preview-test", ""), previewThumbResources))
		assert.True(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, "t=coarse-preview-test"), previewThumbResources))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(strings.Repeat("a", clean.UrlTokenMaxLength+1), ""), previewThumbResources))
	})
	t.Run("SignedSegmentWithUserToken", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, "t="+userToken), previewThumbResources))
		assert.True(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, "t="+conf.PreviewToken()), previewThumbResources))
	})
	t.Run("SignedSegmentWithoutQuery", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, ""), previewThumbResources))

		// A static token named "signed" keeps working without "?t=".
		entity.PreviewToken.Set("test_signed_static", PreviewSignedSegment)
		t.Cleanup(func() { entity.PreviewToken.Unset("test_signed_static") })
		assert.False(t, InvalidPreviewToken(previewCtx(PreviewSignedSegment, ""), previewThumbResources))
	})
	t.Run("OtherSegmentWithSignedQuery", func(t *testing.T) {
		assert.True(t, InvalidPreviewToken(previewCtx("unknown1", "t="+url.QueryEscape(signed)), previewThumbResources))
	})
	t.Run("EmptySegmentQuery", func(t *testing.T) {
		assert.False(t, InvalidPreviewToken(previewCtx("~~~", "t="+userToken), previewThumbResources))
		assert.True(t, InvalidPreviewToken(previewCtx("~~~", "t="+url.QueryEscape(signed)), previewThumbResources))
	})
	t.Run("PublicMode", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePublic)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePasswd) })
		assert.False(t, InvalidPreviewToken(previewCtx("anything", ""), previewThumbResources))
	})
}

func TestInvalidPreviewToken_SignedRevocation(t *testing.T) {
	withPreviewAuth(t)

	t.Run("ClientDisabled", func(t *testing.T) {
		client, _, signed := newPreviewClientSession(t, "*")
		require.False(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))

		client.AuthEnabled = false
		require.NoError(t, client.SaveChanges("AuthEnabled"))
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
	t.Run("ClientDisabledSessionCached", func(t *testing.T) {
		// A row write leaves the session in the cache; the client check still refuses it.
		client, sess, signed := newPreviewClientSession(t, "*")
		_, err := entity.FindSession(sess.ID)
		require.NoError(t, err)
		require.False(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))

		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("auth_enabled", false).Error)
		entity.FlushClientCache()
		_, err = entity.FindSession(sess.ID)
		require.NoError(t, err)
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
	t.Run("ClientDeleted", func(t *testing.T) {
		client, _, signed := newPreviewClientSession(t, "*")
		require.NoError(t, client.Delete())
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
	t.Run("SessionDeleted", func(t *testing.T) {
		_, sess, signed := newPreviewClientSession(t, "*")
		require.NoError(t, sess.Delete())
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
	t.Run("ScopeWithoutResource", func(t *testing.T) {
		_, sess, _ := newPreviewClientSession(t, "albums")
		signed := tokens.PreviewToken(sess.ID, sess.ExpiresAt())
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
		assert.False(t, InvalidPreviewToken(previewCtx(signed, ""), acl.Resources{acl.ResourceAlbums, acl.ResourcePhotos}))
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), acl.Resources{acl.ResourceLabels, acl.ResourcePhotos}))
	})
	t.Run("AccountSession", func(t *testing.T) {
		// A signed value bound to an account session is not accepted.
		alice := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
		require.NoError(t, alice.Create())
		t.Cleanup(func() { _ = alice.Delete() })
		assert.True(t, InvalidPreviewToken(previewCtx(tokens.PreviewToken(alice.ID, alice.ExpiresAt()), ""), previewThumbResources))
	})
}

func TestPreviewTokens(t *testing.T) {
	_, sess, signed := newPreviewClientSession(t, "*")
	download := tokens.SignDownload(sess.ID)

	t.Run("Static", func(t *testing.T) {
		s, v := previewTokens(previewCtx("abc12345", "t=xyz"))
		assert.Equal(t, "", s)
		assert.Equal(t, "abc12345", v)
	})
	t.Run("SignedPath", func(t *testing.T) {
		s, v := previewTokens(previewCtx(signed, ""))
		assert.Equal(t, signed, s)
		assert.Equal(t, "", v)
	})
	t.Run("SignedQuery", func(t *testing.T) {
		s, v := previewTokens(previewCtx(PreviewSignedSegment, "t="+url.QueryEscape(signed)))
		assert.Equal(t, signed, s)
		assert.Equal(t, "", v)
	})
	t.Run("SignedQueryEmptyAfterCleaning", func(t *testing.T) {
		s, v := previewTokens(previewCtx(PreviewSignedSegment, "t=%2F%2F"))
		assert.Equal(t, "", s)
		assert.Equal(t, "", v)
	})
	t.Run("EmptySegment", func(t *testing.T) {
		s, v := previewTokens(previewCtx("", "t="+url.QueryEscape(download)))
		assert.Equal(t, "", s)
		assert.Equal(t, download, v)
	})
}

func TestPreviewSession(t *testing.T) {
	withPreviewAuth(t)
	_, sess, signed := newPreviewClientSession(t, "*")

	t.Run("Success", func(t *testing.T) {
		c := previewCtx(signed, "")
		got := PreviewSession(c)
		require.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
		assert.Same(t, got, PreviewSession(c))
	})
	t.Run("Static", func(t *testing.T) {
		assert.Nil(t, PreviewSession(previewCtx("abc12345", "")))
	})
	t.Run("Forged", func(t *testing.T) {
		assert.Nil(t, PreviewSession(previewCtx(changeByte(signed, len(signed)-2), "")))
	})
}

func TestPreviewRoleDenied(t *testing.T) {
	t.Run("Client", func(t *testing.T) {
		_, sess, _ := newPreviewRoleClientSession(t, "*", acl.RoleClient.String())
		assert.False(t, previewRoleDenied(sess, previewThumbResources))
	})
	t.Run("Instance", func(t *testing.T) {
		_, sess, _ := newPreviewRoleClientSession(t, "*", acl.RoleInstance.String())
		assert.False(t, previewRoleDenied(sess, previewThumbResources))
	})
	t.Run("None", func(t *testing.T) {
		_, sess, _ := newPreviewRoleClientSession(t, "*", acl.RoleAliasNone)
		require.Equal(t, acl.RoleNone, sess.GetClientRole())
		assert.True(t, previewRoleDenied(sess, previewThumbResources))
		assert.True(t, previewRoleDenied(sess, acl.Resources{acl.ResourceAlbums, acl.ResourcePhotos}))
	})
}

func TestInvalidPreviewToken_Role(t *testing.T) {
	withPreviewAuth(t)

	t.Run("None", func(t *testing.T) {
		_, sess, signed := newPreviewRoleClientSession(t, "*", acl.RoleAliasNone)
		assert.Empty(t, SessionPreviewToken(sess))
		assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
	t.Run("Instance", func(t *testing.T) {
		_, sess, signed := newPreviewRoleClientSession(t, "*", acl.RoleInstance.String())
		assert.Equal(t, signed, SessionPreviewToken(sess))
		assert.False(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
	})
}

func TestPreviewOutOfScope(t *testing.T) {
	t.Run("NoScope", func(t *testing.T) {
		assert.False(t, previewOutOfScope(&entity.Session{}, previewThumbResources))
		assert.False(t, previewOutOfScope(&entity.Session{AuthScope: "*"}, previewThumbResources))
	})
	t.Run("InScope", func(t *testing.T) {
		assert.False(t, previewOutOfScope(&entity.Session{AuthScope: "files"}, previewThumbResources))
		assert.False(t, previewOutOfScope(&entity.Session{AuthScope: "folders"}, acl.Resources{acl.ResourceFolders, acl.ResourcePhotos}))
	})
	t.Run("OutOfScope", func(t *testing.T) {
		assert.True(t, previewOutOfScope(&entity.Session{AuthScope: "metrics"}, previewThumbResources))
		assert.True(t, previewOutOfScope(&entity.Session{AuthScope: "albums"}, acl.Resources{acl.ResourceFolders, acl.ResourcePhotos}))
	})
}

func TestSessionPreviewToken(t *testing.T) {
	withPreviewAuth(t)

	t.Run("Nil", func(t *testing.T) {
		assert.Empty(t, SessionPreviewToken(nil))
	})
	t.Run("Account", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice")
		assert.Equal(t, sess.PreviewToken, SessionPreviewToken(&sess))
	})
	t.Run("SignedPreview", func(t *testing.T) {
		_, sess, signed := newPreviewClientSession(t, "photos")
		assert.Equal(t, signed, SessionPreviewToken(sess))
		_, sid, _, ok := tokens.ParseCompact(SessionPreviewToken(sess))
		require.True(t, ok)
		assert.Equal(t, sess.ID, sid)
	})
	t.Run("StoredRandomToken", func(t *testing.T) {
		_, sess, signed := newPreviewClientSession(t, "*")
		sess.PreviewToken = "oldrnd01"
		assert.Equal(t, signed, SessionPreviewToken(sess))
	})
	t.Run("ScopeWithoutPreviews", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("client_metrics")
		assert.Empty(t, SessionPreviewToken(&sess))
	})
}

func TestSessionContentUri(t *testing.T) {
	conf := get.Config()
	prevCdn, prevVideo := conf.Options().CdnUrl, conf.Options().CdnVideo
	conf.Options().CdnUrl, conf.Options().CdnVideo = "https://cdn.example.com/", true
	t.Cleanup(func() { conf.Options().CdnUrl, conf.Options().CdnVideo = prevCdn, prevVideo })
	require.NotEqual(t, conf.ApiUri(), conf.ContentUri())

	t.Run("Client", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("client_metrics")
		assert.Equal(t, conf.ApiUri(), SessionContentUri(&sess))
	})
	t.Run("Account", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice")
		assert.Equal(t, conf.ContentUri(), SessionContentUri(&sess))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, conf.ContentUri(), SessionContentUri(nil))
	})
}

func TestAddTokenHeaders_SignedPreview(t *testing.T) {
	withPreviewAuth(t)

	t.Run("Signed", func(t *testing.T) {
		_, sess, signed := newPreviewClientSession(t, "*")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, sess)
		assert.Equal(t, signed, w.Header().Get("X-Preview-Token"))
		assert.NotEmpty(t, w.Header().Get("X-Download-Token"))
	})
	t.Run("ScopeWithoutPreviews", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("client_metrics")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, &sess)
		assert.Empty(t, w.Header().Get("X-Preview-Token"))
		assert.Empty(t, w.Header().Get("X-Download-Token"))
	})
	t.Run("UnconfiguredSigner", func(t *testing.T) {
		// Without a preview token, a client session without an account still receives its download token.
		_, sess, _ := newPreviewClientSession(t, "*")
		prev := *tokens.Preview
		tokens.Preview.Key = nil
		t.Cleanup(func() { *tokens.Preview = prev })

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, sess)
		assert.Empty(t, w.Header().Get("X-Preview-Token"))
		assert.NotEmpty(t, w.Header().Get("X-Download-Token"))
	})
	t.Run("Nil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, nil)
		assert.Empty(t, w.Header().Get("X-Preview-Token"))
	})
}

func TestPreviewEndpoints_SignedToken(t *testing.T) {
	withPreviewAuth(t)
	_, _, signed := newPreviewClientSession(t, "*")
	scoped := map[string]string{}

	for _, scope := range []string{"albums", "labels", "folders"} {
		_, _, scoped[scope] = newPreviewClientSession(t, scope)
	}

	// Every path names a hash or UID that does not resolve, so an accepted request answers with a
	// placeholder and changes no fixture, while a refused one answers 403.
	const missingHash = "ffffffffffffffffffffffffffffffffffff0001"

	endpoints := []struct {
		name     string
		register func(*gin.RouterGroup)
		path     func(token string) string
		scope    string
	}{
		{"Thumb", GetThumb, func(token string) string { return "/api/v1/t/" + missingHash + "/" + token + "/tile_500" }, ""},
		{"Video", GetVideo, func(token string) string { return "/api/v1/videos/" + missingHash + "/" + token + "/avc" }, ""},
		{"AlbumCover", AlbumCover, func(token string) string { return "/api/v1/albums/at9lxuqxpogzzz01/t/" + token + "/tile_500" }, "albums"},
		{"LabelCover", LabelCover, func(token string) string { return "/api/v1/labels/lt9lxuqxpogzzz01/t/" + token + "/tile_500" }, "labels"},
		{"FolderCover", FolderCover, func(token string) string { return "/api/v1/folders/t/dt9lxuqxpogzzz01/" + token + "/tile_500" }, "folders"},
	}

	for _, e := range endpoints {
		t.Run(e.name, func(t *testing.T) {
			app, router, _ := NewApiTest()
			e.register(router)

			assert.NotEqual(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, e.path(signed)).Code, "path")
			assert.NotEqual(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, e.path(PreviewSignedSegment)+"?t="+url.QueryEscape(signed)).Code, "query")
			assert.Equal(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, e.path(changeByte(signed, len(signed)-2))).Code, "forged")

			// A session scoped to one collection type loads only the covers of that type.
			for scope, token := range scoped {
				if scope == e.scope {
					assert.NotEqual(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, e.path(token)).Code, scope)
				} else {
					assert.Equal(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, e.path(token)).Code, scope)
				}
			}
		})
	}
}

func TestSearchPhotosView_SignedPreview(t *testing.T) {
	conf := withPreviewAuth(t)
	prevCdn := conf.Options().CdnUrl
	conf.Options().CdnUrl = "https://cdn.example.com/"
	t.Cleanup(func() { conf.Options().CdnUrl = prevCdn })

	_, sess, signed := newPreviewClientSession(t, "*")

	app, router, _ := NewApiTest()
	SearchPhotosView(router)

	r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/view?count=1&merged=true", sess.AuthToken())
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	assert.Equal(t, signed, r.Header().Get("X-Preview-Token"))

	src := gjson.Get(r.Body.String(), "0.Thumbs.fit_720.src").String()
	require.NotEmpty(t, src, r.Body.String())
	assert.True(t, strings.HasPrefix(src, conf.ApiUri()+"/t/"), src)
	assert.Contains(t, src, "/"+signed+"/")
}

func TestSearchGeo_SignedPreview(t *testing.T) {
	conf := withPreviewAuth(t)
	prevCdn := conf.Options().CdnUrl
	conf.Options().CdnUrl = "https://cdn.example.com/"
	t.Cleanup(func() { conf.Options().CdnUrl = prevCdn })

	_, sess, signed := newPreviewClientSession(t, "*")

	app, router, _ := NewApiTest()
	SearchGeo(router)

	r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/geo/view?count=1", sess.AuthToken())
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	src := gjson.Get(r.Body.String(), "0.Thumbs.fit_720.src").String()
	require.NotEmpty(t, src, r.Body.String())
	assert.True(t, strings.HasPrefix(src, conf.ApiUri()+"/t/"), src)
	assert.Contains(t, src, "/"+signed+"/")
}

// countSessionQueries counts the queries of the auth_sessions table until the test ends.
func countSessionQueries(t *testing.T) *atomic.Int64 {
	t.Helper()

	name := "test:count_session_queries_" + rnd.Base36(8)
	count := &atomic.Int64{}

	require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (entity.Session{}).TableName() {
			count.Add(1)
		}
	}))
	t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

	return count
}

// countClientQueries counts the queries of the auth_clients table until the test ends.
func countClientQueries(t *testing.T) *atomic.Int64 {
	t.Helper()

	name := "test:count_client_queries_" + rnd.Base36(8)
	count := &atomic.Int64{}

	require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (entity.Client{}).TableName() {
			count.Add(1)
		}
	}))
	t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

	return count
}

// withAuthLimit replaces limiter.Auth with one that allows burst failures, and restores it after the test.
func withAuthLimit(t *testing.T, burst int) {
	t.Helper()
	prev := limiter.Auth
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), burst)
	t.Cleanup(func() { limiter.Auth = prev })
}

func TestSignedTokens_GoneSession(t *testing.T) {
	withPreviewAuth(t)

	// clientIp is the address previewCtx and downloadCtx requests come from.
	clientIp := ClientIP(previewCtx("", ""))

	t.Run("Preview", func(t *testing.T) {
		withAuthLimit(t, 3)
		_, sess, signed := newPreviewClientSession(t, "*")
		require.NoError(t, sess.Delete())

		queries := countSessionQueries(t)

		// A request whose session is gone costs one session query and does not count towards the limit.
		for i := 0; i < 6; i++ {
			assert.True(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
		}

		assert.Equal(t, int64(6), queries.Load())
		assert.False(t, limiter.Auth.Reject(clientIp))
	})
	t.Run("Download", func(t *testing.T) {
		withAuthLimit(t, 3)
		_, sess, _ := newPreviewClientSession(t, "*")
		token := tokens.SignDownload(sess.ID)
		require.NoError(t, sess.Delete())

		queries := countSessionQueries(t)

		for i := 0; i < 6; i++ {
			assert.Nil(t, DownloadSession(downloadCtx("t="+url.QueryEscape(token))))
		}

		assert.Equal(t, int64(6), queries.Load())
		assert.False(t, limiter.Auth.Reject(clientIp))
	})
	t.Run("ValidNotCharged", func(t *testing.T) {
		withAuthLimit(t, 3)
		_, sess, signed := newPreviewClientSession(t, "*")
		download := tokens.SignDownload(sess.ID)

		for i := 0; i < 3; i++ {
			assert.False(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
			assert.NotNil(t, DownloadSession(downloadCtx("t="+url.QueryEscape(download))))
		}

		assert.False(t, limiter.Auth.Reject(clientIp))
	})
	t.Run("ValidOverLimit", func(t *testing.T) {
		// A valid token is not refused by the limit.
		withAuthLimit(t, 3)
		_, sess, signed := newPreviewClientSession(t, "*")
		for !limiter.Auth.Reject(clientIp) {
			limiter.Auth.Reserve(clientIp)
		}

		assert.False(t, InvalidPreviewToken(previewCtx(signed, ""), previewThumbResources))
		got := DownloadSession(downloadCtx("t=" + url.QueryEscape(tokens.SignDownload(sess.ID))))
		require.NotNil(t, got)
		assert.Equal(t, sess.ID, got.ID)
	})
	t.Run("ForgedNotCharged", func(t *testing.T) {
		withAuthLimit(t, 3)
		_, sess, signed := newPreviewClientSession(t, "*")
		forged := changeByte(signed, len(signed)-2)

		// With the session gone, a lookup of the forged token would query the database.
		require.NoError(t, sess.Delete())
		queries := countSessionQueries(t)

		// A forged token costs no session query and does not count towards the authentication limit.
		for i := 0; i < 5; i++ {
			assert.True(t, InvalidPreviewToken(previewCtx(forged, ""), previewThumbResources))
		}

		assert.Equal(t, int64(0), queries.Load())
		assert.False(t, limiter.Auth.Reject(clientIp))
	})
}
