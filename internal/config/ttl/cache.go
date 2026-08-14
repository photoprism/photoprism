package ttl

// Client-facing Cache-Control lifetimes in seconds.
// CacheDefault and CacheVideo are overridden by Config.Propagate.
var (
	// CacheMaxAge is the maximum cache duration (in seconds).
	CacheMaxAge Duration = 31536000 // 365 days is the maximum cache time
	// CacheDefault is the default cache duration (in seconds).
	CacheDefault Duration = 2592000 // 30 days is the default cache time
	// CacheVideo is the cache duration for video streams (in seconds).
	CacheVideo Duration = 21600 // 6 hours for video streams
	// CacheCover is the cache duration for album cover images (in seconds).
	CacheCover Duration = 3600 // 1 hour for album cover images
)

const (
	// CacheCollection is the lifetime of the in-memory caches that hold derived views of library
	// content, such as albums, folders and covers (in seconds). Kept short because these change
	// whenever pictures are added, moved or removed, and not every writer invalidates them.
	CacheCollection Duration = 900 // 15 minutes

	// CacheCollectionCleanup is how often those caches purge expired entries (in seconds).
	// Unlike CacheCover, which is the Cache-Control lifetime for the same responses,
	// this and CacheCollection are server-side lifetimes used with Duration().
	CacheCollectionCleanup Duration = 300 // 5 minutes
)
