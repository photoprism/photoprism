package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoQueryRating(t *testing.T) {
	query := form.NewSearchPhotosGeo("rating:4-5")

	if err := query.ParseQueryString(); err != nil {
		t.Fatal(err)
	}

	results, err := PhotosGeo(query)

	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, results)

	for _, result := range results {
		assert.GreaterOrEqual(t, result.PhotoRating, 4)
		assert.LessOrEqual(t, result.PhotoRating, 5)
	}
}
