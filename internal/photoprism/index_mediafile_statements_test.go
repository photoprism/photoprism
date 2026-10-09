package photoprism

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sqlcount"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// countedIndexStatements runs fn against a counted connection to the config's database and returns its statements.
func countedIndexStatements(t *testing.T, c *config.Config, fn func()) []string {
	t.Helper()

	p, err := sqlcount.OpenGorm(c.DatabaseDriver(), c.DatabaseDSN())
	require.NoError(t, err)

	entity.SetDbProvider(p)

	defer func() {
		entity.SetDbProvider(c)
		_ = p.Close()
	}()

	// Cache hits would make the count depend on what earlier tests looked up.
	entity.FlushCaches()

	p.Counter.Start()
	fn()

	return p.Counter.Stop()
}

// writeStatementsJpeg writes a JPEG without metadata whose pixels depend on seed, so that no other test
// indexes the same file and the keywords derived from its name are always the same.
func writeStatementsJpeg(t *testing.T, fileName, seed string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 64, 48))

	for i := range img.Pix {
		img.Pix[i] = seed[i%len(seed)] + uint8(i) //nolint:gosec // Wraps by design.
	}

	f, err := os.Create(fileName) //nolint:gosec // G304: test-owned path.
	require.NoError(t, err)
	require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 90}))
	require.NoError(t, f.Close())
}

// deleteIndexedPhoto removes a photo created by indexing together with the rows derived from it.
func deleteIndexedPhoto(photoUID string) {
	if photoUID == "" {
		return
	}

	photo := entity.Photo{}

	if err := entity.UnscopedDb().Where("photo_uid = ?", photoUID).First(&photo).Error; err != nil {
		return
	}

	_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.Details{}).Error
	_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.PhotoKeyword{}).Error
	_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.PhotoLabel{}).Error
	_ = entity.UnscopedDb().Where("photo_id = ?", photo.ID).Delete(&entity.File{}).Error
	_ = entity.UnscopedDb().Where("id = ?", photo.ID).Delete(&entity.Photo{}).Error
}

// TestIndex_UserMediaFileStatements pins an upper bound for the statements indexing a single JPEG issues.
func TestIndex_UserMediaFileStatements(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := config.TestConfig()

	var added, updated int

	switch c.DatabaseDriver() {
	case dsn.DriverSQLite3:
		added, updated = 40, 44
	case dsn.DriverMySQL:
		added, updated = 40, 48
	default:
		t.Skipf("no statement ceiling for driver %s", c.DatabaseDriver())
	}

	const seed = "statements"
	dir := filepath.Join(c.OriginalsPath(), seed)
	require.NoError(t, fs.MkdirAll(dir))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	fileName := filepath.Join(dir, "statements.jpg")
	writeStatementsJpeg(t, fileName, seed)

	ind := NewIndex(c, NewConvert(c), NewFiles(), NewPhotos())
	opt := IndexOptionsSingle(c)
	opt.DetectFaces = false
	opt.DetectNsfw = false
	opt.DetectNSFWLabels = false
	opt.GenerateLabels = false

	// Remove a photo an interrupted run may have left at the same path.
	leftover := entity.Photo{}
	if entity.UnscopedDb().Where("photo_path = ? AND photo_name = ?", seed, "statements").First(&leftover).Error == nil {
		deleteIndexedPhoto(leftover.PhotoUID)
	}

	var result IndexResult

	t.Cleanup(func() { deleteIndexedPhoto(result.PhotoUID) })

	index := func() {
		m, mediaErr := NewMediaFile(fileName)
		require.NoError(t, mediaErr)
		result = ind.UserMediaFile(m, opt, filepath.Join(seed, "statements.jpg"), "", entity.Admin.GetUID())
	}

	statements := countedIndexStatements(t, c, index)
	require.Equal(t, IndexAdded, result.Status, "%s", result.Err)
	t.Logf("added: %d", len(statements))
	assert.LessOrEqual(t, len(statements), added, "%q", statements)

	statements = countedIndexStatements(t, c, index)
	require.Equal(t, IndexUpdated, result.Status, "%s", result.Err)
	t.Logf("updated: %d", len(statements))
	assert.LessOrEqual(t, len(statements), updated, "%q", statements)
}
