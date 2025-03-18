package vector

import "math"

// Mean gets the average of a slice of numbers
func Mean(v Vector) float64 {
	s := v.Sum()

	n := float64(len(v))

	return s / n
}

// GeometricMean gets the geometric mean for a slice of numbers
func GeometricMean(v Vector) float64 {
	l := v.Dim()

	if l == 0 {
		return NaN()
	}

	// Get the product of all the numbers
	var p float64
	for _, n := range v {
		if p == 0 {
			p = n
		} else {
			p *= n
		}
	}

	// Calculate the geometric mean
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
