package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestAlbumHasThumb(t *testing.T) {
	// Other tests assign covers through UpdateCovers(), so the fixture value is set explicitly
	// and restored afterwards. The direct update bypasses the hooks that clear the album cache.
	setAlbumThumb := func(t *testing.T, uid, hash string) {
		var current []string

		if err := Db().Model(&entity.Album{}).Where("album_uid = ?", uid).Limit(1).Pluck("thumb", &current).Error; err != nil {
			t.Fatal(err)
		} else if err = Db().Model(&entity.Album{}).Where("album_uid = ?", uid).Update("thumb", hash).Error; err != nil {
			t.Fatal(err)
		}

		entity.FlushAlbumCache()

		t.Cleanup(func() {
			restore := ""

			if len(current) > 0 {
				restore = current[0]
			}

			_ = Db().Model(entity.Album{}).Where("album_uid = ?", uid).Update("thumb", restore).Error
			entity.FlushAlbumCache()
		})
	}

	t.Run("NoThumb", func(t *testing.T) {
		setAlbumThumb(t, "as6sg6bxpogaaba7", "")
		assert.False(t, AlbumHasThumb("as6sg6bxpogaaba7"))
	})
	t.Run("HasThumb", func(t *testing.T) {
		setAlbumThumb(t, "as6sg6bxpogaaba7", "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.True(t, AlbumHasThumb("as6sg6bxpogaaba7"))
	})
	t.Run("StaleThumb", func(t *testing.T) {
		// A hash no client can resolve must not gate the cover query.
		setAlbumThumb(t, "as6sg6bxpogaaba7", "0000000000000000000000000000000000000000")
		assert.False(t, AlbumHasThumb("as6sg6bxpogaaba7"))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.False(t, AlbumHasThumb("as6sg6bxpog00007"))
	})
	t.Run("InvalidUID", func(t *testing.T) {
		assert.False(t, AlbumHasThumb("3765"))
	})
}

func TestAlbumByUID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if album, err := AlbumByUID("as6sg6bxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Christmas 2030", album.AlbumTitle)
		}

		if cached, err := AlbumByUID("as6sg6bxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Christmas 2030", cached.AlbumTitle)
		}
	})
	t.Run("Missing", func(t *testing.T) {
		album, err := AlbumByUID("as6sg6bxpog00007")
		assert.NotNil(t, album)
		assert.Error(t, err, "record not found")
	})
	t.Run("InvalidUID", func(t *testing.T) {
		album, err := AlbumByUID("3765")
		assert.NotNil(t, album)
		assert.Error(t, err, "invalid album uid")
	})
}

