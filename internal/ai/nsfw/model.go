package nsfw

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// Classes is the number of scores the bundled TensorFlow model emits.
const Classes = 5

// Model uses TensorFlow to label drawing, hentai, neutral, porn and sexy images.
type Model struct {
	model     *tf.SavedModel
	modelPath string
	labels    []string
	meta      *tensorflow.ModelInfo
	disabled  bool
	mutex     sync.Mutex
}

// NewModel returns a new detector instance.
func NewModel(modelPath string, meta *tensorflow.ModelInfo, disabled bool) *Model {
	if meta == nil {
		meta = new(tensorflow.ModelInfo)
	}

	return &Model{
		modelPath: modelPath,
		meta:      meta,
		disabled:  disabled,
	}
}

// File returns the decision for a local JPEG file, scored against threshold.
func (m *Model) File(fileName string, threshold float32) (result Result, err error) {
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}

	if fs.MimeType(fileName) != header.ContentTypeJpeg {
		return Unavailable("not a jpeg file"), fmt.Errorf("%s is not a jpeg file", clean.Log(filepath.Base(fileName)))
	}

	var img []byte

	if img, err = os.ReadFile(fileName); err != nil { //nolint:gosec // fileName is provided by trusted callers; reading local test fixtures is intentional
		return Unavailable(clean.Error(err)), err
	}

	return m.Run(img, threshold)
}

// Url returns the decision for an image at an https or data URL, scored against threshold.
func (m *Model) Url(imgUrl string, threshold float32) (result Result, err error) {
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}

	var img []byte

	if img, err = media.ReadUrlImage(imgUrl, scheme.HttpsData); err != nil {
		return Unavailable(clean.Error(err)), err
	}

	return m.Run(img, threshold)
}

// Run returns the decision for an encoded JPEG image, scored against threshold.
func (m *Model) Run(img []byte, threshold float32) (result Result, err error) {
	if m == nil || m.disabled {
		return Unavailable("detector is disabled"), nil
	}

	if loadErr := m.loadModel(); loadErr != nil {
		return Unavailable(clean.Error(loadErr)), loadErr
	}

	defer tensorflow.MaybeCollectTensorMemory()

	// Create input tensor from image.
	input, err := tensorflow.ImageTransform(
		img, fs.ImageJpeg, m.meta.Input.Resolution())

	if err != nil {
		return Unavailable(clean.Error(err)), fmt.Errorf("%s", err)
	}

	// Run inference.
	output, err := m.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.model.Graph.Operation(m.meta.Input.Name).Output(m.meta.Input.OutputIndex): input,
		},
		[]tf.Output{
			m.model.Graph.Operation(m.meta.Output.Name).Output(m.meta.Output.OutputIndex),
		},
		nil)

	if err != nil {
		return Unavailable(clean.Error(err)), fmt.Errorf("%s (run inference)", err.Error())
	}

	if len(output) < 1 {
		return Unavailable("no output"), fmt.Errorf("inference failed, no output")
	}

	scores, ok := output[0].Value().([][]float32)
	if !ok || len(scores) < 1 {
		return Unavailable("unexpected output type"), fmt.Errorf("inference returned an unexpected output type")
	}

	if result, err = m.getScores(scores[0]); err != nil {
		return Unavailable(clean.Error(err)), err
	}

	result = result.Decide(threshold)

	log.Tracef("nsfw: image classified as %+v", result)

	return result, nil
}

// Init initializes tensorflow models if not disabled.
func (m *Model) Init() (err error) {
	if m.disabled {
		return nil
	}

	return m.loadModel()
}

func (m *Model) loadModel() error {
	// Use mutex to prevent the model from being loaded and
	// initialized twice by different indexing workers.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.model != nil {
		// Already loaded
		return nil
	}

	modelName := clean.Log(filepath.Base(m.modelPath))
	log.Infof("nsfw: loading %s", modelName)

	if len(m.meta.Tags) == 0 {
		infos, err := tensorflow.GetModelTagsInfo(m.modelPath)

		switch {
		case err != nil:
			log.Errorf("nsfw: could not get the model info for %s (%s)", modelName, clean.Error(err))
		case len(infos) == 1:
			log.Debugf("nsfw: model info: %+v", infos[0])
			m.meta.Merge(&infos[0])
		case len(infos) > 1:
			log.Warnf("nsfw: found %d metagraphs... that's too many", len(infos))
		default:
			log.Warnf("nsfw: no metagraphs found in %s", modelName)
		}
	}

	// Load saved TensorFlow model from the specified path.
	model, err := tensorflow.SavedModel(m.modelPath, m.meta.Tags)
	if err != nil {
		return err
	}

	if !m.meta.IsComplete() {
		input, output, err := tensorflow.GetInputAndOutputFromSavedModel(model)
		if err != nil {
			log.Errorf("nsfw: could not get info from signatures: %v", err)
			input, output, err = tensorflow.GuessInputAndOutput(model)
			if err != nil {
				return fmt.Errorf("nsfw: %w", err)
			}
		}

		m.meta.Merge(&tensorflow.ModelInfo{
			Input:  input,
			Output: output,
		})
	}

	m.model = model

	if m.meta.Output.OutputsLogits {
		_, err = tensorflow.AddSoftmax(m.model.Graph, m.meta)
		if err != nil {
			return fmt.Errorf("nsfw: could not add softmax (%s)", clean.Error(err))
		}
	}

	return m.loadLabels(m.modelPath)
}

func (m *Model) loadLabels(modelPath string) (err error) {
	m.labels, err = tensorflow.LoadLabels(modelPath, int(m.meta.Output.NumOutputs))
	return err
}

// getScores maps exactly five positional model outputs onto their classes.
func (m *Model) getScores(p []float32) (Result, error) {
	if len(p) != Classes {
		return Result{}, fmt.Errorf("model returned %d values, expected %d", len(p), Classes)
	}

	for _, score := range p {
		if err := ValidateScore(score); err != nil {
			return Result{}, err
		}
	}

	return Result{
		Drawing: p[0],
		Hentai:  p[1],
		Neutral: p[2],
		Porn:    p[3],
		Sexy:    p[4],
	}, nil
}
