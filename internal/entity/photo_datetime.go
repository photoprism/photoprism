package entity

import (
	"time"

	"github.com/photoprism/photoprism/pkg/time/tz"
	"github.com/photoprism/photoprism/pkg/txt"
)

// TrustedTime tests if the photo has a known date and time from a trusted source.
func (m *Photo) TrustedTime() bool {
	if SrcPriority[m.TakenSrc] <= SrcPriority[SrcEstimate] {
		return false
	} else if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		return false
	} else if tz.Name(m.TimeZone) == tz.Local {
		return false
	}

	return true
}

// SetTakenAt changes the photo date if not empty and from the same source.
func (m *Photo) SetTakenAt(utc, local time.Time, zone, source string) {
	if utc.IsZero() || utc.Year() < 1000 || utc.Year() > txt.YearMax {
		return
	}

	// Prevent the existing time from being overwritten by lower priority sources.
	if SrcPriority[source] < SrcPriority[m.TakenSrc] && !m.TakenAt.IsZero() {
		return
	}

	// Normalize time zone string.
	zone = tz.Name(zone)

	// Ignore sub-seconds to avoid jitter.
	utc = tz.TruncateUTC(utc)

	// Default local time to taken if zero or invalid.
	if local.IsZero() || local.Year() < 1000 {
		local = utc
	} else {
		local = tz.TruncateLocal(local)
	}

	// If no zone is specified, assume the current zone or try to determine
	// the time zone based on the time offset. Otherwise, default to Local.
	if source == SrcName && tz.Name(zone) == tz.Local && tz.Name(m.TimeZone) == tz.Local {
		// Assume Local timezone if the time was extracted from a filename.
		zone = tz.Local
	} else if zone == tz.Unknown {
		if m.TimeZone != tz.Unknown {
			zone = m.TimeZone
		} else if !utc.Equal(local) {
			zone = tz.UtcOffset(utc, local, "")
		}

		if zone == tz.Unknown {
			zone = tz.Local
		}
	}

	// Don't update older date.
	if SrcPriority[source] <= SrcPriority[SrcAuto] && !m.TakenAt.IsZero() && utc.After(m.TakenAt) {
		return
	}

	// Use location time zone if it has a higher priority.
	if SrcPriority[source] < SrcPriority[m.PlaceSrc] && m.HasLatLng() {
		if locZone := m.LocationTimeZone(); locZone != "" {
			if zone == tz.UTC {
				local = tz.LocationUTC(utc, tz.Find(locZone))
			}
			zone = locZone
		}
	}

	// Set UTC time and date source.
	m.TakenAt = utc
	m.TakenAtLocal = local
	m.TakenSrc = source
	m.TimeZone = tz.Name(m.TimeZone)

	if zone == tz.UTC && m.TimeZone != tz.Local {
		// Set local time from UTC and keep existing time zone.
		m.TakenAtLocal = m.GetTakenAtLocal()
	} else if zone != tz.Local {
		// Apply new time zone.
		m.TimeZone = zone

		if m.TimeZoneUTC() {
			m.TakenAtLocal = utc
		} else {
			m.TakenAt = m.GetTakenAt()
		}
	} else if m.TimeZoneUTC() {
		m.TakenAtLocal = utc
	} else if !m.TimeZoneLocal() {
		// Keep existing time zone.
		m.TakenAt = m.GetTakenAt()
	}

	m.UpdateDateFields()
}

// TimeZoneUTC tests if the current time zone is UTC.
func (m *Photo) TimeZoneUTC() bool {
	return tz.IsUTC(m.TimeZone)
}

// TimeZoneLocal tests if the current time zone is Local.
func (m *Photo) TimeZoneLocal() bool {
	return tz.IsLocal(m.TimeZone)
}

// UpdateTimeZone applies a new time zone when the source priority allows it and recalculates derived times.
func (m *Photo) UpdateTimeZone(zone string) {
	if zone == "" {
		return
	} else if zone = tz.Name(zone); zone == tz.Local || zone == tz.UTC || zone == tz.Name(m.TimeZone) {
		return
	}

	if SrcPriority[m.TakenSrc] >= SrcPriority[SrcManual] && !tz.IsLocal(m.TimeZone) {
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
	} else if m.TakenSrc != SrcManual && m.TakenSrc != SrcBatch {
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
