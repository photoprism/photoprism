package ttl

// DownloadArchiveAge is how long a generated download archive is kept in the temp directory (in
// seconds). Clients fetch an archive right after creating it, so the lifetime is short.
const DownloadArchiveAge Duration = 3600 // 1 hour
