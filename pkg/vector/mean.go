package vector

import (
	"fmt"
	"math"
)

// Mean gets the average of a slice of numbers
func Mean(v Vector) float64 {
	s := v.Sum()

	n := float64(len(v))

	return s / n
}

// GeometricMean gets the geometric mean for a slice of numbers.
// It is undefined for negative values (returns NaN) and yields 0 if any
// value is 0, since the product of all values is then 0.
func GeometricMean(v Vector) float64 {
	l := v.Dim()

	if l == 0 {
		return NaN()
	}

	// Multiply all values; a zero element correctly drives the product to 0.
	p := 1.0
	for _, n := range v {
		if n < 0 {
			return NaN()
		}
		p *= n
	}

	// Calculate the geometric mean.
	return math.Pow(p, 1/float64(l))
}

// HarmonicMean gets the harmonic mean for a slice of numbers
func HarmonicMean(v Vector) float64 {
	l := v.Dim()

	if l == 0 {
		return NaN()
	}

	// Get the sum of all the numbers reciprocals and return an
	// error for values that cannot be included in harmonic mean
	var p float64
	for _, n := range v {
		if n < 0 {
			return NaN()
		} else if n == 0 {
			return NaN()
		}
		p += 1 / n
	}

	return float64(l) / p
}

// Mean returns the vector's mean value.
func (v Vector) Mean() float64 {
	return Mean(v)
}

// GeometricMean returns the vector's geometric mean value.
func (v Vector) GeometricMean() float64 {
	return GeometricMean(v)
}

// HarmonicMean returns the vector's harmonic mean value.
func (v Vector) HarmonicMean() float64 {
	return HarmonicMean(v)
}

// weightedSum returns the weighted sum of the vector.  This is really only useful in
// calculating the weighted mean.
func (v Vector) weightedSum(w Vector) (float64, error) {
	if len(v) != len(w) {
		return NaN(), fmt.Errorf("length of weights unequal to vector length")
	}

	ws := 0.0

	for i := range v {
		ws += v[i] * w[i]
	}

	return ws, nil
}

// WeightedMean returns the vector's weighted mean value based of the specified weights.
func (v Vector) WeightedMean(w Vector) (float64, error) {
	ws, err := v.weightedSum(w)

	if err != nil {
		return NaN(), err
	}

	sw := w.Sum()

	return ws / sw, nil
}
