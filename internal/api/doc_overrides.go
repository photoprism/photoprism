package api

import "time"

// SwaggerTimeDuration overrides the generated schema for time.Duration to avoid unstable enums
// from the standard library constants (Nanosecond, Minute, etc.). Using a simple integer schema is
// accurate (nanoseconds) and deterministic.
//
//	@name			time.Duration
//	@description	Duration in nanoseconds (int64). Examples: 1000000000 (1s), 60000000000 (1m).
type SwaggerTimeDuration = time.Duration
