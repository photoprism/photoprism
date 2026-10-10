package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/media/colors"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newColorFile creates a file of a picture and removes it afterwards.
func newColorFile(t *testing.T, photo *Photo, m File) *File {
	t.Helper()

	m.PhotoID, m.PhotoUID = photo.ID, photo.PhotoUID
	m.FileUID = rnd.GenerateUID(FileUID)
	m.FileName = "colors/" + m.FileUID + ".jpg"
	m.FileHash = rnd.GenerateUID('h')

	require.NoError(t, m.Create())
	t.Cleanup(func() { _ = UnscopedDb().Unscoped().Delete(&m).Error })

	return &m
}

// findColorFile loads a file row by UID.
func findColorFile(t *testing.T, uid string) (found File) {
	t.Helper()
	require.NoError(t, UnscopedDb().Where("file_uid = ?", uid).First(&found).Error)
	return found
}

// TestColors_Create pins that measured colors and chroma values keep their value when a row is created,
// while fields that are left blank get the "unknown" column default.
func TestColors_Create(t *testing.T) {
	t.Run("Measured", func(t *testing.T) {
		photo := NewPhoto(false)
		photo.PhotoUID = rnd.GenerateUID(PhotoUID)
		photo.PhotoColor = colors.Black.ID()
		createShapePhoto(t, &photo)

		found, err := findShapePhoto(photo.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, int16(16), found.PhotoColor)

		file := newColorFile(t, &photo, File{FileChroma: 1, FileDiff: 512, FileMainColor: "black", FilePrimary: true})
		stored := findColorFile(t, file.FileUID)
		assert.Equal(t, int16(1), stored.FileChroma)
		assert.Equal(t, 512, stored.FileDiff)
	})
	t.Run("Unknown", func(t *testing.T) {
		// Existing SQLite columns have no default, so new pictures start with an unknown color.
		photo := NewPhoto(false)
		assert.Equal(t, int16(-1), photo.PhotoColor)

		photo.PhotoUID = rnd.GenerateUID(PhotoUID)
		createShapePhoto(t, &photo)

		found, err := findShapePhoto(photo.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, int16(-1), found.PhotoColor)

		// A picture created without the constructor gets the column default.
		blank := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID)}
		createShapePhoto(t, blank)

		found, err = findShapePhoto(blank.PhotoUID)
		require.NoError(t, err)
		assert.Equal(t, int16(-1), found.PhotoColor)

		file := newColorFile(t, &photo, File{})
		stored := findColorFile(t, file.FileUID)
		assert.Equal(t, int16(-1), stored.FileChroma)
		assert.Equal(t, -1, stored.FileDiff)
	})
}

// TestColors_Migrations checks that the migrations of stored colors, chroma and diff values map 0 to the
// current values, leave other values unchanged, and change nothing when they run again.
func TestColors_Migrations(t *testing.T) {
	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	ids := []string{"20261010-000001", "20261010-000002", "20261010-000003"}
	statements := make(map[string][]string)

	for _, m := range migrate.Dialects[conn.Driver] {
		statements[m.ID] = m.Statements
	}

	for _, id := range ids {
		require.NotEmpty(t, statements[id], "migration %s", id)
	}

	// newPhoto creates a picture with the given files, then stores the given color, chroma and diff values.
	newPhoto := func(t *testing.T, color int16, files ...File) (*Photo, []*File) {
		photo := NewPhoto(false)
		photo.PhotoUID = rnd.GenerateUID(PhotoUID)
		createShapePhoto(t, &photo)
		require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("photo_color", color).Error)

		result := make([]*File, 0, len(files))

		for _, f := range files {
			chroma, diff := f.FileChroma, f.FileDiff
			file := newColorFile(t, &photo, f)
			require.NoError(t, UnscopedDb().Model(&File{}).Where("id = ?", file.ID).
				UpdateColumns(Values{"file_chroma": chroma, "file_diff": diff}).Error)
			result = append(result, file)
		}

		return &photo, result
	}

	black, blackFiles := newPhoto(t, 0, File{FilePrimary: true, FileMainColor: "black", FileColors: "000000000"})
	grey, greyFiles := newPhoto(t, 0,
		File{FilePrimary: true, FileMainColor: "grey", FileColors: "111111111", FileDiff: 700},
		File{FileMainColor: "black", FileColors: "000000000", FileDiff: 0})
	unknown, unknownFiles := newPhoto(t, 0, File{FilePrimary: true})
	noFiles, _ := newPhoto(t, 0)
	measured, measuredFiles := newPhoto(t, 9, File{FilePrimary: true, FileMainColor: "black", FileColors: "000000000", FileChroma: 32, FileDiff: 512})

	// A NULL color grid stands for colors that were never computed.
	require.NoError(t, UnscopedDb().Exec("UPDATE files SET file_colors = NULL WHERE id = ?", unknownFiles[0].ID).Error)

	// run executes the migrations.
	run := func() {
		for _, id := range ids {
			for _, s := range statements[id] {
				require.NoError(t, UnscopedDb().Exec(s).Error)
			}
		}
	}

	// check asserts the color of each picture, and the chroma and diff of its files.
	check := func() {
		for _, c := range []struct {
			name   string
			photo  *Photo
			color  int16
			files  []*File
			chroma []int16
			diff   []int
		}{
			{"Black", black, 16, blackFiles, []int16{1}, []int{-1}},
			{"Grey", grey, -1, greyFiles, []int16{1, 1}, []int{700, -1}},
			{"Unknown", unknown, -1, unknownFiles, []int16{-1}, []int{-1}},
			{"NoFiles", noFiles, -1, nil, nil, nil},
			{"Measured", measured, 9, measuredFiles, []int16{32}, []int{512}},
		} {
			found, err := findShapePhoto(c.photo.PhotoUID)
			require.NoError(t, err)
			assert.Equal(t, c.color, found.PhotoColor, "color of %s", c.name)

			for i, f := range c.files {
				stored := findColorFile(t, f.FileUID)
				assert.Equal(t, c.chroma[i], stored.FileChroma, "chroma of %s file %d", c.name, i)
				assert.Equal(t, c.diff[i], stored.FileDiff, "diff of %s file %d", c.name, i)
			}
		}
	}

	run()
	check()

	run()
	check()
}
