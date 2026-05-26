package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig_BaseUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.BaseUri(""))
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "", c.BaseUri(""))
	c.options.SiteUrl = "http://foo:2342/foo bar/"
	assert.Equal(t, "/foo%20bar", c.BaseUri(""))
	assert.Equal(t, "/foo%20bar/baz", c.BaseUri("/baz"))
}

func TestConfig_StorageNamespace(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.SiteUrl = "https://example.com/foo/"

	sum := sha256.Sum256([]byte(c.SiteUrl()))
	assert.Equal(t, fmt.Sprintf("%x", sum), c.StorageNamespace())
}

func TestConfig_StaticUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/static", c.StaticUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "/static", c.StaticUri())
	c.options.SiteUrl = "http://foo:2342/foo/"
	assert.Equal(t, "/foo/static", c.StaticUri())
	c.options.CdnUrl = "http://foo:2342/bar"
	assert.Equal(t, "http://foo:2342/bar/foo"+StaticUri, c.StaticUri())
}

func TestConfig_ApiUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.ApiUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.ApiUri())
	c.options.SiteUrl = "http://foo:2342/foo/"
	assert.Equal(t, "/foo"+ApiUri, c.ApiUri())
}

func TestConfig_FrontendUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/library", c.FrontendUri(""))
	assert.Equal(t, "/library/", c.FrontendUri("/"))
	assert.Equal(t, "/library/browse", c.FrontendUri("/browse"))
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "/library", c.FrontendUri(""))
	assert.Equal(t, "/library/", c.FrontendUri("/"))
	assert.Equal(t, "/library/browse", c.FrontendUri("/browse"))
	c.options.SiteUrl = "http://foo:2342/foo/"
	assert.Equal(t, "/foo/library", c.FrontendUri(""))
	assert.Equal(t, "/foo/library/", c.FrontendUri("/"))
	assert.Equal(t, "/foo/library/browse", c.FrontendUri("/browse"))

	c.options.FrontendUri = "/portal/admin/"
	assert.Equal(t, "/foo/portal/admin", c.FrontendUri(""))
	assert.Equal(t, "/foo/portal/admin/", c.FrontendUri("/"))
	assert.Equal(t, "/foo/portal/admin/browse", c.FrontendUri("/browse"))

	c.options.FrontendUri = "  "
	assert.Equal(t, "/foo/library", c.FrontendUri(""))
}

func TestNormalizeFrontendPath(t *testing.T) {
	assert.Equal(t, "", normalizeFrontendPath(""))
	assert.Equal(t, "", normalizeFrontendPath(" "))
	assert.Equal(t, "/library", normalizeFrontendPath("/library"))
	assert.Equal(t, "/portal/admin", normalizeFrontendPath("portal/admin/"))
	assert.Equal(t, "", normalizeFrontendPath("../admin"))
}

func TestConfig_ContentUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.ContentUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.ContentUri())
	c.options.CdnUrl = "http://foo:2342//"
	assert.Equal(t, "http://foo:2342"+ApiUri, c.ContentUri())
}

func TestConfig_VideoUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.VideoUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.VideoUri())
	c.options.CdnUrl = "http://foo:2342//"
	c.options.CdnVideo = true
	assert.Equal(t, "http://foo:2342"+ApiUri, c.VideoUri())
	c.options.CdnVideo = false
	assert.Equal(t, ApiUri, c.VideoUri())
}

func TestConfig_SiteUrl(t *testing.T) {
	c := NewConfig(CliTestContext())

	cases := []struct {
		name string
		in   string
		want string
	}{
		// Baseline behavior.
		{"Default", "", "http://localhost:2342/"},
		{"NoTrailingSlash", "http://superhost", "http://superhost/"},
		{"NonDefaultPort", "http://superhost:2342/", "http://superhost:2342/"},

		// Default-port stripping.
		{"HttpsDefaultPort", "https://app.localssl.dev:443/", "https://app.localssl.dev/"},
		{"HttpDefaultPort", "http://example.com:80/sub", "http://example.com/sub/"},
		{"NonDefaultPortPreserved", "https://example.com:8443/", "https://example.com:8443/"},
		{"MismatchedScheme", "http://example.com:443/", "http://example.com:443/"},

		// Uncommon but well-formed inputs.
		{"IPv6DefaultPort", "https://[::1]:443/", "https://[::1]/"},
		{"IPv6NonDefaultPort", "https://[2001:db8::1]:8443/path", "https://[2001:db8::1]:8443/path/"},
		{"PathPreserved", "https://example.com:443/i/pro-1/", "https://example.com/i/pro-1/"},
		{"QueryStripped", "https://example.com:443/i/pro-1/?lang=de&page=2", "https://example.com/i/pro-1/"},
		{"ForceQueryStripped", "https://example.com/?", "https://example.com/"},
		{"FragmentStripped", "https://example.com/library/#photo123", "https://example.com/library/"},
		{"SurroundingWhitespace", "  https://app.example.com:443/  ", "https://app.example.com/"},
		{"ExtraTrailingSlashes", "https://example.com:443////", "https://example.com/"},
		{"Whitespace", "   ", "http://localhost:2342/"},

		// Unix-socket schemes: port stripping must stay a no-op.
		{"UnixScheme", "unix:///var/run/photoprism.sock", "unix:///var/run/photoprism.sock/"},
		{"HttpUnixScheme", "http+unix:///var/run/photoprism.sock", "http+unix:///var/run/photoprism.sock/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.options.SiteUrl = tc.in
			assert.Equal(t, tc.want, c.SiteUrl())
		})
	}
}

