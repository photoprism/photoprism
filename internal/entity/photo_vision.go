package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SaveVision persists derived metadata after a vision pass by regenerating the title and refreshing label counts.
func (m *Photo) SaveVision() (err error) {
	if m == nil {
		return
	} else if !m.HasID() {
		return fmt.Errorf("photo id is missing")
	}

	err = m.GenerateAndSaveTitle()

	if err != nil {
		log.Errorf("vision: failed to update %s (%s)", m.String(), clean.Error(err))
	} else if labelsErr := UpdateLabelCountsIfNeeded(); labelsErr != nil {
		// Update precalculated label photo counts if needed.
		log.Warnf("vision: failed to update label counts (%s)", labelsErr)
	} else {
		log.Infof("vision: updated %s", m.String())
	}

	return err
}
