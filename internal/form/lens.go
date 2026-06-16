package form

import (
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Lens represents a lens edit form.
type Lens struct {
	LensMake  string `json:"Make"`
	LensModel string `json:"Model"`
}

// NewLens creates a new form struct based on the interface values.
func NewLens(m any) (*Lens, error) {
	frm := &Lens{}
	err := deepcopier.Copy(m).To(frm)
	return frm, err
}

// Validate returns an error if any form values are invalid.
func (frm *Lens) Validate() error {
	lensMake := txt.Clip(frm.LensMake, txt.ClipName)
	lensModel := txt.Clip(frm.LensModel, txt.ClipName)

	if lensMake == "" || lensModel == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	lensSlug := txt.Slug(lensMake + " " + lensModel)

	if lensSlug == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	return nil
}
