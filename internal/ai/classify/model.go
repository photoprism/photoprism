package classify

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// Model represents an ONNX image classification model.
type Model struct {
	session        *onnxruntime.DynamicAdvancedSession
	name           ModelName
	modelPath      string
	labelPath      string
	labels         []string
	disabled       bool
	canonicalOrder bool
	meta           *onnx.ModelInfo
	mean           [onnx.Channels]float32
	scales         [onnx.Channels]float32
	mutex          sync.Mutex
}

// Settings contains the paths and preprocessing contract used to load a classifier.
type Settings struct {
	Name           ModelName
	ModelPath      string
	LabelPath      string
	Info           *onnx.ModelInfo
	CanonicalOrder bool
	Disabled       bool
}

// NewModel returns a new ONNX classification model instance.
func NewModel(settings Settings) *Model {
	info := cloneModelInfo(settings.Info)

	if info == nil {
		info = &onnx.ModelInfo{}
	}

	return &Model{
		name:           NormalizeModelName(settings.Name),
		modelPath:      settings.ModelPath,
		labelPath:      settings.LabelPath,
		meta:           info,
		canonicalOrder: settings.CanonicalOrder,
		disabled:       settings.Disabled,
	}
}

// NewRegisteredModel returns the registered classifier with the specified name.
func NewRegisteredModel(modelsPath string, name ModelName, disabled bool) *Model {
	description := FindModel(name)

	if description == nil {
		return nil
	}

	modelDir := filepath.Join(modelsPath, string(description.Name))

	return NewModel(Settings{
		Name:           description.Name,
		ModelPath:      description.ONNX.FilePath(modelDir),
		LabelPath:      filepath.Join(modelDir, description.LabelFile),
		Info:           description.ONNX,
		CanonicalOrder: description.CanonicalOrder,
		Disabled:       disabled,
	})
}

// Init initializes the classifier unless it is disabled.
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

// File returns matching labels for a local image file.
func (m *Model) File(fileName string, confidenceThreshold int) (Labels, error) {
	if m == nil || m.disabled {
		return nil, nil
	}

	data, err := os.ReadFile(fileName) //nolint:gosec // trusted callers intentionally select local media files
	if err != nil {
		return nil, err
	}

	return m.Run(data, confidenceThreshold)
}

// Url returns matching labels for a remote image.
func (m *Model) Url(imgURL string, confidenceThreshold int) (Labels, error) {
	if m == nil || m.disabled {
		return nil, nil
	}

	data, err := media.ReadUrlImage(imgURL, scheme.HttpsData)
	if err != nil {
		return nil, err
	}

	return m.Run(data, confidenceThreshold)
}

// Run returns matching labels for the specified encoded image.
func (m *Model) Run(data []byte, confidenceThreshold int) (result Labels, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("classify: %s (inference panic)\nstack: %s", recovered, debug.Stack())
		}
	}()

	if m == nil || m.disabled {
		return result, nil
	}

	probabilities, err := m.infer(data)
	if err != nil {
		return nil, err
	}

	result = m.bestLabels(probabilities, confidenceThreshold)
	if len(result) == 0 {
		return Labels{}, nil
	}

	log.Tracef("classify: image classified as %+v", result)

	return result, nil
}

