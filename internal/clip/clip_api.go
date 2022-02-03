package clip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/face"
)

// ClipApi is the interface to talk with an CLIP REST service
type ClipApi struct {
	baseUrl string
}

// EncodeText returns the CLIP embedding for a given text
func (api *ClipApi) EncodeText(text string) (face.Embedding, error) {
	url := api.baseUrl + "/encode_text/" + url.PathEscape(text)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dest []float32
	err = json.Unmarshal(body, &dest)
	if err != nil {
		return nil, err
	}
	return face.NewEmbedding(dest), nil
}

// EncodeText returns the CLIP embedding for a given image
func (api *ClipApi) EncodeImage(image_filename string) (face.Embedding, error) {
	url := api.baseUrl + "/encode_image"
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	fw, err := writer.CreateFormFile("image", image_filename)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(image_filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return nil, err
	}
	writer.Close()
	req, err := http.NewRequest("GET", url, &buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dest []float32
	err = json.Unmarshal(respBody, &dest)
	if err != nil {
		return nil, err
	}
	return face.NewEmbedding(dest), nil
}
