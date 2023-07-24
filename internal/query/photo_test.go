package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
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
		result, err := PhotoPreloadByUID("pt9jtdre2lvl0y12")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PhotoPreloadByUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestMissingPhotos(t *testing.T) {
	result, err := MissingPhotos(15, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.LessOrEqual(t, 1, len(result))
}

func TestArchivedPhotos(t *testing.T) {
	results, err := ArchivedPhotos(15, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(results))

	if len(results) > 1 {
		result := results[0]
		assert.Equal(t, "image", result.PhotoType)
		assert.Equal(t, "pt9jtdre2lvl0y25", result.PhotoUID)
	}
}

func TestPhotosMetadataUpdate(t *testing.T) {
	interval := entity.MetadataUpdateInterval
	result, err := PhotosMetadataUpdate(10, 0, time.Second, interval)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Photos{}, result)
}

func TestOrphanPhotos(t *testing.T) {
	result, err := OrphanPhotos()

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Photos{}, result)
}

// TODO How to verify?
func TestFixPrimaries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		err := FixPrimaries()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFlagHiddenPhotos(t *testing.T) {
	// Set photo quality scores to -1 if files are missing.
	if err := FlagHiddenPhotos(); err != nil {
		t.Fatal(err)
	}
}
