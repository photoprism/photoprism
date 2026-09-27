package entity

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// statementCounter counts the SQL statements the database logger reports.
type statementCounter struct {
	sql  []string
	vars []int
}

// Print counts one statement per SQL log entry, with the number of values bound to it.
func (c *statementCounter) Print(values ...any) {
	if len(values) > 3 && values[0] == "sql" {
		c.sql = append(c.sql, fmt.Sprint(values[3]))

		if len(values) > 4 {
			vars, _ := values[4].([]any)
			c.vars = append(c.vars, len(vars))
		} else {
			c.vars = append(c.vars, 0)
		}
	}
}

// countStatements returns the SQL statements fn issues.
func countStatements(t *testing.T, fn func()) []string {
	t.Helper()

	return captureStatements(t, fn).sql
}

// captureStatements returns the SQL statements fn issues and the values bound to each.
func captureStatements(t *testing.T, fn func()) *statementCounter {
	t.Helper()

	c := &statementCounter{}
	db := Db()
	defer restoreDbLogger(db)()

	db.SetLogger(c)
	db.LogMode(true)

	fn()

	return c
}

// restoreDbLogger returns a function that restores the logger and log mode db has now. GORM keeps both
// unexported, so they are read and written through reflection.
func restoreDbLogger(db *gorm.DB) func() {
	field := func(name string) reflect.Value {
		f := reflect.ValueOf(db).Elem().FieldByName(name)
		return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem() //nolint:gosec // G103: test-only access to unexported GORM fields.
	}

	logger := reflect.New(field("logger").Type()).Elem()
	logger.Set(field("logger"))
	mode := field("logMode").Int()

	return func() {
		field("logger").Set(logger)
		field("logMode").SetInt(mode)
	}
}

// syncTestMarkers stores n automatic markers of subjUID in cluster f on the given file.
func syncTestMarkers(t *testing.T, f *Face, n int, subjUID, fileUID string) []string {
	t.Helper()

	uids := make([]string, 0, n)

	for i := range n {
		dist := 0.5 * f.AcceptDist()
		m := Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        fileUID,
			MarkerType:     MarkerFace,
			SubjUID:        subjUID,
			FaceID:         f.ID,
			FaceDist:       dist,
			EmbeddingsJSON: face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, uint64(20+i))}.JSON(),
			EmbedModel:     f.EmbedModel,
			Size:           face.ClusterSizeThreshold,
			Score:          face.ClusterScore("") + 10,
			MatchedAt:      TimeStamp(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, UnscopedDb().Create(&m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(Marker{}, "marker_uid = ?", m.MarkerUID) })

		uids = append(uids, m.MarkerUID)
	}

	return uids
}

// TestMarker_SyncSubject_Statements pins how many statements naming a marker in an unnamed cluster issues.
func TestMarker_SyncSubject_Statements(t *testing.T) {
	erin := NewSubject("Sync Statements Erin", SubjPerson, SrcManual)
	require.NoError(t, erin.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", erin.SubjUID) })

	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7801)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	uids := syncTestMarkers(t, f, 3, "", "fs6sg6bw45bn0001")
	m := FindMarker(uids[0])
	require.NotNil(t, m)

	statements := countStatements(t, func() {
		changed, err := m.SetName(erin.SubjName, SrcManual)
		require.NoError(t, err)
		require.True(t, changed)
	})

	// The subject lookup, naming the cluster, its automatic markers, and the refresh of their photos.
	assert.Len(t, statements, 4, "%v", statements)

	for _, q := range statements {
		assert.NotContains(t, q, `FROM "faces"`, "the cluster is named without loading it")
		assert.NotContains(t, q, "FROM `faces`", "the cluster is named without loading it")
	}

	assert.Equal(t, erin.SubjUID, FindFace(f.ID).SubjUID)
	assert.Equal(t, []string{erin.SubjUID, erin.SubjUID}, []string{FindMarker(uids[1]).SubjUID, FindMarker(uids[2]).SubjUID})
}

