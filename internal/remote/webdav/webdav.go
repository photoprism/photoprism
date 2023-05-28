/*
Package webdav provides WebDAV file sharing and synchronization.

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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
const Second = time.Second

// MaxRequestDuration is the maximum request duration e.g. for recursive retrieval of large remote directory structures.
const MaxRequestDuration = 30 * time.Minute

// Durations maps Timeout options to specific time durations.
var Durations = map[Timeout]time.Duration{
	TimeoutHigh:    120 * Second,
	TimeoutDefault: 60 * Second,
	TimeoutMedium:  60 * Second,
	TimeoutLow:     30 * Second,
	TimeoutNone:    0,
}
