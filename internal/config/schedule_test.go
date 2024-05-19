package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchedule(t *testing.T) {
	assert.Equal(t, "", Schedule(""))
	assert.Equal(t, DefaultIndexSchedule, Schedule(DefaultIndexSchedule))

	// Random default backup schedule.
	backupSchedule := Schedule(DefaultBackupSchedule)
	assert.Equal(t, backupSchedule, Schedule(backupSchedule))

	// Regular backups at a random time (daily or weekly).
	daily := Schedule(ScheduleDaily)
	weekly := Schedule(ScheduleWeekly)

	assert.Equal(t, daily, Schedule(daily))
	assert.Equal(t, weekly, Schedule(weekly))
}
