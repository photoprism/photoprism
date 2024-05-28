package entity

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestAlbum_Yaml(t *testing.T) {
	t.Run("berlin-2019", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		result, err := m.Yaml()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("YAML: %s", result)
	})
	t.Run("christmas2030", func(t *testing.T) {
		m := AlbumFixtures.Get("christmas2030")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		result, err := m.Yaml()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("YAML: %s", result)
	})
}

func TestAlbum_SaveAsYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		backupPath := fs.Abs("testdata/TestAlbum_SaveAsYaml")

		fileName, relName, err := m.YamlFileName(backupPath)

		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(fileName, "internal/entity/testdata/TestAlbum_SaveAsYaml/album/as6sg6bxpogaaba9.yml"))
		assert.Equal(t, "album/as6sg6bxpogaaba9.yml", relName)

		if err = m.SaveAsYaml(fileName); err != nil {
			t.Fatal(err)
			return
		}

		if err = m.LoadFromYaml(fileName); err != nil {
			t.Error(err)
		}

		if err = os.RemoveAll(backupPath); err != nil {
			t.Error(err)
		}
	})
	t.Run("FilenameEmpty", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		backupPath := fs.Abs("testdata/TestAlbum_SaveAsYaml")

		err := m.SaveAsYaml("")

		assert.Error(t, err)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Error(err)
		}
	})
	t.Run("NoAlbumUID", func(t *testing.T) {
		m := Album{}

		backupPath := fs.Abs("testdata/TestAlbum_SaveAsYaml")

		err := m.SaveAsYaml("")

		assert.Error(t, err)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Error(err)
		}
	})
}

func TestAlbum_YamlFileName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		fileName, relName, err := m.YamlFileName("/foo/bar")

		assert.NoError(t, err)
		assert.Equal(t, "/foo/bar/album/as6sg6bxpogaaba9.yml", fileName)
		assert.Equal(t, "album/as6sg6bxpogaaba9.yml", relName)
	})
	t.Run("NoAlbumUID", func(t *testing.T) {
		m := Album{}

		fileName, relName, err := m.YamlFileName("/foo/bar")

		assert.Error(t, err)
		assert.Equal(t, "", fileName)
		assert.Equal(t, "", relName)
	})
	t.Run("BackupPathEmpty", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		if found := m.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			m = *found
		}

		fileName, relName, err := m.YamlFileName("")

		assert.Error(t, err)
		assert.Equal(t, "", fileName)
		assert.Equal(t, "album/as6sg6bxpogaaba9.yml", relName)
	})
}

func TestAlbum_SaveBackupYaml(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		backupPath := fs.Abs("testdata/TestAlbum_SaveBackupYaml")

		if err := fs.MkdirAll(backupPath); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.SaveBackupYaml(backupPath); err != nil {
			t.Error(err)
		}

		if err := os.RemoveAll(backupPath); err != nil {
			t.Error(err)
		}
	})
	t.Run("NoAlbumUID", func(t *testing.T) {
		m := Album{}

		backupPath := fs.Abs("testdata/TestAlbum_SaveBackupYaml")

		if err := fs.MkdirAll(backupPath); err != nil {
			t.Fatal(err)
			return
		}

		err := m.SaveBackupYaml(backupPath)

		assert.Error(t, err)

		if err := os.RemoveAll(backupPath); err != nil {
			t.Error(err)
		}
	})
	t.Run("BackupPathEmpty", func(t *testing.T) {
		m := AlbumFixtures.Get("berlin-2019")

		err := m.SaveBackupYaml("")

		assert.Error(t, err)
	})
}

func TestAlbum_LoadFromYaml(t *testing.T) {
	t.Run("berlin-2020", func(t *testing.T) {
		fileName := "testdata/album/as6sg6bxpoaaaaaa.yml"

		m := Album{}

		if err := m.LoadFromYaml(fileName); err != nil {
			t.Fatal(err)
		}

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "as6sg6bxpoaaaaaa"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		} else {
			a = *found
		}

		if existingYaml, err := os.ReadFile(fileName); err != nil {
			t.Fatal(err)
		} else if newYaml, err := a.Yaml(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, existingYaml[:50], newYaml[:50])
			assert.Equal(t, a.AlbumUID, m.AlbumUID)
			assert.Equal(t, a.AlbumSlug, m.AlbumSlug)
			assert.Equal(t, a.AlbumType, m.AlbumType)
			assert.Equal(t, a.AlbumTitle, m.AlbumTitle)
			assert.Equal(t, a.AlbumDescription, m.AlbumDescription)
			assert.Equal(t, a.AlbumOrder, m.AlbumOrder)
			assert.Equal(t, a.AlbumCountry, m.AlbumCountry)
			assert.Equal(t, a.CreatedAt, m.CreatedAt)
			assert.Equal(t, a.UpdatedAt, m.UpdatedAt)
			assert.Equal(t, len(a.Photos), len(m.Photos))
		}
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		m := Album{}

		err := m.LoadFromYaml("")

		assert.Error(t, err)
	})
}
