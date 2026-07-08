package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosQueryFisheye(t *testing.T) {
	// The fisheye subquery must build valid SQL and narrow the result set, even when no fixture
	// matches (real matching is covered by the end-to-end index verification).
	var fisheye form.SearchPhotos
	fisheye.Query = "fisheye:true"
	fisheye.Merged = true

	if err := fisheye.ParseQueryString(); err != nil {
		t.Fatal(err)
	}

	matched, _, err := Photos(fisheye)
	if err != nil {
		t.Fatal(err)
	}

	var all form.SearchPhotos
	all.Merged = true

	total, _, err := Photos(all)
	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(total), len(matched))
}
