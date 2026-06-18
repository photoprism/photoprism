/*
Package meta resolver.go normalizes the time-zone, capture-time, and local-time
fields on a Data receiver from any combination of CreatedAt, TakenGps, Lat/Lng,
TimeOffset, MimeType, and TakenNs already populated by an upstream reader.
Shared between the ExifTool JSON path (internal/meta/json_exiftool.go) and the
XMP sidecar path (internal/meta/xmp.go) so both flows produce identical entity
state for the same capture metadata.
*/
package meta

import (
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/time/tz"
)

// ResolveTimeZone normalizes TakenAt, TakenAtLocal, and TimeZone from any
// combination of CreatedAt, TakenGps, Lat/Lng, TimeOffset, MimeType, and
// TakenNs already present on the Data receiver. Safe to call from any
// metadata reader (EXIF, ExifTool JSON, XMP sidecar). Reads no external
// state beyond the receiver; writes only data.TakenAt, data.TakenAtLocal,
// and data.TimeZone. The logName argument is used for diagnostic log
// output only.
func (data *Data) ResolveTimeZone(logName string) {
	hasTimeOffset := false

	// Has Media Create Date?
	if !data.CreatedAt.IsZero() {
		data.TakenAt = data.CreatedAt
	}

	// Fallback to GPS UTC Time?
	if data.TakenAt.IsZero() && data.TakenAtLocal.IsZero() && !data.TakenGps.IsZero() {
		data.TimeZone = tz.UTC
		data.TakenAt = data.TakenGps.UTC()
		data.TakenAtLocal = time.Time{}
	}

	// Check plausibility of the local <> UTC time difference.
	if !data.TakenAt.IsZero() && !data.TakenAtLocal.IsZero() {
		if d := data.TakenAt.Sub(data.TakenAtLocal).Abs(); d > time.Hour*27 {
			log.Infof("metadata: %s has an invalid local time offset (%s)", logName, d.String())
			log.Debugf("metadata: %s was taken at %s, local time %s, create time %s, time zone %s", logName, clean.Log(data.TakenAt.UTC().String()), clean.Log(data.TakenAtLocal.String()), clean.Log(data.CreatedAt.String()), clean.Log(data.TimeZone))
			data.TakenAtLocal = data.TakenAt
			data.TakenAt = data.TakenAt.UTC()
		}
	}

	// Has time zone offset?
	if _, offset := data.TakenAtLocal.Zone(); offset != 0 && !data.TakenAtLocal.IsZero() {
		hasTimeOffset = true
	} else if mt := data.MimeType; mt != "" && data.TakenAtLocal.IsZero() && (mt == MimeVideoMp4 || mt == MimeQuicktime) {
		// Assume default time zone for MP4 & Quicktime videos is UTC.
		// see https://exiftool.org/TagNames/QuickTime.html
		log.Tracef("metadata: default time zone for %s is UTC (%s)", logName, clean.Log(mt))
		data.TimeZone = tz.UTC
		data.TakenAt = data.TakenAt.UTC()
		data.TakenAtLocal = time.Time{}
	}

	// Set time zone and calculate UTC time.
	if data.Lat != 0 && data.Lng != 0 {
		if zone := tz.Position(data.Lat, data.Lng); zone != "" {
			data.TimeZone = zone
		}

		if loc := tz.Find(data.TimeZone); !data.TakenAtLocal.IsZero() {
			if tl, parseErr := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); parseErr == nil {
				data.TakenAtLocal = tz.Strip(data.TakenAtLocal)
				data.TakenAt = tl.Truncate(time.Second).UTC()
			} else {
				log.Errorf("metadata: %s", clean.Error(parseErr)) // this should never happen
			}
		} else if !data.TakenAt.IsZero() {
			if localUtc, parseErr := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAt.In(loc).Format("2006:01:02 15:04:05"), time.UTC); parseErr == nil {
				data.TakenAtLocal = localUtc
				data.TakenAt = data.TakenAt.UTC()
			} else {
				log.Errorf("metadata: %s", clean.Error(parseErr)) // this should never happen
			}
		}
	} else if hasTimeOffset {
		if localUtc, parseErr := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), time.UTC); parseErr == nil {
			data.TakenAtLocal = localUtc.Truncate(time.Second).UTC()
		}

		data.TakenAt = data.TakenAt.Truncate(time.Second).UTC()
	}

	// Set UTC offset as time zone?
	if data.TimeZone != "" && data.TimeZone != tz.Local && data.TimeZone != tz.UTC || data.TakenAt.IsZero() {
		// Don't change existing time zone.
	} else if utcOffset := tz.UtcOffset(data.TakenAt, data.TakenAtLocal, data.TimeOffset); utcOffset != "" {
		data.TimeZone = utcOffset

		if data.TakenAtLocal.IsZero() {
			data.TakenAtLocal = tz.Strip(data.TakenAt)
		}

		data.TakenAt = data.TakenAt.UTC()
		log.Infof("metadata: %s has time offset %s", logName, clean.Log(utcOffset))
	} else if data.TimeOffset != "" {
		log.Infof("metadata: %s has invalid time offset %s", logName, clean.Log(data.TimeOffset))
	}

	// Normalize time zone name.
	data.TimeZone = tz.Name(data.TimeZone)

	// Set local time based on UTC time if empty.
	if data.TakenAtLocal.IsZero() && !data.TakenAt.IsZero() {
		if loc := tz.Find(data.TimeZone); loc.String() == tz.Local {
			data.TakenAtLocal = tz.Strip(data.TakenAt)
			data.TakenAt = data.TakenAt.UTC()
		} else if localUtc, parseErr := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAt.In(loc).Format("2006:01:02 15:04:05"), time.UTC); parseErr == nil {
			data.TakenAtLocal = localUtc
			data.TakenAt = data.TakenAt.UTC()
		} else {
			log.Errorf("metadata: %s", clean.Error(parseErr)) // this should never happen
		}
	}

	// Add nanoseconds to the calculated UTC and local time.
	if data.TakenAt.Nanosecond() == 0 {
		if ns := time.Duration(data.TakenNs); ns > 0 && ns <= time.Second {
			data.TakenAt = data.TakenAt.Truncate(time.Second).UTC().Add(ns)
			data.TakenAtLocal = data.TakenAtLocal.Truncate(time.Second).Add(ns)
		}
	}
}
