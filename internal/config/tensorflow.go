package config

import tf "github.com/tensorflow/tensorflow/tensorflow/go"

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return tf.Version()
}

// TensorFlowDisabled returns true if the use of TensorFlow is disabled for image classification.
func (c *Config) TensorFlowDisabled() bool {
	return c.config.DisableTensorFlow
}
