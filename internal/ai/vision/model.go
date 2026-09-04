package vision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
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
	VersionCloud  = "cloud"
)

// Model represents a computer vision model configuration.
type Model struct {
	Type           ModelType             `yaml:"Type,omitempty" json:"type,omitempty"`
	Default        bool                  `yaml:"Default,omitempty" json:"default,omitempty"`
	Model          string                `yaml:"Model,omitempty" json:"model,omitempty"`
	Name           string                `yaml:"Name,omitempty" json:"name,omitempty"`
	Version        string                `yaml:"Version,omitempty" json:"version,omitempty"`
	Engine         ModelEngine           `yaml:"Engine,omitempty" json:"engine,omitempty"`
	Run            RunType               `yaml:"Run,omitempty" json:"Run,omitempty"` // "auto", "never", "manual", "always", "newly-indexed", "on-schedule"
	System         string                `yaml:"System,omitempty" json:"system,omitempty"`
	Prompt         string                `yaml:"Prompt,omitempty" json:"prompt,omitempty"`
	Format         string                `yaml:"Format,omitempty" json:"format,omitempty"`
	Normalize      NormalizeType         `yaml:"Normalize,omitempty" json:"normalize,omitempty"` // "single-word", "phrase", or "false"
	Schema         string                `yaml:"Schema,omitempty" json:"schema,omitempty"`
	SchemaFile     string                `yaml:"SchemaFile,omitempty" json:"schemaFile,omitempty"`
	Resolution     int                   `yaml:"Resolution,omitempty" json:"resolution,omitempty"`
	TensorFlow     *tensorflow.ModelInfo `yaml:"TensorFlow,omitempty" json:"tensorflow,omitempty"`
	ONNX           *onnx.ModelInfo       `yaml:"ONNX,omitempty" json:"onnx,omitempty"`
	LabelFile      string                `yaml:"LabelFile,omitempty" json:"labelFile,omitempty"`
	CanonicalOrder bool                  `yaml:"CanonicalOrder,omitempty" json:"canonicalOrder,omitempty"`
	Options        *ModelOptions         `yaml:"Options,omitempty" json:"options,omitempty"`
	Service        Service               `yaml:"Service,omitempty" json:"service"`
	Path           string                `yaml:"Path,omitempty" json:"-"`
	Disabled       bool                  `yaml:"Disabled,omitempty" json:"disabled,omitempty"`
	classifyModel  *classify.Model
	faceModel      face.Embedder
	nsfwModel      *nsfw.Model
	nsfwErr        error
	schemaOnce     sync.Once
	schema         string
}

// Models represents a set of computer vision models.
type Models []*Model

// Clone returns model copies with independent lazy state.
func (m Models) Clone() Models {
	result := make(Models, len(m))
	for i, model := range m {
		result[i] = model.Clone()
	}

	return result
}

// GetModel returns the normalized model identifier, name, and version strings
// used in service requests. Callers can always destructure the tuple because
// nil receivers return empty values.
func (m *Model) GetModel() (model, name, version string) {
	if m == nil {
		return "", "", ""
	}

	// Sanitize the configured values without lowercasing: upstream catalogs
	// (Ollama tags, Hugging Face IDs served by OpenAI-compatible endpoints)
	// match identifiers verbatim, so case must round-trip from vision.yml.
	name = clean.Type(m.Name)
	version = clean.Type(m.Version)

	// Build a base name from the highest-priority override:
	// 1) Service-specific override (expanded for env vars)
	// 2) Model-specific override
	// 3) Declarative model name
	serviceModel := m.Service.GetModel()
	switch {
	case serviceModel != "":
		name = serviceModel
	case strings.TrimSpace(m.Model) != "":
		name = clean.Type(m.Model)
	}

	// Return if no model is configured.
	if name == "" {
		return "", "", ""
	}

	// Split "name:version" strings so callers can access versioned models
	// without repeating parsing logic at each call site.
	if parts := strings.SplitN(name, ":", 2); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		name = parts[0]
		version = parts[1]
	}

	// Default to "latest" for non-OpenAI engines when no version was set.
	if version == "" {
		version = VersionLatest
	}

	switch m.Engine {
	case openai.EngineName:
		return name, name, ""
	case ollama.EngineName:
		return strings.Join([]string{name, version}, ":"), name, version
	default:
		return name, name, version
	}
}

