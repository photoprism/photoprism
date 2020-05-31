package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() || m.HasCountry() && m.LocSrc != SrcAuto && m.LocSrc != SrcEstimate {
		// Do nothing.
		return
	}

	unknown := UnknownCountry.ID
	countryCode := unknown

	if code := txt.CountryCode(m.PhotoTitle); code != unknown {
		countryCode = code
	}

	if countryCode == unknown && fs.NonCanonical(m.PhotoName) {
		if code := txt.CountryCode(m.PhotoName); code != unknown {
			countryCode = code
		} else if code := txt.CountryCode(m.PhotoPath); code != unknown {
			countryCode = code
		}
	}

	if countryCode == unknown && m.OriginalName != "" && fs.NonCanonical(m.OriginalName) {
		if code := txt.CountryCode(m.OriginalName); code != UnknownCountry.ID {
			countryCode = code
		}
	}

	if countryCode != unknown {
		m.PhotoCountry = countryCode
		m.LocSrc = SrcEstimate
		log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
	}
}

// EstimatePlace updates the photo with an estimated place and country if possible.
func (m *Photo) EstimatePlace() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() && m.LocSrc != SrcAuto && m.LocSrc != SrcEstimate {
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
		Where("place_id <> '' AND place_id <> 'zz' AND loc_src <> '' AND loc_src <> ?", SrcEstimate).
		Order(gorm.Expr(dateExpr, m.TakenAt)).
		Preload("Place").First(&recentPhoto).Error; err != nil {
		log.Errorf("photo: %s (estimate place)", err.Error())
		m.EstimateCountry()
	} else {
		if days := recentPhoto.TakenAt.Sub(m.TakenAt) / (time.Hour * 24); days < -7 {
			log.Debugf("photo: can't estimate position of %s, %d days time difference", m, -1*days)
		} else if days > -7 {
			log.Debugf("photo: can't estimate position of %s, %d days time difference", m, days)
		} else if recentPhoto.HasPlace() {
			m.Place = recentPhoto.Place
			m.PlaceID = recentPhoto.PlaceID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.LocSrc = SrcEstimate
			log.Debugf("photo: approximate position of %s is %s (id %s)", m, txt.Quote(m.CountryName()), recentPhoto.PlaceID)
		} else if recentPhoto.HasCountry() {
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.LocSrc = SrcEstimate
			log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
		} else {
			m.EstimateCountry()
		}
	}
}

// Maintain photo data, improve if possible.
func (m *Photo) Maintain() error {
	if !m.HasID() {
		return errors.New("photo: can't maintain, id is empty")
	}

	maintained := time.Now()
	m.MaintainedAt = &maintained

	m.EstimatePlace()

	labels := m.ClassifyLabels()

	m.UpdateYearMonth()

	if err := m.UpdateTitle(labels); err != nil {
		log.Warn(err)
	}

	if m.DetailsLoaded() {
		w := txt.UniqueWords(txt.Words(m.Details.Keywords))
		w = append(w, labels.Keywords()...)
		m.Details.Keywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	m.PhotoQuality = m.QualityScore()

	return UnscopedDb().Save(m).Error
}
