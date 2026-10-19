package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestFindCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(FindCommand, []string{"find", "--csv"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "File Name;Mime Type;")
	})
}

func TestFindCommand_FileNameControlCharacters(t *testing.T) {
	f := entity.FileFixturesExampleJPG
	stored, err := query.FileByUID(f.FileUID)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).
			UpdateColumn("file_name", stored.FileName).Error)
	})

	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).
		UpdateColumn("file_name", "1990/Ab\x1bc\x01d\x07e\nf.jpg").Error)

	output, err := RunWithTestContext(FindCommand, []string{"find", "uid:" + f.PhotoUID})

	assert.NoError(t, err)
	assert.Contains(t, output, "1990/Ab[U+001B]c[U+0001]d[U+0007]e f.jpg")
	assert.NotContains(t, output, "\x07")
	assert.NotContains(t, output, "\x1b")

	t.Run("NUL", func(t *testing.T) {
		if entity.DbDialect() == dsn.DialectPostgreSQL {
			t.Skip("PostgreSQL cannot store NUL in text")
		}

		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).
			UpdateColumn("file_name", "1990/a\x00b.jpg").Error)

		nulOutput, nulErr := RunWithTestContext(FindCommand, []string{"find", "uid:" + f.PhotoUID})
		require.NoError(t, nulErr)
		assert.Contains(t, nulOutput, "1990/a[U+0000]b.jpg")
		assert.NotContains(t, nulOutput, "\x00")
	})
	t.Run("JSON", func(t *testing.T) {
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_uid = ?", f.FileUID).
			UpdateColumn("file_name", "1990/a\u202Eb\x7fc.jpg").Error)

		jsonOutput, jsonErr := RunWithTestContext(FindCommand, []string{"find", "--json", "uid:" + f.PhotoUID})
		require.NoError(t, jsonErr)
		assert.Contains(t, jsonOutput, `1990/a\u202eb\u007fc.jpg`)
		assert.NotContains(t, jsonOutput, "\u202E")
	})
}
