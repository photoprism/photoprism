package face

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

// onnxEmbedder generates face embeddings with an ONNX Runtime session. The preprocessing
// values are copied out of the model description so the inner loop stays free of pointer
// hops and of the division that a standard deviation would otherwise cost per pixel.
type onnxEmbedder struct {
	session    *onnxruntime.DynamicAdvancedSession
	model      *EmbeddingModel
	colorOrder onnx.ColorOrder
	mean       [onnx.Channels]float32
	scales     [onnx.Channels]float32
	width      int
	height     int
	dims       int
	mutex      sync.Mutex
}

// NewONNXEmbedder loads an ONNX face embedding model and returns it as an Embedder.
func NewONNXEmbedder(settings EmbedderSettings) (Embedder, error) {
	m := settings.Model

	switch {
	case m == nil:
		return nil, fmt.Errorf("faces: missing embedding model")
	case m.Runtime != RuntimeONNX:
		return nil, fmt.Errorf("faces: %s is not an ONNX embedding model", clean.Log(m.Name))
	case m.ONNX == nil || m.ONNX.Input == nil:
		return nil, fmt.Errorf("faces: %s has no ONNX model description", clean.Log(m.Name))
	case settings.ModelPath == "":
		return nil, fmt.Errorf("faces: embedding model path is empty")
	}

	if _, err := os.Stat(settings.ModelPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	// Embeddings are persisted, so a file that is not the artifact the registry describes
	// must never produce them: channel order and normalization cannot be read from a
	// graph, so the wrong preprocessing would be applied silently and every vector would
	// land in a space nothing else can be compared against.
	if err := m.ONNX.VerifyChecksum(settings.ModelPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	if err := onnx.EnsureRuntime(settings.LibraryPath); err != nil {
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

	threads := settings.Threads

	if threads <= 0 {
		threads = max(runtime.NumCPU()/2, 1)
	}

	if err = sessionOpts.SetIntraOpNumThreads(threads); err != nil {
		return nil, fmt.Errorf("faces: configure intra-op threads: %w", err)
	}

	if err = sessionOpts.SetInterOpNumThreads(InterOpThreads); err != nil {
		return nil, fmt.Errorf("faces: configure inter-op threads: %w", err)
	}

	if err = sessionOpts.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		return nil, fmt.Errorf("faces: optimize session graph: %w", err)
	}

	graph, err := onnx.Inspect(settings.ModelPath, sessionOpts)

	if err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	// A graph that contradicts the registry is describing a different model, so refuse to
	// generate embeddings that would silently be written into an incompatible space.
	if err = m.ONNX.VerifyGraph(graph); err != nil {
		return nil, fmt.Errorf("faces: %s %w", clean.Log(m.Name), err)
	}

	width, height := embedderInputSize(m, graph)
	dims := embedderOutputDims(m, graph)

	if dims != m.Dims {
		return nil, fmt.Errorf("faces: %s returns %d dimensions, expected %d", clean.Log(m.Name), dims, m.Dims)
	}

	session, err := onnxruntime.NewDynamicAdvancedSession(
		settings.ModelPath,
		[]string{graph.Input.Name},
		[]string{graph.Output.Name},
		sessionOpts,
	)

	if err != nil {
		return nil, fmt.Errorf("faces: initialize ONNX embedding session: %w", err)
	}

	log.Infof("faces: loading %s", clean.Log(m.Name))

	return &onnxEmbedder{
		session:    session,
		model:      m,
		colorOrder: m.ONNX.Input.ColorOrder,
		mean:       m.ONNX.Input.Normalization.Mean,
		scales:     m.ONNX.Input.Normalization.Scales(),
		width:      width,
		height:     height,
		dims:       dims,
	}, nil
}

// embedderInputSize returns the crop size the model expects, preferring the graph's own
// static input shape over the registered value.
func embedderInputSize(m *EmbeddingModel, graph *onnx.ModelInfo) (width, height int) {
	width, height = m.InputSize()
	graphWidth, graphHeight := graph.InputSize()

	// Each axis falls back on its own, because a graph may fix one and leave the other
	// dynamic.
	if graphWidth > 0 {
		width = graphWidth
	}

	if graphHeight > 0 {
		height = graphHeight
	}

	return width, height
}

// embedderOutputDims returns the embedding length the graph reports, falling back to the
// registered value when the output shape is dynamic.
func embedderOutputDims(m *EmbeddingModel, graph *onnx.ModelInfo) int {
	if dims := graph.OutputWidth(); dims > 0 {
		return dims
	}

	return m.Dims
}

// ModelName returns the name of the embedding model.
func (e *onnxEmbedder) ModelName() ModelName {
	return e.model.Name
}

// Dims returns the length of the embeddings the model produces.
func (e *onnxEmbedder) Dims() int {
	return e.dims
}

// CropSize returns the input width and height the model expects.
func (e *onnxEmbedder) CropSize() (width, height int) {
	return e.width, e.height
}

// Aligned reports whether the model requires landmark-aligned crops.
func (e *onnxEmbedder) Aligned() bool {
	return e.model.Aligned()
}

// Close releases the ONNX session.
func (e *onnxEmbedder) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.session == nil {
		return nil
	}

	err := e.session.Destroy()
	e.session = nil

	return err
}

// Run returns the embeddings for a prepared face crop.
func (e *onnxEmbedder) Run(img image.Image) Embeddings {
	if img == nil {
		return nil
	}

	blob, err := e.buildBlob(img)

	if err != nil {
		log.Errorf("faces: %s", err)
		return nil
	}

	inputShape := onnxruntime.Shape{1, 3, int64(e.height), int64(e.width)}
	inputTensor, err := onnxruntime.NewTensor(inputShape, blob)

	if err != nil {
		log.Errorf("faces: create input tensor: %s", err)
		return nil
	}

	defer func() {
		if destroyErr := inputTensor.Destroy(); destroyErr != nil {
			log.Debugf("faces: %s (destroy input tensor)", destroyErr)
		}
	}()

	outputTensor, err := onnxruntime.NewEmptyTensor[float32](onnxruntime.Shape{1, int64(e.dims)})

	if err != nil {
		log.Errorf("faces: create output tensor: %s", err)
		return nil
	}

	defer func() {
		if destroyErr := outputTensor.Destroy(); destroyErr != nil {
			log.Debugf("faces: %s (destroy output tensor)", destroyErr)
		}
	}()

	// ONNX Runtime sessions are not documented as safe for concurrent Run calls with
	// preallocated outputs, and indexing workers share a single embedder instance.
	// Reconfiguring the embedder closes this session while workers may still hold it,
	// so the nil check must happen under the lock that Close takes.
	e.mutex.Lock()

	if e.session == nil {
		e.mutex.Unlock()
		log.Warnf("faces: %s was closed while generating embeddings", clean.Log(e.model.Name))

		return nil
	}

	err = e.session.Run([]onnxruntime.Value{inputTensor}, []onnxruntime.Value{outputTensor})
	e.mutex.Unlock()

	if err != nil {
		log.Errorf("faces: run embedding session: %s", err)
		return nil
	}

	data := outputTensor.GetData()

	if len(data) != e.dims {
		log.Errorf("faces: embedding has %d dimensions, expected %d", len(data), e.dims)
		return nil
	}

	result := make([]float32, e.dims)
	copy(result, data)

	return NewEmbeddings([][]float32{result})
}

// buildBlob converts a face crop into the NCHW tensor layout the model expects.
func (e *onnxEmbedder) buildBlob(img image.Image) ([]float32, error) {
	if e.width < 1 || e.height < 1 {
		return nil, fmt.Errorf("invalid embedding model input size %dx%d", e.width, e.height)
	}

	if img.Bounds().Dx() != e.width || img.Bounds().Dy() != e.height {
		img = thumb.Resample(img, e.width, e.height, thumb.ResampleFillCenter)
	}

	bounds := img.Bounds()
	planeSize := e.width * e.height
	blob := make([]float32, planeSize*onnx.Channels)

	rIndex, gIndex, bIndex := e.colorOrder.Indices()

	for y := range e.height {
		for x := range e.width {
			cr, cg, cb, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			channels := [onnx.Channels]float32{
				(float32(cr>>8) - e.mean[0]) * e.scales[0],
				(float32(cg>>8) - e.mean[1]) * e.scales[1],
				(float32(cb>>8) - e.mean[2]) * e.scales[2],
			}

			idx := y*e.width + x
			blob[idx+planeSize*rIndex] = channels[0]
			blob[idx+planeSize*gIndex] = channels[1]
			blob[idx+planeSize*bIndex] = channels[2]
		}
	}

	return blob, nil
}
