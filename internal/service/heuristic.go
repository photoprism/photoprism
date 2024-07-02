package service

import (
	"net/url"
	"strings"
)

type Heuristic struct {
	ServiceType string
	Domains     []string
	Paths       []string
	Method      string
}

var Heuristics = []Heuristic{
	{Facebook, []string{"facebook.com", "www.facebook.com"}, []string{}, "GET"},
	{Twitter, []string{"twitter.com"}, []string{}, "GET"},
	{Flickr, []string{"flickr.com", "www.flickr.com"}, []string{}, "GET"},
	{Instagram, []string{"instagram.com", "www.instagram.com"}, []string{}, "GET"},
	{Telegram, []string{"web.telegram.org", "www.telegram.org", "telegram.org"}, []string{}, "GET"},
	{WhatsApp, []string{"web.whatsapp.com", "www.whatsapp.com", "whatsapp.com"}, []string{}, "GET"},
	{OneDrive, []string{"onedrive.live.com"}, []string{}, "GET"},
	{GDrive, []string{"drive.google.com"}, []string{}, "GET"},
	{GPhotos, []string{"photos.google.com"}, []string{}, "GET"},
	{WebDAV, []string{}, []string{"/", "/webdav/", "/originals/", "/remote.php/dav/files/{user}/", "/remote.php/webdav/", "/dav/files/{user}/", "/servlet/webdav.infostore/"}, "PROPFIND"},
}

func (h Heuristic) MatchDomain(match string) bool {
	if len(h.Domains) == 0 {
		return true
	}

	for _, m := range h.Domains {
		if m == match {
			return true
		}
	}

	return false
}

func (h Heuristic) Discover(rawUrl, user string) *url.URL {
	u, err := url.Parse(rawUrl)

	if err != nil {
		return nil
	}

	if HttpOk(h.Method, u.String()) {
		return u
	}

	for _, p := range h.Paths {
		u.Path = strings.Replace(p, "{user}", user, -1)

		if HttpOk(h.Method, u.String()) {
			return u
		}
	}

	return nil
}
