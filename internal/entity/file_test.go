package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/projection"
)

func TestFile_RegenerateIndex(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		File{ID: 1000000}.RegenerateIndex()
	})
	t.Run("PhotoID", func(t *testing.T) {
		File{PhotoID: 1000039}.RegenerateIndex()
	})
	t.Run("PhotoUID", func(t *testing.T) {
		File{PhotoUID: "pr2xu7myk7wrbk32"}.RegenerateIndex()
	})
	t.Run("FirstFileByHash", func(t *testing.T) {
		f, err := FirstFileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		if err != nil {
			t.Fatal(err)
		}
		f.RegenerateIndex()
	})
	t.Run("All", func(t *testing.T) {
		File{}.RegenerateIndex()
	})
}

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
		assert.Contains(t, filename, fs.ExtJPEG)
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

		assert.Equal(t, "file 123: cannot save file with empty photo id", err.Error())
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
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1910, FileHeight: 1000}
		assert.True(t, file.Panorama())
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
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1500, FileHeight: 1000, FileProjection: projection.Equirectangular.String()}
		assert.True(t, file.Panorama())
	})
	t.Run("transverse-cylindrical", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "jpg", FileSidecar: false, FileWidth: 1500, FileHeight: 1000, FileProjection: projection.TransverseCylindrical.String()}
		assert.True(t, file.Panorama())
	})
	t.Run("sidecar", func(t *testing.T) {
		file := &File{Photo: nil, FileType: "xmp", FileSidecar: true, FileWidth: 3000, FileHeight: 1000}
		assert.False(t, file.Panorama())
	})
}

func TestFile_SetProjection(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		m := &File{}
		m.SetProjection(Unknown)
		assert.True(t, projection.Unknown.Equal(m.FileProjection))
		assert.Equal(t, Unknown, m.FileProjection)
		assert.Equal(t, projection.Unknown.String(), m.FileProjection)
	})
	t.Run(projection.Cubestrip.String(), func(t *testing.T) {
		m := &File{}
		m.SetProjection(projection.Cubestrip.String())
		assert.True(t, projection.Cubestrip.Equal(m.FileProjection))
		assert.Equal(t, projection.Cubestrip.String(), m.FileProjection)
	})
	t.Run(projection.Cylindrical.String(), func(t *testing.T) {
		m := &File{}
		m.SetProjection(projection.Cylindrical.String())
		assert.True(t, projection.Cylindrical.Equal(m.FileProjection))
		assert.Equal(t, projection.Cylindrical.String(), m.FileProjection)
	})
	t.Run(projection.TransverseCylindrical.String(), func(t *testing.T) {
		m := &File{}
		m.SetProjection(projection.TransverseCylindrical.String())
		assert.Equal(t, projection.TransverseCylindrical.String(), m.FileProjection)
	})
	t.Run(projection.PseudocylindricalCompromise.String(), func(t *testing.T) {
		m := &File{}
		m.SetProjection(projection.PseudocylindricalCompromise.String())
		assert.Equal(t, projection.PseudocylindricalCompromise.String(), m.FileProjection)
		assert.Equal(t, projection.PseudocylindricalCompromise, projection.Find(m.FileProjection))
	})
	t.Run("New", func(t *testing.T) {
		m := &File{}
		p := projection.New(" 幸福 Hanzi are logograms developed for the writing of Chinese! Expressions in an index may not ...!")
		m.SetProjection(p.String())
		assert.Equal(t, p.String(), m.FileProjection)
		assert.GreaterOrEqual(t, clean.ClipType, len(m.FileProjection))
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
		assert.Contains(t, filename, fs.ExtJPEG)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "20190115-000000-Berlin-Morning-Mood")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.ExtJPEG)
	})
	t.Run("original name empty", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "", FileName: "sonnenaufgang.jpg"}

		filename := file.OriginalBase(0)

		assert.Contains(t, filename, "sonnenaufgang")
		assert.Contains(t, filename, fs.ExtJPEG)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "sonnenaufgang")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.ExtJPEG)
	})
	t.Run("original name not empty", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "Sonnenaufgang.jpg", FileName: "123.jpg"}

		filename := file.OriginalBase(0)

		assert.Contains(t, filename, "Sonnenaufgang")
		assert.Contains(t, filename, fs.ExtJPEG)

		filename2 := file.OriginalBase(1)
		assert.Contains(t, filename2, "Sonnenaufgang")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.ExtJPEG)
	})
}

