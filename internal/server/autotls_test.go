package server

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/acme"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

func TestAutoTLS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com/"
		conf.Options().TLSEmail = "admin@example.com"

		m, err := AutoTLS(conf)
		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "admin@example.com", m.Email)
		assert.NoError(t, m.HostPolicy(context.Background(), "photos.example.com"))
		assert.Error(t, m.HostPolicy(context.Background(), "other.example.com"))
		assert.NotNil(t, m.Cache)
	})
	t.Run("IDN", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://bücher.example.org/"
		conf.Options().TLSEmail = "admin@example.com"

		m, err := AutoTLS(conf)
		require.NoError(t, err)
		assert.NoError(t, m.HostPolicy(context.Background(), "xn--bcher-kva.example.org"))
		assert.Error(t, m.HostPolicy(context.Background(), "bucher.example.org"))
	})
	t.Run("TrailingDot", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com./"
		conf.Options().TLSEmail = "admin@example.com"

		m, err := AutoTLS(conf)
		require.NoError(t, err)
		assert.NoError(t, m.HostPolicy(context.Background(), "photos.example.com"))
	})
	t.Run("Disabled", func(t *testing.T) {
		for name, tc := range map[string]struct {
			siteUrl, email string
			disable        bool
			want           string
		}{
			"NoEmail":     {siteUrl: "https://photos.example.com/", want: "disabled auto tls"},
			"HttpSiteUrl": {siteUrl: "http://photos.example.com/", email: "admin@example.com", want: "disabled tls"},
			"HttpIP":      {siteUrl: "http://192.0.2.1:2342/", email: "admin@example.com", want: "disabled tls"},
			"IPAddress":   {siteUrl: "https://192.0.2.1/", email: "admin@example.com", want: "site domain name required to enable auto tls"},
			"Localhost":   {siteUrl: "https://localhost:2342/", email: "admin@example.com", want: "site domain name required to enable auto tls"},
			"DisableTLS":  {siteUrl: "https://photos.example.com/", email: "admin@example.com", disable: true, want: "disabled tls"},
		} {
			t.Run(name, func(t *testing.T) {
				conf := config.NewConfig(config.CliTestContext())
				conf.Options().SiteUrl = tc.siteUrl
				conf.Options().TLSEmail = tc.email
				conf.Options().DisableTLS = tc.disable

				m, err := AutoTLS(conf)
				assert.Nil(t, m)
				assert.EqualError(t, err, tc.want)
			})
		}
	})
}

// captureSystemLog replaces the system logger for the duration of a test.
func captureSystemLog(t *testing.T) *logtest.Hook {
	t.Helper()

	orig := event.SystemLog
	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	t.Cleanup(func() { event.SystemLog = orig })

	return hook
}

func TestCertificateWarner(t *testing.T) {
	failing := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return nil, errors.New("acme/autocert: missing certificate")
	}

	t.Run("ReportsSiteDomainOncePerInterval", func(t *testing.T) {
		hook := captureSystemLog(t)
		getCertificate := certificateWarner(failing, "photos.example.com", time.Hour)

		_, err := getCertificate(&tls.ClientHelloInfo{ServerName: "Photos.Example.com"})
		assert.Error(t, err)
		_, _ = getCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com"})

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "failed to obtain certificate for photos.example.com")
		assert.Contains(t, hook.LastEntry().Message, "missing certificate")
	})
	t.Run("ReportsAgainAfterInterval", func(t *testing.T) {
		hook := captureSystemLog(t)
		getCertificate := certificateWarner(failing, "photos.example.com", 0)

		_, _ = getCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com"})
		_, _ = getCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com"})

		assert.Len(t, hook.AllEntries(), 2)
	})
	t.Run("IDN", func(t *testing.T) {
		hook := captureSystemLog(t)
		getCertificate := certificateWarner(failing, "bücher.example.org.", time.Hour)

		_, _ = getCertificate(&tls.ClientHelloInfo{ServerName: "xn--bcher-kva.example.org"})

		assert.Len(t, hook.AllEntries(), 1)
	})
	t.Run("IgnoresOtherNamesAndChallenges", func(t *testing.T) {
		hook := captureSystemLog(t)
		getCertificate := certificateWarner(failing, "photos.example.com", 0)

		_, _ = getCertificate(&tls.ClientHelloInfo{ServerName: "other.example.com"})
		_, _ = getCertificate(&tls.ClientHelloInfo{})
		_, _ = getCertificate(nil)
		_, err := getCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com", SupportedProtos: []string{acme.ALPNProto}})
		assert.Error(t, err)

		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Success", func(t *testing.T) {
		hook := captureSystemLog(t)
		want := &tls.Certificate{}
		getCertificate := certificateWarner(func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return want, nil }, "photos.example.com", 0)

		cert, err := getCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com"})
		assert.NoError(t, err)
		assert.Same(t, want, cert)
		assert.Empty(t, hook.AllEntries())
	})
}
