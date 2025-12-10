package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// VisionYaml returns the path to the computer-vision configuration file,
// preferring an explicit override and otherwise letting fs.ConfigFilePath pick
// the right `.yml`/`.yaml` variant in the config directory.
func (c *Config) VisionYaml() string {
	if c == nil {
		return ""
	}

	if c.options.VisionYaml != "" {
		return fs.Abs(c.options.VisionYaml)
	} else {
		return fs.ConfigFilePath(c.ConfigPath(), "vision", fs.ExtYml)
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

// VisionModelShouldRun reports whether the configured vision model of the
// specified type should execute in a given scheduling context. Face detection
// delegates to FaceEngineShouldRun so detection and embedding stay aligned.
func (c *Config) VisionModelShouldRun(t vision.ModelType, when vision.RunType) bool {
	if c == nil {
		return false
	}

	if t == vision.ModelTypeFace && c.DisableFaces() {
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

	if t == vision.ModelTypeFace {
		return c.FaceEngineShouldRun(when)
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
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path derived from config directory
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
