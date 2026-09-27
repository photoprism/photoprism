package photoprism

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
)

// stampTestPage stores markers already holding their closest cluster, ordered by matching cursor.
func stampTestPage(t *testing.T, name string, count int) (*Faces, *entity.Face, []string) {
	t.Helper()
	w := isolatedTestFaces(t, name)
	require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM markers").Error)
	f := entity.NewFace("", entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(8300)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	uids := consensusTestMarkers(t, f, count, "", entity.SrcAuto, false)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid IN (?)", uids).UpdateColumns(entity.Values{
		"matched_at": nil, "face_dist": 0, "embeddings_json": face.Embeddings{f.Embedding()}.JSON(),
	}).Error)
	sort.Strings(uids)
	return w, f, uids
}

// markerStampScope reports whether an update changes only a marker match timestamp.
func markerStampScope(scope *gorm.Scope) bool {
	if scope.TableName() != (entity.Marker{}).TableName() {
		return false
	}
	attrs, ok := scope.InstanceGet("gorm:update_attrs")
	if !ok {
		return false
	}
	values, ok := attrs.(map[string]interface{})
	if !ok || len(values) != 1 {
		return false
	}
	_, ok = values["matched_at"]
	return ok
}

// TestFaces_MatchFacesStampPages checks page flushes, pass selection, and recoverable stamp failures.
func TestFaces_MatchFacesStampPages(t *testing.T) {
	for _, mode := range []string{"Complete", "QueryError", "StampFailure"} {
		t.Run(mode, func(t *testing.T) {
			w, f, uids := stampTestPage(t, "facesstamps"+mode, 5)
			oldLimit := faceMatchBatchSize
			faceMatchBatchSize = 2
			t.Cleanup(func() { faceMatchBatchSize = oldLimit })
			hook := captureLog(t)
			queries, updates := 0, 0
			observe := true
			entity.Db().Callback().Query().Before("gorm:query").Register("test:stamp-page-query", func(scope *gorm.Scope) {
				if _, ok := scope.Value.(*entity.Markers); !ok || !observe {
					return
				}
				queries++
				if mode != "StampFailure" {
					for _, uid := range uids[:min((queries-1)*faceMatchBatchSize, len(uids))] {
						assert.NotNil(t, entity.FindMarker(uid).MatchedAt, "the prior page is persisted before this query")
					}
				}
				if mode == "QueryError" && queries == 2 {
					_ = scope.Err(errors.New("test page query failure"))
				}
			})
			t.Cleanup(func() { entity.Db().Callback().Query().Remove("test:stamp-page-query") })
			entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:stamp-page-update", func(scope *gorm.Scope) {
				if !markerStampScope(scope) {
					return
				}
				updates++
				if mode == "StampFailure" && updates == 1 {
					_ = scope.Err(errors.New("test page stamp failure"))
				}
			})
			t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:stamp-page-update") })
			cutoff := entity.TimeStamp()
			stats := make(map[string]*faceMatchStats)
			result, err := w.MatchFaces(entity.Faces{*f}, false, nil, stats)
			observe = false
			require.NotNil(t, stats[f.ID])
			if mode == "QueryError" {
				require.ErrorContains(t, err, "test page query failure")
				assert.Equal(t, 1, updates)
				assert.Equal(t, 2, stats[f.ID].matched)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 3, updates, "one stamp statement per page")
				assert.Equal(t, 5, stats[f.ID].matched)
			}
			assert.Equal(t, FacesMatchResult{}, result)
			for i, uid := range uids {
				stored := entity.FindMarker(uid)
				require.NotNil(t, stored)
				if mode == "QueryError" && i >= 2 || mode == "StampFailure" && i < 2 {
					assert.Nil(t, stored.MatchedAt)
				} else {
					assert.NotNil(t, stored.MatchedAt)
				}
			}
			if mode == "StampFailure" {
				assert.Contains(t, strings.Join(loggedMessages(hook, logrus.WarnLevel), "\n"), "test page stamp failure")
			}
			if mode == "Complete" {
				stats = make(map[string]*faceMatchStats)
				result, err = w.MatchFaces(entity.Faces{*f}, false, cutoff, stats)
				require.NoError(t, err)
				assert.Empty(t, stats, "pass two does not revisit markers covered by pass one")
				assert.Equal(t, FacesMatchResult{}, result)
				assert.Equal(t, 3, updates)
			}
		})
	}
}

// TestFaces_MatchFacesStampCancellation checks that an interrupted page persists its collected stamps.
func TestFaces_MatchFacesStampCancellation(t *testing.T) {
	w, f, uids := stampTestPage(t, "facesstampcancel", 4)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", uids[1]).UpdateColumn("face_id", "").Error)
	require.NoError(t, mutex.FacesWorker.Start())
	t.Cleanup(mutex.FacesWorker.Stop)
	canceled := false
	entity.Db().Callback().Update().After("gorm:commit_or_rollback_transaction").Register("test:stamp-cancel", func(scope *gorm.Scope) {
		if m, ok := scope.Value.(*entity.Marker); ok && m.MarkerUID == uids[1] && !canceled {
			canceled = true
			w.Cancel()
		}
	})
	t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:stamp-cancel") })
	_, err := w.MatchFaces(entity.Faces{*f}, false, nil, nil)
	require.ErrorContains(t, err, "worker canceled")
	require.True(t, canceled)
	assert.NotNil(t, entity.FindMarker(uids[0]).MatchedAt, "the collected stamp is flushed on return")
	assert.NotNil(t, entity.FindMarker(uids[1]).MatchedAt)
	assert.Nil(t, entity.FindMarker(uids[2]).MatchedAt)
	assert.Nil(t, entity.FindMarker(uids[3]).MatchedAt)
}
