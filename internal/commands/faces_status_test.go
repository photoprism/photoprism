package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFacesStatusCommand(t *testing.T) {
	output, err := RunWithTestContext(FacesStatusCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Spot-check face-related rows are reported. They are named without the "face-" prefix the
	// subcommand already carries, which is what distinguishes this report from "show config".
	assert.Contains(t, output, "Detector")
	assert.Contains(t, output, "Model")
	assert.Contains(t, output, "Min Size")
	assert.Contains(t, output, "Cluster Distance")
	assert.Contains(t, output, "Enabled")
	assert.NotContains(t, output, "face-size")

	// Non-face options must not leak into the focused report.
	assert.NotContains(t, output, "originals-path")
	assert.NotContains(t, output, "ffmpeg-bin")
	assert.NotContains(t, output, "auth-mode")
}

// TestFacesStatusCommandAlias checks that the name this command was renamed from still resolves,
// so an operator's notes and scripts keep working.
func TestFacesStatusCommandAlias(t *testing.T) {
	assert.Contains(t, FacesStatusCommand.Aliases, "config")
	assert.Equal(t, "status", FacesStatusCommand.Name)
}

func TestFacesStatusCommandJSON(t *testing.T) {
	output, err := RunWithTestContext(FacesStatusCommand, []string{"status", "--json"})

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

	for _, want := range []string{"Detector", "Model", "Min Size", "Cluster Distance"} {
		if _, ok := names[want]; !ok {
			t.Errorf("expected JSON to contain %q, got keys: %v", want, names)
		}
	}
}
