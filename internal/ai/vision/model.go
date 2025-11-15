package vision

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	visionschema "github.com/photoprism/photoprism/internal/ai/vision/schema"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

var modelMutex = sync.Mutex{}

const labelSchemaEnvVar = "PHOTOPRISM_VISION_LABEL_SCHEMA_FILE"

// Default model version strings.
var (
	VersionLatest = "latest"
	VersionMobile = "mobile"
	Version3B     = "3b"
)

// Model represents a computer vision model configuration.
type Model struct {
	Type          ModelType             `yaml:"Type,omitempty" json:"type,omitempty"`
	Default       bool                  `yaml:"Default,omitempty" json:"default,omitempty"`
	Name          string                `yaml:"Name,omitempty" json:"name,omitempty"`
	Version       string                `yaml:"Version,omitempty" json:"version,omitempty"`
	Engine        ModelEngine           `yaml:"Engine,omitempty" json:"engine,omitempty"`
	Run           RunType               `yaml:"Run,omitempty" json:"Run,omitempty"` // "auto", "never", "manual", "always", "newly-indexed", "on-schedule"
	System        string                `yaml:"System,omitempty" json:"system,omitempty"`
	Prompt        string                `yaml:"Prompt,omitempty" json:"prompt,omitempty"`
	Format        string                `yaml:"Format,omitempty" json:"format,omitempty"`
	Schema        string                `yaml:"Schema,omitempty" json:"schema,omitempty"`
	SchemaFile    string                `yaml:"SchemaFile,omitempty" json:"schemaFile,omitempty"`
	Resolution    int                   `yaml:"Resolution,omitempty" json:"resolution,omitempty"`
	TensorFlow    *tensorflow.ModelInfo `yaml:"TensorFlow,omitempty" json:"tensorflow,omitempty"`
	Options       *ApiRequestOptions    `yaml:"Options,omitempty" json:"options,omitempty"`
	Service       Service               `yaml:"Service,omitempty" json:"service,omitempty"`
	Path          string                `yaml:"Path,omitempty" json:"-"`
	Disabled      bool                  `yaml:"Disabled,omitempty" json:"disabled,omitempty"`
	classifyModel *classify.Model
	faceModel     *face.Model
	nsfwModel     *nsfw.Model
	schemaOnce    sync.Once
	schema        string
}

// Models represents a set of computer vision models.
type Models []*Model

