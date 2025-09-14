package config

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VisionYaml returns the vision config YAML filename.
//
// TODO: Call fs.YamlFilePath to use ".yaml" extension for new YAML files, unless a .yml" file already exists.
//
//	return fs.YamlFilePath("vision", c.ConfigPath(), c.options.VisionYaml)
func (c *Config) VisionYaml() string {
	if c.options.VisionYaml != "" {
		return fs.Abs(c.options.VisionYaml)
	} else {
		return filepath.Join(c.ConfigPath(), "vision.yml")
	}
}

// VisionApi checks whether the Computer Vision API endpoints should be enabled.
func (c *Config) VisionApi() bool {
	return c.options.VisionApi && !c.options.Demo
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

// ModelsPath returns the path where the machine learning models are located.
func (c *Config) ModelsPath() string {
	if c.options.ModelsPath != "" {
		return fs.Abs(c.options.ModelsPath)
	}

	if dir := filepath.Join(c.AssetsPath(), fs.ModelsDir); fs.PathExists(dir) {
		c.options.ModelsPath = dir
		return c.options.ModelsPath
	}

	c.options.ModelsPath = fs.FindDir(fs.ModelsPaths)

	return c.options.ModelsPath
}

// NasnetModelPath returns the TensorFlow model path.
func (c *Config) NasnetModelPath() string {
	return filepath.Join(c.ModelsPath(), "nasnet")
}

// FacenetModelPath returns the FaceNet model path.
func (c *Config) FacenetModelPath() string {
	return filepath.Join(c.ModelsPath(), "facenet")
}

// NsfwModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NsfwModelPath() string {
	return filepath.Join(c.ModelsPath(), "nsfw")
}

// DetectNSFW checks if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	return c.options.DetectNSFW
}
