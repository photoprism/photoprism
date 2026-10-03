package encode

import (
	"strconv"
	"strings"
)

// ParseBitrate returns the bitrate in Mbit/s of a value such as "25M", or 0 if it is not a positive number of megabits.
// Values above MaxBitrateLimit are clamped to it.
func ParseBitrate(s string) int {
	s, ok := strings.CutSuffix(strings.TrimSpace(s), "M")

	if !ok {
		return 0
	}

	// Atoi returns the maximum int for values out of range, so they are clamped as well.
	n, err := strconv.Atoi(s)

	switch {
	case n > MaxBitrateLimit:
		return MaxBitrateLimit
	case err != nil || n <= 0:
		return 0
	default:
		return n
	}
}
