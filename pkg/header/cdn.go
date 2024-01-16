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

var (
	CdnMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
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

// AbortCdnRequest checks if the request should not be served through a CDN.
func AbortCdnRequest(req *http.Request) bool {
	if !IsCdn(req) {
		return false
	} else if req.Header.Get(XAuthToken) != "" {
		return true
	} else if req.URL.Path == "/" {
		return true
	}

	return list.Excludes(CdnMethods, req.Method)
}
