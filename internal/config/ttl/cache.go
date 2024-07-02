package ttl

var (
	CacheMaxAge  Duration = 31536000 // 365 days is the maximum cache time
	CacheDefault Duration = 2592000  // 30 days is the default cache time
	CacheVideo   Duration = 21600    // 6 hours for video streams
	CacheCover   Duration = 3600     // 1 hour for album cover images
)