// infer returns the complete validated probability vector for an encoded image.
func (m *Model) infer(data []byte) ([]float32, error) {
	if m == nil || m.disabled {
		return nil, nil
	}

	if err := m.loadModel(); err != nil {
		return nil, err
	}

	img, _, err := fs.DecodeImageData(data)
	if err != nil {
		return nil, err
	}

	blob, err := m.buildBlob(img)
	if err != nil {
		return nil, err
	}

	inputTensor, err := onnxruntime.NewTensor(m.inputShape(), blob)
	if err != nil {
		return nil, fmt.Errorf("classify: create input tensor: %w", err)
	}

	defer func() {
		if destroyErr := inputTensor.Destroy(); destroyErr != nil {
			log.Debugf("classify: %s (destroy input tensor)", destroyErr)
		}
	}()

	outputs := []onnxruntime.Value{nil}

	m.mutex.Lock()
	if m.session == nil {
		m.mutex.Unlock()
		return nil, fmt.Errorf("classify: model was closed while running")
	}

	err = m.session.Run([]onnxruntime.Value{inputTensor}, outputs)
	m.mutex.Unlock()

	defer func() {
		if outputs[0] != nil {
			if destroyErr := outputs[0].Destroy(); destroyErr != nil {
				log.Debugf("classify: %s (destroy output tensor)", destroyErr)
			}
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("classify: %s (run inference)", clean.Error(err))
	}

	if outputs[0] == nil {
		return nil, fmt.Errorf("classify: inference failed, no output")
	}

	outputTensor, ok := outputs[0].(*onnxruntime.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("classify: output tensor has unsupported type")
	}

	probabilities := append([]float32(nil), outputTensor.GetData()...)
	if m.meta.Output.OutputsLogits() {
		if probabilities, err = softmax(probabilities); err != nil {
			return nil, err
		}
	} else if err = validateProbabilities(probabilities); err != nil {
		return nil, err
	}

	if len(probabilities) != len(m.labels) {
		return nil, fmt.Errorf("classify: model returned %d values for %d labels", len(probabilities), len(m.labels))
	}

	return probabilities, nil
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

// loadModel initializes and validates the ONNX graph exactly once.
func (m *Model) loadModel() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.session != nil || m.disabled {
		return nil
	}

	if m.modelPath == "" {
		return fmt.Errorf("classify: model path is empty")
	}

	if _, err := os.Stat(m.modelPath); err != nil { //nolint:gosec // model configuration intentionally selects this local file
		return fmt.Errorf("classify: %w", err)
	}

	if err := m.meta.VerifyChecksum(m.modelPath); err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	if err := onnx.EnsureRuntime(""); err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	sessionOptions, err := onnxruntime.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	defer func() {
		if destroyErr := sessionOptions.Destroy(); destroyErr != nil {
			log.Debugf("classify: %s (destroy session options)", destroyErr)
		}
	}()

	threads := max(runtime.NumCPU()/2, 1)
	if err = sessionOptions.SetIntraOpNumThreads(threads); err != nil {
		return fmt.Errorf("classify: configure intra-op threads: %w", err)
	}

	if err = sessionOptions.SetInterOpNumThreads(1); err != nil {
		return fmt.Errorf("classify: configure inter-op threads: %w", err)
	}

	if err = sessionOptions.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		return fmt.Errorf("classify: optimize session graph: %w", err)
	}

	metadata, err := onnx.Metadata(m.modelPath)
	if err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	metadataInfo, err := onnx.InfoFromMetadata(metadata)
	if err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	m.meta.Merge(metadataInfo)

	graph, err := onnx.Inspect(m.modelPath, sessionOptions)
	if err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	if graph.Output == nil {
		return fmt.Errorf("classify: model has no resolved output")
	} else if graph.Output.Count != 1 {
		return fmt.Errorf("classify: model has %d outputs, expected 1", graph.Output.Count)
	}

	if err = m.meta.VerifyGraph(graph); err != nil {
		return fmt.Errorf("classify: %w", err)
	}

	m.meta.Merge(graph)
	onnx.CompleteResizeMetadata(m.meta)
	if err = m.applyDefaults(); err != nil {
		return err
	}

	if err = m.loadLabels(); err != nil {
		return err
	}

	if err = m.validateLabels(); err != nil {
		return err
	}

	session, err := onnxruntime.NewDynamicAdvancedSession(
		m.modelPath,
		[]string{m.meta.Input.Name},
		[]string{m.meta.Output.Name},
		sessionOptions,
	)
	if err != nil {
		return fmt.Errorf("classify: initialize ONNX session: %w", err)
	}

	m.mean = m.meta.Input.Normalization.Mean
	m.scales = m.meta.Input.Normalization.Scales()
	m.session = session

	log.Infof("classify: loading %s with ONNX Runtime from %s", clean.Log(string(m.name)), clean.Log(onnx.RuntimeLibraryPath()))

	return nil
}

