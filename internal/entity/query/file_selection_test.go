package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/require"
)

// aclSession builds an in-memory session for the named user fixture.
func aclSession(name string) *entity.Session {
	s := &entity.Session{}
	s.SetUser(entity.UserFixtures.Pointer(name))
	return s
}

// visitorSessionWithShares builds an unregistered visitor session that has redeemed the given share
// link tokens, so its SharedUIDs resolve from the matching links exactly as in production.
func visitorSessionWithShares(tokens ...string) *entity.Session {
	s := &entity.Session{}
	s.SetData(&entity.SessionData{Tokens: tokens})
	return s
}

func TestAlbumDownloadSelection(t *testing.T) {
	t.Run("ExcludesArchivedAndHidden", func(t *testing.T) {
		sel := AlbumDownloadSelection(true, true, false, true)
		assert.False(t, sel.Archived)
		assert.False(t, sel.Hidden)
		// Private is deferred to the session scope when a session is identified.
		assert.True(t, sel.Private)
	})
	t.Run("AnonymousExcludesPrivate", func(t *testing.T) {
		sel := AlbumDownloadSelection(true, true, false, false)
		assert.False(t, sel.Private)
		assert.False(t, sel.Archived)
		assert.False(t, sel.Hidden)
	})
}

func TestSelectedFilesForSession(t *testing.T) {
	// Include private pictures in the base selection so the session scope is what filters them.
	o := FileSelection{Private: true, MaxSize: 1024 * MiB}
	private := form.Selection{Photos: []string{"ps6sg6be2lvl0y13"}} // "Photo06", private

	t.Run("AdminSeesPrivate", func(t *testing.T) {
		files, err := SelectedFilesForSession(private, o, aclSession("alice"))
		assert.NoError(t, err)
		assert.NotEmpty(t, files)
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		files, err := SelectedFilesForSession(private, o, aclSession("guest"))
		assert.NoError(t, err)
		assert.Empty(t, files)
	})
	t.Run("NilMatchesSelectedFiles", func(t *testing.T) {
		frm := form.Selection{Photos: []string{"ps6sg6be2lvl0yh7"}}
		base, err := SelectedFiles(frm, o)
		assert.NoError(t, err)
		scoped, err := SelectedFilesForSession(frm, o, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(base), len(scoped))
	})
	t.Run("VisitorSharedFolderSelection", func(t *testing.T) {
		// A visitor selecting pictures shared only through a folder (smart) album must be able to
		// download them, even though the album has no photos_albums rows.
		//nolint:gosec // G101: deterministic fixture share-link token for tests only.
		const folderShareToken = "8jxf3jfn2k"                       // link to the "april-1990" folder album
		frm := form.Selection{Photos: []string{"ps6sg6be2lvl0yh0"}} // "Photo03", path 1990/04
		files, err := SelectedFilesForSession(frm, o, visitorSessionWithShares(folderShareToken))
		assert.NoError(t, err)
		assert.NotEmpty(t, files)
	})
	t.Run("VisitorWrongShareSelection", func(t *testing.T) {
		// Sharing a different smart album must not expose the folder picture for download.
		//nolint:gosec // G101: deterministic fixture share-link token for tests only.
		const stateShareToken = "9jxf3jfn2k" // link to the "california-usa" state album (excludes 1990/04)
		frm := form.Selection{Photos: []string{"ps6sg6be2lvl0yh0"}}
		files, err := SelectedFilesForSession(frm, o, visitorSessionWithShares(stateShareToken))
		assert.NoError(t, err)
		assert.Empty(t, files)
	})
}

