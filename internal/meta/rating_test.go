package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON_Rating(t *testing.T) {
	t.Run("Rated", func(t *testing.T) {
		data, err := JSON("testdata/rating.json", "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 4, data.Rating)
	})
}

func TestXMP_Rating(t *testing.T) {
	t.Run("AttributeForm", func(t *testing.T) {
		data, err := XMP("testdata/rating.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, data.Rating)
	})
	t.Run("RejectedIsUnrated", func(t *testing.T) {
		data, err := XMP("testdata/rating_rejected.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, data.Rating)
	})
}
