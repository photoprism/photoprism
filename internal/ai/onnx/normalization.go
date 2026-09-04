package onnx

// Normalization describes the preprocessing a model expects as (channel - Mean) / StdDev, applied
// per channel to values in the 0-255 range.
//
// Per channel rather than scalar, because ImageNet-pretrained classifiers use a different mean and
// standard deviation for each one, which a single value cannot express.
type Normalization struct {
	Mean   [Channels]float32 `yaml:"Mean,omitempty" json:"mean,omitempty"`
	StdDev [Channels]float32 `yaml:"StdDev,omitempty" json:"stdDev,omitempty"`
}

// Uniform returns a normalization that applies the same mean and standard deviation to
// every channel.
func Uniform(mean, stdDev float32) Normalization {
	return Normalization{
		Mean:   [Channels]float32{mean, mean, mean},
		StdDev: [Channels]float32{stdDev, stdDev, stdDev},
	}
}

// Scales returns the reciprocal of the standard deviation per channel, so that the inner
// preprocessing loop multiplies instead of dividing. Channels without a standard
// deviation stay unscaled.
func (n Normalization) Scales() [Channels]float32 {
	var scales [Channels]float32

	for i, stdDev := range n.StdDev {
		if stdDev == 0 {
			scales[i] = 1
		} else {
			scales[i] = 1 / stdDev
		}
	}

	return scales
}

// IsZero reports whether no normalization was specified.
func (n Normalization) IsZero() bool {
	return n == Normalization{}
}
