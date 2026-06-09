package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFacesConfigCommand(t *testing.T) {
	output, err := RunWithTestContext(FacesConfigCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Spot-check face-related rows are reported.
	assert.Contains(t, output, "face-engine")
	assert.Contains(t, output, "face-size")
	assert.Contains(t, output, "face-cluster-dist")
	assert.Contains(t, output, "facenet-model-path")
	assert.Contains(t, output, "disable-faces")

	// Non-face options must not leak into the focused report.
	assert.NotContains(t, output, "originals-path")
	assert.NotContains(t, output, "ffmpeg-bin")
	assert.NotContains(t, output, "auth-mode")
}

func TestFacesConfigCommandJSON(t *testing.T) {
	output, err := RunWithTestContext(FacesConfigCommand, []string{"config", "--json"})

	if err != nil {
		t.Fatal(err)
	}

	var payload struct {
		Sections []struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
		} `json:"sections"`
	}

	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, output)
	}

	if len(payload.Sections) != 1 {
		t.Fatalf("expected exactly one section, got %d", len(payload.Sections))
	}

	names := make(map[string]string, len(payload.Sections[0].Items))
	for _, item := range payload.Sections[0].Items {
		names[item["name"]] = item["value"]
	}

	for _, want := range []string{"face-engine", "face-size", "face-cluster-dist", "facenet-model-path"} {
		if _, ok := names[want]; !ok {
			t.Errorf("expected JSON to contain %q, got keys: %v", want, names)
		}
	}
}
