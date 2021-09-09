package txt

import (
	"regexp"
	"strings"
	"time"
)

var DateRegexp = regexp.MustCompile("\\D\\d{4}[\\-_]\\d{2}[\\-_]\\d{2,}")
var DatePathRegexp = regexp.MustCompile("\\D\\d{4}/\\d{1,2}/?\\d*")
var DateFilenameRegexp = regexp.MustCompile("\\d{4}\\d{2}\\d{2}")
var DateTimeRegexp = regexp.MustCompile("\\D\\d{4}[\\-_]\\d{2}[\\-_]\\d{2}.{1,4}\\d{2}\\D\\d{2}\\D\\d{2,}")
var DateIntRegexp = regexp.MustCompile("\\d{1,4}")
var YearRegexp = regexp.MustCompile("\\d{4,5}")
var IsDateRegexp = regexp.MustCompile("\\d{4}[\\-_]?\\d{2}[\\-_]?\\d{2}")
var IsDateTimeRegexp = regexp.MustCompile("\\d{4}[\\-_]?\\d{2}[\\-_]?\\d{2}.{1,4}\\d{2}\\D?\\d{2}\\D?\\d{2}")

var (
	YearMin = 1990
	YearMax = time.Now().Year() + 3
)

var (
	DefaultTime = time.Time{}
)

const (
	MonthMin = 1
	MonthMax = 12
	DayMin   = 1
	DayMax   = 31
	HourMin  = 0
	HourMax  = 24
	MinMin   = 0
	MinMax   = 59
	SecMin   = 0
	SecMax   = 59
)

// Time returns a string as time or the zero time instant in case it can not be converted.
func Time(s string) (result time.Time) {
	defer func() {
		if r := recover(); r != nil {
			result = DefaultTime
		}
	}()

	if len(s) < 6 {
		return DefaultTime
	}

	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}

	b := []byte(s)
	handlers := []func([]byte) time.Time{
		DateTimeHandler,
		DateHandler,
		DatePathHandler,
		DateFilenameHandler,
	}

	for _, handler := range handlers {
		result = handler(b)
		if !result.IsZero() {
			break
		}
	}

	return result.UTC()
}

// DateTimeHandler checks whether the data contains a datetime like "2020-01-30_09-57-18"
func DateTimeHandler(b []byte) time.Time {
	found := DateTimeRegexp.Find(b)

	if len(found) == 0 {
		return DefaultTime
	}

	n := DateIntRegexp.FindAll(found, -1)

	if len(n) != 6 {
		return DefaultTime
	}

	year := Int(string(n[0]))
	month := Int(string(n[1]))
	day := Int(string(n[2]))
	hour := Int(string(n[3]))
	min := Int(string(n[4]))
	sec := Int(string(n[5]))

	if !IsValidDate(year, month, day) || !IsValidTime(hour, min, sec) {
		return DefaultTime
	}

	return time.Date(
		year,
		time.Month(month),
		day,
		hour,
		min,
		sec,
		0,
		time.UTC)
}

// DateHandler checks whether the data contains a date like "2020-01-30"
func DateHandler(b []byte) time.Time {
	found := DateRegexp.Find(b)

	if len(found) == 0 {
		return DefaultTime
	}

	n := DateIntRegexp.FindAll(found, -1)

	if len(n) != 3 {
		return DefaultTime
	}

	year := Int(string(n[0]))
	month := Int(string(n[1]))
	day := Int(string(n[2]))

	if !IsValidDate(year, month, day) {
		return DefaultTime
	}

	return time.Date(
		year,
		time.Month(month),
		day,
		0,
		0,
		0,
		0,
		time.UTC)
}

// DatePathHandler checks whether the data contains a date like "2020/01/03"
func DatePathHandler(b []byte) time.Time {
	found := DatePathRegexp.Find(b)

	if len(found) == 0 {
		return DefaultTime
	}

	n := DateIntRegexp.FindAll(found, -1)

	if len(n) < 2 || len(n) > 3 {
		return DefaultTime
	}

	year := Int(string(n[0]))
	month := Int(string(n[1]))

	if !IsValidYearAndMonth(year, month) {
		return DefaultTime
	}

	if len(n) == 2 {
		return time.Date(
			year,
			time.Month(month),
			1,
			0,
			0,
			0,
			0,
			time.UTC)
	} else if day := Int(string(n[2])); IsValidDay(day) {
		return time.Date(
			year,
			time.Month(month),
			day,
			0,
			0,
			0,
			0,
			time.UTC)
	}

	return DefaultTime
}

// DateFilenameHandler checks whether the data contains a date like "20200130"
func DateFilenameHandler(b []byte) time.Time {
	found := DateFilenameRegexp.Find(b)

	if len(found) == 0 {
		return DefaultTime
	}

	match := string(found)

	year := Int(match[0:4])
	month := Int(match[4:6])
	day := Int(match[6:8])

	if !IsValidDate(year, month, day) {
		return DefaultTime
	}

	return time.Date(
		year,
		time.Month(month),
		day,
		0,
		0,
		0,
		0,
		time.UTC)
}

// IsValidTime tests if the given time is plausible
func IsValidTime(hour int, min int, sec int) bool {
	return IsValidHour(hour) && IsValidMinute(min) && IsValidSecond(sec)
}

// IsValidHour tests if the given hour value is plausible
func IsValidHour(hour int) bool {
	return hour >= HourMin && hour <= HourMax
}

// IsValidMinute tests if the given minutes value is plausible
func IsValidMinute(minute int) bool {
	return minute >= MinMin && minute <= MinMax
}

// IsValidSecond tests if the given second value is plausible
func IsValidSecond(second int) bool {
	return second >= SecMin && second <= SecMax
}

// IsValidDate tests if the given date is plausible
func IsValidDate(year int, month int, day int) bool {
	return IsValidYearAndMonth(year, month) && IsValidDay(day)
}

// IsValidYearAndMonth tests if the given year and month are plausible
func IsValidYearAndMonth(year int, month int) bool {
	return IsValidYear(year) && IsValidMonth(month)
}

// IsValidYear tests if the given year is plausible
func IsValidYear(year int) bool {
	return year >= YearMin && year <= YearMax
}

// IsValidMonth tests if the given month is plausible
func IsValidMonth(month int) bool {
	return month >= MonthMin && month <= MonthMax
}

// IsValidDay tests if the given day is plausible
func IsValidDay(day int) bool {
	return day >= DayMin && day <= DayMax
}

// IsTime tests if the string looks like a date and/or time.
func IsTime(s string) bool {
	if s == "" {
		return false
	} else if m := IsDateRegexp.FindString(s); m == s {
		return true
	} else if m := IsDateTimeRegexp.FindString(s); m == s {
		return true
	}

	return Time(s) != DefaultTime
}

// Year tries to find a matching year for a given string e.g. from a file oder directory name.
func Year(s string) int {
	b := []byte(s)

	found := YearRegexp.FindAll(b, -1)

	for _, match := range found {
		year := Int(string(match))

		if year > YearMin && year < YearMax {
			return year
		}
	}

	return 0
}
