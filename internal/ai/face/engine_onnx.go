package face

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	onnxruntime "github.com/yalue/onnxruntime_go"
)

// ONNXOptions configures how the ONNX runtime-backed detector is initialised.
type ONNXOptions struct {
	ModelPath      string
	LibraryPath    string
	Threads        int
	ScoreThreshold float32
	NMSThreshold   float32
}

const (
	DefaultONNXModelFilename  = "scrfd.onnx"
	onnxDefaultScoreThreshold = 0.50
	onnxDefaultNMSThreshold   = 0.40
	onnxDefaultInputSize      = 640
	onnxInputMean             = 127.5
	onnxInputStd              = 128.0
)

// anchorCacheKey uniquely identifies cached anchor center grids.
type anchorCacheKey struct {
	height  int
	width   int
	stride  int
	anchors int
}

// onnxEngine runs face detection using an ONNX Runtime session and SCRFD model.
type onnxEngine struct {
	session        *onnxruntime.DynamicAdvancedSession
	inputName      string
	outputNames    []string
	inputWidth     int
	inputHeight    int
	featStrides    []int
	numAnchors     int
	useKps         bool
	batched        bool
	scoreThreshold float32
	nmsThreshold   float32
	centerMu       sync.Mutex
	centerCache    map[anchorCacheKey][]float32
}

var (
	onnxOnce          sync.Once
	onnxInitErr       error
	onnxExecutableVar = os.Executable
)

// ensureONNXRuntime loads the ONNX runtime shared library and initializes the global environment.
func ensureONNXRuntime(libraryPath string) error {
	onnxOnce.Do(func() {
		candidates := onnxSharedLibraryCandidates(libraryPath)
		var errs []string

		for _, candidate := range candidates {
			onnxruntime.SetSharedLibraryPath(candidate)

			if err := onnxruntime.InitializeEnvironment(); err != nil {
				// Collect errors so we can surface meaningful diagnostics when all options fail.
				errs = append(errs, fmt.Sprintf("%s (%v)", candidate, err))
				continue
			}

			// Successfully initialized; stop retrying.
			onnxInitErr = nil
			return
		}

		if len(errs) == 0 {
			onnxInitErr = errors.New("faces: no ONNX runtime library candidates")
			return
		}

		onnxInitErr = fmt.Errorf("faces: failed to load ONNX runtime: %s", strings.Join(errs, "; "))
	})

	return onnxInitErr
}

// onnxSharedLibraryCandidates lists library paths to try when loading the ONNX runtime.
func onnxSharedLibraryCandidates(explicit string) []string {
	appendUnique := func(list []string, seen map[string]struct{}, values ...string) []string {
		for _, value := range values {
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			list = append(list, value)
			seen[value] = struct{}{}
		}
		return list
	}

	seen := make(map[string]struct{})
	candidates := make([]string, 0, 8)
	candidates = appendUnique(candidates, seen, explicit)
	candidates = appendUnique(candidates, seen,
		"libonnxruntime.so",
		"libonnxruntime.so.1",
		"onnxruntime.so",
	)

	if exePath, err := onnxExecutableVar(); err == nil {
		exeDir := filepath.Dir(exePath)
		rootDir := filepath.Dir(exeDir)

		candidates = appendUnique(candidates, seen,
			filepath.Join(exeDir, "libonnxruntime.so"),
			filepath.Join(exeDir, "lib", "libonnxruntime.so"),
		)

		if rootDir != "" && rootDir != "." && rootDir != exeDir {
			candidates = appendUnique(candidates, seen, filepath.Join(rootDir, "lib", "libonnxruntime.so"))
		}
	}

	return candidates
}

