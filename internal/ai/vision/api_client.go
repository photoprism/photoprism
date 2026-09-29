package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	httpclient "github.com/photoprism/photoprism/pkg/http/client"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/safe"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// PerformApiRequest performs a Vision API request and returns the result.
func PerformApiRequest(apiRequest *ApiRequest, uri, method, key string) (apiResponse *ApiResponse, err error) {
	if apiRequest == nil {
		return apiResponse, errors.New("api request is nil")
	} else if err = validateApiRequestURL(uri); err != nil {
		return apiResponse, invalidUriError(apiRequest.GetResponseFormat(), err)
	}

	data, jsonErr := apiRequest.JSON()

	if jsonErr != nil {
		return apiResponse, jsonErr
	}

	// Bound the total request time, including any 429 retries, to ServiceTimeout.
	ctx, cancel := context.WithTimeout(context.Background(), ServiceTimeout)
	defer cancel()

	// Create HTTP client and a factory that builds a fresh authenticated request
	// per attempt, so a buffered payload is replayed safely when retrying a 429.
	client := http.Client{Timeout: ServiceTimeout, CheckRedirect: noRedirect}
	newReq := func() (*http.Request, error) {
		req, reqErr := http.NewRequestWithContext(ctx, method, uri, bytes.NewReader(data))
		if reqErr != nil {
			return nil, reqErr
		}

		// Add "application/json" content type header.
		header.SetContentType(req, header.ContentTypeJson)

		// Add an authentication header if an access token is provided.
		if key != "" {
			header.SetAuthorization(req, key)
		}

		// Add custom OpenAI organization and project headers.
		if apiRequest.GetResponseFormat() == ApiFormatOpenAI {
			header.SetOpenAIOrg(req, apiRequest.Org)
			header.SetOpenAIProject(req, apiRequest.Project)
		}

		return req, nil
	}

	// Perform API request, retrying transient HTTP 429 responses with bounded
	// exponential backoff while other statuses stay terminal.
	// #nosec G704 URI is validated by validateApiRequestURL before issuing the request.
	clientResp, clientErr := httpclient.Do(ctx, &client, newReq, httpclient.RetryPolicy{
		MaxRetries:      ServiceMaxRetries,
		BaseDelay:       ServiceRetryDelay,
		MaxDelay:        ServiceRetryMaxDelay,
		RetryStatuses:   []int{http.StatusTooManyRequests},
		HonorRetryAfter: true,
	})

	if clientErr != nil {
		return apiResponse, transportError(apiRequest.GetResponseFormat(), clientErr)
	}

	defer func() {
		_ = clientResp.Body.Close()
	}()

	if location := clientResp.Header.Get(header.Location); location != "" && clientResp.StatusCode >= 300 && clientResp.StatusCode < 400 {
		return nil, redirectError(apiRequest.GetResponseFormat(), clientResp.StatusCode, location)
	}

	body, apiErr := io.ReadAll(io.LimitReader(clientResp.Body, MaxResponseBytes+1))
	if apiErr != nil {
		return nil, transportError(apiRequest.GetResponseFormat(), apiErr)
	} else if int64(len(body)) > MaxResponseBytes {
		return nil, fmt.Errorf("vision: response exceeds the maximum size of %d bytes", MaxResponseBytes)
	}

	format := apiRequest.GetResponseFormat()

	if clientResp.StatusCode >= 300 {
		logServiceResponse(serviceError(format, clientResp.StatusCode), body)
	}

	if engine, ok := EngineFor(format); ok && engine.Parser != nil {
		parsed, parseErr := engine.Parser.Parse(context.Background(), apiRequest, body, clientResp.StatusCode)
		if parseErr != nil {
			return nil, responseError(format, clientResp.StatusCode, parseErr)
		}

		if clientResp.StatusCode < 300 && log.IsLevelEnabled(logrus.TraceLevel) {
			log.Tracef("vision: response %q", body)
		}

		return parsed, nil
	}

	apiResponse = &ApiResponse{}

	// Parse and return response, or an error if the request failed.
	switch format {
	case ApiFormatVision:
		if apiErr = json.Unmarshal(body, apiResponse); apiErr != nil || clientResp.StatusCode >= 300 {
			return apiResponse, responseError(format, clientResp.StatusCode, apiErr)
		}
	default:
		return apiResponse, fmt.Errorf("unsupported response format %s", clean.Log(apiRequest.ResponseFormat))
	}

	return apiResponse, nil
}