func TestConfig_SiteHttps(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.False(t, c.SiteHttps())
	})
}

func TestConfig_SiteDomain(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, localhost, c.SiteDomain())
	c.options.SiteUrl = "https://foo.bar.com:2342/"
	assert.Equal(t, "foo.bar.com", c.SiteDomain())
	c.options.SiteUrl = ""
	assert.Equal(t, localhost, c.SiteDomain())
}

func TestConfig_SiteHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "localhost:2342", c.SiteHost())
	c.options.SiteUrl = "https://foo.bar.com:2342/"
	assert.Equal(t, "foo.bar.com:2342", c.SiteHost())
	c.options.SiteUrl = ""
	assert.Equal(t, "localhost:2342", c.SiteHost())
}

func TestConfig_SiteFavicon(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "favicon.ico", filepath.Base(c.SiteFavicon()))
	assert.True(t, fs.FileExistsNotEmpty(c.SiteFavicon()))
}

func TestConfig_SitePreview(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "https://i.photoprism.app/prism?cover=64&style=centered%20dark&caption=none&title=PhotoPrism", c.SitePreview())
	c.options.SitePreview = "http://preview.jpg"
	assert.Equal(t, "http://preview.jpg", c.SitePreview())
	c.options.SitePreview = "preview123.jpg"
	assert.Equal(t, "http://localhost:2342/preview123.jpg", c.SitePreview())
	c.options.SitePreview = "foo/preview123.jpg"
	assert.Equal(t, "http://localhost:2342/foo/preview123.jpg", c.SitePreview())
	c.options.SitePreview = "/foo/preview123.jpg"
	assert.Equal(t, "http://localhost:2342/foo/preview123.jpg", c.SitePreview())
}

func TestConfig_SitePreview_StripsDefaultPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	tmp := t.TempDir()
	c.SetThemePath(tmp)
	previewFile := filepath.Join(c.ThemePath(), "preview.jpg")
	if err := os.WriteFile(previewFile, []byte("test"), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	c.options.SiteUrl = "https://host:443/"
	c.options.SitePreview = "preview.jpg"
	assert.Equal(t, "https://host/_theme/preview.jpg", c.SitePreview())
}

func TestConfig_DownloadUrl_StripsDefaultPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.SiteUrl = "https://host:443/"
	assert.Equal(t, "https://host"+DownloadUri, c.DownloadUrl())
	c.options.SiteUrl = "http://host:80/"
	assert.Equal(t, "http://host"+DownloadUri, c.DownloadUrl())
	c.options.SiteUrl = "https://host:8443/"
	assert.Equal(t, "https://host:8443"+DownloadUri, c.DownloadUrl())
}

func TestConfig_SiteAuthor(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteAuthor())
	c.options.SiteAuthor = "@Jens.Mander"
	assert.Equal(t, "@Jens.Mander", c.SiteAuthor())
	c.options.SiteAuthor = ""
	assert.Equal(t, "", c.SiteAuthor())
}

func TestConfig_SiteTitle(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "PhotoPrism", c.SiteTitle())
	c.options.SiteTitle = "Cats"
	assert.Equal(t, "Cats", c.SiteTitle())
	c.options.SiteTitle = "PhotoPrism"
	assert.Equal(t, "PhotoPrism", c.SiteTitle())
}

func TestConfig_SiteCaption(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteCaption())
	c.options.SiteCaption = "PhotoPrism App"
	assert.Equal(t, "PhotoPrism App", c.SiteCaption())
	c.options.SiteCaption = ""
	assert.Equal(t, "", c.SiteCaption())
}

func TestConfig_SiteDescription(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteDescription())
	c.options.SiteDescription = "My Description!"
	assert.Equal(t, "My Description!", c.SiteDescription())
	c.options.SiteDescription = ""
	assert.Equal(t, "", c.SiteDescription())
}

func TestConfig_LegalInfo(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.LegalInfo())
	assert.Equal(t, "", c.LegalUrl())
	c.options.LegalInfo = "ACME Inc."
	c.options.LegalUrl = "https://example.com/"
	assert.Equal(t, c.options.LegalInfo, c.LegalInfo())
	assert.Equal(t, c.options.LegalUrl, c.LegalUrl())
	c.options.LegalInfo = ""
	c.options.LegalUrl = ""
	assert.Equal(t, "", c.LegalInfo())
	assert.Equal(t, "", c.LegalUrl())
}

func TestConfig_RobotsTxt(t *testing.T) {
	c := NewConfig(CliTestContext())

	result, err := c.RobotsTxt()

	assert.NoError(t, err)
	assert.Equal(t, robotsTxt, result)
}
