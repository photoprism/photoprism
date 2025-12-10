package batch

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewPhotosForm(t *testing.T) {
	t.Run("FromJSON", runNewPhotosFormFromJSON)
	t.Run("FromFixturesAlbumsLabels", runNewPhotosFormFromFixtures)
	t.Run("TwoPhotosMixedFlags1", runNewPhotosFormMixedFlags1)
	t.Run("TwoPhotosMixedFlags2", runNewPhotosFormMixedFlags2)
}

// runNewPhotosFormFromJSON exercises PhotosForm behavior.
func runNewPhotosFormFromJSON(t *testing.T) {
	var photos search.PhotoResults

	dataFile := fs.Abs("./testdata/photos.json")
	data, err := os.ReadFile(dataFile) // #nosec G304 test fixture path is static
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(data, &photos); err != nil {
		t.Fatal(err)
	}

	// Avoid DB access in unit tests by clearing PhotoUIDs (skips preload)
	for i := range photos {
		photos[i].PhotoUID = ""
	}

	frm := NewPhotosForm(photos)

	// Photo metadata.
	assert.Equal(t, "", frm.PhotoType.Value)
	assert.Equal(t, true, frm.PhotoType.Mixed)
	assert.Equal(t, "", frm.PhotoTitle.Value)
	assert.Equal(t, true, frm.PhotoTitle.Mixed)
	assert.Equal(t, "", frm.PhotoCaption.Value)
	assert.Equal(t, true, frm.PhotoCaption.Mixed)
	assert.Equal(t, false, frm.PhotoFavorite.Value)
	assert.Equal(t, true, frm.PhotoFavorite.Mixed)
	assert.Equal(t, false, frm.PhotoPrivate.Value)
	assert.Equal(t, false, frm.PhotoPrivate.Mixed)
	assert.Equal(t, 1000003, frm.CameraID.Value)
	assert.Equal(t, false, frm.CameraID.Mixed)
	assert.Equal(t, 1000000, frm.LensID.Value)
	assert.Equal(t, false, frm.LensID.Mixed)
	assert.Equal(t, 0, frm.PhotoIso.Value)
	assert.Equal(t, true, frm.PhotoIso.Mixed)
	assert.Equal(t, float32(0), frm.PhotoFNumber.Value)
	assert.Equal(t, true, frm.PhotoFNumber.Mixed)
	assert.Equal(t, 0, frm.PhotoFocalLength.Value)
	assert.Equal(t, true, frm.PhotoFocalLength.Mixed)
	assert.Equal(t, float64(0), frm.PhotoLat.Value)
	assert.Equal(t, true, frm.PhotoLat.Mixed)
	assert.Equal(t, float64(0), frm.PhotoLng.Value)
	assert.Equal(t, true, frm.PhotoLng.Mixed)
	assert.Equal(t, 0, frm.PhotoAltitude.Value)
	assert.Equal(t, true, frm.PhotoAltitude.Mixed)

	// Additional details.
	assert.Equal(t, "", frm.DetailsKeywords.Value)
	assert.Equal(t, true, frm.DetailsKeywords.Mixed)
	assert.Equal(t, "", frm.DetailsSubject.Value)
	assert.Equal(t, true, frm.DetailsSubject.Mixed)
	assert.Equal(t, "", frm.DetailsArtist.Value)
	assert.Equal(t, true, frm.DetailsArtist.Mixed)
	assert.Equal(t, "", frm.DetailsCopyright.Value)
	assert.Equal(t, true, frm.DetailsCopyright.Mixed)
	assert.Equal(t, "", frm.DetailsLicense.Value)
	assert.Equal(t, true, frm.DetailsLicense.Mixed)
}

// runNewPhotosFormFromFixtures ensures preloaded fixtures behave.
func runNewPhotosFormFromFixtures(t *testing.T) {

	// Ensure test config and fixtures/DB are initialized.
	_ = config.TestConfig()

	// Build minimal PhotoResults with only UIDs set; NewPhotosForm will preload details.
	photos := search.PhotoResults{
		{PhotoUID: "pqkm36fjqvset9uz"},
		{PhotoUID: "pqkm36fjqvset9uy"},
	}

	frm := NewPhotosForm(photos)

	// Expect albums and labels collected from fixtures to be non-empty.
	if assert.NotNil(t, frm) {
		assert.Greater(t, len(frm.Albums.Items), 0, "expected at least one album item")
		assert.Greater(t, len(frm.Labels.Items), 0, "expected at least one label item")
		// Titles should be non-empty for the first items.
		assert.NotEmpty(t, frm.Albums.Items[0].Title)
		assert.NotEmpty(t, frm.Labels.Items[0].Title)
	}
}

