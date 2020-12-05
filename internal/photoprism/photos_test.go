package photoprism

import (
	"testing"
)

func TestPhotos_Init(t *testing.T) {
	photos := NewPhotos()

	if err := photos.Init(); err != nil {
		t.Fatal(err)
	}
}
