package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

func TestVisionEndpoint(t *testing.T) {
	cases := []struct {
		name   string
		uri    string
		method string
		want   string
	}{
		{
			name:   "Plain",
			uri:    "http://ollama:11434/api/generate",
			method: "POST",
			want:   "POST http://ollama:11434/api/generate",
		},
		{ //nolint:gosec // example URL, the password in it is exactly what this case redacts
			name:   "BasicAuth",
			uri:    "https://vision:secret@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://vision:***@vision.example.com/api/generate",
		},
		{
			// Nothing distinguishes a name in this position from an access token, which several
			// services carry there, so the name goes and the endpoint is still identified by its
			// scheme, host and path.
			name:   "NameWithoutPassword",
			uri:    "https://vision@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://***@vision.example.com/api/generate",
		},
		{
			name:   "QueryIsKept",
			uri:    "https://api.example.com/v1/responses?tier=flex",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?tier=flex",
		},
		{
			name:   "MissingUri",
			uri:    "",
			method: "POST",
			want:   "",
		},
		{
			name:   "MissingMethod",
			uri:    "https://api.example.com/v1/responses",
			method: "",
			want:   "",
		},
		{
			name:   "Unparsable",
			uri:    "://nope",
			method: "POST",
			want:   "POST ?",
		},
		{
			name:   "ApiKeyQuery",
			uri:    "https://api.example.com/v1/responses?api_key=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?api_key=***",
		},
		{
			name:   "AccessTokenQuery",
			uri:    "https://api.example.com/v1/responses?access_token=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?access_token=***",
		},
		{
			name:   "MixedCaseKeyQuery",
			uri:    "https://api.example.com/v1/responses?X-Api-Key=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?X-Api-Key=***",
		},
		{
			name:   "CredentialQueryBesideAKeptOne",
			uri:    "https://api.example.com/v1/responses?tier=flex&token=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?tier=flex&token=***",
		},
		{
			name:   "RepeatedCredentialQuery",
			uri:    "https://api.example.com/v1/responses?secret=one&secret=two",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?secret=***&secret=***",
		},
		{ //nolint:gosec // example URL, the credentials in it are exactly what this case redacts
			name:   "UserinfoAndQueryTogether",
			uri:    "https://vision:notreal@api.example.com/v1?signature=notreal",
			method: "POST",
			want:   "POST https://vision:***@api.example.com/v1?signature=***",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, visionEndpoint(tc.uri, tc.method))
		})
	}
}

// TestVisionInstalled verifies listing inspects artifacts without loading models.
func TestVisionInstalled(t *testing.T) {
	modelsPath := t.TempDir()
	model := vision.NewLabelModel(classify.DefaultModelName())
	assert.Equal(t, report.No, visionInstalled(model, modelsPath))
	description := classify.DefaultModel()
	filename := description.ONNX.FilePath(filepath.Join(modelsPath, string(description.Name)))
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), fs.ModeDir))
	require.NoError(t, os.WriteFile(filename, []byte("artifact"), fs.ModeFile))
	model.DisabledByMode = true
	assert.Equal(t, report.Yes, visionInstalled(model, modelsPath))
	detector := vision.NewNsfwModel(nsfw.ModelYahoo)
	assert.Equal(t, report.No, visionInstalled(detector, modelsPath))
	custom := &vision.Model{Type: vision.ModelTypeNsfw, Name: "custom", Path: "custom/model.onnx"}
	assert.Equal(t, report.No, visionInstalled(custom, modelsPath))
	require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, "custom"), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(modelsPath, "custom/model.onnx"), []byte("artifact"), fs.ModeFile))
	assert.Equal(t, report.Yes, visionInstalled(custom, modelsPath))
	custom.Service = vision.Service{Uri: "https://example.com", Method: "POST"}
	assert.Equal(t, report.NotAssigned, visionInstalled(custom, modelsPath))
	assert.Equal(t, report.NotAssigned, visionInstalled(&vision.Model{Type: vision.ModelTypeCaption}, modelsPath))
}

// disabledFaceVisionYaml keeps the default models but disables face processing.
const disabledFaceVisionYaml = `Models:
- Type: face
  Default: true
  Disabled: true
`

