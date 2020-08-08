// +build NOTENSORFLOW

package config

import (
	"path/filepath"
)

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return "NONE"
}

// TensorFlowOff returns true if TensorFlow should NOT be used for image classification (or anything else).
func (c *Config) TensorFlowOff() bool {
	return true
}

// TensorFlowModelPath returns the TensorFlow model path.
func (c *Config) TensorFlowModelPath() string {
	return filepath.Join(c.AssetsPath(), "nasnet")
}

// NSFWModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NSFWModelPath() string {
	return filepath.Join(c.AssetsPath(), "nsfw")
}
