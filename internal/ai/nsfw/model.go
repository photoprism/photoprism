package nsfw

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// Model runs an ONNX NSFW detector and reduces its output to an unsafe probability.
type Model struct {
	session           *onnxruntime.DynamicAdvancedSession
	name              ModelName
	modelPath         string
	meta              *onnx.ModelInfo
	reduction         Reduction
	unsafeClassIndex  int
	neutralClassIndex int
	defaultThreshold  float32
	disabled          bool
	mean              [onnx.Channels]float32
	scales            [onnx.Channels]float32
	mutex             sync.Mutex
}

// Settings contains the artifact, preprocessing, and output semantics for a detector.
type Settings struct {
	Name              ModelName
	ModelPath         string
	Info              *onnx.ModelInfo
	Reduction         Reduction
	UnsafeClassIndex  int
	NeutralClassIndex int
	DefaultThreshold  float32
	Disabled          bool
}

// NewModel returns an ONNX NSFW detector instance.
func NewModel(settings Settings) *Model {
	info := cloneModelInfo(settings.Info)
	if info == nil {
		info = &onnx.ModelInfo{}
	}
	threshold := settings.DefaultThreshold
	if threshold <= 0 || threshold > 1 {
		threshold = DefaultThreshold
	}
	return &Model{name: NormalizeModelName(settings.Name), modelPath: settings.ModelPath, meta: info,
		reduction: settings.Reduction, unsafeClassIndex: settings.UnsafeClassIndex,
		neutralClassIndex: settings.NeutralClassIndex, defaultThreshold: threshold, disabled: settings.Disabled}
}

// NewRegisteredModel returns a registered detector rooted at modelsPath.
func NewRegisteredModel(modelsPath string, name ModelName, disabled bool) *Model {
	description := FindModel(name)
	if description == nil {
		return nil
	}
	modelDir := filepath.Join(modelsPath, string(description.Name))
	return NewModel(Settings{Name: description.Name, ModelPath: description.ONNX.FilePath(modelDir),
		Info: description.ONNX, Reduction: description.Reduction, UnsafeClassIndex: description.UnsafeClassIndex,
		NeutralClassIndex: description.NeutralClassIndex, DefaultThreshold: description.DefaultThreshold, Disabled: disabled})
}

// Init initializes the detector unless it is disabled.
func (m *Model) Init() error {
	if m == nil || m.disabled {
		return nil
	}
	return m.loadModel()
}

// Close releases the ONNX session.
func (m *Model) Close() error {
	if m == nil {
		return nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.session == nil {
		return nil
	}
	err := m.session.Destroy()
	m.session = nil
	return err
}

// DefaultThreshold returns the detector-specific fallback threshold.
func (m *Model) DefaultThreshold() float32 {
	if m == nil || m.defaultThreshold <= 0 || m.defaultThreshold > 1 {
		return DefaultThreshold
	}
	return m.defaultThreshold
}

// File returns the decision for a local image file.
func (m *Model) File(fileName string, threshold float32) (Result, error) {
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}
	data, err := os.ReadFile(fileName) //nolint:gosec // trusted callers intentionally select local media files
	if err != nil {
		return Unavailable(clean.Error(err)), err
	}
	return m.Run(data, threshold)
}

// Url returns the decision for an image at an HTTPS or data URL.
func (m *Model) Url(imgURL string, threshold float32) (Result, error) {
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}
	data, err := media.ReadUrlImage(imgURL, scheme.HttpsData)
	if err != nil {
		return Unavailable(clean.Error(err)), err
	}
	return m.Run(data, threshold)
}

// Run returns the decision for an encoded image.
func (m *Model) Run(data []byte, threshold float32) (result Result, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("nsfw: %s (inference panic)\nstack: %s", recovered, debug.Stack())
			result = Unavailable(clean.Error(err))
		}
	}()
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}
	probability, err := m.infer(data)
	if err != nil {
		return Unavailable(clean.Error(err)), err
	}
	if threshold <= 0 || threshold > 1 {
		threshold = m.DefaultThreshold()
	}
	result = NewResult(probability, threshold)
	if result.IsUnavailable() {
		return result, fmt.Errorf("nsfw: %s", result.Reason)
	}
	log.Tracef("nsfw: image classified as %+v", result)
	return result, nil
}