// runNewPhotosFormMixedFlags1 captures mixed flag handling.
func runNewPhotosFormMixedFlags1(t *testing.T) {
	photo1 := search.Photo{
		ID:            111115411,
		PhotoUID:      "",
		PhotoType:     "image",
		PhotoTitle:    "Same Title",
		PhotoCountry:  "de",
		PhotoPrivate:  true,
		PhotoPanorama: true,
		PhotoScan:     true,
		PhotoFavorite: false,
		CameraID:      1,
		LensID:        2,
		PhotoAltitude: -10,
		PhotoLat:      48.519234,
		PhotoLng:      9.057997,
		PhotoDay:      4,
		PhotoMonth:    5,
		PhotoYear:     2021,
	}

	photo2 := search.Photo{
		ID:            111115511,
		CreatedAt:     time.Time{},
		TakenAt:       time.Time{},
		TakenAtLocal:  time.Time{},
		PhotoUID:      "",
		PhotoType:     "image",
		PhotoTitle:    "Same Title",
		PhotoCountry:  "ca",
		PhotoPrivate:  false,
		PhotoPanorama: false,
		PhotoScan:     false,
		PhotoFavorite: true,
		CameraID:      1,
		LensID:        2,
		PhotoAltitude: -15,
		PhotoLat:      48.519234,
		PhotoLng:      9.057997,
		PhotoDay:      3,
		PhotoMonth:    5,
		PhotoYear:     2020,
	}

	photos := search.PhotoResults{photo1, photo2}
	frm := NewPhotosForm(photos)

	// Photo metadata.
	assert.Equal(t, "image", frm.PhotoType.Value)
	assert.Equal(t, false, frm.PhotoType.Mixed)
	assert.Equal(t, "Same Title", frm.PhotoTitle.Value)
	assert.Equal(t, false, frm.PhotoTitle.Mixed)
	assert.Equal(t, "", frm.PhotoCaption.Value)
	assert.Equal(t, false, frm.PhotoCaption.Mixed)
	assert.Equal(t, false, frm.PhotoFavorite.Value)
	assert.Equal(t, true, frm.PhotoFavorite.Mixed)
	assert.Equal(t, false, frm.PhotoPrivate.Value)
	assert.Equal(t, true, frm.PhotoPrivate.Mixed)
	assert.Equal(t, false, frm.PhotoScan.Value)
	assert.Equal(t, true, frm.PhotoScan.Mixed)
	assert.Equal(t, false, frm.PhotoPanorama.Value)
	assert.Equal(t, true, frm.PhotoPanorama.Mixed)
	assert.Equal(t, false, frm.CameraID.Mixed)
	assert.Equal(t, 1, frm.CameraID.Value)
	assert.Equal(t, 2, frm.LensID.Value)
	assert.Equal(t, false, frm.LensID.Mixed)
	assert.Equal(t, "", frm.PhotoCountry.Value)
	assert.Equal(t, true, frm.PhotoCountry.Mixed)
	assert.Equal(t, 0, frm.PhotoIso.Value)
	assert.Equal(t, false, frm.PhotoIso.Mixed)
	assert.Equal(t, float32(0), frm.PhotoFNumber.Value)
	assert.Equal(t, false, frm.PhotoFNumber.Mixed)
	assert.Equal(t, 0, frm.PhotoFocalLength.Value)
	assert.Equal(t, false, frm.PhotoFocalLength.Mixed)
	assert.Equal(t, 48.519234, frm.PhotoLat.Value)
	assert.Equal(t, false, frm.PhotoLat.Mixed)
	assert.Equal(t, 9.057997, frm.PhotoLng.Value)
	assert.Equal(t, false, frm.PhotoLng.Mixed)
	assert.Equal(t, 0, frm.PhotoAltitude.Value)
	assert.Equal(t, true, frm.PhotoAltitude.Mixed)
	assert.Equal(t, -2, frm.PhotoDay.Value)
	assert.Equal(t, true, frm.PhotoDay.Mixed)
	assert.Equal(t, 5, frm.PhotoMonth.Value)
	assert.Equal(t, false, frm.PhotoMonth.Mixed)
	assert.Equal(t, -2, frm.PhotoYear.Value)
	assert.Equal(t, true, frm.PhotoYear.Mixed)
}

