package places

import (
	"testing"

	"github.com/photoprism/photoprism/internal/s2"
	"github.com/stretchr/testify/assert"
)

func TestFindLocation(t *testing.T) {
	t.Run("U Berliner Rathaus", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l, err := FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, l.Cached)
		assert.Equal(t, "Alt-Berlin", l.Name())
		assert.Equal(t, "Berlin", l.City())
		assert.Equal(t, "de", l.CountryCode())
	})
}
