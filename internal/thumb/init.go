package thumb

import "github.com/dustin/go-humanize/english"

const (
	MiB               = 1024 * 1024
	GiB               = 1024 * MiB
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

// Init configures the thumb package with the available memory,
// allowed number of workers and image library to be used.
func Init(availableMemory uint64, maxWorkers int, generator string) {
	// Set the maximum amount of cached data allowed
	// before libvips drops cached operations.
	switch {
	case availableMemory >= 4*GiB:
		MaxCacheMem = 512 * MiB
	case availableMemory >= 2*GiB:
		MaxCacheMem = 256 * MiB
	case availableMemory >= 1*GiB:
		MaxCacheMem = 128 * MiB
	case availableMemory <= 0:
		// Use default if free memory could not be detected.
		MaxCacheMem = DefaultCacheMem
	default:
		// Reduce cache size and number of workers if the system seems low on memory.
		MaxCacheMem = 32 * MiB
		maxWorkers = 1
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

	// Set the thumbnail generator library to use.
	switch generator {
	case LibVips, "libvips":
		Generator = LibVips
		log.Debugf("vips: max cache size is %d MB, using up to %s", MaxCacheMem/MiB, english.Plural(NumWorkers, "worker", "workers"))
	default:
		Generator = LibImaging
	}
}

// Shutdown shuts down dependencies like libvips.
func Shutdown() {
	VipsShutdown()
}
