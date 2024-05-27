package entity

import (
	"errors"
	"reflect"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Optimize photo data, improve if possible.
func (m *Photo) Optimize(mergeMeta, mergeUuid, estimatePlace, force bool) (updated bool, merged Photos, err error) {
	if !m.HasID() {
		return false, merged, errors.New("photo: cannot maintain, id is empty")
	}

	current := *m

	if m.HasLatLng() && !m.HasLocation() {
		m.UpdateLocation()
	}

	if original, photos, err := m.Merge(mergeMeta, mergeUuid); err != nil {
		return updated, merged, err
	} else if len(photos) > 0 && original.ID == m.ID {
		merged = photos
	} else if len(photos) > 0 && original.ID != m.ID {
		return false, photos, nil
	}

	// Estimate if feature is enabled and place wasn't set otherwise.
	if estimatePlace && SrcPriority[m.PlaceSrc] <= SrcPriority[SrcEstimate] {
		m.EstimateLocation(force)
	}

	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if err := m.UpdateTitle(labels); err != nil {
		log.Info(err)
	}

	details := m.GetDetails()
	w := txt.UniqueWords(txt.Words(details.Keywords))
	w = append(w, labels.Keywords()...)
	details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	m.PhotoQuality = m.QualityScore()

	checked := Now()

	if reflect.DeepEqual(*m, current) {
		return false, merged, m.Update("CheckedAt", &checked)
	}

	m.CheckedAt = &checked

	return true, merged, m.Save()
}
