package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate] || m.HasLocation() || m.HasPlace() {
		// Ignore.
		return
	}

	unknown := UnknownCountry.ID
	countryCode := unknown

	if code := txt.CountryCode(m.PhotoTitle); code != unknown {
		countryCode = code
	}

	if countryCode == unknown {
		if code := txt.CountryCode(m.PhotoName); code != unknown && !fs.IsGenerated(m.PhotoName) {
			countryCode = code
		} else if code := txt.CountryCode(m.PhotoPath); code != unknown {
			countryCode = code
		}
	}

	if countryCode == unknown && m.OriginalName != "" && !fs.IsGenerated(m.OriginalName) {
		if code := txt.CountryCode(m.OriginalName); code != UnknownCountry.ID {
			countryCode = code
		}
	}

	if countryCode != unknown {
		m.PhotoCountry = countryCode
		m.PlaceSrc = SrcEstimate
		log.Debugf("estimate: probable country for %s is %s", m, txt.Quote(m.CountryName()))
	}
}

// EstimateLocation updates the photo with an estimated place and country if possible.
func (m *Photo) EstimateLocation(force bool) {
	if SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate] {
		// Ignore if location was set otherwise.
		return
	} else if force || m.EstimatedAt == nil {
		// Proceed.
	} else if hours := TimeStamp().Sub(*m.EstimatedAt) / time.Hour; hours < MetadataEstimateInterval {
		// Ignore if estimated less than 7 days ago.
		return
	}

	// Only estimate country if date isn't known with certainty.
	if m.TakenSrc == SrcAuto {
		m.RemoveLocation()
		m.PlaceID = UnknownPlace.ID
		m.PlaceSrc = SrcEstimate
		m.EstimateCountry()
		m.EstimatedAt = TimePointer()
		return
	}

	var err error

	rangeMin := m.TakenAt.Add(-1 * time.Hour * 48)
	rangeMax := m.TakenAt.Add(time.Hour * 48)

	// Find photo with location info taken at a similar time...
	var mostRecent Photos

	switch DbDialect() {
	case MySQL:
		err = UnscopedDb().
			Where("photo_lat <> 0 AND photo_lng <> 0").
			Where("place_src <> '' AND place_src <> ? AND place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz'", SrcEstimate).
			Where("taken_src <> '' AND taken_at BETWEEN CAST(? AS DATETIME) AND CAST(? AS DATETIME)", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", m.TakenAt)).Limit(2).
			Preload("Place").Find(&mostRecent).Error
	case SQLite:
		err = UnscopedDb().
			Where("photo_lat <> 0 AND photo_lng <> 0").
			Where("place_src <> '' AND place_src <> ? AND place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz'", SrcEstimate).
			Where("taken_src <> '' AND taken_at BETWEEN ? AND ?", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(JulianDay(taken_at) - JulianDay(?)) ASC", m.TakenAt)).Limit(2).
			Preload("Place").Find(&mostRecent).Error
	default:
		log.Warnf("estimate: unsupported sql dialect %s", txt.Quote(DbDialect()))
		return
	}

	// Found?
	if err != nil || len(mostRecent) == 0 {
		log.Debugf("estimate: unknown position at %s", m.TakenAt)
		m.RemoveLocation()
		m.EstimateCountry()
	} else if recentPhoto := mostRecent[0]; recentPhoto.HasLocation() && recentPhoto.HasPlace() {
		// Too much time difference?
		if hours := recentPhoto.TakenAt.Sub(m.TakenAt) / time.Hour; hours < -36 || hours > 36 {
			log.Debugf("estimate: skipping %s, %d hours time difference to recent position", m, hours)
		} else if len(mostRecent) == 1 {
			m.RemoveLocation()
			m.Place = recentPhoto.Place
			m.PlaceID = recentPhoto.PlaceID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			m.UpdateTimeZone(recentPhoto.TimeZone)

			log.Debugf("estimate: approximate place of %s is %s (id %s)", m, txt.Quote(m.Place.Label()), recentPhoto.PlaceID)
		} else if recentPhoto.HasPlace() {
			p1 := mostRecent[0]
			p2 := mostRecent[1]

			m.PlaceSrc = SrcEstimate

			movement := geo.NewMovement(p1.Position(), p2.Position(), p1.TakenAt, p2.TakenAt)

			if movement.DistKm < 100 {
				estimate := movement.Position(m.TakenAt)

				m.PhotoLat = float32(estimate.Lat)
				m.PhotoLng = float32(estimate.Lng)

				log.Debugf("estimate: positioned %s at lat %f, lng %f", m, m.PhotoLat, m.PhotoLng)

				m.UpdateLocation()

				if m.Place == nil {
					log.Warnf("estimate: failed updating position of %s", m)
				} else {
					log.Debugf("estimate: approximate place of %s is %s (id %s)", m, txt.Quote(m.Place.Label()), m.PlaceID)
				}
			} else {
				m.RemoveLocation()
				m.Place = recentPhoto.Place
				m.PlaceID = recentPhoto.PlaceID
				m.PhotoCountry = recentPhoto.PhotoCountry
				m.UpdateTimeZone(recentPhoto.TimeZone)
			}
		} else if recentPhoto.HasCountry() {
			m.RemoveLocation()
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			m.UpdateTimeZone(recentPhoto.TimeZone)

			log.Debugf("estimate: probable country for %s is %s", m, txt.Quote(m.CountryName()))
		} else {
			m.RemoveLocation()
			m.EstimateCountry()
		}
	} else {
		log.Warnf("estimate: %s has no location, uid %s", recentPhoto.PhotoName, recentPhoto.PhotoUID)
		m.RemoveLocation()
		m.EstimateCountry()
	}

	m.EstimatedAt = TimePointer()
}
