package maps

import "strings"

// CountryName tries to find a matching country name for a code.
func CountryName(code string) string {
	if code == "" {
		code = "zz"
	} else {
		code = strings.ToLower(code)
	}

	if name, ok := CountryNames[code]; ok {
		return name
	}

	return CountryNames["zz"]
}
