/*
Package webdav provides WebDAV file sharing and synchronization.

It includes the outbound client wrapper used by services and sync workers for
recursive directory discovery, uploads, downloads, and compatibility fallbacks
when remote servers reject recursive PROPFIND requests with "Depth: infinity".

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
package webdav

import (
	"time"

	"github.com/photoprism/photoprism/internal/event"
)

// Global log instance.
var log = event.Log

// Timeout specifies a timeout mode for WebDAV operations.
type Timeout string

// Request Timeout options.
const (
	TimeoutHigh    Timeout = "high"   // 120 * Second
	TimeoutDefault Timeout = ""       // 60 * Second
	TimeoutMedium  Timeout = "medium" // 60 * Second
	TimeoutLow     Timeout = "low"    // 30 * Second
	TimeoutNone    Timeout = "none"   // 0
)

// Second represents a second on which other timeouts are based.
//
//revive:disable-next-line:time-naming // keep exported constant name for API compatibility
const Second = time.Second

// MaxRequestDuration is the maximum request duration for large recursive directory listings, including fallback traversal.
const MaxRequestDuration = 30 * time.Minute

// transferConnectTimeout limits TCP connect time for upload and download requests without imposing a total transfer deadline.
const transferConnectTimeout = 30 * time.Second

// transferTLSHandshakeTimeout limits TLS handshakes for upload and download requests.
const transferTLSHandshakeTimeout = 20 * time.Second

// transferIdleConnTimeout limits how long pooled idle WebDAV connections are kept for transfer operations.
const transferIdleConnTimeout = 90 * time.Second

// transferExpectContinueTimeout limits how long transfer requests wait for an initial 100-continue response.
const transferExpectContinueTimeout = 10 * time.Second

// Durations maps Timeout options to specific time durations.
var Durations = map[Timeout]time.Duration{
	TimeoutHigh:    120 * Second,
	TimeoutDefault: 60 * Second,
	TimeoutMedium:  60 * Second,
	TimeoutLow:     30 * Second,
	TimeoutNone:    0,
}
