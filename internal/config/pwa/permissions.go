package pwa

import (
	"fmt"
	"net/url"

	"github.com/photoprism/photoprism/pkg/list"
)

// Permissions specifies the default web app manifest permissions:
// - https://web.dev/learn/pwa/capabilities
// - https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/manifest.json/permissions
var Permissions = list.List{
	"geolocation",
	"downloads",
	"storage",
	"background",
	"webNavigation",
	"webRequest",
	"clipboardWrite",
}

// OptionalPermissions specifies the optional web app manifest permissions,
// see https://developer.chrome.com/docs/extensions/reference/api/permissions.
var OptionalPermissions = list.List{
	"nativeMessaging",
	"notifications",
}

// HostPermissions returns the URLs for which the app is requesting extra privileges,
// see https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/manifest.json/permissions#host_permissions.
func HostPermissions(siteUrl, cdnUrl string) (hosts []string) {
	if siteUrl != "" {
		uri, err := url.Parse(siteUrl)
		if err != nil {
			// Skip.
		} else if hostName := uri.Hostname(); hostName != "" {
			hosts = append(hosts, fmt.Sprintf("*://%s/*", hostName))
		}
	}

	if cdnUrl != "" && cdnUrl != siteUrl {
		uri, err := url.Parse(cdnUrl)
		if err != nil {
			// Skip.
		} else if hostName := uri.Hostname(); hostName != "" {
			hosts = append(hosts, fmt.Sprintf("*://%s/*", hostName))
		}
	}

	return hosts
}
