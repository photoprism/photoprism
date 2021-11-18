package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_Optimize(t *testing.T) {
	t.Run("update", func(t *testing.T) {
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
	})
	t.Run("photo without id", func(t *testing.T) {
		photo := Photo{}
		result, merged, err := photo.Optimize(false, false, true, false)
		assert.Error(t, err)
		assert.False(t, result)

		if len(merged) > 0 {
			t.Error("no photos should be merged")
		}
	})
}
