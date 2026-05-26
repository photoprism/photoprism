/*
Package disk provides cached helpers for filesystem free-space checks.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

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
package disk

import (
	"sync"
	"time"

	"github.com/photoprism/photoprism/pkg/fs/duf"
)

// CacheTTL controls how long Free results are cached per path.
// Read on every lookup so callers may retune it at runtime.
var CacheTTL = time.Minute

// freeEntry caches a single Free probe with its expiry time.
type freeEntry struct {
	free      uint64
	total     uint64
	expiresAt time.Time
}

var (
	freeMu    sync.RWMutex
	freeCache = map[string]freeEntry{}
)

// Free returns the cached free and total bytes for the filesystem that owns path.
func Free(path string) (free, total uint64, err error) {
	now := time.Now()

	freeMu.RLock()
	entry, ok := freeCache[path]
	freeMu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return entry.free, entry.total, nil
	}

	mount, err := duf.PathInfo(path)
	if err != nil {
		return 0, 0, err
	}

	freeMu.Lock()
	freeCache[path] = freeEntry{
		free:      mount.Free,
		total:     mount.Total,
		expiresAt: now.Add(CacheTTL),
	}
	freeMu.Unlock()

	return mount.Free, mount.Total, nil
}

// FlushFree clears the cache so a freshly freed disk is detected immediately
// without waiting for CacheTTL to expire.
func FlushFree() {
	freeMu.Lock()
	freeCache = map[string]freeEntry{}
	freeMu.Unlock()
}

// SetFree stores a free/total entry for path in the cache, expiring after CacheTTL.
// Allows tests and external probes to inject known values without hitting the filesystem.
func SetFree(path string, free, total uint64) {
	freeMu.Lock()
	freeCache[path] = freeEntry{
		free:      free,
		total:     total,
		expiresAt: time.Now().Add(CacheTTL),
	}
	freeMu.Unlock()
}

// StorageLow reports whether free space on path is below minPct percent of total.
func StorageLow(path string, minPct float64) (free uint64, low bool, err error) {
	if minPct <= 0 || minPct >= 100 {
		free, _, err = Free(path)
		return free, false, err
	}

	free, total, err := Free(path)
	if err != nil {
		return 0, false, err
	}

	if total == 0 {
		return free, false, nil
	}

	return free, float64(free)*100 < minPct*float64(total), nil
}
