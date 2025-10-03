package entity

import (
	"errors"
	"reflect"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Optimize updates picture metadata, enriching titles, keywords, and locations according to the supplied flags.
func (m *Photo) Optimize(mergeMeta, mergeUuid, estimateLocation, force bool) (updated bool, merged Photos, err error) {
	if !m.HasID() {
		return false, merged, errors.New("photo: cannot maintain, id is empty")
	}

	// Keep a snapshot so we can detect whether anything changed.
	current := *m

	if m.HasLatLng() && !m.HasLocation() {
		m.UpdateLocation()
	}

	if original, photos, mergeErr := m.Merge(mergeMeta, mergeUuid); mergeErr != nil {
		return updated, merged, mergeErr
	} else if len(photos) > 0 && original.ID == m.ID {
		merged = photos
	} else if len(photos) > 0 && original.ID != m.ID {
		return false, photos, nil
	}

	// Estimate the location if it is unknown and this feature is enabled.
	if estimateLocation && SrcPriority[m.PlaceSrc] <= SrcPriority[SrcEstimate] {
		m.EstimateLocation(force)
	}

	// Get image classification labels.
	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if updateErr := m.GenerateTitle(labels); updateErr != nil {
		log.Info(updateErr)
	}

	details := m.GetDetails()
	w := txt.UniqueWords(txt.Words(details.Keywords))
	w = append(w, labels.Keywords()...)
	details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if indexErr := m.IndexKeywords(); indexErr != nil {
		log.Errorf("photo: %s", indexErr.Error())
	}

	m.PhotoQuality = m.QualityScore()

	checked := Now()

	// Skip persistence when nothing changed besides the CheckedAt timestamp.
	if reflect.DeepEqual(*m, current) {
		return false, merged, m.Update("CheckedAt", &checked)
	}

	m.CheckedAt = &checked

	// Persist the updated metadata to the database.
	return true, merged, m.Save()
}
