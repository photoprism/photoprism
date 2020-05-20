package txt

import (
	"regexp"
	"strconv"
	"time"
)

var DateRegexp = regexp.MustCompile("\\d{4}[\\-_]\\d{2}[\\-_]\\d{2}")
var DateCanonicalRegexp = regexp.MustCompile("\\d{8}_\\d{6}_\\w{8}\\.")
var DatePathRegexp = regexp.MustCompile("\\d{4}\\/\\d{1,2}\\/?\\d{0,2}")
var DateTimeRegexp = regexp.MustCompile("\\d{4}[\\-_]\\d{2}[\\-_]\\d{2}.{1,4}\\d{2}\\D\\d{2}\\D\\d{2}")
var DateIntRegexp = regexp.MustCompile("\\d{1,4}")

var (
	YearMin = 1990
	YearMax = time.Now().Year() + 3
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
		if err := recover(); err != nil {
			result = time.Time{}
		}
	}()

	b := []byte(s)

	if found := DateCanonicalRegexp.Find(b); len(found) == 25 { // Is it a canonical name like "20120727_093920_97425909.jpg"?
		if date, err := time.Parse("20060102_150405", string(found[0:15])); err == nil {
			result = date.Round(time.Second).UTC()
		}
	} else if found := DateTimeRegexp.Find(b); len(found) > 0 { // Is it a date with time like "2020-01-30_09-57-18"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) != 6 {
			return result
		}

		year := Int(string(n[0]))
		month := Int(string(n[1]))
		day := Int(string(n[2]))
		hour := Int(string(n[3]))
		min := Int(string(n[4]))
		sec := Int(string(n[5]))

		if year < YearMin || year > YearMax || month < MonthMin || month > MonthMax || day < DayMin || day > DayMax {
			return result
		}

		if hour < HourMin || hour > HourMax || min < MinMin || min > MinMax || sec < SecMin || sec > SecMax {
			return result
		}

		result = time.Date(
			year,
			time.Month(month),
			day,
			hour,
			min,
			sec,
			0,
			time.UTC)

	} else if found := DateRegexp.Find(b); len(found) > 0 { // Is it a date only like "2020-01-30"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) != 3 {
			return result
		}

		year := Int(string(n[0]))
		month := Int(string(n[1]))
		day := Int(string(n[2]))

		if year < YearMin || year > YearMax || month < MonthMin || month > MonthMax || day < DayMin || day > DayMax {
			return result
		}

		result = time.Date(
			year,
			time.Month(month),
			day,
			0,
			0,
			0,
			0,
			time.UTC)
	} else if found := DatePathRegexp.Find(b); len(found) > 0 { // Is it a date path like "2020/01/03"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) < 2 || len(n) > 3 {
			return result
		}

		year := Int(string(n[0]))
		month := Int(string(n[1]))

		if year < YearMin || year > YearMax || month < MonthMin || month > MonthMax {
			return result
		}

		if len(n) == 2 {
			result = time.Date(
				year,
				time.Month(month),
				1,
				0,
				0,
				0,
				0,
				time.UTC)
		} else if day := Int(string(n[2])); day >= DayMin && day <= DayMax {
			result = time.Date(
				year,
				time.Month(month),
				day,
				0,
				0,
				0,
				0,
				time.UTC)
		}
	}

	return result.UTC()
}

// Int returns a string as int or 0 if it can not be converted.
func Int(s string) int {
	if s == "" {
		return 0
	}

	result, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0
	}

	return int(result)
}
