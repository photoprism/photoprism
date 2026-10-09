package entity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/entity/legacy"
	"github.com/photoprism/photoprism/internal/entity/sqlcount"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// shapeDeletedAt is the deletion time used by the shape tests.
var shapeDeletedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

// jsonFields returns the top-level fields of the JSON encoding of v.
func jsonFields(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	fields := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(data, &fields))

	return fields
}

// TestDeletedAt_JSON pins how each model encodes DeletedAt in API responses.
func TestDeletedAt_JSON(t *testing.T) {
	deletedAt := shapeDeletedAt
	const deletedJson = `"2026-01-02T03:04:05Z"`

	cases := []struct {
		name    string
		active  any
		deleted any
		// omitted means the field is absent when not deleted, hidden means it is never present.
		omitted bool
		hidden  bool
	}{
		{"Photo", &Photo{}, &Photo{DeletedAt: &deletedAt}, true, false},
		{"File", &File{}, &File{DeletedAt: &deletedAt}, true, false},
		{"Label", &Label{}, &Label{DeletedAt: &deletedAt}, true, false},
		{"Subject", &Subject{}, &Subject{DeletedAt: &deletedAt}, true, false},
		{"User", &User{}, &User{DeletedAt: &deletedAt}, true, false},
		{"Client", &Client{}, &Client{DeletedAt: &deletedAt}, true, false},
		{"LegacyUser", &legacy.User{}, &legacy.User{DeletedAt: &deletedAt}, true, false},
		{"Album", &Album{}, &Album{DeletedAt: &deletedAt}, false, false},
		{"Service", &Service{}, &Service{DeletedAt: &deletedAt}, false, false},
		{"Folder", &Folder{}, &Folder{DeletedAt: &deletedAt}, false, true},
		{"Camera", &Camera{}, &Camera{DeletedAt: &deletedAt}, false, true},
		{"Lens", &Lens{}, &Lens{DeletedAt: &deletedAt}, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := jsonFields(t, tc.active)
			deleted := jsonFields(t, tc.deleted)

			switch {
			case tc.hidden:
				assert.NotContains(t, active, "DeletedAt")
				assert.NotContains(t, deleted, "DeletedAt")
			case tc.omitted:
				assert.NotContains(t, active, "DeletedAt")
				assert.JSONEq(t, deletedJson, string(deleted["DeletedAt"]))
			default:
				assert.JSONEq(t, "null", string(active["DeletedAt"]))
				assert.JSONEq(t, deletedJson, string(deleted["DeletedAt"]))
			}
		})
	}
}

// TestDeletedAt_Yaml pins that sidecars and backups store DeletedAt as a plain timestamp.
func TestDeletedAt_Yaml(t *testing.T) {
	deletedAt := shapeDeletedAt

	cases := []struct {
		name    string
		active  any
		deleted any
		decoded func() any
	}{
		{"Photo", &Photo{}, &Photo{DeletedAt: &deletedAt}, func() any { return &Photo{} }},
		{"Album", &Album{}, &Album{DeletedAt: &deletedAt}, func() any { return &Album{} }},
		{"Folder", &Folder{}, &Folder{DeletedAt: &deletedAt}, func() any { return &Folder{} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active, err := yaml.Marshal(tc.active)
			require.NoError(t, err)
			assert.NotContains(t, string(active), "DeletedAt")

			deleted, err := yaml.Marshal(tc.deleted)
			require.NoError(t, err)
			assert.Contains(t, string(deleted), "\nDeletedAt: 2026-01-02T03:04:05Z\n")

			decoded := tc.decoded()
			require.NoError(t, yaml.Unmarshal(deleted, decoded))
			assert.Equal(t, tc.deleted, decoded)
		})
	}
}

// TestDeletedAt_YamlFile pins that sidecar and backup files with a plain DeletedAt timestamp restore it.
func TestDeletedAt_YamlFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("Photo", func(t *testing.T) {
		fileName := filepath.Join(dir, "photo.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("UID: ps6sg6be2lvl0yh0\nTitle: Archived\nDeletedAt: 2026-01-02T03:04:05Z\n"), 0o600))

		m := &Photo{}
		require.NoError(t, m.LoadFromYaml(fileName))
		require.NotNil(t, m.DeletedAt)
		assert.True(t, shapeDeletedAt.Equal(*m.DeletedAt))
	})
	t.Run("Album", func(t *testing.T) {
		fileName := filepath.Join(dir, "album.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("UID: as6sg6bxpogaaba7\nTitle: Archived\nDeletedAt: 2026-01-02T03:04:05Z\n"), 0o600))

		m := &Album{}
		require.NoError(t, m.LoadFromYaml(fileName))
		require.NotNil(t, m.DeletedAt)
		assert.True(t, shapeDeletedAt.Equal(*m.DeletedAt))
	})
	t.Run("PhotoRoundTrip", func(t *testing.T) {
		fileName := filepath.Join(dir, "roundtrip.yml")
		deletedAt := shapeDeletedAt
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoTitle: "Archived", DeletedAt: &deletedAt}
		require.NoError(t, m.SaveAsYaml(fileName))

		data, err := os.ReadFile(fileName) //nolint:gosec // Test file in a temporary directory.
		require.NoError(t, err)
		assert.Contains(t, string(data), "\nDeletedAt: 2026-01-02T03:04:05Z\n")

		restored := &Photo{}
		require.NoError(t, restored.LoadFromYaml(fileName))
		require.NotNil(t, restored.DeletedAt)
		assert.True(t, shapeDeletedAt.Equal(*restored.DeletedAt))
	})
}

