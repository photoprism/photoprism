/*
Package proxy provides shared defaults and helpers for path-based HTTP proxy routing.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package proxy

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/http/header"
)

const (
	// DefaultPathPrefix defines the default shared-domain proxy route path prefix.
	DefaultPathPrefix = "/i/"
	// Timeout defines the upstream response header and idle timeout.
	Timeout = 60 * time.Second
	// CacheTTL is the positive cache TTL for instance resolution.
	CacheTTL = 60 * time.Second
	// CacheNegativeTTL is the negative cache TTL for missing instances.
	CacheNegativeTTL = 2 * time.Second
	// CacheCleanup is the cleanup interval for the proxy cache.
	CacheCleanup = 1 * time.Minute
)

var (
	// PathPrefix defines the shared-domain proxy route path prefix, e.g. used by the Portal.
	// This startup value can be overridden before route registration.
	PathPrefix = DefaultPathPrefix
	// Methods lists additional methods that proxy routes should support
	// beyond standard methods provided by gin.Engine.Any.
	Methods = []string{
		header.MethodMkcol,
		header.MethodCopy,
		header.MethodMove,
		header.MethodLock,
		header.MethodUnlock,
		header.MethodPropfind,
		header.MethodProppatch,
		header.MethodReport,
		header.MethodSearch,
		header.MethodMkcalendar,
		header.MethodACL,
		header.MethodBind,
		header.MethodUnbind,
		header.MethodRebind,
		header.MethodVersionControl,
		header.MethodCheckout,
		header.MethodUncheckout,
		header.MethodCheckin,
		header.MethodUpdate,
		header.MethodLabel,
		header.MethodMerge,
		header.MethodMkworkspace,
		header.MethodMkactivity,
		header.MethodBaselineControl,
		header.MethodOrderpatch,
	}
)

// NormalizePathPrefix validates and normalizes a proxy path prefix.
func NormalizePathPrefix(prefix string) (string, error) {
	prefix = strings.TrimSpace(prefix)

	if prefix == "" {
		return DefaultPathPrefix, nil
	}

	trimmed := strings.Trim(prefix, "/")

	if trimmed == "" {
		return "", fmt.Errorf("proxy path prefix must not be root")
	}

	if strings.ContainsAny(trimmed, "?#*") {
		return "", fmt.Errorf("proxy path prefix contains invalid characters")
	}

	if strings.ContainsRune(trimmed, '\\') {
		return "", fmt.Errorf("proxy path prefix must not contain backslashes")
	}

	normalized := "/" + trimmed

	// Reject ambiguous prefixes (duplicate slashes, dot segments, parent traversal).
	if path.Clean(normalized) != normalized {
		return "", fmt.Errorf("proxy path prefix contains invalid path segments")
	}

	return normalized + "/", nil
}

// SetPathPrefix sets PathPrefix after validation and normalization.
func SetPathPrefix(prefix string) error {
	normalized, err := NormalizePathPrefix(prefix)

	if err != nil {
		return err
	}

	PathPrefix = normalized

	return nil
}
