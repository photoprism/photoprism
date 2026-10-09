package entity

import (
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// reviseTestFiles are fixture files on different photos, so released markers flag distinct photos.
var reviseTestFiles = []string{"bridge.jpg", "reunion.jpg", "Quality1FavoriteTrue.jpg", "Photo18.jpg"}

// reviseTestState is what a revision leaves behind: marker columns and the photos flagged for maintenance.
type reviseTestState struct {
	Released []string
	Markers  map[string]string
	Photos   []uint
}

// reviseTestMarker stores a marker at dist from cluster f, as a fraction of what the cluster accepts.
func reviseTestMarker(t *testing.T, f *Face, frac float64, seed uint64, fileUID, subjUID, subjSrc string, model face.ModelName) Marker {
	t.Helper()

	dist := frac * f.AcceptDist()
	m := Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        fileUID,
		MarkerType:     MarkerFace,
		MarkerSrc:      SrcImage,
		SubjUID:        subjUID,
		SubjSrc:        subjSrc,
		FaceID:         f.ID,
		FaceDist:       dist,
		EmbeddingsJSON: face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, seed)}.JSON(),
		EmbedModel:     model,
		Size:           face.ClusterSizeThreshold,
		Score:          face.ClusterScore("") + 10,
		MatchedAt:      TimeStamp(),
		W:              0.1,
		H:              0.1,
	}

	require.NoError(t, UnscopedDb().Create(&m).Error)
	t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", m.MarkerUID) })

	return m
}

// reviseTestPhotos returns the photo IDs of the test files and marks them as checked.
func reviseTestPhotos(t *testing.T, fileUIDs []string) []uint {
	t.Helper()

	var files []File
	require.NoError(t, UnscopedDb().Where("file_uid IN (?)", fileUIDs).Find(&files).Error)

	ids := make([]uint, 0, len(files))
	checked := make(map[uint]*time.Time, len(files))

	for _, f := range files {
		var p Photo
		require.NoError(t, UnscopedDb().First(&p, "id = ?", f.PhotoID).Error)
		checked[p.ID] = p.CheckedAt
		ids = append(ids, p.ID)
	}

	t.Cleanup(func() {
		for id, at := range checked {
			UnscopedDb().Model(&Photo{}).Where("id = ?", id).UpdateColumn("checked_at", at)
		}
	})

	return ids
}

// reviseTestReset restores the stored markers and marks the photos as checked again.
func reviseTestReset(t *testing.T, markers []Marker, photos []uint) {
	t.Helper()

	for _, m := range markers {
		require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", m.MarkerUID).UpdateColumns(Values{
			"face_id": m.FaceID, "face_dist": m.FaceDist, "subj_uid": m.SubjUID, "subj_src": m.SubjSrc,
			"matched_at": m.MatchedAt, "updated_at": m.UpdatedAt,
		}).Error)
	}

	require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id IN (?)", photos).
		UpdateColumn("checked_at", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).Error)
}

// reviseTestSnapshot reads the state a revision left, comparing timestamps with the stored originals.
func reviseTestSnapshot(t *testing.T, revised Markers, markers []Marker, photos []uint) reviseTestState {
	t.Helper()

	s := reviseTestState{Markers: make(map[string]string, len(markers))}

	for _, r := range revised {
		s.Released = append(s.Released, r.MarkerUID)
	}

	sort.Strings(s.Released)

	for _, orig := range markers {
		var m Marker
		require.NoError(t, UnscopedDb().First(&m, "marker_uid = ?", orig.MarkerUID).Error)
		s.Markers[m.MarkerUID] = strings.Join([]string{
			m.FaceID, fmtFloat(m.FaceDist), m.SubjUID, m.SubjSrc,
			boolString(m.MatchedAt == nil), boolString(m.UpdatedAt.After(orig.UpdatedAt)),
		}, "|")
	}

	require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id IN (?) AND checked_at IS NULL", photos).
		Order("id").Pluck("id", &s.Photos).Error)

	return s
}

// fmtFloat formats a distance for a snapshot.
func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

// boolString formats a flag for a snapshot.
func boolString(b bool) string {
	if b {
		return "1"
	}

	return "0"
}

// reviseMatchesPerMarker releases what the cluster no longer matches one marker at a time, with
// ClearFace and Unmatched, as the reference the batched ReviseMatches must reproduce.
func reviseMatchesPerMarker(t *testing.T, m *Face) (revised Markers) {
	t.Helper()

	var matches Markers
	require.NoError(t, Db().Where("face_id = ?", m.ID).Where("marker_type = ?", MarkerFace).Find(&matches).Error)

	for _, marker := range matches {
		if !face.SameEmbeddingSpace(marker.EmbedModel, m.EmbedModel) {
			continue
		} else if ok, _ := m.Match(marker.Embeddings(), marker.EmbedModel); ok {
			continue
		}

		updated, err := marker.ClearFace()
		require.NoError(t, err)

		if updated {
			require.NoError(t, marker.Unmatched())
			revised = append(revised, marker)
		}
	}

	return revised
}

