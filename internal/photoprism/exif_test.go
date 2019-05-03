package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetExifData(t *testing.T) {
	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	image1, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")

	assert.Nil(t, err)

	info, err := image1.ExifData()

	assert.Empty(t, err)

	assert.IsType(t, &ExifData{}, info)

	assert.Equal(t, "iPhone SE", info.CameraModel)
}

func TestMediaFile_GetExifData_Slow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	image2, err := NewMediaFile(ctx.ImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	info, err := image2.ExifData()

	assert.Empty(t, err)

	assert.IsType(t, &ExifData{}, info)

	assert.Equal(t, "Canon EOS M10", info.CameraModel)
}