// ModelLoaded reports whether the ONNX session is initialized.
func (m *Model) ModelLoaded() bool {
	if m == nil {
		return false
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.session != nil
}

// infer executes the model and returns a validated unsafe probability.
func (m *Model) infer(data []byte) (float32, error) {
	if err := m.loadModel(); err != nil {
		return 0, err
	}
	img, _, err := fs.DecodeImageData(data)
	if err != nil {
		return 0, err
	}
	blob, err := m.buildBlob(img)
	if err != nil {
		return 0, err
	}
	inputTensor, err := onnxruntime.NewTensor(m.inputShape(), blob)
	if err != nil {
		return 0, fmt.Errorf("nsfw: create input tensor: %w", err)
	}
	defer func() {
		if destroyErr := inputTensor.Destroy(); destroyErr != nil {
			log.Debugf("nsfw: %s (destroy input tensor)", destroyErr)
		}
	}()
	outputs := []onnxruntime.Value{nil}
	m.mutex.Lock()
	if m.session == nil {
		m.mutex.Unlock()
		return 0, fmt.Errorf("nsfw: model was closed while running")
	}
	err = m.session.Run([]onnxruntime.Value{inputTensor}, outputs)
	m.mutex.Unlock()
	defer func() {
		if outputs[0] != nil {
			if destroyErr := outputs[0].Destroy(); destroyErr != nil {
				log.Debugf("nsfw: %s (destroy output tensor)", destroyErr)
			}
		}
	}()
	if err != nil {
		return 0, fmt.Errorf("nsfw: %s (run inference)", clean.Error(err))
	}
	if outputs[0] == nil {
		return 0, fmt.Errorf("nsfw: inference failed, no output")
	}
	outputTensor, ok := outputs[0].(*onnxruntime.Tensor[float32])
	if !ok {
		return 0, fmt.Errorf("nsfw: output tensor has unsupported type")
	}
	return m.reduceOutput(append([]float32(nil), outputTensor.GetData()...))
}

// reduceOutput converts model logits or probabilities to one unsafe probability.
func (m *Model) reduceOutput(values []float32) (float32, error) {
	if m == nil || m.meta == nil || m.meta.Output == nil {
		return 0, fmt.Errorf("nsfw: incomplete model output description")
	}
	if len(values) != m.meta.Output.Width {
		return 0, fmt.Errorf("nsfw: model returned %d values, expected %d", len(values), m.meta.Output.Width)
	}
	if m.reduction == ReductionSigmoidUnsafe {
		if len(values) != 1 {
			return 0, fmt.Errorf("nsfw: sigmoid reduction requires one output")
		}
		value := float64(values[0])
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, fmt.Errorf("nsfw: model returned a non-finite logit")
		}
		return float32(1 / (1 + math.Exp(-value))), nil
	}
	probabilities := values
	var err error
	if m.meta.Output.OutputsLogits() {
		probabilities, err = outputSoftmax(values)
	} else {
		err = validateOutputProbabilities(probabilities)
	}
	if err != nil {
		return 0, err
	}
	var score float32
	switch m.reduction {
	case ReductionSoftmaxUnsafe:
		if m.unsafeClassIndex < 0 || m.unsafeClassIndex >= len(probabilities) {
			return 0, fmt.Errorf("nsfw: unsafe class index %d is out of range", m.unsafeClassIndex)
		}
		score = probabilities[m.unsafeClassIndex]
	case ReductionNeutralComplement:
		if m.neutralClassIndex < 0 || m.neutralClassIndex >= len(probabilities) {
			return 0, fmt.Errorf("nsfw: neutral class index %d is out of range", m.neutralClassIndex)
		}
		score = 1 - probabilities[m.neutralClassIndex]
	default:
		return 0, fmt.Errorf("nsfw: unsupported output reduction %s", m.reduction)
	}
	if err = ValidateScore(score); err != nil {
		return 0, fmt.Errorf("nsfw: %w", err)
	}
	return score, nil
}

