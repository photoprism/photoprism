package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/face"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestFirstFileByHash(t *testing.T) {
	t.Run("not existing file", func(t *testing.T) {
		f, err := FirstFileByHash("xxx")

		assert.EqualError(t, err, "record not found")
		assert.Equal(t, uint(0), f.ID)
	})
	t.Run("existing file", func(t *testing.T) {
		f, err := FirstFileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint(0xf4240), f.ID)
	})
}

func TestFile_ShareFileName(t *testing.T) {
	t.Run("photo with title", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc"}

		filename := file.ShareBase(0)

		assert.Contains(t, filename, "20190115-000000-Berlin-Morning-Mood")
		assert.Contains(t, filename, fs.JpegExt)
	})
	t.Run("photo without title", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: ""}
		file := &File{Photo: photo, FileType: "jpg", PhotoUID: "123", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc"}

		filename := file.ShareBase(0)

		assert.Equal(t, filename, "e98eb86480a72bd585d228a709f0622f90e86cbc.jpg")
	})
	t.Run("photo without photo", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc"}

		filename := file.ShareBase(0)

		assert.Equal(t, "e98eb86480a72bd585d228a709f0622f90e86cbc.jpg", filename)
	})
	t.Run("file without photo", func(t *testing.T) {
		file := FileFixtures.Get("FileWithoutPhoto.mp4")

		filename := file.ShareBase(0)

		assert.Equal(t, "pcad9a68fa6acc5c5ba965adf6ec465ca42fd916.mp4", filename)
	})
	t.Run("file hash < 8", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}

		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98"}

		filename := file.ShareBase(0)

		assert.NotContains(t, filename, "20190115-000000-Berlin-Morning-Mood")
	})
	t.Run("no file uid", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileHash: "e98ijhyt"}

		filename := file.ShareBase(0)

		assert.Equal(t, filename, "e98ijhyt.jpg")
	})
}

func TestFile_Changed(t *testing.T) {
	var deletedAt = time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
	t.Run("different modified times", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix()}
		d := time.Date(2020, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, true, file.Changed(500, d))
	})
	t.Run("different sizes", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 600, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix()}
		d := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, true, file.Changed(500, d))
	})
	t.Run("no change", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix()}
		d := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, false, file.Changed(500, d))
	})
	t.Run("deleted", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix(), DeletedAt: &deletedAt}
		d := time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, false, file.Changed(500, d))
	})
}

func TestFile_Missing(t *testing.T) {
	var deletedAt = time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)

	t.Run("deleted", func(t *testing.T) {
		file := &File{FileMissing: false, Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix(), DeletedAt: &deletedAt}
		assert.Equal(t, true, file.Missing())
	})

	t.Run("missing", func(t *testing.T) {
		file := &File{FileMissing: true, Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix(), DeletedAt: nil}
		assert.Equal(t, true, file.Missing())
	})

	t.Run("not_missing", func(t *testing.T) {
		file := &File{FileMissing: false, Photo: nil, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix(), DeletedAt: nil}
		assert.Equal(t, false, file.Missing())
	})
}

func TestFile_Create(t *testing.T) {
	t.Run("photo id == 0", func(t *testing.T) {
		file := File{PhotoID: 0}

		assert.Error(t, file.Create())
	})
	t.Run("file already exists", func(t *testing.T) {
		file := &File{PhotoID: 123, FileType: "jpg", FileSize: 500, ModTime: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC).Unix()}
		assert.Nil(t, file.Create())
		assert.Error(t, file.Create())
	})
	t.Run("success", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}

		file := &File{Photo: photo, FileType: "jpg", FileSize: 500, PhotoID: 766, FileName: "testname", FileRoot: "xyz"}

		err := file.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFile_Purge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.Equal(t, nil, file.Purge())
	})
}

func TestFile_Found(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.Equal(t, nil, file.Purge())
		assert.Equal(t, true, file.FileMissing)
		err := file.Found()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, file.FileMissing)
	})
}

func TestFile_AllFilesMissing(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		file := FileFixtures.Get("missing.jpg")
		assert.True(t, file.AllFilesMissing())
	})
	t.Run("false", func(t *testing.T) {
		file := FileFixtures.Get("Quality1FavoriteTrue.jpg")
		assert.False(t, file.AllFilesMissing())
	})
}

