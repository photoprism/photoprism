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

func TestLabelHasThumb(t *testing.T) {
	// Other tests assign covers through UpdateCovers(), so the fixture value is set explicitly
	// and restored afterwards.
	setLabelThumb := func(t *testing.T, uid, hash string) {
		var current []string

		if err := Db().Model(&entity.Label{}).Where("label_uid = ?", uid).Limit(1).Pluck("thumb", &current).Error; err != nil {
			t.Fatal(err)
		} else if err = Db().Model(&entity.Label{}).Where("label_uid = ?", uid).Update("thumb", hash).Error; err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			restore := ""

			if len(current) > 0 {
				restore = current[0]
			}

			_ = Db().Model(&entity.Label{}).Where("label_uid = ?", uid).Update("thumb", restore).Error
		})
	}

	t.Run("NoThumb", func(t *testing.T) {
		setLabelThumb(t, "ls6sg6b1wowuy3c5", "")
		assert.False(t, LabelHasThumb("ls6sg6b1wowuy3c5"))
	})
	t.Run("HasThumb", func(t *testing.T) {
		setLabelThumb(t, "ls6sg6b1wowuy3c5", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.True(t, LabelHasThumb("ls6sg6b1wowuy3c5"))
	})
	t.Run("StaleThumb", func(t *testing.T) {
		// A hash no client can resolve must not gate the cover query.
		setLabelThumb(t, "ls6sg6b1wowuy3c5", "0000000000000000000000000000000000000000")
		assert.False(t, LabelHasThumb("ls6sg6b1wowuy3c5"))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.False(t, LabelHasThumb("ls6sg6b1wow00000"))
	})
	t.Run("InvalidUID", func(t *testing.T) {
		assert.False(t, LabelHasThumb("3765"))
		assert.False(t, LabelHasThumb(""))
	})
}

func TestLabelThumbByUID(t *testing.T) {
	t.Run("InvalidUID", func(t *testing.T) {
		for _, uid := range []string{"", "xxx", "as6sg6bxpogaaba8"} {
			_, err := LabelThumbByUID(uid)
			assert.EqualError(t, err, "invalid label uid", "%s", uid)
		}
	})
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
