package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEngine(t *testing.T) {
	cases := map[string]EngineName{
		"":         EngineAuto,
		"AUTO":     EngineAuto,
		"pigo":     EngineONNX,
		"  PIGO  ": EngineONNX,
		"onnx":     EngineONNX,
		"OnNx":     EngineONNX,
		"unknown":  EngineAuto,
		"none":     EngineNone,
	}

	for input, expected := range cases {
		if got := ParseEngine(input); got != expected {
			t.Fatalf("ParseEngine(%q) = %q, expected %q", input, got, expected)
		}
	}
}

func TestActiveEngineName(t *testing.T) {
	assert.Equal(t, EngineNone, ActiveEngineName())
}
