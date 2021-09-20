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
