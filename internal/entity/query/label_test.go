package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestLabelBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := LabelBySlug("flower")

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &entity.Label{}, result)
		assert.Equal(t, "Flower", result.LabelName)
	})

	t.Run("NotFound", func(t *testing.T) {
		label, err := LabelBySlug("111")

		assert.IsType(t, &entity.Label{}, label)
		assert.Error(t, err, "record not found")
		assert.Empty(t, label.ID)
	})
}

func TestLabelByUID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := LabelByUID("ls6sg6b1wowuy3c5")

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &entity.Label{}, result)
		assert.Equal(t, "COW", result.LabelName)
	})

	t.Run("NotFound", func(t *testing.T) {
		result, err := LabelByUID("111")

		assert.IsType(t, &entity.Label{}, result)
		assert.Error(t, err, "record not found")
		assert.Empty(t, result.ID)
	})
}

func TestLabelThumbBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := LabelThumbBySlug("cow")

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &entity.File{}, result)
		assert.Equal(t, "1990/04/bridge2.jpg", result.FileName)
	})
	t.Run("NotFound", func(t *testing.T) {
		result, err := LabelThumbBySlug("no-jpeg")

		if err == nil {
			t.Fatalf("did not expect to find file: %+v", result)
		}

		assert.IsType(t, &entity.File{}, result)
	})
}

func TestLabelThumbByUID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := LabelThumbByUID("ls6sg6b1wowuy3c5")

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &entity.File{}, result)
		assert.Equal(t, "1990/04/bridge2.jpg", result.FileName)
	})
	t.Run("NotFound", func(t *testing.T) {
		result, err := LabelThumbByUID("14")

		assert.IsType(t, &entity.File{}, result)
		assert.Error(t, err, "record not found")
	})
}

func TestPhotoLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := PhotoLabel(uint(1000000), uint(1000001))

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, &entity.PhotoLabel{}, result)
		assert.Equal(t, 38, result.Uncertainty)
	})
	t.Run("NotFound", func(t *testing.T) {
		result, err := PhotoLabel(uint(1000000), uint(1000003))

		assert.IsType(t, &entity.PhotoLabel{}, result)
		assert.Equal(t, "record not found", err.Error())
	})
}
