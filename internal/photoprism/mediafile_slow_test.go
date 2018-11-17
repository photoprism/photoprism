// +build slow

package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetPerceptiveHash_Slow(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	mediaFile1, err := NewMediaFile(conf.GetImportPath() + "/20130203_193332_0AE340D280.jpg")
	assert.Nil(t, err)
	hash1, _ := mediaFile1.GetPerceptualHash()

	assert.Equal(t, "ef95", hash1)

	mediaFile2, err := NewMediaFile(conf.GetImportPath() + "/20130203_193332_0AE340D280_V2.jpg")
	assert.Nil(t, err)
	hash2, _ := mediaFile2.GetPerceptualHash()

	assert.Equal(t, "6f95", hash2)

	distance, _ := mediaFile1.GetPerceptualDistance(hash2)

	assert.Equal(t, 1, distance)

	mediaFile3, err := NewMediaFile(conf.GetImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)
	hash3, _ := mediaFile3.GetPerceptualHash()

	assert.Equal(t, "ad73", hash3)

	distance, _ = mediaFile1.GetPerceptualDistance(hash3)

	assert.Equal(t, 7, distance)
}
