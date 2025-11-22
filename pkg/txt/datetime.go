package txt

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/time/tz"
)

// regex tester: https://regoio.herokuapp.com/

// DateRegexp matches generic date patterns in strings.
var DateRegexp = regexp.MustCompile(`\D\d{4}[\-_]\d{2}[\-_]\d{2,}`)

// DatePathRegexp matches date-like path fragments.
var DatePathRegexp = regexp.MustCompile(`\D\d{4}/\d{1,2}/?\d*`)

// DateTimeRegexp matches datetime patterns in arbitrary strings.
var DateTimeRegexp = regexp.MustCompile(`\D\d{2,4}[\-_]\d{2}[\-_]\d{2}.{1,4}\d{2}\D\d{2}\D\d{2,}`)

// DateWhatsAppRegexp matches WhatsApp media filenames.
var DateWhatsAppRegexp = regexp.MustCompile(`(?:IMG|VID)-(?P<year>\d{4})(?P<month>\d{2})(?P<day>\d{2})-WA`)

// DateIntRegexp matches short numeric date parts.
var DateIntRegexp = regexp.MustCompile(`\d{1,4}`)

// YearRegexp matches 4–5 digit years.
var YearRegexp = regexp.MustCompile(`\d{4,5}`)

// IsDateRegexp detects yyyy-mm-dd style strings.
var IsDateRegexp = regexp.MustCompile(`\d{4}[\-_]?\d{2}[\-_]?\d{2}`)

// IsDateTimeRegexp detects yyyy-mm-dd hh:mm:ss style strings.
var IsDateTimeRegexp = regexp.MustCompile(`\d{4}[\-_]?\d{2}[\-_]?\d{2}.{1,4}\d{2}\D?\d{2}\D?\d{2}`)

// HumanDateTimeRegexp parses human-readable timestamps.
var HumanDateTimeRegexp = regexp.MustCompile(`((?P<day>\d{2})|\D{2})\D((?P<month>\d{2})|\D{2})\D((?P<year>\d{4})|\D{4})\D((?P<h>\d{2})|\D{2})\D((?P<m>\d{2})|\D{2})\D((?P<s>\d{2})|\D{2})(\.(?P<subsec>\d+))?(?P<z>\D)?(?P<zh>\d{2})?\D?(?P<zm>\d{2})?`)

// HumanDateTimeMatch stores named capture group indexes for human datetime regex.
var HumanDateTimeMatch = make(map[string]int)

// ExifDateTimeRegexp parses EXIF-style datetime strings.
var ExifDateTimeRegexp = regexp.MustCompile(`((?P<year>\d{4})|\D{4})\D((?P<month>\d{2})|\D{2})\D((?P<day>\d{2})|\D{2})\D((?P<h>\d{2})|\D{2})\D((?P<m>\d{2})|\D{2})\D((?P<s>\d{2})|\D{2})(\.(?P<subsec>\d+))?(?P<z>\D)?(?P<zh>\d{2})?\D?(?P<zm>\d{2})?`)

// ExifDateTimeMatch stores named capture group indexes for EXIF datetime regex.
var ExifDateTimeMatch = make(map[string]int)

// OneYear represents a duration of 365 days.
const OneYear = time.Hour * 24 * 365

func init() {
	en := ExifDateTimeRegexp.SubexpNames()
	for i := 0; i < len(en); i++ {
		if name := en[i]; name != "" {
			ExifDateTimeMatch[name] = i
		}
	}

	hn := HumanDateTimeRegexp.SubexpNames()
	for i := 0; i < len(hn); i++ {
		if name := hn[i]; name != "" {
			HumanDateTimeMatch[name] = i
		}
	}
}

var (
	// YearMin is the minimum acceptable year.
	YearMin = 1970
	// YearMinShort is the minimum 2-digit year treated as 1900s.
	YearMinShort = 90
	// YearMax is the maximum acceptable year (3 years in future).
	YearMax = time.Now().Add(OneYear * 3).Year()
	// YearShort is the current 2-digit year.
	YearShort = Int(time.Now().Format("06"))
)

const (
	// MonthMin is the minimum allowed month.
	MonthMin = 1
	// MonthMax is the maximum allowed month.
	MonthMax = 12
	// DayMin is the minimum day in a month.
	DayMin = 1
	// DayMax is the maximum day in a month.
	DayMax = 31
	// HourMin is the minimum hour in a day.
	HourMin = 0
	// HourMax is the maximum hour in a day.
	HourMax = 24
	// MinMin is the minimum minute in an hour.
	MinMin = 0
	// MinMax is the maximum minute in an hour.
	MinMax = 59
	// SecMin is the minimum second in a minute.
	SecMin = 0
	// SecMax is the maximum second in a minute.
	SecMax = 59
)

