package batch

import "time"

// ComputeDateChange calculates a new date and output values based on the provided actions and values.
// The returned time is deliberately constructed in UTC: we store TakenAtLocal as a naive timestamp and
// later derive the real UTC instant via Photo.GetTakenAt() using the photo's TimeZone. See SavePhotoForm
// for details on how TakenAt / TakenAtLocal stay consistent in the database.
func ComputeDateChange(
	baseLocal time.Time,
	curYear, curMonth, curDay int,
	dayAct Action, dayVal int,
	monthAct Action, monthVal int,
	yearAct Action, yearVal int,
) (newLocal time.Time, outYear, outMonth, outDay int) {
	outYear = curYear
	outMonth = curMonth
	outDay = curDay

	year := baseLocal.Year()
	if yearAct == ActionUpdate && yearVal > 0 {
		year = yearVal
	}

	month := int(baseLocal.Month())
	if monthAct == ActionUpdate && monthVal > 0 {
		month = monthVal
	}

	day := baseLocal.Day()
	if dayAct == ActionUpdate {
		if dayVal == -1 {
			day = 1
		} else if dayVal > 0 {
			day = dayVal
		}
	}

	// Clamp to last valid day of target month/year.
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}

	newLocal = time.Date(year, time.Month(month), day, baseLocal.Hour(), baseLocal.Minute(), baseLocal.Second(), 0, time.UTC)

	if yearAct == ActionUpdate {
		outYear = yearVal
	}
	if monthAct == ActionUpdate {
		outMonth = monthVal
	}

	if dayAct == ActionUpdate {
		if dayVal == -1 {
			outDay = -1
		} else if dayVal > 0 {
			outDay = day
		}
	} else {
		outDay = day
	}

	return newLocal, outYear, outMonth, outDay
}
