package thumb

import (
	"image"
	"image/color"
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

func TestBytes_GByte(t *testing.T) {
	var b Bytes = 3 * GB
	assert.Equal(t, 3.0, b.GByte())
}

func TestMemSize_ColorModels(t *testing.T) {
	// 10x10 grayscale: 100 bytes
	g := image.NewGray(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(100), MemSize(g))
	// 10x10 gray16: 200 bytes
	g16 := image.NewGray16(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(200), MemSize(g16))
	// 10x10 rgba: 400 bytes
	rgba := image.NewRGBA(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(400), MemSize(rgba))
	// 10x10 rgba64: 800 bytes
	rgba64 := image.NewRGBA64(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(800), MemSize(rgba64))
	// Alpha-only: 100 bytes
	a := image.NewAlpha(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(100), MemSize(a))
	// Alpha16: 200 bytes
	a16 := image.NewAlpha16(image.Rect(0, 0, 10, 10))
	assert.Equal(t, Bytes(200), MemSize(a16))
	// Custom image with NRGBA color model still reports by bounds*BPP switch; use NRGBA to hit 4 bytes path
	nr := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	_ = nr           // already covered by rgba
	_ = color.RGBA{} // avoid unused import warning if any
}