// runNewPhotosFormMixedFlags2 covers camera/lens variance.
func runNewPhotosFormMixedFlags2(t *testing.T) {
	photo1 := search.Photo{
		ID:            111115411,
		PhotoUID:      "",
		PhotoType:     "image",
		PhotoTitle:    "Same Title",
		PhotoCountry:  "de",
		CameraID:      1000001,
		LensID:        1000001,
		PhotoDay:      3,
		PhotoMonth:    5,
		PhotoYear:     2020,
		PhotoAltitude: 105,
	}

	photo2 := search.Photo{
		ID:            111115511,
		CreatedAt:     time.Time{},
		TakenAt:       time.Time{},
		TakenAtLocal:  time.Time{},
		PhotoUID:      "",
		PhotoType:     "image",
		PhotoTitle:    "Same Title",
		PhotoCountry:  "",
		CameraID:      1000000,
		LensID:        1000000,
		PhotoDay:      3,
		PhotoMonth:    6,
		PhotoYear:     2020,
		PhotoAltitude: 105,
	}

	photos := search.PhotoResults{photo1, photo2}
	frm := NewPhotosForm(photos)

	// Photo metadata.
	assert.Equal(t, "image", frm.PhotoType.Value)
	assert.Equal(t, false, frm.PhotoType.Mixed)
	assert.Equal(t, "Same Title", frm.PhotoTitle.Value)
	assert.Equal(t, false, frm.PhotoTitle.Mixed)
	assert.Equal(t, "", frm.PhotoCaption.Value)
	assert.Equal(t, false, frm.PhotoCaption.Mixed)
	assert.Equal(t, false, frm.PhotoFavorite.Value)
	assert.Equal(t, false, frm.PhotoFavorite.Mixed)
	assert.Equal(t, false, frm.PhotoPrivate.Value)
	assert.Equal(t, false, frm.PhotoPrivate.Mixed)
	assert.Equal(t, false, frm.PhotoScan.Value)
	assert.Equal(t, false, frm.PhotoScan.Mixed)
	assert.Equal(t, false, frm.PhotoPanorama.Value)
	assert.Equal(t, false, frm.PhotoPanorama.Mixed)
	assert.Equal(t, true, frm.CameraID.Mixed)
	assert.Equal(t, -2, frm.CameraID.Value)
	assert.Equal(t, true, frm.LensID.Mixed)
	assert.Equal(t, -2, frm.LensID.Value)
	assert.Equal(t, "", frm.PhotoCountry.Value)
	assert.Equal(t, true, frm.PhotoCountry.Mixed)
	assert.Equal(t, 0, frm.PhotoIso.Value)
	assert.Equal(t, false, frm.PhotoIso.Mixed)
	assert.Equal(t, float32(0), frm.PhotoFNumber.Value)
	assert.Equal(t, false, frm.PhotoFNumber.Mixed)
	assert.Equal(t, 0, frm.PhotoFocalLength.Value)
	assert.Equal(t, false, frm.PhotoFocalLength.Mixed)
	assert.Equal(t, 105, frm.PhotoAltitude.Value)
	assert.Equal(t, false, frm.PhotoAltitude.Mixed)
	assert.Equal(t, 3, frm.PhotoDay.Value)
	assert.Equal(t, false, frm.PhotoDay.Mixed)
	assert.Equal(t, -2, frm.PhotoMonth.Value)
	assert.Equal(t, true, frm.PhotoMonth.Mixed)
	assert.Equal(t, 2020, frm.PhotoYear.Value)
	assert.Equal(t, false, frm.PhotoYear.Mixed)
}

func TestNewPhotosFormWithEntities(t *testing.T) {
	t.Run("FallsBack", runNewPhotosFormWithEntitiesFallsBack)
}

// runNewPhotosFormWithEntitiesFallsBack ensures helper falls back correctly.
func runNewPhotosFormWithEntitiesFallsBack(t *testing.T) {
	_ = config.TestConfig()
	photos := search.PhotoResults{
		{PhotoUID: "pqkm36fjqvset9uz"},
		{PhotoUID: "pqkm36fjqvset9uy"},
	}

	legacy := NewPhotosForm(photos)
	withNil := NewPhotosFormWithEntities(photos, nil)

	if assert.NotNil(t, legacy) && assert.NotNil(t, withNil) {
		assert.Equal(t, legacy.PhotoTitle.Value, withNil.PhotoTitle.Value)
		assert.Equal(t, len(legacy.Albums.Items), len(withNil.Albums.Items))
		assert.Equal(t, len(legacy.Labels.Items), len(withNil.Labels.Items))
	}
}
