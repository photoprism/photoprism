package query

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFilesByPath(t *testing.T) {
	t.Run("Holiday", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "Holiday", false)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
	t.Run("SlashHoliday", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "/Holiday", false)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
	t.Run("Num2790Public", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "2790/07", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
	t.Run("HolidayPublic", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "Holiday", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
}

func TestExistingFiles(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		files, err := Files(1000, 0, "/", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(files))
	})
	t.Run("FilesFoundIncludeMissingFalse", func(t *testing.T) {
		files, err := Files(1000, 0, "/", false)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(files))
	})
	t.Run("SearchForFilesPath", func(t *testing.T) {
		files, err := Files(1000, 0, "Photos", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, files)
	})
}

func TestFilesByUID(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		files, err := FilesByUID([]string{"fs6sg6bw45bnlqdw"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", files[0].FileName)
	})
	t.Run("NoFilesFound", func(t *testing.T) {
		files, err := FilesByUID([]string{"fs6sg6bw45bnlxxx"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(files))
	})
	t.Run("InvalidLimit", func(t *testing.T) {
		_, err := FilesByUID([]string{"fs6sg6bw45bnlxxx"}, -100, 0)
		assert.Error(t, err)
	})

	t.Run("Negative limit with offset", func(t *testing.T) {
		_, err := FilesByUID([]string{"fs6sg6bw45bnlqdw"}, -100, 100)
		assert.Error(t, err)
	})

	t.Run("offset and limit", func(t *testing.T) {
		files, err := FilesByUID([]string{"fs6sg6bw45bnlxxx"}, 10, 100)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(files))
	})
}