func TestFile_Save(t *testing.T) {
	t.Run("save without photo", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", PhotoUID: "123", FileUID: "123"}
		err := file.Save()

		if err == nil {
			t.Fatal("error should not be nil")
		}

		if file.ID != 0 {
			t.Fatalf("file id should be 0: %d", file.ID)
		}

		assert.Equal(t, "file: photo id must not be empty (save 123)", err.Error())
	})
	t.Run("success", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}

		file := &File{Photo: photo, FileType: "jpg", FileSize: 500, PhotoID: 766, FileName: "Food", FileRoot: "", UpdatedAt: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC)}

		err := file.Save()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFile_UpdateVideoInfos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{FileType: "jpg", FileWidth: 600, FileName: "VideoUpdate", PhotoID: 1000003}

		assert.Equal(t, "1990/04/bridge2.mp4", FileFixturesExampleBridgeVideo.FileName)
		assert.Equal(t, int(1200), FileFixturesExampleBridgeVideo.FileWidth)

		err := file.UpdateVideoInfos()

		if err != nil {
			t.Fatal(err)
		}

		var files Files

		if err := Db().Where("photo_id = ? AND file_video = 1", file.PhotoID).Find(&files).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, files, 1)

		for _, f := range files {
			assert.Equal(t, "1990/04/bridge2.mp4", f.FileName)
			assert.Equal(t, int(600), f.FileWidth)
		}
	})
}

func TestFile_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{FileType: "jpg", FileSize: 500, FileName: "ToBeUpdated", FileRoot: "", PhotoID: 5678}

		err := file.Save()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "ToBeUpdated", file.FileName)

		err2 := file.Update("FileName", "Happy")

		if err2 != nil {
			t.Fatal(err2)
		}
		assert.Equal(t, "Happy", file.FileName)
	})
}

func TestFile_Links(t *testing.T) {
	t.Run("result", func(t *testing.T) {
		file := FileFixturesExampleBridge
		links := file.Links()
		assert.Equal(t, "5jxf3jfn2k", links[0].LinkToken)
	})
}

func TestFile_NoJPEG(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "xmp", FileSize: 500}
		assert.True(t, file.NoJPEG())
	})
	t.Run("false", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.False(t, file.NoJPEG())
	})
}

func TestFile_Panorama(t *testing.T) {
	t.Run("3000", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 3000, FileHeight: 1000}
		assert.True(t, file.Panorama())
	})
	t.Run("1999", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1999, FileHeight: 1000}
		assert.False(t, file.Panorama())
	})
	t.Run("2000", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 2000, FileHeight: 1000}
		assert.True(t, file.Panorama())
	})
	t.Run("false", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1500, FileHeight: 1000}
		assert.False(t, file.Panorama())
	})
	t.Run("equirectangular", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1500, FileHeight: 1000, FileProjection: ProjectionEquirectangular}
		assert.True(t, file.Panorama())
	})
	t.Run("sidecar", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "xmp", FileSidecar: true, FileWidth: 3000, FileHeight: 1000}
		assert.False(t, file.Panorama())
	})
}

func TestFile_Delete(t *testing.T) {
	t.Run("permanently", func(t *testing.T) {
		file := &File{FileType: "jpg", FileSize: 500, FileName: "ToBePermanentlyDeleted", FileRoot: "", PhotoID: 5678}

		err := file.Save()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "ToBePermanentlyDeleted", file.FileName)

		err2 := file.Delete(true)

		assert.Nil(t, err2)
	})
	t.Run("not permanently", func(t *testing.T) {
		file := &File{FileType: "jpg", FileSize: 500, FileName: "ToBeDeleted", FileRoot: "", PhotoID: 5678}

		err := file.Save()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "ToBeDeleted", file.FileName)

		err2 := file.Delete(false)

		assert.Nil(t, err2)
	})
}

func TestPrimaryFile(t *testing.T) {
	file, err := PrimaryFile("pt9jtdre2lvl0y17")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Holiday/Video.jpg", file.FileName)
}

func TestFile_OriginalBase(t *testing.T) {
	t.Run("original name empty, filename empty", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "", FileName: ""}

		filename := file.OriginalBase(0)

		assert.Contains(t, filename, "20190115-000000-Berlin-Morning-Mood")
		assert.Contains(t, filename, fs.JpegExt)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "20190115-000000-Berlin-Morning-Mood")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.JpegExt)
	})
	t.Run("original name empty", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "", FileName: "sonnenaufgang.jpg"}

		filename := file.OriginalBase(0)

		assert.Contains(t, filename, "sonnenaufgang")
		assert.Contains(t, filename, fs.JpegExt)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "sonnenaufgang")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.JpegExt)
	})
	t.Run("original name not empty", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "Sonnenaufgang.jpg", FileName: "123.jpg"}

		filename := file.OriginalBase(0)

		assert.Contains(t, filename, "Sonnenaufgang")
		assert.Contains(t, filename, fs.JpegExt)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "Sonnenaufgang")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.JpegExt)
	})
}