// Model returns the parsed and normalized identifier, name, and version
// strings. Nil receivers return empty values so callers can destructure the
// tuple without additional nil checks.
func (m *Model) Model() (model, name, version string) {
	if m == nil {
		return "", "", ""
	}

	// Return empty identifier string if no name was set.
	if m.Name == "" {
		return "", "", clean.TypeLowerDash(m.Version)
	}

	// Normalize model name.
	name = clean.TypeLower(m.Name)

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

// IsDefault reports whether the model refers to one of the built-in defaults.
// Nil receivers return false.
func (m *Model) IsDefault() bool {
	if m == nil {
		return false
	}

	if m.Default {
		return true
	}

	if m.TensorFlow == nil {
		return false
	}

	switch m.Type {
	case ModelTypeLabels:
		return m.Name == NasnetModel.Name
	case ModelTypeNsfw:
		return m.Name == NsfwModel.Name
	case ModelTypeFace:
		return m.Name == FacenetModel.Name
	case ModelTypeCaption:
		return m.Name == CaptionModel.Name
	}

	return false
}

// Endpoint returns the remote service request method and endpoint URL. Nil
// receivers return empty strings.
func (m *Model) Endpoint() (uri, method string) {
	if m == nil {
		return uri, method
	}

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

// EndpointKey returns the access token belonging to the remote service
// endpoint, or an empty string for nil receivers.
func (m *Model) EndpointKey() (key string) {
	if m == nil {
		return ""
	}

	if key = m.Service.EndpointKey(); key != "" {
		return key
	}

	ensureEnv()

	return strings.TrimSpace(os.ExpandEnv(ServiceKey))
}

// EndpointFileScheme returns the endpoint API request file scheme type. Nil
// receivers fall back to the global default scheme.
func (m *Model) EndpointFileScheme() (fileScheme scheme.Type) {
	if m == nil {
		return ""
	}

	if fileScheme = m.Service.EndpointFileScheme(); fileScheme != "" {
		return fileScheme
	}

	return ServiceFileScheme
}

// EndpointRequestFormat returns the endpoint API request format. Nil receivers
// fall back to the global default format.
func (m *Model) EndpointRequestFormat() (format ApiFormat) {
	if m == nil {
		return ""
	}

	if format = m.Service.EndpointRequestFormat(); format != "" {
		return format
	}

	return ServiceRequestFormat
}

// EndpointResponseFormat returns the endpoint API response format. Nil
// receivers fall back to the global default format.
func (m *Model) EndpointResponseFormat() (format ApiFormat) {
	if m == nil {
		return ""
	}

	if format = m.Service.EndpointResponseFormat(); format != "" {
		return format
	}

	return ServiceResponseFormat
}

// GetPrompt returns the configured model prompt, using engine defaults when
// none is specified. Nil receivers return an empty string.
func (m *Model) GetPrompt() string {
	if m == nil {
		return ""
	}

	if m.Prompt != "" {
		return m.Prompt
	}

	if defaults := m.engineDefaults(); defaults != nil {
		if prompt := defaults.UserPrompt(m); prompt != "" {
			return prompt
		}
	}

	switch m.Type {
	case ModelTypeCaption:
		return ollama.CaptionPrompt
	case ModelTypeLabels:
		return ollama.LabelPromptDefault
	default:
		return ""
	}
}

// PromptContains returns true if the prompt contains the specified substring.
func (m *Model) PromptContains(s string) bool {
	if s == "" {
		return false
	}

	return strings.Contains(m.GetSystemPrompt()+m.GetPrompt(), s)
}

// GetSystemPrompt returns the configured system prompt, falling back to
// engine defaults when none is specified. Nil receivers return an empty
// string.
func (m *Model) GetSystemPrompt() string {
	if m == nil {
		return ""
	}

	if m.System != "" {
		return m.System
	}

	if defaults := m.engineDefaults(); defaults != nil {
		if system := defaults.SystemPrompt(m); system != "" {
			return system
		}
	}

	switch m.Type {
	case ModelTypeLabels:
		return ollama.LabelSystem
	default:
		return ""
	}
}

// GetFormat returns the configured response format or a sensible default. Nil
// receivers return an empty string.
func (m *Model) GetFormat() string {
	if m == nil {
		return ""
	}

	if f := strings.TrimSpace(strings.ToLower(m.Format)); f != "" {
		return f
	}

	if m.Type == ModelTypeLabels && m.EndpointResponseFormat() == ApiFormatOllama {
		return FormatJSON
	}

	return ""
}

// GetSource returns the default entity src based on the model configuration.
func (m *Model) GetSource() string {
	if m == nil {
		return entity.SrcAuto
	}

	switch m.EngineName() {
	case ollama.EngineName:
		return entity.SrcOllama
	case openai.EngineName:
		return entity.SrcOpenAI
	}

	switch m.EndpointRequestFormat() {
	case ApiFormatOllama:
		return entity.SrcOllama
	case ApiFormatOpenAI:
		return entity.SrcOpenAI
	}

	return entity.SrcImage
}

// GetOptions returns the API request options, applying engine defaults on
// demand. Nil receivers return nil.
func (m *Model) GetOptions() *ApiRequestOptions {
	if m == nil {
		return nil
	}

	var engineDefaults *ApiRequestOptions
	if defaults := m.engineDefaults(); defaults != nil {
		engineDefaults = cloneOptions(defaults.Options(m))
	}

	if m.Options == nil {
		switch m.Type {
		case ModelTypeLabels, ModelTypeCaption, ModelTypeGenerate:
			if engineDefaults == nil {
				engineDefaults = &ApiRequestOptions{}
			}
			normalizeOptions(engineDefaults)
			m.Options = engineDefaults
			return m.Options
		default:
			return nil
		}
	}

	mergeOptionDefaults(m.Options, engineDefaults)
	normalizeOptions(m.Options)

	return m.Options
}

func mergeOptionDefaults(target, defaults *ApiRequestOptions) {
	if target == nil || defaults == nil {
		return
	}

	if target.TopP <= 0 && defaults.TopP > 0 {
		target.TopP = defaults.TopP
	}

	if len(target.Stop) == 0 && len(defaults.Stop) > 0 {
		target.Stop = append([]string(nil), defaults.Stop...)
	}

	if target.MaxOutputTokens <= 0 && defaults.MaxOutputTokens > 0 {
		target.MaxOutputTokens = defaults.MaxOutputTokens
	}

	if strings.TrimSpace(target.Detail) == "" && strings.TrimSpace(defaults.Detail) != "" {
		target.Detail = strings.TrimSpace(defaults.Detail)
	}

	if !target.ForceJson && defaults.ForceJson {
		target.ForceJson = true
	}

	if target.SchemaVersion == "" && defaults.SchemaVersion != "" {
		target.SchemaVersion = defaults.SchemaVersion
	}

	if target.CombineOutputs == "" && defaults.CombineOutputs != "" {
		target.CombineOutputs = defaults.CombineOutputs
	}
}

func normalizeOptions(opts *ApiRequestOptions) {
	if opts == nil {
		return
	}

	if opts.Temperature <= 0 {
		opts.Temperature = DefaultTemperature
	} else if opts.Temperature > MaxTemperature {
		opts.Temperature = MaxTemperature
	}
}

func cloneOptions(opts *ApiRequestOptions) *ApiRequestOptions {
	if opts == nil {
		return nil
	}

	clone := *opts

	if len(opts.Stop) > 0 {
		clone.Stop = append([]string(nil), opts.Stop...)
	}

	return &clone
}

// EngineName returns the normalized engine identifier or infers one from the
// request configuration. Nil receivers return an empty string.
func (m *Model) EngineName() string {
	if m == nil {
		return ""
	}

	if engine := strings.TrimSpace(strings.ToLower(m.Engine)); engine != "" {
		return engine
	}

	uri, method := m.Endpoint()
	if uri != "" && method != "" {
		format := m.EndpointRequestFormat()
		switch format {
		case ApiFormatOllama:
			return ollama.EngineName
		case ApiFormatOpenAI:
			return openai.EngineName
		case ApiFormatVision, "":
			return EngineVision
		default:
			return strings.ToLower(string(format))
		}
	}

	if m.TensorFlow != nil {
		return EngineTensorFlow
	}

	return EngineLocal
}

// ApplyEngineDefaults normalizes the engine name and applies registered engine
// defaults (formats, schemes, resolution) when these are not explicitly configured.
func (m *Model) ApplyEngineDefaults() {
	if m == nil {
		return
	}

	engine := strings.TrimSpace(strings.ToLower(m.Engine))
	if engine == "" {
		return
	}

	if info, ok := EngineInfoFor(engine); ok {
		if m.Service.Uri == "" {
			m.Service.Uri = info.Uri
		}

		if m.Service.RequestFormat == "" {
			m.Service.RequestFormat = info.RequestFormat
		}

		if m.Service.ResponseFormat == "" {
			m.Service.ResponseFormat = info.ResponseFormat
		}

		if info.FileScheme != "" && m.Service.FileScheme == "" {
			m.Service.FileScheme = info.FileScheme
		}

		if info.DefaultResolution > 0 && m.Resolution <= 0 {
			m.Resolution = info.DefaultResolution
		}
	}

	if engine == openai.EngineName && strings.TrimSpace(m.Service.Key) == "" {
		m.Service.Key = "${OPENAI_API_KEY}"
	}

	m.Engine = engine
}

// SchemaTemplate returns the model-specific JSON schema template, if any. Nil
// receivers return an empty string.
func (m *Model) SchemaTemplate() string {
	if m == nil {
		return ""
	}

	m.schemaOnce.Do(func() {
		var schemaText string

		if m.Type == ModelTypeLabels {
			if envFile := strings.TrimSpace(os.Getenv(labelSchemaEnvVar)); envFile != "" {
				path := fs.Abs(envFile)
				if path == "" {
					path = envFile
				}
				if data, err := os.ReadFile(path); err != nil {
					log.Warnf("vision: failed to read schema from %s (%s)", clean.Log(path), err)
				} else {
					schemaText = string(data)
				}
			}
		}

		if schemaText == "" && strings.TrimSpace(m.Schema) != "" {
			schemaText = m.Schema
		}

		if schemaText == "" && strings.TrimSpace(m.SchemaFile) != "" {
			path := fs.Abs(m.SchemaFile)
			if path == "" {
				path = m.SchemaFile
			}
			if data, err := os.ReadFile(path); err != nil {
				log.Warnf("vision: failed to read schema from %s (%s)", clean.Log(path), err)
			} else {
				schemaText = string(data)
			}
		}

		m.schema = strings.TrimSpace(schemaText)

		if m.schema == "" && m.Type == ModelTypeLabels {
			if defaults := m.engineDefaults(); defaults != nil {
				m.schema = strings.TrimSpace(defaults.SchemaTemplate(m))
			}

			if m.schema == "" {
				m.schema = visionschema.LabelsJson(m.PromptContains("nsfw"))
			}
		}
	})

	return m.schema
}

func (m *Model) engineDefaults() EngineDefaults {
	if m == nil {
		return nil
	}

	if engine, ok := EngineFor(m.EndpointRequestFormat()); ok {
		return engine.Defaults
	}

	if info, ok := EngineInfoFor(m.EngineName()); ok {
		if engine, ok := EngineFor(info.RequestFormat); ok {
			return engine.Defaults
		}
	}
	return nil
}

// SchemaInstructions returns a helper string that can be appended to prompts.
// Nil receivers return an empty string.
func (m *Model) SchemaInstructions() string {
	if m == nil {
		return ""
	}

	if schema := m.SchemaTemplate(); schema != "" {
		return fmt.Sprintf("Return JSON that matches this schema:\n%s", schema)
	}

	return ""
}

// ClassifyModel returns the matching classify model instance, if any. Nil
// receivers return nil.
func (m *Model) ClassifyModel() *classify.Model {
	if m == nil {
		return nil
	}

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

		if m.TensorFlow == nil {
			m.TensorFlow = &tensorflow.ModelInfo{}
		}

		// Set default thumbnail resolution if no tags are configured.
		if m.Resolution <= 0 {
			m.Resolution = DefaultResolution
		}

		if m.TensorFlow.Input == nil {
			m.TensorFlow.Input = new(tensorflow.PhotoInput)
		}

		m.TensorFlow.Input.SetResolution(m.Resolution)

		// Try to load custom model based on the configuration values.
		if model := classify.NewModel(GetModelsPath(), m.Path, GetNasnetModelPath(), m.TensorFlow, m.Disabled); model == nil {
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

// FaceModel returns the matching face recognition model instance, if any. Nil
// receivers return nil.
func (m *Model) FaceModel() *face.Model {
	if m == nil {
		return nil
	}

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
		if model := face.NewModel(GetFacenetModelPath(), GetCachePath(), m.Resolution, m.TensorFlow, m.Disabled); model == nil {
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

		if m.TensorFlow == nil {
			m.TensorFlow = &tensorflow.ModelInfo{}
		}

		// Try to load custom model based on the configuration values.
		if model := face.NewModel(GetModelPath(m.Path), GetCachePath(), m.Resolution, m.TensorFlow, m.Disabled); model == nil {
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

// NsfwModel returns the matching nsfw model instance, if any. Nil receivers
// return nil.
func (m *Model) NsfwModel() *nsfw.Model {
	if m == nil {
		return nil
	}

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
		if model := nsfw.NewModel(GetNsfwModelPath(), NsfwModel.TensorFlow, m.Disabled); model == nil {
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

		if m.TensorFlow.Input == nil {
			m.TensorFlow.Input = new(tensorflow.PhotoInput)
		}

		m.TensorFlow.Input.SetResolution(m.Resolution)

		if m.TensorFlow == nil {
			m.TensorFlow = &tensorflow.ModelInfo{}
		}

		// Try to load custom model based on the configuration values.
		if model := nsfw.NewModel(GetModelPath(m.Path), m.TensorFlow, m.Disabled); model == nil {
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

// Clone returns a shallow copy of the model. Nil receivers return nil.
func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}

	c := *m
	return &c
}
