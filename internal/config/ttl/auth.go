package ttl

const (
	// DownloadTokenDefaultAge is the lifetime of signed download tokens (in seconds) unless configured.
	// Kept short so a leaked token expires quickly; tokens overlap (a fresh one is issued on each
	// response), so it only needs to exceed the client's refresh interval.
	DownloadTokenDefaultAge Duration = 3600 // 1 hour

	// DownloadTokenMinAge is the smallest accepted lifetime for signed download tokens (in seconds).
	// Above the client's 10-minute config poll, so an idle client's token cannot lapse before the next
	// refresh; shorter configured values are raised to it.
	DownloadTokenMinAge Duration = 900 // 15 minutes
)

// DownloadToken is the effective lifetime of signed download tokens (in seconds), set by config.Propagate.
var DownloadToken = DownloadTokenDefaultAge
