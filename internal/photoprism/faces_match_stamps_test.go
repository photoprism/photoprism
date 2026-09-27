package photoprism

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

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

// TestFaces_MatchFacesPause checks that only a page that changed a marker pauses the walk.
func TestFaces_MatchFacesPause(t *testing.T) {
	for _, force := range []bool{false, true} {
		for _, mode := range []string{"StampedOnly", "ChangedPage", "NoMatchPage", "AmbiguousPage"} {
			t.Run(fmt.Sprintf("%s/Force%t", mode, force), func(t *testing.T) {
				w, f, uids := stampTestPage(t, fmt.Sprintf("facespause%s%t", mode, force), 5)
				oldLimit, oldPause := faceMatchBatchSize, faceMatchPause
				pauses := 0
				faceMatchBatchSize = 2
				faceMatchPause = func() { pauses++ }
				t.Cleanup(func() { faceMatchBatchSize, faceMatchPause = oldLimit, oldPause })
				faces := entity.Faces{*f}
				switch mode {
				case "ChangedPage":
					require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", uids[1]).UpdateColumns(entity.Values{"face_id": "", "face_dist": -1}).Error)
				case "NoMatchPage":
					require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", uids[3]).UpdateColumns(entity.Values{
						"face_id": "", "face_dist": -1, "embeddings_json": face.Embeddings{face.FixtureEmbedding(8301)}.JSON(),
					}).Error)
				case "AmbiguousPage":
					faces[0].SubjUID = entity.SubjectFixtures.Get("john-doe").SubjUID
					rival := entity.NewFace(entity.SubjectFixtures.Get("jane-doe").SubjUID, entity.SrcManual, face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), face.MatchMarginDefault/2, 8302)}, face.EmbeddingModelName())
					faces = append(faces, *rival)
				}
				result, err := w.MatchFaces(faces, force, nil, nil)
				require.NoError(t, err)
				switch mode {
				case "ChangedPage":
					assert.Equal(t, int64(1), result.Updated)
					assert.Equal(t, f.ID, entity.FindMarker(uids[1]).FaceID)
					assert.Equal(t, 1, pauses, "only the page holding the changed marker pauses")
				case "AmbiguousPage":
					assert.Equal(t, int64(5), result.Ambiguous)
					assert.Equal(t, int64(5), result.Updated)
					assert.Equal(t, 3, pauses, "every page detached a marker")
				default:
					assert.Equal(t, FacesMatchResult{}, result)
					assert.Zero(t, pauses)
				}
				for _, uid := range uids {
					assert.NotNil(t, entity.FindMarker(uid).MatchedAt, "every page is still walked")
				}
			})
		}
	}
}

// TestFaces_MatchFacesCursor checks that both walks page by cursor, without counting markers.
func TestFaces_MatchFacesCursor(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprintf("Force%t", force), func(t *testing.T) {
			w, f, uids := stampTestPage(t, fmt.Sprintf("facescursor%t", force), 5)
			oldLimit := faceMatchBatchSize
			faceMatchBatchSize = 2
			t.Cleanup(func() { faceMatchBatchSize = oldLimit })
			counts, queries := 0, 0
			entity.Db().Callback().RowQuery().After("gorm:row_query").Register("test:marker-count", func(scope *gorm.Scope) {
				if scope.TableName() == (entity.Marker{}).TableName() && strings.Contains(strings.ToLower(scope.SQL), "count(") {
					counts++
				}
			})
			t.Cleanup(func() { entity.Db().Callback().RowQuery().Remove("test:marker-count") })
			// Adds a marker that sorts after all others before the first page, as one detected during
			// the walk would, and removes the first marker before the second, which shifts later rows.
			added := "m" + strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6] + "zzzzzzzzz"
			t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", added) })
			entity.Db().Callback().Query().Before("gorm:query").Register("test:marker-delete", func(scope *gorm.Scope) {
				if _, ok := scope.Value.(*entity.Markers); !ok {
					return
				}
				switch queries++; queries {
				case 1:
					require.NoError(t, entity.UnscopedDb().Create(&entity.Marker{
						MarkerUID: added, FileUID: consensusTestFileUID, MarkerType: entity.MarkerFace, MarkerSrc: entity.SrcImage,
						EmbeddingsJSON: face.Embeddings{f.Embedding()}.JSON(), EmbedModel: f.EmbedModel, FaceDist: -1, W: 0.1, H: 0.1,
					}).Error)
				case 2:
					require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM markers WHERE marker_uid = ?", uids[0]).Error)
				}
			})
			t.Cleanup(func() { entity.Db().Callback().Query().Remove("test:marker-delete") })
			_, err := w.MatchFaces(entity.Faces{*f}, force, nil, nil)
			require.NoError(t, err)
			assert.Zero(t, counts)
			for _, uid := range uids[1:] {
				assert.NotNil(t, entity.FindMarker(uid).MatchedAt, "the walk reaches every remaining marker")
			}
			if force {
				assert.Equal(t, 4, queries, "three pages and the empty one that ends the walk")
				assert.Nil(t, entity.FindMarker(added).MatchedAt, "the full walk ends at the markers it started with")
			} else {
				assert.Equal(t, 4, queries, "the added marker completes the third page")
				assert.NotNil(t, entity.FindMarker(added).MatchedAt, "the unmatched walk takes new markers")
			}
		})
	}
}

// TestFaces_MatchFacesBoundError checks that a full walk returns when its bound cannot be read.
func TestFaces_MatchFacesBoundError(t *testing.T) {
	w, f, uids := stampTestPage(t, "facesbounderror", 2)
	queries := 0
	entity.Db().Callback().Query().Before("gorm:query").Register("test:bound-error", func(scope *gorm.Scope) {
		if _, ok := scope.Value.(*entity.Marker); ok {
			_ = scope.Err(errors.New("test bound failure"))
		}
	})
	t.Cleanup(func() { entity.Db().Callback().Query().Remove("test:bound-error") })
	entity.Db().Callback().Query().Before("gorm:query").Register("test:bound-pages", func(scope *gorm.Scope) {
		if _, ok := scope.Value.(*entity.Markers); ok {
			queries++
		}
	})
	t.Cleanup(func() { entity.Db().Callback().Query().Remove("test:bound-pages") })
	_, err := w.MatchFaces(entity.Faces{*f}, true, nil, nil)
	entity.Db().Callback().Query().Remove("test:bound-error")
	require.ErrorContains(t, err, "test bound failure")
	assert.Zero(t, queries)
	for _, uid := range uids {
		assert.Nil(t, entity.FindMarker(uid).MatchedAt)
	}
}

// TestFaces_MatchFacesEmpty checks that a full walk over no markers reads no page.
func TestFaces_MatchFacesEmpty(t *testing.T) {
	w, f, _ := stampTestPage(t, "facesempty", 1)
	require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM markers").Error)
	// On MariaDB the isolated config shares the package database, so later tests need the fixtures.
	t.Cleanup(entity.ResetTestFixtures)
	queries := 0
	entity.Db().Callback().Query().Before("gorm:query").Register("test:empty-walk", func(scope *gorm.Scope) {
		if _, ok := scope.Value.(*entity.Markers); ok {
			queries++
		}
	})
	t.Cleanup(func() { entity.Db().Callback().Query().Remove("test:empty-walk") })
	result, err := w.MatchFaces(entity.Faces{*f}, true, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, FacesMatchResult{}, result)
	assert.Zero(t, queries)
}
