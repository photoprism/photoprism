package vector

// Centroid returns the element-wise mean (centroid) of the given vectors as a
// new, independent vector. Vectors whose length differs from the first vector
// are ignored, and the mean is taken over the vectors actually included. It
// returns nil when vs is empty or the first vector has no elements.
func Centroid(vs Vectors) Vector {
	if len(vs) == 0 {
		return nil
	}

	dim := len(vs[0])

	if dim == 0 {
		return nil
	}

	result := make(Vector, dim)
	n := 0

	for _, v := range vs {
		if len(v) != dim {
			continue
		}

		for j := range dim {
			result[j] += v[j]
		}

		n++
	}

	inv := 1 / float64(n)

	for j := range result {
		result[j] *= inv
	}

	return result
}
