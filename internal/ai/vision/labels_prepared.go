package vision

import (
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
)

// UsesPreparedLabels reports whether the model uses the bundled S2 photo input path.
func (m *Model) UsesPreparedLabels() bool {
	if m == nil || m.Type != ModelTypeLabels || m.TensorFlow != nil || m.hasService() {
		return false
	}
	return classify.NormalizeModelName(classify.ModelName(m.Name)) == classify.ModelEfficientFormerV2S2 &&
		(m.Resolution <= 0 || thumb.Vision(m.Resolution).Name == thumb.Tile224)
}

var preparedLabelsFunc = generatePreparedLabels

// SetPreparedLabelsFunc overrides prepared label generation for tests.
func SetPreparedLabelsFunc(fn func([]classify.Input, entity.Src) (classify.Labels, error)) {
	if fn == nil {
		preparedLabelsFunc = generatePreparedLabels
	} else {
		preparedLabelsFunc = fn
	}
}

// GeneratePreparedLabels classifies prepared S2 inputs and merges their visible labels.
func GeneratePreparedLabels(inputs []classify.Input, labelSrc entity.Src) (classify.Labels, error) {
	return preparedLabelsFunc(inputs, labelSrc)
}

// generatePreparedLabels runs the configured local S2 classifier.
func generatePreparedLabels(inputs []classify.Input, labelSrc entity.Src) (classify.Labels, error) {
	if Config == nil {
		return nil, fmt.Errorf("vision service is not configured")
	}
	model := Config.Model(ModelTypeLabels)
	if !model.UsesPreparedLabels() {
		return nil, fmt.Errorf("prepared labels require the local S2 model")
	}
	classifier := model.ClassifyModel()
	if classifier == nil {
		return nil, fmt.Errorf("invalid labels model configuration")
	}
	if labelSrc == entity.SrcAuto {
		labelSrc = model.GetSource()
	}
	return preparedLabels(inputs, labelSrc, Config.Thresholds.Confidence, classifier.Predict)
}

// preparedLabels applies per-input predictions before merging their labels.
func preparedLabels(inputs []classify.Input, source entity.Src, confidence int, predict func(classify.Input, int) (classify.Labels, error)) (result classify.Labels, err error) {
	if len(inputs) < 1 || len(inputs) > 2 {
		return nil, fmt.Errorf("one or two prepared inputs required")
	}
	for _, input := range inputs {
		if !input.Prepared {
			return nil, fmt.Errorf("prepared input required")
		}
		labels, predictErr := predict(input, confidence)
		if predictErr != nil {
			return result, predictErr
		}
		result = mergeLabels(result, labels, source)
	}
	sort.Sort(result)
	return result, nil
}
