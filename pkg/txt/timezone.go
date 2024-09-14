package txt

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// TimeZone returns a time zone for the given UTC offset string.
func TimeZone(offset string) *time.Location {
	if offset == "" {
		// Local time.
	} else if offset == "UTC" || offset == "Z" {
		return time.UTC
	} else if seconds, err := TimeOffset(offset); err == nil {
		if h := seconds / 3600; h > 0 || h < 0 {
			return time.FixedZone(fmt.Sprintf("UTC%+d", h), seconds)
		}
	} else if zone, zoneErr := time.LoadLocation(offset); zoneErr == nil {
		return zone
	}

	return time.FixedZone("", 0)
}

// NormalizeUtcOffset returns a normalized UTC time offset string.
func NormalizeUtcOffset(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	switch s {
	case "-12", "-12:00", "UTC-12", "UTC-12:00":
		return "UTC-12"
	case "-11", "-11:00", "UTC-11", "UTC-11:00":
		return "UTC-11"
	case "-10", "-10:00", "UTC-10", "UTC-10:00":
		return "UTC-10"
	case "-9", "-09", "-09:00", "UTC-9", "UTC-09:00":
		return "UTC-9"
	case "-8", "-08", "-08:00", "UTC-8", "UTC-08:00":
		return "UTC-8"
	case "-7", "-07", "-07:00", "UTC-7", "UTC-07:00":
		return "UTC-7"
	case "-6", "-06", "-06:00", "UTC-6", "UTC-06:00":
		return "UTC-6"
	case "-5", "-05", "-05:00", "UTC-5", "UTC-05:00":
		return "UTC-5"
	case "-4", "-04", "-04:00", "UTC-4", "UTC-04:00":
		return "UTC-4"
	case "-3", "-03", "-03:00", "UTC-3", "UTC-03:00":
		return "UTC-3"
	case "-2", "-02", "-02:00", "UTC-2", "UTC-02:00":
		return "UTC-2"
	case "-1", "-01", "-01:00", "UTC-1", "UTC-01:00":
		return "UTC-1"
	case "Z", "UTC", "UTC+0", "UTC-0", "UTC+00:00", "UTC-00:00":
		return time.UTC.String()
	case "01:00", "+1", "+01", "+01:00", "UTC+1", "UTC+01:00":
		return "UTC+1"
	case "02:00", "+2", "+02", "+02:00", "UTC+2", "UTC+02:00":
		return "UTC+2"
	case "03:00", "+3", "+03", "+03:00", "UTC+3", "UTC+03:00":
		return "UTC+3"
	case "04:00", "+4", "+04", "+04:00", "UTC+4", "UTC+04:00":
		return "UTC+4"
	case "05:00", "+5", "+05", "+05:00", "UTC+5", "UTC+05:00":
		return "UTC+5"
	case "06:00", "+6", "+06", "+06:00", "UTC+6", "UTC+06:00":
		return "UTC+6"
	case "07:00", "+7", "+07", "+07:00", "UTC+7", "UTC+07:00":
		return "UTC+7"
	case "08:00", "+8", "+08", "+08:00", "UTC+8", "UTC+08:00":
		return "UTC+8"
	case "09:00", "+9", "+09", "+09:00", "UTC+9", "UTC+09:00":
		return "UTC+9"
	case "10:00", "+10", "+10:00", "UTC+10", "UTC+10:00":
		return "UTC+10"
	case "11:00", "+11", "+11:00", "UTC+11", "UTC+11:00":
		return "UTC+11"
	case "12:00", "+12", "+12:00", "UTC+12", "UTC+12:00":
		return "UTC+12"
	}

	return ""
}

// UtcOffset returns the time difference as UTC offset string.
func UtcOffset(local, utc time.Time, offset string) string {
	if offset = NormalizeUtcOffset(offset); offset != "" {
		return offset
	}

	if local.IsZero() || utc.IsZero() || local == utc {
		return ""
	}

	d := local.Sub(utc).Hours()

	// Return if time difference includes fractions of an hour.
	if math.Abs(d-float64(int64(d))) > 0.1 {
		return ""
	}

	// Check if time difference is within expected range (hours).
	if h := int(d); h == 0 || h < -12 || h > 12 {
		return ""
	} else {
		return fmt.Sprintf("UTC%+d", h)
	}
}

// TimeOffset returns the UTC time offset in seconds or an error if it is invalid.
func TimeOffset(utcOffset string) (seconds int, err error) {
	switch utcOffset {
	case "-12", "-12:00", "UTC-12", "UTC-12:00":
		seconds = -12 * 3600
	case "-11", "-11:00", "UTC-11", "UTC-11:00":
		seconds = -11 * 3600
	case "-10", "-10:00", "UTC-10", "UTC-10:00":
		seconds = -10 * 3600
	case "-9", "-09", "-09:00", "UTC-9", "UTC-09:00":
		seconds = -9 * 3600
	case "-8", "-08", "-08:00", "UTC-8", "UTC-08:00":
		seconds = -8 * 3600
	case "-7", "-07", "-07:00", "UTC-7", "UTC-07:00":
		seconds = -7 * 3600
	case "-6", "-06", "-06:00", "UTC-6", "UTC-06:00":
		seconds = -6 * 3600
	case "-5", "-05", "-05:00", "UTC-5", "UTC-05:00":
		seconds = -5 * 3600
	case "-4", "-04", "-04:00", "UTC-4", "UTC-04:00":
		seconds = -4 * 3600
	case "-3", "-03", "-03:00", "UTC-3", "UTC-03:00":
		seconds = -3 * 3600
	case "-2", "-02", "-02:00", "UTC-2", "UTC-02:00":
		seconds = -2 * 3600
	case "-1", "-01", "-01:00", "UTC-1", "UTC-01:00":
		seconds = -1 * 3600
	case "01:00", "+1", "+01", "+01:00", "UTC+1", "UTC+01:00":
		seconds = 1 * 3600
	case "02:00", "+2", "+02", "+02:00", "UTC+2", "UTC+02:00":
		seconds = 2 * 3600
	case "03:00", "+3", "+03", "+03:00", "UTC+3", "UTC+03:00":
		seconds = 3 * 3600
	case "04:00", "+4", "+04", "+04:00", "UTC+4", "UTC+04:00":
		seconds = 4 * 3600
	case "05:00", "+5", "+05", "+05:00", "UTC+5", "UTC+05:00":
		seconds = 5 * 3600
	case "06:00", "+6", "+06", "+06:00", "UTC+6", "UTC+06:00":
		seconds = 6 * 3600
	case "07:00", "+7", "+07", "+07:00", "UTC+7", "UTC+07:00":
		seconds = 7 * 3600
	case "08:00", "+8", "+08", "+08:00", "UTC+8", "UTC+08:00":
		seconds = 8 * 3600
	case "09:00", "+9", "+09", "+09:00", "UTC+9", "UTC+09:00":
		seconds = 9 * 3600
	case "10:00", "+10", "+10:00", "UTC+10", "UTC+10:00":
		seconds = 10 * 3600
	case "11:00", "+11", "+11:00", "UTC+11", "UTC+11:00":
		seconds = 11 * 3600
	case "12:00", "+12", "+12:00", "UTC+12", "UTC+12:00":
		seconds = 12 * 3600
	case "Z", "UTC", "UTC+0", "UTC-0", "UTC+00:00", "UTC-00:00":
		seconds = 0
	default:
		return 0, fmt.Errorf("invalid UTC offset")
	}

	return seconds, nil
}
