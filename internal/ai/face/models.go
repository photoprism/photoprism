package face

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/fs"
)

// EmbeddingRuntime identifies the inference runtime that generates face embeddings.
type EmbeddingRuntime = string

const (
	// RuntimeTensorFlow generates embeddings through the TensorFlow bindings.
	RuntimeTensorFlow EmbeddingRuntime = "tensorflow"
	// RuntimeONNX generates embeddings through ONNX Runtime.
	RuntimeONNX EmbeddingRuntime = "onnx"
)

// CropAlignment identifies how face crops are prepared before embedding inference.
type CropAlignment = string

const (
	// AlignBox scales the detected face bounding box without using landmarks.
	AlignBox CropAlignment = "box"
	// AlignArcFace5 warps the five detected landmarks onto the standard ArcFace template.
	AlignArcFace5 CropAlignment = "arcface5"
)

// Licenses that apply to the pretrained weights of the supported embedding models.
const (
	// LicenseApache2 allows redistribution under the Apache License 2.0.
	LicenseApache2 = "Apache-2.0"
	// LicenseResearchOnly restricts the weights to non-commercial research unless licensed separately.
	LicenseResearchOnly = "research-only"
	// LicenseUnknown marks weights whose provenance has not been verified.
	LicenseUnknown = "unknown"
)

// ModelName identifies a face embedding model.
type ModelName = string

const (
	// ModelAuto resolves to the first model in AutoModelPreference that is installed.
	ModelAuto ModelName = "auto"
	// ModelNone disables embedding generation so only face regions are detected.
	ModelNone ModelName = "none"
	// ModelFaceNet is the TensorFlow FaceNet model PhotoPrism has shipped since 2021.
	ModelFaceNet ModelName = "facenet"
	// ModelSFace is the OpenCV Zoo SFace model.
	ModelSFace ModelName = "sface"
	// ModelArcFaceR50 is the InsightFace ArcFace WebFace600K ResNet-50 model.
	ModelArcFaceR50 ModelName = "arcface_r50"
	// ModelArcFaceMBF is the InsightFace ArcFace WebFace600K MobileFaceNet model.
	ModelArcFaceMBF ModelName = "arcface_mbf"
)

// EmbeddingModel describes what a face embedding model expects from the pipeline.
// Preprocessing is expressed as (channel - Mean) * Scale, matching the blobFromImage
// convention that the upstream reference implementations use, so the values can be
// compared against them directly.
type EmbeddingModel struct {
	Name       ModelName
	Runtime    EmbeddingRuntime
	Dir        string
	FileName   string
	Width      int
	Height     int
	Dims       int
	ColorOrder tensorflow.ColorChannelOrder
	Mean       float32
	Scale      float32
	Alignment  CropAlignment
	License    string
}

// EmbeddingModels lists the supported face embedding models by name.
//
// The ArcFace entries are recognized so operators can benchmark them, but their
// weights are not installed by "make dep" because InsightFace publishes them for
// non-commercial research only.
var EmbeddingModels = map[ModelName]*EmbeddingModel{
	ModelFaceNet: {
		Name:       ModelFaceNet,
		Runtime:    RuntimeTensorFlow,
		Dir:        "facenet",
		Width:      160,
		Height:     160,
		Dims:       512,
		ColorOrder: tensorflow.RGB,
		Mean:       127.5,
		Scale:      1 / 127.5,
		Alignment:  AlignBox,
		License:    LicenseUnknown,
	},
	ModelSFace: {
		Name:       ModelSFace,
		Runtime:    RuntimeONNX,
		Dir:        "sface",
		FileName:   "face_recognition_sface_2021dec.onnx",
		Width:      112,
		Height:     112,
		Dims:       128,
		ColorOrder: tensorflow.RGB,
		Mean:       0,
		Scale:      1,
		Alignment:  AlignArcFace5,
		License:    LicenseApache2,
	},
	ModelArcFaceR50: {
		Name:       ModelArcFaceR50,
		Runtime:    RuntimeONNX,
		Dir:        "arcface",
		FileName:   "w600k_r50.onnx",
		Width:      112,
		Height:     112,
		Dims:       512,
		ColorOrder: tensorflow.RGB,
		Mean:       127.5,
		Scale:      1 / 127.5,
		Alignment:  AlignArcFace5,
		License:    LicenseResearchOnly,
	},
	ModelArcFaceMBF: {
		Name:       ModelArcFaceMBF,
		Runtime:    RuntimeONNX,
		Dir:        "arcface",
		FileName:   "w600k_mbf.onnx",
		Width:      112,
		Height:     112,
		Dims:       512,
		ColorOrder: tensorflow.RGB,
		Mean:       127.5,
		Scale:      1 / 127.5,
		Alignment:  AlignArcFace5,
		License:    LicenseResearchOnly,
	},
}

