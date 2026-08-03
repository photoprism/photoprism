package face

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

// onnxEmbedder generates face embeddings with an ONNX Runtime session.
type onnxEmbedder struct {
	session    *onnxruntime.DynamicAdvancedSession
	model      *EmbeddingModel
	outputName string
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
	case settings.ModelPath == "":
		return nil, fmt.Errorf("faces: embedding model path is empty")
	}

	if _, err := os.Stat(settings.ModelPath); err != nil {
		return nil, fmt.Errorf("faces: %w", err)
	}

	if err := ensureONNXRuntime(settings.LibraryPath); err != nil {
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

	if err = sessionOpts.SetInterOpNumThreads(threads); err != nil {
		return nil, fmt.Errorf("faces: configure inter-op threads: %w", err)
	}

	if err = sessionOpts.SetGraphOptimizationLevel(onnxruntime.GraphOptimizationLevelEnableAll); err != nil {
		return nil, fmt.Errorf("faces: optimize session graph: %w", err)
	}

	inputInfos, outputInfos, err := onnxruntime.GetInputOutputInfoWithOptions(settings.ModelPath, sessionOpts)

	if err != nil {
		return nil, fmt.Errorf("faces: load ONNX metadata: %w", err)
	}

	if len(inputInfos) == 0 {
		return nil, fmt.Errorf("faces: embedding model has no inputs")
	} else if len(outputInfos) == 0 {
		return nil, fmt.Errorf("faces: embedding model has no outputs")
	}

	width, height := embedderInputSize(inputInfos[0].Dimensions, m)
	dims := embedderOutputDims(outputInfos[0].Dimensions, m)

	// A mismatch means the file is not the model the registry describes, so refuse to
	// generate embeddings that would silently be written into an incompatible space.
	if dims != m.Dims {
		return nil, fmt.Errorf("faces: %s returns %d dimensions, expected %d", clean.Log(m.Name), dims, m.Dims)
	}

	session, err := onnxruntime.NewDynamicAdvancedSession(
		settings.ModelPath,
		[]string{inputInfos[0].Name},
		[]string{outputInfos[0].Name},
		sessionOpts,
	)

	if err != nil {
		return nil, fmt.Errorf("faces: initialise ONNX embedding session: %w", err)
	}

	log.Infof("faces: loading %s", clean.Log(m.Name))

	return &onnxEmbedder{
		session:    session,
		model:      m,
		outputName: outputInfos[0].Name,
		width:      width,
		height:     height,
		dims:       dims,
	}, nil
}

// embedderInputSize returns the crop size the model expects, preferring its own
// static input shape over the registered defaults.
func embedderInputSize(dims []int64, m *EmbeddingModel) (width, height int) {
	width = m.Width
	height = m.Height

	if len(dims) >= 4 {
		if w := int(dims[len(dims)-1]); w > 0 {
			width = w
		}

		if h := int(dims[len(dims)-2]); h > 0 {
			height = h
		}
	}

	return width, height
}

// embedderOutputDims returns the embedding length the model reports, falling back
// to the registered value when the output shape is dynamic.
func embedderOutputDims(dims []int64, m *EmbeddingModel) int {
	if len(dims) > 0 {
		if d := int(dims[len(dims)-1]); d > 0 {
			return d
		}
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
	blob := make([]float32, planeSize*3)

	rIndex, gIndex, bIndex := e.model.ColorOrder.Indices()

	for y := range e.height {
		for x := range e.width {
			cr, cg, cb, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			channels := [3]float32{
				(float32(cr>>8) - e.model.Mean) * e.model.Scale,
				(float32(cg>>8) - e.model.Mean) * e.model.Scale,
				(float32(cb>>8) - e.model.Mean) * e.model.Scale,
			}

			idx := y*e.width + x
			blob[idx+planeSize*rIndex] = channels[0]
			blob[idx+planeSize*gIndex] = channels[1]
			blob[idx+planeSize*bIndex] = channels[2]
		}
	}

	return blob, nil
}
