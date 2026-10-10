package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// resetPortalPeers clears the Portal peer caches before and after a test.
func resetPortalPeers(t *testing.T) {
	t.Helper()
	portalPeers.Flush()
	portalPeersWarned.Flush()
	t.Cleanup(func() {
		portalPeers.Flush()
		portalPeersWarned.Flush()
	})
}

// newPortalPeerContext returns a test context for a request from remoteAddr that trusts no proxy.
func newPortalPeerContext(t *testing.T, remoteAddr, forwardedFor string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, engine := gin.CreateTestContext(httptest.NewRecorder())
	require.NoError(t, engine.SetTrustedProxies(nil))
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/photos", nil)
	c.Request.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		c.Request.Header.Set(header.XForwardedFor, forwardedFor)
	}
	return c
}

func TestRememberPortalPeer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resetPortalPeers(t)
		rememberPortalPeer(newPortalPeerContext(t, "192.0.2.10:12345", "203.0.113.7"))
		_, found := portalPeers.Get("192.0.2.10")
		assert.True(t, found)
		assert.Equal(t, 1, portalPeers.ItemCount())
	})
	t.Run("InvalidPeer", func(t *testing.T) {
		resetPortalPeers(t)
		rememberPortalPeer(nil)
		rememberPortalPeer(newPortalPeerContext(t, "", ""))
		assert.Zero(t, portalPeers.ItemCount())
	})
}

func TestWarnUntrustedPortal(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		resetPortalPeers(t)
		hook := captureSystemLog(t)
		portalPeers.SetDefault("192.0.2.10", struct{}{})

		warnUntrustedPortal(newPortalPeerContext(t, "192.0.2.10:12345", "203.0.113.7"), "192.0.2.10")
		warnUntrustedPortal(newPortalPeerContext(t, "192.0.2.10:12346", "203.0.113.8"), "192.0.2.10")

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Equal(t, "auth: requests forwarded by portal 192.0.2.10 are attributed to its address (see trusted-proxy)", hook.LastEntry().Message)
	})
	t.Run("TrustedPortal", func(t *testing.T) {
		resetPortalPeers(t)
		hook := captureSystemLog(t)
		portalPeers.SetDefault("192.0.2.10", struct{}{})

		warnUntrustedPortal(newPortalPeerContext(t, "192.0.2.10:12345", "203.0.113.7"), "203.0.113.7")
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("NoForwardedAddress", func(t *testing.T) {
		resetPortalPeers(t)
		hook := captureSystemLog(t)
		portalPeers.SetDefault("192.0.2.10", struct{}{})

		warnUntrustedPortal(newPortalPeerContext(t, "192.0.2.10:12345", ""), "192.0.2.10")
		warnUntrustedPortal(nil, "192.0.2.10")
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("OtherPeer", func(t *testing.T) {
		resetPortalPeers(t)
		hook := captureSystemLog(t)
		portalPeers.SetDefault("192.0.2.10", struct{}{})

		warnUntrustedPortal(newPortalPeerContext(t, "198.51.100.20:12345", "203.0.113.7"), "198.51.100.20")
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("AuthAny", func(t *testing.T) {
		resetPortalPeers(t)
		hook := captureSystemLog(t)
		portalPeers.SetDefault("192.0.2.10", struct{}{})

		c := newPortalPeerContext(t, "192.0.2.10:12345", "203.0.113.7")
		require.NotNil(t, AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionView}))
		require.Len(t, hook.AllEntries(), 1)
	})
}