// NewONNXEngine loads the SCRFD model and returns an ONNX-backed DetectionEngine.
func NewONNXEngine(opts ONNXOptions) (DetectionEngine, error) {
	if opts.ModelPath == "" {
		return nil, fmt.Errorf("faces: missing ONNX model path")
	}

	if _, err := os.Stat(opts.ModelPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	if opts.ScoreThreshold <= 0 {
		opts.ScoreThreshold = onnxDefaultScoreThreshold
	}

	if opts.NMSThreshold <= 0 {
		opts.NMSThreshold = onnxDefaultNMSThreshold
	}

	if err := ensureONNXRuntime(opts.LibraryPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	sessionOpts, err := onnxruntime.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}
	defer sessionOpts.Destroy()

	threads := opts.Threads
	if threads == 0 {
		threads = runtime.NumCPU() / 2
		if threads < 1 {
			threads = 1
		}
	}

	if err := sessionOpts.SetIntraOpNumThreads(threads); err != nil {
		return nil, fmt.Errorf("faces: configure intra-op threads: %w", err)
	}

	if err := sessionOpts.SetInterOpNumThreads(threads); err != nil {
		return nil, fmt.Errorf("faces: configure inter-op threads: %w", err)
	}

	if err := sessionOpts.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		return nil, fmt.Errorf("faces: optimize session graph: %w", err)
	}

	inputInfos, outputInfos, err := onnxruntime.GetInputOutputInfoWithOptions(opts.ModelPath, sessionOpts)
	if err != nil {
		return nil, fmt.Errorf("faces: load ONNX metadata: %w", err)
	}

	if len(inputInfos) == 0 {
		return nil, fmt.Errorf("faces: ONNX model has no inputs")
	}

	if len(outputInfos) == 0 {
		return nil, fmt.Errorf("faces: ONNX model has no outputs")
	}

	inputName := inputInfos[0].Name
	inputDims := inputInfos[0].Dimensions

	width := onnxDefaultInputSize
	height := onnxDefaultInputSize

	if len(inputDims) >= 4 {
		if w := int(inputDims[len(inputDims)-1]); w > 0 {
			width = w
		}
		if h := int(inputDims[len(inputDims)-2]); h > 0 {
			height = h
		}
	}

	outputNames := make([]string, len(outputInfos))
	for i, out := range outputInfos {
		outputNames[i] = out.Name
	}

	fmc, numAnchors, useKps, batched, err := deriveONNXLayout(outputInfos)
	if err != nil {
		return nil, err
	}

	featStrides := stridesForFeatureMaps(fmc)

	session, err := onnxruntime.NewDynamicAdvancedSession(opts.ModelPath, []string{inputName}, outputNames, sessionOpts)
	if err != nil {
		return nil, fmt.Errorf("faces: initialise ONNX session: %w", err)
	}

	engine := &onnxEngine{
		session:        session,
		inputName:      inputName,
		outputNames:    outputNames,
		inputWidth:     width,
		inputHeight:    height,
		featStrides:    featStrides,
		numAnchors:     numAnchors,
		useKps:         useKps,
		batched:        batched,
		scoreThreshold: opts.ScoreThreshold,
		nmsThreshold:   opts.NMSThreshold,
		centerCache:    make(map[anchorCacheKey][]float32),
	}

	return engine, nil
}

// deriveONNXLayout infers the number of feature map chains, anchors, and output layout from the model outputs.
func deriveONNXLayout(outputs []onnxruntime.InputOutputInfo) (fmc, anchors int, useKps, batched bool, err error) {
	outCount := len(outputs)

	switch outCount {
	case 6:
		fmc = 3
		anchors = 2
	case 9:
		fmc = 3
		anchors = 2
		useKps = true
	case 10:
		fmc = 5
		anchors = 1
	case 15:
		fmc = 5
		anchors = 1
		useKps = true
	default:
		return 0, 0, false, false, fmt.Errorf("faces: unsupported ONNX output count %d", outCount)
	}

	dims := outputs[0].Dimensions
	if len(dims) == 3 {
		batched = true
	}

	return fmc, anchors, useKps, batched, nil
}

// stridesForFeatureMaps returns SCRFD's default strides for the given number of feature maps.
func stridesForFeatureMaps(fmc int) []int {
	if fmc == 5 {
		return []int{8, 16, 32, 64, 128}
	}

	return []int{8, 16, 32}
}

func (o *onnxEngine) Name() string {
	return EngineONNX
}

func (o *onnxEngine) Close() error {
	if o.session != nil {
		if err := o.session.Destroy(); err != nil {
			return err
		}
		o.session = nil
	}

	return nil
}

