package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*func TestQuery_Photos(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("search with query", func(t *testing.T) {
		query := form.NewPhotoSearch("Title:Reunion")
		result, err := search.Photos(query)

		t.Log(result)
		t.Log(err)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Cake", result[1].PhotoName)
		assert.Equal(t, "COW", result[0].PhotoName)
	})
}*/

func TestPhotoByID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PhotoByID(1000000)
		assert.Nil(t, err)
		assert.Equal(t, 2790, result.PhotoYear)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PhotoByID(99999)
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestPhotoByUUID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PhotoByUUID("pt9jtdre2lvl0y12")
		assert.Nil(t, err)
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PhotoByUUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

func TestPreloadPhotoByUUID(t *testing.T) {
	t.Run("photo found", func(t *testing.T) {
		result, err := PreloadPhotoByUUID("pt9jtdre2lvl0y12")
		assert.Nil(t, err)
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})

	t.Run("no photo found", func(t *testing.T) {
		result, err := PreloadPhotoByUUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}
