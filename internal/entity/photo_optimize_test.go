package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_Optimize(t *testing.T) {
	t.Run("Update", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")

		if updated, merged, err := photo.Optimize(false, false, true, false); err != nil {
			t.Fatal(err)
		} else if !updated {
			t.Error("photo should be updated")
		} else if len(merged) > 0 {
			t.Error("no photos should be merged")
		}

		if updated, merged, err := photo.Optimize(false, false, true, false); err != nil {
			t.Fatal(err)
		} else if updated {
			t.Errorf("photo should NOT be updated, merged: %+v", merged)
		} else if len(merged) > 0 {
			t.Errorf("no photos should be merged")
		}
		t.Cleanup(func() {
			pf := PhotoFixtures.Get("Photo19")
			assert.NoError(t, UnscopedDb().Delete(&Details{}, "photo_id = ?", pf.ID).Error)
			assert.NoError(t, UnscopedDb().Delete(&PhotoKeyword{}, "keyword_id in (select id from keywords where keyword in ('jens','mander','nature','photo19', 'bridge')) and photo_id = ?", pf.ID).Error)
			assert.NoError(t, UnscopedDb().Delete(&Keyword{}, "keyword in ('jens','mander','nature','photo19')").Error)
			assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", pf.ID).UpdateColumns(Values{"photo_title": pf.PhotoTitle, "photo_quality": pf.PhotoQuality, "indexed_at": pf.IndexedAt, "checked_at": pf.CheckedAt, "estimated_at": pf.EstimatedAt}).Error)
		})
	})
	t.Run("PhotoWithoutId", func(t *testing.T) {
		photo := Photo{}
		result, merged, err := photo.Optimize(false, false, true, false)
		assert.Error(t, err)
		assert.False(t, result)

		if len(merged) > 0 {
			t.Error("no photos should be merged")
		}
	})
}
