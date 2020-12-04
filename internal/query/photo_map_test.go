package query

import (
	"testing"
)

func TestIndexedPhotos(t *testing.T) {
	result, err := IndexedPhotos()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("INDEXED Photos: %#v", result)
}