// remoteFaceVisionYaml selects a face entry that a service endpoint runs.
//
//nolint:gosec // example URL, the password in it is exactly what the test redacts
const remoteFaceVisionYaml = `Models:
- Type: face
  Name: facenet
  Service:
    Uri: https://vision:secret@vision.example.com/api/v1/face
    Method: POST
`

// visionListJSON runs "vision ls --json" and returns the objects it prints.
func visionListJSON(t *testing.T) []map[string]string {
	t.Helper()

	output, err := RunWithTestContext(VisionListCommand, []string{"ls", "--json"})
	require.NoError(t, err)

	var rows []map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &rows))

	return rows
}

// visionListObject returns the "vision ls --json" object with the specified option and type.
func visionListObject(t *testing.T, rows []map[string]string, option, modelType string) map[string]string {
	t.Helper()

	for _, row := range rows {
		if row["option"] == option && row["type"] == modelType {
			return row
		}
	}

	t.Fatalf("no %s row for option %s", modelType, option)

	return nil
}

// TestVisionListCommand verifies the vision and face tables in text and Markdown output, and the
// single flat table that CSV, TSV and JSON exports keep.
func TestVisionListCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(VisionListCommand, []string{"ls"})
		require.NoError(t, err)
		require.Contains(t, output, visionModelsTitle)
		require.Contains(t, output, visionFacesTitle)
		assert.Less(t, strings.Index(output, visionModelsTitle), strings.Index(output, visionFacesTitle))
		assert.Contains(t, output, visionFacesNote)

		// No face row in the vision table, and both face models in the face table.
		visionTable, faceTable, _ := strings.Cut(output, visionFacesTitle)
		assert.NotContains(t, visionTable, " face ")
		assert.NotContains(t, visionTable, faceModelOption)
		assert.Contains(t, faceTable, " "+faceDetectionRole+" ")
		assert.Contains(t, faceTable, " "+faceRecognitionRole+" ")
		assert.Contains(t, faceTable, faceDetectorOption)
		assert.Contains(t, faceTable, faceModelOption)
		assert.Less(t, strings.Index(faceTable, faceDetectionRole), strings.Index(faceTable, faceRecognitionRole))
	})
	t.Run("Markdown", func(t *testing.T) {
		output, err := RunWithTestContext(VisionListCommand, []string{"ls", "--md"})
		require.NoError(t, err)
		assert.Contains(t, output, "### "+visionModelsTitle+"\n")
		assert.Contains(t, output, "### "+visionFacesTitle+"\n")
		assert.Contains(t, output, "| "+faceDetectorOption+" |")
		assert.Contains(t, output, visionFacesNote)
	})
	for _, tc := range []struct {
		name string
		flag string
		sep  string
	}{
		{name: "CSV", flag: "--csv", sep: ";"},
		{name: "TSV", flag: "--tsv", sep: "\t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := RunWithTestContext(VisionListCommand, []string{"ls", tc.flag})
			require.NoError(t, err)
			header := strings.Join(visionListCols, tc.sep)
			assert.Equal(t, 1, strings.Count(output, header))
			assert.True(t, strings.HasSuffix(header, tc.sep+"Option"))
			assert.NotContains(t, output, visionModelsTitle)
			assert.NotContains(t, output, visionFacesNote)
			assert.Contains(t, output, tc.sep+faceDetectorOption+"\n")
			assert.Contains(t, output, tc.sep+faceModelOption+"\n")
			assert.Contains(t, output, tc.sep+visionYamlOption+"\n")
		})
	}
	t.Run("JSON", func(t *testing.T) {
		rows := visionListJSON(t)
		require.NotEmpty(t, rows)

		keys := []string{"model", "type", "engine", "provider", "endpoint", "format", "normalize",
			"resolution", "options", "schedule", "status", "installed", "option"}

		for _, row := range rows {
			assert.Len(t, row, len(keys))

			for _, key := range keys {
				assert.Contains(t, row, key)
			}
		}

		detector := visionListObject(t, rows, faceDetectorOption, vision.ModelTypeFace)
		recognition := visionListObject(t, rows, faceModelOption, vision.ModelTypeFace)
		assert.Equal(t, rows[len(rows)-2], detector)
		assert.Equal(t, rows[len(rows)-1], recognition)
		visionListObject(t, rows, visionYamlOption, vision.ModelTypeLabels)

		for _, row := range rows[:len(rows)-2] {
			assert.NotEqual(t, vision.ModelTypeFace, row["type"])
		}
	})
	t.Run("LogCount", func(t *testing.T) {
		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hook := test.NewLocal(logger)
		t.Cleanup(hook.Reset)

		rows := visionListJSON(t)
		var messages []string

		for _, entry := range hook.AllEntries() {
			messages = append(messages, entry.Message)
		}

		assert.Contains(t, messages, fmt.Sprintf("found %d models", len(rows)))
	})
	t.Run("FaceEntryDisabled", func(t *testing.T) {
		withVisionYaml(t, disabledFaceVisionYaml)
		rows := visionListJSON(t)
		assert.Equal(t, report.Disabled, visionListObject(t, rows, faceDetectorOption, vision.ModelTypeFace)["status"])
		assert.Equal(t, report.Disabled, visionListObject(t, rows, faceModelOption, vision.ModelTypeFace)["status"])
	})
	t.Run("RemoteFaceEntry", func(t *testing.T) {
		withVisionYaml(t, remoteFaceVisionYaml)
		rows := visionListJSON(t)

		// The entry stays in the vision table with its endpoint, without the password.
		remote := visionListObject(t, rows, visionYamlOption, vision.ModelTypeFace)
		assert.Equal(t, "facenet", remote["model"])
		assert.Equal(t, "POST https://vision:***@vision.example.com/api/v1/face", remote["endpoint"])

		recognition := visionListObject(t, rows, faceModelOption, vision.ModelTypeFace)
		assert.Empty(t, recognition["endpoint"])
		assert.Empty(t, recognition["provider"])
		assert.Equal(t, report.NotAssigned, recognition["installed"])

		output, err := RunWithTestContext(VisionListCommand, []string{"ls"})
		require.NoError(t, err)
		visionTable, _, _ := strings.Cut(output, visionFacesTitle)
		assert.Contains(t, visionTable, "POST https://vision:***@vision.example.com/api/v1/face")
		assert.NotContains(t, output, "secret")
	})
}

