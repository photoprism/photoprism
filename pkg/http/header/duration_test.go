package header

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *testing.T) {
	var (
		day   = time.Hour * 24
		week  = day * 7
		month = day * 31
		year  = day * 365
	)

	assert.Equal(t, int(day.Seconds()), DurationDay)
	assert.Equal(t, int(week.Seconds()), DurationWeek)
	assert.Equal(t, int(month.Seconds()), DurationMonth)
	assert.Equal(t, int(year.Seconds()), DurationYear)
}
