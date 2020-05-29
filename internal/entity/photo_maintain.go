package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimatePosition updates the photo with an estimated geolocation if possible.
func (m *Photo) EstimatePosition() {
	var recentPhoto Photo

	if result := UnscopedDb().
		Where("place_id <> '' AND place_id <> 'zz' AND loc_src <> '' AND loc_src <> ?", SrcEstimate).
		Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", m.TakenAt)).
		Preload("Place").First(&recentPhoto); result.Error == nil {
		if recentPhoto.HasPlace() {
			m.Place = recentPhoto.Place
			m.PlaceID = recentPhoto.PlaceID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.LocSrc = SrcEstimate
			log.Debugf("prism: approximate position of %s is %s", m.PhotoUID, recentPhoto.PlaceID)
		} else if recentPhoto.HasCountry() {
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.LocSrc = SrcEstimate
			log.Debugf("prism: probable country for %s is %s", m.PhotoUID, recentPhoto.PhotoCountry)
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

	if m.UnknownCountry() && m.LocSrc == SrcAuto || m.UnknownLocation() && m.LocSrc == SrcEstimate {
		m.EstimatePosition()
	}

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
