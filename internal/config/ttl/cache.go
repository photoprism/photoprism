package ttl

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
