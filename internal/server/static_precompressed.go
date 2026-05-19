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

// PrecompressedStatic serves files from dir, preferring the precompressed
// `.zst`/`.gz` sibling when Accept-Encoding allows. Range requests always go
// through the identity representation (Range offsets are over identity bytes).
// Vary: Accept-Encoding is emitted only when a sibling actually exists; when
// HttpCompressionPreferences is empty the handler degrades to identity-only.
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

		// Vary: Accept-Encoding only when an encoded sibling could be selected.
		if len(prefs) > 0 {
			for _, suffix := range precompressedExtensions {
				if siblingExists(fs, rel+suffix) {
					c.Writer.Header().Add(header.Vary, header.AcceptEncoding)
					break
				}
			}
		}

		// Range requests and disabled compression both fall through to identity.
		if c.Request.Header.Get(header.Range) != "" || len(prefs) == 0 {
			http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), f)
			_ = f.Close()
			return
		}

		// Negotiate and stream the encoded sibling when one exists.
		switch encoding := NegotiateEncoding(c.Request.Header.Get(header.AcceptEncoding), prefs); encoding {
		case EncodingZstd, EncodingGzip:
			suffix := precompressedExtensions[encoding]
			if suffix != "" && siblingExists(fs, rel+suffix) {
				_ = f.Close()
				if servePrecompressed(c, fs, rel+suffix, encoding, info.Name(), info.ModTime()) {
					return
				}
				// Sibling vanished between probe and Open — fall back to identity.
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

// servePrecompressed streams the precompressed sibling via http.ServeContent,
// advertising Content-Encoding and deriving Content-Type from the identity
// filename. Uses identityModTime (not the sibling's) so Last-Modified +
// If-Modified-Since revalidation stays aligned across encodings.
// Returns false if the sibling can no longer be opened so the caller can fall back.
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

	// Set headers before http.ServeContent so its 304 path can strip them.
	if ct := contentTypeForName(identityName); ct != "" {
		c.Writer.Header().Set(header.ContentType, ct)
	}
	c.Writer.Header().Set(header.ContentEncoding, encoding)
	// http.ServeContent won't set Content-Length when Content-Encoding is set;
	// set it explicitly from the sibling's on-disk size.
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

// contentTypeForName returns the MIME type for a filename via the stdlib
// mime package, or "" when the extension is unknown.
func contentTypeForName(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}
