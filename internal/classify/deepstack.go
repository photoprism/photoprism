package classify

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
)

type DeepStackResponse struct {
	Success     bool
	Predictions []DeepStackPredictions
}

type DeepStackPredictions struct {
	Confidence float32
	Label      string
}

// DeepStack is a wrapper for deepstack REST API.
type DeepStack struct {
	disabled       bool
	apiEndpointUrl string
	labels         []string
}

// New returns new DeepStack instance.
func DeepStackNew(apiEndpointUrl string, disabled bool) *DeepStack {
	apiEndpointUrl = strings.TrimSuffix(apiEndpointUrl, "/")
	return &DeepStack{apiEndpointUrl: apiEndpointUrl, disabled: disabled}
}

// Init tests the DeepStack API connection
func (t *DeepStack) DeepStackInit() (err error) {
	if t.disabled {
		return nil
	}

	return t.testDeepStackApiConnection()
}

// File returns matching labels for a jpeg media file.
func (t *DeepStack) DeepStackFile(filename string) (result Labels, err error) {
	if t.disabled {
		return result, nil
	}

	pathDetectionUrl := fmt.Sprintf("%v/%v", t.apiEndpointUrl, DeepStackApiPathDetection)

	log.Debugf("classify: processing %s via DeepStack, API: %s", filename, pathDetectionUrl)
	//imageBuffer, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var client http.Client

	response, err := UploadMultipartFile(&client, pathDetectionUrl, "image", filename)

	if err != nil {
		return nil, err
	}

	//== defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var data DeepStackResponse
	json.Unmarshal(body, &data)

	for _, prediction := range data.Predictions {

		labelText := strings.ToLower(prediction.Label)
		p := prediction.Confidence

		rule, _ := Rules.Find(labelText)

		// discard labels that don't met the threshold
		if p < rule.Threshold {
			continue
		}

		// Get rule label name instead of t.labels name if it exists
		if rule.Label != "" {
			labelText = rule.Label
		}

		labelText = strings.TrimSpace(labelText)

		uncertainty := 100 - int(math.Round(float64(p*100)))

		result = append(result, Label{Name: labelText, Source: SrcImage, Uncertainty: uncertainty, Priority: rule.Priority, Categories: rule.Categories})
	}

	// Sort by probability
	sort.Sort(result)

	// Return the best labels only.
	if l := len(result); l < 5 {
		return result[:l], err
	} else {
		return result[:5], err
	}
}

// Uploads a file to a server using form based multi-part upload
func UploadMultipartFile(client *http.Client, uri, key, path string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())

	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		w, err := mwriter.CreateFormFile(key, path)
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}
		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mwriter.Close(); err != nil {
			errchan <- err
			return
		}
	}()

	resp, err := client.Do(req)
	merr := <-errchan

	if err != nil || merr != nil {
		return resp, fmt.Errorf("http error: %v, multipart error: %v", err, merr)
	}

	return resp, nil
}

func (t *DeepStack) testDeepStackApiConnection() error {
	//TODO check connectivity, not implemented
	return nil
}
