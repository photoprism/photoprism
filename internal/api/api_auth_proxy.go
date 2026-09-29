package api

import (
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// portalPeers holds the addresses from which a Portal-signed JWT was accepted.
var portalPeers = gc.New(time.Hour, 10*time.Minute)

// portalPeersWarned holds the Portal addresses an untrusted proxy warning was logged for.
var portalPeersWarned = gc.New(24*time.Hour, time.Hour)

// rememberPortalPeer records the TCP peer of a request with a verified Portal-signed JWT.
func rememberPortalPeer(c *gin.Context) {
	if c == nil || c.Request == nil {
		return
	}

	if peer := header.IP(c.RemoteIP(), header.UnknownIP); peer != header.UnknownIP {
		portalPeers.SetDefault(peer, struct{}{})
	}
}

// warnUntrustedPortal logs a system warning, at most once a day per address, if a request
// forwarded by a Portal is attributed to the Portal's own address.
func warnUntrustedPortal(c *gin.Context, clientIP string) {
	if c == nil || c.Request == nil || c.Request.Header.Get(header.XForwardedFor) == "" {
		return
	}

	peer := header.IP(c.RemoteIP(), header.UnknownIP)

	if peer == header.UnknownIP || peer != clientIP {
		return
	} else if _, found := portalPeers.Get(peer); !found {
		return
	} else if portalPeersWarned.Add(peer, struct{}{}, gc.DefaultExpiration) != nil {
		return
	}

	event.SystemWarn([]string{"auth", "requests forwarded by portal %s are attributed to its address (see trusted-proxy)"}, peer)
}
