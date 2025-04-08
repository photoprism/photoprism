package config

import (
	"os"
	"path/filepath"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VisionYaml returns the vision config YAML filename.
func (c *Config) VisionYaml() string {
	if c.options.VisionYaml != "" {
		return fs.Abs(c.options.VisionYaml)
	} else {
		return filepath.Join(c.ConfigPath(), "vision.yml")
	}
}

// VisionApi checks whether the Computer Vision API endpoints should be enabled.
func (c *Config) VisionApi() bool {
	return c.options.VisionApi
}

// VisionUri returns the remote computer vision service URI, e.g. https://example.com/api/v1/vision.
func (c *Config) VisionUri() string {
	return clean.Uri(c.options.VisionUri)
}

// VisionKey returns the remote computer vision service access token, if any.
func (c *Config) VisionKey() string {
	// Try to read access token from file if c.options.VisionKey is not set.
	if c.options.VisionKey != "" {
		return clean.Password(c.options.VisionKey)
	} else if fileName := FlagFilePath("VISION_KEY"); fileName == "" {
		// No access token set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read vision key from %s (%s)", fileName, err)
		return ""
	} else {
		return clean.Password(string(b))
	}
}

// TensorFlowVersion returns the TenorFlow framework version.
func (c *Config) TensorFlowVersion() string {
	return tf.Version()
}

// NasnetModelPath returns the TensorFlow model path.
func (c *Config) NasnetModelPath() string {
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