// loadModel initializes and validates the ONNX graph exactly once.
func (m *Model) loadModel() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.session != nil || m.disabled {
		return nil
	}
	if m.modelPath == "" {
		return fmt.Errorf("nsfw: model path is empty")
	}
	if _, err := os.Stat(m.modelPath); err != nil { //nolint:gosec // model configuration intentionally selects this local file
		return fmt.Errorf("nsfw: %w", err)
	}
	if err := m.meta.VerifyChecksum(m.modelPath); err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	if err := onnx.EnsureRuntime(""); err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	options, err := onnxruntime.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	defer func() {
		if destroyErr := options.Destroy(); destroyErr != nil {
			log.Debugf("nsfw: %s (destroy session options)", destroyErr)
		}
	}()
	if err = options.SetIntraOpNumThreads(max(runtime.NumCPU()/2, 1)); err != nil {
		return fmt.Errorf("nsfw: configure intra-op threads: %w", err)
	}
	if err = options.SetInterOpNumThreads(1); err != nil {
		return fmt.Errorf("nsfw: configure inter-op threads: %w", err)
	}
	if err = options.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		return fmt.Errorf("nsfw: optimize session graph: %w", err)
	}
	metadata, err := onnx.Metadata(m.modelPath)
	if err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	metadataInfo, err := onnx.InfoFromMetadata(metadata)
	if err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	m.meta.Merge(metadataInfo)
	graph, err := onnx.Inspect(m.modelPath, options)
	if err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	if graph.Output == nil {
		return fmt.Errorf("nsfw: model has no resolved output")
	}
	if graph.Output.Count != 1 {
		return fmt.Errorf("nsfw: model has %d outputs, expected 1", graph.Output.Count)
	}
	if err = m.meta.VerifyGraph(graph); err != nil {
		return fmt.Errorf("nsfw: %w", err)
	}
	m.meta.Merge(graph)
	onnx.CompleteResizeMetadata(m.meta)
	if err = m.validateDescription(); err != nil {
		return err
	}
	session, err := onnxruntime.NewDynamicAdvancedSession(m.modelPath, []string{m.meta.Input.Name}, []string{m.meta.Output.Name}, options)
	if err != nil {
		return fmt.Errorf("nsfw: initialize ONNX session: %w", err)
	}
	m.mean = m.meta.Input.Normalization.Mean
	m.scales = m.meta.Input.Normalization.Scales()
	m.session = session
	log.Infof("nsfw: loading %s with ONNX Runtime", clean.Log(string(m.name)))
	return nil
}

// validateDescription rejects incomplete preprocessing and output semantics.
func (m *Model) validateDescription() error {
	if m.meta == nil || m.meta.Input == nil || m.meta.Output == nil {
		return fmt.Errorf("nsfw: incomplete ONNX model description")
	}
	input := m.meta.Input
	if input.Width <= 0 || input.Height <= 0 {
		return fmt.Errorf("nsfw: invalid model input size %dx%d", input.Width, input.Height)
	}
	if input.Layout != onnx.LayoutNCHW && input.Layout != onnx.LayoutNHWC {
		return fmt.Errorf("nsfw: unsupported input layout %s", input.Layout)
	}
	if !input.ColorOrder.Valid() || input.Normalization.IsZero() || input.Resize.IsZero() {
		return fmt.Errorf("nsfw: model preprocessing is incomplete")
	}
	if m.meta.Output.Width <= 0 || m.meta.Output.Count != 1 || m.meta.Output.Logits == nil {
		return fmt.Errorf("nsfw: model output description is incomplete")
	}
	if m.reduction == "" {
		return fmt.Errorf("nsfw: model output reduction is missing")
	}
	if m.reduction == ReductionSoftmaxUnsafe && (m.unsafeClassIndex < 0 || m.unsafeClassIndex >= m.meta.Output.Width) {
		return fmt.Errorf("nsfw: unsafe class index %d is out of range", m.unsafeClassIndex)
	}
	if m.reduction == ReductionNeutralComplement && (m.neutralClassIndex < 0 || m.neutralClassIndex >= m.meta.Output.Width) {
		return fmt.Errorf("nsfw: neutral class index %d is out of range", m.neutralClassIndex)
	}
	return nil
}

