package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoMap_Get(t *testing.T) {
	t.Run("get existing photo", func(t *testing.T) {
		r := PhotoFixtures.Get("19800101_000002_D640C559")
		assert.Equal(t, "ps6sg6be2lvl0yh7", r.PhotoUID)
		assert.Equal(t, "27900704_070228_D6D51B6C", r.PhotoName)
		assert.IsType(t, Photo{}, r)
	})
	t.Run("get not existing photo", func(t *testing.T) {
		r := PhotoFixtures.Get("TestName")
		assert.Equal(t, "TestName", r.PhotoName)
		assert.IsType(t, Photo{}, r)
	})
}

func TestPhotoMap_Pointer(t *testing.T) {
	t.Run("get existing photo pointer", func(t *testing.T) {
		r := PhotoFixtures.Pointer("19800101_000002_D640C559")
		assert.Equal(t, "ps6sg6be2lvl0yh7", r.PhotoUID)
		assert.Equal(t, "27900704_070228_D6D51B6C", r.PhotoName)
		assert.IsType(t, &Photo{}, r)
	})
	t.Run("get not existing photo pointer", func(t *testing.T) {
		r := PhotoFixtures.Pointer("TestName2")
		assert.Equal(t, "TestName2", r.PhotoName)
		assert.IsType(t, &Photo{}, r)
	})
}
