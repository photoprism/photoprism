package face

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder for ONNX engine input
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"
	xdraw "golang.org/x/image/draw"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ONNXOptions configures how the ONNX runtime-backed detector is initialized.
type ONNXOptions struct {
	ModelPath      string
	LibraryPath    string
	Threads        int
	ScoreThreshold float32
	NMSThreshold   float32
}

const (
	// DefaultONNXModelFilename is the bundled ONNX model name used when none is provided.
	DefaultONNXModelFilename  = "scrfd.onnx"
	onnxDefaultScoreThreshold = 0.50
	onnxDefaultNMSThreshold   = 0.40

	// maxDetectorInputSize bounds the input geometry a detector graph may declare. The
	// detector runs on a thumbnail and the blob is sized from these values, so an axis
	// beyond this describes a graph that cannot be run rather than a wider export.
	maxDetectorInputSize = 4096
)

// detectorInputSize returns the geometry to run the specified detector at, given what its
// graph declares. A dynamic axis reports zero and falls back to the registered size; an axis
// larger than a detector input can be is refused, because the blob is sized from it.
func detectorInputSize(detector *Detector, graphWidth, graphHeight int) (width, height int, err error) {
	defaultWidth, defaultHeight := detector.ONNX.InputSize()

	width, height = graphWidth, graphHeight

	if width < 1 {
		width = defaultWidth
	}

	if height < 1 {
		height = defaultHeight
	}

	if width > maxDetectorInputSize || height > maxDetectorInputSize {
		return 0, 0, fmt.Errorf("faces: detector input %dx%d exceeds %d", width, height, maxDetectorInputSize)
	}

	return width, height, nil
}

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
	colorOrder     onnx.ColorOrder
	mean           [onnx.Channels]float32
	scales         [onnx.Channels]float32
	featStrides    []int
	numAnchors     int
	batched        bool
	useKps         bool
	decode         DecodeKind
	detector       DetectorName
	scoreThreshold float32
	nmsThreshold   float32
	sessionMu      sync.Mutex
	centerMu       sync.Mutex
	centerCache    map[anchorCacheKey][]float32
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

	// Which detector this is decides the channel order and normalization, which cannot be read
	// from a graph. An unknown artifact falls back to the default rather than refusing, so a
	// re-export still runs, and its layout is still derived from the graph.
	detector := DetectorForFile(opts.ModelPath)

	if detector == nil {
		detector = DefaultDetector()
		log.Warnf("faces: unrecognized detector %s, assuming %s preprocessing", clean.Log(filepath.Base(opts.ModelPath)), detector.Name)
	}

	// Operators may point MODELS_PATH at another export, whose layout is read from the graph,
	// so a checksum other than the registered one is reported and accepted.
	if err := detector.ONNX.VerifyChecksum(opts.ModelPath); err != nil {
		log.Warnf("faces: %s", err)
	}

	if err := onnx.EnsureRuntime(opts.LibraryPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	sessionOpts, err := onnxruntime.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}
	defer func() {
		if destroyErr := sessionOpts.Destroy(); destroyErr != nil {
			log.Debugf("faces: %s (destroy session options)", destroyErr)
		}
	}()

	threads := opts.Threads
	if threads == 0 {
		threads = max(runtime.NumCPU()/2, 1)
	}

	if err := sessionOpts.SetIntraOpNumThreads(threads); err != nil {
		return nil, fmt.Errorf("faces: configure intra-op threads: %w", err)
	}

	if err := sessionOpts.SetInterOpNumThreads(InterOpThreads); err != nil {
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
	graphWidth, graphHeight, _ := onnx.InputGeometry(inputInfos[0].Dimensions)

	width, height, err := detectorInputSize(detector, graphWidth, graphHeight)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("faces: initialize ONNX session: %w", err)
	}

	engine := &onnxEngine{
		session:        session,
		inputName:      inputName,
		outputNames:    outputNames,
		inputWidth:     width,
		inputHeight:    height,
		colorOrder:     detector.ONNX.Input.ColorOrder,
		mean:           detector.ONNX.Input.Normalization.Mean,
		scales:         detector.ONNX.Input.Normalization.Scales(),
		decode:         detector.Decode,
		detector:       detector.Name,
		featStrides:    featStrides,
		numAnchors:     numAnchors,
		batched:        batched,
		useKps:         useKps,
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
	case 12:
		// YuNet: cls, obj, bbox and kps at three strides, one prior per cell.
		fmc = 3
		anchors = 1
		useKps = true
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

// Detector returns the name of the detection model this session loaded.
func (o *onnxEngine) Detector() DetectorName {
	return o.detector
}

// Close releases the ONNX session.
func (o *onnxEngine) Close() error {
	o.sessionMu.Lock()
	defer o.sessionMu.Unlock()

	if o.session == nil {
		return nil
	}

	err := o.session.Destroy()
	o.session = nil

	return err
}

// Detect identifies faces in the provided image using the ONNX runtime session.
func (o *onnxEngine) Detect(fileName string, minSize int) (Faces, error) {
	img, _, err := fs.DecodeImageFile(fileName)
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
	defer func() {
		if destroyErr := tensor.Destroy(); destroyErr != nil {
			log.Debugf("faces: %s (destroy input tensor)", destroyErr)
		}
	}()

	inputs := []onnxruntime.Value{tensor}
	outputs := make([]onnxruntime.Value, len(o.outputNames))

	// Reconfiguring the detector closes this session while indexing workers may still
	// hold it, so the nil check has to happen under the lock that Close takes.
	o.sessionMu.Lock()

	if o.session == nil {
		o.sessionMu.Unlock()
		return Faces{}, fmt.Errorf("faces: detector was closed while detecting")
	}

	err = o.session.Run(inputs, outputs)
	o.sessionMu.Unlock()

	if err != nil {
		return Faces{}, fmt.Errorf("faces: run session: %w", err)
	}
	for _, out := range outputs {
		if out != nil {
			value := out
			defer func() {
				if destroyErr := value.Destroy(); destroyErr != nil {
					log.Debugf("faces: %s (destroy output tensor)", destroyErr)
				}
			}()
		}
	}

	var detections []onnxDetection

	if o.decode == DecodeYuNet {
		detections, err = o.parseYuNetDetections(outputs, detScale, width, height)
	} else {
		detections, err = o.parseDetections(outputs, detScale, width, height)
	}
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

		if det.hasKps {
			f.Eyes, f.Landmarks = LandmarkAreas(det.kps, size)
		}

		result.Append(f)
	}

	return result, nil
}