// Detect identifies faces in the provided image using the ONNX runtime session.
func (o *onnxEngine) Detect(fileName string, findLandmarks bool, minSize int) (Faces, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return Faces{}, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return Faces{}, err
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width == 0 || height == 0 {
		return Faces{}, fmt.Errorf("faces: invalid image dimensions")
	}

	blob, detScale, err := o.buildBlob(img)
	if err != nil {
		return Faces{}, err
	}

	shape := onnxruntime.Shape{1, 3, int64(o.inputHeight), int64(o.inputWidth)}
	tensor, err := onnxruntime.NewTensor(shape, blob)
	if err != nil {
		return Faces{}, fmt.Errorf("faces: create tensor: %w", err)
	}
	defer tensor.Destroy()

	inputs := []onnxruntime.Value{tensor}
	outputs := make([]onnxruntime.Value, len(o.outputNames))
	if err := o.session.Run(inputs, outputs); err != nil {
		return Faces{}, fmt.Errorf("faces: run session: %w", err)
	}
	for _, out := range outputs {
		if out != nil {
			defer out.Destroy()
		}
	}

	detections, err := o.parseDetections(outputs, detScale, width, height)
	if err != nil {
		return Faces{}, err
	}

	filtered := nonMaxSuppression(detections, o.nmsThreshold)
	result := make(Faces, 0, len(filtered))

	for _, det := range filtered {
		faceWidth := det.x2 - det.x1
		faceHeight := det.y2 - det.y1
		size := int(math.Max(float64(faceWidth), float64(faceHeight)))
		if size < minSize {
			continue
		}

		row := int((det.y1 + det.y2) * 0.5)
		col := int((det.x1 + det.x2) * 0.5)
		score := int(math.Round(float64(det.score * 100)))
		if score > 100 {
			score = 100
		} else if score < 0 {
			score = 0
		}

		f := Face{
			Rows:  height,
			Cols:  width,
			Score: score,
			Area:  NewArea("face", row, col, size),
		}

		result.Append(f)
	}

	return result, nil
}

// buildBlob normalises the input image into the tensor layout expected by SCRFD.
func (o *onnxEngine) buildBlob(img image.Image) ([]float32, float32, error) {
	inputWidth := o.inputWidth
	inputHeight := o.inputHeight

	if inputWidth < 1 {
		inputWidth = onnxDefaultInputSize
	}

	if inputHeight < 1 {
		inputHeight = onnxDefaultInputSize
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return nil, 0, fmt.Errorf("faces: invalid image dimensions")
	}

	imRatio := float32(height) / float32(width)
	modelRatio := float32(inputHeight) / float32(inputWidth)

	newHeight := inputHeight
	newWidth := inputWidth
	if imRatio > modelRatio {
		newHeight = inputHeight
		newWidth = int(float32(newHeight) / imRatio)
	} else {
		newWidth = inputWidth
		newHeight = int(float32(newWidth) * imRatio)
	}

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	resized := imaging.Resize(img, newWidth, newHeight, imaging.Linear)

	planeSize := inputWidth * inputHeight
	blob := make([]float32, planeSize*3)

	for y := 0; y < inputHeight; y++ {
		for x := 0; x < inputWidth; x++ {
			idx := y*inputWidth + x
			var r, g, b float32
			if x < newWidth && y < newHeight {
				cr, cg, cb, _ := resized.At(x, y).RGBA()
				r = float32(uint8(cr >> 8))
				g = float32(uint8(cg >> 8))
				b = float32(uint8(cb >> 8))
			}

			blob[idx] = (r - onnxInputMean) / onnxInputStd
			blob[idx+planeSize] = (g - onnxInputMean) / onnxInputStd
			blob[idx+planeSize*2] = (b - onnxInputMean) / onnxInputStd
		}
	}

	detScale := float32(newHeight) / float32(height)

	return blob, detScale, nil
}

