package header

// Durations in seconds, e.g. to set a maximum cache age.
const (
	DurationMinute int = 60
	DurationHour       = DurationMinute * 60 // One hour in seconds
	DurationDay        = DurationHour * 24   // One day in seconds
	DurationWeek       = DurationDay * 7     // One week in seconds
	DurationMonth      = DurationDay * 31    // About one month in seconds
	DurationYear       = DurationDay * 365   // 365 days in seconds
)
