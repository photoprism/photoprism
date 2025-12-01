package config

import (
	"time"

	"github.com/photoprism/photoprism/pkg/time/unix"
)

// ApiUri defines the standard path for handling REST requests.
const ApiUri = "/api/v1"

// DownloadUri defines the file download URI based on the ApiUri.
const DownloadUri = ApiUri + "/dl"

// LibraryUri defines the path for user interface routes.
const LibraryUri = "/library"

// StaticUri defines the standard path for serving static content.
const StaticUri = "/static"

// CustomStaticUri defines the standard path for serving custom static content.
const CustomStaticUri = "/c/static"

// ThemeUri defines the optional theme URI path for serving theme assets.
const ThemeUri = "/_theme"

// DefaultIndexSchedule defines the default indexing schedule in cron format.
const DefaultIndexSchedule = "" // e.g. "0 */3 * * *" for every 3 hours

// DefaultAutoIndexDelay sets the default delay (in seconds) before background indexing starts.
const DefaultAutoIndexDelay = 300 // 5 Minutes
// DefaultAutoImportDelay sets the default delay (in seconds) before background imports start (-1 disables).
const DefaultAutoImportDelay = -1 // Disabled

// MinWakeupInterval is the minimum allowed interval for the background worker.
const MinWakeupInterval = time.Minute // 1 Minute
// MaxWakeupInterval is the maximum allowed interval for the background worker.
const MaxWakeupInterval = time.Hour * 24 // 1 Day
// DefaultWakeupIntervalSeconds is the default worker interval in seconds.
const DefaultWakeupIntervalSeconds = int(15 * 60) // 15 Minutes
// DefaultWakeupInterval is the default worker interval as a duration.
const DefaultWakeupInterval = time.Second * time.Duration(DefaultWakeupIntervalSeconds)

// MegaByte defines a megabyte in bytes.
const MegaByte = 1000 * 1000 // 1,000,000 Bytes

// GigaByte defines gigabyte in bytes.
const GigaByte = MegaByte * 1000 // 1,000,000,000 Bytes

// MinMem defines the minimum amount of system memory required.
const MinMem = GigaByte

// RecommendedMem defines the recommended amount of system memory.
const RecommendedMem = 3 * GigaByte // 3,000,000,000 Bytes

// DefaultResolutionLimit defines the default resolution limit.
const DefaultResolutionLimit = 150 // 150 Megapixels

// serialName defines the name of the unique storage serial.
const serialName = "serial"

// DefaultSessionMaxAge defines the standard session expiration time in seconds.
const DefaultSessionMaxAge = unix.Week * 2

// DefaultSessionTimeout defines the standard session idle time in seconds.
const DefaultSessionTimeout = unix.Week

// DefaultSessionCache defines the default session cache duration in seconds.
const DefaultSessionCache = unix.Minute * 15

// Product feature tags used to automatically generate documentation.
const (
	Pro        = "pro"
	Portal     = "portal"
	Plus       = "plus"
	Essentials = "essentials"
	Community  = "ce"
)