// IsTime tests if the string looks like a date and/or time.
func IsTime(s string) bool {
	if s == "" {
		return false
	} else if m := IsDateRegexp.FindString(s); m == s {
		return true
	} else if m = IsDateTimeRegexp.FindString(s); m == s {
		return true
	}

	return false
}

// DateTime formats a time pointer as a human-readable datetime string.
func DateTime(t *time.Time) string {
	if t == nil {
		return ""
	} else if t.IsZero() {
		return ""
	}

	return t.UTC().Format("2006-01-02 15:04:05")
}

// UnixTime formats a unix time as a human-readable datetime string.
func UnixTime(t int64) string {
	if t == 0 {
		return ""
	}

	timeStamp := time.Unix(t, 0)

	if timeStamp.IsZero() {
		return ""
	}

	return timeStamp.UTC().Format("2006-01-02 15:04:05")
}

// ParseTimeUTC parses a UTC timestamp and returns a valid time.Time if possible.
func ParseTimeUTC(s string) (t time.Time) {
	return ParseTime(s, "")
}

// ParseTime parses a timestamp and returns a valid time.Time if possible.
func ParseTime(s, timeZone string) (t time.Time) {
	defer func() {
		if r := recover(); r != nil {
			// Panic? Return unknown time.
			t = time.Time{}
		}
	}()

	// Ignore defaults.
	if DateTimeDefault(s) {
		return time.Time{}
	}

	s = strings.TrimLeft(s, " ")

	// Timestamp too short?
	if len(s) < 4 {
		return time.Time{}
	} else if len(s) > 50 {
		// Clip to max length.
		s = s[:50]
	}

	// Pad short timestamp with whitespace at the end.
	s = fmt.Sprintf("%-19s", s)

	var v map[string]int

	m := ExifDateTimeRegexp.FindStringSubmatch(s)

	// Pattern doesn't match? Return unknown time.
	if len(m) == 0 {
		if m = HumanDateTimeRegexp.FindStringSubmatch(s); len(m) == 0 {
			return time.Time{}
		} else {
			v = HumanDateTimeMatch
		}
	} else {
		v = ExifDateTimeMatch
	}

	// Ignore timestamps without year, month, and day.
	if Int(m[v["year"]]) == 0 && Int(m[v["month"]]) == 0 && Int(m[v["day"]]) == 0 {
		return time.Time{}
	}

	loc := tz.Find(timeZone)
	zone := tz.TimeUTC

	// Location found?
	if timeZone != "" {
		zone = loc
		timeZone = loc.String()
	}

	// Does the timestamp contain a time zone offset?
	z := m[v["z"]]                     // Supported values, if not empty: Z, +, -
	zh := IntVal(m[v["zh"]], 0, 23, 0) // Hours.
	zm := IntVal(m[v["zm"]], 0, 59, 0) // Minutes.

	// Valid time zone offset found?
	if offset := (zh*60 + zm) * 60; offset > 0 && offset <= 86400 {
		// Offset timezone name example: UTC+03:30
		switch z {
		case "+":
			// Positive offset relative to UTC.
			zone = time.FixedZone(fmt.Sprintf("UTC+%02d:%02d", zh, zm), offset)
		case "-":
			// Negative offset relative to UTC.
			zone = time.FixedZone(fmt.Sprintf("UTC-%02d:%02d", zh, zm), -1*offset)
		}
	}

	var ns int

	if subSec := m[v["subsec"]]; subSec != "" {
		ns = Int(subSec + strings.Repeat("0", 9-len(subSec)))
	} else {
		ns = 0
	}

	// Create rounded timestamp from parsed input values.
	// Year 0 is treated separately as it has a special meaning in exiftool. Golang
	// does not seem to accept value 0 for the year, but considers a date to be
	// "zero" when year is 1.
	year := IntVal(m[v["year"]], 0, YearMax, time.Now().Year())
	if year == 0 {
		year = 1
	}

	month := IntVal(m[v["month"]], 1, 12, 1)
	day := IntVal(m[v["day"]], 1, 31, 1)

	t = time.Date(
		year,
		time.Month(month),
		day,
		IntVal(m[v["h"]], 0, 23, 0),
		IntVal(m[v["m"]], 0, 59, 0),
		IntVal(m[v["s"]], 0, 59, 0),
		ns,
		zone)

	if timeZone != "" && timeZone != tz.Local && loc != zone {
		return t.In(loc)
	}

	return t
}
