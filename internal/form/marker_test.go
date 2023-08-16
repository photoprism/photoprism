package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var m = struct {
			SubjSrc       string
			MarkerName    string
			MarkerReview  bool
			MarkerInvalid bool
		}{
			SubjSrc:       "manual",
			MarkerName:    "Foo",
			MarkerReview:  true,
			MarkerInvalid: true,
		}

		f, err := NewMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "manual", f.SubjSrc)
		assert.Equal(t, "Foo", f.MarkerName)
		assert.Equal(t, true, f.MarkerReview)
		assert.Equal(t, true, f.MarkerInvalid)
	})
}

func TestMarker_Validate(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		frm := Marker{}
		assert.Error(t, frm.Validate())
	})
	t.Run("False", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Nil(t, frm.Validate())
	})
	t.Run("FileUID", func(t *testing.T) {
		frm := Marker{
			FileUID:       "rygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
	t.Run("Area", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             1.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Jens Mander",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
	t.Run("Name", func(t *testing.T) {
		frm := Marker{
			FileUID:       "frygcme3hc9re8nc",
			MarkerType:    "face",
			X:             0.303519,
			Y:             0.260742,
			W:             0.548387,
			H:             0.365234,
			SubjSrc:       "manual",
			MarkerName:    "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer...",
			MarkerReview:  false,
			MarkerInvalid: false,
		}
		assert.Error(t, frm.Validate())
	})
}
