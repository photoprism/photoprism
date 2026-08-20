package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFolderCoverByUID(t *testing.T) {
	t.Run("Num1990Num04", func(t *testing.T) {
		if result, err := FolderCoverByUID("dqo63pn2f87f02xj"); err != nil {
			t.Fatal(err)
		} else if result.FileUID == "" {
			t.Fatal("result must not be empty")
		} else if result.FileUID != "fs6sg6bw15bnlqdw" {
			t.Errorf("wrong result: %#v", result)
		}
	})
	t.Run("Num2007Twelve", func(t *testing.T) {
		if result, err := FolderCoverByUID("dqo63pn2f87f02oi"); err != nil {
			t.Fatal(err)
		} else if result.FileUID == "" {
			t.Fatal("result must not be empty")
		} else if result.FileUID != "fs6sg6bqhhinlplk" {
			t.Errorf("wrong result: %#v", result)
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		for _, uid := range []string{"", "xxx", "as6sg6bxpogaaba8"} {
			_, err := FolderCoverByUID(uid)
			assert.EqualError(t, err, "invalid folder uid", "%s", uid)
		}
	})
}

func TestFoldersByPath(t *testing.T) {
	before, err := FoldersByRoot(entity.RootOriginals, true)
	require.NoError(t, err)
	defer func() {
		after, err := FoldersByRoot(entity.RootOriginals, true)
		require.NoError(t, err)
		for _, afterFolder := range after {
			found := false
			for _, beforeFolder := range before {
				if beforeFolder.FolderUID == afterFolder.FolderUID {
					found = true
					break
				}
			}
			if !found {
				require.NoError(t, entity.UnscopedDb().Delete(&afterFolder).Error)
			}
		}
	}()
	t.Run("Root", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootOriginals, "testdata", "", false)

		t.Logf("folders: %+v", folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, folders, 1)
	})
	t.Run("Subdirectory", func(t *testing.T) {
		folders, err := FoldersByPath(entity.RootOriginals, "testdata", "directory", false)

		t.Logf("folders: %+v", folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, folders, 2)
	})
}

func TestAlbumFolders(t *testing.T) {
	t.Run("Root", func(t *testing.T) {
		folders, err := AlbumFolders(1)

		if assert.Nil(t, err) {
			assert.GreaterOrEqual(t, len(folders), 1)

			t.Logf("folders: %+v", folders)
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		folders, err := AlbumFolders(9999999)

		if assert.Nil(t, err) {
			assert.NotNil(t, folders)
			assert.Len(t, folders, 0)
		}
	})
}

func TestUpdateFolderDates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		actual := entity.FindFolder("/", "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("1990/04")).Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("2007/12")).Error)
		}()
		records, err := UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 2, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 17, actual.FolderDay)
		records, err = UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 0, records)
	})
	t.Run("MaxWithTimeOffset", func(t *testing.T) {
		actual := entity.FindFolder("/", "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("1990/04")).Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("2007/12")).Error)
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

		records, err := UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 2, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 18, actual.FolderDay)
		records, err = UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 0, records)
	})
	t.Run("DeleteRecord", func(t *testing.T) {
		actual := entity.FindFolder("/", "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("1990/04")).Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("2007/12")).Error)
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

		records, err := UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 2, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 18, actual.FolderDay)
		require.NoError(t, photoPhoto17.Archive())
		records, err = UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 1, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 17, actual.FolderDay)
	})
	t.Run("InvalidDateFromFolder", func(t *testing.T) {
		actual := entity.FindFolder("/", "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		actual.FolderDay = 0
		actual.FolderMonth = 0
		actual.FolderYear = 0
		require.NoError(t, entity.UnscopedDb().Save(actual).Error)

		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("1990/04")).Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("2007/12")).Error)
		}()
		records, err := UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 2, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 17, actual.FolderDay)
		records, err = UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 0, records)
	})
	t.Run("FolderOnAnotherRoot", func(t *testing.T) {
		actual := entity.FindFolder("/", "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		actual.FolderDay = 0
		actual.FolderMonth = 0
		actual.FolderYear = 0
		require.NoError(t, entity.UnscopedDb().Save(actual).Error)

		importF := entity.NewFolder(entity.RootImport, "1990/04", time.Date(1990, 4, 1, 0, 0, 0, 0, time.UTC))
		importF.FolderDay = 0
		importF.FolderMonth = 0
		importF.FolderYear = 0
		require.NoError(t, entity.Db().Save(&importF).Error)
		assert.Equal(t, 0, importF.FolderDay)

		defer func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("1990/04")).Error)
			require.NoError(t, entity.UnscopedDb().Save(entity.FolderFixtures.Pointer("2007/12")).Error)
			require.NoError(t, entity.UnscopedDb().Delete(&importF).Error)
		}()

		records, err := UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 2, records)
		actual = entity.FindFolder("/", "1990/04")
		assert.Equal(t, 17, actual.FolderDay)
		actual = entity.FindFolder(entity.RootImport, "1990/04")
		assert.Equal(t, 0, actual.FolderDay)
		records, err = UpdateFolderDates()
		require.NoError(t, err)
		assert.Equal(t, 0, records)
	})
}

func TestFoldersByRoot(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		actual, err := FoldersByRoot(entity.RootOriginals, true)
		require.NoError(t, err)
		assert.Len(t, actual, 4, "Include Deleted")
		actual, err = FoldersByRoot(entity.RootOriginals, false)
		require.NoError(t, err)
		assert.Len(t, actual, 3, "Exclude Deleted")
	})
}