func TestFileByPhotoUID(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		file, err := FileByPhotoUID("ps6sg6be2lvl0y11")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Germany/bridge.jpg", file.FileName)
	})
	t.Run("NoFilesFound", func(t *testing.T) {
		file, err := FileByPhotoUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestVideoByPhotoUID(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		file, err := VideoByPhotoUID("ps6sg6be2lvl0yh0")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "1990/04/bridge2.mp4", file.FileName)
	})
	t.Run("NoFilesFound", func(t *testing.T) {
		file, err := VideoByPhotoUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestDocumentByPhotoUID(t *testing.T) {
	t.Run("FileFound", func(t *testing.T) {
		file, err := DocumentByPhotoUID("ps6sg6byk7wrbk48")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "pdf", file.FileType)
		assert.Equal(t, "education/university/BSc-Thesis.pdf", file.FileName)
	})
	t.Run("NoFileFound", func(t *testing.T) {
		_, err := DocumentByPhotoUID("111")

		assert.Error(t, err)
	})
	t.Run("EmptyUID", func(t *testing.T) {
		_, err := DocumentByPhotoUID("")

		assert.Error(t, err)
	})
}

func TestFileByUID(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		file, err := FileByUID("fs6sg6bw45bnlqdw")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", file.FileName)
	})
	t.Run("NoFilesFound", func(t *testing.T) {
		file, err := FileByUID("111")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByHash(t *testing.T) {
	t.Run("FilesFound", func(t *testing.T) {
		file, err := FileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", file.FileName)
	})
	t.Run("NoFilesFound", func(t *testing.T) {
		file, err := FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestSetPhotoPrimary(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, false, entity.FileFixturesExampleXMP.FilePrimary)

		err := SetPhotoPrimary("ps6sg6be2lvl0yh7", "fs6sg6bwhhbnlqdn")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NoFileUid", func(t *testing.T) {
		err := SetPhotoPrimary("ps6sg6be2lvl0yh7", "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NoUid", func(t *testing.T) {
		err := SetPhotoPrimary("", "")

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("FileMissing", func(t *testing.T) {
		err := SetPhotoPrimary("ps6sg6be2lvl0y22", "")

		if err == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err.Error(), "cannot find primary file")
	})
}

func TestSetFileError(t *testing.T) {
	assert.Equal(t, "", entity.FileFixturesExampleXMP.FileError)

	SetFileError("fs6sg6bwhhbnlqdn", "errorFromTest")

	// TODO How to assert
	// assert.Equal(t, true, entity.FileFixturesExampleXMP.FilePrimary)
}

func TestRenameFile(t *testing.T) {
	t.Run("EmptyName", func(t *testing.T) {
		err := RenameFile("xxx", "", "yyy", "yyy")

		if err == nil {
			t.Fatal(err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "2790/02/Photo01.xmp", entity.FileFixturesExampleXMP.FileName)
		assert.Equal(t, "/", entity.FileFixturesExampleXMP.FileRoot)
		err := RenameFile("/", "exampleXmpFile.xmp", "test-root", "yyy.jpg")

		if err != nil {
			t.Fatal(err)
		}
		// TODO how to assert?
		// assert.Equal(t, "", entity.FileFixturesExampleXMP.FileName)
	})

}

func TestIndexedFiles(t *testing.T) {
	if err := entity.AddDuplicate(
		"Photo18.jpg",
		entity.RootSidecar,
		"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		661858,
		time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
	); err != nil {
		t.Fatal(err)
	}

	result, err := IndexedFiles()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("indexed files: %#v", result)
}

func TestOrphanFiles(t *testing.T) {
	files, err := OrphanFiles()

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Files{}, files)
}

// TestFilesByPath_OmitsMarkers pins that the folder listing carries no markers. Its response is
// cached across sessions under a key that names none of them, so a marker list resolved for
// whoever asked first must not be what the next session reads.
func TestFilesByPath_OmitsMarkers(t *testing.T) {
	files, err := FilesByPath(100, 0, entity.RootOriginals, "1990/04", false)
	require.NoError(t, err)
	require.NotEmpty(t, files, "the fixture has to hold files for this to mean anything")

	for i := range files {
		assert.True(t, files[i].OmitMarkers, files[i].FileUID)
		assert.Empty(t, *files[i].Markers())
	}

	// The key stays, carrying an empty list: a non-nil pointer is never omitted, and an absent
	// field would be a wider response change than this needs to be.
	b, err := json.Marshal(files)
	require.NoError(t, err)

	var res []struct {
		Markers []map[string]any
	}

	require.NoError(t, json.Unmarshal(b, &res))

	for i := range res {
		assert.Empty(t, res[i].Markers)
	}
}

func TestFilesByPhotoIDs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("19800101_000002_D640C559")

		files, err := FilesByPhotoIDs([]uint{photo.ID})

		require.NoError(t, err)
		require.NotEmpty(t, files)

		for _, f := range files {
			assert.Equal(t, photo.ID, f.PhotoID)
		}
	})
	t.Run("Duplicates", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("19800101_000002_D640C559")

		once, err := FilesByPhotoIDs([]uint{photo.ID})
		require.NoError(t, err)

		twice, err := FilesByPhotoIDs([]uint{photo.ID, 0, photo.ID})
		require.NoError(t, err)

		assert.Len(t, twice, len(once))
		assert.NotEmpty(t, twice[0].FileName)
	})
	t.Run("None", func(t *testing.T) {
		files, err := FilesByPhotoIDs(nil)

		require.NoError(t, err)
		assert.Empty(t, files)
	})
}

// TestOriginalsByPhotoID verifies that only the originals of a picture are returned, with all columns.
func TestOriginalsByPhotoID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := entity.PhotoFixtures.Get("19800101_000002_D640C559")

		files, err := OriginalsByPhotoID(photo.ID)

		require.NoError(t, err)
		require.NotEmpty(t, files)

		for _, f := range files {
			assert.Equal(t, photo.ID, f.PhotoID)
			assert.Equal(t, entity.RootOriginals, f.FileRoot)
			assert.False(t, f.FileSidecar)
			assert.False(t, f.FileMissing)
			assert.NotEmpty(t, f.FileUID)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		files, err := OriginalsByPhotoID(0)

		require.NoError(t, err)
		assert.Empty(t, files)
	})
}
