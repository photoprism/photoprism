package vector

import "fmt"

// Product returns a vector of element-wise products of two input vectors.
func Product(a, b Vector) (Vector, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("vector dimensions do not match (%d, %d)", len(a), len(b))
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
		return NaN(), err
	}

	return p.Sum(), nil
}
