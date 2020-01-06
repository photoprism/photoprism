package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

// Example for using database fixtures defined in assets/resources/examples/fixtures.sql
func TestCamera_FirstOrCreate(t *testing.T) {
	t.Run("iphone_5", func(t *testing.T) {
		camera := entity.NewCamera("iPhone 5", "Apple")
		c := config.TestConfig()
		camera.FirstOrCreate(c.Db())
		assert.Equal(t, "TEST FIXTURE", camera.CameraNotes)
	})
}
