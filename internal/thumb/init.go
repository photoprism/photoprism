package thumb

import "github.com/dustin/go-humanize/english"

const (
	// MiB represents one mebibyte.
	MiB = 1024 * 1024
	// GiB represents one gibibyte.
	GiB = 1024 * MiB
	// DefaultCacheMem specifies the default libvips cache memory limit.
	DefaultCacheMem = 128 * MiB
	// DefaultCacheSize is the default number of cached operations.
	DefaultCacheSize = 128
	// DefaultCacheFiles is the default number of cached files.
	DefaultCacheFiles = 16
	// DefaultWorkers is the default worker count when not specified.
	DefaultWorkers = 1
)

var (
	// MaxCacheMem is the maximum memory libvips may use for caching.
	MaxCacheMem = DefaultCacheMem
	// MaxCacheSize limits the number of cached operations.
	MaxCacheSize = DefaultCacheSize
	// MaxCacheFiles limits the number of cached files.
	MaxCacheFiles = DefaultCacheFiles
	// NumWorkers defines the number of libvips worker threads.
	NumWorkers = DefaultWorkers
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
	switch {
	case maxWorkers > 0:
		NumWorkers = maxWorkers // explicitly configured
	case maxWorkers < 0:
		NumWorkers = 0 // use libvips default
	default:
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
