package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestCounts_Refresh(t *testing.T) {
	entity.ValidateFixtures(t)
	counts := Counts{Cameras: 0, Photos: 0}
	assert.Equal(t, counts.Cameras, 0)
	assert.Equal(t, counts.Photos, 0)
	counts.Refresh()
	assert.Greater(t, counts.Cameras, 0)
	assert.Greater(t, counts.Photos, 0)
	assert.Greater(t, counts.Albums, 0)

}
