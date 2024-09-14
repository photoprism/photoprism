package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaceMap_Get(t *testing.T) {
	t.Run("get existing face", func(t *testing.T) {
		r := FaceFixtures.Get("jane-doe")
		assert.Equal(t, "js6sg6b1h1njaaab", r.SubjUID)
		assert.Equal(t, "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7", r.ID)
		assert.IsType(t, Face{}, r)
	})
	t.Run("get not existing location", func(t *testing.T) {
		r := FaceFixtures.Get("xxx")
		assert.Equal(t, UnknownID, r.ID)
		assert.IsType(t, Face{}, r)
	})
}

func TestFaceMap_Pointer(t *testing.T) {
	t.Run("get existing face", func(t *testing.T) {
		r := FaceFixtures.Pointer("jane-doe")
		assert.Equal(t, "js6sg6b1h1njaaab", r.SubjUID)
		assert.Equal(t, "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7", r.ID)
		assert.IsType(t, &Face{}, r)
	})
	t.Run("get not existing location", func(t *testing.T) {
		r := FaceFixtures.Pointer("xxx")
		assert.Equal(t, UnknownID, r.ID)
		assert.IsType(t, &Face{}, r)
	})
}
