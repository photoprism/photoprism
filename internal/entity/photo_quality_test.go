package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPhoto_QualityScore(t *testing.T) {
	t.Run("PhotoFixture19800101_000002_D640C559", func(t *testing.T) {
		assert.Equal(t, 3, PhotoFixtures.Pointer("19800101_000002_D640C559").QualityScore())
	})
	t.Run("PhotoFixturePhoto01 - favorite true - taken at before 2008", func(t *testing.T) {
		assert.Equal(t, 7, PhotoFixtures.Pointer("Photo01").QualityScore())
	})
	t.Run("PhotoFixturePhoto06 - taken at after 2012 - resolution 2", func(t *testing.T) {
		assert.Equal(t, 3, PhotoFixtures.Pointer("Photo06").QualityScore())
	})
	t.Run("PhotoFixturePhoto07 - score < 3 bit edited", func(t *testing.T) {
		assert.Equal(t, 3, PhotoFixtures.Pointer("Photo07").QualityScore())
	})
	t.Run("PhotoFixturePhoto15 - description with non-photographic", func(t *testing.T) {
		assert.Equal(t, 2, PhotoFixtures.Pointer("Photo15").QualityScore())
	})
	// digikam test that fails in gorm2.
	t.Run("digikam test", func(t *testing.T) {
		//digikam := NewPhoto(true)
		digikam := Photo{ //ID: 9800000,
			//PhotoUID:         "PSKS0MPLV8V124RH",
			TakenAt:          time.Date(2020, 10, 17, 17, 48, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2020, 10, 17, 12, 48, 0, 0, time.UTC),
			TakenSrc:         "meta",
			PhotoType:        "image",
			TypeSrc:          "",
			PhotoTitle:       "Bismarckviertel / Berlin / 2020",
			TitleSrc:         "",
			PhotoDescription: "",
			DescriptionSrc:   "meta",
			PhotoPath:        "2020/10",
			PhotoName:        "20201017_154824_43C25EB3",
			OriginalName:     "digikam",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         "Europe/Berlin",
			Place:            PlaceFixtures.Pointer("Germany"),
			PlaceID:          PlaceFixtures.Pointer("Germany").ID,
			PlaceSrc:         "xmp",
			Cell:             CellFixtures.Pointer("Neckarbrücke"),
			CellID:           CellFixtures.Pointer("Neckarbrücke").ID,
			CellAccuracy:     0,
			PhotoAltitude:    84,
			PhotoLat:         52.46052169777778,
			PhotoLng:         13.331401824722223,
			PhotoCountry:     "de",
			PhotoYear:        2020,
			PhotoMonth:       10,
			PhotoDay:         17,
			PhotoIso:         100,
			PhotoExposure:    "1/50",
			PhotoFNumber:     1.8,
			PhotoFocalLength: 27,
			PhotoQuality:     0,
			PhotoResolution:  0,
			Camera:           CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     "",
			CameraSrc:        "meta",
			Lens:             LensFixtures.Pointer("lens-f-380"),
			LensID:           LensFixtures.Pointer("lens-f-380").ID,
			Details:          nil, // DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []Keyword{
				KeywordFixtures.Get("bridge"),
			},
			Albums: []Album{
				AlbumFixtures.Get("holiday-2030"),
			},
			Files: []File{},
			/*			Labels: []PhotoLabel{
						LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
						LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
					},*/
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			EditedAt:   nil,
			CheckedAt:  nil,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 0,
		}
		// Create a valid record
		digikam.Create()
		// validate that the QualityScore is correct
		assert.Equal(t, 3, digikam.QualityScore())

		beforeId := uint(0)
		if res := Db().Raw("SELECT COALESCE(MAX(id),0) FROM `errors`").Scan(&beforeId); res.Error != nil {
			t.Log(res.Error)
			t.FailNow()
		}

		// Replicate the data scenario for the json file.
		// As created in UserMediaFile for a non primary file
		digikam.Camera = &UnknownCamera
		digikam.Cell = &UnknownLocation
		digikam.Place = &UnknownPlace
		// Save the record, which doesn't do anything in GormV1
		digikam.Save()

		// Validate that the QualityScore stays the same.
		assert.Equal(t, 3, digikam.QualityScore())

		// Clear the 3 errors that were created.
		expectedError := "%threw photo.Save has inconsistent%"
		if res := Db().Where("id > ? AND error_message like ?", beforeId, expectedError).Delete(&Error{}); res.Error != nil {
			t.Log(res.Error)
			t.FailNow()
		} else {
			assert.Equal(t, int64(3), res.RowsAffected)
		}
	})
}

func TestPhoto_UpdateQuality(t *testing.T) {
	t.Run("Hidden", func(t *testing.T) {
		p := &Photo{PhotoQuality: -1} // ToDo: Does this need ID: 1,?
		err := p.UpdateQuality()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, -1, p.PhotoQuality)
	})
	t.Run("Favorite", func(t *testing.T) {
		p := &Photo{PhotoQuality: 0, PhotoFavorite: true}
		/*
			p := &Photo{ID: 1, PhotoQuality: 0, PhotoFavorite: true}
			Db().Create(p)
			// Make it look like the gorm1 tests as they aren't updated by BeforeCreate
			p.TakenAt = time.Date(0000, 1, 1, 0, 0, 0, 0, time.UTC)
		*/
		err := p.UpdateQuality()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 4, p.PhotoQuality)
	})

	t.Run("no PK provided", func(t *testing.T) {
		p := &Photo{PhotoQuality: 0, PhotoFavorite: true}
		err := p.UpdateQuality()
		assert.ErrorContains(t, err, "No PK provided")
	})
}
