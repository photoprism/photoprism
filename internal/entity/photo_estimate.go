package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() || m.HasCountry() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
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
func (m *Photo) EstimatePlace() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
		// Do nothing.
		return
	}

	var recentPhoto Photo
	var dateExpr string

	switch DbDialect() {
	case MySQL:
		dateExpr = "ABS(DATEDIFF(taken_at, ?)) ASC"
	case SQLite:
		dateExpr = "ABS(JulianDay(taken_at) - JulianDay(?)) ASC"
	default:
		log.Errorf("photo: unknown sql dialect %s", DbDialect())
		return
	}

	if err := UnscopedDb().
		Where("place_id <> '' AND place_id <> 'zz' AND place_src <> '' AND place_src <> ?", SrcEstimate).
		Order(gorm.Expr(dateExpr, m.TakenAt)).
		Preload("Place").First(&recentPhoto).Error; err != nil {
		log.Debugf("photo: can't estimate place at %s", m.TakenAt)
		m.EstimateCountry()
	} else {
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
}