// visionListValue returns the value of the named column in a "vision ls" row.
func visionListValue(t *testing.T, row []string, col string) string {
	t.Helper()

	for i, name := range visionListCols {
		if name == col {
			return row[i]
		}
	}

	t.Fatalf("unknown column %s", col)

	return ""
}

// TestVisionListRow verifies that rows report what each model runs with, including the face
// embedding model FACE_MODEL selects and the execution provider of local ONNX models.
func TestVisionListRow(t *testing.T) {
	settings := visionListSettings{
		ModelsPath: t.TempDir(),
		Provider:   onnx.ProviderCUDA,
		FaceModel:  face.ModelSFace,
		FaceActive: true,
		FaceRun:    vision.RunAuto,
	}

	remote := &vision.Model{Type: vision.ModelTypeLabels, Engine: "ollama", Name: "qwen3-vl:4b-instruct", Run: vision.RunOnSchedule}
	remote.ApplyEngineDefaults()

	t.Run("DefaultLabels", func(t *testing.T) {
		row := visionListRow(vision.NewLabelModel(classify.DefaultModelName()), settings)
		require.Len(t, row, len(visionListCols))
		assert.Equal(t, string(classify.DefaultModelName()), visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
		assert.Empty(t, visionListValue(t, row, "Options"))
		assert.Equal(t, "auto", visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
	})
	t.Run("NamedLabels", func(t *testing.T) {
		row := visionListRow(&vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S1)}, settings)
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
	})
	t.Run("NamedDetector", func(t *testing.T) {
		row := visionListRow(&vision.Model{Type: vision.ModelTypeNsfw, Name: string(nsfw.ModelFalconsai)}, settings)
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
	})
	t.Run("DisabledByMode", func(t *testing.T) {
		model := vision.NewNsfwModel(nsfw.ModelYahoo)
		model.DisabledByMode = true
		assert.Equal(t, report.Disabled, visionListValue(t, visionListRow(model, settings), "Status"))
	})
	t.Run("Remote", func(t *testing.T) {
		row := visionListRow(remote, settings)
		assert.Equal(t, "qwen3-vl:4b-instruct", visionListValue(t, row, "Model"))
		assert.Equal(t, "ollama", visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Contains(t, visionListValue(t, row, "Endpoint"), "POST ")
		assert.Equal(t, vision.RunOnSchedule, visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceModel", func(t *testing.T) {
		row := visionListRow(vision.FacenetModel.Clone(), settings)
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "112", visionListValue(t, row, "Resolution"))
		assert.Empty(t, visionListValue(t, row, "Options"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceNone", func(t *testing.T) {
		none := settings
		none.FaceModel, none.FaceActive, none.FaceRun = face.ModelNone, false, vision.RunNever
		row := visionListRow(vision.FacenetModel.Clone(), none)
		assert.Equal(t, face.ModelNone, visionListValue(t, row, "Model"))
		assert.Empty(t, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Resolution"))
		assert.Equal(t, vision.RunNever, visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Disabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceNetEntry", func(t *testing.T) {
		facenet := settings
		facenet.FaceModel = face.ModelFaceNet
		model := vision.FacenetModel.Clone()
		row := visionListRow(model, facenet)
		assert.Equal(t, face.ModelFaceNet, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineTensorFlow, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Equal(t, "160", visionListValue(t, row, "Resolution"))
		assert.Equal(t, `{"tags":"serve"}`, visionListValue(t, row, "Options"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))

		// A custom entry loads from its own path, which the registry knows nothing about.
		model.Name = "facenet_custom"
		row = visionListRow(model, facenet)
		assert.Equal(t, "facenet_custom", visionListValue(t, row, "Model"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceEntryDisabled", func(t *testing.T) {
		model := vision.FacenetModel.Clone()
		model.Disabled = true
		assert.Equal(t, report.Disabled, visionListValue(t, visionListRow(model, settings), "Status"))
	})
	t.Run("FaceEndpoint", func(t *testing.T) {
		model := &vision.Model{Type: vision.ModelTypeFace, Name: "facenet", Run: vision.RunManual,
			Service: vision.Service{Uri: "https://vision.example.com/api/v1/face", Method: "POST"}}
		row := visionListRow(model, settings)
		assert.Equal(t, "facenet", visionListValue(t, row, "Model"))
		assert.Equal(t, "POST https://vision.example.com/api/v1/face", visionListValue(t, row, "Endpoint"))
		assert.Equal(t, "auto", visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
}

// visionFaceSettings returns listing settings with both face models in force.
func visionFaceSettings(t *testing.T) visionListSettings {
	return visionListSettings{
		ModelsPath:     t.TempDir(),
		Provider:       onnx.ProviderCPU,
		FaceModel:      face.ModelSFace,
		FaceActive:     true,
		FaceDetector:   face.DetectorYuNet,
		DetectorActive: true,
		FaceRun:        vision.RunAuto,
	}
}

// TestPrintVisionList verifies the sections of text and Markdown output and the flat exports.
func TestPrintVisionList(t *testing.T) {
	s := visionFaceSettings(t)
	rows := visionListRows([]*vision.Model{vision.NewLabelModel(classify.DefaultModelName())}, s)
	faceRows := visionFaceRows([]*vision.Model{vision.FacenetModel.Clone()}, s)

	t.Run("Default", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printVisionList(report.Default, rows, faceRows) })
		require.NoError(t, err)
		assert.Contains(t, output, visionModelsTitle+"\n")
		assert.Contains(t, output, visionFacesTitle+"\n")
		assert.Contains(t, output, visionFacesNote)
		assert.Equal(t, 1, strings.Count(strings.ReplaceAll(output, visionFacesNote, ""), faceDetectorOption))
		assert.NotContains(t, output, visionYamlOption)
		assert.NotContains(t, output, "###")
	})
	t.Run("NoVisionRows", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printVisionList(report.Markdown, nil, faceRows) })
		require.NoError(t, err)
		assert.NotContains(t, output, visionModelsTitle)
		assert.Contains(t, output, "### "+visionFacesTitle+"\n")
	})
	t.Run("JSON", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printVisionList(report.JSON, rows, faceRows) })
		require.NoError(t, err)
		var objects []map[string]string
		require.NoError(t, json.Unmarshal([]byte(output), &objects))
		require.Len(t, objects, 3)
		assert.Equal(t, []string{visionYamlOption, faceDetectorOption, faceModelOption},
			[]string{objects[0]["option"], objects[1]["option"], objects[2]["option"]})
	})
	t.Run("InvalidFormat", func(t *testing.T) {
		var err error
		capture.Output(func() { err = printVisionList(report.Format("invalid"), rows, faceRows) })
		assert.Error(t, err)
	})
}

// TestVisionListRows verifies that local face entries are left to the face table, while the
// other entries, including a face entry that a service runs, stay in the vision table.
func TestVisionListRows(t *testing.T) {
	s := visionFaceSettings(t)
	local := vision.FacenetModel.Clone()
	remote := &vision.Model{Type: vision.ModelTypeFace, Name: "facenet",
		Service: vision.Service{Uri: "https://vision.example.com/api/v1/face", Method: "POST"}}
	labels := vision.NewLabelModel(classify.DefaultModelName())

	t.Run("Success", func(t *testing.T) {
		rows := visionListRows([]*vision.Model{labels, nil, local, remote}, s)
		require.Len(t, rows, 2)
		assert.Equal(t, vision.ModelTypeLabels, visionListValue(t, rows[0], "Type"))
		assert.Equal(t, vision.ModelTypeFace, visionListValue(t, rows[1], "Type"))
		assert.Equal(t, "POST https://vision.example.com/api/v1/face", visionListValue(t, rows[1], "Endpoint"))

		for _, row := range rows {
			assert.Equal(t, visionYamlOption, visionListValue(t, row, "Option"))
		}
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, visionListRows(nil, s))
		assert.Empty(t, visionListRows([]*vision.Model{local}, s))
	})
	t.Run("VisionUri", func(t *testing.T) {
		withVisionServiceUri(t, "https://vision:secret@vision.example.com/api/v1")
		rows := visionListRows([]*vision.Model{vision.FacenetModel.Clone()}, s)
		require.Len(t, rows, 1)
		assert.Equal(t, "POST https://vision:***@vision.example.com/api/v1/face", visionListValue(t, rows[0], "Endpoint"))
	})
}

// withVisionServiceUri sets the global vision service URI that every entry without its own
// endpoint uses, restoring it afterwards.
func withVisionServiceUri(t *testing.T, uri string) {
	t.Helper()

	serviceUri, serviceMethod := vision.ServiceUri, vision.ServiceMethod
	t.Cleanup(func() { vision.ServiceUri, vision.ServiceMethod = serviceUri, serviceMethod })
	vision.ServiceUri, vision.ServiceMethod = uri, "POST"
}

// TestVisionFaceEntry verifies the face entry follows vision.Config.Model.
func TestVisionFaceEntry(t *testing.T) {
	first := &vision.Model{Type: vision.ModelTypeFace, Name: "first"}
	second := &vision.Model{Type: vision.ModelTypeFace, Name: "second"}
	disabled := &vision.Model{Type: vision.ModelTypeFace, Name: "disabled", Disabled: true}
	byMode := &vision.Model{Type: vision.ModelTypeFace, Name: "by_mode", DisabledByMode: true}
	labels := vision.NewLabelModel(classify.DefaultModelName())

	t.Run("LastEnabled", func(t *testing.T) {
		models := []*vision.Model{first, second, byMode, disabled, labels, nil}
		assert.Same(t, second, visionFaceEntry(models))
		assert.Same(t, (&vision.ConfigValues{Models: []*vision.Model{first, second, byMode, disabled, labels}}).Model(vision.ModelTypeFace), visionFaceEntry(models))
	})
	t.Run("AllDisabled", func(t *testing.T) {
		assert.Same(t, byMode, visionFaceEntry([]*vision.Model{disabled, byMode, labels}))
	})
	t.Run("None", func(t *testing.T) {
		assert.Nil(t, visionFaceEntry([]*vision.Model{labels, nil}))
		assert.Nil(t, visionFaceEntry(nil))
	})
}

// TestVisionDetectorRow verifies the detector row follows face-detector, the face entry, and
// whether faces are processed at all.
func TestVisionDetectorRow(t *testing.T) {
	t.Run("Installed", func(t *testing.T) {
		s := visionFaceSettings(t)
		detector := face.FindDetector(face.DetectorYuNet)
		require.NoError(t, os.MkdirAll(filepath.Dir(detector.Path(s.ModelsPath)), fs.ModeDir))
		require.NoError(t, os.WriteFile(detector.Path(s.ModelsPath), []byte("artifact"), fs.ModeFile))
		row := visionDetectorRow(true, s)
		require.Len(t, row, len(visionListCols))
		assert.Equal(t, face.DetectorYuNet, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.ModelTypeFace, visionListValue(t, row, "Type"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cpu", visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Endpoint"))
		assert.Empty(t, visionListValue(t, row, "Format"))
		assert.Empty(t, visionListValue(t, row, "Options"))
		assert.Equal(t, "640", visionListValue(t, row, "Resolution"))
		assert.Equal(t, "auto", visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.Yes, visionListValue(t, row, "Installed"))
		assert.Equal(t, faceDetectorOption, visionListValue(t, row, "Option"))
	})
	t.Run("NotInstalled", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceDetector, s.DetectorActive = face.DetectorSCRFD, false
		row := visionDetectorRow(true, s)
		assert.Equal(t, face.DetectorSCRFD, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, report.Disabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
	})
	t.Run("None", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceDetector, s.DetectorActive, s.FaceRun = face.DetectorNone, false, vision.RunNever
		row := visionDetectorRow(true, s)
		assert.Equal(t, face.DetectorNone, visionListValue(t, row, "Model"))
		assert.Empty(t, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Resolution"))
		assert.Equal(t, vision.RunNever, visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Disabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("EntryDisabled", func(t *testing.T) {
		assert.Equal(t, report.Disabled, visionListValue(t, visionDetectorRow(false, visionFaceSettings(t)), "Status"))
	})
}

// TestVisionRecognitionRow verifies the recognition row reports the embedding model, or the
// face-model name when a service endpoint returns the vectors.
func TestVisionRecognitionRow(t *testing.T) {
	t.Run("Local", func(t *testing.T) {
		row := visionRecognitionRow(vision.FacenetModel.Clone(), visionFaceSettings(t))
		require.Len(t, row, len(visionListCols))
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cpu", visionListValue(t, row, "Provider"))
		assert.Equal(t, "112", visionListValue(t, row, "Resolution"))
		assert.Equal(t, "default", visionListValue(t, row, "Format"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
		assert.Equal(t, faceModelOption, visionListValue(t, row, "Option"))
	})
	t.Run("FaceNet", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceModel = face.ModelFaceNet
		row := visionRecognitionRow(vision.FacenetModel.Clone(), s)
		assert.Equal(t, face.ModelFaceNet, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineTensorFlow, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Equal(t, "160", visionListValue(t, row, "Resolution"))
		assert.Equal(t, `{"tags":"serve"}`, visionListValue(t, row, "Options"))
	})
	t.Run("Remote", func(t *testing.T) {
		model := &vision.Model{Type: vision.ModelTypeFace, Name: "facenet", Engine: "vision", Resolution: 160,
			Service: vision.Service{Uri: "https://vision.example.com/api/v1/face", Method: "POST"}}
		row := visionRecognitionRow(model, visionFaceSettings(t))
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Equal(t, "vision", visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Endpoint"))
		assert.Empty(t, visionListValue(t, row, "Resolution"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
		assert.Equal(t, faceModelOption, visionListValue(t, row, "Option"))
	})
	t.Run("VisionUri", func(t *testing.T) {
		withVisionServiceUri(t, "https://vision.example.com/api/v1")
		row := visionRecognitionRow(vision.FacenetModel.Clone(), visionFaceSettings(t))
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Endpoint"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("NoEntry", func(t *testing.T) {
		row := visionRecognitionRow(nil, visionFaceSettings(t))
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Equal(t, report.Disabled, visionListValue(t, row, "Status"))
		assert.False(t, vision.FacenetModel.Disabled)
	})
}

// TestVisionFaceRows verifies that each face row reports its own status, while both share the
// schedule, which FACE_RUN sets and a missing detector stops.
func TestVisionFaceRows(t *testing.T) {
	entry := vision.FacenetModel.Clone()

	status := func(rows [][]string) []string {
		require.Len(t, rows, 2)
		assert.Equal(t, faceDetectorOption, visionListValue(t, rows[0], "Option"))
		assert.Equal(t, faceModelOption, visionListValue(t, rows[1], "Option"))
		assert.Equal(t, visionListValue(t, rows[0], "Schedule"), visionListValue(t, rows[1], "Schedule"))

		return []string{visionListValue(t, rows[0], "Status"), visionListValue(t, rows[1], "Status")}
	}

	t.Run("Default", func(t *testing.T) {
		rows := visionFaceRows([]*vision.Model{entry}, visionFaceSettings(t))
		assert.Equal(t, []string{report.Enabled, report.Enabled}, status(rows))
	})
	t.Run("FaceModelNone", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceModel, s.FaceActive = face.ModelNone, false
		rows := visionFaceRows([]*vision.Model{entry}, s)
		assert.Equal(t, []string{report.Enabled, report.Disabled}, status(rows))
		assert.Equal(t, report.NotAssigned, visionListValue(t, rows[1], "Installed"))
	})
	t.Run("FaceDetectorNone", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceDetector, s.DetectorActive, s.FaceRun = face.DetectorNone, false, vision.RunNever
		rows := visionFaceRows([]*vision.Model{entry}, s)
		assert.Equal(t, []string{report.Disabled, report.Enabled}, status(rows))
		assert.Equal(t, vision.RunNever, visionListValue(t, rows[0], "Schedule"))
	})
	t.Run("DisableFaces", func(t *testing.T) {
		s := visionFaceSettings(t)
		s.FaceActive, s.DetectorActive, s.FaceRun = false, false, vision.RunNever
		assert.Equal(t, []string{report.Disabled, report.Disabled}, status(visionFaceRows([]*vision.Model{entry}, s)))
	})
	t.Run("EntryDisabled", func(t *testing.T) {
		disabled := vision.FacenetModel.Clone()
		disabled.Disabled = true
		rows := visionFaceRows([]*vision.Model{disabled}, visionFaceSettings(t))
		assert.Equal(t, []string{report.Disabled, report.Disabled}, status(rows))
	})
	t.Run("EntryDisabledByMode", func(t *testing.T) {
		byMode := vision.FacenetModel.Clone()
		byMode.DisabledByMode = true
		rows := visionFaceRows([]*vision.Model{byMode}, visionFaceSettings(t))
		assert.Equal(t, []string{report.Disabled, report.Disabled}, status(rows))
	})
	t.Run("NoEntry", func(t *testing.T) {
		rows := visionFaceRows(nil, visionFaceSettings(t))
		assert.Equal(t, []string{report.Disabled, report.Disabled}, status(rows))
	})
}

// TestVisionFaceTable verifies that flat face rows are projected onto the face table with roles.
func TestVisionFaceTable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rows := visionFaceTable(visionFaceRows([]*vision.Model{vision.FacenetModel.Clone()}, visionFaceSettings(t)))
		require.Len(t, rows, 2)
		assert.Equal(t, []string{face.DetectorYuNet, faceDetectionRole, vision.EngineONNX, "cpu", "640", "auto",
			report.Enabled, report.No, faceDetectorOption}, rows[0])
		assert.Equal(t, []string{face.ModelSFace, faceRecognitionRole, vision.EngineONNX, "cpu", "112", "auto",
			report.Enabled, report.No, faceModelOption}, rows[1])
	})
	t.Run("UnknownOption", func(t *testing.T) {
		row := make([]string, len(visionListCols))
		row[visionListColumn("Option")] = visionYamlOption
		rows := visionFaceTable([][]string{row})
		require.Len(t, rows, 1)
		assert.Empty(t, rows[0][1])
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, visionFaceTable(nil))
	})
}

// TestVisionListColumn verifies column lookups in the flat table.
func TestVisionListColumn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, 0, visionListColumn("Model"))
		assert.Equal(t, len(visionListCols)-1, visionListColumn("Option"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Panics(t, func() { visionListColumn("Role") })
	})
}

// TestVisionRemote verifies that only a model with a service endpoint is reported as remote.
func TestVisionRemote(t *testing.T) {
	assert.True(t, visionRemote(&vision.Model{Type: vision.ModelTypeFace,
		Service: vision.Service{Uri: "https://vision.example.com/api/v1/face", Method: "POST"}}))
	assert.False(t, visionRemote(vision.FacenetModel.Clone()))
	assert.False(t, visionRemote(nil))
}

// TestVisionRunText verifies that the automatic schedule is named.
func TestVisionRunText(t *testing.T) {
	assert.Equal(t, "auto", visionRunText(vision.RunAuto))
	assert.Equal(t, vision.RunNever, visionRunText(vision.RunNever))
	assert.Equal(t, vision.RunOnSchedule, visionRunText(vision.RunOnSchedule))
}

// TestNewVisionListSettings verifies the settings follow the instance configuration.
func TestNewVisionListSettings(t *testing.T) {
	t.Run("FaceModelNone", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().FaceModel = face.ModelNone
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.ModelNone, settings.FaceModel)
		assert.False(t, settings.FaceActive)
	})
	t.Run("FaceModelNotInstalled", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ModelsPath = t.TempDir()
		conf.Options().FaceModel = face.ModelAuraFace
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.ModelAuraFace, settings.FaceModel)
		assert.False(t, settings.FaceActive)
	})
	t.Run("FaceModelAutoNotInstalled", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ModelsPath = t.TempDir()
		conf.Options().FaceModel = face.ModelAuto
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.ModelNone, settings.FaceModel)
		assert.False(t, settings.FaceActive)
	})

	t.Run("FaceDetectorNone", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().FaceDetector = face.DetectorNone
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.DetectorNone, settings.FaceDetector)
		assert.False(t, settings.DetectorActive)
	})
	t.Run("FaceDetectorNotInstalled", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ModelsPath = t.TempDir()
		conf.Options().FaceDetector = face.DetectorYuNet
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.DetectorYuNet, settings.FaceDetector)
		assert.False(t, settings.DetectorActive)
	})
	t.Run("FaceDetectorAutoNotInstalled", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ModelsPath = t.TempDir()
		conf.Options().FaceDetector = face.DetectorAuto
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.DetectorNone, settings.FaceDetector)
		assert.False(t, settings.DetectorActive)
	})

	conf := get.Config()
	settings := newVisionListSettings(conf)
	assert.Equal(t, conf.FaceDetector(), settings.FaceDetector)
	assert.Equal(t, conf.ModelsPath(), settings.ModelsPath)
	assert.Equal(t, conf.OnnxProvider(), settings.Provider)
	assert.Equal(t, conf.FaceEngineRunType(), settings.FaceRun)

	disabled := conf.Options().DisableFaces
	t.Cleanup(func() { conf.Options().DisableFaces = disabled })
	conf.Options().DisableFaces = true
	settings = newVisionListSettings(conf)
	assert.False(t, settings.FaceActive)
	assert.False(t, settings.DetectorActive)
	assert.Equal(t, vision.RunNever, settings.FaceRun)
}

// TestVisionRegisteredWidth verifies the input width of registered models named in vision.yml.
func TestVisionRegisteredWidth(t *testing.T) {
	assert.Equal(t, 224, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelRepViTM10)}))
	assert.Equal(t, 224, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeNsfw, Name: string(nsfw.ModelYahoo)}))
	assert.Zero(t, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeLabels, Name: "custom_21k"}))
	assert.Zero(t, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeCaption, Name: string(nsfw.ModelYahoo)}))
}