func TestFile_DownloadName(t *testing.T) {
	t.Run("DownloadNameFile", func(t *testing.T) {
		photo := &Photo{TakenAtLocal: time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), PhotoTitle: "Berlin / Morning Mood"}
		file := &File{Photo: photo, FileType: "jpg", FileUID: "foobar345678765", FileHash: "e98eb86480a72bd585d228a709f0622f90e86cbc", OriginalName: "originalName.jpg", FileName: "filename.jpg"}

		filename := file.DownloadName(customize.DownloadNameFile, 0)
		assert.Contains(t, filename, "filename")
		assert.Contains(t, filename, fs.ExtJPEG)

		filename2 := file.DownloadName(customize.DownloadNameOriginal, 1)
		assert.Contains(t, filename2, "originalName")
		assert.Contains(t, filename2, "(1)")
		assert.Contains(t, filename2, fs.ExtJPEG)

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
	t.Run("Primary", func(t *testing.T) {
		file := &File{FileUID: "fqzuh65p4sjk3kdn", FileHash: "346b3897eec9ef75e35fbf0bbc4c83c55ca41e31", FileType: "jpg", FileWidth: 720, FileName: "FacesTest", PhotoID: 1000003, FilePrimary: true}

		faces := face.Faces{face.Face{
			Rows:       480,
			Cols:       720,
			Score:      45,
			Area:       face.NewArea("face", 250, 200, 10),
			Eyes:       face.Areas{face.NewArea("eye_l", 240, 195, 1), face.NewArea("eye_r", 240, 205, 1)},
			Landmarks:  face.Areas{face.NewArea("a", 250, 185, 2), face.NewArea("b", 250, 215, 2)},
			Embeddings: face.Embeddings{{0.012816238, 0.0054710666, 0.06963101, 0.037285835, 0.04412884, 0.017333193, 0.03373656, -0.033069234, 0.025952332, -0.0035901065, -0.029420156, 0.07464688, -0.043232113, 0.060328174, 0.028897963, -0.027495274, -0.02622295, 0.038605634, -0.030962847, 0.05343173, -0.05042871, 0.010407827, 0.014773584, 0.04305641, -0.045918103, 0.014705811, 0.0031296816, 0.08703609, 0.012646829, 0.040463835, 0.080548696, -0.04496776, 0.032542497, -0.046235666, 0.0018886769, -0.09422433, -0.006701393, 0.0601084, 0.05649471, -0.02277308, 0.048038833, -0.022022927, -0.024692882, -0.0067683067, 0.02597589, -0.026766079, -0.04489042, -0.060946267, 0.052194964, 0.0098239435, -0.063547, 0.008626338, -0.041202333, -0.057555206, -0.05206756, -0.007974572, -0.036597952, -0.04232167, 0.0064586936, 0.011131428, -0.076106876, -0.014716604, 0.027977718, 0.060634963, 0.0046368516, -0.035929997, -0.079733424, -0.051017676, -0.03521493, -0.0062531913, -0.030387852, 0.101194955, -0.027980363, -0.010152243, -0.005128962, -0.026926627, 0.008371125, -0.088778615, 0.022396773, -0.025815062, -0.0027552384, -0.049987435, -0.019902563, -0.024667386, 0.064883195, -0.010091326, -0.024541432, -0.03390568, -0.04975766, -0.05255319, 0.0462333, -0.062871166, 0.070803925, -0.020970127, 0.012365979, -0.048543453, 0.027297763, 0.02785581, 0.09220687, -0.021206442, 0.015040259, 0.11726589, 0.00079200073, 0.08544253, 0.08694455, -0.037786104, -0.09956117, 0.07473473, 0.086737245, 0.02916126, 0.0355523, 0.067868374, -0.056218974, 0.007066174, 0.046310645, -0.025015457, -0.019863142, -0.018884404, 0.00076502684, -0.0699868, 0.043558553, 0.11221989, -0.036503807, -0.07346668, 0.023614183, 0.008353507, 0.05629068, -0.05628395, -0.030611087, 0.013364313, -0.014508443, 0.013493559, 0.061809033, 0.06598724, -0.03538405, -0.08597677, -0.06253287, -0.032587055, 0.030790405, 0.031729434, 0.0349981, 0.09145327, 0.012044479, 0.09593962, -0.011460096, -0.014851449, -0.041916795, 0.0037967677, -0.028313408, -0.016944066, -0.023236271, 0.046519253, 0.09026307, -0.014203754, 0.0048228586, 0.012194195, -0.062746234, -0.02189861, -0.030368697, 0.004226377, -0.044146035, 0.04542304, 0.046805177, 0.03882082, -0.06006401, 0.06286592, 0.03714168, -0.011287339, -0.0129849315, 0.01757729, -0.031555075, 0.005606887, 0.0045785876, 0.031963747, 0.040269036, 0.033833507, -0.06477002, -0.0039275866, 0.079373375, 0.044617712, 0.012070597, -0.06322144, 0.011061547, -0.006825576, 0.033158936, -0.10063759, -0.016583988, 0.008227036, 0.05604638, -0.0039418507, 0.030264255, 0.006545456, -0.046788998, -0.06612186, 0.019110108, 0.010173552, -0.0015304928, -0.02745248, 0.08436771, -0.05111628, 0.03491268, -0.018905425, 0.009436589, -0.071091056, 0.06312779, 0.055885248, -0.008187491, 0.013967105, 0.049851406, -0.046775173, -0.05380721, -0.02520902, 0.048415728, -0.053037673, 0.08821214, -0.04349023, -0.002511317, -0.013129268, -0.04000359, -0.0100794975, -0.0659472, 0.044489317, -0.03651276, 0.0032823374, -0.004647774, -0.019675476, 0.11854173, 0.035627883, 0.015952459, -0.017490689, -0.009468227, -0.034936972, -0.0040077316, -0.014501512, -0.040732305, -0.004475036, 0.026295308, 0.11893579, -0.012221011, 0.01921595, 0.003704211, -0.00081420684, 0.031362444, 0.021526098, -0.03796045, -0.04051389, 0.08994492, 0.020430982, -0.13368618, 0.059530005, 0.02978135, -0.020171875, 0.07243986, 0.08047519, 0.014236827, 0.023928184, -0.056827813, 0.030533543, -0.01695773, 0.0019564428, 0.019315101, 0.048426118, 0.012069902, 0.014532966, 0.07157925, -0.00082132005, -0.03102693, 0.05207618, 0.033050887, -0.06816059, 0.037159886, 0.012156096, 0.0906456, 0.05786973, 0.021087963, -0.03615757, -0.0006905898, 0.0062891473, 0.054622658, -0.02605763, -0.050890833, -0.00017370642, -0.010385195, 0.022578984, 0.001822225, -0.045328267, 0.015035055, 0.05529688, 0.046605356, -0.0007772419, -0.09158666, -0.039371215, -0.0026332953, 0.022653094, 0.077683136, -0.027678892, -0.07956019, -0.08317627, 0.012950206, -0.04643972, 0.027308058, -0.007675166, 0.009162879, -0.0064983773, -0.0073145335, 0.041186735, -0.027793638, -0.00047516363, 0.014808601, 0.052241515, 0.07800082, 0.048793413, 0.018123679, 0.06639319, 0.0056572245, -0.0023089426, -0.0012806753, 0.07676211, -0.08715853, -0.02962473, -0.009583457, -0.028001878, -0.0037823156, 0.048585914, -0.017176645, -0.028013503, -0.04553737, -0.04014757, -0.012503475, -0.098679036, -0.031309552, -0.07011677, 0.0286711, -0.007448121, -0.03362688, 0.014612736, 0.006140878, 0.050224025, -0.03131365, 0.017277641, -0.012991993, -0.045904126, 0.006959225, 0.044762693, -0.0052471757, -0.009494742, 0.020247253, -0.025165197, -0.007513343, -0.007732138, -0.03059627, -0.027137207, 0.030832471, -0.0006397405, 0.026458293, 0.048394475, -0.014066572, -0.008397393, 0.030369336, -0.0018644024, -0.08373501, 0.02299318, 0.08410273, 0.03791566, 0.016693544, -0.022285318, -0.107647866, 0.008533737, 0.05805777, 0.063223496, 0.043848637, -0.033787355, 0.013578734, 0.020149017, 0.059982095, -0.016969858, 0.04481642, 0.027871825, 0.037242968, 0.04364479, -0.05280717, 0.008205654, -0.03536789, 0.020066299, 0.02891452, 0.029394835, 0.09834288, 0.03443311, 0.038843676, -0.023331352, -0.0022070059, 0.039741606, 0.033018216, 0.04989029, 0.035506245, 0.026467659, 0.034031004, 0.029856045, -0.06866382, -0.0496181, 0.063887335, -0.02873221, 0.024889331, 0.01833896, -0.010304041, -0.048351713, -0.083444275, -0.030584292, -0.092650875, 0.012108162, -0.022506258, 0.014489741, -0.037093587, -0.0041784635, 0.08624283, -0.012284314, -0.014817595, -0.0073567405, -0.013233772, -0.07208923, 0.049182527, -0.019994823, 0.006094942, -0.014795295, -0.017715558, -0.021894615, -0.01329216, 0.0032535691, -0.061918758, -0.0027641011, -0.04525581, 0.051380426, -0.027817326, 0.040541418, 0.020033667, 0.027792405, -0.059075374, 0.026320897, 0.012968171, -0.002865264, -0.017004456, 0.041212566, 0.0038082711, -0.08282011, -0.052709907, -0.041330304, 0.06054631, -0.08095043, 0.017253665, 0.066494696, -0.0356273, -0.059468318, 0.032792054, 0.10238864, 0.029640062, 0.06367693, -0.000065915876, -0.07408563, -0.035968937, -0.06602596, 0.024129247, 0.002624706, -0.0044429703, -0.038953166, -0.02367998, -0.009588521, 0.031618122, -0.063372254, 0.05579818, -0.00065322284, -0.012777491, -0.04045443, -0.015359356, -0.08424052, 0.016582847, 0.04319089, 0.03904139, -0.004957754, 0.03633682, 0.016728338, 0.0071737715, 0.07263827, -0.059946816, 0.020960696, 0.05819421, -0.0047716517, -0.00028777352, -0.044942997, 0.019640505, -0.0060415184, -0.009499886, 0.03395488, -0.05268878, -0.040615927, 0.05501862, 0.0143708, -0.084489234, -0.046911728, -0.04033474, 0.050277222, 0.04054947, 0.014454217, -0.023438897, -0.05800994, -0.029950928, 0.0032126154, 0.0017874262, 0.025801007, 0.08680619, 0.017868958, 0.0035924045, -0.04201902}},
		}}

		file.AddFaces(faces)

		assert.Equal(t, 1, len(*file.Markers()))

		if err := file.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, file.FileMissing)
		assert.NotEmpty(t, file.FileUID)
		assert.NotEmpty(t, file.Markers())
	})
	t.Run("NoEmbeddings", func(t *testing.T) {
		file := &File{FileUID: "fqzuh65p4sjk3kd1", FileHash: "146b3897eec9ef75e35fbf0bbc4c83c55ca41e31", FileType: "jpg", FileWidth: 720, FileName: "FacesTest", PhotoID: 1000003, FilePrimary: false}

		faces := face.Faces{face.Face{
			Rows:      480,
			Cols:      720,
			Score:     45,
			Area:      face.NewArea("face", 250, 200, 10),
			Eyes:      face.Areas{face.NewArea("eye_l", 240, 195, 1), face.NewArea("eye_r", 240, 205, 1)},
			Landmarks: face.Areas{face.NewArea("a", 250, 185, 2), face.NewArea("b", 250, 215, 2)},
		}}

		file.AddFaces(faces)

		assert.Equal(t, 0, len(*file.Markers()))
	})
}

