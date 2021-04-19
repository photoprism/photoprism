package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_IdenticalIdentical(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 1, len(result))
	})

	t.Run("unstacked photo", func(t *testing.T) {
		photo := &Photo{PhotoStack: IsUnstacked, PhotoName: "testName"}

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(result))
	})
	t.Run("success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		result, err := photo.Identical(true, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
}

func TestPhoto_Merge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		original, merged, err := photo.Merge(true, false)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1000023, int(original.ID))
		assert.Equal(t, 1000024, int(merged[0].ID))
	})
}
