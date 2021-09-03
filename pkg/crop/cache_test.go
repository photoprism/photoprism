package crop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	t.Run("Crop160", func(t *testing.T) {
		result, err := FileName("147da9f0261e2d81e9a52b266f1945556588bb78", "/example", 160, 160, "042008007010")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "/example/1/4/7/147da9f0261e2d81e9a52b266f1945556588bb78_160x160_crop_042008007010.jpg", result)
	})
	t.Run("InvalidSize", func(t *testing.T) {
		result, err := FileName("147da9f0261e2d81e9a52b266f1945556588bb78", "/example", 15000, 160, "042008007010")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "crop: invalid size 15000x160", err.Error())
		assert.Empty(t, result)
	})
	t.Run("InvalidHash", func(t *testing.T) {
		result, err := FileName("147", "/example", 160, 160, "042008007010")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "crop: invalid file hash 147", err.Error())
		assert.Empty(t, result)
	})
	t.Run("InvalidPath", func(t *testing.T) {
		result, err := FileName("147da9f0261e2d81e9a52b266f1945556588bb78", "", 160, 160, "042008007010")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Equal(t, "crop: cache path missing", err.Error())
		assert.Empty(t, result)
	})
}
