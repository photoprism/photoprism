package header

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/list"
)

// Content Delivery Network (CDN) headers.
const (
	CdnHost         = "Cdn-Host"
	CdnMobileDevice = "Cdn-Mobiledevice"
	CdnServerZone   = "Cdn-Serverzone"
	CdnServerID     = "Cdn-Serverid"
	CdnConnectionID = "Cdn-Connectionid"
)

// IsCdn checks whether the request seems to come from a CDN.
func IsCdn(req *http.Request) bool {
	if req == nil {
		return false
	} else if req.Header == nil || req.URL == nil {
		return false
	}

	if req.Header.Get(CdnHost) != "" {
		return true
	}

	return false
}

// BlockCdn checks whether the request should be blocked for CDNs.
func BlockCdn(req *http.Request) bool {
	if !IsCdn(req) {
		return false
	}

	if req.URL.Path == "/" {
		return true
	}

	return list.Excludes(SafeMethods, req.Method)
}
