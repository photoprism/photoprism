package vision

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type Files = []string

// ApiRequest represents a Vision API service request.
type ApiRequest struct {
	Id             string    `form:"id" yaml:"Id,omitempty" json:"id,omitempty"`
	Model          string    `form:"model" yaml:"Model,omitempty" json:"model,omitempty"`
	Url            string    `form:"url" yaml:"Url,omitempty" json:"url,omitempty"`
	Images         Files     `form:"images" yaml:"Images,omitempty" json:"images,omitempty"`
	responseFormat ApiFormat `form:"-"`
}

// NewApiRequest returns a new service API request with the specified format and payload.
func NewApiRequest(requestFormat ApiFormat, files Files, fileScheme scheme.Type) (result *ApiRequest, err error) {
	if len(files) == 0 {
		return result, errors.New("missing files")
	}

	switch requestFormat {
	case ApiFormatUrl:
		return NewApiRequestUrl(files[0], fileScheme)
	case ApiFormatImages, ApiFormatVision:
		return NewApiRequestImages(files, fileScheme)
	default:
		return result, errors.New("invalid request format")
	}
}

// NewApiRequestUrl returns a new Vision API request with the specified image Url as payload.
func NewApiRequestUrl(fileName string, fileScheme scheme.Type) (result *ApiRequest, err error) {
	var imgUrl string

	switch fileScheme {
	case scheme.Https:
		// Return if no thumbnail filenames were given.
		if !fs.FileExistsNotEmpty(fileName) {
			return result, errors.New("invalid image file name")
		}

		// Generate a random token for the remote service to download the file.
		fileUuid := rnd.UUID()

		if err = download.Register(fileUuid, fileName); err != nil {
			return result, fmt.Errorf("%s (create download url)", err)
		}

		imgUrl = fmt.Sprintf("%s/%s", DownloadUrl, fileUuid)
	case scheme.Data:
		var u *url.URL
		if u, err = url.Parse(fileName); err != nil {
			return result, fmt.Errorf("%s (invalid image url)", err)
		} else if !slices.Contains(scheme.HttpsHttp, u.Scheme) {
			return nil, fmt.Errorf("unsupported image url scheme %s", clean.Log(u.Scheme))
		} else {
			imgUrl = u.String()
		}
	default:
		return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Url:            imgUrl,
		responseFormat: ApiFormatVision,
	}, nil
}

// NewApiRequestImages returns a new Vision API request with the specified images as payload.
func NewApiRequestImages(images Files, fileScheme scheme.Type) (*ApiRequest, error) {
	imageUrls := make(Files, len(images))

	if fileScheme == scheme.Https && !strings.HasPrefix(DownloadUrl, "https://") {
		log.Tracef("vision: file request scheme changed from https to data because https is not configured")
		fileScheme = scheme.Data
	}

	for i := range images {
		switch fileScheme {
		case scheme.Https:
			fileUuid := rnd.UUID()
			if err := download.Register(fileUuid, images[i]); err != nil {
				return nil, fmt.Errorf("%s (create download url)", err)
			} else {
				imageUrls[i] = fmt.Sprintf("%s/%s", DownloadUrl, fileUuid)
			}
		case scheme.Data:
			if file, err := os.Open(images[i]); err != nil {
				return nil, fmt.Errorf("%s (create data url)", err)
			} else {
				imageUrls[i] = media.DataUrl(file)
			}
		default:
			return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
		}
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Images:         imageUrls,
		responseFormat: ApiFormatVision,
	}, nil
}

// GetId returns the request ID string and generates a random ID if none was set.
func (r *ApiRequest) GetId() string {
	if r.Id == "" {
		r.Id = rnd.UUID()
	}

	return r.Id
}

// GetResponseFormat returns the expected response format type.
func (r *ApiRequest) GetResponseFormat() ApiFormat {
	if r.responseFormat == "" {
		return ApiFormatVision
	}

	return r.responseFormat
}

// JSON returns the request data as JSON-encoded bytes.
func (r *ApiRequest) JSON() ([]byte, error) {
	return json.Marshal(*r)
}
