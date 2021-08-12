package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarker(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var m = struct {
			Ref           string
			RefSrc        string
			MarkerSrc     string
			MarkerType    string
			MarkerScore   int
			MarkerInvalid bool
			MarkerLabel   string
		}{
			Ref:           "3h59wvth837b5vyiub35",
			RefSrc:        "meta",
			MarkerSrc:     "image",
			MarkerType:    "Face",
			MarkerScore:   100,
			MarkerInvalid: true,
			MarkerLabel:   "Foo",
		}

		f, err := NewMarker(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "3h59wvth837b5vyiub35", f.Ref)
		assert.Equal(t, "meta", f.RefSrc)
		assert.Equal(t, "image", f.MarkerSrc)
		assert.Equal(t, "Face", f.MarkerType)
		assert.Equal(t, 100, f.MarkerScore)
		assert.Equal(t, true, f.MarkerInvalid)
		assert.Equal(t, "Foo", f.MarkerLabel)
	})
}