func TestFile_DownloadName(t *testing.T) {
	t.Run("DownloadNameFile", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "originalName.jpg", FileName: "filename.jpg"}

		filename := file.DownloadName(DownloadNameFile, 0)
		assert.Contains(t, filename, "filename")
		assert.Contains(t, filename, fs.JpegExt)

		filename2 := file.DownloadName(DownloadNameOriginal, 1)
		assert.Contains(t, filename2, "originalName")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.JpegExt)

		filename3 := file.DownloadName("xxx", 0)
		assert.Contains(t, filename3, "20190115-000000-Berlin-Morning-Mood")
	})
}

func TestFile_Undelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.Equal(t, nil, file.Purge())
		assert.Equal(t, true, file.FileMissing)
		err := file.Undelete()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, file.FileMissing)
	})
	t.Run("file not missing", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSize: 500}
		assert.Equal(t, false, file.FileMissing)
		err := file.Undelete()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, file.FileMissing)
	})
}

func TestFile_AddFaces(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		file := &File{FileType: "jpg", FileWidth: 720, FileName: "FacesTest", PhotoID: 1000003}

		faces := face.Faces{face.Face{
			Rows:      480,
			Cols:      720,
			Score:     45,
			Area:      face.NewArea("face", 250, 200, 10),
			Eyes:      face.Areas{face.NewArea("eye_l", 240, 195, 1), face.NewArea("eye_r", 240, 205, 1)},
			Landmarks: face.Areas{face.NewArea("a", 250, 185, 2), face.NewArea("b", 250, 215, 2)},
		}}

		file.AddFaces(faces)

		if err := file.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, file.FileMissing)
		assert.NotEmpty(t, file.FileUID)
		assert.NotEmpty(t, file.Markers())
	})
}

func TestFile_FaceCount(t *testing.T) {
	t.Run("FileFixturesExampleBridge", func(t *testing.T) {
		file := FileFixturesExampleBridge

		result := file.FaceCount()

		assert.GreaterOrEqual(t, result, 3)
	})
}

func TestFile_Rename(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := FileFixtures.Get("exampleFileName.jpg")

		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", m.FileName)
		assert.Equal(t, RootOriginals, m.FileRoot)
		assert.Equal(t, false, m.FileMissing)
		assert.Nil(t, m.DeletedAt)

		p := m.RelatedPhoto()

		assert.Equal(t, "2790/07", p.PhotoPath)
		assert.Equal(t, "27900704_070228_D6D51B6C", p.PhotoName)

		if err := m.Rename("x/y/newName.jpg", "newRoot", "x/y", "newBase"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "x/y/newName.jpg", m.FileName)
		assert.Equal(t, "newRoot", m.FileRoot)
		assert.Equal(t, false, m.FileMissing)
		assert.Nil(t, m.DeletedAt)
		assert.Equal(t, "x/y", p.PhotoPath)
		assert.Equal(t, "newBase", p.PhotoName)

		if err := m.Rename("2790/07/27900704_070228_D6D51B6C.jpg", RootOriginals, "2790/07", "27900704_070228_D6D51B6C"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", m.FileName)
		assert.Equal(t, RootOriginals, m.FileRoot)
		assert.Equal(t, false, m.FileMissing)
		assert.Nil(t, m.DeletedAt)
		assert.Equal(t, "2790/07", p.PhotoPath)
		assert.Equal(t, "27900704_070228_D6D51B6C", p.PhotoName)
	})
}

func TestFile_SubjectNames(t *testing.T) {
	t.Run("Photo27.jpg", func(t *testing.T) {
		m := FileFixtures.Get("Photo27.jpg")

		names := m.SubjectNames()
		t.Log(len(names))
		if len(names) != 1 {
			t.Errorf("there should be one name: %#v", names)
		} else {
			assert.Equal(t, "Actress A", names[0])
		}
	})
	t.Run("Video.jpg", func(t *testing.T) {
		m := FileFixtures.Get("Video.jpg")

		names := m.SubjectNames()
		t.Log(len(names))
		if len(names) != 1 {
			t.Errorf("there should be one name: %#v", names)
		} else {
			assert.Equal(t, "Actor A", names[0])
		}
	})
	t.Run("bridge.jpg", func(t *testing.T) {
		m := FileFixtures.Get("bridge.jpg")

		names := m.SubjectNames()

		if len(names) != 2 {
			t.Errorf("two names expected: %#v", names)
		} else {
			assert.Equal(t, []string{"Corn McCornface", "Jens Mander"}, names)
		}
	})
}
