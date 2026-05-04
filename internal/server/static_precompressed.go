/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

*/

package server

import (
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// precompressedExtensions maps each supported content-encoding to the
// sibling file suffix produced by frontend/scripts/precompress.js.
var precompressedExtensions = map[string]string{
	EncodingZstd: ".zst",
	EncodingGzip: ".gz",
}

// PrecompressedStatic returns a Gin handler that serves files from dir,
// preferring the precompressed `.zst` or `.gz` sibling produced by
// frontend/scripts/precompress.js when the client's Accept-Encoding allows.
//
// The handler always serves the identity representation for Range requests,
// since the byte offsets in a successful Range response correspond to
// identity bytes — serving a precompressed slice would scramble the offsets
// the client asked for and contradict any subsequent Range continuation.
//
// When at least one precompressed sibling exists for the requested file,
// the handler emits Vary: Accept-Encoding regardless of which encoding wins,
// so shared caches keep separate entries per coding.
//
// When the operator disables compression (HttpCompressionPreferences is
// empty), the handler degenerates to identity-only serving and never picks
// a precompressed sibling, matching the runtime middleware's contract that
// PHOTOPRISM_HTTP_COMPRESSION=none disables every encoded code path.
func PrecompressedStatic(conf *config.Config, dir string) gin.HandlerFunc {
	fs := http.Dir(dir)
	prefs := conf.HttpCompressionPreferences()

	return func(c *gin.Context) {
		// Gin only registers this handler for GET/HEAD, but guard anyway so
		// odd verbs (e.g. via an upstream rewrite) get a clean 405 instead of
		// a half-served body.
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		rel := c.Param("filepath")
		if rel == "" {
			rel = "/"
		}

		// http.Dir.Open rejects path-traversal attempts, returning os.ErrNotExist
		// for cleaned paths that escape the root.
		f, err := fs.Open(rel)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			_ = f.Close()
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Advertise Accept-Encoding as a Vary axis only when the negotiator
		// could realistically select an encoded sibling for this resource —
		// i.e. compression is enabled by config AND a sibling exists on disk.
		// With PHOTOPRISM_HTTP_COMPRESSION=none the response is always
		// identity regardless of Accept-Encoding, so Vary would mislead
		// shared caches into keeping redundant per-encoding entries that
		// all hold the same identity bytes.
		if len(prefs) > 0 {
			for _, suffix := range precompressedExtensions {
				if siblingExists(fs, rel+suffix) {
					c.Writer.Header().Add(header.Vary, header.AcceptEncoding)
					break
				}
			}
		}

		// Range requests must serve the identity representation — see
		// statusBypassesCompression in compress.go for the same reasoning
		// applied to the runtime middleware's 206 bypass. With compression
		// disabled, fall straight through to identity as well.
		if c.Request.Header.Get(header.Range) != "" || len(prefs) == 0 {
			http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), f)
			_ = f.Close()
			return
		}

		// Negotiate an encoding the client accepts and the operator prefers.
		// On a hit, close the identity handle and stream the sibling instead.
		switch encoding := NegotiateEncoding(c.Request.Header.Get(header.AcceptEncoding), prefs); encoding {
		case EncodingZstd, EncodingGzip:
			suffix := precompressedExtensions[encoding]
			if suffix != "" && siblingExists(fs, rel+suffix) {
				_ = f.Close()
				if servePrecompressed(c, fs, rel+suffix, encoding, info.Name(), info.ModTime()) {
					return
				}
				// servePrecompressed returns false if the sibling vanished
				// between the existence probe and the Open call — fall back
				// to identity by re-opening below.
				if reopen, openErr := fs.Open(rel); openErr == nil {
					http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), reopen)
					_ = reopen.Close()
				} else {
					c.AbortWithStatus(http.StatusNotFound)
				}
				return
			}
		}

		http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), f)
		_ = f.Close()
	}
}

// servePrecompressed streams the precompressed sibling at siblingPath via
// http.ServeContent, advertising Content-Encoding and deriving Content-Type
// from the identity filename so clients see e.g. application/javascript
// rather than application/zstd. Delegating to http.ServeContent preserves
// the conditional-request semantics that gin.Static gave us before this
// refactor — the handler emits Last-Modified on every successful response
// and returns 304 Not Modified when an If-Modified-Since header matches
// the identity's last-modified time. (http.ServeContent also honors
// If-None-Match when the caller sets an ETag; this handler does not, so
// the active validator pair is Last-Modified + If-Modified-Since.)
//
// identityModTime is the canonical resource version (the identity file's
// mtime); using it instead of the sibling's mtime keeps revalidation
// aligned across encodings — both the identity and every sibling share a
// single 304 boundary.
//
// Range requests are not expected here: the caller short-circuits to the
// identity path on `Range:` precisely so http.ServeContent doesn't slice
// the encoded byte stream.
//
// Returns false (and serves nothing) if the sibling can no longer be
// opened so the caller can fall back to identity.
func servePrecompressed(
	c *gin.Context,
	fs http.FileSystem,
	siblingPath, encoding, identityName string,
	identityModTime time.Time,
) bool {
	sibling, err := fs.Open(siblingPath)
	if err != nil {
		return false
	}
	defer func() { _ = sibling.Close() }()

	info, err := sibling.Stat()
	if err != nil || info.IsDir() {
		return false
	}

	// Set headers before http.ServeContent so its conditional-request
	// branches (304 Not Modified) can clear them via writeNotModified.
	if ct := contentTypeForName(identityName); ct != "" {
		c.Writer.Header().Set(header.ContentType, ct)
	}
	c.Writer.Header().Set(header.ContentEncoding, encoding)
	// http.ServeContent declines to set Content-Length when Content-Encoding
	// is non-empty (it can't infer on-wire size from a transformed stream),
	// so set it explicitly from the sibling's actual size. ServeContent's
	// 304 path strips Content-Length via writeNotModified, which is correct.
	c.Writer.Header().Set(header.ContentLength, strconv.FormatInt(info.Size(), 10))

	http.ServeContent(c.Writer, c.Request, identityName, identityModTime, sibling)
	return true
}

// siblingExists reports whether a regular file exists at the given path
// within the static filesystem. Directories are not considered siblings.
func siblingExists(fs http.FileSystem, p string) bool {
	f, err := fs.Open(p)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

// contentTypeForName returns the MIME type for a filename based on its
// extension, or an empty string when no mapping is known. Defers to the
// stdlib mime package which understands the common web/application types
// (.js, .css, .json, .svg, …) plus anything registered by the host.
func contentTypeForName(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}
