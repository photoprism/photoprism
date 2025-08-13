package vision

import (
	"fmt"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

var modelMutex = sync.Mutex{}

// Default model version strings.
var (
	VersionLatest = "latest"
	VersionMobile = "mobile"
)

// Model represents a computer vision model configuration.
type Model struct {
	Type          ModelType             `yaml:"Type,omitempty" json:"type,omitempty"`
	Default       bool                  `yaml:"Default,omitempty" json:"default,omitempty"`
	Name          string                `yaml:"Name,omitempty" json:"name,omitempty"`
	Version       string                `yaml:"Version,omitempty" json:"version,omitempty"`
	System        string                `yaml:"System,omitempty" json:"system,omitempty"`
	Prompt        string                `yaml:"Prompt,omitempty" json:"prompt,omitempty"`
	Resolution    int                   `yaml:"Resolution,omitempty" json:"resolution,omitempty"`
	Meta          *tensorflow.ModelInfo `yaml:"Meta,omitempty" json:"meta,omitempty"`
	Service       Service               `yaml:"Service,omitempty" json:"Service,omitempty"`
	Path          string                `yaml:"Path,omitempty" json:"-"`
	Disabled      bool                  `yaml:"Disabled,omitempty" json:"disabled,omitempty"`
	classifyModel *classify.Model
	faceModel     *face.Model
	nsfwModel     *nsfw.Model
}

// Models represents a set of computer vision models.
type Models []*Model

// Model returns the parsed and normalized model identifier, name, and version strings.
func (m *Model) Model() (model, name, version string) {
	// Return empty identifier string if no name was set.
	if m.Name == "" {
		return "", "", clean.TypeLowerDash(m.Version)
	}

	// Normalize model name.
	name = clean.TypeLowerDash(m.Name)

	// Split name to check if it contains the version.
	s := strings.SplitN(name, ":", 2)

	// Return if name contains both model name and version.
	if len(s) == 2 && s[0] != "" && s[1] != "" {
		return name, s[0], s[1]
	}

	// Normalize model version.
	version = clean.TypeLowerDash(m.Version)

	// Default to "latest" if no specific version was set.
	if version == "" {
		version = VersionLatest
	}

	// Create model identifier from model name and version.
	model = strings.Join([]string{s[0], version}, ":")

	// Return normalized model identifier, name, and version.
	return model, name, version
}

// Endpoint returns the remote service request method and endpoint URL, if any.
func (m *Model) Endpoint() (uri, method string) {
	if uri, method = m.Service.Endpoint(); uri != "" && method != "" {
		return uri, method
	} else if ServiceUri == "" {
		return "", ""
	} else if serviceType := clean.TypeLowerUnderscore(m.Type); serviceType == "" {
		return "", ""
	} else {
		return fmt.Sprintf("%s/%s", ServiceUri, serviceType), ServiceMethod
	}
}

// EndpointKey returns the access token belonging to the remote service endpoint, if any.
func (m *Model) EndpointKey() (key string) {
	if key = m.Service.EndpointKey(); key != "" {
		return key
	} else {
		return ServiceKey
	}
}

// EndpointFileScheme returns the endpoint API request file scheme type.
func (m *Model) EndpointFileScheme() (fileScheme scheme.Type) {
	if fileScheme = m.Service.EndpointFileScheme(); fileScheme != "" {
		return fileScheme
	}

	return ServiceFileScheme
}

// EndpointRequestFormat returns the endpoint API request format.
func (m *Model) EndpointRequestFormat() (format ApiFormat) {
	if format = m.Service.EndpointRequestFormat(); format != "" {
		return format
	}

	return ServiceRequestFormat
}

// EndpointResponseFormat returns the endpoint API response format.
func (m *Model) EndpointResponseFormat() (format ApiFormat) {
	if format = m.Service.EndpointResponseFormat(); format != "" {
		return format
	}

	return ServiceResponseFormat
}

// ClassifyModel returns the matching classify model instance, if any.
func (m *Model) ClassifyModel() *classify.Model {
	// Use mutex to prevent models from being loaded and
	// initialized twice by different indexing workers.
	modelMutex.Lock()
	defer modelMutex.Unlock()

	// Return the existing model instance if it has already been created.
	if m.classifyModel != nil {
		return m.classifyModel
	}

	switch m.Name {
	case "":
		log.Warnf("vision: missing name, model instance cannot be created")
		return nil
	case NasnetModel.Name, "nasnet":
		// Load and initialize the Nasnet image classification model.
		if model := classify.NewNasnet(GetModelsPath(), m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init nasnet model)", err)
			return nil
		} else {
			m.classifyModel = model
		}
	default:
		// Set model path from model name if no path is configured.
		if m.Path == "" {
			m.Path = clean.Path(clean.TypeLowerUnderscore(m.Name))
		}

		if m.Meta == nil {
			m.Meta = &tensorflow.ModelInfo{}
		}

		// Set default thumbnail resolution if no tags are configured.
		if m.Resolution <= 0 {
			m.Resolution = DefaultResolution
		}

		if m.Meta.Input == nil {
			m.Meta.Input = new(tensorflow.PhotoInput)
		}

		m.Meta.Input.SetResolution(m.Resolution)

		// Try to load custom model based on the configuration values.
		if model := classify.NewModel(GetModelsPath(), m.Path, GetNasnetModelPath(), m.Meta, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.classifyModel = model
		}
	}

	return m.classifyModel
}

