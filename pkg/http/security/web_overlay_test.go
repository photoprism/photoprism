package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestOverlayHasAmbiguousPath(t *testing.T) {
	tests := []struct {
		name        string
		requestPath string
		escapedPath string
		blocked     bool
	}{
		{name: "RootAllowed", requestPath: "/", escapedPath: "/", blocked: false},
		{name: "DirectoryAllowed", requestPath: "/docs/", escapedPath: "/docs/", blocked: false},
		{name: "DoubleSlashBlocked", requestPath: "/docs//index.html", escapedPath: "/docs//index.html", blocked: true},
		{name: "TraversalBlocked", requestPath: "/docs/../env", escapedPath: "/docs/../env", blocked: true},
		{name: "HiddenSegmentBlocked", requestPath: "/.env", escapedPath: "/.env", blocked: true},
		{name: "SpecialSegmentBlocked", requestPath: "/@secrets.txt", escapedPath: "/@secrets.txt", blocked: true},
		{name: "EncodedDotProbeBlocked", requestPath: "/%2eenv", escapedPath: "/%2eenv", blocked: true},
		{name: "EncodedSlashProbeBlocked", requestPath: "/foo%2fbar", escapedPath: "/foo%2fbar", blocked: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.blocked, OverlayHasAmbiguousPath(tc.requestPath, tc.escapedPath))
		})
	}
}

func TestOverlayRelativePath(t *testing.T) {
	t.Run("NoBasePath", func(t *testing.T) {
		relPath, ok := OverlayRelativePath("/assets/app.js", "")
		assert.True(t, ok)
		assert.Equal(t, "assets/app.js", relPath)
	})
	t.Run("BasePathRoot", func(t *testing.T) {
		relPath, ok := OverlayRelativePath("/i/acme", "i/acme")
		assert.True(t, ok)
		assert.Equal(t, "", relPath)
	})
	t.Run("BasePathAsset", func(t *testing.T) {
		relPath, ok := OverlayRelativePath("/i/acme/assets/app.js", "i/acme")
		assert.True(t, ok)
		assert.Equal(t, "assets/app.js", relPath)
	})
	t.Run("OutsideBasePath", func(t *testing.T) {
		relPath, ok := OverlayRelativePath("/assets/app.js", "i/acme")
		assert.False(t, ok)
		assert.Equal(t, "", relPath)
	})
}

func TestOverlayPathBlocked(t *testing.T) {
	tests := []struct {
		name    string
		webPath string
		blocked bool
	}{
		{name: "PublicFileAllowed", webPath: "public.txt", blocked: false},
		{name: "NestedPublicFileAllowed", webPath: "docs/index.html", blocked: false},
		{name: "HiddenFileBlocked", webPath: ".env", blocked: true},
		{name: "NestedHiddenFileBlocked", webPath: "foo/.env", blocked: true},
		{name: "SpecialFileBlocked", webPath: "@secret.txt", blocked: true},
		{name: "SpecialDirBlocked", webPath: "__MACOSX/test.txt", blocked: true},
		{name: "BlockedByName", webPath: "options.yml", blocked: true},
		{name: "BlockedByNameCaseInsensitive", webPath: "Options.YML", blocked: true},
		{name: "BlockedByNameAuthJson", webPath: "auth.json", blocked: true},
		{name: "BlockedByNameJoinToken", webPath: "join_token", blocked: true},
		{name: "BlockedByExtensionPem", webPath: "tls/server.pem", blocked: true},
		{name: "BlockedByExtensionToml", webPath: "docs/public.toml", blocked: true},
		{name: "BlockedByExtensionSQL", webPath: "backup/database.sql", blocked: true},
		{name: "BlockedByPrefix", webPath: "node/secrets/token.txt", blocked: true},
		{name: "BlockedByPrefixConfigPortal", webPath: "config/portal/options.yml", blocked: true},
		{name: "BlockedByPrefixCertificates", webPath: "config/certificates/fullchain.pem", blocked: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.blocked, OverlayPathBlocked(tc.webPath))
		})
	}
}

func TestOverlayResolveFile(t *testing.T) {
	t.Run("ResolvesFileInOverlayRoot", func(t *testing.T) {
		webDir := t.TempDir()
		webFile := filepath.Join(webDir, "public.txt")
		require.NoError(t, os.WriteFile(webFile, []byte("ok"), fs.ModeFile))

		resolved, ok := OverlayResolveFile(webDir, "public.txt")
		require.True(t, ok)
		assert.NotEmpty(t, resolved)
		assert.FileExists(t, resolved)
	})
	t.Run("MissingFileReturnsFalse", func(t *testing.T) {
		webDir := t.TempDir()
		resolved, ok := OverlayResolveFile(webDir, "missing.txt")
		assert.False(t, ok)
		assert.Equal(t, "", resolved)
	})
	t.Run("SymlinkEscapeReturnsFalse", func(t *testing.T) {
		rootDir := t.TempDir()
		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		require.NoError(t, os.WriteFile(outsideFile, []byte("secret"), fs.ModeFile))

		linkPath := filepath.Join(rootDir, "leak.txt")
		if err := os.Symlink(outsideFile, linkPath); err != nil {
			t.Skipf("symlink setup failed: %v", err)
		}

		resolved, ok := OverlayResolveFile(rootDir, "leak.txt")
		assert.False(t, ok)
		assert.Equal(t, "", resolved)
	})
}
