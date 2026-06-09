package security

import "testing"

// TestHashPath verifies deterministic hashing and path lookup behavior.
func TestHashPath(t *testing.T) {
	t.Run("Deterministic", func(t *testing.T) {
		const path = "/wp-login.php"
		a := HashPath(path)
		b := HashPath(path)

		if a == 0 {
			t.Fatalf("expected non-zero hash for %q", path)
		}

		if a != b {
			t.Fatalf("expected deterministic hash for %q", path)
		}
	})
	t.Run("IsScanPath", func(t *testing.T) {
		if !IsScanPath("/wp-login.php") {
			t.Fatalf("expected scanner path to be detected")
		}

		if !IsScanPath("/auth.json") {
			t.Fatalf("expected sensitive overlay file probe to be detected")
		}

		if !IsScanPath("/config/options.yml") {
			t.Fatalf("expected sensitive overlay path probe to be detected")
		}

		if !IsScanPath("/.htaccess") {
			t.Fatalf("expected common config probe to be detected")
		}

		if !IsScanPath("/cwclass.php") {
			t.Fatalf("expected web shell probe to be detected")
		}

		if !IsScanPath("/var/task/serverless.yml") {
			t.Fatalf("expected serverless config probe to be detected")
		}

		for _, sensitiveSimplePath := range []string{
			"/node/secrets",
			"/config/portal",
			"/config/certificates",
		} {
			if !IsScanPath(sensitiveSimplePath) {
				t.Fatalf("expected sensitive simple path %q to be detected", sensitiveSimplePath)
			}
		}

		if IsScanPath("/library") {
			t.Fatalf("expected non-scanner path to be ignored")
		}

		for _, safePath := range []string{"/headers", "/health", "/healthz", "/hello"} {
			if IsScanPath(safePath) {
				t.Fatalf("expected safe path %q to be ignored", safePath)
			}
		}

		for _, safeWellKnown := range []string{
			"/.well-known/openid-configuration",
			"/.well-known/jwks",
			"/.well-known/jwks.json",
			"/.well-known/openid-configuration/jwks",
			"/.well-known/openid-configuration/jwks.json",
			"/.well-known/ai-plugin.json",
			"/.well-known/assetlinks.json",
			"/.well-known/acme-challenge/token-abc123",
			"/well-known/apple-app-site-association",
			"/well-known/pki-validation/token.txt",
		} {
			if IsScanPath(safeWellKnown) {
				t.Fatalf("expected safe .well-known path %q to be ignored", safeWellKnown)
			}
		}

		for _, safePath := range []string{
			"/home",
			"/mosts",
			"/html",
			"/hub",
			"/i/",
			"/icons",
			"/image",
			"/img",
			"/index",
			"/info",
			"/jobs",
			"/m",
			"/misc",
			"/plugin",
			"/i/pro-1/library/login",
			"/api/ping",
			"/api/v1/feature",
			"/api/v1/getbaseconfig",
			"/api/v1/info",
			"/api/v1/instance",
			"/library/login",
			"/service/rest/swagger.json",
			"/squid.svg",
			"/video/404.mp4",
			"/font/icon.eot",
			"/font/open-sans/0-255.pbf",
		} {
			if IsScanPath(safePath) {
				t.Fatalf("expected low-signal or proxy path %q to be ignored", safePath)
			}
		}
	})
}
