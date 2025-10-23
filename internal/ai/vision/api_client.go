package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// PerformApiRequest performs a Vision API request and returns the result.
func PerformApiRequest(apiRequest *ApiRequest, uri, method, key string) (apiResponse *ApiResponse, err error) {
	if apiRequest == nil {
		return apiResponse, errors.New("api request is nil")
	}

	data, jsonErr := apiRequest.JSON()

	if jsonErr != nil {
		return apiResponse, jsonErr
	}

	// Create HTTP client and authenticated service API request.
	client := http.Client{Timeout: ServiceTimeout}
	req, reqErr := http.NewRequest(method, uri, bytes.NewReader(data))

	// Add "application/json" content type header.
	header.SetContentType(req, header.ContentTypeJson)

	// Add an authentication header if an access token is configured.
	if key != "" {
		header.SetAuthorization(req, key)
	}

	if reqErr != nil {
		return apiResponse, reqErr
	}

	// Perform API request.
	clientResp, clientErr := client.Do(req)

	if clientErr != nil {
		return apiResponse, clientErr
	}

	defer func() {
		_ = clientResp.Body.Close()
	}()

	body, apiErr := io.ReadAll(clientResp.Body)
	if apiErr != nil {
		return nil, apiErr
	}

	format := apiRequest.GetResponseFormat()

	if engine, ok := EngineFor(format); ok && engine.Parser != nil {
		if clientResp.StatusCode >= 300 {
			log.Debugf("vision: %s (status code %d)", body, clientResp.StatusCode)
		}

		parsed, parseErr := engine.Parser.Parse(context.Background(), apiRequest, body, clientResp.StatusCode)
		if parseErr != nil {
			return nil, parseErr
		}

		return parsed, nil
	}

	apiResponse = &ApiResponse{}

	// Parse and return response, or an error if the request failed.
	switch format {
	case ApiFormatVision:
		if apiErr = json.Unmarshal(body, apiResponse); apiErr != nil {
			return apiResponse, apiErr
		} else if clientResp.StatusCode >= 300 {
			log.Debugf("vision: %s (status code %d)", body, clientResp.StatusCode)
		}
	default:
		return apiResponse, fmt.Errorf("unsupported response format %s", clean.Log(apiRequest.ResponseFormat))
	}

	return apiResponse, nil
}

func decodeOllamaResponse(data []byte) (*ApiResponseOllama, error) {
	resp := &ApiResponseOllama{}
	dec := json.NewDecoder(bytes.NewReader(data))

	for {
		var chunk ApiResponseOllama
		if err := dec.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		*resp = chunk
	}

	return resp, nil
}

func parseOllamaLabels(raw string) ([]LabelResult, error) {
	cleaned := clean.JSON(raw)
	if cleaned == "" {
		return nil, nil
	}

	var payload struct {
		Labels []LabelResult `json:"labels"`
	}

	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		return nil, err
	}

	return payload.Labels, nil
}
