package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhotoUnstack(t *testing.T) {
	t.Run("UnstackXmpSidecarFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bw45bnlqdw/unstack")
		// Sidecar files can not be unstacked.
		assert.Equal(t, http.StatusBadRequest, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
	t.Run("UnstackBridge3Jpg", func(t *testing.T) {
		app, router, c := NewApiTest()
		PhotoUnstack(router)
		require.NoError(t, fs.Copy("./testdata/london_160x160.jpg", filepath.Join(c.Options().OriginalsPath, "London", "bridge3.jpg"), true))
		require.NoError(t, fs.Copy("./testdata/face_160x160.jpg", filepath.Join(c.Options().OriginalsPath, "1990", "04", "bridge2.jpg"), true))
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bwhhbnlqdn/unstack")
		assert.Equal(t, http.StatusOK, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
	t.Run("CaptureFiles", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)

		photo := entity.PhotoFixtures.Get("Photo04")

		for fileName, originalName := range map[string]string{
			"2022/insta360/VID_20220625_140410_10_008.insv": "",
			"2022/insta360/LRV_20220625_140410_11_008.insv": "",
			"2026/09/20260925_135937_07784009.00001.insv":   "VID_20220625_140410_10_008.insv",
		} {
			file := entity.File{
				FileUID:      rnd.GenerateUID(entity.FileUID),
				PhotoID:      photo.ID,
				PhotoUID:     photo.PhotoUID,
				FileName:     fileName,
				OriginalName: originalName,
				FileRoot:     entity.RootOriginals,
				FileHash:     rnd.GenerateUID(entity.FileUID),
			}
			require.NoError(t, file.Create())
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.File{}, "file_uid = ?", file.FileUID).Error })

			r := PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+file.FileUID+"/unstack")
			assert.Equal(t, http.StatusBadRequest, r.Code, fileName)

			var result entity.File
			require.NoError(t, entity.UnscopedDb().First(&result, "file_uid = ?", file.FileUID).Error)
			assert.Equal(t, photo.ID, result.PhotoID, fileName)
		}
	})
	t.Run("LookAlikeFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)

		// A file whose name only resembles a capture file passes the check and is not found on disk.
		photo := entity.PhotoFixtures.Get("Photo04")
		file := entity.File{
			FileUID:  rnd.GenerateUID(entity.FileUID),
			PhotoID:  photo.ID,
			PhotoUID: photo.PhotoUID,
			FileName: "2022/insta360/VID_20220625_140410_10_008.mp4",
			FileRoot: entity.RootOriginals,
			FileHash: rnd.GenerateUID(entity.FileUID),
		}
		require.NoError(t, file.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&file).Error })

		r := PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+file.FileUID+"/unstack")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("CaptureNeighbor", func(t *testing.T) {
		// Other files stacked with a capture are unstacked without its lens files, which stay together.
		cases := []struct {
			photoName string
			target    string
			status    int
		}{
			{"VID_20220625_140410_00_008", "VID_20220625_140410_10_008.mp4", 0},
			{"VID_20220625_140410_00_008", "VID_20220625_140410_10_008.insv.jpg", 0},
			{"VID_20220625_140410_00_008", "VID_20220625_140410_00_008.insv", http.StatusBadRequest},
			{"IMG_9999", "VID_20220625_140410_00_008.mp4", 0},
		}

		for _, tc := range cases {
			app, router, conf := NewApiTest()
			PhotoUnstack(router)

			folder := "unstackcapture" + rnd.Base36(6)
			dir := filepath.Join(conf.OriginalsPath(), folder)
			require.NoError(t, fs.MkdirAll(dir))
			t.Cleanup(func() {
				_ = os.RemoveAll(dir)
				_ = entity.UnscopedDb().Where("file_name LIKE ?", folder+"/%").Delete(&entity.File{}).Error
				_ = entity.UnscopedDb().Where("photo_path = ?", folder).Delete(&entity.Photo{}).Error
			})

			left, right := "VID_20220625_140410_00_008.insv", "VID_20220625_140410_10_008.insv"
			names := []string{left, right}

			if tc.target != left {
				names = append(names, tc.target)
			}

			for _, name := range names {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(name), fs.ModeFile))
			}

			photo := entity.NewUserPhoto(true, "")
			photo.PhotoPath = folder
			photo.PhotoName = tc.photoName
			require.NoError(t, photo.Create())

			newFile := func(name, root string, primary bool) *entity.File {
				f := &entity.File{
					FileUID:     rnd.GenerateUID(entity.FileUID),
					PhotoID:     photo.ID,
					PhotoUID:    photo.PhotoUID,
					FileName:    folder + "/" + name,
					FileRoot:    root,
					FilePrimary: primary,
					FileHash:    rnd.GenerateUID(entity.FileUID),
				}
				require.NoError(t, f.Create())
				return f
			}

			newFile(left+".jpg", entity.RootSidecar, true)
			leftFile := newFile(left, entity.RootOriginals, false)
			rightFile := newFile(right, entity.RootOriginals, false)
			target := leftFile

			if tc.target != left {
				target = newFile(tc.target, entity.RootOriginals, false)
			}

			r := PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+target.FileUID+"/unstack")

			if tc.status != 0 {
				assert.Equal(t, tc.status, r.Code, tc.target)
			} else {
				assert.NotEqual(t, http.StatusBadRequest, r.Code, tc.target)
			}

			for _, lens := range []*entity.File{leftFile, rightFile} {
				var result entity.File
				require.NoError(t, entity.UnscopedDb().First(&result, "file_uid = ?", lens.FileUID).Error)
				assert.Equal(t, photo.ID, result.PhotoID, tc.target)
			}
		}
	})
	t.Run("RelatedEdit", func(t *testing.T) {
		app, router, conf := NewApiTest()
		PhotoUnstack(router)

		// Related files with another base name, such as edits, are unstacked together.
		folder := "unstackedit" + rnd.Base36(6)
		dir := filepath.Join(conf.OriginalsPath(), folder)
		require.NoError(t, fs.MkdirAll(dir))
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
			_ = entity.UnscopedDb().Where("file_name LIKE ?", folder+"/%").Delete(&entity.File{}).Error
			_ = entity.UnscopedDb().Where("photo_path = ?", folder).Delete(&entity.Photo{}).Error
		})
		require.NoError(t, fs.Copy("./testdata/london_160x160.jpg", filepath.Join(dir, "IMG_1234.jpg"), true))
		require.NoError(t, fs.Copy("./testdata/face_160x160.jpg", filepath.Join(dir, "IMG_E1234.jpg"), true))

		photo := entity.NewUserPhoto(true, "")
		photo.PhotoPath = folder
		photo.PhotoName = "other"
		require.NoError(t, photo.Create())

		newFile := func(name, root string, primary bool) *entity.File {
			f := &entity.File{
				FileUID:     rnd.GenerateUID(entity.FileUID),
				PhotoID:     photo.ID,
				PhotoUID:    photo.PhotoUID,
				FileName:    folder + "/" + name,
				FileRoot:    root,
				FilePrimary: primary,
				FileHash:    rnd.GenerateUID(entity.FileUID),
			}
			require.NoError(t, f.Create())
			return f
		}

		newFile("other.jpg", entity.RootSidecar, true)
		target := newFile("IMG_1234.jpg", entity.RootOriginals, false)
		edit := newFile("IMG_E1234.jpg", entity.RootOriginals, false)

		PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+target.FileUID+"/unstack")

		var targetResult, editResult entity.File
		require.NoError(t, entity.UnscopedDb().First(&targetResult, "file_uid = ?", target.FileUID).Error)
		require.NoError(t, entity.UnscopedDb().First(&editResult, "file_uid = ?", edit.FileUID).Error)
		require.NotEqual(t, photo.ID, targetResult.PhotoID)
		assert.Equal(t, targetResult.PhotoID, editResult.PhotoID)
	})
	t.Run("ImportedCapture", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)

		// Lens files renamed on import are identified by their original names.
		photo := entity.PhotoFixtures.Get("Photo04")
		files := []entity.File{
			{FileName: "2026/09/20260925_135937_07784009.insv", OriginalName: "VID_20220625_140410_00_008.insv"},
			{FileName: "2026/09/20260925_135937_07784009.00001.insv", OriginalName: "VID_20220625_140410_10_008.insv"},
		}

		for i := range files {
			files[i].FileUID = rnd.GenerateUID(entity.FileUID)
			files[i].PhotoID = photo.ID
			files[i].PhotoUID = photo.PhotoUID
			files[i].FileRoot = entity.RootOriginals
			files[i].FileHash = rnd.GenerateUID(entity.FileUID)
			require.NoError(t, files[i].Create())
			uid := files[i].FileUID
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.File{}, "file_uid = ?", uid).Error })
		}

		for _, file := range files {
			r := PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+file.FileUID+"/unstack")
			assert.Equal(t, http.StatusBadRequest, r.Code, file.OriginalName)
		}
	})
	t.Run("SingleFileCapture", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)

		// Photos with lens codes are single-file captures and pass the check.
		photo := entity.PhotoFixtures.Get("Photo04")

		for _, fileName := range []string{
			"2023/insta360/IMG_20231015_101112_00_123.insp",
			"2023/insta360/IMG_20231015_101112_10_124.insp",
		} {
			file := entity.File{
				FileUID:  rnd.GenerateUID(entity.FileUID),
				PhotoID:  photo.ID,
				PhotoUID: photo.PhotoUID,
				FileName: fileName,
				FileRoot: entity.RootOriginals,
				FileHash: rnd.GenerateUID(entity.FileUID),
			}
			require.NoError(t, file.Create())
			uid := file.FileUID
			t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.File{}, "file_uid = ?", uid).Error })

			r := PerformRequest(app, "POST", "/api/v1/photos/"+photo.PhotoUID+"/files/"+file.FileUID+"/unstack")
			assert.Equal(t, http.StatusNotFound, r.Code, fileName)
		}
	})
	t.Run("NotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/xxx/unstack")
		assert.Equal(t, http.StatusNotFound, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
}
