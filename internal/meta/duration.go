package meta

import (
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Duration converts a metadata string to a valid duration.
func Duration(s string) (result time.Duration) {
	if s == "" {
		return 0
	}

	s = clean.Duration(s)

	if txt.IsFloat(s) {
		result = time.Duration(txt.Float(s) * 1e9)
	} else if n := strings.Split(strings.TrimSpace(s), ":"); len(n) == 3 {
		hr, _ := strconv.Atoi(n[0])
		min, _ := strconv.Atoi(n[1])
		sec, _ := strconv.Atoi(n[2])

		result = time.Duration(hr)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
	} else if d, err := time.ParseDuration(s); err == nil {
		result = d
	}

	return result.Round(10e6)
}