func TestFileSelection(t *testing.T) {
	none := form.Selection{Photos: []string{}}

	one := form.Selection{Photos: []string{"ps6sg6be2lvl0yh8"}}

	two := form.Selection{Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"}}

	albums := form.Selection{Albums: []string{"as6sg6bxpogaaba9", "as6sg6bitoga0004", "as6sg6bxpogaaba8", "as6sg6bxpogaaba7"}}

	months := form.Selection{Albums: []string{"as6sg6bipogaabj9"}}

	folders := form.Selection{Albums: []string{"as6sg6bipogaaba1", "as6sg6bipogaabj8"}}

	states := form.Selection{Albums: []string{"as6sg6bipogaab11", "as6sg6bipotaab12", "asjv2cw2eikl3cb3"}}

	many := form.Selection{
		Files:  []string{"fs6sg6bw45bnlqdw"},
		Photos: []string{"ps6sg6be2lvl0y21", "ps6sg6be2lvl0y19", "ps6sg6byk7wrbk38", "ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"},
	}

	t.Run("EmptySelection", func(t *testing.T) {
		sel := DownloadSelection(true, false, true)
		if results, err := SelectedFiles(none, sel); err == nil {
			t.Fatal("error expected")
		} else {
			assert.Empty(t, results)
		}
	})
	t.Run("DownloadSelectionRawSidecarPrivate", func(t *testing.T) {
		sel := DownloadSelection(true, true, false)
		if results, err := SelectedFiles(one, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 2)
		}
	})
	t.Run("DownloadSelectionRawOriginals", func(t *testing.T) {
		sel := DownloadSelection(true, false, true)
		if results, err := SelectedFiles(two, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 2)
		}
	})
	t.Run("AlbumDownloadSidecar", func(t *testing.T) {
		// Regression for the album download Sidecar option: an album selection must include
		// real sidecar files (e.g. XMP) when Sidecar is enabled and exclude them when not,
		// matching the multi-file download path.
		album := form.Selection{Albums: []string{"as6sg6bxpogaaba9"}} // contains Photo01
		const xmpUID = "fs6sg6bw45bn0003"                             // Photo01.xmp sidecar

		with, err := SelectedFiles(album, DownloadSelection(true, true, false))
		if err != nil {
			t.Fatal(err)
		}
		var foundWith bool
		for _, f := range with {
			if f.FileUID == xmpUID {
				foundWith = true
			}
		}
		assert.True(t, foundWith, "album selection must include the XMP sidecar when enabled")

		without, err := SelectedFiles(album, DownloadSelection(true, false, false))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range without {
			assert.NotEqual(t, xmpUID, f.FileUID, "album selection must exclude the XMP sidecar when disabled")
		}
	})
	t.Run("ShareSelectionOriginals", func(t *testing.T) {
		sel := ShareSelection(false, true)
		if results, err := SelectedFiles(many, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareSelectionPrimary", func(t *testing.T) {
		sel := ShareSelection(true, true)
		if results, err := SelectedFiles(many, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 6)
		}
	})
	t.Run("ShareAlbums", func(t *testing.T) {
		sel := ShareSelection(true, true)
		if results, err := SelectedFiles(albums, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 10)
		}
	})
	t.Run("ShareMonths", func(t *testing.T) {
		sel := ShareSelection(true, true)
		if results, err := SelectedFiles(months, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 0)
		}
	})
	t.Run("ShareFoldersOriginals", func(t *testing.T) {
		sel := ShareSelection(true, true)
		if results, err := SelectedFiles(folders, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareFolders", func(t *testing.T) {
		sel := ShareSelection(false, true)
		if results, err := SelectedFiles(folders, sel); err != nil {
			t.Fatal(err)
		} else {
			log.Debugf("ShareFolders Results: %#v", results)
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareStatesOriginals", func(t *testing.T) {
		sel := ShareSelection(true, true)
		if results, err := SelectedFiles(states, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 5)
		}
	})
	t.Run("ShareStates", func(t *testing.T) {
		sel := ShareSelection(false, true)
		if results, err := SelectedFiles(states, sel); err != nil {
			t.Fatal(err)
		} else {
			log.Debugf("ShareStates Result: %#v", results[0])
			assert.Len(t, results, 5)
		}
	})
}

// TestShareSelection_OmitTypes verifies that sharing converted files omits every image format
// except JPEG, so a newly supported type does not reach a share as its original by default.
func TestShareSelection_OmitTypes(t *testing.T) {
	// Named one by one rather than read back from the format table the selection is built from, so
	// a format that stops being an image is caught here instead of quietly leaving the omit list.
	pinned := []fs.Type{
		fs.ImagePng,
		fs.ImageWebp,
		fs.ImageTiff,
		fs.ImageAvif,
		fs.ImageHeic,
		fs.ImageBmp,
		fs.ImageGif,
		fs.ImagePsd,
		fs.ImageJpegXL,
		fs.ImageCineon,
	}

	t.Run("Converted", func(t *testing.T) {
		omit := ShareSelection(false, true).OmitTypes

		for _, fileType := range pinned {
			assert.Containsf(t, omit, fileType.String(), "%s must not be shared as the original", fileType)
		}

		assert.NotContains(t, omit, fs.ImageJpeg.String(), "jpeg is the format that is shared")
	})
	t.Run("CoversEveryImageType", func(t *testing.T) {
		omit := ShareSelection(false, true).OmitTypes

		for _, fileType := range media.FileTypes(media.Image) {
			if fileType == fs.ImageJpeg {
				continue
			}

			assert.Containsf(t, omit, fileType.String(), "%s must not be shared as the original", fileType)
		}
	})
	t.Run("Originals", func(t *testing.T) {
		assert.Empty(t, ShareSelection(true, true).OmitTypes)
	})
	t.Run("OriginalsWithoutYaml", func(t *testing.T) {
		assert.Equal(t, []string{fs.SidecarYaml.String()}, ShareSelection(true, false).OmitTypes)
	})
	t.Run("ConvertedWithoutYaml", func(t *testing.T) {
		sel := ShareSelection(false, false)
		assert.Contains(t, sel.OmitMedia, media.Sidecar.String())
		assert.Equal(t, ShareSelection(false, true).OmitTypes, sel.OmitTypes)
	})
}

