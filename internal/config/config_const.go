package config

import (
	"time"
)

// ApiUri defines the standard path for handling REST requests.
const ApiUri = "/api/v1"

// StaticUri defines the standard path for serving static content.
const StaticUri = "/static"

// CustomStaticUri defines the standard path for serving custom static content.
const CustomStaticUri = "/c/static"

// DefaultAutoIndexDelay and DefaultAutoImportDelay set the default safety delay duration
// before starting to index/import in the background.
const DefaultAutoIndexDelay = int(5 * 60)  // 5 Minutes
const DefaultAutoImportDelay = int(3 * 60) // 3 Minutes

// MinWakeupInterval and MaxWakeupInterval limit the interval duration
// in which the background worker can be invoked.
const MinWakeupInterval = time.Minute             // 1 Minute
const MaxWakeupInterval = time.Hour * 24          // 1 Day
const DefaultWakeupIntervalSeconds = int(15 * 60) // 15 Minutes
const DefaultWakeupInterval = time.Second * time.Duration(DefaultWakeupIntervalSeconds)

// Megabyte defines a megabyte in bytes.
const Megabyte = 1000 * 1000 // 1,000,000 Bytes

// Gigabyte defines gigabyte in bytes.
const Gigabyte = Megabyte * 1000 // 1,000,000,000 Bytes

// MinMem defines the minimum amount of system memory required.
const MinMem = Gigabyte

// RecommendedMem defines the recommended amount of system memory.
const RecommendedMem = 3 * Gigabyte // 3,000,000,000 Bytes

// DefaultResolutionLimit defines the default resolution limit.
const DefaultResolutionLimit = 150 // 150 Megapixels

// serialName defines the name of the unique storage serial.
const serialName = "serial"

// UnixHour defines one hour in UnixTime.
const UnixHour int64 = 3600

// UnixDay defines one day in UnixTime.
const UnixDay = UnixHour * 24

// UnixWeek defines one week in UnixTime.
const UnixWeek = UnixDay * 7

// DefaultSessionMaxAge defines the standard session expiration time in seconds.
const DefaultSessionMaxAge = UnixWeek * 2

// DefaultSessionTimeout defines the standard session idle time in seconds.
const DefaultSessionTimeout = UnixWeek

// Product feature tags used to automatically generate documentation.
const (
	Pro        = "pro"
	Plus       = "plus"
	Essentials = "essentials"
)
