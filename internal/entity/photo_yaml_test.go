package entity

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPhoto_Yaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		result, err := m.Yaml()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("YAML: %s", result)
	})
}

func TestPhoto_SaveAsYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		fileName := filepath.Join(t.TempDir(), ".photoprism_test.yml")

		if err := m.SaveAsYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := m.LoadFromYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FilenameEmpty", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		err := m.SaveAsYaml("")

		assert.Error(t, err)
	})
	t.Run("NoPhotoUID", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		fileName := filepath.Join(os.TempDir(), ".photoprism_test.yml")

		err := m.SaveAsYaml(fileName)

		assert.Error(t, err)
	})
}

func TestPhoto_YamlFileName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()
		fileName, relative, err := m.YamlFileName("xxx", "yyy")
		assert.NoError(t, err)
		assert.Equal(t, "xxx/2790/02/yyy/Photo01.yml", fileName)
		assert.Equal(t, "2790/02/Photo01.yml", relative)

		if err := os.RemoveAll("xxx"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_SaveSidecarYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		m.PreloadFiles()

		basePath := t.TempDir()
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.SaveSidecarYaml(originalsPath, sidecarPath); err != nil {
			t.Error(err)
		}

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoNameEmpty", func(t *testing.T) {
		m := Photo{}
		m.PreloadFiles()

		basePath := t.TempDir()
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
	t.Run("PhotoUIDEmpty", func(t *testing.T) {
		m := Photo{PhotoName: "testphoto"}
		m.PreloadFiles()

		basePath := t.TempDir()
		originalsPath := filepath.Join(basePath, "originals")
		sidecarPath := filepath.Join(basePath, "sidecar")

		t.Logf("originalsPath: %s", originalsPath)
		t.Logf("sidecarPath: %s", sidecarPath)

		if err := fs.MkdirAll(originalsPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := fs.MkdirAll(sidecarPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveSidecarYaml(originalsPath, sidecarPath)

		assert.Error(t, err)

		if err := os.RemoveAll(basePath); err != nil {
			t.Error(err)
		}
	})
}

func TestPhoto_LoadFromYaml(t *testing.T) {
	t.Run("EmptyFilename", func(t *testing.T) {
		m := Photo{}

		err := m.LoadFromYaml("")

		assert.Error(t, err)
	})

	t.Run("GormV1Format", func(t *testing.T) {
		filePath := filepath.Join(os.TempDir())

		if err := os.MkdirAll(filePath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		fileName := filepath.Join(filePath, ".gormv1_format.yml")

		newYaml := []byte("UID: as6sg6bipotaajfa\nDeletedAt: 2025-06-30T10:33:49Z\nType: moment\nTitle: Walking Cows\nAltitude: 0\nOriginalName: test/folder/image_123445\nCreatedAt: 2020-01-01T00:00:00Z\nUpdatedAt: 2025-06-30T10:33:49Z\n")
		err := os.WriteFile(fileName, newYaml, 0644)
		assert.NoError(t, err)

		photoToCheck := Photo{}

		err = photoToCheck.LoadFromYaml(fileName)
		assert.NoError(t, err)

		assert.Equal(t, "as6sg6bipotaajfa", photoToCheck.PhotoUID)
		assert.Equal(t, "moment", photoToCheck.PhotoType)
		assert.Equal(t, "Walking Cows", photoToCheck.PhotoTitle)
		assert.Equal(t, 0, photoToCheck.PhotoAltitude)
		assert.Equal(t, "test/folder/image_123445", photoToCheck.OriginalName)
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), photoToCheck.CreatedAt)
		assert.Equal(t, time.Date(2025, 6, 30, 10, 33, 49, 0, time.UTC), photoToCheck.UpdatedAt)
		assert.Equal(t, gorm.DeletedAt{Time: time.Date(2025, 6, 30, 10, 33, 49, 0, time.UTC), Valid: true}, photoToCheck.DeletedAt)

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GormV2Format", func(t *testing.T) {
		filePath := t.TempDir()

		if err := os.MkdirAll(filePath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		fileName := filepath.Join(filePath, ".gormv2_format.yml")

		newYaml := []byte("UID: as6sg6bipotaajfa\nType: moment\nTitle: Flying Cows\nAltitude: 100\nOriginalName: test/folder/image_123446\nCreatedAt: 2020-01-01T00:00:00Z\nUpdatedAt: 2025-06-30T10:33:49Z\nDeletedAt:\n  time: 2025-06-30T10:33:50Z\n  valid: true\n")
		err := os.WriteFile(fileName, newYaml, 0644)
		assert.NoError(t, err)

		photoToCheck := Photo{}

		err = photoToCheck.LoadFromYaml(fileName)
		assert.NoError(t, err)

		assert.Equal(t, "as6sg6bipotaajfa", photoToCheck.PhotoUID)
		assert.Equal(t, "moment", photoToCheck.PhotoType)
		assert.Equal(t, "Flying Cows", photoToCheck.PhotoTitle)
		assert.Equal(t, 100, photoToCheck.PhotoAltitude)
		assert.Equal(t, "test/folder/image_123446", photoToCheck.OriginalName)
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), photoToCheck.CreatedAt)
		assert.Equal(t, time.Date(2025, 6, 30, 10, 33, 49, 0, time.UTC), photoToCheck.UpdatedAt)
		assert.Equal(t, gorm.DeletedAt{Time: time.Date(2025, 6, 30, 10, 33, 50, 0, time.UTC), Valid: true}, photoToCheck.DeletedAt)

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GormV1Format_Bad", func(t *testing.T) {
		filePath := t.TempDir()

		if err := os.MkdirAll(filePath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		fileName := filepath.Join(filePath, ".gormv1_format_bad.yml")

		newYaml := []byte("UID: as6sg6bipotaajfa\nDeletedAt: 2025-06-30T10:33:49Z\nType: moment\nTitle: Walking Cows\nAltitude: GroundLevel\nOriginalName: test/folder/image_123445\nCreatedAt: 2020-01-01T00:00:00Z\nUpdatedAt: 2025-06-30T10:33:49Z\n")
		err := os.WriteFile(fileName, newYaml, 0644)
		assert.NoError(t, err)

		photoToCheck := Photo{}

		err = photoToCheck.LoadFromYaml(fileName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "!!timestamp")
		assert.Contains(t, err.Error(), "!!str")

		assert.Equal(t, "as6sg6bipotaajfa", photoToCheck.PhotoUID)
		assert.Equal(t, "moment", photoToCheck.PhotoType)
		assert.Equal(t, "Walking Cows", photoToCheck.PhotoTitle)
		assert.Equal(t, 0, photoToCheck.PhotoAltitude)
		assert.Equal(t, "test/folder/image_123445", photoToCheck.OriginalName)
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), photoToCheck.CreatedAt)
		assert.Equal(t, time.Date(2025, 6, 30, 10, 33, 49, 0, time.UTC), photoToCheck.UpdatedAt)
		assert.Equal(t, gorm.DeletedAt{}, photoToCheck.DeletedAt)

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GormV2Format_Bad", func(t *testing.T) {
		filePath := t.TempDir()

		if err := os.MkdirAll(filePath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		fileName := filepath.Join(filePath, ".gormv2_format_bad.yml")

		newYaml := []byte("UID: as6sg6bipotaajfa\nType: moment\nTitle: Flying Cows\nAltitude: Flying\nOriginalName: test/folder/image_123446\nCreatedAt: 2020-01-01T00:00:00Z\nUpdatedAt: 2025-06-30T10:33:49Z\nDeletedAt:\n  time: 2025-06-30T10:33:50Z\n  valid: true\n")
		err := os.WriteFile(fileName, newYaml, 0644)
		assert.NoError(t, err)

		photoToCheck := Photo{}

		err = photoToCheck.LoadFromYaml(fileName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "!!str")

		assert.Equal(t, "as6sg6bipotaajfa", photoToCheck.PhotoUID)
		assert.Equal(t, "moment", photoToCheck.PhotoType)
		assert.Equal(t, "Flying Cows", photoToCheck.PhotoTitle)
		assert.Equal(t, 0, photoToCheck.PhotoAltitude)
		assert.Equal(t, "test/folder/image_123446", photoToCheck.OriginalName)
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), photoToCheck.CreatedAt)
		assert.Equal(t, time.Date(2025, 6, 30, 10, 33, 49, 0, time.UTC), photoToCheck.UpdatedAt)
		assert.Equal(t, gorm.DeletedAt{Time: time.Date(2025, 6, 30, 10, 33, 50, 0, time.UTC), Valid: true}, photoToCheck.DeletedAt)

		if err := os.Remove(fileName); err != nil {
			t.Fatal(err)
		}
	})

}
