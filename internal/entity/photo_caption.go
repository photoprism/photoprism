package entity

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/pkg/txt"
)

// HasCaption checks if the photo has a caption.
func (m *Photo) HasCaption() bool {
	return !m.NoCaption()
}

// NoCaption returns true if the photo has no caption.
func (m *Photo) NoCaption() bool {
	return strings.TrimSpace(m.GetCaption()) == ""
}

// ShouldGenerateCaption checks if a caption should be generated for this model.
func (m *Photo) ShouldGenerateCaption(src Src, force bool) bool {
	return SrcPriority[src] >= SrcPriority[m.CaptionSrc] && (m.NoCaption() || force)
}

// GetCaption returns the photo caption, if any.
func (m *Photo) GetCaption() string {
	return m.PhotoCaption
}

// GetCaptionSrc returns the caption source, if any.
func (m *Photo) GetCaptionSrc() string {
	return m.CaptionSrc
}

// SetCaption stores the supplied caption when it is non-empty and the source priority is sufficient.
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

// GenerateCaption builds an automatic caption from the supplied names when CaptionSrc is auto and enough names are known.
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
	}

	captionSrcPriority := SrcPriority[m.GetCaptionSrc()]

	if captionSrcPriority < SrcPriority[SrcImage] {
		return nil
	}

	start := time.Now()

	var uncertainty int

	if captionSrcPriority < SrcPriority[SrcMeta] {
		uncertainty = 20
	} else {
		uncertainty = 15
	}

	keywords := txt.UniqueKeywords(m.GetCaption())

	var labelIds []uint

	for _, w := range keywords {
		if label, err := FindLabel(w, true); err == nil {
			if label.Skip() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, uncertainty, SrcCaption))
		}
	}

	if err := Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", SrcCaption, m.ID, labelIds).Delete(&PhotoLabel{}).Error; err != nil {
		return err
	}

	log.Debugf("index: updated %s [%s]", english.Plural(len(labelIds), "caption label", "caption labels"), time.Since(start))

	return nil
}
