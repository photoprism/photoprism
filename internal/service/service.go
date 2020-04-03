/*
Package service implements a remote service abstraction.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package service

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

var log = event.Log
var client = &http.Client{}

const (
	TypeWeb          = "web"
	TypeWebDAV       = "webdav"
	TypeFacebook     = "facebook"
	TypeTwitter      = "twitter"
	TypeFlickr       = "flickr"
	TypeInstagram    = "instagram"
	TypeEyeEm        = "eyeem"
	TypeTelegram     = "telegram"
	TypeWhatsApp     = "whatsapp"
	TypeGooglePhotos = "gphotos"
	TypeGoogleDrive  = "gdrive"
	TypeOneDrive     = "onedrive"
)

type Account struct {
	AccName string
	AccURL  string
	AccType string
	AccKey  string
	AccUser string
	AccPass string
}

type Heuristic struct {
	ServiceType string
	Domains     []string
	Paths       []string
	Method      string
}

var Heuristics = []Heuristic{
	{TypeFacebook, []string{"facebook.com", "www.facebook.com"}, []string{}, "GET"},
	{TypeTwitter, []string{"twitter.com"}, []string{}, "GET"},
	{TypeFlickr, []string{"flickr.com", "www.flickr.com"}, []string{}, "GET"},
	{TypeInstagram, []string{"instagram.com", "www.instagram.com"}, []string{}, "GET"},
	{TypeEyeEm, []string{"eyeem.com", "www.eyeem.com"}, []string{}, "GET"},
	{TypeTelegram, []string{"web.telegram.org", "www.telegram.org", "telegram.org"}, []string{}, "GET"},
	{TypeWhatsApp, []string{"web.whatsapp.com", "www.whatsapp.com", "whatsapp.com"}, []string{}, "GET"},
	{TypeOneDrive, []string{"onedrive.live.com"}, []string{}, "GET"},
	{TypeGoogleDrive, []string{"drive.google.com"}, []string{}, "GET"},
	{TypeGooglePhotos, []string{"photos.google.com"}, []string{}, "GET"},
	{TypeWebDAV, []string{}, []string{"/", "/webdav", "/remote.php/dav/files/{user}", "/remote.php/webdav", "/dav/files/{user}", "/servlet/webdav.infostore/"}, "PROPFIND"},
	{TypeWeb, []string{}, []string{}, "GET"},
}

func HttpOk(method, rawUrl string) bool {
	req, err := http.NewRequest(method, rawUrl, nil)

	if err != nil {
		return false
	}

	if resp, err := client.Do(req); err != nil {
		return false
	} else if resp.StatusCode < 400 {
		return true
	}

	return false
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
		strings.Replace(p, "{user}", user, -1)
		u.Path = p

		if HttpOk(h.Method, u.String()) {
			return u
		}
	}

	return nil
}

func Discover(rawUrl, user, pass string) (result Account, err error) {
	if rawUrl == "" {
		return result, errors.New("service URL is empty")
	}

	u, err := url.Parse(rawUrl)

	if err != nil {
		return result, err
	}

	u.Host = strings.ToLower(u.Host)

	result.AccUser = u.User.Username()
	result.AccPass, _ = u.User.Password()

	// Extract user info
	if user != "" {
		result.AccUser = user
	}

	if pass != "" {
		result.AccPass = pass
	}

	if user != "" || pass != "" {
		u.User = url.UserPassword(result.AccUser, result.AccPass)
	}

	// Set default scheme
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	for _, h := range Heuristics {
		if !h.MatchDomain(u.Host) {
			continue
		}

		if serviceUrl := h.Discover(u.String(), result.AccUser); serviceUrl != nil {
			serviceUrl.User = nil

			if w := txt.Keywords(serviceUrl.Host); len(w) > 0 {
				result.AccName = strings.Title(w[0])
			} else {
				result.AccName = serviceUrl.Host
			}

			result.AccType = string(h.ServiceType)
			result.AccURL = serviceUrl.String()

			return result, nil
		}
	}

	return result, errors.New("could not connect")
}
