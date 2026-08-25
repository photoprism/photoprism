package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFaces_Uncertainty(t *testing.T) {
	t.Run("HighestScoreDecides", func(t *testing.T) {
		f := Faces{Face{Score: 96}, Face{Score: 70}}
		assert.Equal(t, 1, f.Uncertainty())
	})
	t.Run("Confident", func(t *testing.T) {
		f := Faces{Face{Score: 91}, Face{Score: 91}}
		assert.Equal(t, 5, f.Uncertainty())
	})
	t.Run("Ordinary", func(t *testing.T) {
		f := Faces{Face{Score: 76}, Face{Score: 76}}
		assert.Equal(t, 20, f.Uncertainty())
	})
	t.Run("AtTheFloor", func(t *testing.T) {
		f := Faces{Face{Score: 50}, Face{Score: 50}}
		assert.Equal(t, 50, f.Uncertainty())
	})
	t.Run("NoFaces", func(t *testing.T) {
		assert.Equal(t, 100, Faces{}.Uncertainty())
	})
}

// TestScoreUncertainty pins that the confident end is reachable. The detectors cap a score at
// 100, so a scale whose steps start above that reports nothing better than its fourth step.
func TestScoreUncertainty(t *testing.T) {
	assert.Equal(t, 1, ScoreUncertainty(100))
	assert.Equal(t, 1, ScoreUncertainty(96))
	assert.Equal(t, 5, ScoreUncertainty(95))
	// The steps below the default detector's cutoff are reachable only through a detector that
	// scores lower, a marker no detector produced, or a migration, which detects beneath it on
	// purpose. Stated as the relationship rather than as the numbers either side, because the
	// cutoff is a calibration that moves.
	floor := FindDetector(DetectorYuNet).MinScore
	assert.Less(t, ScoreUncertainty(floor+1), ScoreUncertainty(floor),
		"a more confident face must be less uncertain")
	assert.Equal(t, 45, ScoreUncertainty(FindDetector(DetectorSCRFD).MinScore+1))
	assert.Equal(t, 50, ScoreUncertainty(0), "a marker no detector scored stays at the least certain step")
	assert.Equal(t, 50, ScoreUncertainty(9), "and so does one a migration found below every cutoff")
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
