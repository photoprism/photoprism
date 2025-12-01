package tz

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Offset returns the UTC time offset in seconds or an error if it is invalid.
func Offset(utcOffset string) (seconds int, err error) {
	switch utcOffset {
	case "Z", "UTC", "UTC+0", "UTC-0", "UTC+00:00", "UTC-00:00":
		seconds = 0
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
	default:
		return 0, fmt.Errorf("invalid UTC offset")
	}

	return seconds, nil
}

// NormalizeUtcOffset returns a normalized UTC time offset string.
func NormalizeUtcOffset(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	switch s {
	case "0", "Z", "UTC", "Etc/UTC", "Etc/UTC+0", "Etc/UTC+00", "Etc/UTC+00:00", "UTC+0", "UTC-0", "UTC+00", "UTC-00", "UTC+00:00", "UTC-00:00", "Zulu", "Etc/Zulu":
		return UTC
	case "Etc/GMT", "Etc/GMT+00:00", "Etc/GMT-00:00", "Etc/GMT+00", "Etc/GMT-00", "Etc/GMT+0", "GMT+0", "GMT", "gmt":
		return GMT
	case "Etc/GMT+01:00", "Etc/GMT+01", "GMT+1", "01:00", "+1", "+01", "+01:00", "UTC+1", "UTC+01:00":
		return "UTC+1"
	case "Etc/GMT+02:00", "Etc/GMT+02", "GMT+2", "02:00", "+2", "+02", "+02:00", "UTC+2", "UTC+02:00":
		return "UTC+2"
	case "Etc/GMT+03:00", "Etc/GMT+03", "GMT+3", "03:00", "+3", "+03", "+03:00", "UTC+3", "UTC+03:00":
		return "UTC+3"
	case "Etc/GMT+04:00", "Etc/GMT+04", "GMT+4", "04:00", "+4", "+04", "+04:00", "UTC+4", "UTC+04:00":
		return "UTC+4"
	case "Etc/GMT+05:00", "Etc/GMT+05", "GMT+5", "05:00", "+5", "+05", "+05:00", "UTC+5", "UTC+05:00":
		return "UTC+5"
	case "Etc/GMT+05:45", "GMT+05:45", "UTC+05:45", "Z+05:45":
		return AsiaKathmandu
	case "Etc/GMT+06:00", "Etc/GMT+06", "GMT+6", "06:00", "+6", "+06", "+06:00", "UTC+6", "UTC+06:00":
		return "UTC+6"
	case "Etc/GMT+07:00", "Etc/GMT+07", "GMT+7", "07:00", "+7", "+07", "+07:00", "UTC+7", "UTC+07:00":
		return "UTC+7"
	case "Etc/GMT+08:00", "Etc/GMT+08", "GMT+8", "08:00", "+8", "+08", "+08:00", "UTC+8", "UTC+08:00":
		return "UTC+8"
	case "Etc/GMT+09:00", "Etc/GMT+09", "GMT+9", "09:00", "+9", "+09", "+09:00", "UTC+9", "UTC+09:00":
		return "UTC+9"
	case "Etc/GMT+10:00", "Etc/GMT+10", "GMT+10", "10:00", "+10", "+10:00", "UTC+10", "UTC+10:00":
		return "UTC+10"
	case "Etc/GMT+11:00", "Etc/GMT+11", "GMT+11", "11:00", "+11", "+11:00", "UTC+11", "UTC+11:00":
		return "UTC+11"
	case "Etc/GMT+12:00", "Etc/GMT+12", "GMT+12", "12:00", "+12", "+12:00", "UTC+12", "UTC+12:00":
		return "UTC+12"
	case "Etc/GMT-12:00", "Etc/GMT-12", "GMT-12", "-12", "-12:00", "UTC-12", "UTC-12:00":
		return "UTC-12"
	case "Etc/GMT-11:00", "Etc/GMT-11", "GMT-11", "-11", "-11:00", "UTC-11", "UTC-11:00":
		return "UTC-11"
	case "Etc/GMT-10:00", "Etc/GMT-10", "GMT-10", "-10", "-10:00", "UTC-10", "UTC-10:00":
		return "UTC-10"
	case "Etc/GMT-09:00", "Etc/GMT-09", "GMT-9", "-9", "-09", "-09:00", "UTC-9", "UTC-09:00":
		return "UTC-9"
	case "Etc/GMT-08:00", "Etc/GMT-08", "GMT-8", "-8", "-08", "-08:00", "UTC-8", "UTC-08:00":
		return "UTC-8"
	case "Etc/GMT-07:00", "Etc/GMT-07", "GMT-7", "-7", "-07", "-07:00", "UTC-7", "UTC-07:00":
		return "UTC-7"
	case "Etc/GMT-06:00", "Etc/GMT-06", "GMT-6", "-6", "-06", "-06:00", "UTC-6", "UTC-06:00":
		return "UTC-6"
	case "Etc/GMT-05:00", "Etc/GMT-05", "GMT-5", "-5", "-05", "-05:00", "UTC-5", "UTC-05:00":
		return "UTC-5"
	case "Etc/GMT-04:00", "Etc/GMT-04", "GMT-4", "-4", "-04", "-04:00", "UTC-4", "UTC-04:00":
		return "UTC-4"
	case "Etc/GMT-03:00", "Etc/GMT-03", "GMT-3", "-3", "-03", "-03:00", "UTC-3", "UTC-03:00":
		return "UTC-3"
	case "Etc/GMT-02:00", "Etc/GMT-02", "GMT-2", "-2", "-02", "-02:00", "UTC-2", "UTC-02:00":
		return "UTC-2"
	case "Etc/GMT-01:00", "Etc/GMT-01", "GMT-1", "-1", "-01", "-01:00", "UTC-1", "UTC-01:00":
		return "UTC-1"
	}

	return ""
}

// UtcOffset returns the time difference as UTC offset string.
func UtcOffset(utc, local time.Time, offset string) string {
	if offset = NormalizeUtcOffset(offset); offset != "" {
		return offset
	}
	if utc.IsZero() || local.Equal(utc) {
		return ""
	}

	utc = utc.Truncate(time.Second)

	if local.IsZero() {
		if _, sec := utc.Zone(); sec == 0 {
			return ""
		} else if h := sec / 3600; h != 0 && h >= -12 && h <= 12 {
			return fmt.Sprintf("UTC%+d", h)
		}
	}

	local = local.Truncate(time.Second)

	d := local.Sub(utc).Hours()

	// Return if time difference includes fractions of an hour.
	if math.Abs(d-float64(int64(d))) > 0.1 {
		return ""
	}

	// Check if time difference is within expected range (hours).
	if h := int(d); h != 0 && h >= -12 && h <= 12 {
		return fmt.Sprintf("UTC%+d", h)
	}

	return ""
}
