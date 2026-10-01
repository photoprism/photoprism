package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	})
}
