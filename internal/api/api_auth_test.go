package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestAuth(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		header.SetAuthorization(c.Request, session.PublicAuthToken)

		// Check auth token.
		authToken := AuthToken(c)
		assert.Equal(t, session.PublicAuthToken, authToken)

		// Check successful authorization in public mode.
		s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.GetUserName())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))

		// Check failed authorization in public mode.
		s = Auth(c, acl.ResourceUsers, acl.ActionUpload)
		assert.NotNil(t, s)
		assert.Equal(t, "", s.GetUserName())
		assert.Equal(t, "", s.ID)
		assert.Equal(t, http.StatusForbidden, s.HttpStatus())
		assert.True(t, s.Abort(c))
	})
}

func TestAuthAny(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		header.SetAuthorization(c.Request, session.PublicAuthToken)

		// Check auth token.
		authToken := AuthToken(c)
		assert.Equal(t, session.PublicAuthToken, authToken)

		// Check successful authorization in public mode.
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionUpdate})
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.GetUserName())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))

		// Check failed authorization in public mode.
		s = AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionUpload})
		assert.NotNil(t, s)
		assert.Equal(t, "", s.GetUserName())
		assert.Equal(t, "", s.ID)
		assert.Equal(t, http.StatusForbidden, s.HttpStatus())
		assert.True(t, s.Abort(c))

		// Check successful authorization with multiple actions in public mode.
		s = AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionUpload, acl.ActionView})
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.GetUserName())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))
	})
}

func TestAuthToken(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// No headers have been set, so no token should be returned.
		token := AuthToken(c)
		assert.Equal(t, "", token)
	})
	t.Run("BearerToken", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		header.SetAuthorization(c.Request, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := header.BearerToken(c)
		assert.Equal(t, authToken, bearerToken)
	})
	t.Run("Header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		c.Request.Header.Add(header.XAuthToken, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := header.BearerToken(c)
		assert.Equal(t, "", bearerToken)
	})
}

func TestAuthAnyPortalJWT(t *testing.T) {
	fx := newPortalJWTFixture(t, "ok")

	spec := fx.defaultClaimsSpec()
	token := fx.issue(t, spec)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.5:1234"
	c.Request = req

	s := AuthAny(c, acl.ResourceCluster, acl.Permissions{acl.ActionView})
	require.NotNil(t, s)
	assert.True(t, s.IsClient())
	assert.Equal(t, http.StatusOK, s.HttpStatus())
	assert.Contains(t, s.AuthScope, "cluster")
	assert.Equal(t, fmt.Sprintf("portal:%s", fx.clusterUUID), s.AuthIssuer)
	assert.Equal(t, "portal:client-test", s.ClientUID)
	assert.False(t, s.Abort(c))

	// Audience mismatch should reject the token once the node UUID changes.
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.RemoteAddr = "10.0.0.5:1234"
	c.Request = req2
	fx.nodeConf.Options().NodeUUID = rnd.UUID()
	get.SetConfig(fx.nodeConf)
	s = AuthAny(c, acl.ResourceCluster, acl.Permissions{acl.ActionView})
	require.NotNil(t, s)
	assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	assert.True(t, s.Abort(c))
}

func TestAuthAnyPortalJWT_MissingScope(t *testing.T) {
	fx := newPortalJWTFixture(t, "missing-scope")
	spec := fx.defaultClaimsSpec()
	spec.Scope = []string{"vision"}
	token := fx.issue(t, spec)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.5:1234"
	c.Request = req

	s := AuthAny(c, acl.ResourceCluster, acl.Permissions{acl.ActionView})
	require.NotNil(t, s)
	assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	assert.True(t, s.Abort(c))
}

func TestAuthAnyPortalJWT_InvalidIssuer(t *testing.T) {
	fx := newPortalJWTFixture(t, "invalid-issuer")
	spec := fx.defaultClaimsSpec()
	spec.Issuer = "https://portal.invalid.test"
	token := fx.issue(t, spec)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.5:1234"
	c.Request = req

	s := AuthAny(c, acl.ResourceCluster, acl.Permissions{acl.ActionView})
	require.NotNil(t, s)
	assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	assert.True(t, s.Abort(c))
}

func TestAuthAnyPortalJWT_NoJWKSConfigured(t *testing.T) {
	fx := newPortalJWTFixture(t, "no-jwks")
	fx.nodeConf.SetJWKSUrl("")
	get.SetConfig(fx.nodeConf)

	spec := fx.defaultClaimsSpec()
	token := fx.issue(t, spec)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.5:1234"
	c.Request = req

	s := AuthAny(c, acl.ResourceCluster, acl.Permissions{acl.ActionView})
	require.NotNil(t, s)
	assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	assert.True(t, s.Abort(c))
}

type portalJWTFixture struct {
	nodeConf    *config.Config
	issuer      *clusterjwt.Issuer
	clusterUUID string
	nodeUUID    string
}

func newPortalJWTFixture(t *testing.T, suffix string) portalJWTFixture {
	t.Helper()

	origConf := get.Config()
	t.Cleanup(func() { get.SetConfig(origConf) })

	nodeConf := config.NewMinimalTestConfigWithDb("auth-any-portal-jwt-"+suffix, t.TempDir())

	nodeConf.Options().NodeRole = cluster.RoleInstance
	nodeConf.Options().Public = false
	clusterUUID := rnd.UUID()
	nodeConf.Options().ClusterUUID = clusterUUID
	nodeUUID := nodeConf.NodeUUID()
	nodeConf.Options().PortalUrl = "https://portal.example.test"

	portalConf := config.NewMinimalTestConfigWithDb("auth-any-portal-jwt-issuer-"+suffix, t.TempDir())

	portalConf.Options().NodeRole = cluster.RolePortal
	portalConf.Options().ClusterUUID = clusterUUID

	mgr, err := clusterjwt.NewManager(portalConf)
	require.NoError(t, err)
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	jwksBytes, err := json.Marshal(mgr.JWKS())
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBytes)
	}))
	t.Cleanup(srv.Close)

	nodeConf.SetJWKSUrl(srv.URL + "/.well-known/jwks.json")
	get.SetConfig(nodeConf)

	return portalJWTFixture{
		nodeConf:    nodeConf,
		issuer:      clusterjwt.NewIssuer(mgr),
		clusterUUID: clusterUUID,
		nodeUUID:    nodeUUID,
	}
}

func (fx portalJWTFixture) defaultClaimsSpec() clusterjwt.ClaimsSpec {
	return clusterjwt.ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", fx.clusterUUID),
		Subject:  "portal:client-test",
		Audience: fmt.Sprintf("node:%s", fx.nodeUUID),
		Scope:    []string{"cluster", "vision"},
	}
}

func (fx portalJWTFixture) issue(t *testing.T, spec clusterjwt.ClaimsSpec) string {
	t.Helper()
	token, err := fx.issuer.Issue(spec)
	require.NoError(t, err)
	return token
}
