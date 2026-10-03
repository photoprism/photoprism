package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
)

// downloadEndpoint describes one signed-token download route for the scope regression table.
type downloadEndpoint struct {
	name     string
	register func(*gin.RouterGroup)
	path     string
}

// downloadEndpoints returns the four routes that authorize through AuthDownload. Each path names an
// entity that does not resolve, so a request the scope admits ends in 404 and a refused one in 403.
func downloadEndpoints() []downloadEndpoint {
	return []downloadEndpoint{
		{"PhotoDownload", GetPhotoDownload, "/api/v1/photos/pt9jtdre2lvl0y11/dl"},
		{"FileDownload", GetDownload, "/api/v1/dl/3cad9168fa6acc5c5c2965ddf6ec465ca42fd818"},
		{"AlbumDownload", DownloadAlbum, "/api/v1/albums/at9lxuqxpogaaba7/dl"},
		{"ZipDownload", ZipDownload, "/api/v1/zip/nonexistent.zip"},
	}
}

func TestAuthDownloadScope(t *testing.T) {
	// A scoped session can hold a signed download token, so every download endpoint applies the session
	// scope to its own resource. A wildcard scope is unaffected.
	refused := entity.SessionFixtures.Get("alice_app_password_shares")
	wildcard := entity.SessionFixtures.Get("gandalf_app_password_full_access")

	t.Run("SignedTokensResolveBothFixtures", func(t *testing.T) {
		// Pins the signer: without this a broken SignDownload would turn every refusal below into a
		// forged-token 403 that proves nothing about the scope gate.
		conf := get.Config()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		for _, s := range []entity.Session{refused, wildcard} {
			got := DownloadSession(downloadCtx("t=" + tokens.SignDownload(s.ID)))
			if assert.NotNil(t, got) {
				assert.Equal(t, s.ID, got.ID)
			}
		}
	})

	for _, e := range downloadEndpoints() {
		t.Run(e.name+"RefusesOutOfScopeSession", func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			defer conf.SetAuthMode(config.AuthModePublic)
			e.register(router)
			r := PerformRequest(app, http.MethodGet, e.path+"?t="+tokens.SignDownload(refused.ID))
			assert.Equal(t, http.StatusForbidden, r.Code)
		})
		t.Run(e.name+"AdmitsWildcardScope", func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			defer conf.SetAuthMode(config.AuthModePublic)
			e.register(router)
			r := PerformRequest(app, http.MethodGet, e.path+"?t="+tokens.SignDownload(wildcard.ID))
			assert.Equal(t, http.StatusNotFound, r.Code)
		})
	}
}

func TestAuthDownloadScopeServesTheSameOriginal(t *testing.T) {
	// One account and one original, so the only difference between the two requests below is the
	// scope: the refusal cannot be a missing file or a role the account does not hold.
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	conf.Options().OriginalsPath = t.TempDir()
	conf.Propagate()
	GetPhotoDownload(router)

	const uid = "ps6sg6be2lvl0y13"
	file, err := query.FileByPhotoUID(uid)
	require.NoError(t, err)
	original := CreateTestOriginal(t, file)

	wildcard := entity.SessionFixtures.Get("alice_app_password_full_access")
	granted := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+tokens.SignDownload(wildcard.ID))
	require.Equal(t, http.StatusOK, granted.Code)
	assert.Equal(t, original, granted.Body.Bytes())

	refused := entity.SessionFixtures.Get("alice_app_password_shares")
	denied := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+tokens.SignDownload(refused.ID))
	assert.Equal(t, http.StatusForbidden, denied.Code)
}