// createShapePhoto creates a photo and registers a cleanup that removes it with its details row.
func createShapePhoto(t *testing.T, m *Photo) {
	t.Helper()

	require.NoError(t, m.Create())
	t.Cleanup(func() {
		_ = UnscopedDb().Where("photo_id = ?", m.ID).Delete(&Details{}).Error
		_ = UnscopedDb().Unscoped().Delete(m).Error
	})
}

// findShapePhoto loads a photo row by UID, including soft-deleted ones.
func findShapePhoto(uid string) (found Photo, err error) {
	err = UnscopedDb().Where("photo_uid = ?", uid).First(&found).Error
	return found, err
}

// assertWholeSeconds checks that a timestamp carries no fractional seconds.
func assertWholeSeconds(t *testing.T, name string, ts time.Time) {
	t.Helper()
	assert.False(t, ts.IsZero(), "%s is set", name)
	assert.Equal(t, 0, ts.Nanosecond(), "%s is truncated to seconds", name)
}

// TestTimestamps_Seconds pins that generated timestamps are whole seconds in memory and in the database.
func TestTimestamps_Seconds(t *testing.T) {
	t.Run("AlbumCreate", func(t *testing.T) {
		m := NewAlbum("Shape Timestamps", AlbumManual)

		// Leave both timestamps to the ORM.
		m.CreatedAt = time.Time{}
		m.UpdatedAt = time.Time{}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		assertWholeSeconds(t, "CreatedAt", m.CreatedAt)
		assertWholeSeconds(t, "UpdatedAt", m.UpdatedAt)

		found := FindAlbum(Album{AlbumUID: m.AlbumUID})
		require.NotNil(t, found)
		assert.True(t, m.CreatedAt.Equal(found.CreatedAt), "stored CreatedAt equals the value in memory")
		assert.True(t, m.UpdatedAt.Equal(found.UpdatedAt), "stored UpdatedAt equals the value in memory")
	})
	t.Run("PhotoSave", func(t *testing.T) {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoTitle: "Shape Timestamps", TakenSrc: SrcAuto}
		createShapePhoto(t, m)

		assertWholeSeconds(t, "CreatedAt", m.CreatedAt)

		m.PhotoTitle = "Shape Timestamps Updated"
		require.NoError(t, m.Save())
		assertWholeSeconds(t, "UpdatedAt", m.UpdatedAt)

		found, err := findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.True(t, m.UpdatedAt.Equal(found.UpdatedAt), "stored UpdatedAt equals the value in memory")
	})
}