// TestMarker_SyncSubject_RefreshesPhotos pins that re-syncing a stale automatic marker in a cluster that
// already carries the person flags its picture for a metadata refresh.
func TestMarker_SyncSubject_RefreshesPhotos(t *testing.T) {
	fred := NewSubject("Sync Refresh Fred", SubjPerson, SrcManual)
	require.NoError(t, fred.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", fred.SubjUID) })

	f := NewFace(fred.SubjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(7802)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	file := FileFixtures.Get("exampleFileName.jpg")
	photo := FindPhoto(Photo{PhotoUID: file.PhotoUID})
	require.NotNil(t, photo)
	checked := photo.CheckedAt
	t.Cleanup(func() { UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("checked_at", checked) })

	named := syncTestMarkers(t, f, 1, fred.SubjUID, "fs6sg6bw45bn0001")
	stale := syncTestMarkers(t, f, 1, "", file.FileUID)

	now := TimeStamp()
	require.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("checked_at", now).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	changed, err := m.SetName(fred.SubjName, SrcManual)
	require.NoError(t, err)
	require.True(t, changed)

	assert.Equal(t, fred.SubjUID, FindMarker(stale[0]).SubjUID, "the stale sibling follows the cluster's person")

	refreshed := FindPhoto(Photo{PhotoUID: file.PhotoUID})
	require.NotNil(t, refreshed)
	assert.Nil(t, refreshed.CheckedAt, "and its picture is flagged for a refresh")
}

// TestMarker_SyncSubject_WithoutRelated pins that syncing without related markers, as matching does,
// neither loads a cluster named after someone else nor reports the marker to it.
func TestMarker_SyncSubject_WithoutRelated(t *testing.T) {
	gina := NewSubject("Sync Without Gina", SubjPerson, SrcManual)
	require.NoError(t, gina.Create())
	hank := NewSubject("Sync Without Hank", SubjPerson, SrcManual)
	require.NoError(t, hank.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid IN (?)", []string{gina.SubjUID, hank.SubjUID}) })

	f := NewFace(gina.SubjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(7803)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	uids := syncTestMarkers(t, f, 1, hank.SubjUID, "fs6sg6bw45bn0001")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", uids[0]).UpdateColumn("subj_src", SrcManual).Error)
	m := FindMarker(uids[0])
	require.NotNil(t, m)

	statements := countStatements(t, func() {
		require.NoError(t, m.SyncSubject(false))
	})

	for _, q := range statements {
		assert.NotContains(t, q, `FROM "faces"`)
		assert.NotContains(t, q, "FROM `faces`")
	}

	assert.Equal(t, f.ID, FindMarker(uids[0]).FaceID)
	assert.Zero(t, FindFace(f.ID).Collisions)
}

// TestMarker_SetFace_NamingRelinksSiblings pins that matching a marker named by hand into an unnamed
// cluster names the cluster and relinks its automatic markers.
func TestMarker_SetFace_NamingRelinksSiblings(t *testing.T) {
	ivy := NewSubject("Sync Review Ivy", SubjPerson, SrcManual)
	require.NoError(t, ivy.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", ivy.SubjUID) })

	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7811)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	sibling := syncTestMarkers(t, f, 1, "", "fs6sg6bw45bn0001")
	named := syncTestMarkers(t, f, 1, "", "fs6sg6bw45bn0001")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"face_id": "", "subj_src": SrcManual, "marker_name": ivy.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	require.Empty(t, m.FaceID)
	require.Empty(t, m.SubjUID)

	_, err := m.SetFace(f, 0.5*f.AcceptDist())
	require.NoError(t, err)

	assert.Equal(t, ivy.SubjUID, m.SubjUID)
	assert.Equal(t, ivy.SubjUID, FindFace(f.ID).SubjUID, "cluster named")
	assert.Equal(t, ivy.SubjUID, FindMarker(sibling[0]).SubjUID, "automatic sibling follows the named cluster")
}

// TestMarker_SyncSubject_MatchingDoesNotRelink pins that syncing without related markers names an
// unnamed cluster but leaves its automatic markers alone.
func TestMarker_SyncSubject_MatchingDoesNotRelink(t *testing.T) {
	jo := NewSubject("Sync Review Jo", SubjPerson, SrcManual)
	require.NoError(t, jo.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", jo.SubjUID) })

	f := NewFace("", SrcAuto, face.Embeddings{face.FixtureEmbedding(7812)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	sibling := syncTestMarkers(t, f, 1, "", "fs6sg6bw45bn0001")
	named := syncTestMarkers(t, f, 1, jo.SubjUID, "fs6sg6bw45bn0001")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"subj_src": SrcManual, "marker_name": jo.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)
	require.NoError(t, m.SyncSubject(false))

	assert.Equal(t, jo.SubjUID, FindFace(f.ID).SubjUID)
	assert.Empty(t, FindMarker(sibling[0]).SubjUID, "matching does not relink siblings")
}

// TestMarker_SyncSubject_NoRefreshWithoutChange pins that no picture is refreshed when no marker changed.
func TestMarker_SyncSubject_NoRefreshWithoutChange(t *testing.T) {
	kim := NewSubject("Sync Review Kim", SubjPerson, SrcManual)
	require.NoError(t, kim.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Subject{}, "subj_uid = ?", kim.SubjUID) })

	f := NewFace(kim.SubjUID, SrcAuto, face.Embeddings{face.FixtureEmbedding(7813)}, face.EmbeddingModelName())
	require.NoError(t, f.Create())
	t.Cleanup(func() { UnscopedDb().Delete(Face{}, "id = ?", f.ID) })

	syncTestMarkers(t, f, 1, kim.SubjUID, "fs6sg6bw45bn0001")
	named := syncTestMarkers(t, f, 1, kim.SubjUID, "fs6sg6bw45bn0001")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", named[0]).
		UpdateColumns(Values{"subj_src": SrcManual, "marker_name": kim.SubjName}).Error)

	m := FindMarker(named[0])
	require.NotNil(t, m)

	statements := countStatements(t, func() { require.NoError(t, m.SyncSubject(true)) })

	related := false

	for _, q := range statements {
		related = related || strings.Contains(q, `UPDATE "markers"`) || strings.Contains(q, "UPDATE `markers`")
		assert.False(t, strings.Contains(q, "photos"), "no refresh when nothing changed: %s", q)
	}

	assert.True(t, related, "the related markers were checked")
}

// TestRestoreDbLogger pins that capturing statements leaves the logger and log mode a caller had set.
func TestRestoreDbLogger(t *testing.T) {
	db := Db()
	t.Cleanup(restoreDbLogger(db))

	outer := &statementCounter{}
	db.SetLogger(outer)
	db.LogMode(true)

	inner := captureStatements(t, func() { FindFace("RESTOREDBLOGGER") })
	assert.NotEmpty(t, inner.sql, "the capture sees its own statements")

	FindFace("RESTOREDBLOGGERAFTER")
	require.NotEmpty(t, outer.sql, "the previous logger is active again")
	assert.Contains(t, outer.sql[len(outer.sql)-1], "faces")

	// A quiet log mode stays quiet after a capture.
	quiet := &statementCounter{}
	db.SetLogger(quiet)
	db.LogMode(false)
	captureStatements(t, func() { FindFace("RESTOREDBLOGGERQUIET") })
	FindFace("RESTOREDBLOGGERQUIETAFTER")
	assert.Empty(t, quiet.sql, "the previous log mode is active again")
}
