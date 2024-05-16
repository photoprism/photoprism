package thumb

var (
	ConcurrencyLevel = 1
	MaxCacheFiles    = 0
	MaxCacheMem      = 0
	MaxCacheSize     = 0
)

// Shutdown shuts down dependencies like libvips.
func Shutdown() {
	VipsShutdown()
}
