package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPhotoSelection(t *testing.T) {
	t.Run("no photos selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{},
		}
		r, err := PhotoSelection(f)
		assert.Equal(t, "no photos selected", err.Error())
		assert.Empty(t, r)
	})
	t.Run("photos selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"},
		}
		r, err := PhotoSelection(f)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(r))
		assert.IsType(t, []entity.Photo{}, r)
	})
}
