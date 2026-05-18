package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var checkedTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

func TestInitDBLengths(t *testing.T) {
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()
	log.Info("Expect data to long Error or SQLSTATE from dbtest_fieldlength_test")

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
		expectedCount++

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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
				assert.Contains(t, result.Error.Error(), "value too long")
			}
		}

		actualCount := int64(0)
		stmt.Model(m).Count(&actualCount)

		// Cleanup, Skip soft delete!
		result2 := entity.UnscopedDb().Delete(n)
		assert.NoError(t, result2.Error, "UnscopedDb().Delete()")

		assert.Equal(t, expectedCount, actualCount)
	})
	t.Run("PhotoExceedMaxPhotoCaption", func(t *testing.T) {
		if strings.Contains(entity.DbDialect(), "sqlite") {
			t.Skip("sqlite doesn't support max length testing")
		}
		m := &entity.Photo{}
		n := &entity.Photo{ID: 99887766,
			// UUID:
			PhotoUID:      "1234567890123456789012345678901234567890123456789012345678901234",
			TakenAt:       time.Date(2008, 7, 1, 10, 0, 0, 0, time.UTC),
			TakenAtLocal:  time.Date(2008, 7, 1, 12, 0, 0, 0, time.UTC),
			TakenSrc:      "12345678",
			PhotoType:     "12345678",
			TypeSrc:       "12345678",
			PhotoTitle:    strings.Repeat("1234567890", 20),
			TitleSrc:      "12345678",
			PhotoCaption:  strings.Repeat("1234567890", 409) + "1234567",
			CaptionSrc:    "12345678",
			PhotoPath:     strings.Repeat("1234567890", 102) + "1234",
			PhotoName:     strings.Repeat("1234567890", 25) + "12345",
			OriginalName:  strings.Repeat("1234567890", 75) + "12345",
			PhotoFavorite: false,

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
				assert.Contains(t, result.Error.Error(), schema.NamingStrategy{}.ColumnName("", "PhotoCaption"))
			} else if strings.Contains(entity.DbDialect(), "postgres") {
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
				assert.Contains(t, result.Error.Error(), "value too long")
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
		if strings.Contains(entity.DbDialect(), "postgres") {
			t.Skip("postgres doesn't support max length testing on bytes")
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
		expectedCount++

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

	t.Run("AlbumSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		a := entity.NewAlbum("PhotosPrism", entity.AlbumFolder)
		a.AlbumSlug = txt.Clip(sTxt, 80)
		assert.Len(t, a.AlbumSlug, 80)
		assert.NotEqual(t, sTxt, a.AlbumSlug)
		require.NoError(t, a.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&a).Error)

		a = entity.NewAlbum("PhotosPrism", entity.AlbumFolder)
		a.AlbumSlug = txt.Clip(sTxt, 160)
		assert.Len(t, a.AlbumSlug, 160)
		assert.NotEqual(t, sTxt, a.AlbumSlug)
		require.NoError(t, a.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&a).Error)

		a = entity.NewAlbum("PhotosPrism", entity.AlbumFolder)
		a.AlbumSlug = txt.Clip(sTxt, 161)
		assert.Len(t, a.AlbumSlug, 161)
		assert.NotEqual(t, sTxt, a.AlbumSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, a.Create(), "Data too long")
		default:
			require.NoError(t, a.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&a).Error)
		}
	})

	t.Run("CameraSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)
		c := entity.NewCamera("PhotosPrism", "PhotoPrism Test Model 80")
		c.CameraSlug = txt.Clip(sTxt, 80)
		assert.Len(t, c.CameraSlug, 80)
		assert.NotEqual(t, sTxt, c.CameraSlug)
		require.NoError(t, c.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&c).Error)

		c = entity.NewCamera("PhotosPrism", "PhotoPrism Test Model 160")
		c.CameraSlug = txt.Clip(sTxt, 160)
		assert.Len(t, c.CameraSlug, 160)
		assert.NotEqual(t, sTxt, c.CameraSlug)
		require.NoError(t, c.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&c).Error)

		c = entity.NewCamera("PhotosPrism", "PhotoPrism Test Model 162")
		c.CameraSlug = txt.Clip(sTxt, 162)
		assert.Len(t, c.CameraSlug, 162)
		assert.NotEqual(t, sTxt, c.CameraSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, c.Create(), "Data too long")
		default:
			require.NoError(t, c.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&c).Error)
		}

		// The following string is randomly generated via https://pinkylam.me/generator/chinese-lorem-ipsum/
		s = "息戊玩員方大！飛女節兄抓央更三黑院都亮到，向友服面那故收共穴哪戶國色？六回多丟，每告晚友苦貓常和辛是汁去苦卜「犬丁風門反一支蝴」話怪更這校！蝸隻少七話就直弓「活吉天」氣久背真掃朋拉牙送別飛兄。林午頁坡急蛋奶就。手陽土，長七上步蝶訴石尼文邊鼻抄打時尺聽重道，尤蛋珠。"
		c = entity.NewCamera("PhotosPrism", "PhotoPrism Test Model 162 Unicode")
		c.CameraSlug = txt.Clip(slug.Make(s), 162)
		assert.Len(t, c.CameraSlug, 162)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, c.Create(), "Data too long")
		default:
			require.NoError(t, c.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&c).Error)
		}
	})

	t.Run("CountrySlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		c := entity.NewCountry("TC", "Test Country")
		c.CountrySlug = txt.Clip(sTxt, 80)
		assert.Len(t, c.CountrySlug, 80)
		assert.NotEqual(t, sTxt, c.CountrySlug)
		require.NoError(t, c.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&c).Error)

		c = entity.NewCountry("TC", "Test Country")
		c.CountrySlug = txt.Clip(sTxt, 160)
		assert.Len(t, c.CountrySlug, 160)
		assert.NotEqual(t, sTxt, c.CountrySlug)
		require.NoError(t, c.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&c).Error)

		c = entity.NewCountry("TC", "Test Country")
		c.CountrySlug = txt.Clip(sTxt, 161)
		assert.Len(t, c.CountrySlug, 161)
		assert.NotEqual(t, sTxt, c.CountrySlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, c.Create(), "Data too long")
		default:
			require.NoError(t, c.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&c).Error)
		}
	})

	t.Run("LabelSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		l := entity.NewLabel("TestingLabelLength", 5)
		l.LabelSlug = txt.Clip(sTxt, 80)
		l.CustomSlug = txt.Clip(sTxt, 80)
		assert.Len(t, l.LabelSlug, 80)
		assert.NotEqual(t, sTxt, l.LabelSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLabel("TestingLabelLength", 5)
		l.LabelSlug = txt.Clip(sTxt, 160)
		l.CustomSlug = txt.Clip(sTxt, 160)
		assert.Len(t, l.LabelSlug, 160)
		assert.NotEqual(t, sTxt, l.LabelSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLabel("TestingLabelLength", 5)
		l.LabelSlug = txt.Clip(sTxt, 161)
		l.CustomSlug = txt.Clip(sTxt, 161)
		assert.Len(t, l.LabelSlug, 161)
		assert.NotEqual(t, sTxt, l.LabelSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, l.Create(), "Data too long")
		default:
			require.NoError(t, l.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&l).Error)
		}
	})

	t.Run("LensSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		l := entity.NewLens("TestingLensLength", "Test Lens")
		l.LensSlug = txt.Clip(sTxt, 80)
		assert.Len(t, l.LensSlug, 80)
		assert.NotEqual(t, sTxt, l.LensSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLens("TestingLensLength", "Test Lens")
		l.LensSlug = txt.Clip(sTxt, 160)
		assert.Len(t, l.LensSlug, 160)
		assert.NotEqual(t, sTxt, l.LensSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLens("TestingLensLength", "Test Lens")
		l.LensSlug = txt.Clip(sTxt, 161)
		assert.Len(t, l.LensSlug, 161)
		assert.NotEqual(t, sTxt, l.LensSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, l.Create(), "Data too long")
		default:
			require.NoError(t, l.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&l).Error)
		}
	})

	t.Run("ShareSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		l := entity.NewLink(rnd.GenerateUID(entity.AlbumUID), false, false)
		l.ShareSlug = txt.Clip(sTxt, 80)
		assert.Len(t, l.ShareSlug, 80)
		assert.NotEqual(t, sTxt, l.ShareSlug)
		require.NoError(t, l.Save())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLink(rnd.GenerateUID(entity.AlbumUID), false, false)
		l.ShareSlug = txt.Clip(sTxt, 160)
		assert.Len(t, l.ShareSlug, 160)
		assert.NotEqual(t, sTxt, l.ShareSlug)
		require.NoError(t, l.Save())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewLink(rnd.GenerateUID(entity.AlbumUID), false, false)
		l.ShareSlug = txt.Clip(sTxt, 161)
		assert.Len(t, l.ShareSlug, 161)
		assert.NotEqual(t, sTxt, l.ShareSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, l.Save(), "Data too long")
		default:
			require.NoError(t, l.Save())
			require.NoError(t, entity.UnscopedDb().Delete(&l).Error)
		}
	})

	t.Run("SubjSlug", func(t *testing.T) {
		s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec et purus pulvinar, iaculis enim at, placerat mauris. Phasellus lacus turpis, egestas id cursus condimentum, facilisis malesuada dui cras."
		sTxt := slug.Make(s)

		l := entity.NewSubject("TestingLensLength", entity.SubjPerson, entity.SrcManual)
		l.SubjSlug = txt.Clip(sTxt, 80)
		assert.Len(t, l.SubjSlug, 80)
		assert.NotEqual(t, sTxt, l.SubjSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewSubject("TestingLensLength", entity.SubjPerson, entity.SrcManual)
		l.SubjSlug = txt.Clip(sTxt, 160)
		assert.Len(t, l.SubjSlug, 160)
		assert.NotEqual(t, sTxt, l.SubjSlug)
		require.NoError(t, l.Create())
		require.NoError(t, entity.UnscopedDb().Delete(&l).Error)

		l = entity.NewSubject("TestingLensLength", entity.SubjPerson, entity.SrcManual)
		l.SubjSlug = txt.Clip(sTxt, 161)
		assert.Len(t, l.SubjSlug, 161)
		assert.NotEqual(t, sTxt, l.SubjSlug)
		switch entity.DbDialect() {
		case dsn.DialectMySQL:
			require.ErrorContains(t, l.Create(), "Data too long")
		default:
			require.NoError(t, l.Create())
			require.NoError(t, entity.UnscopedDb().Delete(&l).Error)
		}
	})

	log.Info("End expecting data to long Error or SQLSTATE from dbtest_fieldlength_test")
}