// TestZeroValues_Persist pins that clearing a value persists it rather than keeping the old one.
func TestZeroValues_Persist(t *testing.T) {
	t.Run("PhotoTitleCaption", func(t *testing.T) {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoTitle: "Shape Title", PhotoCaption: "Shape Caption", TitleSrc: SrcManual, CaptionSrc: SrcManual}
		createShapePhoto(t, m)

		m.PhotoTitle = ""
		m.PhotoCaption = ""
		m.PhotoFavorite = false
		require.NoError(t, m.Save())

		found, err := findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "", found.PhotoTitle)
		assert.Equal(t, "", found.PhotoCaption)
	})
	t.Run("PhotoUpdate", func(t *testing.T) {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoCaption: "Shape Caption", PhotoFavorite: true, PhotoQuality: 3}
		createShapePhoto(t, m)

		m.PhotoCaption = ""
		m.PhotoFavorite = false
		m.PhotoQuality = 0
		require.NoError(t, Update(m, "ID", "PhotoUID"))

		found, err := findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "", found.PhotoCaption)
		assert.False(t, found.PhotoFavorite)
		assert.Equal(t, 0, found.PhotoQuality)
	})
	t.Run("FileDiffChroma", func(t *testing.T) {
		photo := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID)}
		createShapePhoto(t, photo)

		m := &File{
			PhotoID:    photo.ID,
			PhotoUID:   photo.PhotoUID,
			FileUID:    rnd.GenerateUID(FileUID),
			FileName:   "shape/zero-values.jpg",
			FileRoot:   RootOriginals,
			FileHash:   rnd.GenerateUID('h'),
			FileDiff:   500,
			FileChroma: 12,
		}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Unscoped().Delete(m).Error })

		m.FileDiff = 0
		m.FileChroma = 0
		require.NoError(t, m.Save())

		found := File{}
		require.NoError(t, UnscopedDb().Where("file_uid = ?", m.FileUID).First(&found).Error)
		assert.Equal(t, 0, found.FileDiff)
		assert.Equal(t, int16(0), found.FileChroma)
	})
	t.Run("DefaultTaggedUpdate", func(t *testing.T) {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), TimeZone: "Europe/Berlin"}
		createShapePhoto(t, m)

		m.TimeZone = ""
		require.NoError(t, m.Save())

		found, err := findShapePhoto(m.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "", found.TimeZone)
	})
	t.Run("DefaultTaggedAlbumSave", func(t *testing.T) {
		m := NewAlbum("Shape Country", AlbumManual)
		m.AlbumCountry = "de"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.AlbumCountry = ""
		require.NoError(t, m.Save())

		found := FindAlbum(Album{AlbumUID: m.AlbumUID})
		require.NotNil(t, found)
		assert.Equal(t, "", found.AlbumCountry)
	})
	t.Run("DefaultTaggedCreate", func(t *testing.T) {
		photo := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID)}
		createShapePhoto(t, photo)

		// A blank field with a default tag gets the default, in memory and in the database.
		assert.Equal(t, "Local", photo.TimeZone)
		assert.Equal(t, UnknownID, photo.PhotoCountry)

		m := &File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileUID: rnd.GenerateUID(FileUID), FileName: "shape/default-root.jpg", FileHash: rnd.GenerateUID('h')}
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Unscoped().Delete(m).Error })
		assert.Equal(t, RootOriginals, m.FileRoot)

		found := File{}
		require.NoError(t, UnscopedDb().Where("file_uid = ?", m.FileUID).First(&found).Error)
		assert.Equal(t, RootOriginals, found.FileRoot)

		stored, err := findShapePhoto(photo.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, "Local", stored.TimeZone)
		assert.Equal(t, UnknownID, stored.PhotoCountry)
	})
	t.Run("AlbumDescription", func(t *testing.T) {
		m := NewAlbum("Shape Description", AlbumManual)
		m.AlbumDescription = "Shape Description"
		m.AlbumFavorite = true
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.AlbumDescription = ""
		m.AlbumFavorite = false
		require.NoError(t, m.Save())

		found := FindAlbum(Album{AlbumUID: m.AlbumUID})
		require.NotNil(t, found)
		assert.Equal(t, "", found.AlbumDescription)
		assert.False(t, found.AlbumFavorite)
	})
}

// countedStatements runs fn against a counted connection to the test database and returns its statements.
func countedStatements(t *testing.T, fn func()) []string {
	t.Helper()

	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	p, err := sqlcount.OpenGorm(conn.Driver, conn.Dsn)
	require.NoError(t, err)

	SetDbProvider(p)

	defer func() {
		SetDbProvider(conn)
		_ = p.Close()
	}()

	p.Counter.Start()
	fn()

	return p.Counter.Stop()
}

// statementCeiling returns the statement limit for the test database driver, which differ because
// MariaDB reports no affected rows for an update that changes nothing.
func statementCeiling(t *testing.T, sqlite, mysql int) int {
	t.Helper()

	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	switch conn.Driver {
	case "sqlite3":
		return sqlite
	case "mysql":
		return mysql
	default:
		t.Fatalf("no statement ceiling for driver %q", conn.Driver)
		return 0
	}
}

// TestPhoto_SaveStatements pins an upper bound for the statements a photo update issues.
func TestPhoto_SaveStatements(t *testing.T) {
	m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoTitle: "Shape Statements", TitleSrc: SrcManual}
	createShapePhoto(t, m)

	t.Run("Changed", func(t *testing.T) {
		statements := countedStatements(t, func() {
			m.PhotoTitle = "Shape Statements Changed"
			require.NoError(t, m.Save())
		})

		assert.LessOrEqual(t, len(statements), statementCeiling(t, 3, 4), "%q", statements)
	})
	t.Run("Unchanged", func(t *testing.T) {
		statements := countedStatements(t, func() {
			require.NoError(t, m.Save())
		})

		assert.LessOrEqual(t, len(statements), statementCeiling(t, 3, 5), "%q", statements)
	})
}

// TestPhoto_PreloadListShape pins how the preloaded lists of a photo without related rows encode.
func TestPhoto_PreloadListShape(t *testing.T) {
	m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoTitle: "Shape Preload"}
	createShapePhoto(t, m)

	found, err := findShapePhoto(m.PhotoUID)
	require.NoError(t, err)

	found.PreloadMany()
	found.PreloadLabels()
	fields := jsonFields(t, &found)

	expected := map[string]string{"Files": "[]", "Albums": "[]", "Labels": "[]"}

	for name, value := range expected {
		assert.JSONEq(t, value, string(fields[name]), name)
	}
}
