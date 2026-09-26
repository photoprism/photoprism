package meta

import (
	"errors"
	"strconv"
	"strings"
)

// exposureMaxSeconds is the longest exposure time that is stored, in seconds.
const exposureMaxSeconds = 1e6

// normalizeExposure returns an exposure time in the format of formatExposure, so that the values reported
// by ExifTool, Exif, and XMP are stored alike. Values that cannot be parsed are returned unchanged, unless
// they exceed the length of the database column.
func normalizeExposure(s string) string {
	if s = strings.TrimSpace(s); s == "" {
		return ""
	}

	if n, d, found := strings.Cut(s, "/"); found {
		num, numErr := strconv.ParseFloat(strings.TrimSpace(n), 64)
		den, denErr := strconv.ParseFloat(strings.TrimSpace(d), 64)

		if parsedFloat(numErr) && parsedFloat(denErr) {
			if den == 0 {
				return ""
			}

			return formatExposure(num / den)
		}
	} else if secs, err := strconv.ParseFloat(s, 64); parsedFloat(err) {
		return formatExposure(secs)
	}

	if len(s) > 64 {
		return ""
	}

	return s
}

// parsedFloat reports whether strconv.ParseFloat returned a number, which is infinite if it is out of range.
func parsedFloat(err error) bool {
	return err == nil || errors.Is(err, strconv.ErrRange)
}