func TestFile_ValidFaceCount(t *testing.T) {
	t.Run("FileFixturesExampleBridge", func(t *testing.T) {
		file := FileFixturesExampleBridge

		result := file.ValidFaceCount()

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

func TestFile_UnsavedMarkers(t *testing.T) {
	t.Run("bridge2.jpg", func(t *testing.T) {
		m := FileFixtures.Get("bridge2.jpg")
		assert.Equal(t, "ft2es49w15bnlqdw", m.FileUID)
		assert.False(t, m.UnsavedMarkers())

		markers := m.Markers()

		assert.Equal(t, 1, m.ValidFaceCount())
		assert.Equal(t, 1, markers.ValidFaceCount())
		assert.Equal(t, 1, markers.DetectedFaceCount())
		assert.False(t, m.UnsavedMarkers())
		assert.False(t, markers.Unsaved())

		newMarker := *NewMarker(m, cropArea1, "lt9k3pw1wowuy1c1", SrcManual, MarkerFace, 100, 65)

		markers.Append(newMarker)

		assert.Equal(t, 1, m.ValidFaceCount())
		assert.Equal(t, 2, markers.ValidFaceCount())
		assert.Equal(t, 1, markers.DetectedFaceCount())
		assert.True(t, m.UnsavedMarkers())
		assert.True(t, markers.Unsaved())
	})
}

func TestFile_ReplaceHash(t *testing.T) {
	t.Run("exampleFileName.jpg", func(t *testing.T) {
		m := FileFixtures.Get("exampleFileName.jpg")

		if err := m.ReplaceHash(""); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFile_SetHDR(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := FileFixtures.Get("exampleFileName.jpg")

		assert.Equal(t, false, m.IsHDR())
		m.SetHDR(false)
		assert.Equal(t, false, m.IsHDR())
		m.SetHDR(true)
		assert.Equal(t, true, m.IsHDR())
		m.ResetHDR()
		assert.Equal(t, false, m.IsHDR())
	})
}

func TestFile_SetColorProfile(t *testing.T) {
	t.Run("DisplayP3", func(t *testing.T) {
		m := FileFixtures.Get("exampleFileName.jpg")

		assert.Equal(t, "", m.ColorProfile())
		assert.True(t, m.HasColorProfile(colors.Default))
		assert.False(t, m.HasColorProfile(colors.ProfileDisplayP3))

		m.SetColorProfile(string(colors.ProfileDisplayP3))

		assert.Equal(t, "Display P3", m.ColorProfile())
		assert.False(t, m.HasColorProfile(colors.Default))
		assert.True(t, m.HasColorProfile(colors.ProfileDisplayP3))

		m.SetColorProfile("")

		assert.Equal(t, "Display P3", m.ColorProfile())
		assert.False(t, m.HasColorProfile(colors.Default))
		assert.True(t, m.HasColorProfile(colors.ProfileDisplayP3))

		m.ResetColorProfile()

		assert.Equal(t, "", m.ColorProfile())
		assert.True(t, m.HasColorProfile(colors.Default))
		assert.False(t, m.HasColorProfile(colors.ProfileDisplayP3))
	})
}

func TestFile_SetFPS(t *testing.T) {
	t.Run("FileDuration", func(t *testing.T) {
		m := File{FileDuration: time.Second * 60}

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 0.0, m.FileFPS)
		assert.Equal(t, 0, m.FileFrames)

		m.SetFPS(10)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 10.0, m.FileFPS)
		assert.Equal(t, 600, m.FileFrames)

		m.SetFPS(20)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 20.0, m.FileFPS)
		assert.Equal(t, 600, m.FileFrames)

		m.FileFrames = 0
		m.SetFPS(20)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 20.0, m.FileFPS)
		assert.Equal(t, 1200, m.FileFrames)
	})
}

