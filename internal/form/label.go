package form

import (
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Label represents a label edit form.
type Label struct {
	LabelName        string `json:"Name"`
	LabelPriority    int    `json:"Priority"`
	LabelFavorite    bool   `json:"Favorite"`
	LabelDescription string `json:"Description"`
	LabelNotes       string `json:"Notes"`
	Thumb            string `json:"Thumb"`
	ThumbSrc         string `json:"ThumbSrc"`
	Uncertainty      int    `json:"Uncertainty"`
}

// NewLabel creates a new form struct based on the interface values.
func NewLabel(m interface{}) (*Label, error) {
	frm := &Label{}
	err := deepcopier.Copy(m).To(frm)
	return frm, err
}

// Validate returns an error if any form values are invalid.
func (frm *Label) Validate() error {
	labelName := txt.Clip(clean.NameCapitalized(frm.LabelName), txt.ClipName)

	if labelName == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	labelSlug := txt.Slug(labelName)

	if labelSlug == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	return nil
}
