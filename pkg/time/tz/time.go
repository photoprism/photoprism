package tz

import (
	"time"
)

var (
	// TimeUTC provides the UTC location.
	TimeUTC = time.UTC
	// TimeLocal provides the configured local location.
	TimeLocal = time.FixedZone(Local, 0)
)
