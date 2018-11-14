// +build slow

package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetExifData_Slow(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	image2, err := NewMediaFile(conf.GetImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	info, err := image2.GetExifData()

	assert.Empty(t, err)

	assert.IsType(t, &ExifData{}, info)

	assert.Equal(t, "Canon EOS M10", info.CameraModel)
}
