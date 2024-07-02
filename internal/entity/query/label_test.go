package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelBySlug(t *testing.T) {
	t.Run("file found", func(t *testing.T) {
		label, err := LabelBySlug("flower")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Flower", label.LabelName)
	})

	t.Run("no file found", func(t *testing.T) {
		label, err := LabelBySlug("111")

		assert.Error(t, err, "record not found")
		assert.Empty(t, label.ID)
	})
}

func TestLabelByUID(t *testing.T) {
	t.Run("file found", func(t *testing.T) {
		label, err := LabelByUID("ls6sg6b1wowuy3c5")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "COW", label.LabelName)
	})

	t.Run("no file found", func(t *testing.T) {
		label, err := LabelByUID("111")

		assert.Error(t, err, "record not found")
		assert.Empty(t, label.ID)
	})
}

func TestLabelThumbBySlug(t *testing.T) {
	t.Run("file found", func(t *testing.T) {
		file, err := LabelThumbBySlug("cow")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04/bridge2.jpg", file.FileName)
	})

	t.Run("no file found", func(t *testing.T) {
		file, err := LabelThumbBySlug("no-jpeg")

		if err == nil {
			t.Fatalf("did not expect to find file: %+v", file)
		}
	})
}

func TestLabelThumbByUID(t *testing.T) {
	t.Run("file found", func(t *testing.T) {
		file, err := LabelThumbByUID("ls6sg6b1wowuy3c5")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04/bridge2.jpg", file.FileName)
	})

	t.Run("no file found", func(t *testing.T) {
		file, err := LabelThumbByUID("14")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestPhotoLabel(t *testing.T) {
	t.Run("photo label found", func(t *testing.T) {
		r, err := PhotoLabel(uint(1000000), uint(1000001))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 38, r.Uncertainty)
	})
	t.Run("no photo label found", func(t *testing.T) {
		r, err := PhotoLabel(uint(1000000), uint(1000003))
		assert.Equal(t, "record not found", err.Error())
		t.Log(r)
	})
}
