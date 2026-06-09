package vector

import "math"

// EuclideanDist returns the Euclidean distance between the vectors,
func (v Vector) EuclideanDist(w Vector) float64 {
	return EuclideanDist(v, w)
}

// CosineSimilarity returns the cosine similarity between two vectors,
// ranging from -1 (opposite) to 1 (identical).
func (v Vector) CosineSimilarity(w Vector) float64 {
	return CosineSimilarity(v, w)
}

// CosineDist returns the cosine distance between two vectors (1 - cosine similarity).
func (v Vector) CosineDist(w Vector) float64 {
	return CosineDist(v, w)
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

// CosineSimilarity returns the cosine similarity between two vectors, ranging
// from -1 (opposite) to 1 (identical). It returns NaN when the dimensions
// differ and 0 when either operand is a zero vector.
func CosineSimilarity(a, b Vector) float64 {
	if a.Dim() != b.Dim() {
		return NaN()
	}

	var sum, s1, s2 float64

	for i := range a {
		sum += a[i] * b[i]
		s1 += a[i] * a[i]
		s2 += b[i] * b[i]
	}

	if s1 == 0 || s2 == 0 {
		return 0.0
	}

	return sum / (math.Sqrt(s1) * math.Sqrt(s2))
}

// CosineDist returns the cosine distance between two vectors, defined as
// 1 - CosineSimilarity. Identical vectors yield 0; it returns NaN when the
// dimensions differ.
func CosineDist(a, b Vector) float64 {
	return 1.0 - CosineSimilarity(a, b)
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
