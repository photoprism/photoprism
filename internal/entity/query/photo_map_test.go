package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestIndexedPhotos(t *testing.T) {
	entity.ValidateFixtures(t)
	result, err := IndexedPhotos()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("INDEXED Photos: %#v", result)
}