// validateLabels rejects output widths, vocabularies, and canonical orders that cannot align.
func (m *Model) validateLabels() error {
	if m == nil || m.meta == nil || m.meta.Output == nil {
		return fmt.Errorf("classify: incomplete model output description")
	}
	if m.meta.Output.Width != len(m.labels) {
		return fmt.Errorf("classify: model output width %d does not match %d labels", m.meta.Output.Width, len(m.labels))
	}
	if !m.canonicalOrder {
		return nil
	}
	if len(m.labels) != ImageNetClasses {
		return fmt.Errorf("classify: canonical ImageNet model has %d labels, expected %d", len(m.labels), ImageNetClasses)
	}
	if strings.EqualFold(strings.TrimSpace(m.labels[0]), "background") {
		return fmt.Errorf("classify: canonical ImageNet model contains a background class")
	}
	return nil
}

// applyDefaults fills semantic values that cannot be inferred from an ONNX graph.
func (m *Model) applyDefaults() error {
	if m.meta.Input == nil || m.meta.Output == nil {
		return fmt.Errorf("classify: incomplete ONNX model description")
	}

	input := m.meta.Input
	if input.Width <= 0 || input.Height <= 0 {
		return fmt.Errorf("classify: invalid model input size %dx%d", input.Width, input.Height)
	}

	if input.Layout == onnx.LayoutUndefined {
		input.Layout = onnx.LayoutNCHW
		log.Warnf("classify: model input layout is not declared, assuming NCHW")
	}

	if input.Layout != onnx.LayoutNCHW && input.Layout != onnx.LayoutNHWC {
		return fmt.Errorf("classify: unsupported input layout %s", input.Layout)
	}

	if !input.ColorOrder.Valid() {
		input.ColorOrder = onnx.RGB
		log.Warnf("classify: model channel order is not declared, assuming RGB")
	}

	if input.Normalization.IsZero() {
		input.Normalization = imageNetNormalization()
		log.Warnf("classify: model normalization is not declared, assuming ImageNet mean and standard deviation")
	}

	if input.Resize.IsZero() {
		input.Resize = onnx.Resize{
			Mode:          onnx.ResizeCenterCrop,
			ShortEdge:     imageNetShortEdge(input.Width, input.Height),
			Interpolation: onnx.InterpolationBicubic,
		}
		log.Warnf("classify: model resize convention is not declared, assuming an ImageNet center crop")
	}

	if m.meta.Output.Width <= 0 {
		return fmt.Errorf("classify: model output width is dynamic")
	}

	if m.meta.Output.Logits == nil {
		m.meta.Output.Logits = onnx.Bool(true)
		log.Warnf("classify: model output type is not declared, assuming raw logits")
	}

	if m.meta.Output.Count <= 0 {
		m.meta.Output.Count = 1
	}

	return nil
}

// loadLabels loads and validates the model's explicitly selected label file.
func (m *Model) loadLabels() error {
	if m.labelPath == "" {
		return fmt.Errorf("classify: label file path is empty")
	}

	labels, err := readLabels(m.labelPath)
	if err != nil {
		return fmt.Errorf("classify: could not load labels: %w", err)
	}
	if len(labels) != m.meta.Output.Width {
		return fmt.Errorf("classify: %s contains %d labels, expected %d", filepath.Base(m.labelPath), len(labels), m.meta.Output.Width)
	}

	m.labels = labels
	log.Infof("classify: loading model labels from %s", clean.Log(m.labelPath))

	return nil
}

// readLabels returns the newline-separated labels in fileName.
func readLabels(fileName string) ([]string, error) {
	file, err := os.Open(fileName) //nolint:gosec // model configuration intentionally selects this local file
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Debugf("classify: %s (close labels file)", closeErr)
		}
	}()

	labels := make([]string, 0, ImageNetClasses)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}

// bestLabels returns at most five labels that pass the configured thresholds.
func (m *Model) bestLabels(probabilities []float32, confidenceThreshold int) Labels {
	result := make(Labels, 0, 5)

	for i, probability := range probabilities {
		if i >= len(m.labels) {
			break
		}

		confidence := int(math.Round(float64(probability * 100)))
		if confidence < confidenceThreshold {
			continue
		}

		labelText := strings.ToLower(m.labels[i])
		rule, _ := Rules.Find(labelText)

		if probability < rule.Threshold {
			continue
		}

		if rule.Label != "" {
			labelText = rule.Label
		}

		labelText = strings.TrimSpace(labelText)
		result = append(result, Label{
			Name:        labelText,
			Source:      SrcImage,
			Uncertainty: 100 - confidence,
			Priority:    rule.Priority,
			Categories:  rule.Categories,
		})
	}

	sort.Sort(result)
	if len(result) > 5 {
		return result[:5]
	}

	return result
}

