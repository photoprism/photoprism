package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VisionYaml returns the vision config YAML filename.
//
// TODO: Call fs.YamlFilePath to use ".yaml" extension for new YAML files, unless a .yml" file already exists.
//
//	return fs.YamlFilePath("vision", c.ConfigPath(), c.options.VisionYaml)
func (c *Config) VisionYaml() string {
	if c == nil {
		return ""
	}

	if c.options.VisionYaml != "" {
		return fs.Abs(c.options.VisionYaml)
	} else {
		return filepath.Join(c.ConfigPath(), "vision.yml")
	}
}

// VisionSchedule returns the cron schedule configured for the vision worker, or "" if disabled.
func (c *Config) VisionSchedule() string {
	if c == nil {
		return ""
	}

	return Schedule(c.options.VisionSchedule)
}

// VisionFilter returns the search filter to use for scheduled vision runs.
func (c *Config) VisionFilter() string {
	if c == nil {
		return ""
	}

	return strings.TrimSpace(c.options.VisionFilter)
}

// VisionModelShouldRun checks when the specified model type should run.
func (c *Config) VisionModelShouldRun(t vision.ModelType, when vision.RunType) bool {
	if c == nil {
		return false
	}

	if t == vision.ModelTypeLabels && c.DisableClassification() {
		return false
	}

	if t == vision.ModelTypeNsfw && !c.DetectNSFW() {
		return false
	}

	if vision.Config == nil {
		return false
	}

	return vision.Config.ShouldRun(t, when)
}

// VisionApi checks whether the Computer Vision API endpoints should be enabled.
func (c *Config) VisionApi() bool {
	if c == nil {
		return false
	}

	return c.options.VisionApi && !c.options.Demo
}

// VisionUri returns the remote computer vision service URI, e.g. https://example.com/api/v1/vision.
func (c *Config) VisionUri() string {
	if c == nil {
		return ""
	}

	return clean.Uri(c.options.VisionUri)
}

// VisionKey returns the remote computer vision service access token, if any.
func (c *Config) VisionKey() string {
	if c == nil {
		return ""
	}

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
	if c == nil {
		return ""
	}

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
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "nasnet")
}

// FacenetModelPath returns the FaceNet model path.
func (c *Config) FacenetModelPath() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "facenet")
}

// NsfwModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NsfwModelPath() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "nsfw")
}

// DetectNSFW checks if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	if c == nil {
		return false
	}

	return c.options.DetectNSFW
}
