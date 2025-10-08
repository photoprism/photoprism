package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

func TestAuthAnyJWT(t *testing.T) {
	t.Run("ClusterScope", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-success")
		spec := fx.defaultClaimsSpec()
		token := fx.issue(t, spec)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.10:12345"
		c.Request = req

		session := authAnyJWT(c, "192.0.2.10", token, acl.ResourceCluster, nil)
		require.NotNil(t, session)
		assert.Equal(t, http.StatusOK, session.HttpStatus())
		assert.Equal(t, spec.Subject, session.ClientUID)
		assert.Contains(t, session.AuthScope, "cluster")
		assert.Equal(t, spec.Issuer, session.AuthIssuer)
	})
	t.Run("ClusterCIDRAllowed", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-cidr-allow")
		spec := fx.defaultClaimsSpec()
		token := fx.issue(t, spec)

		origCIDR := fx.nodeConf.Options().ClusterCIDR
		fx.nodeConf.Options().ClusterCIDR = "192.0.2.0/24"
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().ClusterCIDR = origCIDR
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.10:2222"
		c.Request = req

		session := authAnyJWT(c, "192.0.2.10", token, acl.ResourceCluster, nil)
		require.NotNil(t, session)
		assert.Equal(t, spec.Subject, session.ClientUID)
	})
	t.Run("ClusterCIDRBlocked", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-cidr-block")
		spec := fx.defaultClaimsSpec()
		token := fx.issue(t, spec)

		origCIDR := fx.nodeConf.Options().ClusterCIDR
		fx.nodeConf.Options().ClusterCIDR = "192.0.2.0/24"
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().ClusterCIDR = origCIDR
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "203.0.113.10:2222"
		c.Request = req

		assert.Nil(t, authAnyJWT(c, "203.0.113.10", token, acl.ResourceCluster, nil))
	})
	t.Run("JWTScopeDefaultRejectsOtherResources", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-scope-default-reject")
		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{"photos"}
		token := fx.issue(t, spec)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/photos", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.60:1001"
		c.Request = req

		assert.Nil(t, authAnyJWT(c, "192.0.2.60", token, acl.ResourcePhotos, nil))
	})
	t.Run("JWTScopeAllowed", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-scope-allow")
		token := fx.issue(t, fx.defaultClaimsSpec())

		orig := fx.nodeConf.Options().JWTScope
		fx.nodeConf.Options().JWTScope = "cluster vision"
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().JWTScope = orig
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.30:1001"
		c.Request = req

		sess := authAnyJWT(c, "192.0.2.30", token, acl.ResourceCluster, nil)
		require.NotNil(t, sess)
	})
	t.Run("JWTScopeAllowsSuperset", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-scope-reject")
		token := fx.issue(t, fx.defaultClaimsSpec())

		orig := fx.nodeConf.Options().JWTScope
		fx.nodeConf.Options().JWTScope = "cluster"
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().JWTScope = orig
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.40:1001"
		c.Request = req

		sess := authAnyJWT(c, "192.0.2.40", token, acl.ResourceCluster, nil)
		require.NotNil(t, sess)
	})
	t.Run("JWTScopeCustomResource", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-scope-custom")
		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{"photos"}
		token := fx.issue(t, spec)

		origScope := fx.nodeConf.Options().JWTScope
		fx.nodeConf.Options().JWTScope = "photos"
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().JWTScope = origScope
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/photos", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.0.2.50:2001"
		c.Request = req

		_, verifyErr := get.VerifyJWT(c.Request.Context(), token, clusterjwt.ExpectedClaims{
			Issuer:   fmt.Sprintf("portal:%s", fx.clusterUUID),
			Audience: fmt.Sprintf("node:%s", fx.nodeUUID),
			Scope:    []string{"photos"},
			JWKSURL:  fx.nodeConf.JWKSUrl(),
		})
		require.NoError(t, verifyErr)

		sess := authAnyJWT(c, "192.0.2.50", token, acl.ResourcePhotos, nil)
		require.NotNil(t, sess)
	})
	t.Run("VisionScope", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-vision")
		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{"vision"}
		token := fx.issue(t, spec)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/vision/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "198.18.0.5:8080"
		c.Request = req

		session := authAnyJWT(c, "198.18.0.5", token, acl.ResourceVision, nil)
		require.NotNil(t, session)
		assert.Equal(t, http.StatusOK, session.HttpStatus())
		assert.Contains(t, session.AuthScope, "vision")
		assert.Equal(t, spec.Issuer, session.AuthIssuer)
	})
	t.Run("RejectsMalformedOrUnknown", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-invalid")
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-without-dots")
		req.RemoteAddr = "192.0.2.10:12345"
		c.Request = req

		assert.Nil(t, authAnyJWT(c, "192.0.2.10", "invalid-token-without-dots", acl.ResourceCluster, nil))

		// Ensure we also bail out when JWKS URL is not configured.
		fx.nodeConf.SetJWKSUrl("")
		get.SetConfig(fx.nodeConf)
		assert.Nil(t, authAnyJWT(c, "192.0.2.10", "", acl.ResourceCluster, nil))
	})
	t.Run("NoIssuerMatch", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-no-issuer")
		spec := fx.defaultClaimsSpec()
		token := fx.issue(t, spec)

		// Remove all issuer candidates.
		origPortal := fx.nodeConf.Options().PortalUrl
		origSite := fx.nodeConf.Options().SiteUrl
		origClusterUUID := fx.nodeConf.Options().ClusterUUID
		fx.nodeConf.Options().PortalUrl = ""
		fx.nodeConf.Options().SiteUrl = ""
		fx.nodeConf.Options().ClusterUUID = ""
		get.SetConfig(fx.nodeConf)
		t.Cleanup(func() {
			fx.nodeConf.Options().PortalUrl = origPortal
			fx.nodeConf.Options().SiteUrl = origSite
			fx.nodeConf.Options().ClusterUUID = origClusterUUID
			get.SetConfig(fx.nodeConf)
		})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "203.0.113.5:2222"
		c.Request = req

		assert.Nil(t, authAnyJWT(c, "203.0.113.5", token, acl.ResourceCluster, nil))
	})
	t.Run("UnsupportedResource", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "cluster-jwt-unsupported")
		token := fx.issue(t, fx.defaultClaimsSpec())

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/cluster/theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "198.51.100.7:9999"
		c.Request = req

		assert.Nil(t, authAnyJWT(c, "198.51.100.7", token, acl.ResourcePhotos, nil))
	})
}

