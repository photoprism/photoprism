package face

import "math"

func EuclidianDistance(face1 []float32, face2 []float32) float64 {
	var dist float64
	// TODO use more efficient implementation
	// either with TF or some go library, and batch processing
	for k := 0; k < 512; k++ {
		dist += math.Pow(float64(face1[k]-face2[k]), 2)
	}
	return math.Sqrt(dist)
}
