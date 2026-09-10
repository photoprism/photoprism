package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFilesByPath(t *testing.T) {
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
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
	entity.ValidateFixtures(t)
	file := entity.FileFixtures.Get("bridge3.jpg")
	var actual entity.File
	require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", file.ID).First(&actual).Error)
	assert.Equal(t, "", actual.FileError)

	SetFileError(file.FileUID, "errorFromTest")
	t.Cleanup(func() {
		require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", file.ID).UpdateColumns(entity.Values{"file_error": "", "photo_taken_at": file.PhotoTakenAt, "time_index": file.TimeIndex, "media_id": file.MediaID}).Error)
	})
	require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", file.ID).First(&actual).Error)
	assert.Equal(t, "errorFromTest", actual.FileError)
}

func TestRenameFile(t *testing.T) {
	entity.ValidateFixtures(t)
	t.Run("EmptyName", func(t *testing.T) {
		err := RenameFile("xxx", "", "yyy", "yyy")

		if err == nil {
			t.Fatal(err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		var actual entity.File
		require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", entity.FileFixturesExampleXMP.ID).First(&actual).Error)
		assert.Equal(t, "2790/02/Photo01.xmp", actual.FileName)
		assert.Equal(t, "/", actual.FileRoot)
		err := RenameFile(entity.FileFixturesExampleXMP.FileRoot, entity.FileFixturesExampleXMP.FileName, "test-root", "yyy.jpg")

		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", entity.FileFixturesExampleXMP.ID).UpdateColumns(entity.Values{"time_index": entity.FileFixturesExampleXMP.TimeIndex, "media_id": entity.FileFixturesExampleXMP.MediaID, "file_name": entity.FileFixturesExampleXMP.FileName, "file_root": entity.FileFixturesExampleXMP.FileRoot}).Error)
		})
		require.NoError(t, Db().Model(&entity.File{}).Where("id = ?", entity.FileFixturesExampleXMP.ID).First(&actual).Error)
		assert.Equal(t, "yyy.jpg", actual.FileName)
		assert.Equal(t, "test-root", actual.FileRoot)
	})

}

func TestIndexedFiles(t *testing.T) {
	entity.ValidateFixtures(t)
	if err := entity.AddDuplicate(
		"Photo18.jpg",
		entity.RootSidecar,
		"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		661858,
		time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
	); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, UnscopedDb().Delete(&entity.Duplicate{}, "1=1").Error)
	})

	result, err := IndexedFiles()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("indexed files: %#v", result)
}

func TestOrphanFiles(t *testing.T) {
	entity.ValidateFixtures(t)
	files, err := OrphanFiles()

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Files{}, files)
}
