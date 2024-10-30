package thumb

import "github.com/dustin/go-humanize/english"

const (
	MiB               = 1024 * 1024
	GiB               = 1024 * MiB
	DefaultCacheMem   = 128 * MiB
	DefaultCacheSize  = 128
	DefaultCacheFiles = 16
	DefaultWorkers    = 1
)

var (
	MaxCacheMem   = DefaultCacheMem
	MaxCacheSize  = DefaultCacheSize
	MaxCacheFiles = DefaultCacheFiles
	NumWorkers    = DefaultWorkers
)

// Init initializes the package config based on the available memory,
// the allowed number of workers and the image processing library to be used.
func Init(availableMemory uint64, maxWorkers int, imgLib string) {
	// Set the maximum amount of cached data allowed
	// before libvips drops cached operations.
	switch {
	case availableMemory >= 4*GiB:
		MaxCacheMem = 512 * MiB
	case availableMemory >= 1*GiB:
		MaxCacheMem = 256 * MiB
	case availableMemory <= 0:
		// Use default if free memory could not be detected.
		MaxCacheMem = DefaultCacheMem
	default:
		// Reduce cache size and number of workers if the system seems low on memory.
		MaxCacheMem = 64 * MiB
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

	// Set the image processing library.
	switch imgLib {
	case LibVips, "libvips":
		Library = LibVips
		log.Debugf("vips: max cache size is %d MB, using up to %s", MaxCacheMem/MiB, english.Plural(NumWorkers, "worker", "workers"))
	default:
		Library = LibImaging
	}
}

// Shutdown shuts down dependencies like libvips.
func Shutdown() {
	VipsShutdown()
}
