package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_TimeZone(t *testing.T) {
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		img, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		assert.Nil(t, err)

		zone, err := img.TimeZone()

		assert.Nil(t, err)
		assert.Equal(t, "Africa/Johannesburg", zone)
	})
}
