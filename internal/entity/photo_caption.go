package entity

import (
	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/txt"
)

// HasCaption checks if the photo has a caption.
func (m *Photo) HasCaption() bool {
	return !m.NoCaption()
}

// NoCaption returns true if the photo has no caption.
func (m *Photo) NoCaption() bool {
	return m.GetCaption() == ""
}

// GetCaption returns the photo caption, if any.
func (m *Photo) GetCaption() string {
	return m.PhotoCaption
}

// GetCaptionSrc returns the caption source, if any.
func (m *Photo) GetCaptionSrc() string {
	return m.CaptionSrc
}

// SetCaption sets the specified caption if is not empty and from the same source.
func (m *Photo) SetCaption(caption, source string) {
	newCaption := txt.Clip(caption, txt.ClipLongText)

	if newCaption == "" {
		return
	}

	if (SrcPriority[source] < SrcPriority[m.CaptionSrc]) && m.HasCaption() {
		return
	}

	m.PhotoCaption = newCaption
	m.CaptionSrc = source
}

// GenerateCaption generates the caption from the specified list of at least 3 names if CaptionSrc is auto.
func (m *Photo) GenerateCaption(names []string) {
	if m.CaptionSrc != SrcAuto {
		return
	}

	// Generate caption from the specified list of names.
	if len(names) > 3 {
		m.PhotoCaption = txt.JoinNames(names, false)
	} else {
		m.PhotoCaption = ""
	}
}

// UpdateCaptionLabels updates the labels assigned based on the photo caption.
func (m *Photo) UpdateCaptionLabels() error {
	if m == nil {
		return nil
	} else if !m.HasCaption() {
		return nil
	} else if SrcPriority[m.GetCaptionSrc()] < SrcPriority[SrcMeta] {
		return nil
	}

	keywords := txt.UniqueKeywords(m.GetCaption())

	var labelIds []uint

	for _, w := range keywords {
		if label, err := FindLabel(w, true); err == nil {
			if label.Skip() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, 15, classify.SrcCaption))
		}
	}

	return Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", classify.SrcCaption, m.ID, labelIds).Delete(&PhotoLabel{}).Error
}
