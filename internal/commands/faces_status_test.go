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

	// The state goes above the tables, because it can run to several lines and can be an error,
	// neither of which fits a value column.
	assert.Contains(t, output, "Face detection and recognition are")

	// One table per group, named by the option "show config" reports, so the two can be read
	// against each other and against the flag list.
	assert.Contains(t, output, "GLOBAL OPTIONS")
	assert.Contains(t, output, "FACE DETECTION")
	assert.Contains(t, output, "FACE RECOGNITION")
	assert.Contains(t, output, "face-detector")
	assert.Contains(t, output, "face-model")
	assert.Contains(t, output, "face-size")
	assert.Contains(t, output, "face-cluster-dist")

	// The notes carry what a value cannot say: which detector "auto" resolved to, and that a
	// cutoff no option holds was calibrated for it rather than chosen.
	assert.Contains(t, output, "Detector: ")
	assert.Contains(t, output, "Model: ")

	// The deprecated runtime option is reported only while it still decides something.
	assert.NotContains(t, output, "face-engine")

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
		Status   []string `json:"status"`
		Sections []struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
			Note  string              `json:"note"`
		} `json:"sections"`
	}

	if err = json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, output)
	}

	// A caller parsing the report must get the status and the notes too, or it would have to
	// re-derive from the values exactly what the tables cannot state.
	assert.NotEmpty(t, payload.Status)

	titles := make([]string, 0, len(payload.Sections))
	names := make(map[string]string)

	for _, section := range payload.Sections {
		titles = append(titles, section.Title)

		for _, item := range section.Items {
			names[item["name"]] = item["value"]
		}
	}

	assert.Equal(t, []string{"Global Options", "Face Detection", "Face Recognition"}, titles)

	for _, want := range []string{"face-detector", "face-model", "face-size", "face-cluster-dist"} {
		assert.Contains(t, names, want)
	}
}
