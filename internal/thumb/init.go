package thumb

import "github.com/dustin/go-humanize/english"

const (
	MiB               = 1024 * 1024
	DefaultCacheMem   = 64 * MiB
	DefaultCacheSize  = 100
	DefaultCacheFiles = 0
	DefaultWorkers    = 1
)

var (
	MaxCacheMem   = DefaultCacheMem
	MaxCacheSize  = DefaultCacheSize
	MaxCacheFiles = DefaultCacheFiles
	NumWorkers    = DefaultWorkers
)

// Init configures the thumb package based on available memory and allowed number of workers.
func Init(availableMemory uint64, maxWorkers int) {
	// Set the maximum amount of cached data allowed
	// before libvips drops cached operations.
	switch {
	case availableMemory > 4:
		MaxCacheMem = 512 * MiB
	case availableMemory > 2:
		MaxCacheMem = 256 * MiB
	case availableMemory > 1:
		MaxCacheMem = 128 * MiB
	default:
		MaxCacheMem = DefaultCacheMem
	}

	// Set the number of worker threads that libvips can use.
	if maxWorkers > 0 {
		// Using the specified number of workers.
		NumWorkers = maxWorkers
	} else if maxWorkers < 0 {
		// Using built-in default.
		NumWorkers = 0
	} else {
		// Default to one worker.
		NumWorkers = DefaultWorkers
	}

	log.Debugf("vips: using up to %d MB of cache and %s", MaxCacheMem/MiB, english.Plural(NumWorkers, "worker", "workers"))
}

// Shutdown shuts down dependencies like libvips.
func Shutdown() {
	VipsShutdown()
}