// parseDetections decodes model outputs into bounding boxes in the original image space.
func (o *onnxEngine) parseDetections(values []onnxruntime.Value, detScale float32, origWidth, origHeight int) ([]onnxDetection, error) {
	fmc := len(o.featStrides)
	detections := make([]onnxDetection, 0, 32)

	for level, stride := range o.featStrides {
		scoreTensor, ok := values[level].(*onnxruntime.Tensor[float32])
		if !ok {
			return nil, fmt.Errorf("faces: unexpected tensor type for scores")
		}

		bboxTensor, ok := values[level+fmc].(*onnxruntime.Tensor[float32])
		if !ok {
			return nil, fmt.Errorf("faces: unexpected tensor type for boxes")
		}

		scores := scoreTensor.GetData()
		boxes := bboxTensor.GetData()

		height := o.inputHeight / stride
		width := o.inputWidth / stride
		cells := height * width
		anchors := o.numAnchors
		expected := cells * anchors

		switch {
		case len(scores) == expected:
			// already aligned
		case len(scores) == expected*2:
			trimmed := make([]float32, expected)
			copy(trimmed, scores[len(scores)-expected:])
			scores = trimmed
		default:
			return nil, fmt.Errorf("faces: unexpected score tensor size %d (expected %d)", len(scores), expected)
		}

		if len(boxes) != expected*4 {
			return nil, fmt.Errorf("faces: mismatch between scores and boxes")
		}

		centers := o.anchorCenters(height, width, stride, anchors)

		for idx, score := range scores {
			if score < o.scoreThreshold {
				continue
			}

			cx := centers[idx*2]
			cy := centers[idx*2+1]
			boxOffset := idx * 4
			left := boxes[boxOffset] * float32(stride)
			top := boxes[boxOffset+1] * float32(stride)
			right := boxes[boxOffset+2] * float32(stride)
			bottom := boxes[boxOffset+3] * float32(stride)

			x1 := clampFloat32((cx-left)/detScale, 0, float32(origWidth))
			y1 := clampFloat32((cy-top)/detScale, 0, float32(origHeight))
			x2 := clampFloat32((cx+right)/detScale, 0, float32(origWidth))
			y2 := clampFloat32((cy+bottom)/detScale, 0, float32(origHeight))

			if x2 <= x1 || y2 <= y1 {
				continue
			}

			detections = append(detections, onnxDetection{
				x1:    x1,
				y1:    y1,
				x2:    x2,
				y2:    y2,
				score: score,
			})
		}
	}

	return detections, nil
}

// anchorCenters returns cached anchor centers for the given feature map shape.
func (o *onnxEngine) anchorCenters(height, width, stride, anchors int) []float32 {
	key := anchorCacheKey{height: height, width: width, stride: stride, anchors: anchors}

	o.centerMu.Lock()
	cached, ok := o.centerCache[key]
	if ok {
		o.centerMu.Unlock()
		return cached
	}

	centers := make([]float32, height*width*anchors*2)
	idx := 0
	for y := 0; y < height; y++ {
		cy := float32(y * stride)
		for x := 0; x < width; x++ {
			cx := float32(x * stride)
			for a := 0; a < anchors; a++ {
				centers[idx] = cx
				centers[idx+1] = cy
				idx += 2
			}
		}
	}

	o.centerCache[key] = centers
	o.centerMu.Unlock()
	return centers
}

// onnxDetection stores a single detection candidate in image coordinates.
type onnxDetection struct {
	x1    float32
	y1    float32
	x2    float32
	y2    float32
	score float32
}

// nonMaxSuppression filters overlapping detection boxes using IoU thresholding.
func nonMaxSuppression(boxes []onnxDetection, threshold float32) []onnxDetection {
	if len(boxes) == 0 {
		return nil
	}

	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].score > boxes[j].score
	})

	picked := make([]onnxDetection, 0, len(boxes))
	suppressed := make([]bool, len(boxes))

	for i := 0; i < len(boxes); i++ {
		if suppressed[i] {
			continue
		}

		current := boxes[i]
		picked = append(picked, current)

		for j := i + 1; j < len(boxes); j++ {
			if suppressed[j] {
				continue
			}

			if iou(current, boxes[j]) > threshold {
				suppressed[j] = true
			}
		}
	}

	return picked
}

// iou calculates the intersection-over-union score for two detections.
func iou(a, b onnxDetection) float32 {
	x1 := float32(math.Max(float64(a.x1), float64(b.x1)))
	y1 := float32(math.Max(float64(a.y1), float64(b.y1)))
	x2 := float32(math.Min(float64(a.x2), float64(b.x2)))
	y2 := float32(math.Min(float64(a.y2), float64(b.y2)))

	w := x2 - x1
	h := y2 - y1
	if w <= 0 || h <= 0 {
		return 0
	}

	inter := w * h
	areaA := (a.x2 - a.x1) * (a.y2 - a.y1)
	areaB := (b.x2 - b.x1) * (b.y2 - b.y1)
	union := areaA + areaB - inter
	if union <= 0 {
		return 0
	}

	return inter / union
}

// clampFloat32 bounds v to the inclusive range [min, max].
func clampFloat32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
