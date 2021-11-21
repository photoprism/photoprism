package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if m.HasLocation() || m.HasPlace() || m.HasCountry() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
		// Do nothing.
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
		log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
	}
}

// EstimatePlace updates the photo with an estimated place and country if possible.
func (m *Photo) EstimatePlace(force bool) {
	if m.HasLocation() || m.HasPlace() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
		// Don't estimate if location is known.
		return
	} else if force || m.EstimatedAt == nil {
		// Proceed.
	} else if hours := TimeStamp().Sub(*m.EstimatedAt) / time.Hour; hours < MetadataEstimateInterval {
		// Ignore if estimated less than 7 days ago.
		return
	}

	// Only estimate country if date isn't known with certainty.
	if m.TakenSrc == SrcAuto {
		m.PlaceID = UnknownPlace.ID
		m.PlaceSrc = SrcEstimate
		m.EstimateCountry()
		m.EstimatedAt = TimePointer()
		return
	}

	var err error

	rangeMin := m.TakenAt.Add(-1 * time.Hour * 72)
	rangeMax := m.TakenAt.Add(time.Hour * 72)

	// Find photo with location info taken at a similar time...
	var recentPhoto Photo

	switch DbDialect() {
	case MySQL:
		err = UnscopedDb().
			Where("place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz' AND place_src <> '' AND place_src <> ?", SrcEstimate).
			Where("taken_at BETWEEN CAST(? AS DATETIME) AND CAST(? AS DATETIME)", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", m.TakenAt)).
			Preload("Place").First(&recentPhoto).Error
	case SQLite:
		err = UnscopedDb().
			Where("place_id IS NOT NULL AND place_id <> '' AND place_id <> 'zz' AND place_src <> '' AND place_src <> ?", SrcEstimate).
			Where("taken_at BETWEEN ? AND ?", rangeMin, rangeMax).
			Order(gorm.Expr("ABS(JulianDay(taken_at) - JulianDay(?)) ASC", m.TakenAt)).
			Preload("Place").First(&recentPhoto).Error
	default:
		log.Warnf("photo: unsupported sql dialect %s", txt.Quote(DbDialect()))
		return
	}

	// Found?
	if err != nil {
		log.Debugf("photo: can't estimate place at %s", m.TakenAt)
		m.EstimateCountry()
	} else {
		// Too much time difference?
		if hours := recentPhoto.TakenAt.Sub(m.TakenAt) / time.Hour; hours < -36 || hours > 36 {
			log.Debugf("photo: can't estimate position of %s, %d hours time difference", m, hours)
		} else if recentPhoto.HasPlace() {
			m.Place = recentPhoto.Place
			m.PlaceID = recentPhoto.PlaceID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			m.UpdateTimeZone(recentPhoto.TimeZone)

			log.Debugf("photo: approximate position of %s is %s (id %s)", m, txt.Quote(m.CountryName()), recentPhoto.PlaceID)
		} else if recentPhoto.HasCountry() {
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			m.UpdateTimeZone(recentPhoto.TimeZone)

			log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
		} else {
			m.EstimateCountry()
		}
	}

	m.EstimatedAt = TimePointer()
}