// TestFace_ReviseMatches_Batch pins that the batched revision releases the same markers, stores the
// same columns and flags the same photos as releasing them one at a time.
func TestFace_ReviseMatches_Batch(t *testing.T) {
	ann := NewSubject("Revise Batch Ann", SubjPerson, SrcManual)
	require.NoError(t, ann.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Subject{}, "subj_uid = ?", ann.SubjUID) })

	f := NewFace(ann.SubjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(7901)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

	fileUIDs := make([]string, len(reviseTestFiles))

	for i, name := range reviseTestFiles {
		fileUIDs[i] = FileFixtures.Get(name).FileUID
	}

	photos := reviseTestPhotos(t, fileUIDs)
	require.Len(t, photos, len(reviseTestFiles))

	// Kept, released automatic, released by hand, released from XMP, and one from another embedding
	// space; the last file holds only markers that stay, so its photo must not be flagged.
	var markers []Marker
	add := func(frac float64, file int, subjSrc string, model face.ModelName) {
		markers = append(markers, reviseTestMarker(t, f, frac, uint64(100+len(markers)), fileUIDs[file], ann.SubjUID, subjSrc, model))
	}

	add(0.3, 0, SrcAuto, f.EmbedModel)
	add(0.35, 3, SrcAuto, f.EmbedModel)
	add(0.4, 3, SrcManual, f.EmbedModel)
	add(0.9, 0, SrcAuto, f.EmbedModel)
	add(0.92, 1, SrcAuto, f.EmbedModel)
	add(0.95, 1, SrcManual, f.EmbedModel)
	add(0.97, 2, SrcXmp, f.EmbedModel)
	add(0.98, 2, SrcAuto, face.ModelArcFaceR50)

	// Narrow the cluster so only markers within 0.5 of what it accepted still match.
	f.CollisionRadius = 0.5 * f.AcceptDist()
	require.NoError(t, f.Updates(Values{"collision_radius": f.CollisionRadius}))
	f = FindFace(f.ID)
	require.NotNil(t, f)

	// Backdate the stored stamps, so a changed updated_at shows at second resolution.
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range markers {
		markers[i].UpdatedAt = past
		markers[i].MatchedAt = &past
	}

	reviseTestReset(t, markers, photos)
	want := reviseTestSnapshot(t, reviseMatchesPerMarker(t, f), markers, photos)

	reviseTestReset(t, markers, photos)
	UpdateFaces.Store(false)
	revised, err := f.ReviseMatches()
	require.NoError(t, err)
	assert.True(t, UpdateFaces.Load(), "a release flags the faces for maintenance")
	got := reviseTestSnapshot(t, revised, markers, photos)

	require.Len(t, want.Released, 4, "reference releases the four outside the narrowed cluster")
	assert.Equal(t, want, got)
	assert.Len(t, got.Photos, 3, "the photo holding only kept markers is not flagged")

	for _, r := range revised {
		assert.Empty(t, r.FaceID)
		assert.Equal(t, -1.0, r.FaceDist)
		assert.Nil(t, r.MatchedAt)
		assert.Equal(t, r.SubjSrc != SrcAuto, r.SubjUID != "", "automatic names are cleared in the returned copy")
	}
}

