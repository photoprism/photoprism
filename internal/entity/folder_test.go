package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestNewFolder(t *testing.T) {
	t.Run("Num2020Num05", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "2020/05", time.Now().UTC())
		assert.Equal(t, RootOriginals, folder.Root)
		assert.Equal(t, "2020/05", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, "", folder.FolderDescription)
		assert.Equal(t, "", folder.FolderType)
		assert.Equal(t, sortby.Name, folder.FolderOrder)
		assert.IsType(t, "", folder.FolderUID)
		assert.Equal(t, false, folder.FolderFavorite)
		assert.Equal(t, false, folder.FolderIgnore)
		assert.Equal(t, false, folder.FolderWatch)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("Num2020Num05Num01", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/01/", time.Now().UTC())
		assert.Equal(t, "2020/05/01", folder.Path)
		assert.Equal(t, "May 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("Num2020Num05Num23", func(t *testing.T) {
		folder := NewFolder(RootImport, "/2020/05/23/", time.Now().UTC())
		assert.Equal(t, "2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("NormalizesBackslashes", func(t *testing.T) {
		folder := NewFolder(RootImport, `\2020\05\23\`, time.Now().UTC())
		assert.Equal(t, "2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("Num2020Num05Num23IcelandNum2020", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/2020/05/23/Iceland 2020", time.Now().UTC())
		assert.Equal(t, "2020/05/23/Iceland 2020", folder.Path)
		assert.Equal(t, "Iceland 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "is", folder.FolderCountry)
	})
	t.Run("LondonNum2020Num05Num23", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "/London/2020/05/23", time.Now().UTC())
		assert.Equal(t, "London/2020/05/23", folder.Path)
		assert.Equal(t, "May 23, 2020", folder.FolderTitle)
		assert.Equal(t, 2020, folder.FolderYear)
		assert.Equal(t, 5, folder.FolderMonth)
		assert.Equal(t, "zz", folder.FolderCountry)
	})
	t.Run("RootOriginalsNoDir", func(t *testing.T) {
		folder := NewFolder(RootOriginals, "", time.Time{})
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("RootOriginalsRootDir", func(t *testing.T) {
		folder := NewFolder(RootOriginals, RootPath, time.Time{})
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Originals", folder.FolderTitle)
		assert.Equal(t, 0, folder.FolderYear)
		assert.Equal(t, 0, folder.FolderMonth)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
	t.Run("NoRootWithRootDir", func(t *testing.T) {
		folder := NewFolder("", RootPath, time.Now().UTC())
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "", folder.FolderTitle)
		assert.Equal(t, UnknownID, folder.FolderCountry)
	})
}

func TestFirstOrCreateFolder(t *testing.T) {
	t.Run("ExistingRootFolder", func(t *testing.T) {
		folder := NewFolder(RootOriginals, RootPath, time.Now().UTC())
		result := FirstOrCreateFolder(&folder)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if folder.FolderTitle != "Originals" {
			t.Errorf("FolderTitle should be 'Originals'")
		}

		if folder.FolderCountry != UnknownID {
			t.Errorf("FolderCountry should be 'zz'")
		}

		found := FindFolder(RootOriginals, RootPath)

		if found == nil {
			t.Fatal("found must not be nil")
		}

		if found.FolderTitle != "Originals" {
			t.Errorf("FolderTitle should be 'Originals'")
		}

		if found.FolderCountry != UnknownID {
			t.Errorf("FolderCountry should be 'zz'")
		}
	})
	t.Run("ReturnsSoftDeletedOnCreateConflict", func(t *testing.T) {
		folderPath := "first-or-create-soft-deleted-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		now := Now()
		if err := UnscopedDb().Model(Folder{}).Where("root = ? AND path = ?", RootOriginals, folderPath).UpdateColumn("deleted_at", now).Error; err != nil {
			t.Fatal(err)
		}

		createCandidate := NewFolder(RootOriginals, folderPath, time.Now().UTC())
		result := FirstOrCreateFolder(&createCandidate)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if result.DeletedAt == nil {
			t.Fatal("expected soft-deleted folder from unscoped conflict lookup")
		}
	})
}

func TestFolder_SetValuesFromPath(t *testing.T) {
	t.Run("Root", func(t *testing.T) {
		folder := NewFolder("new", "", time.Now().UTC())
		folder.SetValuesFromPath()
		assert.Equal(t, "New", folder.FolderTitle)
	})
}

func TestFolder_Slug(t *testing.T) {
	t.Run("Root", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach", Root: "sidecar", Path: "ugly/beach"}
		assert.Equal(t, "ugly-beach", folder.Slug())
	})
}

func TestFolder_Title(t *testing.T) {
	t.Run("Root", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach"}
		assert.Equal(t, "Beautiful beach", folder.Title())
	})
}

func TestFolder_RootPath(t *testing.T) {
	t.Run("Rainbow", func(t *testing.T) {
		folder := Folder{FolderTitle: "Beautiful beach", Root: "/", Path: "rainbow"}
		assert.Equal(t, "/rainbow", folder.RootPath())
	})
}

func TestFindFolder(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		assert.Nil(t, FindFolder("vvfgt", "jgfuyf"))
	})
	t.Run("PathNameIsRootPath", func(t *testing.T) {
		assert.Nil(t, FindFolder("vvfgt", RootPath))
	})
	t.Run("FindsSoftDeleted", func(t *testing.T) {
		folderPath := "find-folder-soft-deleted-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		now := Now()

		if err := UnscopedDb().Model(Folder{}).Where("root = ? AND path = ?", RootOriginals, folderPath).UpdateColumn("deleted_at", now).Error; err != nil {
			t.Fatal(err)
		}

		found := FindFolder(RootOriginals, folderPath)

		if found == nil {
			t.Fatal("expected folder lookup result")
		}

		assert.NotNil(t, found.DeletedAt)
	})
	t.Run("NormalizesBackslashes", func(t *testing.T) {
		folderPath := "find-folder-backslash-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		found := FindFolder(RootOriginals, strings.ReplaceAll(folderPath, "/", `\`))

		if found == nil {
			t.Fatal("expected folder lookup result")
		}

		assert.Equal(t, folderPath, found.Path)
	})
}

func TestFolder_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		folder := NewFolder("oldRoot", "oldPath", time.Now().UTC())

		assert.Equal(t, "oldRoot", folder.Root)
		assert.Equal(t, "oldPath", folder.Path)

		err := folder.Updates(Folder{Root: "newRoot", Path: "newPath"})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "newRoot", folder.Root)
		assert.Equal(t, "newPath", folder.Path)
	})
}

func TestFolder_SetForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		formValues := Folder{FolderTitle: "Beautiful beach"}

		folderForm, err := form.NewFolder(formValues)

		if err != nil {
			t.Fatal(err)
		}

		folder := NewFolder("oldRoot", "oldPath", time.Now().UTC())

		assert.Equal(t, "oldRoot", folder.Root)
		assert.Equal(t, "oldPath", folder.Path)
		assert.Equal(t, "OldPath", folder.FolderTitle)

		err = folder.SetForm(folderForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", folder.Root)
		assert.Equal(t, "", folder.Path)
		assert.Equal(t, "Beautiful beach", folder.FolderTitle)
	})
}

func TestFolder_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		folder := Folder{FolderTitle: "Holiday 2020", Root: RootOriginals, Path: "2020/Greece"}
		err := folder.Create()

		if err != nil {
			t.Fatal(err)
		}

		result := FindFolder(RootOriginals, "2020/Greece")

		assert.Equal(t, "2020-greece", result.Slug())
		assert.Equal(t, "Holiday 2020", result.Title())
	})
	t.Run("EmojiSubFolderDoesNotOverwriteParentAlbum", func(t *testing.T) {
		parentPath := "emoji-collision-parent-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		childPath := parentPath + "/🍷"

		parentFolder := NewFolder(RootOriginals, parentPath, time.Now().UTC())

		if err := parentFolder.Create(); err != nil {
			t.Fatal(err)
		}

		childFolder := NewFolder(RootOriginals, childPath, time.Now().UTC())

		if err := childFolder.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path IN (?)", RootOriginals, []string{parentPath, childPath}).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{parentPath, childPath}).Delete(Album{}).Error
		})

		parentAlbum := FindFolderAlbum(parentPath)

		if parentAlbum == nil {
			t.Fatal("expected parent folder album")
		}

		childAlbum := FindFolderAlbum(childPath)

		if childAlbum == nil {
			t.Fatal("expected child folder album")
		}

		assert.NotEqual(t, parentAlbum.ID, childAlbum.ID)
		assert.Equal(t, parentPath, parentAlbum.AlbumPath)
		assert.Equal(t, childPath, childAlbum.AlbumPath)
		assert.Equal(t, parentFolder.Title(), parentAlbum.AlbumTitle)
		assert.Equal(t, childFolder.Title(), childAlbum.AlbumTitle)
	})
	t.Run("EmojiSubFolderRepairsParentCollisionTitle", func(t *testing.T) {
		cases := []struct {
			name      string
			childName string
		}{
			{name: "Wine", childName: "🍷"},
			{name: "Puzzle", childName: "🧩"},
			{name: "BeachUmbrella", childName: "⛱️"},
			{name: "Blossom", childName: "🌸"},
			{name: "WorkKiss", childName: "Work 😘"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				parentPath := "emoji-collision-repair-" + txt.Slug(tc.childName+"-"+time.Now().UTC().Format(time.RFC3339Nano))
				childPath := parentPath + "/" + tc.childName

				staleAlbum := NewFolderAlbum(txt.Title(parentPath), childPath, `path:"`+childPath+`" public:true`)

				if staleAlbum == nil {
					t.Fatal("expected stale folder album")
				}

				if err := staleAlbum.Create(); err != nil {
					t.Fatal(err)
				}

				childFolder := NewFolder(RootOriginals, childPath, time.Now().UTC())

				if err := childFolder.Create(); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, childPath).Delete(Folder{}).Error
					_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, childPath).Delete(Album{}).Error
				})

				childAlbum := FindFolderAlbum(childPath)

				if childAlbum == nil {
					t.Fatal("expected child folder album")
				}

				assert.Equal(t, childFolder.Title(), childAlbum.AlbumTitle)
				assert.NotEqual(t, txt.Title(parentPath), childAlbum.AlbumTitle)
			})
		}
	})
	t.Run("ExistingFolderDuplicateCreateDoesNotReconcileAlbum", func(t *testing.T) {
		folderPath := "existing-folder-recreate-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano)) + "/child"
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		album := FindFolderAlbum(folderPath)

		if album == nil {
			t.Fatal("expected folder album")
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}

		if found := FindFolderAlbum(folderPath); found != nil {
			t.Fatal("expected folder album to be deleted")
		}

		rescanFolder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := rescanFolder.Create(); err == nil {
			t.Fatal("expected duplicate folder create error")
		}

		if recreated := FindFolderAlbum(folderPath); recreated != nil {
			t.Fatal("expected duplicate folder create to skip reconciliation")
		}
	})
	t.Run("ExistingFolderRecreatesMissingAlbumOnRescanRepair", func(t *testing.T) {
		folderPath := "existing-folder-recreate-repair-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano)) + "/child"
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		album := FindFolderAlbum(folderPath)

		if album == nil {
			t.Fatal("expected folder album")
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}

		if found := FindFolderAlbum(folderPath); found != nil {
			t.Fatal("expected folder album to be deleted")
		}

		reconciled, err := ReconcileOriginalsFolderAlbums(folderPath)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, reconciled)

		recreated := FindFolderAlbum(folderPath)

		if recreated == nil {
			t.Fatal("expected folder album to be recreated")
		}

		assert.Equal(t, folderPath, recreated.AlbumPath)
		assert.Equal(t, folder.Title(), recreated.AlbumTitle)
	})
	t.Run("ExistingFolderRepairsAlbumTitleOnRescanRepair", func(t *testing.T) {
		parentPath := "existing-folder-repair-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		folderPath := parentPath + "/child-folder"
		folder := NewFolder(RootOriginals, folderPath, time.Now().UTC())

		if err := folder.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path = ?", RootOriginals, folderPath).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path = ?", AlbumFolder, folderPath).Delete(Album{}).Error
		})

		album := FindFolderAlbum(folderPath)

		if album == nil {
			t.Fatal("expected folder album")
		}

		// Simulate stale collision state where the child album title shows the
		// parent folder name.
		parentTitle := txt.Title(parentPath)

		if err := album.Update("AlbumTitle", parentTitle); err != nil {
			t.Fatal(err)
		}

		afterUpdate := FindFolderAlbum(folderPath)

		if afterUpdate == nil {
			t.Fatal("expected folder album after title update")
		}

		assert.Equal(t, parentTitle, afterUpdate.AlbumTitle)

		reconciled, err := ReconcileOriginalsFolderAlbums(parentPath)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, reconciled)

		repaired := FindFolderAlbum(folderPath)

		if repaired == nil {
			t.Fatal("expected repaired folder album")
		}

		assert.Equal(t, folder.Title(), repaired.AlbumTitle)
		assert.NotEqual(t, parentTitle, repaired.AlbumTitle)
	})
	t.Run("ExistingEmojiSiblingFoldersRepairSwappedTitleOnRescanRepair", func(t *testing.T) {
		parentPath := "ins-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		pathMirror := parentPath + "/🪞"
		pathWine := parentPath + "/🍷"

		folderMirror := NewFolder(RootOriginals, pathMirror, time.Now().UTC())
		folderWine := NewFolder(RootOriginals, pathWine, time.Now().UTC())

		if err := folderMirror.Create(); err != nil {
			t.Fatal(err)
		}

		if err := folderWine.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path IN (?)", RootOriginals, []string{pathMirror, pathWine}).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathWine}).Delete(Album{}).Error
		})

		// Start from a stale collision state:
		// - only one album row exists for pathMirror
		// - title/filter point to pathWine
		_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathWine}).Delete(Album{}).Error

		stale := &Album{
			AlbumType:   AlbumFolder,
			AlbumSlug:   legacyFolderAlbumSlug(pathMirror),
			AlbumPath:   pathMirror,
			AlbumFilter: `path:"` + pathWine + `" public:true`,
			CreatedAt:   Now(),
			UpdatedAt:   Now(),
		}
		stale.SetTitle(folderWine.Title())

		if err := stale.Create(); err != nil {
			t.Fatal(err)
		}

		reconciled, err := ReconcileOriginalsFolderAlbums(parentPath)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, reconciled)

		mirrorAlbum := FindFolderAlbum(pathMirror)

		if mirrorAlbum == nil {
			t.Fatal("expected mirror folder album")
		}

		wineAlbum := FindFolderAlbum(pathWine)

		if wineAlbum == nil {
			t.Fatal("expected wine folder album")
		}

		assert.Equal(t, folderMirror.Title(), mirrorAlbum.AlbumTitle)
		assert.Equal(t, folderWine.Title(), wineAlbum.AlbumTitle)
		assert.Equal(t, pathMirror, mirrorAlbum.AlbumPath)
		assert.Equal(t, pathWine, wineAlbum.AlbumPath)
	})
	t.Run("SingleFolderRescanRepairsSiblingCollisionScope", func(t *testing.T) {
		parentPath := "ins-single-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		pathMirror := parentPath + "/🪞"
		pathWine := parentPath + "/🍷"

		folderMirror := NewFolder(RootOriginals, pathMirror, time.Now().UTC())
		folderWine := NewFolder(RootOriginals, pathWine, time.Now().UTC())

		if err := folderMirror.Create(); err != nil {
			t.Fatal(err)
		}

		if err := folderWine.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path IN (?)", RootOriginals, []string{pathMirror, pathWine}).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathWine}).Delete(Album{}).Error
		})

		// Start from a stale collision state where only one sibling album row
		// exists and points to the wrong folder metadata.
		_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathWine}).Delete(Album{}).Error

		stale := &Album{
			AlbumType:   AlbumFolder,
			AlbumSlug:   legacyFolderAlbumSlug(pathMirror),
			AlbumPath:   pathMirror,
			AlbumFilter: `path:"` + pathWine + `" public:true`,
			CreatedAt:   Now(),
			UpdatedAt:   Now(),
		}
		stale.SetTitle(folderWine.Title())

		if err := stale.Create(); err != nil {
			t.Fatal(err)
		}

		// Reconcile a single missing sibling path. This should widen scope to
		// the parent folder so both siblings get repaired together.
		reconciled, err := ReconcileOriginalsFolderAlbums(pathWine)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, reconciled)

		mirrorAlbum := FindFolderAlbum(pathMirror)

		if mirrorAlbum == nil {
			t.Fatal("expected mirror folder album")
		}

		wineAlbum := FindFolderAlbum(pathWine)

		if wineAlbum == nil {
			t.Fatal("expected wine folder album")
		}

		assert.Equal(t, folderMirror.Title(), mirrorAlbum.AlbumTitle)
		assert.Equal(t, folderWine.Title(), wineAlbum.AlbumTitle)
		assert.Equal(t, pathMirror, mirrorAlbum.AlbumPath)
		assert.Equal(t, pathWine, wineAlbum.AlbumPath)
	})
	t.Run("SingleDeepFolderRescanRepairsAncestorCollisionScope", func(t *testing.T) {
		parentPath := "ins-deep-" + txt.Slug(time.Now().UTC().Format(time.RFC3339Nano))
		pathMirror := parentPath + "/🍷"
		pathEmoji := parentPath + "/🪞"
		pathPumpkin := pathEmoji + "/🎃"
		pathLink := pathPumpkin + "/🔗"
		pathFish := pathLink + "/🐠"

		folderMirror := NewFolder(RootOriginals, pathMirror, time.Now().UTC())
		folderEmoji := NewFolder(RootOriginals, pathEmoji, time.Now().UTC())
		folderPumpkin := NewFolder(RootOriginals, pathPumpkin, time.Now().UTC())
		folderLink := NewFolder(RootOriginals, pathLink, time.Now().UTC())
		folderFish := NewFolder(RootOriginals, pathFish, time.Now().UTC())

		for _, folder := range []*Folder{&folderMirror, &folderEmoji, &folderPumpkin, &folderLink, &folderFish} {
			if err := folder.Create(); err != nil {
				t.Fatal(err)
			}
		}

		t.Cleanup(func() {
			_ = UnscopedDb().Where("root = ? AND path IN (?)", RootOriginals, []string{pathMirror, pathEmoji, pathPumpkin, pathLink, pathFish}).Delete(Folder{}).Error
			_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathEmoji, pathPumpkin, pathLink, pathFish}).Delete(Album{}).Error
		})

		// Start from a stale collision state where a deep emoji-only folder has
		// overwritten a sibling album outside its immediate parent subtree.
		_ = UnscopedDb().Where("album_type = ? AND album_path IN (?)", AlbumFolder, []string{pathMirror, pathFish}).Delete(Album{}).Error

		stale := &Album{
			AlbumType:   AlbumFolder,
			AlbumSlug:   legacyFolderAlbumSlug(pathMirror),
			AlbumPath:   pathMirror,
			AlbumFilter: `path:"` + pathFish + `" public:true`,
			CreatedAt:   Now(),
			UpdatedAt:   Now(),
		}
		stale.SetTitle(folderFish.Title())

		if err := stale.Create(); err != nil {
			t.Fatal(err)
		}

		reconciled, err := ReconcileOriginalsFolderAlbums(pathFish)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, reconciled)

		mirrorAlbum := FindFolderAlbum(pathMirror)

		if mirrorAlbum == nil {
			t.Fatal("expected mirror folder album")
		}

		fishAlbum := FindFolderAlbum(pathFish)

		if fishAlbum == nil {
			t.Fatal("expected fish folder album")
		}

		assert.Equal(t, folderMirror.Title(), mirrorAlbum.AlbumTitle)
		assert.Equal(t, folderFish.Title(), fishAlbum.AlbumTitle)
		assert.Equal(t, pathMirror, mirrorAlbum.AlbumPath)
		assert.Equal(t, pathFish, fishAlbum.AlbumPath)
	})
}
