package form

import (
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Camera represents a camera edit form.
type Camera struct {
	CameraMake  string `json:"Make"`
	CameraModel string `json:"Model"`
}

// NewCamera creates a new form struct based on the interface values.
func NewCamera(m any) (*Camera, error) {
	frm := &Camera{}
	err := deepcopier.Copy(m).To(frm)
	return frm, err
}

// Validate returns an error if any form values are invalid.
func (frm *Camera) Validate() error {
	cameraMake := txt.Clip(frm.CameraMake, txt.ClipName)
	cameraModel := txt.Clip(frm.CameraModel, txt.ClipName)

	if cameraMake == "" || cameraModel == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	cameraSlug := txt.Slug(cameraMake + " " + cameraModel)

	if cameraSlug == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	return nil
}
