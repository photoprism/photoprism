package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// holdFacesLock takes the migration lock for the duration of a test.
func holdFacesLock(t *testing.T, conf *config.Config) {
	t.Helper()

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)
}

// expireFacesLock writes a lock file whose holder stopped renewing it, which is what a killed
// migration leaves behind.
func expireFacesLock(t *testing.T, conf *config.Config) {
	t.Helper()

	state := mutex.FileLockState{
		Action:    "faces migration",
		PID:       os.Getpid(),
		Host:      "test",
		UpdatedAt: time.Now().Add(-time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	b, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(conf.FacesLockFile(), b, fs.ModeFile))
	t.Cleanup(func() { _ = os.Remove(conf.FacesLockFile()) })
}

// markerFixtureUIDs returns the file and marker of the fixture photo the write tests edit.
func markerFixtureUIDs(t *testing.T, app http.Handler, router *gin.RouterGroup) (fileUID, markerUID string) {
	t.Helper()

	GetPhoto(router)

	r := PerformRequest(app, "GET", "/api/v1/photos/ps6sg6be2lvl0y11")
	require.Equal(t, http.StatusOK, r.Code)

	fileUID = gjson.Get(r.Body.String(), "Files.0.UID").String()
	markerUID = gjson.Get(r.Body.String(), "Files.0.Markers.0.UID").String()

	require.NotEmpty(t, fileUID)
	require.NotEmpty(t, markerUID)

	return fileUID, markerUID
}

// TestFaceMigrationRefusesWrites covers the gate on the marker and subject write paths. A migration
// replaces every cluster in one transaction, so an edit made while it runs is overwritten with no
// error anywhere - the refusal is what makes that visible.
func TestFaceMigrationRefusesWrites(t *testing.T) {
	t.Run("UpdateMarker", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateMarker(router)
		_, markerUID := markerFixtureUIDs(t, app, router)

		holdFacesLock(t, conf)

		r := PerformRequestWithBody(app, "PUT", fmt.Sprintf("/api/v1/markers/%s", markerUID), `{"SubjSrc": "manual", "Name": "Jens Mander"}`)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.Contains(t, r.Body.String(), "migration")
	})
	t.Run("ClearMarkerSubject", func(t *testing.T) {
		app, router, conf := NewApiTest()
		ClearMarkerSubject(router)
		_, markerUID := markerFixtureUIDs(t, app, router)

		holdFacesLock(t, conf)

		r := PerformRequestWithBody(app, "DELETE", fmt.Sprintf("/api/v1/markers/%s/subject", markerUID), "")
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("CreateMarker", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateMarker(router)
		fileUID, _ := markerFixtureUIDs(t, app, router)

		holdFacesLock(t, conf)

		body := fmt.Sprintf(`{"FileUID": %q, "MarkerType": "face", "X": 0.3, "Y": 0.26, "W": 0.54, "H": 0.36}`, fileUID)
		r := PerformRequestWithBody(app, "POST", "/api/v1/markers", body)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("UpdateSubject", func(t *testing.T) {
		app, router, conf := NewApiTest()
		UpdateSubject(router)

		holdFacesLock(t, conf)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Name": "Renamed While Migrating"}`)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("ExpiredLockDoesNotRefuse", func(t *testing.T) {
		// A killed migration leaves its file behind, and a file-exists check would wedge every
		// write path until somebody deleted it by hand.
		app, router, conf := NewApiTest()
		UpdateMarker(router)
		_, markerUID := markerFixtureUIDs(t, app, router)

		expireFacesLock(t, conf)

		r := PerformRequestWithBody(app, "PUT", fmt.Sprintf("/api/v1/markers/%s", markerUID), `{"MarkerInvalid": false}`)
		assert.NotEqual(t, http.StatusConflict, r.Code)
	})
	t.Run("NoLockDoesNotRefuse", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateSubject(router)

		r := PerformRequestWithBody(app, "PUT", "/api/v1/subjects/js6sg6b1qekk9jx8", `{"Name": "Not Migrating"}`)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