// TestFace_ReviseMatches_Batches pins that more released markers than two batches hold are all released
// with two statements per batch, none binding more markers than a batch holds.
func TestFace_ReviseMatches_Batches(t *testing.T) {
	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7902)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

	fileUID := FileFixtures.Get("bridge.jpg").FileUID
	photos := reviseTestPhotos(t, []string{fileUID})

	n := 2*BatchSize() + 1
	markers := make([]Marker, 0, n)

	for i := range n {
		markers = append(markers, Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        fileUID,
			MarkerType:     MarkerFace,
			MarkerSrc:      SrcImage,
			FaceID:         f.ID,
			FaceDist:       0.9 * f.AcceptDist(),
			EmbeddingsJSON: face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), 0.9*f.AcceptDist(), uint64(200+i))}.JSON(),
			EmbedModel:     f.EmbedModel,
			MatchedAt:      TimeStamp(),
		})
	}

	for i := range markers {
		require.NoError(t, UnscopedDb().Create(&markers[i]).Error)
	}

	t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "face_id = ? OR marker_uid IN (?)", f.ID, markerUIDs(markers)) })

	f.CollisionRadius = 0.5 * f.AcceptDist()
	require.NoError(t, f.Updates(Values{"collision_radius": f.CollisionRadius}))
	reviseTestReset(t, nil, photos)

	var revised Markers
	c := captureStatements(t, func() {
		var err error
		revised, err = f.ReviseMatches()
		require.NoError(t, err)
	})

	assert.Len(t, revised, n)
	assert.Len(t, c.sql, 1+3*2, "one select, then a marker and a photo update per batch: %v", c.sql)

	// A batch of UIDs plus, on the marker update, the cluster ID and the column values.
	for i, vars := range c.vars {
		assert.LessOrEqual(t, vars, BatchSize()+8, "statement %d binds one batch at most", i)
	}

	var left int64
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("face_id = ?", f.ID).Count(&left).Error)
	assert.Zero(t, left, "every marker is released")

	var flagged int64
	require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id IN (?) AND checked_at IS NULL", photos).Count(&flagged).Error)
	assert.Equal(t, int64(len(photos)), flagged, "their photo is flagged")
}

// TestFace_releaseMarkers pins that a release writes only markers still in the cluster.
func TestFace_releaseMarkers(t *testing.T) {
	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7903)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

	other := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7904)}, face.EmbeddingModelName())
	require.NoError(t, other.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", other.ID) })

	fileUID := FileFixtures.Get("bridge.jpg").FileUID
	reviseTestPhotos(t, []string{fileUID})

	t.Run("Empty", func(t *testing.T) {
		c := captureStatements(t, func() {
			released, err := f.releaseMarkers(nil, Now())
			require.NoError(t, err)
			assert.Zero(t, released)
		})
		assert.Empty(t, c.sql)
	})
	t.Run("OtherCluster", func(t *testing.T) {
		mine := reviseTestMarker(t, f, 0.9, 301, fileUID, "", SrcAuto, f.EmbedModel)
		moved := reviseTestMarker(t, other, 0.3, 302, fileUID, "", SrcAuto, other.EmbedModel)

		released, err := f.releaseMarkers([]string{mine.MarkerUID, moved.MarkerUID}, Now())
		require.NoError(t, err)
		assert.Equal(t, 1, released, "only the marker still in the cluster counts")

		assert.Empty(t, FindMarker(mine.MarkerUID).FaceID)
		kept := FindMarker(moved.MarkerUID)
		assert.Equal(t, other.ID, kept.FaceID, "a marker another cluster holds now stays there")
		assert.NotNil(t, kept.MatchedAt)
	})
}

// TestRefreshMarkerPhotos pins that only the photos of the specified markers are flagged.
func TestRefreshMarkerPhotos(t *testing.T) {
	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7905)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Face{}, "id = ?", f.ID) })

	a, b := FileFixtures.Get("bridge.jpg").FileUID, FileFixtures.Get("reunion.jpg").FileUID
	photos := reviseTestPhotos(t, []string{a, b})
	require.Len(t, photos, 2)

	flagged := reviseTestMarker(t, f, 0.3, 311, a, "", SrcAuto, f.EmbedModel)
	reviseTestMarker(t, f, 0.3, 312, b, "", SrcAuto, f.EmbedModel)

	t.Run("Empty", func(t *testing.T) {
		assert.NoError(t, refreshMarkerPhotos(nil))
	})
	t.Run("Success", func(t *testing.T) {
		reviseTestReset(t, nil, photos)
		require.NoError(t, refreshMarkerPhotos([]string{flagged.MarkerUID}))

		var ids []uint
		require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id IN (?) AND checked_at IS NULL", photos).Pluck("id", &ids).Error)
		assert.Equal(t, []uint{FileFixtures.Get("bridge.jpg").PhotoID}, ids)
	})
}

// markerUIDs returns the UIDs of the specified markers.
func markerUIDs(markers []Marker) []string {
	uids := make([]string, len(markers))

	for i, m := range markers {
		uids[i] = m.MarkerUID
	}

	return uids
}

// TestBatchSize pins the number of values a batch binds per database dialect.
func TestBatchSize(t *testing.T) {
	switch dialect := DbDialect(); dialect {
	case dsn.DialectSQLite:
		assert.Equal(t, 333, BatchSize())
	case dsn.DialectMySQL:
		assert.Equal(t, 1000, BatchSize())
	case dsn.DialectPostgreSQL:
		assert.Equal(t, 1000, BatchSize())
	default:
		t.Fatalf("unexpected dialect %s", dialect)
	}
}
