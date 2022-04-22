package txt

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

// ExpandYear converts a string to a year and expands two-digit years if possible.
func ExpandYear(s string) int {
	l := len(s)

	if l < 2 || l == 3 || l > 4 {
		return -1
	} else if !IsUInt(s) {
		return -1
	}

	year := Int(s)

	if l == 4 {
		return year
	} else if year >= 1 && year <= YearShort {
		year = year + 2000
	} else if year >= YearMinShort && year < 100 {
		year = year + 1900
	} else {
		return -1
	}

	return year
}
