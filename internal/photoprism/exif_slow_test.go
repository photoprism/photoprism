// +build slow

package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetExifData_Slow(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	image2, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	info, err := image2.GetExifData()

	assert.Empty(t, err)

	assert.IsType(t, &ExifData{}, info)

	assert.Equal(t, "Canon EOS M10", info.CameraModel)
}
