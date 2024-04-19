package remote

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
	{ServiceFacebook, []string{"facebook.com", "www.facebook.com"}, []string{}, "GET"},
	{ServiceTwitter, []string{"twitter.com"}, []string{}, "GET"},
	{ServiceFlickr, []string{"flickr.com", "www.flickr.com"}, []string{}, "GET"},
	{ServiceInstagram, []string{"instagram.com", "www.instagram.com"}, []string{}, "GET"},
	{ServiceEyeEm, []string{"eyeem.com", "www.eyeem.com"}, []string{}, "GET"},
	{ServiceTelegram, []string{"web.telegram.org", "www.telegram.org", "telegram.org"}, []string{}, "GET"},
	{ServiceWhatsApp, []string{"web.whatsapp.com", "www.whatsapp.com", "whatsapp.com"}, []string{}, "GET"},
	{ServiceOneDrive, []string{"onedrive.live.com"}, []string{}, "GET"},
	{ServiceGDrive, []string{"drive.google.com"}, []string{}, "GET"},
	{ServiceGPhotos, []string{"photos.google.com"}, []string{}, "GET"},
	{ServiceWebDAV, []string{}, []string{"/", "/webdav/", "/originals/", "/remote.php/dav/files/{user}/", "/remote.php/webdav/", "/dav/files/{user}/", "/servlet/webdav.infostore/"}, "PROPFIND"},
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
