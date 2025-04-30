package service

import (
	"net/url"
	"slices"
	"strings"
)

type Headers = map[string]string

// Heuristic represents a heuristic for detecting a remote service type, e.g. WebDAV.
type Heuristic struct {
	Type    Type
	Domains []string
	Paths   []string
	Method  string
	Headers Headers
}

// Heuristics for common remote service types.
var Heuristics = []Heuristic{
	{Type: Facebook, Domains: []string{"facebook.com", "www.facebook.com"}, Paths: []string{}, Method: "GET"},
	{Type: Twitter, Domains: []string{"twitter.com"}, Paths: []string{}, Method: "GET"},
	{Type: Flickr, Domains: []string{"flickr.com", "www.flickr.com"}, Paths: []string{}, Method: "GET"},
	{Type: Instagram, Domains: []string{"instagram.com", "www.instagram.com"}, Paths: []string{}, Method: "GET"},
	{Type: Telegram, Domains: []string{"web.telegram.org", "www.telegram.org", "telegram.org"}, Paths: []string{}, Method: "GET"},
	{Type: WhatsApp, Domains: []string{"web.whatsapp.com", "www.whatsapp.com", "whatsapp.com"}, Paths: []string{}, Method: "GET"},
	{Type: OneDrive, Domains: []string{"onedrive.live.com"}, Paths: []string{}, Method: "GET"},
	{Type: GDrive, Domains: []string{"drive.google.com"}, Paths: []string{}, Method: "GET"},
	{Type: GPhotos, Domains: []string{"photos.google.com"}, Paths: []string{}, Method: "GET"},
	{Type: WebDAV,
		Domains: []string{},
		Paths:   []string{"/", "/webdav/", "/originals/", "/remote.php/dav/files/{user}/", "/remote.php/webdav/", "/dav/files/{user}/", "/servlet/webdav.infostore/"},
		Method:  "PROPFIND",
		Headers: Headers{"Depth": "1"},
	},
}

func (h Heuristic) MatchDomain(match string) bool {
	if len(h.Domains) == 0 {
		return true
	}

	return slices.Contains(h.Domains, match)
}

func (h Heuristic) Discover(rawUrl, user string) *url.URL {
	u, err := url.Parse(rawUrl)

	if err != nil {
		return nil
	}

	if h.TestRequest(h.Method, u.String()) {
		return u
	}

	for _, p := range h.Paths {
		u.Path = strings.Replace(p, "{user}", user, -1)

		if h.TestRequest(h.Method, u.String()) {
			return u
		}
	}

	return nil
}
