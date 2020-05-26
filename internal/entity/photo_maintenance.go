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
		Where("place_uid <> '' && place_uid <> 'zz'").
		Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", m.TakenAt)).
		Preload("Place").First(&recentPhoto); result.Error == nil {
		if recentPhoto.HasPlace() {
			m.Place = recentPhoto.Place
			m.PlaceUID = recentPhoto.PlaceUID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.LocSrc = SrcEstimate
			log.Debugf("prism: approximate position of %s is %s", m.PhotoUID, recentPhoto.PlaceUID)
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

	if m.NoLocation() && (m.LocSrc == SrcAuto || m.LocSrc == SrcEstimate) {
		m.EstimatePosition()
	}

	labels := m.ClassifyLabels()

	m.UpdateYearMonth()

	if err := m.UpdateTitle(labels); err != nil {
		log.Warn(err)
	}

	if m.DetailsLoaded() {
		w := txt.UniqueKeywords(m.Details.Keywords)
		w = append(w, labels.Keywords()...)
		m.Details.Keywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := m.IndexKeywords(); err != nil {
		log.Error(err)
	}

	m.PhotoQuality = m.QualityScore()

	return UnscopedDb().Save(m).Error
}
