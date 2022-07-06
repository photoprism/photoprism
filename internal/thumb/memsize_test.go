package thumb

import (
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
)

func TestMemSize(t *testing.T) {
	src := "testdata/example.jpg"

	assert.FileExists(t, src)

	img, err := imaging.Open(src, imaging.AutoOrientation(true))

	if err != nil {
		t.Fatal(err)
	}

	result := MemSize(img)

	assert.InEpsilon(t, 1464, result.KByte(), 1)
	assert.InEpsilon(t, 1.430511474, result.MByte(), 0.1)
	assert.Equal(t, "1.5 MB", result.String())
}
