package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestSanitizeVisionSource(t *testing.T) {
	cases := map[string]entity.Src{
		"":        entity.SrcAuto,
		"auto":    entity.SrcAuto,
		"AUTO":    entity.SrcAuto,
		"default": entity.SrcDefault,
		"DEFAULT": entity.SrcDefault,
		"image":   entity.SrcImage,
		"ollama":  entity.SrcOllama,
		"openai":  entity.SrcOpenAI,
		"vision":  entity.SrcVision,
	}

	for input, expected := range cases {
		result, err := sanitizeVisionSource(input)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	}

	if _, err := sanitizeVisionSource("meta"); err == nil {
		t.Fatalf("expected error for unsupported source")
	}
}

func TestVisionSourceUsage(t *testing.T) {
	display := visionSourceUsage()

	for _, name := range []string{"auto", "default", "image", "ollama", "openai", "vision"} {
		if !strings.Contains(display, name) {
			t.Fatalf("expected usage to list %s", name)
		}
	}
}
