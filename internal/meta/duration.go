package meta

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var DurationSecondsRegexp = regexp.MustCompile("[0-9\\.]+")

// StringToDuration converts a metadata string to a valid duration.
func StringToDuration(s string) (d time.Duration) {
	if s == "" {
		return d
	}

	s = strings.TrimSpace(s)

	if s == "" {
		return d
	}

	sec := DurationSecondsRegexp.FindAllString(s, -1)

	if len(sec) == 1 {
		secFloat, _ := strconv.ParseFloat(sec[0], 64)
		d = time.Duration(secFloat) * time.Second
	} else if n := strings.Split(s, ":"); len(n) == 3 {
		h, _ := strconv.Atoi(n[0])
		m, _ := strconv.Atoi(n[1])
		s, _ := strconv.Atoi(n[2])

		d = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
	} else if pd, err := time.ParseDuration(s); err != nil {
		d = pd
	}

	return d
}