// serviceName returns the sanitized service name for the specified response format.
func serviceName(format ApiFormat) string {
	if name := clean.TypeLowerUnderscore(format); name != "" {
		return name
	}

	return "remote"
}

// serviceError returns the error for a failed service request, which names the service and status only.
func serviceError(format ApiFormat, code int) error {
	return fmt.Errorf("%s service request failed (status %d)", serviceName(format), code)
}

// redirectError returns the error for a service redirect, and writes its target to the system log
// with the userinfo and query redacted.
func redirectError(format ApiFormat, code int, location string) error {
	err := fmt.Errorf("%s service request failed (status %d, redirect not followed)", serviceName(format), code)
	logServiceResponse(err, []byte("location "+redirectTarget(location)))

	return err
}

// redirectTarget returns the redirect location without its userinfo, query, and fragment.
func redirectTarget(location string) string {
	u, err := url.Parse(location)

	if err != nil {
		return clean.UriQueriesRedacted(clean.UriRedactedText(location))
	}

	u.User = nil
	u.Fragment, u.RawFragment = "", ""

	if u.RawQuery != "" || u.ForceQuery {
		u.RawQuery, u.ForceQuery = clean.UriRedactedValue, false
	}

	return u.String()
}

// invalidUriError returns the error for a service URI that is not a valid request URL, and writes
// the cause to the system log with the userinfo and queries redacted.
func invalidUriError(format ApiFormat, cause error) error {
	err := fmt.Errorf("%s service request failed (invalid service uri)", serviceName(format))
	logServiceResponse(err, []byte(clean.UriQueriesRedacted(clean.UriRedactedText(cause.Error()))))

	return err
}

// noRedirect returns the redirect response to the caller instead of following it.
func noRedirect(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

// transportError returns the error for a request that received no response, and writes its cause
// to the system log with the userinfo and queries in any URI it holds redacted.
func transportError(format ApiFormat, cause error) error {
	reason := "connection error"

	if netErr := net.Error(nil); errors.As(cause, &netErr) && netErr.Timeout() {
		reason = "timeout"
	}

	err := fmt.Errorf("%s service request failed (%s)", serviceName(format), reason)
	logServiceResponse(err, []byte(clean.UriQueriesRedacted(clean.UriRedactedText(cause.Error()))))

	return err
}

// responseError returns the error for a response that failed or could not be parsed. Below status 300,
// it writes the parse error to the system log, as its text may quote the response.
func responseError(format ApiFormat, code int, parseErr error) error {
	if code >= 300 {
		return serviceError(format, code)
	}

	err := fmt.Errorf("%s service returned an invalid response (status %d)", serviceName(format), code)

	if parseErr != nil {
		logServiceResponse(err, []byte(parseErr.Error()))
	}

	return err
}

// logServiceResponse writes the text of a failed service response to the system log, clipped and quoted.
func logServiceResponse(err error, text []byte) {
	var truncated string

	clipped := clip.Bytes(string(text), txt.ClipLongText)

	if len(clipped) < len(bytes.TrimSpace(text)) {
		truncated = fmt.Sprintf(" (truncated from %s)", humanize.Bytes(uint64(len(text))))
	}

	event.SystemError([]string{"vision", "%s", "%q%s"}, err, clipped, truncated)
}

// validateApiRequestURL checks that outbound API requests only use HTTP(S) URLs with a host.
func validateApiRequestURL(rawURL string) error {
	_, err := safe.URL(rawURL)
	return err
}

func decodeOllamaResponse(data []byte) (*ollama.Response, error) {
	resp := &ollama.Response{}
	dec := json.NewDecoder(bytes.NewReader(data))

	for {
		var chunk ollama.Response
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
