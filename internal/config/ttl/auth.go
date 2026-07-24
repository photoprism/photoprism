package ttl

var (
	// DownloadToken is the lifetime of signed download tokens (in seconds). Kept short so a leaked
	// token expires quickly; because the tokens overlap (a fresh one is issued on each response) it
	// only needs to exceed the client's token-refresh interval.
	DownloadToken Duration = 3600 // 1 hour
)