// buildBlob converts an image into the model's normalized tensor layout.
func (m *Model) buildBlob(img image.Image) ([]float32, error) {
	if img == nil || m.meta == nil || m.meta.Input == nil {
		return nil, fmt.Errorf("classify: invalid model input")
	}

	input := m.meta.Input
	img = resizeInput(img, input.Width, input.Height, input.Resize)

	if img.Bounds().Dx() != input.Width || img.Bounds().Dy() != input.Height {
		return nil, fmt.Errorf("classify: resized image is %dx%d, expected %dx%d", img.Bounds().Dx(), img.Bounds().Dy(), input.Width, input.Height)
	}

	bounds := img.Bounds()
	planeSize := input.Width * input.Height
	blob := make([]float32, planeSize*onnx.Channels)
	rIndex, gIndex, bIndex := input.ColorOrder.Indices()

	for y := range input.Height {
		for x := range input.Width {
			red, green, blue, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			channels := [onnx.Channels]float32{
				(float32(red>>8) - m.mean[0]) * m.scales[0],
				(float32(green>>8) - m.mean[1]) * m.scales[1],
				(float32(blue>>8) - m.mean[2]) * m.scales[2],
			}

			pixel := y*input.Width + x
			if input.Layout == onnx.LayoutNHWC {
				base := pixel * onnx.Channels
				blob[base+rIndex] = channels[0]
				blob[base+gIndex] = channels[1]
				blob[base+bIndex] = channels[2]
			} else {
				blob[pixel+planeSize*rIndex] = channels[0]
				blob[pixel+planeSize*gIndex] = channels[1]
				blob[pixel+planeSize*bIndex] = channels[2]
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

// resizeInput applies the model-specific resize and crop convention.
func resizeInput(img image.Image, width, height int, resize onnx.Resize) image.Image {
	filter := resizeFilter(resize.Interpolation)

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
			shortEdge = imageNetShortEdge(width, height)
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
		return thumb.ResampleWithFilter(img, width, height, thumb.ResampleFillCenter, filter)
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

// resizeFilter maps ONNX preprocessing names to PhotoPrism's resampling filters.
func resizeFilter(interpolation onnx.Interpolation) thumb.ResampleFilter {
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

// softmax converts raw logits to finite probabilities with a stable maximum subtraction.
func softmax(logits []float32) ([]float32, error) {
	if len(logits) == 0 {
		return nil, fmt.Errorf("classify: model returned no logits")
	}

	maxLogit := float64(logits[0])
	for _, logit := range logits {
		value := float64(logit)
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("classify: model returned a non-finite logit")
		}
		if value > maxLogit {
			maxLogit = value
		}
	}

	probabilities := make([]float32, len(logits))
	var sum float64
	for i, logit := range logits {
		value := math.Exp(float64(logit) - maxLogit)
		probabilities[i] = float32(value)
		sum += value
	}

	if sum == 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return nil, fmt.Errorf("classify: softmax normalization failed")
	}

	for i := range probabilities {
		probabilities[i] = float32(float64(probabilities[i]) / sum)
	}

	if err := validateProbabilities(probabilities); err != nil {
		return nil, err
	}

	return probabilities, nil
}

// validateProbabilities rejects non-finite or unnormalized model output.
func validateProbabilities(probabilities []float32) error {
	if len(probabilities) == 0 {
		return fmt.Errorf("classify: model returned no probabilities")
	}

	var sum float64
	for _, probability := range probabilities {
		value := float64(probability)
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return fmt.Errorf("classify: model returned an invalid probability")
		}
		sum += value
	}

	if math.Abs(sum-1) > 1e-4 {
		return fmt.Errorf("classify: probabilities sum to %.8f, expected 1", sum)
	}

	return nil
}

// imageNetShortEdge derives the conventional 87.5 percent center-crop edge.
func imageNetShortEdge(width, height int) int {
	return int(math.Round(float64(max(width, height)) / 0.875))
}

// cloneModelInfo returns a private deep copy of an ONNX model description.
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