func TestAuthDownloadScopeUnchangedBranches(t *testing.T) {
	// The branches that carry no session scope must not be refused.
	for _, e := range downloadEndpoints() {
		t.Run(e.name+"PublicMode", func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePublic)
			e.register(router)
			r := PerformRequest(app, http.MethodGet, e.path+"?t="+conf.DownloadToken())
			assert.NotEqual(t, http.StatusForbidden, r.Code)
		})
		t.Run(e.name+"CoarseToken", func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			defer conf.SetAuthMode(config.AuthModePublic)
			orig := tokens.CoarseDownload
			tokens.CoarseDownload = "coarse-scope-token"
			defer func() { tokens.CoarseDownload = orig }()
			e.register(router)
			r := PerformRequest(app, http.MethodGet, e.path+"?t=coarse-scope-token")
			assert.NotEqual(t, http.StatusForbidden, r.Code)
		})
		t.Run(e.name+"ShareLinkVisitor", func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			defer conf.SetAuthMode(config.AuthModePublic)
			visitor := entity.SessionFixtures.Get("visitor")
			// A share-link session inherits its scope from the visitor account, which is unrestricted.
			assert.False(t, visitor.SetUser(&entity.Visitor).HasScope())
			e.register(router)
			r := PerformRequest(app, http.MethodGet, e.path+"?t="+tokens.SignDownload(visitor.ID))
			assert.NotEqual(t, http.StatusForbidden, r.Code)
		})
	}
}

func TestDownloadOutOfScope(t *testing.T) {
	pictures := acl.Resources{acl.ResourcePhotos}
	originals := acl.Resources{acl.ResourceFiles, acl.ResourcePhotos}

	t.Run("ScopedSessionRefusedForForeignResource", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice_app_password_shares")
		assert.True(t, downloadOutOfScope(downloadCtx(""), &sess, pictures))
		assert.True(t, downloadOutOfScope(downloadCtx(""), &sess, originals))
	})
	t.Run("WildcardScopeAdmitted", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("gandalf_app_password_full_access")
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, pictures))
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, originals))
	})
	t.Run("EmptyScopeAdmitted", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("visitor")
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, pictures))
	})
	t.Run("PhotoScopeAdmitsBoth", func(t *testing.T) {
		// The by-hash endpoint accepts either resource, so a picture scope reaches it.
		sess := entity.SessionFixtures.Get("alice_token_scope")
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, pictures))
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, originals))
	})
	t.Run("FileScopeAdmitsOriginalsOnly", func(t *testing.T) {
		sess := entity.Session{AuthScope: "files"}
		assert.False(t, downloadOutOfScope(downloadCtx(""), &sess, originals))
		assert.True(t, downloadOutOfScope(downloadCtx(""), &sess, pictures))
	})
	t.Run("HeaderAuthorizedSessionExempt", func(t *testing.T) {
		// A cluster JWT reaches AuthDownload only after authAnyJWT matched its scope against full file
		// access, so its narrow scope must not be checked a second time against another resource.
		sess := entity.Session{AuthScope: "cluster vision", GrantType: authn.GrantJwtBearer.String()}
		c := downloadCtx("")
		assert.True(t, downloadOutOfScope(c, &sess, pictures))
		c.Set(downloadHeaderAuthKey, true)
		assert.False(t, downloadOutOfScope(c, &sess, pictures))
	})
}

func TestSessionDownloadToken(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Empty(t, SessionDownloadToken(nil))
	})
	t.Run("OutOfScope", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice_app_password_shares")
		assert.Empty(t, SessionDownloadToken(&sess))
	})
	t.Run("WildcardScope", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("gandalf_app_password_full_access")
		assert.NotEmpty(t, SessionDownloadToken(&sess))
	})
}

func TestAddTokenHeadersScope(t *testing.T) {
	conf := get.Config()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	newCtx := func() (*httptest.ResponseRecorder, *gin.Context) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/albums", nil)
		return w, c
	}

	t.Run("OutOfScopeSessionGetsNoDownloadToken", func(t *testing.T) {
		w, c := newCtx()
		sess := entity.SessionFixtures.Get("alice_app_password_shares")
		AddTokenHeaders(c, &sess)
		assert.Equal(t, sess.PreviewToken, w.Header().Get("X-Preview-Token"))
		assert.Empty(t, w.Header().Get("X-Download-Token"))
	})
	t.Run("WildcardScopeGetsDownloadToken", func(t *testing.T) {
		w, c := newCtx()
		sess := entity.SessionFixtures.Get("gandalf_app_password_full_access")
		AddTokenHeaders(c, &sess)
		assert.NotEmpty(t, w.Header().Get("X-Download-Token"))
	})
}
