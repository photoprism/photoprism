package vision

import (
	"net/http"
	"path"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/clean"
)

var modelMutex = sync.Mutex{}

// Model represents a computer vision model configuration.
type Model struct {
	Name          string   `yaml:"Name,omitempty" json:"name,omitempty"`
	Version       string   `yaml:"Version,omitempty" json:"version,omitempty"`
	Resolution    int      `yaml:"Resolution,omitempty" json:"resolution,omitempty"`
	Uri           string   `yaml:"Uri,omitempty" json:"-"`
	Key           string   `yaml:"Key,omitempty" json:"-"`
	Method        string   `yaml:"Method,omitempty" json:"-"`
	Path          string   `yaml:"Path,omitempty" json:"-"`
	Format        string   `yaml:"Format,omitempty" json:"-"`
	Tags          []string `yaml:"Tags,omitempty" json:"-"`
	Disabled      bool     `yaml:"Disabled,omitempty" json:"-"`
	classifyModel *classify.Model
}

// Models represents a set of computer vision models.
type Models []*Model

// Endpoint returns the remote service request method and endpoint URL, if any.
func (m *Model) Endpoint(name string) (method, uri string) {
	if m.Uri == "" && ServiceUri == "" {
		return "", ""
	}

	if m.Method != "" {
		method = m.Method
	} else {
		method = http.MethodPost
	}

	if m.Uri != "" {
		return m.Uri, method
	} else {
		return path.Join(ServiceUri, name), method
	}
}

// EndpointKey returns the access token belonging to the remote service endpoint, if any.
func (m *Model) EndpointKey() string {
	if m.Key != "" {
		return m.Key
	} else if ServiceKey != "" {
		return ServiceKey
	}

	return ""
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
		if model := classify.NewNasnet(AssetsPath, m.Disabled); model == nil {
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
			m.Path = clean.TypeLowerUnderscore(m.Name)
		}

		// Set default thumbnail resolution if no tags are configured.
		if m.Resolution <= 0 {
			m.Resolution = DefaultResolution
		}

		// Set default tag if no tags are configured.
		if len(m.Tags) == 0 {
			m.Tags = []string{"serve"}
		}

		// Try to load custom model based on the configuration values.
		if model := classify.NewModel(AssetsPath, m.Path, m.Resolution, m.Tags, m.Disabled); model == nil {
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
