package classify

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
)

const (
	DefaultResolution       = 224
	ExternalModelsTestLabel = "PHOTOPRISM_TEST_EXTERNAL_MODELS"
)

var baseUrl = "https://dl.photoprism.app/tensorflow/models"

// To avoid downloading everything again and again...
//var baseUrl = "http://host.docker.internal:8000"

type ModelTestCase struct {
	Info   *tensorflow.ModelInfo
	Labels string
}

var modelsInfo = map[string]*ModelTestCase{
	"efficientnet-v2-tensorflow2-imagenet1k-b0-classification-v2.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	"efficientnet-v2-tensorflow2-imagenet1k-m-classification-v2.tar.gz": {
		Info: &tensorflow.ModelInfo{

			Input: &tensorflow.PhotoInput{
				Height: 480,
				Width:  480,
			},
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	"efficientnet-v2-tensorflow2-imagenet21k-b0-classification-v1.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
		Labels: "labels-imagenet21k.txt",
	},
	"inception-v3-tensorflow2-classification-v2.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Input: &tensorflow.PhotoInput{
				Height: 299,
				Width:  299,
			},
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	"resnet-v2-tensorflow2-101-classification-v2.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	"resnet-v2-tensorflow2-152-classification-v2.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	"vision-transformer-tensorflow2-vit-b16-classification-v1.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Input: &tensorflow.PhotoInput{
				Intervals: []tensorflow.Interval{
					{
						Start: -1.0,
						End:   1.0,
					},
				},
			},
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	/* Not correctly uploaded
	"vit-base-patch16-google-250811.tar.gz": {
		Info: &tensorflow.ModelInfo{
			Output: &tensorflow.ModelOutput{
				OutputsLogits: true,
			},
		},
	},
	*/
}

func isSafePath(target, baseDir string) bool {

	// Resolve the absolute path of the target
	absTarget := filepath.Join(baseDir, target)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absTarget, absBase)
}

func TestExternalModel_AllModels(t *testing.T) {

	if os.Getenv(ExternalModelsTestLabel) == "" {
		t.Skipf("Skipping external model tests. To test them add set env var %s=true",
			ExternalModelsTestLabel)
	}

	tmpPath, err := os.MkdirTemp("", "*-photoprism")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpPath)

	for k, v := range modelsInfo {
		t.Run(k, func(*testing.T) {
			log.Infof("vision: testing model %s", k)

			downloadedModel := downloadRemoteModel(t, fmt.Sprintf("%s/%s", baseUrl, k), tmpPath)
			log.Infof("vision: model downloaded to %s", downloadedModel)

			if v.Labels != "" {
				modelPath := filepath.Join(tmpPath, downloadedModel)

				t.Logf("vision: model path is %s", modelPath)
				downloadLabels(t, fmt.Sprintf("%s/%s", baseUrl, v.Labels), modelPath)
			}

			model := NewModel(tmpPath, downloadedModel, modelPath, v.Info, false)
			if err := model.loadModel(); err != nil {
				t.Fatal(err)
			}

			if model.meta.Input.IsDynamic() {
				model.meta.Input.SetResolution(DefaultResolution)
			}

			testModel_LabelsFromFile(t, model)
			testModel_Run(t, model)
		})
	}
}

