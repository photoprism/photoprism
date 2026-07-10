package meta

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestXMP_OverridesPreExistingMetaValues covers the meta-layer half of
// the SrcXmp > SrcMeta contract: every value the sidecar supplies
// overwrites the stale embedded-path value. Per-field SrcMeta → SrcXmp
// transitions are enforced by the entity setters that consume meta.Data.
func TestXMP_OverridesPreExistingMetaValues(t *testing.T) {
	data := Data{
		Title:       "Stale Embedded Title",
		Caption:     "Stale Embedded Caption",
		Lat:         40.0,
		Lng:         -74.0,
		Altitude:    99.9,
		CameraMake:  "WrongMake",
		CameraModel: "WrongModel",
	}

	// multi-rdf-description.xmp spreads ten fields across four sibling
	// rdf:Description blocks.
	if err := data.XMP("testdata/xmp/synthetic/multi-rdf-description.xmp"); err != nil {
		t.Fatalf("data.XMP: %v", err)
	}

	// Sidecar values must win.
	assert.Equal(t, "Multi-Description Fixture", data.Title)
	assert.Equal(t, "PhotoPrism", data.Artist)
	assert.Equal(t, "SyntheticCam", data.CameraMake)
	assert.Equal(t, "SC-1 Mark II", data.CameraModel)
	assert.InDelta(t, 52.5076, data.Lat, 1e-3)
	assert.InDelta(t, 13.4095, data.Lng, 1e-3)
	assert.InDelta(t, 34.5, data.Altitude, 1e-3)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", data.DocumentID)
	assert.Equal(t, "xmp.iid:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", data.InstanceID)

	// Caption was absent from the fixture; the reader only overwrites
	// when it has a non-empty replacement, so the stale value stays.
	assert.Equal(t, "Stale Embedded Caption", data.Caption)
}

// TestXMP_FillsPreviouslyEmptyFields covers the partial-sidecar →
// full-sidecar migration: a DB written by the old reader had only the
// four ✅ preserved fields set; the new reader fills in License, GPS,
// xmpMM IDs and the rest of the high-priority fields on re-index.
func TestXMP_FillsPreviouslyEmptyFields(t *testing.T) {
	data := Data{
		Title:   "XMP Test - Aurora",
		Caption: "Test fixture for digiKam XMP sidecar",
	}

	if err := data.XMP("testdata/xmp/digikam/aurora.jpg.xmp"); err != nil {
		t.Fatalf("data.XMP: %v", err)
	}

	// Previously-set values must round-trip without corruption.
	assert.Equal(t, "XMP Test - Aurora", data.Title)
	assert.Equal(t, "Test fixture for digiKam XMP sidecar", data.Caption)

	// Previously-empty fields must now be populated by the new reader.
	assert.Equal(t, "CC-BY-SA 4.0", data.License)
	assert.Equal(t, "254738CA43CD69C01101874D65B006B4", data.DocumentID)
	assert.Equal(t, "xmp.iid:de0fb44d-bdc8-4cc3-aa50-9aefe4992b34", data.InstanceID)
	assert.Equal(t, "(C) 2026 PhotoPrism — Test fixture", data.Copyright)
	assert.Equal(t, "sRGB IEC61966-2.1", data.ColorProfile)
	assert.NotEmpty(t, data.Subject, "dc:subject Bag must populate Subject")
	assert.Empty(t, data.Keywords, "dc:subject must not populate Keywords")
}

// TestXMP_GpsCompositionEndToEnd asserts that a 2-component Adobe-form
// sidecar populates Lat/Lng/Altitude as floats while leaving the
// ExifTool-format GPSLatitude/GPSLongitude string fields untouched.
func TestXMP_GpsCompositionEndToEnd(t *testing.T) {
	// Pre-populated ExifTool-format strings must survive the XMP path.
	data := Data{
		GPSLatitude:  `40 deg 25' 12.0" N`,
		GPSLongitude: `74 deg 0' 21.6" W`,
	}

	tmp := filepath.Join(t.TempDir(), "adobe-gps.xmp")
	if err := writeFile(t, tmp, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:GPSLatitude>52,30.4567N</exif:GPSLatitude>
   <exif:GPSLongitude>13,24.5678E</exif:GPSLongitude>
   <exif:GPSAltitude>3450/100</exif:GPSAltitude>
   <exif:GPSAltitudeRef>0</exif:GPSAltitudeRef>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`); err != nil {
		t.Fatal(err)
	}
	if err := data.XMP(tmp); err != nil {
		t.Fatalf("data.XMP: %v", err)
	}

	// Float fields populated from the XMP path.
	assert.InDelta(t, 52.5076, data.Lat, 1e-3)
	assert.InDelta(t, 13.4095, data.Lng, 1e-3)
	assert.InDelta(t, 34.5, data.Altitude, 1e-3)

	// String fields keep the ExifTool DMS format — the XMP path
	// populates Lat/Lng/Altitude floats only.
	assert.Equal(t, `40 deg 25' 12.0" N`, data.GPSLatitude)
	assert.Equal(t, `74 deg 0' 21.6" W`, data.GPSLongitude)
}
