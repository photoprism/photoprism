package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
)

const Accuracy1Km = 1000

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate] || m.HasLocation() || m.HasPlace() {
		// Keep existing location data.
		return
	} else if m.IsNonPhotographic() {
		// Don't estimate the country if the picture doesn't appear to be photographic.
		return
	}

	// Reset country.
	unknown := UnknownCountry.ID
	countryCode := unknown

	// Try to guess country from photo title.
	if code := txt.CountryCode(m.PhotoTitle); code != unknown {
		countryCode = code
	} else if code = txt.CountryCode(m.PhotoDescription); code != unknown && m.DescriptionSrc != SrcAuto {
		countryCode = code
	}

	// Try to guess country from filename and path.
	if countryCode == unknown {
		if code := txt.CountryCode(m.PhotoName); code != unknown && !fs.IsGenerated(m.PhotoName) {
			countryCode = code
		} else if code = txt.CountryCode(m.PhotoPath); code != unknown {
			countryCode = code
		}
	}

	// Try to guess country from original filename.
	if countryCode == unknown && m.OriginalName != "" && !fs.IsGenerated(m.OriginalName) {
		if code := txt.CountryCode(m.OriginalName); code != UnknownCountry.ID {
			countryCode = code
		}
	}

	// Set new country?
	if countryCode != unknown {
		m.PhotoCountry = countryCode
		m.PlaceSrc = SrcEstimate
		m.EstimatedAt = TimeStamp()
		log.Debugf("photo: estimated country for %s is %s", m, clean.Log(m.CountryName()))
	}
}

// EstimateLocation updates the photo with an estimated place and country if possible.
func (m *Photo) EstimateLocation(force bool) {
	if SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate] || m.HasLocation() && m.PlaceSrc == SrcAuto {
		// Keep existing data.
		return
	} else if force || m.EstimatedAt == nil {
		// Proceed.
	} else if hours := Now().Sub(*m.EstimatedAt); hours < MetadataEstimateInterval {
		// Ignore if location has been estimated recently (in the last 7 days by default).
		return
	}

	m.EstimatedAt = TimeStamp()

	// Remove the location estimate if the picture doesn't appear to be photographic.
	if m.IsNonPhotographic() {
		m.RemoveLocation(SrcEstimate, false)
		m.RemoveLocationLabels()
		return
	}

	// Only estimate the country if TakenAt is unreliable or the picture has no camera metadata.
	if SrcPriority[m.TakenSrc] <= SrcPriority[SrcName] || m.UnknownCamera() {
		m.RemoveLocation(SrcEstimate, false)
		m.RemoveLocationLabels()
		m.EstimateCountry()
		return
	}

	// Estimate the location based on nearby pictures if the date and camera model are known.
	var err error

	rangeMin := m.TakenAt.Add(-1 * time.Hour * 37)
	rangeMax := m.TakenAt.Add(time.Hour * 37)

	var mostRecent Photos

	switch DbDialect() {
	case MySQL:
		err = UnscopedDb().
			Where("photo_lat <> 0 AND photo_lng <> 0").
			Where("place_src <> '' AND place_src <> ? AND place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz'", SrcEstimate).
			Where("taken_src <> '' AND taken_at BETWEEN CAST(? AS DATETIME) AND CAST(? AS DATETIME)", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(TIMESTAMPDIFF(SECOND, taken_at, ?))", m.TakenAt)).Limit(2).
			Preload("Place").Find(&mostRecent).Error
	case SQLite3:
		err = UnscopedDb().
			Where("photo_lat <> 0 AND photo_lng <> 0").
			Where("place_src <> '' AND place_src <> ? AND place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz'", SrcEstimate).
			Where("taken_src <> '' AND taken_at BETWEEN ? AND ?", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(JulianDay(taken_at) - JulianDay(?))", m.TakenAt)).Limit(2).
			Preload("Place").Find(&mostRecent).Error
	default:
		log.Warnf("photo: unsupported sql dialect %s", clean.Log(DbDialect()))
		return
	}

	if err != nil {
		log.Warnf("photo: %s while estimating position", err)
	}

	// Found?
	if len(mostRecent) == 0 {
		log.Debugf("photo: unknown position at %s", m.TakenAt)
		m.RemoveLocation(SrcEstimate, false)
		m.RemoveLocationLabels()
		m.EstimateCountry()
	} else if recentPhoto := mostRecent[0]; recentPhoto.HasLocation() && recentPhoto.HasPlace() {
		// Too much time difference?
		if hours := recentPhoto.TakenAt.Sub(m.TakenAt) / time.Hour; hours < -36 || hours > 36 {
			log.Debugf("photo: skipping %s, %d hours time difference to recent position", m, hours)
			m.RemoveLocation(SrcEstimate, false)
			m.RemoveLocationLabels()
			m.EstimateCountry()
		} else if len(mostRecent) == 1 || m.UnknownCamera() {
			m.AdoptPlace(&recentPhoto, SrcEstimate, false)
		} else {
			p1 := mostRecent[0]
			p2 := mostRecent[1]

			movement := geo.NewMovement(p1.Position(), p2.Position())

			// Ignore inaccurate coordinate estimates.
			if estimate := movement.EstimatePosition(m.TakenAt); movement.Km() < 100 && estimate.Accuracy < Accuracy1Km {
				m.SetPosition(estimate, SrcEstimate, false)
			} else {
				m.AdoptPlace(&recentPhoto, SrcEstimate, false)
			}
		}
	} else if recentPhoto.HasCountry() {
		log.Debugf("photo: estimated country for %s is %s", m, clean.Log(m.CountryName()))
		m.RemoveLocation(SrcEstimate, false)
		m.RemoveLocationLabels()
		m.PhotoCountry = recentPhoto.PhotoCountry
		m.PlaceSrc = SrcEstimate
		m.UpdateTimeZone(recentPhoto.TimeZone)
	} else {
		log.Warnf("photo: %s has no location, uid %s", recentPhoto.PhotoName, recentPhoto.PhotoUID)
		m.RemoveLocation(SrcEstimate, false)
		m.RemoveLocationLabels()
		m.EstimateCountry()
	}
}
