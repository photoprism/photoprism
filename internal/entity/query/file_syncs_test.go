package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFileSyncs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, err := FileSyncs(uint(1000001), "downloaded", 10)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileSync{}, r)
		}
	})
	t.Run("SearchForAllFileSyncs", func(t *testing.T) {
		r, err := FileSyncs(0, "", 10)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileSync{}, r)
		}
	})
	t.Run("FewestErrorsFirst", func(t *testing.T) {
		const serviceID = uint(1999999)
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", serviceID).Error)
		})
		// Inserted out of order, so neither sort key matches the row order.
		for _, c := range []struct {
			name string
			errs int
		}{{"/c.jpg", 0}, {"/a.jpg", 3}, {"/b.jpg", 0}, {"/d.jpg", 1}} {
			f := entity.NewFileSync(serviceID, c.name)
			f.Status, f.Errors = entity.FileSyncNew, c.errs
			if err := f.Create(); err != nil {
				t.Fatal(err)
			}
		}

		r, err := FileSyncs(serviceID, entity.FileSyncNew, 10)
		if err != nil {
			t.Fatal(err)
		}

		var names []string
		for _, f := range r {
			names = append(names, f.RemoteName)
		}
		assert.Equal(t, []string{"/b.jpg", "/c.jpg", "/d.jpg", "/a.jpg"}, names)

		// A row without an error count sorts with the files not tried yet.
		if err = UnscopedDb().Model(&entity.FileSync{}).Where("service_id = ? AND remote_name = ?", serviceID, "/a.jpg").
			UpdateColumn("errors", gorm.Expr("NULL")).Error; err != nil {
			t.Fatal(err)
		}

		r, err = FileSyncs(serviceID, entity.FileSyncNew, 10)
		if err != nil {
			t.Fatal(err)
		}

		names = names[:0]
		for _, f := range r {
			names = append(names, f.RemoteName)
		}
		assert.Equal(t, []string{"/a.jpg", "/b.jpg", "/c.jpg", "/d.jpg"}, names)
	})
}

func TestCountQueuedFileSyncs(t *testing.T) {
	t.Run("Counts", func(t *testing.T) {
		const serviceID, otherID = uint(1999998), uint(1999997)
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id IN (?)", []uint{serviceID, otherID}).Error)
		})
		for _, c := range []struct {
			service uint
			name    string
			status  string
			errs    int
		}{
			{serviceID, "/untried-a.jpg", entity.FileSyncNew, 0},
			{serviceID, "/untried-b.jpg", entity.FileSyncNew, 0},
			{serviceID, "/failed-a.jpg", entity.FileSyncNew, 1},
			{serviceID, "/failed-b.jpg", entity.FileSyncNew, 7},
			{serviceID, "/failed-c.jpg", entity.FileSyncNew, 2},
			{serviceID, "/done.jpg", entity.FileSyncDownloaded, 0},
			{serviceID, "/given-up.jpg", entity.FileSyncFailed, 4},
			{otherID, "/other.jpg", entity.FileSyncNew, 0},
		} {
			f := entity.NewFileSync(c.service, c.name)
			f.Status, f.Errors = c.status, c.errs
			if err := f.Create(); err != nil {
				t.Fatal(err)
			}
		}

		counts, err := CountQueuedFileSyncs(serviceID, -1)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, QueuedFileSyncCounts{Queued: 5, Untried: 2, Failed: 3, Retryable: 5, Errors: 10}, counts)

		// With a retry limit, only files at or below it may be tried again.
		counts, err = CountQueuedFileSyncs(serviceID, 2)
		if err != nil {
			t.Fatal(err)
		}

		assert.EqualValues(t, 4, counts.Retryable)
	})
	t.Run("NullErrors", func(t *testing.T) {
		// Rows written by earlier releases may hold no error count, which counts as not tried.
		const serviceID = uint(1999996)
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", serviceID).Error)
		})
		f := entity.NewFileSync(serviceID, "/legacy.jpg")
		if err := f.Create(); err != nil {
			t.Fatal(err)
		}
		if err := UnscopedDb().Model(&entity.FileSync{}).Where("service_id = ?", serviceID).UpdateColumn("errors", gorm.Expr("NULL")).Error; err != nil {
			t.Fatal(err)
		}

		counts, err := CountQueuedFileSyncs(serviceID, 3)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, QueuedFileSyncCounts{Queued: 1, Untried: 1, Retryable: 1}, counts)
	})
	t.Run("None", func(t *testing.T) {
		counts, err := CountQueuedFileSyncs(uint(1999995), -1)
		assert.NoError(t, err)
		assert.Equal(t, QueuedFileSyncCounts{}, counts)
	})
}
