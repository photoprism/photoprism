package vector

import (
	"fmt"
	"math"
)

// Copy returns a copy of the vector.
func (v Vector) Copy() Vector {
	y := make(Vector, len(v))
	copy(y, v)
	return y
}

// Dim returns the number of values (dimension).
func (v Vector) Dim() int {
	return len(v)
}

// Sum returns the sum of the vector values.
func (v Vector) Sum() float64 {
	s := 0.0

	for _, f := range v {
		s += f
	}

	return s
}

// weightedSum returns the weighted sum of the vector.  This is really only useful in
// calculating the weighted mean.
func (v Vector) weightedSum(w Vector) (float64, error) {
	if len(v) != len(w) {
		return Epsilon, fmt.Errorf("Length of weights unequal to vector length")
	}

	ws := 0.0

	for i := range v {
		ws += v[i] * w[i]
	}

	return ws, nil
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

// WeightedMean returns the vector's weighted mean value based of the specified weights.
func (v Vector) WeightedMean(w Vector) (float64, error) {
	ws, err := v.weightedSum(w)

	if err != nil {
		return Epsilon, err
	}

	sw := w.Sum()

	return ws / sw, nil
}

// Sd calculates the vector's standard deviation.
func (v Vector) Sd() float64 {
	return math.Sqrt(v.Variance())
}

// Variance calculates the vector's variance.
func (v Vector) Variance() float64 {
	return v.variance(v.Mean())
}

// EuclideanDist returns the Euclidean distance between the vectors,
func (v Vector) EuclideanDist(w Vector) float64 {
	return EuclideanDist(v, w)
}

// CosineDist returns the cosine distance between two vectors.
func (v Vector) CosineDist(w Vector) float64 {
	return CosineDist(v, w)
}

// Norm returns the vector size (magnitude),
// see https://builtin.com/data-science/vector-norms.
func (v Vector) Norm(pow float64) float64 {
	return Norm(v, pow)
}

// EuclideanNorm returns the Euclidean vector size (magnitude),
// see https://builtin.com/data-science/vector-norms.
func (v Vector) EuclideanNorm() float64 {
	return v.Norm(2.0)
}

func (v Vector) variance(mean float64) float64 {
	n := float64(len(v))

	if n == 1 {
		return 0
	} else if n < 2 {
		n = 2
	}

	ss := 0.0

	for _, f := range v {
		ss += math.Pow(f-mean, 2.0)
	}

	return ss / (n - 1)
}

// Product returns a vector of element-wise products of two input vectors.
func Product(a, b Vector) (Vector, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("vector dimentions do not match (%d, %d)", len(a), len(b))
	}

	p := make(Vector, len(a))

	for i := range a {
		p[i] = a[i] * b[i]
	}

	return p, nil
}

// DotProduct returns the dot product of two vectors.
func DotProduct(a, b Vector) (float64, error) {
	p, err := Product(a, b)

	if err != nil {
		return Epsilon, err
	}

	return p.Sum(), nil
}

// Norm returns the size of the vector (use pow = 2.0 for the Euclidean distance),
// see https://builtin.com/data-science/vector-norms.
func Norm(v Vector, pow float64) float64 {
	s := 0.0

	for _, f := range v {
		s += math.Pow(f, pow)
	}

	return math.Pow(s, 1/pow)
}

// EuclideanDist returns the Euclidean distance between multiple vectors.
func EuclideanDist(a, b Vector) float64 {
	if a.Dim() != b.Dim() {
		return NaN()
	}

	var (
		s, t float64
	)

	for i := range a {
		t = a[i] - b[i]
		s += t * t
	}

	return math.Sqrt(s)
}

// CosineDist returns the CosineDist distance between two vectors.
func CosineDist(a, b Vector) float64 {
	if a.Dim() != b.Dim() {
		return NaN()
	}

	var sum, s1, s2 float64

	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
		s1 += math.Pow(a[i], 2)
		s2 += math.Pow(b[i], 2)
	}

	if s1 == 0 || s2 == 0 {
		return 0.0
	}

	return sum / (math.Sqrt(s1) * math.Sqrt(s2))
}

// CosineDists returns the cosine distances between two sets of vectors.
func CosineDists(x, y Vectors) Vectors {
	result := make(Vectors, len(x))

	for i, a := range x {
		result[i] = make([]float64, len(y))

		for j, b := range y {
			result[i][j] = CosineDist(a, b)
		}
	}

	return result
}

// Cor returns the Pearson correlation between two vectors.
func Cor(a, b Vector) (float64, error) {
	n := float64(len(a))
	xy, err := Product(a, b)

	if err != nil {
		return Epsilon, err
	}

	sx := a.Sd()
	sy := b.Sd()

	mx := a.Mean()
	my := b.Mean()

	r := (xy.Sum() - n*mx*my) / ((n - 1) * sx * sy)

	return r, nil
}