// IsCloud reports whether the model runs as a cloud service rather than on local hardware.
// Each signal comes from the model, never from an engine-wide default, so a configuration that
// reaches both a local instance and a cloud service classifies each entry on its own.
func (m *Model) IsCloud() bool {
	if m == nil {
		return false
	}

	_, name, version := m.GetModel()

	// The "cloud" tag marks a model even when a local instance proxies the request.
	if version == VersionCloud {
		return true
	}

	// OpenAI-compatible local servers run open-weight models under their own names.
	if m.Engine == openai.EngineName && openai.IsCloudModel(name) {
		return true
	}

	uri, _ := m.Endpoint()

	return ollama.IsCloudUrl(uri)
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

	switch m.Type {
	case ModelTypeLabels:
		return m.ONNX != nil && m.Name == NasnetModel.Name
	case ModelTypeNsfw:
		return m.TensorFlow != nil && m.Name == NsfwModel.Name
	case ModelTypeFace:
		return m.TensorFlow != nil && m.Name == FacenetModel.Name
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

// ApplyService updates the ApiRequest with service-specific
// values when configured.
func (m *Model) ApplyService(apiRequest *ApiRequest) {
	if m == nil || apiRequest == nil {
		return
	}

	if m.Engine == openai.EngineName {
		apiRequest.Org = m.Service.EndpointOrg()
		apiRequest.Project = m.Service.EndpointProject()
		apiRequest.Tier = m.Service.EndpointTier()
	}

	if think := m.Service.EndpointThink(); think != "" {
		apiRequest.Think = think
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
func (m *Model) GetOptions() *ModelOptions {
	if m == nil {
		return nil
	}

	var engineDefaults *ModelOptions
	if defaults := m.engineDefaults(); defaults != nil {
		engineDefaults = cloneOptions(defaults.Options(m))
	}

	if m.Options == nil {
		switch m.Type {
		case ModelTypeLabels, ModelTypeCaption, ModelTypeGenerate:
			if engineDefaults == nil {
				engineDefaults = &ModelOptions{}
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

func mergeOptionDefaults(target, defaults *ModelOptions) {
	if target == nil || defaults == nil {
		return
	}

	if target.TopP <= 0 && defaults.TopP > 0 {
		target.TopP = defaults.TopP
	}

	if target.Temperature <= 0 && defaults.Temperature > 0 {
		target.Temperature = defaults.Temperature
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

func normalizeOptions(opts *ModelOptions) {
	if opts == nil {
		return
	}

	if opts.Temperature > MaxTemperature {
		opts.Temperature = MaxTemperature
	}
}

func cloneOptions(opts *ModelOptions) *ModelOptions {
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
			return strings.ToLower(format)
		}
	}

	if m.TensorFlow != nil {
		return EngineTensorFlow
	}

	if m.ONNX != nil {
		return EngineONNX
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
		if strings.TrimSpace(m.Model) == "" && strings.TrimSpace(m.Name) == "" {
			m.Model = info.DefaultModel
		}

		if strings.TrimSpace(m.Service.Uri) == "" {
			m.Service.Uri = info.Uri
		}

		if strings.TrimSpace(m.Service.RequestFormat) == "" {
			m.Service.RequestFormat = info.RequestFormat
		}

		if strings.TrimSpace(m.Service.ResponseFormat) == "" {
			m.Service.ResponseFormat = info.ResponseFormat
		}

		if strings.TrimSpace(m.Service.FileScheme) == "" && info.FileScheme != "" {
			m.Service.FileScheme = info.FileScheme
		}

		if m.Resolution <= 0 && info.DefaultResolution > 0 {
			m.Resolution = info.DefaultResolution
		}

		if strings.TrimSpace(m.Service.Key) == "" && info.DefaultKey != "" {
			m.Service.Key = info.DefaultKey
		}

		if strings.TrimSpace(m.Service.Think) == "" && info.DefaultThink != "" {
			m.Service.Think = info.DefaultThink
		}
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
				if schemaFromFile, err := readSchemaFile(envFile); err != nil {
					log.Warnf("vision: failed to read schema from %s (%s)", clean.Log(envFile), err)
				} else {
					schemaText = schemaFromFile
				}
			}
		}

		if schemaText == "" && strings.TrimSpace(m.Schema) != "" {
			schemaText = m.Schema
		}

		if schemaText == "" && strings.TrimSpace(m.SchemaFile) != "" {
			if schemaFromFile, err := readSchemaFile(m.SchemaFile); err != nil {
				log.Warnf("vision: failed to read schema from %s (%s)", clean.Log(m.SchemaFile), err)
			} else {
				schemaText = schemaFromFile
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

// readSchemaFile resolves and reads a schema file path from config or env.
func readSchemaFile(filePath string) (string, error) {
	path := fs.Abs(filePath)
	if path == "" {
		path = filePath
	}

	path = filepath.Clean(path)

	if path == "" {
		return "", fmt.Errorf("schema path is empty")
	}

	if _, err := fs.StatFile(path); err != nil {
		return "", err
	}

	// #nosec G304,G703 schema path is validated with Clean + fs.StatFile and comes from trusted config/env.
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
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

	switch {
	case m.Name == "":
		log.Warnf("vision: missing name, model instance cannot be created")
		return nil
	case classify.FindModel(classify.ModelName(m.Name)) != nil:
		// Load and initialize a registered ONNX image classification model.
		if model := classify.NewRegisteredModel(GetModelsPath(), classify.ModelName(m.Name), m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			m.Disabled = true
			log.Warnf("vision: %s (disable %s model)", err, clean.Log(m.Name))
			return nil
		} else {
			m.classifyModel = model
		}
	default:
		// Set model path from model name if no path is configured.
		if m.Path == "" {
			m.Path = clean.Path(clean.TypeLowerUnderscore(m.Name))
		}

		if m.ONNX == nil {
			m.ONNX = &onnx.ModelInfo{}
		}

		if m.ONNX.Input == nil && m.Resolution > 0 {
			m.ONNX.Input = &onnx.Input{Width: m.Resolution, Height: m.Resolution}
		} else if m.ONNX.Input != nil && m.Resolution > 0 {
			if m.ONNX.Input.Width <= 0 {
				m.ONNX.Input.Width = m.Resolution
			}
			if m.ONNX.Input.Height <= 0 {
				m.ONNX.Input.Height = m.Resolution
			}
		}

		modelPath := filepath.Join(GetModelsPath(), clean.Path(m.Path))
		modelDir := modelPath
		if !strings.EqualFold(filepath.Ext(modelPath), ".onnx") {
			modelDir = modelPath
			if m.ONNX.File == "" {
				m.ONNX.File = filepath.Base(m.Path) + ".onnx"
			}
			modelPath = m.ONNX.FilePath(modelDir)
		} else {
			modelDir = filepath.Dir(modelPath)
		}

		labelFile := m.LabelFile
		if labelFile == "" {
			labelFile = "labels.txt"
		}
		if !filepath.IsAbs(labelFile) {
			labelFile = filepath.Join(modelDir, labelFile)
		}

		// Try to load a custom ONNX model based on the configuration values.
		if model := classify.NewModel(classify.Settings{
			Name:           classify.ModelName(m.Name),
			ModelPath:      modelPath,
			LabelPath:      labelFile,
			Info:           m.ONNX,
			CanonicalOrder: m.CanonicalOrder,
			Disabled:       m.Disabled,
		}); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			m.Disabled = true
			log.Warnf("vision: %s (disable %s)", err, m.Path)
			return nil
		} else {
			m.classifyModel = model
		}
	}

	return m.classifyModel
}

// FaceModel returns the matching face recognition model instance, if any. Nil
// receivers return nil.
func (m *Model) FaceModel() face.Embedder {
	if m == nil {
		return nil
	}

	// FACE_MODEL=none turns embedding generation off, so no model may be loaded even
	// when vision.yml still schedules face processing to detect regions.
	if face.EmbeddingsDisabled() {
		return nil
	}

	// A library whose stored vectors were produced by another model has to be migrated
	// rather than added to, so nothing is embedded until it is. Detection keeps running,
	// because DetectFaces returns its markers instead of failing when this hands out none.
	if face.EmbeddingsBlocked() {
		return nil
	}

	return m.faceEmbedder()
}

// MigrationFaceModel returns the face embedding model instance for a migration, the one caller
// the gates above do not apply to: it writes every vector in its own target's space, so a gate
// against mixing spaces would only stop the work that resolves the mismatch.
func (m *Model) MigrationFaceModel() face.Embedder {
	if m == nil {
		return nil
	}

	return m.faceEmbedder()
}

// faceEmbedder returns the face embedding model instance, loading it when needed.
func (m *Model) faceEmbedder() face.Embedder {
	// An ONNX embedding model selected with FACE_MODEL takes precedence: vision.yml
	// only schedules when faces are processed, while the model itself is per instance.
	if embedder := face.ActiveEmbedder(); embedder != nil {
		return embedder
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
		if model := face.NewModel(face.ModelFaceNet, GetFacenetModelPath(), GetCachePath(), m.Resolution, m.TensorFlow, m.Disabled); model == nil {
			return nil
		} else if err := model.Init(); err != nil {
			log.Errorf("vision: %s (init %s)", err, m.Path)
			return nil
		} else {
			m.faceModel = model
		}
	default:
		// FACE_MODEL is authoritative for which model produces embeddings, and every
		// supported one needs code that knows its preprocessing contract, so there is
		// nothing useful to configure per installation here. Loading it anyway keeps an
		// existing vision.yml working, but its vectors are recorded under the configured
		// model's name rather than this one.
		log.Warnf("vision: custom face model %s in vision.yml is deprecated, select a model with FACE_MODEL instead",
			clean.Log(m.Name))

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
		if model := face.NewModel(face.NormalizeModelName(m.Name), GetModelPath(m.Path), GetCachePath(), m.Resolution, m.TensorFlow, m.Disabled); model == nil {
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

	// Cache initialization failure separately from the operator-controlled Disabled setting.
	if m.nsfwErr != nil {
		return nil
	}

	switch m.Name {
	case "":
		log.Warnf("vision: missing name, model instance cannot be created")
		return nil
	case NsfwModel.Name, "nsfw":
		// Load and initialize the bundled TensorFlow detector.
		model := nsfw.NewModel(GetNsfwModelPath(), NsfwModel.TensorFlow, m.Disabled)

		if err := model.Init(); err != nil {
			m.nsfwErr = err
			log.Warnf("vision: %s (init %s model)", clean.Error(err), clean.Log(m.Name))
			return nil
		}

		m.nsfwModel = model
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

		if m.TensorFlow.Input == nil {
			m.TensorFlow.Input = new(tensorflow.PhotoInput)
		}

		m.TensorFlow.Input.SetResolution(m.Resolution)

		// Try to load custom model based on the configuration values.
		model := nsfw.NewModel(GetModelPath(m.Path), m.TensorFlow, m.Disabled)

		if err := model.Init(); err != nil {
			m.nsfwErr = err
			log.Warnf("vision: %s (init %s)", clean.Error(err), clean.Log(m.Path))
			return nil
		}

		m.nsfwModel = model
	}

	return m.nsfwModel
}

// Clone returns a copy of the model with its own lazily derived state. Nil receivers return nil.
//
// The schema is reset rather than carried over, so a clone derives one from its own fields instead
// of inheriting whatever was computed for the model it came from.
func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}

	//nolint:govet // Copying the guard is safe because the copy is reset before it is used.
	c := *m

	c.schemaOnce = sync.Once{}
	c.schema = ""
	c.classifyModel = nil
	c.faceModel = nil
	c.nsfwModel = nil
	c.nsfwErr = nil
	c.Options = cloneOptions(m.Options)
	if m.TensorFlow != nil {
		tensorFlowInfo := *m.TensorFlow
		tensorFlowInfo.Tags = append([]string(nil), m.TensorFlow.Tags...)
		if m.TensorFlow.Input != nil {
			input := *m.TensorFlow.Input
			input.Intervals = append([]tensorflow.Interval(nil), m.TensorFlow.Input.Intervals...)
			for i := range input.Intervals {
				if input.Intervals[i].Mean != nil {
					mean := *input.Intervals[i].Mean
					input.Intervals[i].Mean = &mean
				}
				if input.Intervals[i].StdDev != nil {
					stdDev := *input.Intervals[i].StdDev
					input.Intervals[i].StdDev = &stdDev
				}
			}
			input.Shape = append([]tensorflow.ShapeComponent(nil), m.TensorFlow.Input.Shape...)
			tensorFlowInfo.Input = &input
		}
		if m.TensorFlow.Output != nil {
			output := *m.TensorFlow.Output
			tensorFlowInfo.Output = &output
		}
		c.TensorFlow = &tensorFlowInfo
	}
	if m.ONNX != nil {
		onnxInfo := *m.ONNX
		if m.ONNX.Input != nil {
			input := *m.ONNX.Input
			onnxInfo.Input = &input
		}
		if m.ONNX.Output != nil {
			output := *m.ONNX.Output
			if m.ONNX.Output.Logits != nil {
				logits := *m.ONNX.Output.Logits
				output.Logits = &logits
			}
			onnxInfo.Output = &output
		}
		c.ONNX = &onnxInfo
	}

	return &c
}
