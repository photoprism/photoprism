package server

import (
	"crypto/tls"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/idna"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/dns"
)

// certificateWarnInterval is the minimum time between warnings about failed certificate requests.
var certificateWarnInterval = time.Minute

// AutoTLS returns the Let's Encrypt certificate manager for the site domain if automatic HTTPS is enabled.
func AutoTLS(conf *config.Config) (*autocert.Manager, error) {
	if !conf.AutoTLS() {
		switch {
		case conf.TLSEmail() == "":
			return nil, errors.New("disabled auto tls")
		case !conf.SiteHttps():
			return nil, errors.New("disabled tls")
		case !dns.IsPublicName(conf.SiteDomain()):
			return nil, errors.New("site domain name required to enable auto tls")
		default:
			return nil, errors.New("disabled tls")
		}
	}

	// Create Let's Encrypt cert manager.
	m := &autocert.Manager{
		Email:      conf.TLSEmail(),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(strings.TrimSuffix(conf.SiteDomain(), ".")),
		Cache:      autocert.DirCache(conf.CertificatesPath()),
	}

	return m, nil
}

// certificateWarner wraps getCertificate so that failed requests for the site domain are reported
// to the system log at most once per interval; other names and challenge handshakes are not reported.
func certificateWarner(getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error), domain string, interval time.Duration) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain = strings.TrimSuffix(domain, ".")

	if ascii, err := idna.Lookup.ToASCII(domain); err == nil {
		domain = ascii
	}

	var mu sync.Mutex
	var last time.Time

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := getCertificate(hello)

		if err == nil || hello == nil || !strings.EqualFold(strings.TrimSuffix(hello.ServerName, "."), domain) {
			return cert, err
		} else if len(hello.SupportedProtos) == 1 && hello.SupportedProtos[0] == acme.ALPNProto {
			return cert, err
		}

		mu.Lock()
		report := last.IsZero() || time.Since(last) >= interval

		if report {
			last = time.Now()
		}

		mu.Unlock()

		if report {
			event.SystemWarn([]string{"server", "auto tls", "failed to obtain certificate for %s", "%s"}, clean.Log(domain), clean.Error(err))
		}

		return cert, err
	}
}
