package vector

import "math"

// Sd calculates the vector's standard deviation.
func (v Vector) Sd() float64 {
	return math.Sqrt(v.Variance())
}

// Variance calculates the vector's variance.
func (v Vector) Variance() float64 {
	return v.variance(v.Mean())
}

// variance returns the sample variance around the given mean.
// Empty and single-element vectors have zero variance by convention,
// which also avoids a division by zero in the n-1 denominator.
func (v Vector) variance(mean float64) float64 {
	n := float64(len(v))

	if n < 2 {
		return 0
	}

	ss := 0.0

	for _, f := range v {
		d := f - mean
		ss += d * d
	}

	return ss / (n - 1)
}

// Cor returns the Pearson correlation between two vectors.
func Cor(a, b Vector) (float64, error) {
	n := float64(len(a))
	xy, err := Product(a, b)

	if err != nil {
		return NaN(), err
	}

	sx := a.Sd()
	sy := b.Sd()

	mx := a.Mean()
	my := b.Mean()

	r := (xy.Sum() - n*mx*my) / ((n - 1) * sx * sy)

	return r, nil
}
