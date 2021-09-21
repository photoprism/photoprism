package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFace_Size(t *testing.T) {
	t.Run("8", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  1,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, 8, f.Size())
	})
}

func TestFace_Dim(t *testing.T) {
	t.Run("3", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  3,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, float32(3), f.Dim())
	})
	t.Run("1", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  0,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, float32(1), f.Dim())
	})
}

func TestFace_EmbeddingsJSON(t *testing.T) {
	t.Run("no result", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  1,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, []byte{0x6e, 0x75, 0x6c, 0x6c}, f.EmbeddingsJSON())
	})
}

func TestFace_CropArea(t *testing.T) {
	t.Run("Position", func(t *testing.T) {
		f := Face{
			Cols:  1000,
			Rows:  600,
			Score: 125,
			Area: Area{
				Name:  "face",
				Col:   400,
				Row:   250,
				Scale: 200,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		t.Logf("marker: %#v", f.CropArea())
	})
}
