package photoprism

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/geo/s2"
	"github.com/stretchr/testify/assert"
)

func TestPhotos_Init(t *testing.T) {
	photos := NewPhotos()

	if err := photos.Init(); err != nil {
		t.Fatal(err)
	}
}

func TestPhotos_Find(t *testing.T) {
	photos := NewPhotos()
	if err := photos.Init(); err != nil {
		t.Fatal(err)
	}

	r := photos.Find(time.Date(2020, 11, 11, 9, 7, 18, 0, time.UTC), s2.TokenPrefix+"85d1ea7d382")
	assert.Equal(t, uint(0), r)
}
