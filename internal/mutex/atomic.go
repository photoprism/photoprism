package mutex

import (
	"sync/atomic"
)

// Restart signals that the application should be restarted,
// e.g. after an update or a config changes.
var Restart = atomic.Bool{}

// TempArchives signals that download archives may exist in the temp directory, so workers can skip
// scanning it. Starts true to cover archives left by a previous process, is set when an archive is
// created, and is cleared by a sweep that finds none left.
var TempArchives = atomic.Bool{}

func init() {
	TempArchives.Store(true)
}