// TestShareSelection_Yaml checks that sharing originals includes YAML sidecar files only when enabled.
func TestShareSelection_Yaml(t *testing.T) {
	photo := entity.NewPhoto(false)
	if err := photo.Save(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})
	for i, name := range []string{"share-control.jpg", "share-control.yml"} {
		fileType, mediaType := fs.ImageJpeg, media.Image
		if fs.FileType(name) == fs.SidecarYaml {
			fileType, mediaType = fs.SidecarYaml, media.Sidecar
		}
		file := entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileName: name, FileType: fileType.String(), MediaType: mediaType.String(),
			FileRoot: entity.RootOriginals, FileHash: fmt.Sprintf("%040d", i+4200)}
		if err := file.Create(); err != nil {
			t.Fatal(err)
		}
	}
	selection := form.Selection{Photos: []string{photo.PhotoUID}}
	names := func(t *testing.T, yaml bool) (result []string) {
		files, err := SelectedFiles(selection, ShareSelection(true, yaml))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			result = append(result, f.FileName)
		}
		return result
	}
	assert.ElementsMatch(t, []string{"share-control.jpg", "share-control.yml"}, names(t, true))
	assert.ElementsMatch(t, []string{"share-control.jpg"}, names(t, false))
}

// TestSelectedFilesForSessionYaml checks export eligibility without changing internal selections.
func TestSelectedFilesForSessionYaml(t *testing.T) {
	photo := entity.NewPhoto(false)
	if err := photo.Save(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})
	file := entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileName: "selection-control.yml", FileType: "yml", FileRoot: entity.RootOriginals, FileHash: "ce1a3d09c704ab1c7519c0edee4a835f661c3aa1"}
	if err := file.Create(); err != nil {
		t.Fatal(err)
	}
	selection := form.Selection{Photos: []string{photo.PhotoUID}}
	options := DownloadSelection(true, true, true)
	internal, err := SelectedFiles(selection, options)
	assert.NoError(t, err)
	assert.Len(t, internal, 1)
	nonDownload := options
	nonDownload.Download = false
	unchanged, err := SelectedFilesForSession(selection, nonDownload, nil)
	assert.NoError(t, err)
	assert.Len(t, unchanged, 1)
	anonymous, err := SelectedFilesForSession(selection, options, nil)
	assert.NoError(t, err)
	assert.Empty(t, anonymous)
	reader, err := SelectedFilesForSession(selection, options, aclSession("alice"))
	assert.NoError(t, err)
	assert.Len(t, reader, 1)
	_, err = SelectedFilesForSession(form.Selection{}, options, aclSession("alice"))
	assert.Error(t, err)
}

func TestSelectedFiles_SubfolderContainment(t *testing.T) {
	base := "zz-like-" + rnd.Base36(6)
	folder := likeTestFolder(t, base+"_a!b!%")
	likeTestFolder(t, base+"_a!b!%/sub")
	likeTestFolder(t, base+"Xa!b!Y/sub")
	likeTestFolder(t, base+"_a!b!Z/sub")
	likeTestFolder(t, base+"_A!b!%/sub")
	inFolder := likeTestPhoto(t, base+"_a!b!%", "in-folder")
	inSubfolder := likeTestPhoto(t, base+"_a!b!%/sub", "in-subfolder")
	sibling := likeTestPhoto(t, base+"Xa!b!Y/sub", "sibling")
	likeTestPhoto(t, base+"_a!b!Z/sub", "sibling-z")
	likeTestPhoto(t, base+"_A!b!%/sub", "sibling-case")

	files, err := SelectedFiles(form.Selection{Files: []string{folder.FolderUID}}, DownloadSelection(true, true, true))
	require.NoError(t, err)

	var uids []string

	for _, f := range files {
		uids = append(uids, f.PhotoUID)
	}

	assert.ElementsMatch(t, []string{inFolder.PhotoUID, inSubfolder.PhotoUID}, uids)
	assert.NotContains(t, uids, sibling.PhotoUID)
}