// buildBlob converts an image into the detector's normalized tensor layout.
func (m *Model) buildBlob(img image.Image) ([]float32, error) {
	if img == nil || m.meta == nil || m.meta.Input == nil {
		return nil, fmt.Errorf("nsfw: invalid model input")
	}
	input := m.meta.Input
	img = resizeModelInput(img, input.Width, input.Height, input.Resize)
	if img.Bounds().Dx() != input.Width || img.Bounds().Dy() != input.Height {
		return nil, fmt.Errorf("nsfw: resized image is %dx%d, expected %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), input.Width, input.Height)
	}
	bounds := img.Bounds()
	planeSize := input.Width * input.Height
	blob := make([]float32, planeSize*onnx.Channels)
	rIndex, gIndex, bIndex := input.ColorOrder.Indices()
	for y := range input.Height {
		for x := range input.Width {
			red, green, blue, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			channels := [onnx.Channels]float32{(float32(red>>8) - m.mean[0]) * m.scales[0],
				(float32(green>>8) - m.mean[1]) * m.scales[1], (float32(blue>>8) - m.mean[2]) * m.scales[2]}
			pixel := y*input.Width + x
			if input.Layout == onnx.LayoutNHWC {
				base := pixel * onnx.Channels
				blob[base+rIndex], blob[base+gIndex], blob[base+bIndex] = channels[0], channels[1], channels[2]
			} else {
				blob[pixel+planeSize*rIndex], blob[pixel+planeSize*gIndex], blob[pixel+planeSize*bIndex] = channels[0], channels[1], channels[2]
			}
		}
	}
	return blob, nil
}

// inputShape returns the tensor shape matching the configured layout.
func (m *Model) inputShape() onnxruntime.Shape {
	if m.meta.Input.Layout == onnx.LayoutNHWC {
		return onnxruntime.Shape{1, int64(m.meta.Input.Height), int64(m.meta.Input.Width), onnx.Channels}
	}
	return onnxruntime.Shape{1, onnx.Channels, int64(m.meta.Input.Height), int64(m.meta.Input.Width)}
}

// resizeModelInput applies the declared resize convention.
func resizeModelInput(img image.Image, width, height int, resize onnx.Resize) image.Image {
	filter := modelResizeFilter(resize.Interpolation)
	switch resize.Mode {
	case onnx.ResizeStretch:
		return thumb.ResampleWithFilter(img, width, height, thumb.ResampleResize, filter)
	case onnx.ResizePad:
		fitted := thumb.ResampleWithFilter(img, width, height, thumb.ResampleFit, filter)
		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.NRGBA{A: 255}}, image.Point{}, draw.Src)
		offset := image.Pt((width-fitted.Bounds().Dx())/2, (height-fitted.Bounds().Dy())/2)
		draw.Draw(dst, image.Rectangle{Min: offset, Max: offset.Add(fitted.Bounds().Size())}, fitted, fitted.Bounds().Min, draw.Src)
		return dst
	case onnx.ResizeCenterCrop:
		shortEdge := resize.ShortEdge
		if shortEdge <= 0 {
			shortEdge = max(width, height)
		}
		bounds := img.Bounds()
		var resized image.Image
		if bounds.Dx() <= bounds.Dy() {
			resized = thumb.ResampleWithFilter(img, shortEdge, 0, thumb.ResampleResize, filter)
		} else {
			resized = thumb.ResampleWithFilter(img, 0, shortEdge, thumb.ResampleResize, filter)
		}
		return centerCrop(resized, width, height)
	default:
		return img
	}
}

// centerCrop returns a centered crop with the requested dimensions.
func centerCrop(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	left := bounds.Min.X + max((bounds.Dx()-width)/2, 0)
	top := bounds.Min.Y + max((bounds.Dy()-height)/2, 0)
	rect := image.Rect(left, top, left+min(width, bounds.Dx()), top+min(height, bounds.Dy())).Intersect(bounds)
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, image.Rect(0, 0, rect.Dx(), rect.Dy()), img, rect.Min, draw.Src)

	return dst
}

// modelResizeFilter maps ONNX interpolation names to thumbnail filters.
func modelResizeFilter(interpolation onnx.Interpolation) thumb.ResampleFilter {
	switch interpolation {
	case onnx.InterpolationNearest:
		return thumb.ResampleNearest
	case onnx.InterpolationLinear:
		return thumb.ResampleLinear
	case onnx.InterpolationLanczos:
		return thumb.ResampleLanczos
	default:
		return thumb.ResampleCubic
	}
}

// outputSoftmax converts finite logits to probabilities using stable maximum subtraction.
func outputSoftmax(values []float32) ([]float32, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("nsfw: model returned no output values")
	}
	maximum := values[0]
	for _, value := range values {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("nsfw: model returned a non-finite logit")
		}
		if value > maximum {
			maximum = value
		}
	}
	probabilities := make([]float32, len(values))
	var sum float64
	for i, value := range values {
		exp := math.Exp(float64(value - maximum))
		probabilities[i] = float32(exp)
		sum += exp
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return nil, fmt.Errorf("nsfw: model returned invalid logits")
	}
	for i := range probabilities {
		probabilities[i] = float32(float64(probabilities[i]) / sum)
	}
	return probabilities, validateOutputProbabilities(probabilities)
}

// validateOutputProbabilities verifies a complete probability distribution.
func validateOutputProbabilities(values []float32) error {
	if len(values) == 0 {
		return fmt.Errorf("nsfw: model returned no probabilities")
	}
	var sum float64
	for _, value := range values {
		if err := ValidateScore(value); err != nil {
			return fmt.Errorf("nsfw: %w", err)
		}
		sum += float64(value)
	}
	if math.Abs(sum-1) > 1e-4 {
		return fmt.Errorf("nsfw: probabilities sum to %.6f, expected 1", sum)
	}
	return nil
}

// cloneModelInfo returns an independent copy of info.
func cloneModelInfo(info *onnx.ModelInfo) *onnx.ModelInfo {
	if info == nil {
		return nil
	}
	clone := *info
	if info.Input != nil {
		input := *info.Input
		clone.Input = &input
	}
	if info.Output != nil {
		output := *info.Output
		if info.Output.Logits != nil {
			logits := *info.Output.Logits
			output.Logits = &logits
		}
		clone.Output = &output
	}
	return &clone
}
