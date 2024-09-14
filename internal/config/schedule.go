package config

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/robfig/cron/v3"

	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	ScheduleDaily  = "daily"
	ScheduleWeekly = "weekly"
)

// Schedule evaluates a schedule config value and returns it, or an empty string if it is invalid. Cron schedules consist
// of 5 space separated values: minute, hour, day of month, month and day of week, e.g. "0 12 * * *" for daily at noon.
func Schedule(s string) string {
	if s == "" {
		return ""
	}

	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case ScheduleDaily:
		return fmt.Sprintf("%d %d * * *", rand.IntN(60), rand.IntN(24))
	case ScheduleWeekly:
		return fmt.Sprintf("%d %d * * 0", rand.IntN(60), rand.IntN(24))
	}

	// Example: "0 12 * * *" stands for daily at noon.
	if _, err := cron.ParseStandard(s); err != nil {
		log.Warnf("config: invalid schedule %s (%s)", clean.Log(s), err)
		return ""
	}

	return s
}
