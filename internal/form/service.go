package form

import (
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/service"
)

// Service represents a remote service form for uploading, downloading or syncing media files.
type Service struct {
	AccName       string `json:"AccName"`
	AccOwner      string `json:"AccOwner"`
	AccURL        string `json:"AccURL"`
	AccType       string `json:"AccType"`
	AccKey        string `json:"AccKey"`
	AccUser       string `json:"AccUser"`
	AccPass       string `json:"AccPass"`
	AccTimeout    string `json:"AccTimeout"` // Request timeout: default, high, medium, low, none
	AccError      string `json:"AccError"`
	AccShare      bool   `json:"AccShare"`   // Manual upload enabled, see SharePath, ShareSize, and ShareExpires.
	AccSync       bool   `json:"AccSync"`    // Background sync enabled, see SyncDownload and SyncUpload.
	RetryLimit    int    `json:"RetryLimit"` // Maximum number of failed requests.
	SharePath     string `json:"SharePath"`
	ShareSize     string `json:"ShareSize"`
	ShareExpires  int    `json:"ShareExpires"`
	SyncPath      string `json:"SyncPath"`
	SyncInterval  int    `json:"SyncInterval"`
	SyncUpload    bool   `json:"SyncUpload"`
	SyncDownload  bool   `json:"SyncDownload"`
	SyncFilenames bool   `json:"SyncFilenames"`
	SyncRaw       bool   `json:"SyncRaw"`
}

// NewService creates a new service form.
func NewService(m interface{}) (f Service, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}

// Discovery performs automatic service discovery.
func (f *Service) Discovery() error {
	acc, err := service.Discover(f.AccURL, f.AccUser, f.AccPass)

	if err != nil {
		return err
	}

	err = deepcopier.Copy(acc).To(f)

	return err
}
