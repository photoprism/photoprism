package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotos_Photos(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {

		photo1 := PhotoFixtures.Get("Photo08")
		photo2 := PhotoFixtures.Get("Photo07")

		photos := Photos{photo1, photo2}

		r := photos.Photos()

		assert.Equal(t, 2, len(r))
	})
}