// AutoModelPreference lists embedding models in the order that ModelAuto prefers them.
//
// FaceNet stays first while it is the shipped default: resolving "auto" to a different
// model would silently start writing embeddings from an incompatible vector space into
// libraries that were clustered with FaceNet.
var AutoModelPreference = []ModelName{ModelFaceNet, ModelSFace, ModelArcFaceR50, ModelArcFaceMBF}

// NormalizeModelName lowercases a model name and accepts hyphens in place of underscores.
func NormalizeModelName(s string) ModelName {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "-", "_")
}

// ParseModelName returns the supported model name matching s, or ModelAuto when unknown.
func ParseModelName(s string) ModelName {
	name := NormalizeModelName(s)

	switch name {
	case ModelAuto, ModelNone:
		return name
	}

	if _, found := EmbeddingModels[name]; found {
		return name
	}

	return ModelAuto
}

// KnownModelName reports whether s names a supported embedding model, "auto", or "none".
func KnownModelName(s string) bool {
	name := NormalizeModelName(s)

	switch name {
	case ModelAuto, ModelNone:
		return true
	}

	_, found := EmbeddingModels[name]

	return found
}

// FindEmbeddingModel returns the model registered under the specified name, or nil when unknown.
func FindEmbeddingModel(name ModelName) *EmbeddingModel {
	return EmbeddingModels[NormalizeModelName(name)]
}

// EmbeddingModelNames returns the names of all supported embedding models in alphabetical order.
func EmbeddingModelNames() []ModelName {
	names := make([]ModelName, 0, len(EmbeddingModels))

	for name := range EmbeddingModels {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ModelUsageString lists the accepted FACE_MODEL values for use in CLI help text.
// It is generated from the registry so the help can never advertise a model that
// has been renamed or removed.
func ModelUsageString() string {
	return strings.Join(append([]ModelName{ModelAuto, ModelNone}, EmbeddingModelNames()...), ", ")
}

// FilePath returns the absolute model path within the specified models directory.
// TensorFlow SavedModel bundles are loaded from their directory, so entries without
// a file name resolve to the directory itself.
func (m *EmbeddingModel) FilePath(modelsPath string) string {
	if m == nil || modelsPath == "" {
		return ""
	}

	if m.FileName == "" {
		return filepath.Join(modelsPath, m.Dir)
	}

	return filepath.Join(modelsPath, m.Dir, m.FileName)
}

// Installed reports whether the model files exist in the specified models directory.
func (m *EmbeddingModel) Installed(modelsPath string) bool {
	if m == nil {
		return false
	}

	path := m.FilePath(modelsPath)

	if path == "" {
		return false
	}

	if m.FileName == "" {
		return fs.PathExists(path)
	}

	return fs.FileExists(path)
}

// Aligned reports whether the model requires landmark-aligned crops.
func (m *EmbeddingModel) Aligned() bool {
	return m != nil && m.Alignment == AlignArcFace5
}

// String returns the model name, or "none" for nil receivers.
func (m *EmbeddingModel) String() string {
	if m == nil {
		return ModelNone
	}

	return m.Name
}
