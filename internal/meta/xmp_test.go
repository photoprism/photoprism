package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"trimmer.io/go-xmp/xmp"
)

func TestXMP(t *testing.T) {
	data, err := XMP("testdata/IMG_20200101_172822.xmp")

	if err != nil {
		t.Fatal(err)
	}

	info := data.exifInfo

	assert.Equal(t, "Michael Mayer", info.Artist)
	assert.Equal(t, xmp.StringList{"Michael Mayer"}, info.ArtistXMP)
	assert.Equal(t, "This is an (edited) legal notice", info.Copyright)
	assert.Equal(t, "2020-01-01T17:28:23Z", info.DateTime.Value().Format("2006-01-02T15:04:05Z"))
	assert.Equal(t, "Example file for development", info.ImageDescription)
	assert.Equal(t, 2736, info.ImageLength)
	assert.Equal(t, 3648, info.ImageWidth)
	assert.Equal(t, "HUAWEI", info.Make)
	assert.Equal(t, "ELE-L29", info.Model)
	assert.Equal(t, 0, int(info.Orientation))
	assert.Equal(t, 1.0, info.GPSAltitude.Value()) // TODO: Is this correct?

	// TODO: Values are empty - why?
	// assert.Equal(t, "52.459690093888895", info.GPSLongitudeCoord.Value())
	// assert.Equal(t, "13.321831703055555", info.GPSLatitudeCoord.Value())
	// assert.Equal(t, 27, info.FocalLengthIn35mmFilm)

	assert.Equal(t, "Michael Mayer", data.Artist)
	assert.Equal(t, "2020-01-01T17:28:23Z", data.TakenAtLocal.Format("2006-01-02T15:04:05Z"))

	// TODO: Is wrong because lat / lng are missing (value empty)
	// assert.Equal(t, "2020-01-01T16:28:23Z", data.TakenAt.Format("2006-01-02T15:04:05Z"))

	assert.Equal(t, "Example file for development", data.Description)
	assert.Equal(t, "This is an (edited) legal notice", data.Copyright)
	assert.Equal(t, 2736, data.Height)
	assert.Equal(t, 3648, data.Width)
	assert.Equal(t, "HUAWEI", data.CameraMake)
	assert.Equal(t, "ELE-L29", data.CameraModel)
	assert.Equal(t, 0, data.Orientation)

	// TODO: Values are empty - why?
	// assert.Equal(t, 27, data.FocalLength)
	// assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
}
