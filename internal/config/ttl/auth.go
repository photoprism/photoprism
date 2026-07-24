package ttl

var (
	// DownloadToken is the lifetime of signed download tokens (in seconds). Kept short so a leaked
	// token expires quickly; because the tokens overlap (a fresh one is issued on each response) it
	// only needs to exceed the client's token-refresh interval.
	DownloadToken Duration = 3600 // 1 hour
)

// DownloadTokenMinAge is the smallest sensible lifetime for signed download tokens (in seconds). It sits
// above the client's 10-minute config-refresh interval so an idle client's token never lapses before the
// next poll refreshes it; a smaller configured value is raised to this floor.
const DownloadTokenMinAge Duration = 900 // 15 minutes (> the 10-minute client config-poll interval)
