package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaces_Uncertainty(t *testing.T) {
	t.Run("maxScore = 310", func(t *testing.T) {
		f := Faces{Face{Score: 310}, Face{Score: 210}}
		assert.Equal(t, 1, f.Uncertainty())
	})
	t.Run("maxScore = 210", func(t *testing.T) {
		f := Faces{Face{Score: 210}, Face{Score: 210}}
		assert.Equal(t, 5, f.Uncertainty())
	})
	t.Run("maxScore = 66", func(t *testing.T) {
		f := Faces{Face{Score: 66}, Face{Score: 66}}
		assert.Equal(t, 20, f.Uncertainty())
	})
	t.Run("maxScore = 10", func(t *testing.T) {
		f := Faces{Face{Score: 10}, Face{Score: 10}}
		assert.Equal(t, 50, f.Uncertainty())
	})
}

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

func TestFaces_Contains(t *testing.T) {
	t.Run("Contained", func(t *testing.T) {
		a := Face{
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

		b := Face{
			Cols:  1000,
			Rows:  600,
			Score: 34,
			Area: Area{
				Name:  "face",
				Col:   100,
				Row:   100,
				Scale: 50,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}

		c := Face{
			Cols:  1000,
			Rows:  600,
			Score: 125,
			Area: Area{
				Name:  "face",
				Col:   125,
				Row:   125,
				Scale: 25,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}

		d := Face{
			Cols:  1000,
			Rows:  600,
			Score: 125,
			Area: Area{
				Name:  "face",
				Col:   110,
				Row:   110,
				Scale: 50,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}

		faces := Faces{a, b}

		assert.True(t, faces.Contains(a))
		assert.True(t, faces.Contains(b))
		assert.False(t, faces.Contains(c))
		assert.True(t, faces.Contains(d))
	})
	t.Run("NotContained", func(t *testing.T) {
		a := Face{
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

		b := Face{
			Cols:  1000,
			Rows:  600,
			Score: 34,
			Area: Area{
				Name:  "face",
				Col:   100,
				Row:   100,
				Scale: 50,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}

		c := Face{
			Cols:  1000,
			Rows:  600,
			Score: 125,
			Area: Area{
				Name:  "face",
				Col:   900,
				Row:   500,
				Scale: 25,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}

		faces := Faces{a}

		assert.False(t, faces.Contains(b))
		assert.False(t, faces.Contains(c))
	})
}
