package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// reportRows indexes a Report result by field name for easier assertions.
func reportRows(t *testing.T) map[string][]string {
	t.Helper()

	rows, cols := Report(&Data{})

	assert.Equal(t, []string{"Field", "Type", "Exiftool", "Adobe XMP", "DCMI"}, cols)
	assert.NotEmpty(t, rows)

	result := make(map[string][]string, len(rows))

	for _, row := range rows {
		assert.Len(t, row, len(cols))
		result[row[0]] = row
	}

	return result
}

func TestReport(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rows := reportRows(t)
		assert.Equal(t, "text", rows["Title"][1])
		assert.Equal(t, "timestamp", rows["TakenAt"][1])
		assert.Equal(t, "flag", rows["Flash"][1])
		assert.Equal(t, "list", rows["Keywords"][1])
	})
	t.Run("XmpTags", func(t *testing.T) {
		rows := reportRows(t)
		assert.Equal(t, "dc:title, photoshop:Headline", rows["Title"][3])
		assert.Equal(t, "photoshop:DateCreated, exif:DateTimeOriginal, xmp:CreateDate", rows["TakenAt"][3])
		assert.Equal(t, "tiff:Model", rows["CameraModel"][3])
		assert.Equal(t, "exifEX:LensModel, aux:Lens, aux:LensID", rows["LensModel"][3])
		assert.Equal(t, "xmpRights:UsageTerms", rows["License"][3])
	})
	t.Run("GpsFromXmp", func(t *testing.T) {
		// The XMP reader supplies coordinates, so these must not report as
		// Exiftool-only, which would read as "XMP GPS is unsupported".
		rows := reportRows(t)
		assert.Equal(t, "exif:GPSLatitude", rows["GPSLatitude"][3])
		assert.Equal(t, "exif:GPSLongitude", rows["GPSLongitude"][3])
		assert.Equal(t, "exif:GPSAltitude", rows["Altitude"][3])
	})
	t.Run("DcmiTags", func(t *testing.T) {
		rows := reportRows(t)
		assert.Equal(t, "title, title.Alt", rows["Title"][4])
		assert.Equal(t, "description, description.Alt", rows["Caption"][4])
		assert.Equal(t, "rights, rights.Alt", rows["Copyright"][4])
		assert.Equal(t, "creator", rows["Artist"][4])
		assert.Equal(t, "subject", rows["Subject"][4])
		assert.Equal(t, "identifier", rows["DocumentID"][4])
	})
	t.Run("ExcludedFields", func(t *testing.T) {
		// Fields tagged meta:"-" or report:"-" have no row at all.
		rows := reportRows(t)
		for _, name := range []string{"Lat", "Lng", "TakenNs", "TimeZone", "Orientation", "MimeType", "Warning"} {
			assert.NotContains(t, rows, name)
		}
	})
	t.Run("KeywordsHaveNoXmpSource", func(t *testing.T) {
		// XMP contributes only derived keywords (flash/panorama/hdr); dc:subject
		// maps to Subject instead, so Keywords reports no XMP source.
		rows := reportRows(t)
		assert.Empty(t, rows["Keywords"][3])
		assert.NotEmpty(t, rows["Subject"][3])
	})
	t.Run("InvalidType", func(t *testing.T) {
		rows, cols := Report("not a struct")
		assert.Empty(t, rows)
		assert.Len(t, cols, 5)
	})
}
