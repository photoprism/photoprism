package config

import tf "github.com/tensorflow/tensorflow/tensorflow/go"

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return tf.Version()
}

// DisableTensorFlow returns true if TensorFlow should not be used for image classification.
func (c *Config) DisableTensorFlow() bool {
	return c.params.DisableTensorFlow
}
