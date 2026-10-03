package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestUnresolvedServiceUri checks that labels, nsfw and caption models whose own service URI does
// not resolve return an error.
func TestUnresolvedServiceUri(t *testing.T) {
	images := Files{fs.Abs("./testdata/cat_224x224.jpg")}

	for _, tc := range []struct {
		name  string
		model *Model
		run   func() error
		err   string
	}{
		{"LabelsTensorFlowName", &Model{Type: ModelTypeLabels, Name: "nasnet", Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}, func() error {
			_, err := labelsInternal(images, media.SrcLocal, entity.SrcImage)
			return err
		}, "service uri of labels model does not resolve"},
		{"LabelsEngine", &Model{Type: ModelTypeLabels, Name: "gemma3:4b", Engine: ollama.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}, func() error {
			_, err := labelsInternal(images, media.SrcLocal, entity.SrcImage)
			return err
		}, "service uri of labels model does not resolve"},
		{"NsfwEngine", &Model{Type: ModelTypeNsfw, Name: "qwen3-vl:4b", Engine: ollama.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}, func() error {
			_, err := nsfwInternalContext(images, media.SrcLocal, nsfwThresholdIndex)
			return err
		}, "service uri of nsfw model does not resolve"},
		{"CaptionEngine", &Model{Type: ModelTypeCaption, Name: "gemma3:4b", Engine: ollama.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}, func() error {
			_, _, err := captionInternal(images, media.SrcLocal)
			return err
		}, "service uri of caption model does not resolve"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
			_, _ = captureLogs(t)
			resetUnresolvedUriWarnings(t)

			prevConfig := Config
			t.Cleanup(func() { Config = prevConfig })
			Config = &ConfigValues{Models: Models{tc.model}, Thresholds: DefaultThresholds}

			var err error
			require.NotPanics(t, func() { err = tc.run() })
			assert.EqualError(t, err, tc.err)
		})
	}
}

// TestModel_NsfwModelWithoutTensorFlow checks that a custom nsfw model without runtime settings does not panic.
func TestModel_NsfwModelWithoutTensorFlow(t *testing.T) {
	_, _ = captureLogs(t)

	model := &Model{Type: ModelTypeNsfw, Name: "custom-nsfw-missing", Path: "custom-nsfw-missing"}

	require.NotPanics(t, func() { assert.Nil(t, model.NsfwModel()) })
}
