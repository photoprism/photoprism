package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

// TrustedTime tests if the photo has a known date and time from a trusted source.
func (m *Photo) TrustedTime() bool {
	if SrcPriority[m.TakenSrc] <= SrcPriority[SrcEstimate] {
		return false
	} else if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		return false
	} else if m.TimeZone == "" {
		return false
	}

	return true
}

// SetTakenAt changes the photo date if not empty and from the same source.
func (m *Photo) SetTakenAt(taken, local time.Time, zone, source string) {
	if taken.IsZero() || taken.Year() < 1000 || taken.Year() > txt.YearMax {
		return
	}

	if SrcPriority[source] < SrcPriority[m.TakenSrc] && !m.TakenAt.IsZero() {
		return
	}

	// Remove time zone if time was extracted from file name.
	if source == SrcName {
		zone = ""
	}

	// Round times to avoid jitter.
	taken = taken.UTC().Truncate(time.Second)

	// Default local time to taken if zero or invalid.
	if local.IsZero() || local.Year() < 1000 {
		local = taken
	} else {
		local = local.Truncate(time.Second)
	}

	// Don't update older date.
	if SrcPriority[source] <= SrcPriority[SrcAuto] && !m.TakenAt.IsZero() && taken.After(m.TakenAt) {
		return
	}

	// Set UTC time and date source.
	m.TakenAt = taken
	m.TakenAtLocal = local
	m.TakenSrc = source

	if zone == time.UTC.String() && m.TimeZone != "" {
		// Location exists, set local time from UTC.
		m.TakenAtLocal = m.GetTakenAtLocal()
	} else if zone != "" {
		// Apply new time zone.
		m.TimeZone = zone
		m.TakenAt = m.GetTakenAt()
	} else if m.TimeZoneUTC() {
		m.TimeZone = zone
		// Keep UTC?
		if m.TimeZoneUTC() {
			m.TakenAtLocal = taken
		}
	} else if m.TimeZone != "" {
		// Apply existing time zone.
		m.TakenAt = m.GetTakenAt()
	}

	m.UpdateDateFields()
}

// TimeZoneUTC tests if the current time zone is UTC.
func (m *Photo) TimeZoneUTC() bool {
	return strings.EqualFold(m.TimeZone, time.UTC.String())
}

// UpdateTimeZone updates the time zone.
func (m *Photo) UpdateTimeZone(zone string) {
	if zone == "" || zone == time.UTC.String() || zone == m.TimeZone {
		return
	}

	if SrcPriority[m.TakenSrc] >= SrcPriority[SrcManual] && m.TimeZone != "" {
		return
	}

	if m.TimeZoneUTC() {
		m.TimeZone = zone
		m.TakenAtLocal = m.GetTakenAtLocal()
	} else {
		m.TimeZone = zone
		m.TakenAt = m.GetTakenAt()
	}
}

// UpdateDateFields updates internal date fields.
func (m *Photo) UpdateDateFields() {
	if m.TakenAt.IsZero() || m.TakenAt.Year() < 1000 {
		return
	}

	if m.TakenAtLocal.IsZero() || m.TakenAtLocal.Year() < 1000 {
		m.TakenAtLocal = m.TakenAt
	}

	// Set date to unknown if file system date is about the same as indexing time.
	if m.TakenSrc == SrcAuto && m.TakenAt.After(m.CreatedAt.Add(-24*time.Hour)) {
		m.PhotoYear = UnknownYear
		m.PhotoMonth = UnknownMonth
		m.PhotoDay = UnknownDay
	} else if m.TakenSrc != SrcManual {
		m.PhotoYear = m.TakenAtLocal.Year()
		m.PhotoMonth = int(m.TakenAtLocal.Month())
		m.PhotoDay = m.TakenAtLocal.Day()
	}

	// Update photo_taken_at column in related files.
	Log("photo", "update date fields",
		UnscopedDb().Model(File{}).
			Where("photo_id = ? AND photo_taken_at <> ?", m.ID, m.TakenAtLocal).
			Updates(File{PhotoTakenAt: m.TakenAtLocal}).Error,
	)
}
