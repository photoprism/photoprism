package vision

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/service/http/header"
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

	apiResponse = &ApiResponse{}

	// Parse and return response, or an error if the request failed.
	switch apiRequest.GetResponseFormat() {
	case ApiFormatVision:
		if apiJson, apiErr := io.ReadAll(clientResp.Body); apiErr != nil {
			return apiResponse, apiErr
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			return apiResponse, apiErr
		} else if clientResp.StatusCode >= 300 {
			log.Debugf("vision: %s (status code %d)", apiJson, clientResp.StatusCode)
		}
	case ApiFormatOllama:
		ollamaResponse := &ApiResponseOllama{}

		if apiJson, apiErr := io.ReadAll(clientResp.Body); apiErr != nil {
			return apiResponse, apiErr
		} else if apiErr = json.Unmarshal(apiJson, ollamaResponse); apiErr != nil {
			return apiResponse, apiErr
		} else if clientResp.StatusCode >= 300 {
			log.Debugf("vision: %s (status code %d)", apiJson, clientResp.StatusCode)
		}

		apiResponse.Id = apiRequest.Id
		apiResponse.Code = clientResp.StatusCode
		apiResponse.Model = &Model{
			Name: ollamaResponse.Model,
		}

		apiResponse.Result.Caption = &CaptionResult{
			Text:   ollamaResponse.Response,
			Source: entity.SrcImage,
		}
	default:
		return apiResponse, fmt.Errorf("unsupported response format %s", clean.Log(apiRequest.responseFormat))
	}

	return apiResponse, nil
}
