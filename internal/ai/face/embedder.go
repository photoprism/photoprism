package face

import (
	"image"
	"sync"
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
	embedderMu       sync.RWMutex
	activeEmbedder   Embedder
	embedderSettings EmbedderSettings
	embedderLoaded   bool
	configuredModel  = ModelDetect
	embedderErr      error
)

// UseEmbedder replaces the active embedding model and returns the previous instance.
func UseEmbedder(embedder Embedder) (previous Embedder) {
	embedderMu.Lock()
	previous = activeEmbedder
	activeEmbedder = embedder
	// What is active no longer came from ConfigureEmbedder, so the recorded settings do not
	// describe it and must not let the next call keep it.
	embedderLoaded = false
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

// EmbedderError returns the error that prevented the configured embedding model from
// loading, or nil when none was requested or it loaded successfully. Both cases report
// ModelNone, so this is what tells a broken model apart from a disabled one.
func EmbedderError() error {
	embedderMu.RLock()
	err := embedderErr
	embedderMu.RUnlock()

	return err
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
// A model that failed to load reports the same, so use EmbedderError to tell them apart.
func EmbeddingsDisabled() bool {
	return ConfiguredModel() == ModelNone
}

// ConfigureEmbedder initializes the embedding model described by the settings.
//
// Only ONNX models are instantiated here. TensorFlow models stay with the vision
// subsystem so that custom entries in vision.yml keep their configured model path
// and graph metadata.
func ConfigureEmbedder(settings EmbedderSettings) error {
	// Loading the same model again produces an identical session, so the active one is kept: each
	// load reads the weights, verifies the checksum, and creates an inference session. A file
	// replaced under an unchanged path keeps the copy that was verified when it was loaded.
	if reuseEmbedder(settings) {
		return nil
	}

	var newEmbedder Embedder
	var initErr error

	name := NormalizeModelName(settings.Name)

	if name == "" {
		name = settings.Model.String()
	}

	if settings.Model != nil && settings.Model.Runtime == RuntimeONNX {
		newEmbedder, initErr = NewONNXEmbedder(settings)
	}

	// A model whose embedder failed to initialize must not stay recorded as the source
	// of new embeddings: callers would fall back to another runtime and stamp its
	// vectors with a name from an incompatible embedding space.
	if initErr != nil {
		name = ModelNone
	}

	embedderMu.Lock()
	previous := activeEmbedder
	activeEmbedder = newEmbedder
	embedderSettings = settings
	embedderLoaded = newEmbedder != nil && initErr == nil
	configuredModel = name
	embedderErr = initErr
	embedderMu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}

	return initErr
}

// reuseEmbedder reports whether the active embedding model was loaded from these settings and
// can serve them again. A model that failed to load is never reusable, so an attempt that ran
// before its weights were installed is retried rather than remembered.
func reuseEmbedder(settings EmbedderSettings) bool {
	embedderMu.RLock()
	defer embedderMu.RUnlock()

	return embedderLoaded && activeEmbedder != nil && embedderSettings == settings
}