func TestJwtIssuerCandidates(t *testing.T) {
	t.Run("IncludesAllSources", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ClusterUUID = "11111111-1111-4111-8111-111111111111"
		conf.Options().PortalUrl = "https://portal.example.test/"
		conf.Options().SiteUrl = "https://site.example.test/base/"

		orig := get.Config()
		get.SetConfig(conf)
		t.Cleanup(func() { get.SetConfig(orig) })

		cands := jwtIssuerCandidates(conf)
		assert.Equal(t, []string{
			"portal:11111111-1111-4111-8111-111111111111",
			"https://portal.example.test",
			"https://site.example.test/base",
		}, cands)
	})
	t.Run("DefaultsToLocalhost", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ClusterUUID = ""
		conf.Options().PortalUrl = ""
		conf.Options().SiteUrl = ""

		assert.Equal(t, []string{"http://localhost:2342"}, jwtIssuerCandidates(conf))
	})
}

func TestShouldAttemptJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	c.Request = req

	assert.True(t, shouldAttemptJWT(c, "a.b.c"))
	assert.False(t, shouldAttemptJWT(nil, "a.b.c"))
	assert.False(t, shouldAttemptJWT(c, "invalidtoken"))
	assert.False(t, shouldAttemptJWT(c, ""))
}

func TestNodeAllowsJWT(t *testing.T) {
	fx := newPortalJWTFixture(t, "node-allows")
	conf := fx.nodeConf

	assert.True(t, shouldAllowJWT(conf, "192.0.2.9"))

	origCIDR := conf.Options().ClusterCIDR
	conf.Options().ClusterCIDR = "192.0.2.0/24"
	assert.True(t, shouldAllowJWT(conf, "192.0.2.25"))
	assert.False(t, shouldAllowJWT(conf, "203.0.113.1"))
	conf.Options().ClusterCIDR = origCIDR

	origJWKS := conf.JWKSUrl()
	conf.SetJWKSUrl("")
	assert.False(t, shouldAllowJWT(conf, "192.0.2.25"))
	conf.SetJWKSUrl(origJWKS)

	assert.False(t, shouldAllowJWT(nil, "192.0.2.25"))
}

func TestExpectedClaimsFor(t *testing.T) {
	fx := newPortalJWTFixture(t, "expected-claims")

	claims := expectedClaimsFor(fx.nodeConf, "cluster")
	assert.Equal(t, fmt.Sprintf("node:%s", fx.nodeUUID), claims.Audience)
	assert.Equal(t, []string{"cluster"}, claims.Scope)
	assert.Equal(t, fx.nodeConf.JWKSUrl(), claims.JWKSURL)

	noScope := expectedClaimsFor(fx.nodeConf, "")
	assert.Nil(t, noScope.Scope)
}

func TestVerifyTokenFromPortal(t *testing.T) {
	fx := newPortalJWTFixture(t, "verify-token")
	spec := fx.defaultClaimsSpec()
	token := fx.issue(t, spec)

	expected := expectedClaimsFor(fx.nodeConf, clean.Scope("cluster"))
	claims := verifyTokenFromPortal(context.Background(), token, expected, []string{"wrong", spec.Issuer})
	require.NotNil(t, claims)
	assert.Equal(t, spec.Issuer, claims.Issuer)
	assert.Equal(t, spec.Subject, claims.Subject)

	nilClaims := verifyTokenFromPortal(context.Background(), token, expected, []string{"wrong"})
	assert.Nil(t, nilClaims)
}

func TestSessionFromJWTClaims(t *testing.T) {
	claims := &clusterjwt.Claims{
		Scope: "cluster vision",
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:  "portal:test",
			Subject: "portal:client",
			ID:      "token-id",
		},
	}

	sess := sessionFromJWTClaims(claims, "192.0.2.100")
	require.NotNil(t, sess)
	assert.Equal(t, http.StatusOK, sess.HttpStatus())
	assert.Equal(t, "portal:client", sess.ClientUID)
	assert.Equal(t, clean.Scope("cluster vision"), sess.AuthScope)
	assert.Equal(t, "portal:test", sess.AuthIssuer)
	assert.Equal(t, "token-id", sess.AuthID)
	assert.Equal(t, "192.0.2.100", sess.ClientIP)
}
