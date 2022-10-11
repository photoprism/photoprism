package server

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/acme/autocert"

	"github.com/photoprism/photoprism/internal/config"
)

// AutoTLS enables automatic HTTPS via Let's Encrypt.
func AutoTLS(conf *config.Config) (*autocert.Manager, error) {
	var siteDomain, tlsEmail, certDir string

	// Enable automatic HTTPS via Let's Encrypt?
	if !conf.SiteHttps() {
		return nil, fmt.Errorf("default site url does not use https")
	} else if siteDomain = conf.SiteDomain(); !strings.Contains(siteDomain, ".") {
		return nil, fmt.Errorf("no fully qualified site domain")
	} else if tlsEmail = conf.AutoTLS(); tlsEmail == "" {
		return nil, fmt.Errorf("automatic tls disabled")
	} else if certDir = conf.CertsConfigPath(); certDir == "" {
		return nil, fmt.Errorf("https certificate cache directory is missing")
	}

	// Create Let's Encrypt cert manager.
	m := &autocert.Manager{
		Email:      tlsEmail,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(siteDomain),
		Cache:      autocert.DirCache(certDir),
	}

	return m, nil
}