func TestAlbumCoverByUID(t *testing.T) {
	t.Run("ExistingUidDefaultAlbum", func(t *testing.T) {
		file, err := AlbumCoverByUID("as6sg6bxpogaaba8", true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2023/11/IMG_57.jpg", file.FileName)
	})
	t.Run("ExistingUidFolderAlbum", func(t *testing.T) {
		file, err := AlbumCoverByUID("as6sg6bipogaaba1", true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04/bridge2.jpg", file.FileName)
	})
	t.Run("ExistingUidEmptyMomentAlbum", func(t *testing.T) {
		file, err := AlbumCoverByUID("as6sg6bitoga0005", true)

		assert.EqualError(t, err, "no cover found", err)
		assert.Equal(t, "", file.FileName)
	})
	t.Run("NotExistingUid", func(t *testing.T) {
		file, err := AlbumCoverByUID("3765", true)
		assert.Error(t, err, "record not found")
		t.Log(file)
	})
	t.Run("ExistingUidEmptyMonthAlbum", func(t *testing.T) {
		file, err := AlbumCoverByUID("as6sg6bipogaabj9", true)

		assert.EqualError(t, err, "no cover found", err)
		assert.Equal(t, "", file.FileName)
	})
}

func TestUpdateAlbumDates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 1, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 17, album.AlbumDay)
		actual, err = UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})
	t.Run("ZeroDay", func(t *testing.T) {
		// A stored day of 0 must be repaired even when it would resolve to the newest
		// picture date if the components were normalized rather than rejected.
		album := entity.Album{
			AlbumUID: "as6sg6bxpogaabz1", AlbumType: entity.AlbumFolder,
			AlbumTitle: "Zero Day", AlbumSlug: "zero-day-album", AlbumPath: "1990/03",
			AlbumYear: 1990, AlbumMonth: 4, AlbumDay: 0,
		}
		require.NoError(t, entity.UnscopedDb().Create(&album).Error)
		photo := entity.Photo{
			PhotoUID:     "ps6sg6bxpogaabz1",
			PhotoName:    "zerodayalbum",
			PhotoPath:    "1990/03",
			TakenAt:      time.Date(1990, 3, 31, 9, 0, 0, 0, time.UTC),
			TakenAtLocal: time.Date(1990, 3, 31, 9, 0, 0, 0, time.UTC),
			TakenSrc:     entity.SrcMeta,
			PhotoQuality: 3,
		}
		require.NoError(t, entity.UnscopedDb().Create(&photo).Error)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM albums WHERE album_uid = ?", "as6sg6bxpogaabz1").Error)
			require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM photos WHERE photo_uid = ?", "ps6sg6bxpogaabz1").Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		_, err := UpdateAlbumDates()
		require.NoError(t, err)
		actual := entity.FindAlbum(entity.Album{AlbumUID: "as6sg6bxpogaabz1"})
		assert.Equal(t, 1990, actual.AlbumYear)
		assert.Equal(t, 3, actual.AlbumMonth)
		assert.Equal(t, 31, actual.AlbumDay)
	})
	t.Run("MaxWithTimeOffset", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		photoPhoto17 := entity.PhotoFixtures.Get("Photo17")
		// Set a timestamp that would sort (by string) BEFORE 1990-04-18 01:00:00+08:00 in SQLite, but is actually a GREATER date.
		photoPhoto17.TakenAtLocal = time.Date(1990, 4, 18, 1, 0, 0, 0, time.UTC)
		photoPhoto17.TakenSrc = entity.SrcMeta
		photoPhoto17.PhotoQuality = 4

		entity.Db().Save(&photoPhoto17)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.PhotoFixtures.Pointer("Photo17")).Error)
		}()

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 1, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 18, album.AlbumDay)
		actual, err = UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})
	t.Run("TwoAlbums", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		album.AlbumDay = 0
		album.AlbumMonth = 0
		album.AlbumYear = 0
		require.NoError(t, entity.Db().Save(&album).Error)
		album2 := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("2016-04").AlbumUID})
		assert.Equal(t, 0, album2.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("2016-04")).Error)
		}()
		photoPhoto12 := entity.PhotoFixtures.Get("Photo12")
		photoPhoto12.TakenAtLocal = time.Date(2016, 4, 18, 1, 0, 0, 0, time.UTC)
		photoPhoto12.TakenSrc = entity.SrcMeta
		photoPhoto12.PhotoQuality = 4
		photoPhoto12.PhotoPath = "2016/04"

		entity.Db().Save(&photoPhoto12)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.PhotoFixtures.Pointer("Photo12")).Error)
		}()

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 2, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 17, album.AlbumDay)
		album2 = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("2016-04").AlbumUID})
		assert.Equal(t, 18, album2.AlbumDay)
		actual, err = UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})
	t.Run("BlankAlbumPath", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		album.AlbumPath = ""
		album.AlbumDay = 0
		album.AlbumMonth = 0
		album.AlbumYear = 0
		require.NoError(t, entity.Db().Save(&album).Error)
		photoPhoto17 := entity.PhotoFixtures.Get("Photo17")
		// Set a timestamp that would sort (by string) BEFORE 1990-04-18 01:00:00+08:00 in SQLite, but is actually a GREATER date.
		photoPhoto17.TakenAtLocal = time.Date(1990, 4, 18, 1, 0, 0, 0, time.UTC)
		photoPhoto17.TakenSrc = entity.SrcMeta
		photoPhoto17.PhotoQuality = 4
		photoPhoto17.PhotoPath = ""

		entity.Db().Save(&photoPhoto17)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.PhotoFixtures.Pointer("Photo17")).Error)
		}()

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 0, album.AlbumDay)
		assert.Equal(t, 0, album.AlbumMonth)
		assert.Equal(t, 0, album.AlbumYear)
		actual, err = UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})
	t.Run("DeleteRecord", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		photoPhoto17 := entity.PhotoFixtures.Get("Photo17")
		// Set a timestamp that would sort (by string) BEFORE 1990-04-18 01:00:00+08:00 in SQLite, but is actually a GREATER date.
		photoPhoto17.TakenAtLocal = time.Date(1990, 4, 18, 1, 0, 0, 0, time.UTC)
		photoPhoto17.TakenSrc = entity.SrcMeta
		photoPhoto17.PhotoQuality = 4

		entity.Db().Save(&photoPhoto17)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.PhotoFixtures.Pointer("Photo17")).Error)
		}()

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 1, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 18, album.AlbumDay)
		require.NoError(t, photoPhoto17.Archive())
		actual, err = UpdateAlbumDates()
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 17, album.AlbumDay)
		require.NoError(t, err)
		assert.Equal(t, 1, actual)
	})
	t.Run("InvalidDateFromAlbum", func(t *testing.T) {
		album := entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 11, album.AlbumDay)
		defer func() {
			require.NoError(t, entity.Db().Save(entity.AlbumFixtures.Pointer("april-1990")).Error)
		}()
		album.AlbumDay = 0
		album.AlbumMonth = 0
		album.AlbumYear = 0
		require.NoError(t, album.Save())

		actual, err := UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 1, actual)
		album = entity.FindAlbum(entity.Album{AlbumUID: entity.AlbumFixtures.Get("april-1990").AlbumUID})
		assert.Equal(t, 17, album.AlbumDay)
		actual, err = UpdateAlbumDates()
		require.NoError(t, err)
		assert.Equal(t, 0, actual)
	})
}

func TestUpdateMissingAlbumEntries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := UpdateMissingAlbumEntries(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbumEntryFound(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := AlbumEntryFound("ps6sg6bexxvl0yh0"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbums(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		results, err := Albums(0, 3)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, results, 3)
	})
}

func TestAlbumsByUID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		results, err := AlbumsByUID([]string{"as6sg6bxpogaaba7", "as6sg6bxpogaaba8"}, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, results, 2)
	})
	t.Run("IncludeDeleted", func(t *testing.T) {
		results, err := AlbumsByUID([]string{"as6sg6bxpogaaba7", "as6sg6bxpogaaba8"}, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, results, 2)
	})
}

func TestAlbumsByType(t *testing.T) {
	t.Run("AlbumFolder", func(t *testing.T) {
		actual, err := AlbumsByType(entity.AlbumFolder, false)
		require.NoError(t, err)
		assert.Len(t, actual, 4)
	})
	t.Run("AlbumManual", func(t *testing.T) {
		actual, err := AlbumsByType(entity.AlbumManual, true)
		require.NoError(t, err)
		assert.Len(t, actual, 22)

		m := entity.AlbumFixtures.Pointer("christmas2030")
		require.NoError(t, m.Delete())
		defer func() {
			require.NoError(t, m.Restore())
		}()
		actual, err = AlbumsByType(entity.AlbumManual, false)
		require.NoError(t, err)
		assert.Len(t, actual, 21)
	})
}
