package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
)

// visitorSessionToken resolves the auth token of the pre-seeded "visitor"
// session fixture; the fixture grants access to the shared album
// as6sg6bxpogaaba8 which contains ps6sg6be2lvl0yh7.
var visitorSessionToken = func() string {
	s := entity.SessionFixtures.Get("visitor")
	return s.AuthToken()
}()

// permittedSidebarKeys lists the JSON keys that BuildPhotoSidebar must
// always emit so the frontend can render the short view.
var permittedSidebarKeys = []string{
	"UID", "Type", "Title", "Caption",
	"TakenAt", "TakenAtLocal", "TimeZone", "Year", "Month", "Day",
	"Lat", "Lng",
}

// restrictedSidebarKeys lists the top-level JSON keys that must not appear
// in the reduced sidebar response. Nested Files[*] fields are not covered
// here — see the TODO on PhotoSidebar.Files.
var restrictedSidebarKeys = []string{
	"Iso", "Exposure", "FNumber", "FocalLength",
	"Altitude", "Country",
	"Camera", "CameraID", "CameraSerial",
	"Lens", "LensID",
	"Cell", "CellID", "Place", "PlaceID",
	"Details", "Labels", "Albums", "Keywords",
	"Path", "Description",
	"CreatedBy", "Faces",
}

// TestBuildPhotoSidebar verifies that the DTO copies the permitted fields
// from an entity.Photo and omits every field that must not leak to roles
// without full sidebar access.
func TestBuildPhotoSidebar(t *testing.T) {
	taken := time.Date(2024, 5, 15, 12, 30, 45, 0, time.UTC)

	src := entity.Photo{
		ID:               42,
		PhotoUID:         "ps6sg6be2lvl0yh9",
		PhotoType:        "image",
		PhotoTitle:       "Permitted Title",
		PhotoCaption:     "Permitted Caption",
		PhotoDescription: "Permitted Description",
		PhotoPath:        "2024/05",
		PhotoName:        "20240515_123045_ABCDEF01",
		OriginalName:     "IMG_0001.jpg",
		TakenAt:          taken,
		TakenAtLocal:     taken,
		TimeZone:         "UTC",
		PhotoYear:        2024,
		PhotoMonth:       5,
		PhotoDay:         15,
		PhotoLat:         52.5200,
		PhotoLng:         13.4050,
		PhotoAltitude:    123,
		PhotoCountry:     "de",
		PhotoIso:         400,
		PhotoExposure:    "1/250",
		PhotoFNumber:     2.8,
		PhotoFocalLength: 50,
		PhotoFaces:       3,
		PhotoDuration:    0,
		CameraID:         7,
		CameraSerial:     "SECRET-SERIAL-1234",
		LensID:           9,
		CreatedBy:        "u00000000000000a",
		Details: &entity.Details{
			PhotoID:   42,
			Keywords:  "hidden,keywords",
			Notes:     "Private notes",
			Subject:   "Subject line",
			Artist:    "Jane Photographer",
			Copyright: "All rights reserved",
			License:   "CC BY-NC 4.0",
		},
		Camera: &entity.Camera{CameraName: "Canon EOS R5", CameraMake: "Canon", CameraModel: "EOS R5"},
		Lens:   &entity.Lens{LensName: "RF 50mm F1.2L", LensMake: "Canon", LensModel: "RF 50mm F1.2L"},
		Cell:   &entity.Cell{ID: "s2:abcdef"},
		Place:  &entity.Place{ID: "de:berlin", PlaceLabel: "Berlin, Germany"},
		Labels: []entity.PhotoLabel{{PhotoID: 42, LabelID: 1}},
		Albums: []entity.Album{{AlbumTitle: "Vacation 2024"}},
	}

	dto := BuildPhotoSidebar(src)

	t.Run("PermittedFieldsCopied", func(t *testing.T) {
		assert.Equal(t, "ps6sg6be2lvl0yh9", dto.UID)
		assert.Equal(t, "image", dto.Type)
		assert.Equal(t, "Permitted Title", dto.Title)
		assert.Equal(t, "Permitted Caption", dto.Caption)
		assert.Equal(t, taken, dto.TakenAt)
		assert.Equal(t, taken, dto.TakenAtLocal)
		assert.Equal(t, "UTC", dto.TimeZone)
		assert.Equal(t, 2024, dto.Year)
		assert.Equal(t, 5, dto.Month)
		assert.Equal(t, 15, dto.Day)
		assert.InDelta(t, 52.5200, dto.Lat, 0.0001)
		assert.InDelta(t, 13.4050, dto.Lng, 0.0001)
	})

	t.Run("JSONOmitsRestrictedFields", func(t *testing.T) {
		buf, err := json.Marshal(dto)
		assert.NoError(t, err)

		for _, key := range permittedSidebarKeys {
			assert.Truef(t, gjson.GetBytes(buf, key).Exists(), "expected permitted key %q to be present", key)
		}
		for _, key := range restrictedSidebarKeys {
			assert.Falsef(t, gjson.GetBytes(buf, key).Exists(), "restricted key %q must not be present in JSON", key)
		}
	})
}

// TestBuildPhotoSidebar_EmptyPhoto ensures the builder handles a
// zero-valued entity.Photo without panicking and without leaking
// restricted keys via struct defaults.
func TestBuildPhotoSidebar_EmptyPhoto(t *testing.T) {
	dto := BuildPhotoSidebar(entity.Photo{})

	buf, err := json.Marshal(dto)
	assert.NoError(t, err)

	for _, key := range permittedSidebarKeys {
		assert.Truef(t, gjson.GetBytes(buf, key).Exists(), "expected permitted key %q to be present", key)
	}
	for _, key := range restrictedSidebarKeys {
		assert.Falsef(t, gjson.GetBytes(buf, key).Exists(), "restricted key %q must not be present", key)
	}
}