func downloadLabels(t *testing.T, url, dst string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	output, err := os.Create(filepath.Join(dst, "labels.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func downloadRemoteModel(t *testing.T, url, tmpPath string) (model string) {
	t.Logf("Downloading %s to %s", url, tmpPath)

	modelPath := strings.TrimSuffix(path.Base(url), ".tar.gz")
	tmpPath = filepath.Join(tmpPath, modelPath)
	os.MkdirAll(tmpPath, 0755)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("Invalid status code for url %s: %d", url, resp.StatusCode)
	}

	uncompressedBody, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tarReader := tar.NewReader(uncompressedBody)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("could not extract the file: %v", err)
		}

		target := filepath.Join(tmpPath, header.Name)
		if !isSafePath(target, tmpPath) {
			t.Fatalf("The model file contains an invalid path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(target, 0755); err != nil {
				t.Fatalf("could not make the dir %s: %v", header.Name, err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				t.Fatalf("could not create file %s: %v", header.Name, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				t.Fatalf("could not copy file %s: %v", header.Name, err)
			}

			rootPath, fileName := filepath.Split(header.Name)
			if fileName == "saved_model.pb" {
				model = filepath.Join(modelPath, rootPath)
			}
			outFile.Close()
		default:
			t.Fatalf("could not extract file. Unknown type %v in %s",
				header.Typeflag,
				header.Name)
		}
	}

	return
}

func containsAny(s string, substrings []string) bool {
	for i := range substrings {
		if strings.Contains(s, substrings[i]) {
			return true
		}
	}
	return false
}

func assertContainsAny(t *testing.T, s string, substrings []string) {
	assert.Truef(t, containsAny(s, substrings),
		"The result [%s] does not contain any of %v",
		s, substrings)
}

func testModel_LabelsFromFile(t *testing.T, tensorFlow *Model) {
	testName := func(name string) string {
		return fmt.Sprintf("%s/%s", tensorFlow.name, name)
	}

	t.Run(testName("chameleon_lime.jpg"), func(t *testing.T) {
		result, err := tensorFlow.File(examplesPath+"/chameleon_lime.jpg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.GreaterOrEqual(t, len(result), 1)

		if len(result) != 1 {
			t.Logf("Expected 1 result, but found %d", len(result))
			t.Logf("Results: %#v", result)
		}

		if len(result) > 0 {
			assert.Contains(t, result[0].Name, "chameleon")
			//assert.Equal(t, 7, result[0].Uncertainty)
		}
	})
	t.Run(testName("cat_224.jpeg"), func(t *testing.T) {
		result, err := tensorFlow.File(examplesPath+"/cat_224.jpeg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.GreaterOrEqual(t, len(result), 1)

		if len(result) != 1 {
			t.Logf("Expected 1 result, but found %d", len(result))
			t.Logf("Results: %#v", result)
		}

		if len(result) > 0 {
			assertContainsAny(t, result[0].Name, []string{"cat", "kitty"})
		}
	})
	t.Run(testName("cat_720.jpeg"), func(t *testing.T) {
		result, err := tensorFlow.File(examplesPath+"/cat_720.jpeg", 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		//assert.Equal(t, 3, len(result))
		assert.GreaterOrEqual(t, len(result), 1)

		// t.Logf("labels: %#v", result)
		if len(result) != 3 {
			t.Logf("Expected 3 result, but found %d", len(result))
			t.Logf("Results: %#v", result)
		}

		if len(result) > 0 {
			assertContainsAny(t, result[0].Name, []string{"cat", "kitty"})
		}
	})
	t.Run(testName("green.jpg"), func(t *testing.T) {
		result, err := tensorFlow.File(examplesPath+"/green.jpg", 10)

		t.Logf("labels: %#v", result)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.GreaterOrEqual(t, len(result), 1)

		if len(result) != 1 {
			t.Logf("Expected 1 result, but found %d", len(result))
			t.Logf("Results: %#v", result)
		}

		if len(result) > 0 {
			assert.Equal(t, "outdoor", result[0].Name)
		}
	})
	t.Run(testName("not existing file"), func(t *testing.T) {
		result, err := tensorFlow.File(examplesPath+"/notexisting.jpg", 10)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Empty(t, result)
	})
	t.Run(testName("disabled true"), func(t *testing.T) {
		tensorFlow.disabled = true
		defer func() { tensorFlow.disabled = false }()

		result, err := tensorFlow.File(examplesPath+"/chameleon_lime.jpg", 10)
		assert.Nil(t, err)

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, result)
		assert.IsType(t, Labels{}, result)
		assert.Equal(t, 0, len(result))

		t.Log(result)
	})
}

func testModel_Run(t *testing.T, tensorFlow *Model) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testName := func(name string) string {
		return fmt.Sprintf("%s/%s", tensorFlow.name, name)
	}

	t.Run(testName("chameleon_lime.jpg"), func(t *testing.T) {
		if imageBuffer, err := os.ReadFile(examplesPath + "/chameleon_lime.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.GreaterOrEqual(t, len(result), 1)

			if len(result) != 1 {
				t.Logf("Expected 1 result, but found %d", len(result))
				t.Logf("Results: %#v", result)
			}

			if len(result) > 0 {
				assert.Contains(t, result[0].Name, "chameleon")
			}
		}
	})
	t.Run(testName("dog_orange.jpg"), func(t *testing.T) {
		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.NotNil(t, result)

			if err != nil {
				t.Fatal(err)
			}

			assert.IsType(t, Labels{}, result)
			assert.GreaterOrEqual(t, len(result), 1)

			if len(result) != 1 {
				t.Logf("Expected 1 result, but found %d", len(result))
				t.Logf("Results: %#v", result)
			}

			if len(result) > 0 {
				assertContainsAny(t, result[0].Name, []string{"dog", "corgi"})
			}
		}
	})
	t.Run(testName("Random.docx"), func(t *testing.T) {
		if imageBuffer, err := os.ReadFile(examplesPath + "/Random.docx"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)
			assert.Empty(t, result)
			assert.Error(t, err)
		}
	})
	t.Run(testName("6720px_white.jpg"), func(t *testing.T) {
		if imageBuffer, err := os.ReadFile(examplesPath + "/6720px_white.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			if err != nil {
				t.Fatal(err)
			}

			assert.Empty(t, result)
		}
	})
	t.Run(testName("disabled true"), func(t *testing.T) {
		tensorFlow.disabled = true
		defer func() { tensorFlow.disabled = false }()
		if imageBuffer, err := os.ReadFile(examplesPath + "/dog_orange.jpg"); err != nil {
			t.Error(err)
		} else {
			result, err := tensorFlow.Run(imageBuffer, 10)

			t.Log(result)

			assert.Nil(t, result)

			assert.Nil(t, err)
			assert.IsType(t, Labels{}, result)
			assert.Equal(t, 0, len(result))
		}
	})
}
