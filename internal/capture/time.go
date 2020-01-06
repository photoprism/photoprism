package capture

import (
	"time"
)

func Time(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf("%s [%s]", name, elapsed)
}
