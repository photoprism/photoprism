package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelValues(t *testing.T) {
	t.Run("NoInterface", func(t *testing.T) {
		m := Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		assert.Error(t, err)
		assert.IsType(t, Map{}, values)
		assert.Len(t, keys, 0)
	})
	t.Run("NewPhoto", func(t *testing.T) {
		m := &Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, Map{}, values)
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 2)
		assert.NotNil(t, values)
		assert.IsType(t, Map{}, values)
	})
	t.Run("NewFace", func(t *testing.T) {
		m := &Face{}
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, Map{}, values)
	})
	t.Run("ExistingFace", func(t *testing.T) {
		m := FaceFixtures.Pointer("john-doe")
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, Map{}, values)
	})
}
