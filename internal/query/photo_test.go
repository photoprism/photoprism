package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoByID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PhotoByID(1000000)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2790, result.PhotoYear)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PhotoByID(99999)
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestPhotoByUID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PhotoByUID("pt9jtdre2lvl0y12")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PhotoByUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestPreloadPhotoByUID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PreloadPhotoByUID("pt9jtdre2lvl0y12")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PreloadPhotoByUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestMissingPhotos(t *testing.T) {
	r, err := MissingPhotos(15, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, 1, len(r))
}

func TestResetPhotosQuality(t *testing.T) {
	err := ResetPhotosQuality()
	if err != nil {
		t.Fatal(err)
	}
}
