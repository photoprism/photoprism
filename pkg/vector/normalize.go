package vector

import "math"

// Normalize scales the vector to unit length (L2 norm) in place.
// A zero vector (including an empty one) is left unchanged to avoid
// a division by zero.
func (v Vector) Normalize() {
	var sum float64

	for _, f := range v {
		sum += f * f
	}

	if sum == 0 {
		return
	}

	inv := 1 / math.Sqrt(sum)

	for i := range v {
		v[i] *= inv
	}
}

// Normalized returns an L2-normalized copy of the vector,
// leaving the receiver unchanged.
func (v Vector) Normalized() Vector {
	c := v.Copy()
	c.Normalize()
	return c
}