func TestFile_SetFrames(t *testing.T) {
	t.Run("FileDuration", func(t *testing.T) {
		m := File{FileDuration: time.Second * 60}

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 0.0, m.FileFPS)
		assert.Equal(t, 0, m.FileFrames)

		m.SetFrames(120)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 2.0, m.FileFPS)
		assert.Equal(t, 120, m.FileFrames)

		m.SetFrames(30)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 2.0, m.FileFPS)
		assert.Equal(t, 30, m.FileFrames)

		m.FileFPS = 0
		m.SetFrames(30)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 0.5, m.FileFPS)
		assert.Equal(t, 30, m.FileFrames)
	})
}

func TestFile_SetDuration(t *testing.T) {
	t.Run("FileFPS", func(t *testing.T) {
		m := File{FileFPS: 20}

		assert.Equal(t, time.Duration(0), m.FileDuration)
		assert.Equal(t, 20.0, m.FileFPS)
		assert.Equal(t, 0, m.FileFrames)

		m.SetDuration(time.Second * 10)

		assert.Equal(t, time.Second*10, m.FileDuration)
		assert.Equal(t, 20.0, m.FileFPS)
		assert.Equal(t, 200, m.FileFrames)

		m.SetDuration(time.Minute)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 20.0, m.FileFPS)
		assert.Equal(t, 200, m.FileFrames)
	})
	t.Run("FileFrames", func(t *testing.T) {
		m := File{FileFrames: 600}

		assert.Equal(t, time.Duration(0), m.FileDuration)
		assert.Equal(t, 0.0, m.FileFPS)
		assert.Equal(t, 600, m.FileFrames)

		m.SetDuration(time.Minute)

		assert.Equal(t, time.Minute, m.FileDuration)
		assert.Equal(t, 10.0, m.FileFPS)
		assert.Equal(t, 600, m.FileFrames)

		m.FileFPS = 0
		m.FileFrames = 0

		m.SetDuration(time.Hour)
		m.SetFrames(216000)

		assert.Equal(t, time.Hour, m.FileDuration)
		assert.Equal(t, 60.0, m.FileFPS)
		assert.Equal(t, 216000, m.FileFrames)
	})
}

