package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileMap_Get(t *testing.T) {
	t.Run("GetExistingFile", func(t *testing.T) {
		r := FileFixtures.Get("exampleFileName.jpg")
		assert.Equal(t, "fs6sg6bw45bnlqdw", r.FileUID)
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", r.FileName)
		assert.IsType(t, File{}, r)
	})
	t.Run("GetNotExistingFile", func(t *testing.T) {
		r := FileFixtures.Get("TestName")
		assert.Equal(t, "TestName", r.FileName)
		assert.IsType(t, File{}, r)
	})
}

func TestFileMap_Pointer(t *testing.T) {
	t.Run("GetExistingFile", func(t *testing.T) {
		r := FileFixtures.Pointer("exampleFileName.jpg")
		assert.Equal(t, "fs6sg6bw45bnlqdw", r.FileUID)
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", r.FileName)
		assert.IsType(t, &File{}, r)
	})
	t.Run("GetNotExistingFile", func(t *testing.T) {
		r := FileFixtures.Pointer("TestName")
		assert.Equal(t, "TestName", r.FileName)
		assert.IsType(t, &File{}, r)
	})
}
