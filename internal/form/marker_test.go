package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var m = struct {
			SubjectSrc    string
			MarkerName    string
			Review        bool
			MarkerInvalid bool
		}{
			SubjectSrc:    "manual",
			MarkerName:    "Foo",
			Review:        true,
			MarkerInvalid: true,
		}

		f, err := NewMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "manual", f.SubjectSrc)
		assert.Equal(t, "Foo", f.MarkerName)
		assert.Equal(t, true, f.Review)
		assert.Equal(t, true, f.MarkerInvalid)
	})
}