// buildBlob normalizes the input image into the tensor layout expected by SCRFD.
func (o *onnxEngine) buildBlob(img image.Image) ([]float32, float32, error) {
	// detectorInputSize resolved these before the session was created, so they are known
	// to be within bounds and non-zero here.
	inputWidth := o.inputWidth
	inputHeight := o.inputHeight

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return nil, 0, fmt.Errorf("faces: invalid image dimensions")
	}

	imRatio := float32(height) / float32(width)
	modelRatio := float32(inputHeight) / float32(inputWidth)

	var newHeight, newWidth int
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

	resized := resizeLinearImage(img, newWidth, newHeight)

	planeSize := inputWidth * inputHeight
	blob := make([]float32, planeSize*onnx.Channels)

	rIndex, gIndex, bIndex := o.colorOrder.Indices()

	for y := 0; y < inputHeight; y++ {
		for x := 0; x < inputWidth; x++ {
			idx := y*inputWidth + x
			var r, g, b float32
			if x < newWidth && y < newHeight {
				cr, cg, cb, _ := resized.At(x, y).RGBA()
				r = float32((cr >> 8) & 0xff)
				g = float32((cg >> 8) & 0xff)
				b = float32((cb >> 8) & 0xff)
			}

			// The padded area is normalized like every other pixel, so the value the model
			// sees there stays the one it was trained to treat as empty.
			blob[idx+planeSize*rIndex] = (r - o.mean[0]) * o.scales[0]
			blob[idx+planeSize*gIndex] = (g - o.mean[1]) * o.scales[1]
			blob[idx+planeSize*bIndex] = (b - o.mean[2]) * o.scales[2]
		}
	}

	detScale := float32(newHeight) / float32(height)

	return blob, detScale, nil
}

// resizeLinearImage rescales an image with a lightweight linear filter for ONNX preprocessing.
func resizeLinearImage(img image.Image, width, height int) image.Image {
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	return dst
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
		landmarks := o.landmarkData(values, level, fmc, expected)

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

			det := onnxDetection{
				x1:    x1,
				y1:    y1,
				x2:    x2,
				y2:    y2,
				score: score,
			}

			if landmarks != nil {
				kpsOffset := idx * NumLandmarks * 2

				// Keypoints are kept unclamped, unlike the box above. similarityTransform
				// is a least-squares fit over all five, so snapping one point that the
				// detector placed outside the frame rotates and scales the whole crop.
				// Sampling out of bounds is already handled by transparent black.
				for p := range NumLandmarks {
					det.kps[p*2] = (cx + landmarks[kpsOffset+p*2]*float32(stride)) / detScale
					det.kps[p*2+1] = (cy + landmarks[kpsOffset+p*2+1]*float32(stride)) / detScale
				}

				det.hasKps = true
			}

			detections = append(detections, det)
		}
	}

	return detections, nil
}

// landmarkData returns the raw landmark predictions for a feature map level, or nil
// when the model has no landmark outputs. An unexpected tensor layout degrades to nil
// so detection keeps working and embedding falls back to unaligned crops.
func (o *onnxEngine) landmarkData(values []onnxruntime.Value, level, fmc, expected int) []float32 {
	if !o.useKps {
		return nil
	}

	index := level + fmc*2

	if index >= len(values) {
		return nil
	}

	tensor, ok := values[index].(*onnxruntime.Tensor[float32])

	if !ok {
		log.Debugf("faces: unexpected tensor type for landmarks")
		return nil
	}

	data := tensor.GetData()

	if len(data) != expected*NumLandmarks*2 {
		log.Debugf("faces: unexpected landmark tensor size %d (expected %d)", len(data), expected*NumLandmarks*2)
		return nil
	}

	return data
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
	for y := range height {
		cy := float32(y * stride)
		for x := range width {
			cx := float32(x * stride)
			for range anchors {
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
	x1     float32
	y1     float32
	x2     float32
	y2     float32
	score  float32
	kps    [NumLandmarks * 2]float32
	hasKps bool
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

	for i := range boxes {
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
