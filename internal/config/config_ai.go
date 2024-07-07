package config

import (
	"path/filepath"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return tf.Version()
}

// TensorFlowModelPath returns the TensorFlow model path.
func (c *Config) TensorFlowModelPath() string {
	return filepath.Join(c.AssetsPath(), "nasnet")
}

// FaceNetModelPath returns the FaceNet model path.
func (c *Config) FaceNetModelPath() string {
	return filepath.Join(c.AssetsPath(), "facenet")
}

// NSFWModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NSFWModelPath() string {
	return filepath.Join(c.AssetsPath(), "nsfw")
}

// DetectNSFW checks if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	return c.options.DetectNSFW
}

// UploadNSFW checks if NSFW photos can be uploaded.
func (c *Config) UploadNSFW() bool {
	return c.options.UploadNSFW
}
