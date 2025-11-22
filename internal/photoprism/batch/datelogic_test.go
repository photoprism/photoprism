package batch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComputeDateChange(t *testing.T) {
	t.Run("LeapYearClamping", func(t *testing.T) {
		// Feb 29, 2024 → 2025 should clamp to Feb 28, 2025
		base := time.Date(2024, 2, 29, 11, 29, 54, 0, time.UTC)
		newLocal, y, m, d := ComputeDateChange(base, 2024, 2, 29, ActionNone, 0, ActionNone, 0, ActionUpdate, 2025)

		assert.Equal(t, 2025, y)
		assert.Equal(t, 2, m)
		assert.Equal(t, 28, d)
		assert.Equal(t, 2025, newLocal.Year())
		assert.Equal(t, time.February, newLocal.Month())
		assert.Equal(t, 28, newLocal.Day())
	})
	t.Run("DayUnknown_YearNotForcedUnknown", func(t *testing.T) {
		// Setting Day=-1 should NOT force Year=-1
		base := time.Date(2023, 11, 12, 9, 0, 0, 0, time.UTC)
		newLocal, y, m, d := ComputeDateChange(base, 2023, 11, 12, ActionUpdate, -1, ActionNone, 0, ActionUpdate, 2000)

		assert.Equal(t, 2000, y)
		assert.Equal(t, 11, m)
		assert.Equal(t, -1, d)
		assert.Equal(t, 1, newLocal.Day()) // day used for TakenAtLocal when unknown
	})
	t.Run("MonthUnknown_KeepsYearValue", func(t *testing.T) {
		// Setting Month=-1 should NOT force Year=-1
		base := time.Date(2020, 4, 30, 8, 0, 0, 0, time.UTC)
		newLocal, y, m, d := ComputeDateChange(base, 2020, 4, 30, ActionNone, 0, ActionUpdate, -1, ActionUpdate, 2000)

		assert.Equal(t, 2000, y)
		assert.Equal(t, -1, m)
		assert.Equal(t, 30, d)
		assert.Equal(t, 30, newLocal.Day())
		assert.Equal(t, time.April, newLocal.Month())
	})
	t.Run("MixedMonth_UpdateDayClampsPerPhoto", func(t *testing.T) {
		// Day 31 → March (31 days) should work
		baseMar := time.Date(2020, 3, 20, 11, 29, 54, 0, time.UTC)
		newLocal1, y1, m1, d1 := ComputeDateChange(baseMar, 2020, 3, 20, ActionUpdate, 31, ActionNone, 0, ActionNone, 0)
		assert.Equal(t, 2020, y1)
		assert.Equal(t, 3, m1)
		assert.Equal(t, 31, d1)
		assert.Equal(t, 31, newLocal1.Day())

		// Day 31 → April (30 days) should clamp to 30
		baseApr := time.Date(2020, 4, 20, 11, 29, 54, 0, time.UTC)
		newLocal2, y2, m2, d2 := ComputeDateChange(baseApr, 2020, 4, 20, ActionUpdate, 31, ActionNone, 0, ActionNone, 0)
		assert.Equal(t, 2020, y2)
		assert.Equal(t, 4, m2)
		assert.Equal(t, 30, d2)
		assert.Equal(t, 30, newLocal2.Day())
	})
	t.Run("UnknownCurrentMonth_DayUpdateClampsPerPhoto", func(t *testing.T) {
		// Base in March, current month unknown, update Day=31
		baseMar := time.Date(2020, 3, 20, 11, 29, 54, 0, time.UTC)
		newLocal1, y1, m1, d1 := ComputeDateChange(baseMar, 2020, -1, 20, ActionUpdate, 31, ActionNone, 0, ActionNone, 0)
		assert.Equal(t, 2020, y1)
		assert.Equal(t, -1, m1) // keep unknown
		assert.Equal(t, 31, d1)
		assert.Equal(t, time.March, newLocal1.Month())
		assert.Equal(t, 31, newLocal1.Day())

		// Base in April, current month unknown, update Day=31 → clamp to 30
		baseApr := time.Date(2020, 4, 20, 11, 29, 54, 0, time.UTC)
		newLocal2, y2, m2, d2 := ComputeDateChange(baseApr, 2020, -1, 20, ActionUpdate, 31, ActionNone, 0, ActionNone, 0)
		assert.Equal(t, 2020, y2)
		assert.Equal(t, -1, m2) // keep unknown
		assert.Equal(t, 30, d2) // clamp to Apr 30
		assert.Equal(t, time.April, newLocal2.Month())
		assert.Equal(t, 30, newLocal2.Day())
	})
	t.Run("MixedDay_UpdateMonthToFeb2020", func(t *testing.T) {
		// Photo 1: Day 20 → Feb works fine
		base1 := time.Date(2020, 4, 20, 11, 29, 54, 0, time.UTC)
		newLocal1, y1, m1, d1 := ComputeDateChange(base1, 2020, 4, 20, ActionNone, 0, ActionUpdate, 2, ActionNone, 0)
		assert.Equal(t, 2020, y1)
		assert.Equal(t, 2, m1)
		assert.Equal(t, 20, d1)
		assert.Equal(t, time.February, newLocal1.Month())
		assert.Equal(t, 20, newLocal1.Day())

		// Photo 2: Day 31 → Feb 2020 (leap) should clamp to 29
		base2 := time.Date(2020, 3, 31, 11, 29, 54, 0, time.UTC)
		newLocal2, y2, m2, d2 := ComputeDateChange(base2, 2020, 3, 31, ActionNone, 0, ActionUpdate, 2, ActionNone, 0)
		assert.Equal(t, 2020, y2)
		assert.Equal(t, 2, m2)
		assert.Equal(t, 29, d2)
		assert.Equal(t, time.February, newLocal2.Month())
		assert.Equal(t, 29, newLocal2.Day())
	})
	t.Run("Mixed_UpdateYearTo2021", func(t *testing.T) {
		// Photo 1: Feb 29, 2020 → 2021 should clamp to Feb 28, 2021
		base1 := time.Date(2020, 2, 29, 11, 29, 54, 0, time.UTC)
		newLocal1, y1, m1, d1 := ComputeDateChange(base1, 2020, 2, 29, ActionNone, 0, ActionNone, 0, ActionUpdate, 2021)
		assert.Equal(t, 2021, y1)
		assert.Equal(t, 2, m1)
		assert.Equal(t, 28, d1)
		assert.Equal(t, 2021, newLocal1.Year())
		assert.Equal(t, 28, newLocal1.Day())

		// Photo 2: Mar 31, 2020 → 2021 should stay Mar 31, 2021
		base2 := time.Date(2020, 3, 31, 11, 29, 54, 0, time.UTC)
		newLocal2, y2, m2, d2 := ComputeDateChange(base2, 2020, 3, 31, ActionNone, 0, ActionNone, 0, ActionUpdate, 2021)
		assert.Equal(t, 2021, y2)
		assert.Equal(t, 3, m2)
		assert.Equal(t, 31, d2)
		assert.Equal(t, 2021, newLocal2.Year())
		assert.Equal(t, 31, newLocal2.Day())
	})
	t.Run("AllUnknown_KeepBaseYearMonth_SetDayOne", func(t *testing.T) {
		// Current values: 2024-04-30 13:29:54 (local)
		base := time.Date(2024, 4, 30, 13, 29, 54, 0, time.UTC)

		// User sets Day, Month, Year to unknown (-1)
		newLocal, y, m, d := ComputeDateChange(
			base,
			2024, 4, 30,
			ActionUpdate, -1, // Day -> unknown
			ActionUpdate, -1, // Month -> unknown
			ActionUpdate, -1, // Year -> unknown
		)

		// Output components should be marked unknown
		assert.Equal(t, -1, y)
		assert.Equal(t, -1, m)
		assert.Equal(t, -1, d)

		// TakenAtLocal recomputed using base year/month and day=1, preserving time
		assert.Equal(t, 2024, newLocal.Year())
		assert.Equal(t, time.April, newLocal.Month())
		assert.Equal(t, 1, newLocal.Day())
		assert.Equal(t, 13, newLocal.Hour())
		assert.Equal(t, 29, newLocal.Minute())
		assert.Equal(t, 54, newLocal.Second())
	})
	t.Run("PreservesClockForNonUTCBase", func(t *testing.T) {
		loc := time.FixedZone("UTC-7", -7*3600)
		base := time.Date(2023, 6, 15, 22, 45, 30, 0, loc)
		newLocal, y, m, d := ComputeDateChange(
			base,
			2023, 6, 15,
			ActionUpdate, 18,
			ActionNone, 0,
			ActionNone, 0,
		)

		assert.Equal(t, 2023, y)
		assert.Equal(t, 6, m)
		assert.Equal(t, 18, d)
		// Returned time is UTC but keeps the local clock so SavePhotoForm can reapply the zone.
		assert.Equal(t, time.UTC, newLocal.Location())
		assert.Equal(t, base.Hour(), newLocal.Hour())
		assert.Equal(t, base.Minute(), newLocal.Minute())
		assert.Equal(t, base.Second(), newLocal.Second())
	})
}
