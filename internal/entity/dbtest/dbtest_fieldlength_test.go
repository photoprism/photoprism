package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var editTime = time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC)
var deleteTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
var checkedTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

func TestInitDBLengths(t *testing.T) {
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	t.Run("PhotoMaxVarLengths", func(t *testing.T) {
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)
		expectedCount += 1

		// Prevent the creation of the child records as it prevents cleanup.
		result := stmt.Omit(clause.Associations).Create(n)
		assert.NoError(t, result.Error)

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error)

		assert.Equal(t, expectedCount, actualCount)
	})

	// Can't test PhotoUID as it's generated in code

	t.Run("PhotoExceedMaxTakenSrc", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "123456789",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "TakenSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})

	t.Run("PhotoExceedMaxPhotoType", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "123456789",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoType"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})

	t.Run("PhotoExceedMaxTypeSrc", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "123456789",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "TypeSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxTitleSrc", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "123456789",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "TitleSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoTitle", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 21),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoTitle"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoDescription", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "1234567",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoDescription"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoPath", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "12345",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoPath"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoName", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "123456",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoName"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxOriginalName", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "123456",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "OriginalName"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxTimeZone", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "12345",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "TimeZone"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPlaceSrc", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "123456789",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PlaceSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoExposure", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "12345",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoExposure"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxCameraSerial", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 17),
			CameraSrc:        "12345678",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "CameraSerial"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxCameraSrc", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "123456789",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "CameraSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})

	t.Run("PhotoExceedMaxCameraSrcUniCode", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "゠ァアィカケパドョ",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)

		result := stmt.Omit(clause.Associations).Create(n)
		assert.Error(t, result.Error, "Create record")
		if result.Error != nil {

			if strings.Contains(entity.DbDialect(), "mysql") {
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "CameraSrc"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value to long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})

	t.Run("PhotoCameraSrcUniCode", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:         "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:          time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:         "12345678",
			PhotoType:        "12345678",
			TypeSrc:          "12345678",
			PhotoTitle:       strings.Repeat("1234567890", 20),
			TitleSrc:         "12345678",
			PhotoDescription: strings.Repeat("1234567890", 409) + "123456",
			DescriptionSrc:   "12345678",
			PhotoPath:        strings.Repeat("1234567890", 102) + "1234",
			PhotoName:        strings.Repeat("1234567890", 25) + "12345",
			OriginalName:     strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite:    false,
			//PhotoSingle
			PhotoPrivate:     false,
			PhotoScan:        false,
			PhotoPanorama:    false,
			TimeZone:         strings.Repeat("1234567890", 6) + "1234",
			Place:            &entity.UnknownPlace,
			PlaceID:          entity.UnknownPlace.ID,
			PlaceSrc:         "12345678",
			Cell:             &entity.UnknownLocation,
			CellID:           entity.UnknownLocation.ID,
			CellAccuracy:     0,
			PhotoAltitude:    0,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoCountry:     entity.UnknownPlace.CountryCode(),
			PhotoYear:        2790,
			PhotoMonth:       7,
			PhotoDay:         4,
			PhotoIso:         200,
			PhotoExposure:    strings.Repeat("1234567890", 6) + "1234",
			PhotoFNumber:     5,
			PhotoFocalLength: 50,
			PhotoQuality:     3,
			PhotoResolution:  2,
			Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
			CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
			CameraSerial:     strings.Repeat("1234567890", 16),
			CameraSrc:        "゠ァ", // Unicode is using 4 bytes per character "゠ァアィカケパド",
			Lens:             entity.LensFixtures.Pointer("lens-f-380"),
			LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
			Details:          entity.DetailsFixtures.Pointer("lake", 1000000),
			Keywords: []entity.Keyword{
				entity.KeywordFixtures.Get("bridge"),
			},
			Albums: []entity.Album{
				entity.AlbumFixtures.Get("holiday-2030"),
			},
			Files: []entity.File{},
			Labels: []entity.PhotoLabel{
				entity.LabelFixtures.PhotoLabel(1000000, "flower", 38, "image"),
				entity.LabelFixtures.PhotoLabel(1000000, "cake", 38, "manual"),
			},
			CreatedAt:  time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
			EditedAt:   nil,
			CheckedAt:  &checkedTime,
			DeletedAt:  gorm.DeletedAt{},
			PhotoColor: 9,
			PhotoStack: 0,
			PhotoFaces: 3}

		stmt := entity.Db()

		expectedCount := int64(0)
		stmt.Model(m).Count(&expectedCount)
		expectedCount += 1

		result := stmt.Omit(clause.Associations).Create(n)
		assert.NoError(t, result.Error, "Create record")
		if result.Error != nil {
			log.Errorf("Error detected %v", result.Error)
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})

}
