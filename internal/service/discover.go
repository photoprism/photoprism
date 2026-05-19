package service

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/safe"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Account represents remote service details.
type Account struct {
	AccName string
	AccURL  string
	AccType string
	AccKey  string
	AccUser string
	AccPass string
}

// Discover performs a service lookup based on the URL and credentials provided and returns an Account if successful.
func Discover(rawUrl, user, pass, servicesCIDR string) (result Account, err error) {
	if rawUrl == "" {
		return result, errors.New("service URL is empty")
	}

	var allowedCIDRs []*net.IPNet

	if allowedCIDRs, err = ParseCIDRs(servicesCIDR); err != nil {
		return result, err
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
	} else {
		u.Scheme = strings.ToLower(u.Scheme)
	}

	if validatedURL, validateErr := safe.URL(u.String()); validateErr != nil {
		if errors.Is(validateErr, safe.ErrSchemeNotAllowed) {
			return result, errors.New("unsupported service URL scheme")
		}
		return result, validateErr
	} else {
		u = validatedURL
	}

	for _, h := range Heuristics {
		if !h.MatchDomain(u.Host) {
			continue
		}

		if serviceUrl := h.Discover(u.String(), result.AccUser, allowedCIDRs); serviceUrl != nil {
			serviceUrl.User = nil

			if w := txt.Keywords(serviceUrl.Host); len(w) > 0 {
				result.AccName = txt.Title(w[0])
			} else {
				result.AccName = serviceUrl.Host
			}

			result.AccType = h.Type
			result.AccURL = serviceUrl.String()

			return result, nil
		}
	}

	return result, errors.New("could not connect")
}
