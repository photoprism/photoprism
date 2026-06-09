package vector

import "math"

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

// Norm returns the size of the vector (use pow = 2.0 for the Euclidean distance),
// see https://builtin.com/data-science/vector-norms. Absolute values are used so
// that odd powers (e.g. the L1 norm) stay well-defined for negative components.
func Norm(v Vector, pow float64) float64 {
	s := 0.0

	for _, f := range v {
		s += math.Pow(math.Abs(f), pow)
	}

	return math.Pow(s, 1/pow)
}
