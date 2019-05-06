package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_RelatedFiles(t *testing.T) {
	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	mediaFile, err := NewMediaFile(ctx.ImportPath() + "/raw/20140717_154212_1EC48F8489.cr2")

	assert.Nil(t, err)

	expectedBaseFilename := ctx.ImportPath() + "/raw/20140717_154212_1EC48F8489"

	related, _, err := mediaFile.RelatedFiles()

	assert.Nil(t, err)

	assert.Len(t, related, 3)

	for _, result := range related {
		t.Logf("Filename: %s", result.Filename())

		filename := result.Filename()

		extension := result.Extension()

		baseFilename := filename[0 : len(filename)-len(extension)]

		assert.Equal(t, expectedBaseFilename, baseFilename)
	}
}

func TestMediaFile_RelatedFiles_Ordering(t *testing.T) {
	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	mediaFile, err := NewMediaFile(ctx.ImportPath() + "/20130203_193332_0AE340D280.jpg")

	assert.Nil(t, err)

	related, _, err := mediaFile.RelatedFiles()

	assert.Nil(t, err)

	assert.Len(t, related, 2)

	for _, result := range related {
		filename := result.Filename()
		t.Logf("Filename: %s", filename)
	}
}

func TestMediaFile_EditedFilename(t *testing.T) {
	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	mediaFile1, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)
	assert.Equal(t, ctx.ImportPath()+"/iphone/IMG_E6788.JPG", mediaFile1.EditedFilename())

	/* TODO: Add example files to import.zip
	mediaFile2, err := NewMediaFile("/foo/bar/IMG_E1234.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "", mediaFile2.EditedFilename())
	*/

	mediaFile3, err := NewMediaFile(ctx.ImportPath() + "/raw/20140717_154212_1EC48F8489.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "", mediaFile3.EditedFilename())
}

func TestMediaFile_MimeType(t *testing.T) {
	ctx := config.TestConfig()

	ctx.InitializeTestData(t)

	image1, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)
	assert.Equal(t, "image/jpeg", image1.MimeType())

	image2, err := NewMediaFile(ctx.ImportPath() + "/raw/20140717_154212_1EC48F8489.cr2")
	assert.Nil(t, err)
	assert.Equal(t, "application/octet-stream", image2.MimeType())
}

func TestMediaFile_Exists(t *testing.T) {
	ctx := config.TestConfig()

	mediaFile, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)
	assert.NotNil(t, mediaFile)
	assert.True(t, mediaFile.Exists())

	mediaFile, err = NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788_XYZ.JPG")
	assert.NotNil(t, err)
	assert.Nil(t, mediaFile)
}
