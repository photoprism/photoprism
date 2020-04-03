package form

import (
	"database/sql"

	"github.com/photoprism/photoprism/internal/service"
	"github.com/ulule/deepcopier"
)

// Account represents a remote service account form for uploading, downloading or syncing media files.
type Account struct {
	AccName      string       `json:"AccName"`
	AccOwner     string       `json:"AccOwner"`
	AccURL       string       `json:"AccURL"`
	AccType      string       `json:"AccType"`
	AccKey       string       `json:"AccKey"`
	AccUser      string       `json:"AccUser"`
	AccPass      string       `json:"AccPass"`
	AccError     string       `json:"AccError"`
	AccShare     bool         `json:"AccShare"`
	AccSync      bool         `json:"AccSync"`
	RetryLimit   uint         `json:"RetryLimit"`
	SharePath    string       `json:"SharePath"`
	ShareSize    string       `json:"ShareSize"`
	ShareExpires uint         `json:"ShareExpires"`
	SyncPath     string       `json:"SyncPath"`
	SyncInterval uint         `json:"SyncInterval"`
	SyncUpload   bool         `json:"SyncUpload"`
	SyncDownload bool         `json:"SyncDownload"`
	SyncDelete   bool         `json:"SyncDelete"`
	SyncRaw      bool         `json:"SyncRaw"`
	SyncVideo    bool         `json:"SyncVideo"`
	SyncSidecar  bool         `json:"SyncSidecar"`
	SyncStart    sql.NullTime `json:"SyncStart"`
}

func NewAccount(m interface{}) (f Account, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}

func (f *Account) ServiceDiscovery() error {
	acc, err := service.Discover(f.AccURL, f.AccUser, f.AccPass)

	if err != nil {
		return err
	}

	err = deepcopier.Copy(acc).To(f)

	return err
}
