package workers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestShare_StartExpiredHidden checks that expired copies in folders that count as hidden are removed.
func TestShare_StartExpiredHidden(t *testing.T) {
	var mu sync.Mutex
	var deleted []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mu.Lock()
			deleted = append(deleted, r.URL.Path)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Share Expired " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav",
		AccShare: true, SharePath: "/@Family", ShareExpires: 60, AccTimeout: "low", RetryLimit: -1}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	share := entity.NewFileShare(entity.FileFixturesExampleJPG.ID, a.ID, "/@Family/x.jpg")
	share.Status = entity.FileShareShared
	require.NoError(t, share.Create())
	require.NoError(t, entity.Db().Model(&entity.FileShare{}).Where("service_id = ?", a.ID).
		UpdateColumn("updated_at", time.Now().Add(-2*time.Hour)).Error)

	require.NoError(t, NewShare(config.TestConfig()).Start())

	mu.Lock()
	assert.Contains(t, deleted, "/@Family/x.jpg")
	mu.Unlock()

	var stored entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ?", a.ID).First(&stored).Error)
	assert.Equal(t, entity.FileShareRemoved, stored.Status)
}
