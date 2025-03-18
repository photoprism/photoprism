package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var album = struct {
			LabelName     string
			Uncertainty   int
			LabelPriority int
			LabelFavorite bool
		}{
			LabelName:     "New Label",
			Uncertainty:   50,
			LabelPriority: -5,
			LabelFavorite: false,
		}

		result, err := NewLabel(album)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &Label{}, result)
		assert.Equal(t, "New Label", result.LabelName)
		assert.Equal(t, 50, result.Uncertainty)
		assert.Equal(t, -5, result.LabelPriority)
		assert.Equal(t, false, result.LabelFavorite)
	})
	t.Run("Favorite", func(t *testing.T) {
		var album = struct {
			LabelName     string
			Uncertainty   int
			LabelPriority int
			LabelFavorite bool
		}{
			LabelName:     "Foo",
			Uncertainty:   10,
			LabelPriority: 5,
			LabelFavorite: true,
		}

		result, err := NewLabel(album)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &Label{}, result)
		assert.Equal(t, "Foo", result.LabelName)
		assert.Equal(t, 10, result.Uncertainty)
		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, true, result.LabelFavorite)
	})
}
