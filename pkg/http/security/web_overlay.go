package security

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// overlayBlockedFileNames lists sensitive file base names that must never be
// served by the web overlay, even when present under /storage/web.
var overlayBlockedFileNames = map[string]struct{}{
	"id_ed25519":     {},
	"id_ed25519.pub": {},
	"id_rsa":         {},
	"id_rsa.pub":     {},
	"config.toml":    {},
	"config.yaml":    {},
	"config.yml":     {},
	"defaults.yaml":  {},
	"defaults.yml":   {},
	"options.yaml":   {},
	"options.yml":    {},
	"hub.yml":        {},
	"hub.yaml":       {},
	"auth.json":      {},
	"join_token":     {},
	"client_secret":  {},
}

// overlayBlockedFileExt lists sensitive file extensions that must never be
// served by the web overlay, even when present under /storage/web.
var overlayBlockedFileExt = map[string]struct{}{
	".key":  {},
	".rsa":  {},
	".jwk":  {},
	".pem":  {},
	".sql":  {},
	".toml": {},
}

// overlayBlockedPathPrefixes lists sensitive slash paths rooted under the
// overlay tree that must always be denied (path-prefix match).
var overlayBlockedPathPrefixes = []string{
	"node/secrets",
	"config/portal",
	"config/certificates",
}

// OverlayHasAmbiguousPath returns true when a request path is unsafe due to
// ambiguity, traversal indicators, hidden/special segments, or encoded probes.
func OverlayHasAmbiguousPath(requestPath, escapedPath string) bool {
	if requestPath == "" {
		return true
	}

	if strings.Contains(requestPath, "\\") || strings.Contains(requestPath, "//") {
		return true
	}

	if strings.Contains(requestPath, "/./") || strings.Contains(requestPath, "/../") ||
		strings.HasSuffix(requestPath, "/.") || strings.HasSuffix(requestPath, "/..") {
		return true
	}

	cleanPath := path.Clean(requestPath)
	if cleanPath == "." {
		return true
	} else if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	// Allow only trailing slash normalization; reject all other material changes.
	if requestPath != cleanPath && requestPath != cleanPath+"/" {
		return true
	}

	if relPath := strings.Trim(cleanPath, "/"); relPath != "" {
		for _, segment := range strings.Split(relPath, "/") {
			if segment == "" || segment == "." || segment == ".." || fs.FileNameHidden(segment) {
				return true
			}
		}
	}

	lowerEscaped := strings.ToLower(escapedPath)
	return strings.Contains(lowerEscaped, "%2e") || strings.Contains(lowerEscaped, "%2f") || strings.Contains(lowerEscaped, "%5c")
}

// OverlayRelativePath maps a request path to an overlay-relative path while
// enforcing the configured base path scope.
func OverlayRelativePath(requestPath, webBase string) (string, bool) {
	if requestPath == "" {
		return "", false
	}

	absPath := strings.Trim(strings.ReplaceAll(requestPath, "\\", "/"), "/")

	switch {
	case webBase != "" && absPath == webBase:
		// Request targets the base path itself (e.g. "/i/acme"), so map it to index.
		return "", true
	case webBase != "" && strings.HasPrefix(absPath, webBase+"/"):
		// Request is under the configured base path; strip the prefix for web storage lookup.
		return strings.TrimPrefix(absPath, webBase+"/"), true
	case webBase != "":
		// Ignore requests outside the configured base path when a base path is enforced.
		return "", false
	default:
		return absPath, true
	}
}

// OverlayPathBlocked returns true when an overlay path targets hidden/special
// path segments or files and directories blocked by denylist rules.
func OverlayPathBlocked(webPath string) bool {
	relPath := strings.Trim(strings.ReplaceAll(webPath, "\\", "/"), "/")
	if relPath == "" {
		return false
	}

	segments := strings.Split(relPath, "/")

	for _, segment := range segments {
		if segment == "" || fs.FileNameHidden(segment) {
			return true
		}
	}

	lowerPath := strings.ToLower(relPath)
	lowerBase := path.Base(lowerPath)

	// Block requests for sensitive file names like "auth.json".
	if _, blocked := overlayBlockedFileNames[lowerBase]; blocked {
		return true
	}

	// Block file extensions commonly used for private keys or backups.
	if _, blocked := overlayBlockedFileExt[path.Ext(lowerBase)]; blocked {
		return true
	}

	// Additionally check for sensitive path prefixes that should not be served.
	for _, prefix := range overlayBlockedPathPrefixes {
		if lowerPath == prefix || strings.HasPrefix(lowerPath, prefix+"/") {
			return true
		}
	}

	return false
}

// OverlayResolveFile resolves an overlay file path and returns false when the
// target cannot be served safely from inside the canonical overlay root.
func OverlayResolveFile(webDir, webPath string) (string, bool) {
	if webDir == "" || webPath == "" {
		return "", false
	}

	overlayPath := strings.Trim(strings.ReplaceAll(webPath, "\\", "/"), "/")
	if overlayPath == "" {
		return "", false
	}

	candidate := filepath.Join(webDir, filepath.FromSlash(overlayPath))
	if !fs.FileExists(candidate) {
		return "", false
	}

	rootRealPath, err := filepath.EvalSymlinks(webDir)
	if err != nil {
		rootRealPath = webDir
	}

	rootAbsPath, err := filepath.Abs(rootRealPath)
	if err != nil {
		return "", false
	}

	fileRealPath, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", false
	}

	fileAbsPath, err := filepath.Abs(fileRealPath)
	if err != nil {
		return "", false
	}

	relPath, err := filepath.Rel(rootAbsPath, fileAbsPath)
	if err != nil {
		return "", false
	} else if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", false
	}

	return fileAbsPath, true
}
