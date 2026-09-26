package txt

import (
	"strings"
	"time"
)

// DateFromFilePath returns a string as time or the zero time instant in case it can not be converted.
func DateFromFilePath(s string) (result time.Time) {
	defer func() {
		if r := recover(); r != nil {
			result = time.Time{}
		}
	}()

	if len(s) < 6 {
		return time.Time{}
	}

	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}

	// Is it a camera file name with a compact date and time like "IMG_20190101_120000.jpg"?
	if taken := dateFromCameraName(s[strings.LastIndex(s, "/")+1:]); !taken.IsZero() {
		return taken
	}

	b := []byte(s)

	if found := DateTimeRegexp.Find(b); len(found) > 0 { // Is it a date with time like "2020-01-30_09-57-18"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) < 6 {
			return result
		}

		year := ExpandYear(string(n[0]))
		month := Int(string(n[1]))
		day := Int(string(n[2]))
		hour := Int(string(n[3]))
		minute := Int(string(n[4]))
		sec := Int(string(n[5]))

		// Perform date plausibility check.
		if year < YearMin || year > YearMax || month < MonthMin || month > MonthMax || day < DayMin || day > DayMax {
			return result
		}

		// Perform time plausibility check.
		if hour < HourMin || hour > HourMax || minute < MinMin || minute > MinMax || sec < SecMin || sec > SecMax {
			return result
		}

		result = time.Date(
			year,
			time.Month(month),
			day,
			hour,
			minute,
			sec,
			0,
			time.UTC)

	} else if found = DateRegexp.Find(b); len(found) > 0 { // Is it a date only like "2020-01-30"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) != 3 {
			return result
		}

		year := ExpandYear(string(n[0]))
		month := Int(string(n[1]))
		day := Int(string(n[2]))

		// Perform date plausibility check.
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
	} else if found = DatePathRegexp.Find(b); len(found) > 0 { // Is it a date path like "2020/01/03"?
		n := DateIntRegexp.FindAll(found, -1)

		if len(n) < 2 || len(n) > 3 {
			return result
		}

		year := ExpandYear(string(n[0]))
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
	} else if found = DateWhatsAppRegexp.Find(b); len(found) > 0 { // Is it a WhatsApp date path like "VID-20191120-WA0001.jpg"?
		match := DateWhatsAppRegexp.FindSubmatch(b)

		if len(match) != 4 {
			return result
		}

		matchMap := make(map[string]string)
		for i, name := range DateWhatsAppRegexp.SubexpNames() {
			if i != 0 {
				matchMap[name] = string(match[i])
			}
		}

		year := ExpandYear(matchMap["year"])
		month := Int(matchMap["month"])
		day := Int(matchMap["day"])

		// Perform date plausibility check.
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
	}

	return result.UTC()
}

// dateFromCameraName returns the date and time encoded in a camera file name, or the zero time if the
// name does not start with IMG_, VID_ or LRV_ followed by a plausible compact date and time.
func dateFromCameraName(base string) time.Time {
	m := DateCameraRegexp.FindStringSubmatch(base)

	if len(m) != 7 {
		return time.Time{}
	}

	year, month, day := Int(m[1]), Int(m[2]), Int(m[3])
	hour, minute, sec := Int(m[4]), Int(m[5]), Int(m[6])

	if year < YearMin || year > YearMax || month < MonthMin || month > MonthMax || day < DayMin || day > DayMax ||
		hour < HourMin || hour > HourMax || minute < MinMin || minute > MinMax || sec < SecMin || sec > SecMax {
		return time.Time{}
	}

	result := time.Date(year, time.Month(month), day, hour, minute, sec, 0, time.UTC)

	// Reject dates that do not exist, such as February 30, and device defaults of an unset clock.
	if result.Day() != day || DateTimeDefault(result.Format(time.DateTime)) {
		return time.Time{}
	}

	return result
}