// FaceModel returns the matching face model instance, if any.
func (m *Model) FaceModel() *face.Model {
	// Use mutex to prevent models from being loaded and
	// initialized twice by different indexing workers.
	modelMutex.Lock()
	defer modelMutex.Unlock()

	// Return the existing model instance if it has already been created.
	if m.faceModel != nil {
		return m.faceModel
	}

	switch m.Name {
	case "":
		log.Warnf("vision: missing name, model instance cannot be created")
		return nil
	case FacenetModel.Name, "facenet":
		// Load and initialize the Nasnet image classification model.
		if model := face.NewModel(GetFacenetModelPath(), GetCachePath(), m.Resolution, m.Meta, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.faceModel = model
		}
	default:
		// Set model path from model name if no path is configured.
		if m.Path == "" {
			m.Path = clean.Path(clean.TypeLowerUnderscore(m.Name))
		}

		// Set default thumbnail resolution if no tags are configured.
		if m.Resolution <= 0 {
			m.Resolution = DefaultResolution
		}

		if m.Meta == nil {
			m.Meta = &tensorflow.ModelInfo{}
		}

		// Try to load custom model based on the configuration values.
		if model := face.NewModel(GetModelPath(m.Path), GetCachePath(), m.Resolution, m.Meta, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.faceModel = model
		}
	}

	return m.faceModel
}

// NsfwModel returns the matching nsfw model instance, if any.
func (m *Model) NsfwModel() *nsfw.Model {
	// Use mutex to prevent models from being loaded and
	// initialized twice by different indexing workers.
	modelMutex.Lock()
	defer modelMutex.Unlock()

	// Return the existing model instance if it has already been created.
	if m.nsfwModel != nil {
		return m.nsfwModel
	}

	switch m.Name {
	case "":
		log.Warnf("vision: missing name, model instance cannot be created")
		return nil
	case NsfwModel.Name, "nsfw":
		// Load and initialize the Nasnet image classification model.
		if model := nsfw.NewModel(GetNsfwModelPath(), NsfwModel.Meta, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.nsfwModel = model
		}
	default:
		// Set model path from model name if no path is configured.
		if m.Path == "" {
			m.Path = clean.Path(clean.TypeLowerUnderscore(m.Name))
		}

		// Set default thumbnail resolution if no tags are configured.
		if m.Resolution <= 0 {
			m.Resolution = DefaultResolution
		}

		if m.Meta.Input == nil {
			m.Meta.Input = new(tensorflow.PhotoInput)
		}

		m.Meta.Input.SetResolution(m.Resolution)

		if m.Meta == nil {
			m.Meta = &tensorflow.ModelInfo{}
		}

		// Try to load custom model based on the configuration values.
		if model := nsfw.NewModel(GetModelPath(m.Path), m.Meta, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.nsfwModel = model
		}
	}

	return m.nsfwModel
}
