package alg

import (
	"math"
	"testing"
)

func TestEuclideanDist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if d := EuclideanDist([]float64{0, 0}, []float64{3, 4}); d != 5 {
			t.Errorf("expected 5, got %f", d)
		}
	})
	t.Run("LongerFirst", func(t *testing.T) {
		if d := EuclideanDist([]float64{1, 2, 3}, []float64{1}); !math.IsNaN(d) {
			t.Errorf("expected NaN, got %f", d)
		}
	})
	t.Run("ShorterFirst", func(t *testing.T) {
		// The reverse order used to return a plausible truncated distance instead of
		// panicking, so both directions are pinned.
		if d := EuclideanDist([]float64{1}, []float64{1, 2, 3}); !math.IsNaN(d) {
			t.Errorf("expected NaN, got %f", d)
		}
	})
}

func TestEuclideanDistSquared(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if d := EuclideanDistSquared([]float64{0, 0}, []float64{3, 4}); d != 25 {
			t.Errorf("expected 25, got %f", d)
		}
	})
	t.Run("LongerFirst", func(t *testing.T) {
		if d := EuclideanDistSquared([]float64{1, 2, 3}, []float64{1}); !math.IsNaN(d) {
			t.Errorf("expected NaN, got %f", d)
		}
	})
	t.Run("ShorterFirst", func(t *testing.T) {
		if d := EuclideanDistSquared([]float64{1}, []float64{1, 2, 3}); !math.IsNaN(d) {
			t.Errorf("expected NaN, got %f", d)
		}
	})
}
