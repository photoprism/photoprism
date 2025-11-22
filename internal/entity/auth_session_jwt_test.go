package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewSessionFromJWT(t *testing.T) {
	issuedAt := time.Date(2025, time.October, 28, 9, 0, 0, 0, time.UTC)
	notBefore := issuedAt.Add(15 * time.Second)
	expiresAt := issuedAt.Add(5 * time.Minute)
	claims := &JWT{
		Token:     "eyJhbGciOiJFZERTQSIsImtpZCI6Imp3dDVjcHUxN242Z2oyIn0.eyJpc3MiOiJwb3J0YWw6Y2JhYTAyNzYtMDdkMy00M2FjLWI0MjAtMjVlMjYwMWIwYWQ0Iiwic3ViIjoicG9ydGFsOmNzNWNwdTE3bjZnajJxbzUiLCJzY29wZSI6ImNsdXN0ZXIgdmlzaW9uIiwianRpIjoiand0NWNwdTE3bjZnajIifQ.64yKsi6ixGmE3j_BL_WckqHQzXHp7018mCVDciGHXxyXcDL4kZPVJg4hKWdAl95IcC-fL5sTl9p2TBQnpSGeDg",
		ID:        "jwt5cpu17n6gj2",
		Issuer:    "portal:cbaa0276-07d3-43ac-b420-25e2601b0ad4",
		Subject:   "portal:cs5cpu17n6gj2qo5",
		Scope:     "cluster vision",
		IssuedAt:  &issuedAt,
		NotBefore: &notBefore,
		ExpiresAt: &expiresAt,
	}

	expectedIP := "192.0.2.100"
	expectedAgent := "PhotoPrism Portal (1.2510.29)"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = expectedIP + ":12345"
	req.Header.Set(header.UserAgent, expectedAgent)
	c.Request = req

	sess := NewSessionFromJWT(c, claims)

	require.NotNil(t, sess)
	assert.Equal(t, http.StatusOK, sess.HttpStatus())
	assert.True(t, rnd.IsSessionID(sess.ID))
	assert.Equal(t, authn.GrantJwtBearer.String(), sess.GrantType)
	assert.Equal(t, authn.MethodJWT.String(), sess.AuthMethod)
	assert.Equal(t, authn.ProviderAccessToken.String(), sess.AuthProvider)
	assert.Equal(t, "portal:cs5cpu17n6gj2qo5", sess.GetClientName())
	assert.Equal(t, clean.Scope("cluster vision"), sess.AuthScope)
	assert.Equal(t, "portal:cbaa0276-07d3-43ac-b420-25e2601b0ad4", sess.AuthIssuer)
	assert.Equal(t, claims.ID, sess.AuthID)
	assert.Equal(t, claims.ID, sess.RefID)
	assert.True(t, rnd.IsRefID(sess.RefID))
	assert.Equal(t, expectedIP, sess.ClientIP)
	assert.Equal(t, expectedAgent, sess.UserAgent)
	assert.Equal(t, claims.Token, sess.AuthToken())
	assert.True(t, sess.GetUser().IsUnknown())
	assert.Equal(t, issuedAt, sess.CreatedAt)
	assert.Equal(t, notBefore, sess.UpdatedAt)
	assert.Equal(t, notBefore.Unix(), sess.LastActive)
	assert.Equal(t, expiresAt.Unix(), sess.SessExpires)
}
