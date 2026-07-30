package face

import (
	"image"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Embedder generates face embeddings from prepared face crops.
type Embedder interface {
	// ModelName returns the name of the embedding model.
	ModelName() ModelName
	// Dims returns the length of the embeddings the model produces.
	Dims() int
	// CropSize returns the input width and height the model expects.
	CropSize() (width, height int)
	// Aligned reports whether the model requires landmark-aligned crops.
	Aligned() bool
	// Run returns the embeddings for a prepared face crop.
	Run(img image.Image) Embeddings
	// Close releases the resources held by the model.
	Close() error
}

// EmbedderSettings captures the configuration required to initialize an embedder.
type EmbedderSettings struct {
	Name        ModelName
	Model       *EmbeddingModel
	ModelPath   string
	Threads     int
	LibraryPath string
}

var (
	embedderMu      sync.RWMutex
	activeEmbedder  Embedder
	configuredModel = ModelAuto
)

// UseEmbedder replaces the active embedding model and returns the previous instance.
func UseEmbedder(embedder Embedder) (previous Embedder) {
	embedderMu.Lock()
	previous = activeEmbedder
	activeEmbedder = embedder
	embedderMu.Unlock()

	return previous
}

// ActiveEmbedder returns the currently configured embedding model, or nil when none is active.
func ActiveEmbedder() Embedder {
	embedderMu.RLock()
	embedder := activeEmbedder
	embedderMu.RUnlock()

	return embedder
}

// ConfiguredModel returns the embedding model name that was last configured.
func ConfiguredModel() ModelName {
	embedderMu.RLock()
	name := configuredModel
	embedderMu.RUnlock()

	return name
}

// ExpectedDims returns the embedding length that the configured model produces.
// Persistence uses it to reject vectors from a different model instead of mixing
// incompatible embedding spaces in the same library.
func ExpectedDims() int {
	if embedder := ActiveEmbedder(); embedder != nil {
		if dims := embedder.Dims(); dims > 0 {
			return dims
		}
	}

	if m := FindEmbeddingModel(ConfiguredModel()); m != nil {
		return m.Dims
	}

	return len(NullEmbedding)
}

// SamplesModel names the embedding model that the bundled children and background
// reference samples were generated with.
const SamplesModel = ModelFaceNet

// SamplesComparable reports whether the bundled reference samples can be compared with
// embeddings from the configured model. Vectors are not comparable across models even
// when their length matches, so the samples would otherwise produce arbitrary verdicts.
func SamplesComparable() bool {
	name := EmbeddingModelName()

	return name == "" || name == SamplesModel
}

// EmbeddingModelName returns the model name to record with newly generated embeddings.
// It is empty when the model cannot be determined, so unknown provenance is never
// mistaken for a specific model.
func EmbeddingModelName() ModelName {
	if embedder := ActiveEmbedder(); embedder != nil {
		return embedder.ModelName()
	}

	if name := ConfiguredModel(); FindEmbeddingModel(name) != nil {
		return name
	}

	return ""
}

// EmbeddingsDisabled reports whether the configuration turns off embedding generation,
// which lets callers skip embeddings without reporting a broken model configuration.
func EmbeddingsDisabled() bool {
	return ConfiguredModel() == ModelNone
}

// ConfigureEmbedder initializes the embedding model described by the settings.
//
// Only ONNX models are instantiated here. TensorFlow models stay with the vision
// subsystem so that custom entries in vision.yml keep their configured model path
// and graph metadata.
func ConfigureEmbedder(settings EmbedderSettings) error {
	var newEmbedder Embedder
	var initErr error

	name := NormalizeModelName(settings.Name)

	if name == "" {
		name = settings.Model.String()
	}

	if settings.Model != nil && settings.Model.Runtime == RuntimeONNX {
		newEmbedder, initErr = NewONNXEmbedder(settings)
	}

	embedderMu.Lock()
	previous := activeEmbedder
	activeEmbedder = newEmbedder
	configuredModel = name
	embedderMu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}

	// The bundled reference samples belong to one vector space, so tell operators when
	// the child and background filters stop applying instead of failing quietly.
	if initErr == nil && !SamplesComparable() {
		log.Warnf("faces: children and background samples do not apply to %s, so those filters are inactive", clean.Log(name))
	}

	return initErr
}