func TestFile_Bitrate(t *testing.T) {
	t.Run("HasDuration", func(t *testing.T) {
		m := File{FileDuration: 1e9 * 20.302, FileSize: 1826192}

		assert.InEpsilon(t, 0.719, m.Bitrate(), 0.01)
	})
	t.Run("NoDuration", func(t *testing.T) {
		m := File{FileDuration: 0, FileSize: 1826192}

		assert.Equal(t, float64(0), m.Bitrate())
	})
	t.Run("NoSize", func(t *testing.T) {
		m := File{FileDuration: 1e9 * 20.302, FileSize: 0}

		assert.Equal(t, float64(0), m.Bitrate())
	})
}

func TestFile_Orientation(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		m := File{FileOrientation: 8}
		assert.Equal(t, 8, m.Orientation())
	})
	t.Run("Empty", func(t *testing.T) {
		m := File{FileOrientation: 0}
		assert.Equal(t, 0, m.Orientation())
	})
	t.Run("Invalid", func(t *testing.T) {
		m := File{FileOrientation: 10}
		assert.Equal(t, 0, m.Orientation())
	})
	t.Run("Negative", func(t *testing.T) {
		m := File{FileOrientation: -1}
		assert.Equal(t, 0, m.Orientation())
	})
}

func TestFile_SetOrientation(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		m := File{FileOrientation: 8}
		assert.Equal(t, 8, m.Orientation())
		assert.Equal(t, "", m.FileOrientationSrc)
		m.SetOrientation(1, SrcManual)
		assert.Equal(t, 1, m.Orientation())
		assert.Equal(t, SrcManual, m.FileOrientationSrc)
	})
	t.Run("Invalid", func(t *testing.T) {
		m := File{FileOrientation: 8}
		assert.Equal(t, 8, m.Orientation())
		assert.Equal(t, "", m.FileOrientationSrc)
		m.SetOrientation(-1, SrcManual)
		assert.Equal(t, 8, m.Orientation())
		assert.Equal(t, "", m.FileOrientationSrc)
	})
}
