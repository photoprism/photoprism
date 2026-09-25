package workers

import (
	"errors"
	"net/url"
	"testing"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/photoprism/photoprism/internal/mutex"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestNewSync(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
}

func TestSync_Start(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)

	if err := mutex.SyncWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.SyncWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncRefresh_RemoteErrorText(t *testing.T) {
	// A remote server chooses its own error body, which reaches the log through the sync worker.
	// The rendered line must stay one line, in one field, whatever the body contains.
	remote := errors.New("500 Internal Server Error: denied\nsync: › admin › granted\u202e \x07")

	rendered := clean.Error(remote)

	assert.NotContains(t, rendered, "\n", "a remote body must not add a line")
	assert.NotContains(t, rendered, "\r", "a remote body must not add a line")
	assert.NotContains(t, rendered, "›", "a remote body must not add a field separator")
	assert.NotContains(t, rendered, "\u202e", "a remote body must not carry a bidi override")
	assert.NotContains(t, rendered, "\x07", "a remote body must not carry a control character")
	assert.Contains(t, rendered, "denied", "the reason must survive")
}

func TestSyncRefresh_RemoteCredentials(t *testing.T) {
	// A sync account URL may carry credentials, and the remote error repeats the URL it failed on.
	//nolint:gosec // G101: Example credential in a fixture URL, which is the subject of the test.
	remote := &url.Error{Op: "PROPFIND", URL: "https://sync-user:notreal@dav.example.com/photos", Err: errors.New("timeout")}

	rendered := clean.Error(remote)

	assert.NotContains(t, rendered, "notreal", "the credential must not survive")
}
